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
	"encoding/json"
	"errors"
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

type (
	// ScanOutputDirectories are the directory paths for output of scan
	ScanOutputDirectories struct {
		ShardScanReportDirectoryPath       string
		ExecutionCheckFailureDirectoryPath string
		CorruptedExecutionDirectoryPath    string
	}

	// ShardScanOutputFiles are the files produced for a scan of a single shard
	ShardScanOutputFiles struct {
		ShardScanReportFile       *os.File
		ExecutionCheckFailureFile *os.File
		CorruptedExecutionFile    *os.File
	}


)

// AdminDBScan is used to scan over all executions in database and detect corruptions
func AdminDBScan(c *cli.Context) {
	lowerShardBound := c.Int(FlagLowerShardBound)
	upperShardBound := c.Int(FlagUpperShardBound)
	numShards := upperShardBound - lowerShardBound
	startingRPS := c.Int(FlagStartingRPS)
	targetRPS := c.Int(FlagRPS)
	scaleUpSeconds := c.Int(FlagRPSScaleUpSeconds)
	scanWorkerCount := c.Int(FlagConcurrency)
	executionsPageSize := c.Int(FlagPageSize)
	scanReportRate := c.Int(FlagReportRate)
	if numShards < scanWorkerCount {
		scanWorkerCount = numShards
	}

	// get a list of checkers here or an error if combination is not valid
	var checkers []AdminDBCheck

	payloadSerializer := persistence.NewPayloadSerializer()
	rateLimiter := getRateLimiter(startingRPS, targetRPS, scaleUpSeconds)
	session := connectToCassandra(c)
	defer session.Close()
	historyStore := cassp.NewHistoryV2PersistenceFromSession(session, loggerimpl.NewNopLogger())
	branchDecoder := codec.NewThriftRWEncoder()
	scanOutputDirectories := createScanOutputDirectories()

	reports := make(chan *ShardScanReport)
	for i := 0; i < scanWorkerCount; i++ {
		go func(workerIdx int) {
			for shardID := lowerShardBound; shardID < upperShardBound; shardID++ {
				if shardID%scanWorkerCount == workerIdx {
					reports <- scanShard(
						session,
						shardID,
						scanOutputDirectories,
						rateLimiter,
						executionsPageSize,
						payloadSerializer,
						historyStore,
						branchDecoder)
				}
			}
		}(i)
	}

	startTime := time.Now()
	progressReport := &ProgressReport{}
	for i := 0; i < numShards; i++ {
		report := <-reports
		includeShardInProgressReport(report, progressReport, startTime)
		if i%scanReportRate == 0 || i == numShards-1 {
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
	scanOutputDirectories *ScanOutputDirectories,
	limiter *quotas.DynamicRateLimiter,
	executionsPageSize int,
	payloadSerializer persistence.PayloadSerializer,
	historyStore persistence.HistoryStore,
	branchDecoder *codec.ThriftRWEncoder,
) *ShardScanReport {
	outputFiles, closeFn := createShardScanOutputFiles(shardID, scanOutputDirectories)
	report := &ShardScanReport{
		ShardID: shardID,
	}
	checkFailureWriter := NewBufferedWriter(outputFiles.ExecutionCheckFailureFile)
	corruptedExecutionWriter := NewBufferedWriter(outputFiles.CorruptedExecutionFile)
	defer func() {
		checkFailureWriter.Flush()
		corruptedExecutionWriter.Flush()
		recordShardScanReport(outputFiles.ShardScanReportFile, report)
		deleteEmptyFiles(outputFiles.CorruptedExecutionFile, outputFiles.ExecutionCheckFailureFile, outputFiles.ShardScanReportFile)
		closeFn()
	}()
	execStore, err := cassp.NewWorkflowExecutionPersistence(shardID, session, loggerimpl.NewNopLogger())
	if err != nil {
		report.Failure = &ShardScanReportFailure{
			Note:    "failed to create execution store",
			Details: err.Error(),
		}
		return report
	}

	var token []byte
	isFirstIteration := true
	for isFirstIteration || len(token) != 0 {
		isFirstIteration = false
		req := &persistence.ListConcreteExecutionsRequest{
			PageSize:  executionsPageSize,
			PageToken: token,
		}
		resp, err := (limiter, &report.TotalDBRequests, execStore, req)
		if err != nil {
			report.Failure = &ShardScanReportFailure{
				Note:    "failed to call ListConcreteExecutions",
				Details: err.Error(),
			}
			return report
		}
		token = resp.NextPageToken
		for _, e := range resp.Executions {
			if e == nil || e.ExecutionInfo == nil {
				continue
			}
			if report.Scanned == nil {
				report.Scanned = &ShardScanReportExecutionsScanned{}
			}
			report.Scanned.TotalExecutionsCount++
			historyVerificationResult, history, historyBranch := verifyHistoryExists(
				e,
				branchDecoder,
				corruptedExecutionWriter,
				checkFailureWriter,
				shardID,
				limiter,
				historyStore,
				&report.TotalDBRequests,
				execStore,
				payloadSerializer)
			switch historyVerificationResult {
			case VerificationResultNoCorruption:
				// nothing to do just keep checking other conditions
			case VerificationResultDetectedCorruption:
				report.Scanned.CorruptedExecutionsCount++
				report.Scanned.CorruptionTypeBreakdown.TotalHistoryMissing++
				if executionOpen(e.ExecutionInfo) {
					report.Scanned.OpenCorruptions.TotalOpen++
				}
				continue
			case VerificationResultCheckFailure:
				report.Scanned.ExecutionCheckFailureCount++
				continue
			}

			if history == nil || historyBranch == nil {
				continue
			}

			firstHistoryEventVerificationResult := verifyFirstHistoryEvent(
				e,
				historyBranch,
				corruptedExecutionWriter,
				checkFailureWriter,
				shardID,
				payloadSerializer,
				history)
			switch firstHistoryEventVerificationResult {
			case VerificationResultNoCorruption:
				// nothing to do just keep checking other conditions
			case VerificationResultDetectedCorruption:
				report.Scanned.CorruptionTypeBreakdown.TotalInvalidFirstEvent++
				report.Scanned.CorruptedExecutionsCount++
				if executionOpen(e.ExecutionInfo) {
					report.Scanned.OpenCorruptions.TotalOpen++
				}
				continue
			case VerificationResultCheckFailure:
				report.Scanned.ExecutionCheckFailureCount++
				continue
			}

			currentExecutionVerificationResult := verifyCurrentExecution(
				e,
				corruptedExecutionWriter,
				checkFailureWriter,
				shardID,
				historyBranch,
				execStore,
				limiter,
				&report.TotalDBRequests)
			switch currentExecutionVerificationResult {
			case VerificationResultNoCorruption:
				// nothing to do just keep checking other conditions
			case VerificationResultDetectedCorruption:
				report.Scanned.CorruptionTypeBreakdown.TotalOpenExecutionInvalidCurrentExecution++
				report.Scanned.CorruptedExecutionsCount++
				if executionOpen(e.ExecutionInfo) {
					report.Scanned.OpenCorruptions.TotalOpen++
				}
				continue
			case VerificationResultCheckFailure:
				report.Scanned.ExecutionCheckFailureCount++
				continue
			}
		}
	}
	return report
}

func deleteEmptyFiles(files ...*os.File) {
	shouldDelete := func(filepath string) bool {
		fi, err := os.Stat(filepath)
		return err == nil && fi.Size() == 0
	}
	for _, f := range files {
		if shouldDelete(f.Name()) {
			os.Remove(f.Name())
		}
	}
}

func createShardScanOutputFiles(shardID int, sod *ScanOutputDirectories) (*ShardScanOutputFiles, func()) {
	executionCheckFailureFile, err := os.Create(fmt.Sprintf("%v/%v", sod.ExecutionCheckFailureDirectoryPath, constructFileNameFromShard(shardID)))
	if err != nil {
		ErrorAndExit("failed to create executionCheckFailureFile", err)
	}
	shardScanReportFile, err := os.Create(fmt.Sprintf("%v/%v", sod.ShardScanReportDirectoryPath, constructFileNameFromShard(shardID)))
	if err != nil {
		ErrorAndExit("failed to create shardScanReportFile", err)
	}
	corruptedExecutionFile, err := os.Create(fmt.Sprintf("%v/%v", sod.CorruptedExecutionDirectoryPath, constructFileNameFromShard(shardID)))
	if err != nil {
		ErrorAndExit("failed to create corruptedExecutionFile", err)
	}

	deferFn := func() {
		executionCheckFailureFile.Close()
		shardScanReportFile.Close()
		corruptedExecutionFile.Close()
	}
	return &ShardScanOutputFiles{
		ShardScanReportFile:       shardScanReportFile,
		ExecutionCheckFailureFile: executionCheckFailureFile,
		CorruptedExecutionFile:    corruptedExecutionFile,
	}, deferFn
}

func constructFileNameFromShard(shardID int) string {
	return fmt.Sprintf("shard_%v.json", shardID)
}

func createScanOutputDirectories() *ScanOutputDirectories {
	now := time.Now().Unix()
	sod := &ScanOutputDirectories{
		ShardScanReportDirectoryPath:       fmt.Sprintf("./scan_%v/shard_scan_report", now),
		ExecutionCheckFailureDirectoryPath: fmt.Sprintf("./scan_%v/execution_check_failure", now),
		CorruptedExecutionDirectoryPath:    fmt.Sprintf("./scan_%v/corrupted_execution", now),
	}
	if err := os.MkdirAll(sod.ShardScanReportDirectoryPath, 0777); err != nil {
		ErrorAndExit("failed to create ShardScanFailureDirectoryPath", err)
	}
	if err := os.MkdirAll(sod.ExecutionCheckFailureDirectoryPath, 0777); err != nil {
		ErrorAndExit("failed to create ExecutionCheckFailureDirectoryPath", err)
	}
	if err := os.MkdirAll(sod.CorruptedExecutionDirectoryPath, 0777); err != nil {
		ErrorAndExit("failed to create CorruptedExecutionDirectoryPath", err)
	}
	fmt.Println("scan results located under: ", fmt.Sprintf("./scan_%v", now))
	return sod
}

func recordShardScanReport(file *os.File, ssr *ShardScanReport) {
	data, err := json.Marshal(ssr)
	if err != nil {
		ErrorAndExit("failed to marshal ShardScanReport", err)
	}
	writeToFile(file, string(data))
}

func writeToFile(file *os.File, message string) {
	if _, err := file.WriteString(fmt.Sprintf("%v\r\n", message)); err != nil {
		ErrorAndExit("failed to write to file", err)
	}
}

func getRateLimiter(startRPS int, targetRPS int, scaleUpSeconds int) *quotas.DynamicRateLimiter {
	if startRPS >= targetRPS {
		ErrorAndExit("startRPS is greater than target RPS", nil)
	}
	if scaleUpSeconds == 0 {
		return quotas.NewDynamicRateLimiter(func() float64 { return float64(targetRPS) })
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

func getHistoryBranch(
	e *persistence.InternalListConcreteExecutionsEntity,
	payloadSerializer persistence.PayloadSerializer,
	branchDecoder *codec.ThriftRWEncoder,
) (*shared.HistoryBranch, error) {
	branchTokenBytes := e.ExecutionInfo.BranchToken
	if len(branchTokenBytes) == 0 {
		if e.VersionHistories == nil {
			return nil, errors.New("failed to get branch token")
		}
		vh, err := payloadSerializer.DeserializeVersionHistories(e.VersionHistories)
		if err != nil {
			return nil, err
		}
		branchTokenBytes = vh.GetHistories()[vh.GetCurrentVersionHistoryIndex()].GetBranchToken()
	}
	var branch shared.HistoryBranch
	if err := branchDecoder.Decode(branchTokenBytes, &branch); err != nil {
		return nil, err
	}
	return &branch, nil
}
