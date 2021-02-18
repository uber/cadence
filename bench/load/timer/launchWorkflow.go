// Copyright (c) 2017-2021 Uber Technologies Inc.

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

package timer

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/cadence"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/uber/cadence/bench/lib"
	"github.com/uber/cadence/bench/load/common"
)

const (
	// TestName is the test name for timer test
	TestName = "timer"

	// LauncherWorkflowName is the workflow name for launching timer test
	LauncherWorkflowName = "timer-load-test-workflow"
)

type (
	launcherActivityProgress struct {
		WorkflowStarted int
		NextStartID     int
	}
)

// RegisterLauncher registers workflows and activities for timer load launching
func RegisterLauncher(w worker.Worker) {
	w.RegisterWorkflowWithOptions(launcherWorkflow, workflow.RegisterOptions{Name: LauncherWorkflowName})
	w.RegisterActivity(launcherActivity)
	w.RegisterActivity(validationActivity)
}

func launcherWorkflow(ctx workflow.Context, config lib.TimerTestConfig) (float64, error) {
	testTimeout := time.Duration(workflow.GetInfo(ctx).ExecutionStartToCloseTimeoutSeconds) * time.Second

	if testTimeout <= time.Duration(config.LongestTimerDurationInSeconds+config.TimerTimeoutInSeconds)*time.Second {
		return 0, cadence.NewCustomError("Test timeout too short, need to be longer than LongestTimerDuration + TimerTimeout")
	}

	totalLaunchCount := config.TotalTimerCount / config.TimerPerWorkflow
	workflowsPerActivity := totalLaunchCount / config.RoutineCount
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Hour,
		StartToCloseTimeout:    testTimeout,
		HeartbeatTimeout:       20 * time.Second,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumAttempts:    10,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	startTime := workflow.Now(ctx)
	futures := make([]workflow.Future, 0, config.RoutineCount)
	for i := 0; i != config.RoutineCount; i++ {
		futures = append(futures, workflow.ExecuteActivity(ctx, launcherActivity, i, workflowsPerActivity, startTime, config))
	}

	var totalStartedWorkflow int
	for _, future := range futures {
		var workflowStarted int
		if err := future.Get(ctx, &workflowStarted); err != nil {
			return 0, err
		}

		totalStartedWorkflow += workflowStarted
	}

	// wait until the timer fires and timerWorkflow completes
	waitDuration := startTime.Add(time.Duration(config.LongestTimerDurationInSeconds+config.TimerTimeoutInSeconds) * time.Second).Sub(workflow.Now(ctx))
	if err := workflow.NewTimer(ctx, waitDuration).Get(ctx, nil); err != nil {
		return 0, err
	}

	// start the validation phase

	// 1. check if enough workflow are started
	if float64(totalStartedWorkflow) < common.DefaultAvailabilityThreshold*float64(totalLaunchCount) {
		return 0, cadence.NewCustomError(
			common.ErrReasonValidationFailed,
			fmt.Sprintf("Too few workflows are started. Expected: %v, actual: %v", totalLaunchCount, totalStartedWorkflow),
		)
	}

	// 2. run an activity to check the status of started workflow by querying visibility records
	// this checks 3 things:
	//     (1) number of workflow in visibility matches the one returned from launcher activity
	//     (2) if all started workflow has closed (meaning fire has fired and no workflow stuck).
	//     (3) how many timers are fired within the MaxTimerLatency limit
	validationActivityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2,
			MaximumAttempts:          5,
			NonRetriableErrorReasons: []string{common.ErrReasonValidationFailed},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, validationActivityOptions)

	var firedInTimeRatio float64
	if err := workflow.ExecuteActivity(ctx, validationActivity, startTime.Add(-10*time.Second).UnixNano(), totalStartedWorkflow).Get(ctx, &firedInTimeRatio); err != nil {
		return 0, err
	}

	return firedInTimeRatio, nil
}

