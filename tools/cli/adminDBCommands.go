package cli

import (
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	cassp "github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/service/history"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type scanFiles struct {
	failedToRunCheck *os.File
	startEventCorruptFile *os.File
}

// AdminShowWorkflow shows history
func AdminDBScan(c *cli.Context) {
	session := connectToCassandra(c)
	defer session.Close()
	waitGroup := sync.WaitGroup{}
	numCoroutine := 40
	waitGroup.Add(numCoroutine)

	failedToRunCheck, err := os.Create(fmt.Sprintf("failedToRunCheck.json"))
	if err != nil {
		panic(err.Error())
	}
	defer failedToRunCheck.Close()

	startEventCorruptFile, err := os.Create(fmt.Sprintf("startEventCorruptFile.json"))
	if err != nil {
		panic(err.Error())
	}
	defer startEventCorruptFile.Close()

	scanFiles := &scanFiles{
		failedToRunCheck: failedToRunCheck,
		startEventCorruptFile: startEventCorruptFile,
	}

	shard := make([]int, 0)
	for i := 0; i < 16384; i++ {
		shard = append(shard, i)
	}

	lock := &sync.Mutex{}

	var processingCount int64
	for i := 0; i < numCoroutine; i++ {
		go func(numRoutine int) {
			for {
				var shardID int
				lock.Lock()
				if len(shard) == 0 {
					lock.Unlock()
					break
				} else {
					shardID = shard[0]
					shard = shard[1:]
					lock.Unlock()
				}

				scanHelper(session, shardID, scanFiles, &processingCount)
			}
			waitGroup.Done()
		}(i)
	}
	waitGroup.Wait()
}

func scanHelper(session *gocql.Session, shardID int, scanFiles *scanFiles, count *int64) {
	serializer := persistence.NewPayloadSerializer()
	exeStore, err := cassp.NewWorkflowExecutionPersistence(shardID, session, loggerimpl.NewNopLogger())
	if err != nil {
		panic(err)
	}
	historyStore := cassp.NewHistoryV2PersistenceFromSession(session, loggerimpl.NewNopLogger())
	branchDecoder := codec.NewThriftRWEncoder()
	pageSize := 100
	var token []byte
	req := &persistence.ListConcreteExecutionsRequest{
		PageSize: pageSize,
		PageToken: token,
	}
	resp, err := exeStore.ListConcreteExecutions(req)
	if err != nil {
		scanFiles.dbOperationFailedFile.WriteString(fmt.Sprintf("call to ListConcreteExecutions failed: %v\r\n", err))
		return
	}
	for len(resp.NextPageToken) != 0 {
		// validate current batch
		for _, e := range resp.ExecutionInfos {
			var branch shared.HistoryBranch
			err := branchDecoder.Decode(e.BranchToken, &branch)
			if err != nil {
				scanFiles.dbOperationFailedFile.WriteString(fmt.Sprintf("failed to decode branch token: %v\r\n", err))
				continue
			}
			readHistoryBranchReq := &persistence.InternalReadHistoryBranchRequest{
				TreeID: branch.GetTreeID(),
				BranchID: branch.GetBranchID(),
				MinNodeID: 1,
				MaxNodeID: 20,
				ShardID: shardID,
				PageSize: pageSize,
			}
			_, err = historyStore.ReadHistoryBranch(readHistoryBranchReq)
			if err != nil && err == gocql.ErrNotFound {
				// here you actually want to write the corrupted record
				scanFiles.startEventCorruptFile.WriteString()
			}


			if err != nil {
				scanFiles.dbOperationFailedFile.WriteString(fmt.Sprintf("failed to read history branch: %v\r\n", err))
				continue
			}

		}


		// sleep
		// get the next batch
		req := &persistence.ListConcreteExecutionsRequest{
			PageSize: pageSize,
			PageToken: resp.NextPageToken,
		}
		resp, err = exeStore.ListConcreteExecutions(req)
		if err != nil {
			// TODO: figure out the correct thing to do in this case
			panic(err)
		}
	}






	for doPaging := true; doPaging; doPaging = len(token) > 0 {
		query := session.Query(getConcreteWorkflowQuery, shardID)
		iter := query.PageSize(pageSize).PageState(token).Iter()
		if iter == nil {
			panic("iter is nil")
		}

		var toBeScanned []ToBeScan

		for cont := true; cont; {
			execResult := make(map[string]interface{})
			if cont = iter.MapScan(execResult); cont {
				// filter default run id
				runID := execResult["run_id"].(gocql.UUID).String()
				if runID == "30000000-0000-f000-f000-000000000001" {
					continue
				}
				if _, ok := execResult["execution"]; ok {
					//rowCount++
					info := createWorkflowExecutionInfo(execResult["execution"].(map[string]interface{}))
					shardID := execResult["shard_id"].(int)
					toBeScanned = append(toBeScanned, ToBeScan{
						ShardID:        shardID,
						DomainID:       info.DomainID,
						WorkflowID:     info.WorkflowID,
						RunID:          runID,
						BranchToken:    info.BranchToken,
						LastUpdateTime: info.LastUpdatedTimestamp.String(),
						NextEventID:    info.NextEventID,
						CloseStatus:    info.CloseStatus,
						Notes:          Notes{},
					})
					//bs , _ := json.Marshal(info)
					//fmt.Println(string(bs))
				}
			}
		}

		token = make([]byte, len(iter.PageState()))
		copy(token, iter.PageState())

		//////Verify
		for _, execution := range toBeScanned {
			if len(execution.BranchToken) == 0 {
				// this is an error case that should not happen, all histories should be migrated
				// to events v2 and therefore all histories should have branch token
				execution.Notes.Note = "execution does not have BranchToken"
				bs, _ := json.Marshal(execution)
				scanFiles.dbOperationFailedFile.WriteString(string(bs))
				scanFiles.dbOperationFailedFile.WriteString("\r\n")
				continue
			}
			var branch shared.HistoryBranch
			err := Decode(execution.BranchToken, &branch)
			if err != nil {
				// this is an error that should not happen, we should always be able to decode branch token
				execution.Notes.Error = err.Error()
				execution.Notes.Note = "failed to decode branch token"
				bs, _ := json.Marshal(execution)
				scanFiles.dbOperationFailedFile.WriteString(string(bs))
				scanFiles.dbOperationFailedFile.WriteString("\r\n")
				continue
			}

			execution.TreeID = branch.GetTreeID()
			execution.BranchID = branch.GetBranchID()


			// now at this point you have everything you need to read history
			_, historyBatches, err := history.GetAllHistory(historyMgr, historyV2Mgr, nil, logger,
				true, execution.DomainID, execution.WorkflowID, execution.RunID,
				1, 2, storeVersion, execution.BranchToken)

			// if you determined history does not exist then treat this workflow as corrupted based on the startEvent
			if _, ok := err.(*shared.EntityNotExistsError); ok || (len(historyBatches) == 0 && err == nil) {
				if err != nil {
					execution.Notes.Error = err.Error()
				}
				bs, _ := json.Marshal(execution)
				scanFiles.startEventCorruptFile.WriteString(string(bs))
				scanFiles.startEventCorruptFile.WriteString("\r\n")
				continue

			}

			if err != nil {
				execution.Notes.Error = err.Error()
				bs, _ := json.Marshal(execution)
				scanFiles.startEventVerificationFailedFile.WriteString(string(bs))
				scanFiles.startEventVerificationFailedFile.WriteString("\r\n")
				continue
			}

		}
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt64(count, int64(len(result)))
		// TODO: add progress emitting number of verification fails so far
		printProgress(int(atomic.LoadInt64(count)), 148000000)
	}
	fmt.Println("Done with Shard: " + strconv.Itoa(shardID))
}

func printProgress(current int, total int) {
	fmt.Printf("\r\t\t%v\t\t/\t\t%v\t\t", current, total)
}
