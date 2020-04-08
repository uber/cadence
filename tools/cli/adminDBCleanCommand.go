package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	cassp "github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/common/quotas"
	"github.com/urfave/cli"
	"math"
	"os"
	"time"
)

type (
	// ShardCleanReport represents the result of cleaning a single shard
	ShardCleanReport struct {
		ShardID         int
		TotalDBRequests int64
		Handled *ShardCleanReportHandled
		Failure *ShardCleanReportFailure
	}

	// ShardCleanReportHandled is the part of ShardCleanReport of executions which were read from corruption file
	// and were attempted to be deleted
	ShardCleanReportHandled struct {
		TotalExecutionsCount int64
		SuccessfullyCleanedCount int64
		FailedCleanedCount int64
	}

	// ShardCleanReportFailure is the part of ShardCleanReport that indicates a failure to clean some or all
	// of the executions found in corruption file
	ShardCleanReportFailure struct {
		Note string
		Details string
	}

	// CleanProgressReport represents the aggregate progress of the clean job.
	// It is periodically printed to stdout
	CleanProgressReport struct {
		NumberOfShardsFinished     int
		TotalExecutionsCount       int64
		SuccessfullyCleanedCount int64
		FailedCleanedCount int64
		TotalDBRequests            int64
		DatabaseRPS                float64
		NumberOfShardCleanFailures int64
		ShardsPerHour              float64
		ExecutionsPerHour          float64
	}

	// CleanOutputDirectories are the directory paths for output of clean
	CleanOutputDirectories struct {
		ShardCleanReportDirectoryPath       string
		SuccessfullyCleanedDirectoryPath    string
		FailedCleanedDirectoryPath          string
	}

	// ShardCleanOutputFiles are the files produced for a clean of a single shard
	ShardCleanOutputFiles struct {
		ShardCleanReportFile       *os.File
		SuccessfullyCleanedFile    *os.File
		FailedCleanedFile          *os.File
	}
)

func AdminDBClean(c *cli.Context) {
	session := connectToCassandra(c)
	execStore, err := cassp.NewWorkflowExecutionPersistence(0, session, loggerimpl.NewNopLogger())
	if err != nil {
		panic(err)
	}
	req := &persistence.DeleteWorkflowExecutionRequest{
		DomainID: "foo",
		WorkflowID: "bar",
		RunID: "baz",
	}
	err = execStore.DeleteWorkflowExecution(req)
	fmt.Println("andrew got error from deleting workflow execution: ", err)



	//lowerShardBound := c.Int(FlagLowerShardBound)
	//upperShardBound := c.Int(FlagUpperShardBound)
	//numShards := upperShardBound - lowerShardBound
	//startingRPS := c.Int(FlagStartingRPS)
	//targetRPS := c.Int(FlagRPS)
	//scaleUpSeconds := c.Int(FlagRPSScaleUpSeconds)
	//scanWorkerCount := c.Int(FlagConcurrency)
	//scanReportRate := c.Int(FlagReportRate)
	//if numShards < scanWorkerCount {
	//	scanWorkerCount = numShards
	//}
	//inputDirectory := getRequiredOption(c, FlagInputDirectory)
	//
	//rateLimiter := getRateLimiter(startingRPS, targetRPS, scaleUpSeconds)
	//session := connectToCassandra(c)
	//defer session.Close()
	//cleanOutputDirectories := createCleanOutputDirectories()

	// start concurrency number of workers
	// each worker will call handleShard
		// handle shard will construct the file in question
		// if the file does not exist return right away with a zero value ShardDeleteReport
		// if the file exists then read it line by line
		// for each line deserialize it into a CorruptExecution
		// if there is an error doing this then increment delete failed count and move on to the next one
		// if you successfully deserialize it then switch on the corruption type
			// if the corruption if of type history missing or history start event invalid then delete current mutable state and concrete mutable state
			// if the corruption is of type current execution error then delete concrete mutable state and move on
		// return report
	// in the main we will fetch these reports and we will always get numShards reports even if they are empty
	// every X reports will will add this report into a sum progress report and emit it to stdout

	//reports := make(chan *ShardCleanReport)
	//for i := 0; i < scanWorkerCount; i++ {
	//	go func(workerIdx int) {
	//		for shardID := lowerShardBound; shardID < upperShardBound; shardID++ {
	//			if shardID%scanWorkerCount == workerIdx {
	//				reports <- handleCleanShard(
	//					rateLimiter,
	//					session,
	//					cleanOutputDirectories,
	//					inputDirectory,
	//					shardID)
	//			}
	//		}
	//	}(i)
	//}
	//
	//startTime := time.Now()
	//progressReport := &CleanProgressReport{}
	//for i := 0; i < numShards; i++ {
	//	report := <-reports
	//	includeShardCleanInProgressReport(report, progressReport, startTime)
	//	if i%scanReportRate == 0 {
	//		reportBytes, err := json.MarshalIndent(*progressReport, "", "\t")
	//		if err != nil {
	//			ErrorAndExit("failed to print progress", err)
	//		}
	//		fmt.Println(string(reportBytes))
	//	}
	//}
}

