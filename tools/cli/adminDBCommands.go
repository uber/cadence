// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gocql/gocql"
	"github.com/urfave/cli"
	"github.com/uber/cadence/common"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	cassp "github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/common/quotas"
)

const (
	historyPageSize = 50
	failedToRunCheckFilename = "failedToRunCheck.json"
	startEventCorruptedFilename = "startEventCorrupted.json"
)

type (
	scanFiles struct {
		failedToRunCheckFile  *os.File
		startEventCorruptedFile *os.File
	}

	// CorruptedExecutionEntity is a corrupted execution
	CorruptedExecutionEntity struct {
		ShardID      int
		DomainID     string
		WorkflowID   string
		RunID        string
		NextEventID  int64
		TreeID       string
		BranchID     string
		CloseStatus  int
		Note         string
	}

	// ScanFailure is a scan failure
	ScanFailure struct {
		Note string
		Details string
	}

	// ProgressReport contains metadata about the scan for all shards which have been finished
	ProgressReport struct {
		NumberOfShardsFinished           int
		NumberOfExecutions               int
		NumberOfCorruptedExecutions      int
		NumberOfFailedChecks             int
		NumberOfShardsFailedToFinishScan int
	}

	// ShardReport contains metadata about the scan for a single shard
	ShardReport struct {
		NumberOfExecutions          int
		NumberOfCorruptedExecutions int
		NumberOfFailedChecks        int
		FailedToFinishScan          bool
	}
)

// AdminDBScan is used to scan over all executions in database and detect corruptions
func AdminDBScan(c *cli.Context) {
	numShards := c.Int(FlagNumberOfShards)
	startingRPS := c.Int(FlagStartingRPS)
	targetRPS := c.Int(FlagRPS)
	scaleUpSeconds := c.Int(FlagRPSScaleUpSeconds)
	scanWorkerCount := c.Int(FlagGoRoutineCount)
	executionsPageSize := c.Int(FlagPageSize)
	scanReportRate := c.Int(FlagScanReportRate)

	payloadSerializer := persistence.NewPayloadSerializer()
	rateLimiter := getRateLimiter(startingRPS, targetRPS, scaleUpSeconds)
	scanFiles, deferFn := createScanFiles()
	session := connectToCassandra(c)
	historyStore := cassp.NewHistoryV2PersistenceFromSession(session, loggerimpl.NewNopLogger())
	branchDecoder := codec.NewThriftRWEncoder()
	defer func() {
		deferFn()
		session.Close()
	}()

	shardReports := make(chan *ShardReport)
	for i := 0; i < scanWorkerCount; i++ {
		go func(workerIdx int) {
			for shardID := 0; shardID < numShards; shardID++ {
				if shardID%scanWorkerCount == workerIdx {
					shardReports <- scanShard(session, shardID, scanFiles, rateLimiter, executionsPageSize, payloadSerializer, historyStore, branchDecoder)
				}
			}
		}(i)
	}

	progressReport := &ProgressReport{}
	for i := 0; i < numShards; i++ {
		report := <-shardReports
		includeShardInProgressReport(report, progressReport)
		if i%scanReportRate == 0 {
			reportBytes, err := json.MarshalIndent(*progressReport, "", "\t")
			if err != nil {
				ErrorAndExit("failed to print progress", err)
			}
			fmt.Println(string(reportBytes))
		}
	}
}

func scanShard(
	session *gocql.Session,
	shardID int,
	scanFiles *scanFiles,
	limiter *quotas.DynamicRateLimiter,
	executionsPageSize int,
	payloadSerializer persistence.PayloadSerializer,
	historyStore persistence.HistoryStore,
	branchDecoder *codec.ThriftRWEncoder,
) *ShardReport {
	execStore, err := cassp.NewWorkflowExecutionPersistence(shardID, session, loggerimpl.NewNopLogger())
	if err != nil {
		ErrorAndExit("failed to create execution persistence", err)
	}
	var token []byte
	report := &ShardReport{}
	isFirstIteration := true
	for isFirstIteration || len(token) != 0 {
		isFirstIteration = false
		req := &persistence.ListConcreteExecutionsRequest{
			PageSize:  executionsPageSize,
			PageToken: token,
		}
		limiter.Wait(context.Background())
		resp, err := execStore.ListConcreteExecutions(req)
		if err != nil {
			report.FailedToFinishScan = true
			writeToFile(scanFiles.failedToRunCheckFile, fmt.Sprintf("call to ListConcreteExecutions failed: %v", err))
			return report
		}
		token = resp.NextPageToken

		for _, e := range resp.ExecutionInfos {
			verifyExecution(report, e, branchDecoder, scanFiles, shardID, limiter, historyStore, payloadSerializer)
		}
	}
	return report
}

