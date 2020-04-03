package cli

import (
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
)

const (
	scanWorkerCount            = 40
	numberOfShards             = 16384
	delayBetweenExecutionPages = 50 * time.Millisecond
	executionPageSize          = 100
	historyPageSize            = 50
)

type (
	scanFiles struct {
		failedToRunCheckFile  *os.File
		startEventCorruptFile *os.File
		progressFile          *os.File
	}

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

	ExecutionScanEntityMetadata struct {
		Message  string
		ErrorMsg string
	}

	ScanShardReport struct {
		ShardId                     int
		NumberOfExecutions          int
		NumberOfCorruptedExecutions int
		NumberOfFailedChecks        int
		FailedToListExecutions      bool
	}
)

// AdminDBScan is used to scan over all executions in database and detect corruptions
func AdminDBScan(c *cli.Context) {
	scanFiles, deferFn := createScanFiles()
	session := connectToCassandra(c)
	defer func() {
		deferFn()
		session.Close()
	}()

	shardReports := make(chan *ScanShardReport)
	for i := 0; i < scanWorkerCount; i++ {
		go func(workerIdx int) {
			for shardId := 0; shardId < numberOfShards; shardId++ {
				if shardId%scanWorkerCount == workerIdx {
					shardReports <- scanShard(session, shardId, scanFiles)
				}
			}
		}(i)
	}

	combinedShardReport := &ScanShardReport{}
	for i := 0; i < numberOfShards; i++ {
		report := <-shardReports
		combineShardReport(report, combinedShardReport)
		writeScanReportToFile(scanFiles.progressFile, *report, *combinedShardReport)
	}
}

func scanShard(session *gocql.Session, shardId int, scanFiles *scanFiles) *ScanShardReport {
	execStore, err := cassp.NewWorkflowExecutionPersistence(shardId, session, loggerimpl.NewNopLogger())
	if err != nil {
		ErrorAndExit("failed to create execution persistence", err)
	}
	historyStore := cassp.NewHistoryV2PersistenceFromSession(session, loggerimpl.NewNopLogger())
	branchDecoder := codec.NewThriftRWEncoder()
	var token []byte
	scanShardReport := &ScanShardReport{
		ShardId: shardId,
	}
	isFirstIteration := true
	for isFirstIteration || len(token) != 0 {
		isFirstIteration = false
		req := &persistence.ListConcreteExecutionsRequest{
			PageSize:  executionPageSize,
			PageToken: token,
		}
		resp, err := execStore.ListConcreteExecutions(req)
		if err != nil {
			writeToFile(scanFiles.failedToRunCheckFile, fmt.Sprintf("call to ListConcreteExecutions failed: %v", err))
			scanShardReport.FailedToListExecutions = true
			return scanShardReport
		}
		token = resp.NextPageToken

		for _, e := range resp.ExecutionInfos {
			scanShardReport.NumberOfExecutions++
			var branch shared.HistoryBranch
			err := branchDecoder.Decode(e.BranchToken, &branch)
			if err != nil {
				scanShardReport.NumberOfFailedChecks++
				writeToFile(scanFiles.failedToRunCheckFile, fmt.Sprintf("failed to decode branch token: %v", err))
				continue
			}
			readHistoryBranchReq := &persistence.InternalReadHistoryBranchRequest{
				TreeID:    branch.GetTreeID(),
				BranchID:  branch.GetBranchID(),
				MinNodeID: 1,
				MaxNodeID: 20,
				ShardID:   shardId,
				PageSize:  historyPageSize,
			}
			_, err = historyStore.ReadHistoryBranch(readHistoryBranchReq)
			if err != nil {
				if err == gocql.ErrNotFound {
					metadata := ExecutionScanEntityMetadata{
						Message:  "Detected workflow was corrupted based on missing history",
						ErrorMsg: err.Error(),
					}
					scanShardReport.NumberOfCorruptedExecutions++
					writeExecutionToFile(scanFiles.startEventCorruptFile, shardId, branch.GetTreeID(), branch.GetBranchID(), metadata, e)
				} else {
					metadata := ExecutionScanEntityMetadata{
						Message:  "Checking corruption based on start event failed",
						ErrorMsg: err.Error(),
					}
					scanShardReport.NumberOfFailedChecks++
					writeExecutionToFile(scanFiles.failedToRunCheckFile, shardId, branch.GetTreeID(), branch.GetBranchID(), metadata, e)
				}
			}
		}
		time.Sleep(delayBetweenExecutionPages)
	}
	return scanShardReport
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
		ErrorAndExit("failed to marshal exeuction", err)
	}
	writeToFile(file, string(data))
}

func writeScanReportToFile(
	file *os.File,
	report ScanShardReport,
	combined ScanShardReport,
) {
	reportData, err := json.Marshal(report)
	if err != nil {
		ErrorAndExit("failed to marshal scanShardReport", err)
	}
	combinedData, err := json.Marshal(combined)
	if err != nil {
		ErrorAndExit("failed to marshal combined scanShardReport", err)
	}
	writeToFile(file, string(reportData))
	writeToFile(file, string(combinedData))
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
	progressFile, err := os.Create(fmt.Sprintf("progressFile.json"))
	if err != nil {
		ErrorAndExit("failed to create file", err)
	}
	deferFn := func() {
		failedToRunCheckFile.Close()
		startEventCorruptFile.Close()
		progressFile.Close()
	}

	return &scanFiles{
		failedToRunCheckFile:  failedToRunCheckFile,
		startEventCorruptFile: startEventCorruptFile,
		progressFile:          progressFile,
	}, deferFn
}

func combineShardReport(report *ScanShardReport, combined *ScanShardReport) {
	if combined.ShardId < report.ShardId {
		combined.ShardId = report.ShardId
	}
	combined.FailedToListExecutions = combined.FailedToListExecutions || report.FailedToListExecutions
	combined.NumberOfCorruptedExecutions = combined.NumberOfCorruptedExecutions + report.NumberOfCorruptedExecutions
	combined.NumberOfFailedChecks = combined.NumberOfFailedChecks + report.NumberOfFailedChecks
	combined.NumberOfExecutions = combined.NumberOfExecutions + report.NumberOfExecutions
}