func handleCleanShard(
	limiter *quotas.DynamicRateLimiter,
	session *gocql.Session,
	outputDirectories *CleanOutputDirectories,
	inputDirectory string,
	shardID int,
) *ShardCleanReport {
	outputFiles, closeFn := createShardCleanOutputFiles(shardID, outputDirectories)
	report := &ShardCleanReport{
		ShardID: shardID,
	}
	failedCleanWriter := NewAdminDBCorruptedExecutionBufferedWriter(outputFiles.FailedCleanedFile)
	successfullyCleanWriter := NewAdminDBCorruptedExecutionBufferedWriter(outputFiles.SuccessfullyCleanedFile)
	defer func() {
		failedCleanWriter.Flush()
		successfullyCleanWriter.Flush()
		recordShardCleanReport(outputFiles.ShardCleanReportFile, report)
		deleteEmptyFiles(outputFiles.ShardCleanReportFile, outputFiles.SuccessfullyCleanedFile, outputFiles.FailedCleanedFile)
		closeFn()
	}()
	shardCorruptedFile, err := getShardCorruptedFile(inputDirectory, shardID)
	if err != nil {
		report.Failure = &ShardCleanReportFailure{
			Note:    "failed to get corruption file",
			Details: err.Error(),
		}
		return report
	}
	//execStore, err := cassp.NewWorkflowExecutionPersistence(shardID, session, loggerimpl.NewNopLogger())
	if err != nil {
		report.Failure = &ShardCleanReportFailure{
			Note:    "failed to create execution store",
			Details: err.Error(),
		}
		return report
	}

	scanner := bufio.NewScanner(shardCorruptedFile)
	for scanner.Scan() {
		if report.Handled == nil {
			report.Handled = &ShardCleanReportHandled{}
		}
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		var ce CorruptedExecution
		err := json.Unmarshal([]byte(line), &ce)
		if err != nil {
			report.Handled.FailedCleanedCount++
			continue
		}

		//deleteConcreteReq := &persistence.DeleteWorkflowExecutionRequest{
		//	DomainID: ce.DomainID,
		//	WorkflowID: ce.WorkflowID,
		//	RunID: ce.RunID,
		//}
		//err := execStore.DeleteWorkflowExecution(deleteConcreteReq)
		//if err != nil {
		//
		//}
	}
	return nil
}

func getShardCorruptedFile(inputDir string, shardID int) (*os.File, error) {
	filepath := fmt.Sprintf("%v/%v", inputDir, constructFileNameFromShard(shardID))
	return os.Open(filepath)
}

