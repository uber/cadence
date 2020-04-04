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

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	cassp "github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/common/quotas"
)

const (
	historyPageSize = 50
)

type (
	scanFiles struct {
		failedToRunCheckFile  *os.File
		startEventCorruptFile *os.File
	}

	// ExecutionScanEntity is the execution entity which gets written to output file from scan
	ExecutionScanEntity struct {
		ShardID      int
		DomainID     string
		WorkflowID   string
		RunID        string
		NextEventID  int64
		TreeID       string
		BranchID     string
		CloseStatus  int
		ScanMetadata ExecutionScanEntityMetadata
	}

	// ExecutionScanEntityMetadata is the metadata from scanning the execution that gets written to output file
	ExecutionScanEntityMetadata struct {
		Message  string
		ErrorMsg string
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
	numShards := getRequiredIntOption(c, FlagNumberOfShards)
	startingRPS := getRequiredIntOption(c, FlagStartingRPS)
	targetRPS := getRequiredIntOption(c, FlagRPS)
	scaleUpSeconds := getRequiredIntOption(c, FlagRPSScaleUpSeconds)
	scanWorkerCount := getRequiredIntOption(c, FlagGoRoutineCount)
	executionsPageSize := getRequiredIntOption(c, FlagPageSize)
	scanReportRate := getRequiredIntOption(c, FlagScanReportRate)

	rateLimiter := getRateLimiter(startingRPS, targetRPS, scaleUpSeconds)
	scanFiles, deferFn := createScanFiles()
	session := connectToCassandra(c)
	defer func() {
		deferFn()
		session.Close()
	}()

	shardReports := make(chan *ShardReport)
	for i := 0; i < scanWorkerCount; i++ {
		go func(workerIdx int) {
			for shardID := 0; shardID < numShards; shardID++ {
				if shardID%scanWorkerCount == workerIdx {
					shardReports <- scanShard(session, shardID, scanFiles, rateLimiter, executionsPageSize)
				}
			}
		}(i)
	}

	progressReport := &ProgressReport{}
	for i := 0; i < numShards; i++ {
		report := <-shardReports
		includeShardInProgressReport(report, progressReport)
		if i%scanReportRate == 0 {
			fmt.Printf("%+v", progressReport)
		}
	}
}

func scanShard(
	session *gocql.Session,
	shardID int,
	scanFiles *scanFiles,
	limiter *quotas.DynamicRateLimiter,
	executionsPageSize int,
) *ShardReport {
	execStore, err := cassp.NewWorkflowExecutionPersistence(shardID, session, loggerimpl.NewNopLogger())
	if err != nil {
		ErrorAndExit("failed to create execution persistence", err)
	}
	historyStore := cassp.NewHistoryV2PersistenceFromSession(session, loggerimpl.NewNopLogger())
	branchDecoder := codec.NewThriftRWEncoder()
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
			report.NumberOfExecutions++
			var branch shared.HistoryBranch
			err := branchDecoder.Decode(e.BranchToken, &branch)
			if err != nil {
				report.NumberOfFailedChecks++
				writeToFile(scanFiles.failedToRunCheckFile, fmt.Sprintf("failed to decode branch token: %v", err))
				continue
			}
			readHistoryBranchReq := &persistence.InternalReadHistoryBranchRequest{
				TreeID:    branch.GetTreeID(),
				BranchID:  branch.GetBranchID(),
				MinNodeID: 1,
				MaxNodeID: 20,
				ShardID:   shardID,
				PageSize:  historyPageSize,
			}
			limiter.Wait(context.Background())
			_, err = historyStore.ReadHistoryBranch(readHistoryBranchReq)
			if err != nil {
				if err == gocql.ErrNotFound {
					metadata := ExecutionScanEntityMetadata{
						Message:  "Detected workflow was corrupted based on missing history",
						ErrorMsg: err.Error(),
					}
					report.NumberOfCorruptedExecutions++
					writeExecutionToFile(scanFiles.startEventCorruptFile, shardID, branch.GetTreeID(), branch.GetBranchID(), metadata, e)
				} else {
					metadata := ExecutionScanEntityMetadata{
						Message:  "Checking corruption based on start event failed",
						ErrorMsg: err.Error(),
					}
					report.NumberOfFailedChecks++
					writeExecutionToFile(scanFiles.failedToRunCheckFile, shardID, branch.GetTreeID(), branch.GetBranchID(), metadata, e)
				}
			}
		}
	}
	return report
}

func writeToFile(file *os.File, message string) {
	if _, err := file.WriteString(fmt.Sprintf("%v\r\n", message)); err != nil {
		ErrorAndExit("failed to write to file", err)
	}
}

func writeExecutionToFile(
	file *os.File,
	shardID int,
	treeID string,
	branchID string,
	metadata ExecutionScanEntityMetadata,
	info *persistence.InternalWorkflowExecutionInfo,
) {
	exec := ExecutionScanEntity{
		ShardID:      shardID,
		DomainID:     info.DomainID,
		WorkflowID:   info.WorkflowID,
		RunID:        info.RunID,
		NextEventID:  info.NextEventID,
		TreeID:       treeID,
		BranchID:     branchID,
		CloseStatus:  info.CloseStatus,
		ScanMetadata: metadata,
	}
	data, err := json.Marshal(exec)
	if err != nil {
		ErrorAndExit("failed to marshal execution", err)
	}
	writeToFile(file, string(data))
}

func createScanFiles() (*scanFiles, func()) {
	failedToRunCheckFile, err := os.Create(fmt.Sprintf("failedToRunCheck.json"))
	if err != nil {
		ErrorAndExit("failed to create file", err)
	}
	startEventCorruptFile, err := os.Create(fmt.Sprintf("startEventCorruptFile.json"))
	if err != nil {
		ErrorAndExit("failed to create file", err)
	}
	deferFn := func() {
		failedToRunCheckFile.Close()
		startEventCorruptFile.Close()
	}

	return &scanFiles{
		failedToRunCheckFile:  failedToRunCheckFile,
		startEventCorruptFile: startEventCorruptFile,
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