func launcherActivity(
	ctx context.Context,
	routineID int,
	launchCount int,
	startTime time.Time,
	config lib.TimerTestConfig,
) (int, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Start launcher activity", zap.Int("routineID", routineID))
	cadenceClient := ctx.Value(lib.CtxKeyCadenceClient).(lib.CadenceClient)
	runtimeContext := ctx.Value(lib.CtxKeyRuntimeContext).(*lib.RuntimeContext)
	numTaskList := runtimeContext.Bench.NumTaskLists
	input := WorkflowParams{
		TimerCount:           config.TimerPerWorkflow,
		EarliesTimerFireTime: startTime.Add(time.Duration(config.ShortestTimerDurationInSeconds) * time.Second),
		LatestTimerFireTime:  startTime.Add(time.Duration(config.LongestTimerDurationInSeconds) * time.Second),
		MaxTimerLatency:      time.Duration(config.MaxTimerLatencyInSeconds) * time.Second,
	}
	workflowOptions := client.StartWorkflowOptions{
		ExecutionStartToCloseTimeout:    72 * time.Hour,
		DecisionTaskStartToCloseTimeout: 5 * time.Minute,
	}

	var progress launcherActivityProgress
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &progress); err != nil {
			logger.Error("Failed to get activity heartbeat details", zap.Int("routineID", routineID), zap.Error(err))
			progress = launcherActivityProgress{
				WorkflowStarted: 0,
				NextStartID:     0,
			}
		} else {
			logger.Info("Successfully loaded activity progress", zap.Int("routineID", routineID), zap.Any("progress", progress))
		}
	}

	for i := progress.NextStartID; i < launchCount; i++ {
		workflowOptions.ID = fmt.Sprintf("%s-%d-%s", TestName, routineID, uuid.New())
		workflowOptions.TaskList = common.GetTaskListName(rand.Intn(numTaskList))

		_ = common.RetryOp(func() error {
			startCtx, cancel := context.WithTimeout(ctx, common.DefaultContextTimeout)
			_, err := cadenceClient.StartWorkflow(startCtx, workflowOptions, timerWorkflowName, input)
			cancel()

			if err == nil || cadence.IsWorkflowExecutionAlreadyStartedError(err) {
				progress.WorkflowStarted++
				return nil
			}

			logger.Error("Failed to start workflow execution", zap.Error(err))
			return err
		}, nil)

		progress.NextStartID++
		activity.RecordHeartbeat(ctx, progress)
		time.Sleep(time.Duration(75+rand.Intn(50)) * time.Millisecond)
	}

	logger.Info("Completed launcher workflow", zap.Int("routineID", routineID))
	return progress.WorkflowStarted, nil
}

func validationActivity(
	ctx context.Context,
	startTimeNanos int64,
	startedWorkflow int,
) (float64, error) {
	cc := ctx.Value(lib.CtxKeyCadenceClient).(lib.CadenceClient)

	domain := activity.GetInfo(ctx).WorkflowDomain

	// 1. check if enough workflows are started
	query := fmt.Sprintf("WorkflowType = '%s' and StartTime > %v", timerWorkflowName, startTimeNanos)
	request := &shared.CountWorkflowExecutionsRequest{
		Domain: &domain,
		Query:  &query,
	}
	resp, err := cc.CountWorkflow(ctx, request)
	if err != nil {
		return 0, err
	}
	actualLaunchedCount := resp.GetCount()
	if actualLaunchedCount < int64(startedWorkflow) {
		return 0, cadence.NewCustomError(common.ErrReasonValidationFailed, fmt.Sprintf("Expected to start %v workflow, actual started: %v", startedWorkflow, actualLaunchedCount))
	}

	// 2. check if timer has fired and workflow has closed
	query = fmt.Sprintf("WorkflowType = '%s' and StartTime > %v and CloseTime != missing", timerWorkflowName, startTimeNanos)
	request.Query = &query
	resp, err = cc.CountWorkflow(ctx, request)
	if err != nil {
		return 0, err
	}
	if resp.GetCount() != actualLaunchedCount {
		return 0, cadence.NewCustomError(common.ErrReasonValidationFailed, fmt.Sprintf("Not all workflows are closed, started: %v, closed: %v", actualLaunchedCount, resp.GetCount()))
	}

	// 3. check # of timers that are fired within max latency
	query = fmt.Sprintf("WorkflowType = '%s' and StartTime > %v and CloseStatus = 0", timerWorkflowName, startTimeNanos)
	request.Query = &query
	resp, err = cc.CountWorkflow(ctx, request)
	if err != nil {
		return 0, err
	}
	firedInTime := resp.GetCount()
	if firedInTime < int64(float64(actualLaunchedCount)*common.DefaultAvailabilityThreshold) {
		return 0, cadence.NewCustomError(common.ErrReasonValidationFailed, fmt.Sprintf("Too few timers fired within latency limit, expected: %v, actual: %v", actualLaunchedCount, firedInTime))
	}

	return float64(firedInTime) / float64(actualLaunchedCount), nil
}