func includeShardCleanInProgressReport(report *ShardCleanReport, progressReport *CleanProgressReport, startTime time.Time) {
	progressReport.NumberOfShardsFinished++
	progressReport.TotalDBRequests += report.TotalDBRequests
	if report.Failure != nil {
		progressReport.NumberOfShardCleanFailures++
	}

	if report.Handled != nil {
		progressReport.TotalExecutionsCount += report.Handled.TotalExecutionsCount
		progressReport.FailedCleanedCount += report.Handled.FailedCleanedCount
		progressReport.SuccessfullyCleanedCount += report.Handled.SuccessfullyCleanedCount
	}

	pastTime := time.Now().Sub(startTime)
	hoursPast := float64(pastTime) / float64(time.Hour)
	progressReport.ShardsPerHour = math.Round(float64(progressReport.NumberOfShardsFinished) / hoursPast)
	progressReport.ExecutionsPerHour = math.Round(float64(progressReport.TotalExecutionsCount) / hoursPast)
	secondsPast := float64(pastTime) / float64(time.Second)
	progressReport.DatabaseRPS = math.Round(float64(progressReport.TotalDBRequests) / secondsPast)
}

func createShardCleanOutputFiles(shardID int, cod *CleanOutputDirectories) (*ShardCleanOutputFiles, func()) {
	shardCleanReportFile, err := os.Create(fmt.Sprintf("%v/%v", cod.ShardCleanReportDirectoryPath, constructFileNameFromShard(shardID)))
	if err != nil {
		ErrorAndExit("failed to create ShardCleanReportFile", err)
	}
	successfullyCleanedFile, err := os.Create(fmt.Sprintf("%v/%v", cod.SuccessfullyCleanedDirectoryPath, constructFileNameFromShard(shardID)))
	if err != nil {
		ErrorAndExit("failed to create SuccessfullyCleanedFile", err)
	}
	failedCleanedFile, err := os.Create(fmt.Sprintf("%v/%v", cod.FailedCleanedDirectoryPath, constructFileNameFromShard(shardID)))
	if err != nil {
		ErrorAndExit("failed to create FailedCleanedFile", err)
	}

	deferFn := func() {
		shardCleanReportFile.Close()
		successfullyCleanedFile.Close()
		failedCleanedFile.Close()
	}
	return &ShardCleanOutputFiles{
		ShardCleanReportFile:       shardCleanReportFile,
		SuccessfullyCleanedFile: successfullyCleanedFile,
		FailedCleanedFile:    failedCleanedFile,
	}, deferFn
}

func createCleanOutputDirectories() *CleanOutputDirectories {
	now := time.Now().Unix()
	cod := &CleanOutputDirectories{
		ShardCleanReportDirectoryPath:       fmt.Sprintf("./clean_%v/shard_scan_report", now),
		SuccessfullyCleanedDirectoryPath: fmt.Sprintf("./clean_%v/successfully_cleaned", now),
		FailedCleanedDirectoryPath:    fmt.Sprintf("./clean_%v/failed_cleaned", now),
	}
	if err := os.MkdirAll(cod.ShardCleanReportDirectoryPath, 0766); err != nil {
		ErrorAndExit("failed to create ShardCleanReportDirectoryPath", err)
	}
	if err := os.MkdirAll(cod.SuccessfullyCleanedDirectoryPath, 0766); err != nil {
		ErrorAndExit("failed to create SuccessfullyCleanedDirectoryPath", err)
	}
	if err := os.MkdirAll(cod.FailedCleanedDirectoryPath, 0766); err != nil {
		ErrorAndExit("failed to create FailedCleanedDirectoryPath", err)
	}
	fmt.Println("clean results located under: ", fmt.Sprintf("./clean_%v", now))
	return cod
}

func recordShardCleanReport(file *os.File, sdr *ShardCleanReport) {
	data, err := json.Marshal(sdr)
	if err != nil {
		ErrorAndExit("failed to marshal ShardCleanReport", err)
	}
	writeToFile(file, string(data))
}