package concurrentexec

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/bench/lib"
	"github.com/uber/cadence/bench/load/common"
)

const (
	// TestName is the test name for concurrent execution bench test
	TestName = "concurrent-execution"

	// LauncherWorkflowName is the workflow name for launching concurrent execution bench test
	LauncherWorkflowName = "concurrent-execution-test-workflow"
)

type (
	batchResult struct {
		score     float64
		errString string
	}
)

// RegisterLauncher registers workflows for launching concurrent execution load
func RegisterLauncher(w worker.Worker) {
	w.RegisterWorkflowWithOptions(concurrentExecTestWorkflow, workflow.RegisterOptions{Name: LauncherWorkflowName})
}

func concurrentExecTestWorkflow(
	ctx workflow.Context,
	config lib.ConcurrentExecTestConfig,
) (float64, error) {
	batchCompletionCh := workflow.NewChannel(ctx)

	workflow.Go(ctx, func(ctx workflow.Context) {
		startBatchWorkflow(ctx, config, batchCompletionCh)
	})

	totalScore := 0.0
	for i := 0; i != config.TotalBatches; i++ {
		var result batchResult
		batchCompletionCh.Receive(ctx, &result)
		if len(result.errString) != 0 {
			return 0, fmt.Errorf("batch workflow execution failed: %v", result.errString)
		}
		totalScore += result.score
	}

	avgScore := totalScore / float64(config.TotalBatches)
	if avgScore < common.DefaultAvailabilityThreshold {
		return 0, fmt.Errorf("batch workflow score too low, expected: %v, actual: %v", common.DefaultAvailabilityThreshold, avgScore)
	}

	return avgScore, nil
}

func startBatchWorkflow(
	ctx workflow.Context,
	config lib.ConcurrentExecTestConfig,
	batchCompletionCh workflow.Channel,
) {
	var numTaskList int
	var err error
	if numTaskList, err = getTaskListNum(ctx); err != nil {
		batchCompletionCh.Send(ctx, batchResult{
			errString: "Failed to getTaskListNum, error: " + err.Error(),
		})
		return
	}

	numConcurrentBatches := config.TotalBatches / config.Concurrency
	for i := 0; i != numConcurrentBatches; i++ {
		workflow.Go(ctx, func(ctx workflow.Context) {
			startConcurrentBatches(ctx, config, batchCompletionCh, i*config.Concurrency, numTaskList)
		})

		if i != numConcurrentBatches-1 {
			timer := workflow.NewTimer(ctx, time.Duration(config.BatchPeriodInSeconds)*time.Second)
			if err := timer.Get(ctx, nil); err != nil {
				batchCompletionCh.Send(ctx, batchResult{
					errString: "Failed to start batch workflow, error: " + err.Error(),
				})
				return
			}
		}
	}
}

func startConcurrentBatches(
	ctx workflow.Context,
	config lib.ConcurrentExecTestConfig,
	batchCompletionCh workflow.Channel,
	batchStartIdx int,
	numTaskList int,
) {
	parentWorkflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	now := workflow.Now(ctx).UnixNano()

	childFutures := make([]workflow.Future, 0, config.Concurrency)
	for batchIdx := batchStartIdx; batchIdx != batchStartIdx+config.Concurrency; batchIdx++ {
		cwo := workflow.ChildWorkflowOptions{
			WorkflowID:                   parentWorkflowID + "-batch-" + strconv.Itoa(batchIdx),
			TaskList:                     common.GetTaskListName(rand.Intn(numTaskList)),
			ExecutionStartToCloseTimeout: time.Duration(config.BatchTimeoutInSeconds) * time.Second,
			TaskStartToCloseTimeout:      time.Minute,
			ParentClosePolicy:            client.ParentClosePolicyTerminate,
			WorkflowIDReusePolicy:        client.WorkflowIDReusePolicyAllowDuplicate,
		}
		childCtx := workflow.WithChildOptions(ctx, cwo)
		childFuture := workflow.ExecuteChildWorkflow(childCtx, batchWorkflowName, config, now)
		childFutures = append(childFutures, childFuture)
	}

	for _, childFuture := range childFutures {
		var batchScore float64
		if err := childFuture.Get(ctx, &batchScore); err != nil {
			batchCompletionCh.Send(ctx, batchResult{
				errString: "Batch workflow failed, error: " + err.Error(),
			})
		} else {
			batchCompletionCh.Send(ctx, batchResult{
				score: batchScore,
			})
		}
	}
}