func verifyExecution(
	report *ShardReport,
	execution *persistence.InternalWorkflowExecutionInfo,
	branchDecoder *codec.ThriftRWEncoder,
	scanFiles *scanFiles,
	shardID int,
	limiter *quotas.DynamicRateLimiter,
	historyStore persistence.HistoryStore,
	payloadSerializer persistence.PayloadSerializer,
) {
	report.NumberOfExecutions++
	var branch shared.HistoryBranch
	err := branchDecoder.Decode(execution.BranchToken, &branch)
	if err != nil {
		report.NumberOfFailedChecks++
		writeToFile(scanFiles.failedToRunCheckFile, fmt.Sprintf("failed to decode branch token: %v", err))
		return
	}
	readHistoryBranchReq := &persistence.InternalReadHistoryBranchRequest{
		TreeID:    branch.GetTreeID(),
		BranchID:  branch.GetBranchID(),
		MinNodeID: common.FirstEventID,
		MaxNodeID: common.EndEventID,
		ShardID:   shardID,
		PageSize:  historyPageSize,
	}
	limiter.Wait(context.Background())
	history, err := historyStore.ReadHistoryBranch(readHistoryBranchReq)
	if err != nil {
		if err == gocql.ErrNotFound {
			report.NumberOfCorruptedExecutions++
			recordCorruptedWorkflow(
				scanFiles.startEventCorruptedFile,
				shardID,
				branch.GetTreeID(),
				branch.GetBranchID(),
				"got not found error",
				execution)
		} else {
			report.NumberOfFailedChecks++
			recordScanFailure(scanFiles.failedToRunCheckFile, "got error from read history other than not exists error", err)
		}
	} else if history == nil || len(history.History) == 0 {
		ErrorAndExit("got no error from fetching history but history is empty", nil)
	} else {
		firstEvent := history.History[0]
		historyEvent, err := payloadSerializer.DeserializeEvent(firstEvent)
		if err != nil {
			report.NumberOfFailedChecks++
			recordScanFailure(scanFiles.failedToRunCheckFile, "got error decoding history event", err)
		} else if historyEvent.GetEventId() != common.FirstEventID || historyEvent.GetEventType() != shared.EventTypeWorkflowExecutionStarted {
			report.NumberOfCorruptedExecutions++
			recordCorruptedWorkflow(
				scanFiles.startEventCorruptedFile,
				shardID,
				branch.GetTreeID(),
				branch.GetBranchID(),
				"got workflow with incorrect first event",
				execution)
		}
	}
}

func recordCorruptedWorkflow(
	file *os.File,
	shardID int,
	treeID string,
	branchID string,
	note string,
	info *persistence.InternalWorkflowExecutionInfo,
) {
	cee := CorruptedExecutionEntity{
		ShardID:      shardID,
		DomainID:     info.DomainID,
		WorkflowID:   info.WorkflowID,
		RunID:        info.RunID,
		NextEventID:  info.NextEventID,
		TreeID:       treeID,
		BranchID:     branchID,
		CloseStatus:  info.CloseStatus,
		Note:         note,
	}
	data, err := json.Marshal(cee)
	if err != nil {
		ErrorAndExit("failed to marshal CorruptedExecutionEntity", err)
	}
	writeToFile(file, string(data))
}

func recordScanFailure(file *os.File, note string, err error) {
	sf := ScanFailure{
		Note: note,
	}
	if err != nil {
		sf.Details = err.Error()
	}
	data, marshalErr := json.Marshal(sf)
	if marshalErr != nil {
		ErrorAndExit("failed to marshal ScanFailure", marshalErr)
	}
	writeToFile(file, string(data))
}

func writeToFile(file *os.File, message string) {
	if _, err := file.WriteString(fmt.Sprintf("%v\r\n", message)); err != nil {
		ErrorAndExit("failed to write to file", err)
	}
}

func createScanFiles() (*scanFiles, func()) {
	failedToRunCheckFile, err := os.Create(fmt.Sprintf(failedToRunCheckFilename))
	if err != nil {
		ErrorAndExit("failed to create file", err)
	}
	startEventCorruptedFile, err := os.Create(fmt.Sprintf(startEventCorruptedFilename))
	if err != nil {
		ErrorAndExit("failed to create file", err)
	}
	deferFn := func() {
		failedToRunCheckFile.Close()
		startEventCorruptedFile.Close()
	}

	return &scanFiles{
		failedToRunCheckFile:  failedToRunCheckFile,
		startEventCorruptedFile: startEventCorruptedFile,
	}, deferFn
}

func includeShardInProgressReport(report *ShardReport, progressReport *ProgressReport) {
	progressReport.NumberOfCorruptedExecutions += report.NumberOfCorruptedExecutions
	progressReport.NumberOfFailedChecks += report.NumberOfFailedChecks
	progressReport.NumberOfExecutions += report.NumberOfExecutions
	progressReport.NumberOfShardsFinished++
	if report.FailedToFinishScan {
		progressReport.NumberOfShardsFailedToFinishScan++
	}
}

func getRateLimiter(startRPS int, targetRPS int, scaleUpSeconds int) *quotas.DynamicRateLimiter {
	if startRPS >= targetRPS {
		ErrorAndExit("startRPS is greater than target RPS", nil)
	}
	rpsIncreasePerSecond := (targetRPS - startRPS) / scaleUpSeconds
	startTime := time.Now()
	rpsFn := func() float64 {
		secondsPast := int(time.Now().Sub(startTime).Seconds())
		if secondsPast >= scaleUpSeconds {
			return float64(targetRPS)
		}
		return float64((rpsIncreasePerSecond * secondsPast) + startRPS)
	}
	return quotas.NewDynamicRateLimiter(rpsFn)
}
