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

package cancellation

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/pborman/uuid"
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
	// TestName is the test name for cancellation test
	TestName = "cancellation"

	// LauncherWorkflowName is the workflow name for launching cancellation load test
	LauncherWorkflowName = "cancellation-load-test-workflow"
)

const (
	sleepWorkflowName = "cancellation-sleep-workflow"

	waitDurationInMilliSeconds       = 75
	waitDurationJitterInMilliSeconds = 50

	minActivityTimeout = time.Minute

	defaultDurationBeforeCancellation = 10 * time.Second
	defaultWorkflowSleepDuration      = 20 * time.Second
	defaultDurationBeforeValidation   = 30 * time.Second
)

type (
	launcherActivityResult struct {
		StartAvailability  float64
		CancelAvailability float64
	}

	startWorkflowProgress struct {
		TotalStartWorkflowCall     int
		SucceededStartWorkflowCall int
		WorkflowStarted            int
		NextStartID                int
	}
)

// RegisterLauncher registers workflows for launching cancellation load
func RegisterLauncher(w worker.Worker) {
	w.RegisterWorkflowWithOptions(launcherWorkflow, workflow.RegisterOptions{Name: LauncherWorkflowName})
	w.RegisterActivity(launcherActivity)
	w.RegisterActivity(validationActivity)
}

// RegisterWorker registers workflows for cancellation test
func RegisterWorker(w worker.Worker) {
	w.RegisterWorkflowWithOptions(sleepWorkflow, workflow.RegisterOptions{Name: sleepWorkflowName})
}

func launcherWorkflow(
	ctx workflow.Context,
	config lib.CancellationTestConfig,
) error {
	workflowPerActivity := config.TotalLaunchCount / config.Concurrency

	activityStartToCloseTimeout := time.Duration(workflowPerActivity/(1000/waitDurationInMilliSeconds)) * time.Second * common.DefaultMaxRetryCount
	if activityStartToCloseTimeout < minActivityTimeout {
		activityStartToCloseTimeout = minActivityTimeout
	}

	launcherActivityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    activityStartToCloseTimeout,
		HeartbeatTimeout:       30 * time.Second,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumAttempts:    10,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, launcherActivityOptions)

	startTime := workflow.Now(ctx)

	futures := make([]workflow.Future, 0, config.Concurrency)
	for i := 0; i < config.Concurrency; i++ {
		future := workflow.ExecuteActivity(ctx, launcherActivity, config)
		futures = append(futures, future)
	}

	avgStartAvailability := 0.0
	avgCancelAvailability := 0.0
	for _, future := range futures {
		var result launcherActivityResult
		if err := future.Get(ctx, &result); err != nil {
			return fmt.Errorf("launcherActivity failed: %v", err)
		}
		avgStartAvailability += result.StartAvailability
		avgCancelAvailability += result.CancelAvailability
	}
	avgStartAvailability /= float64(config.Concurrency)
	avgCancelAvailability /= float64(config.Concurrency)

	// validate availability
	if avgStartAvailability < common.DefaultAvailabilityThreshold {
		return fmt.Errorf("startWorkflow availability too low, required: %v, actual: %v", common.DefaultAvailabilityThreshold, avgStartAvailability)
	}
	if avgCancelAvailability < common.DefaultAvailabilityThreshold {
		return fmt.Errorf("cancelWorkflow availability too low, required: %v, actual: %v", common.DefaultAvailabilityThreshold, avgCancelAvailability)
	}

	// validate if there's stuck workflow using visibility records

	// give the system some time to propagate ES records and wait for workflow that failed to cancel to finish
	if err := workflow.Sleep(ctx, defaultDurationBeforeValidation); err != nil {
		return fmt.Errorf("launcher workflow sleep failed: %v", err)
	}

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
	// move startTime backward by 10 secs to account for the time drift between worker and cadence hosts if any
	return workflow.ExecuteActivity(ctx, validationActivity, config, startTime.Add(-time.Second*10).UnixNano()).Get(ctx, nil)
}

func launcherActivity(
	ctx context.Context,
	config lib.CancellationTestConfig,
) (launcherActivityResult, error) {
	if config.ContextTimeoutInSeconds == 0 {
		config.ContextTimeoutInSeconds = int(common.DefaultContextTimeout.Seconds())
	}

	startRPS := 1000 / waitDurationInMilliSeconds
	workflowIDChan := make(chan string, startRPS*int(defaultDurationBeforeCancellation.Seconds()))

	var startAvailability float64
	go func() {
		startAvailability = startWorkflow(ctx, &config, workflowIDChan)
	}()

	time.Sleep(defaultDurationBeforeCancellation)

	cancelAvailability := cancelWorkflow(ctx, &config, workflowIDChan)

	return launcherActivityResult{
		StartAvailability:  startAvailability,
		CancelAvailability: cancelAvailability,
	}, nil
}

func startWorkflow(
	ctx context.Context,
	config *lib.CancellationTestConfig,
	workflowIDChan chan string,
) float64 {
	defer close(workflowIDChan)

	logger := activity.GetLogger(ctx)

	var progress startWorkflowProgress
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &progress); err != nil {
			// explicitly reset value to 0 if failed to get details, in case the implementation changed the value
			logger.Error("Failed to get activity heartbeat details", zap.Error(err))
			progress = startWorkflowProgress{
				TotalStartWorkflowCall:     0,
				SucceededStartWorkflowCall: 0,
				WorkflowStarted:            0,
				NextStartID:                0,
			}
		}
	}

	cc := ctx.Value(lib.CtxKeyCadenceClient).(lib.CadenceClient)
	rc := ctx.Value(lib.CtxKeyRuntimeContext).(*lib.RuntimeContext)
	numTaskList := rc.Bench.NumTaskLists

	for i := progress.NextStartID; i < config.TotalLaunchCount/config.Concurrency; i++ {
		select {
		case <-ctx.Done():
			// unable to start specified # of workflows in time.
			// Workflow will receive timeout error, the value we return here is irrelevant.
			logger.Error("Failed to start all workflows in time", zap.Any("progress", progress))
			return 0
		default:
		}

		startWorkflowOptions := client.StartWorkflowOptions{
			ExecutionStartToCloseTimeout: 2 * defaultWorkflowSleepDuration,
			ID:                           fmt.Sprintf("%s-%s", TestName, uuid.New()),
			TaskList:                     common.GetTaskListName(rand.Intn(numTaskList)),
		}

		if err := common.RetryOp(func() error {
			progress.TotalStartWorkflowCall++

			startCtx, cancel := context.WithTimeout(ctx, time.Duration(config.ContextTimeoutInSeconds)*time.Second)
			we, err := cc.StartWorkflow(startCtx, startWorkflowOptions, sleepWorkflowName, defaultWorkflowSleepDuration)
			cancel()

			if err == nil || cadence.IsWorkflowExecutionAlreadyStartedError(err) {
				workflowIDChan <- we.ID
				progress.SucceededStartWorkflowCall++
				return nil
			}
			if common.IsServiceBusyError(err) {
				// do not count service busy as failure
				progress.SucceededStartWorkflowCall++
			}

			logger.Error("Failed to start workflow execution", zap.Error(err))
			return err
		}, nil); err == nil {
			progress.WorkflowStarted++
		}

		// successfully started the workflow or gave up after several retries
		progress.NextStartID++
		activity.RecordHeartbeat(ctx, progress)
		time.Sleep(time.Duration(waitDurationInMilliSeconds+rand.Intn(waitDurationJitterInMilliSeconds)) * time.Millisecond)
	}

	availability := float64(progress.SucceededStartWorkflowCall) / float64(progress.TotalStartWorkflowCall)
	logger.Info("Completed start workflow", zap.Float64("availability", availability), zap.Int("workflow-started", progress.WorkflowStarted))
	return availability
}

func cancelWorkflow(
	ctx context.Context,
	config *lib.CancellationTestConfig,
	workflowIDChan chan string,
) float64 {
	logger := activity.GetLogger(ctx)

	totalCancelWorkflowCall := 0
	succeededCancelWorkflowCall := 0
	workflowCancelled := 0

	cc := ctx.Value(lib.CtxKeyCadenceClient).(lib.CadenceClient)

	for {
		select {
		case <-ctx.Done():
			// unable to start specified # of workflows in time.
			// Workflow will receive timeout error, the value we return here is irrelevant.
			logger.Error("Failed to cancel all workflows in time")
			return 0
		case workflowID := <-workflowIDChan:
			if workflowID == "" {
				// channel has closed, all workflow cancelled
				availability := float64(succeededCancelWorkflowCall) / float64(totalCancelWorkflowCall)
				logger.Info("Completed cancel workflow", zap.Float64("availability", availability), zap.Int("workflow-cancelled", workflowCancelled))
				return availability
			}

			if err := common.RetryOp(func() error {
				totalCancelWorkflowCall++

				cancelCtx, cancel := context.WithTimeout(ctx, time.Duration(config.ContextTimeoutInSeconds)*time.Second)
				err := cc.CancelWorkflow(cancelCtx, workflowID, "")
				cancel()

				if err == nil || common.IsCancellationAlreadyRequestedError(err) || common.IsEntityNotExistsError(err) {
					succeededCancelWorkflowCall++
					return nil
				}
				if common.IsServiceBusyError(err) {
					// do not count service busy as failure
					succeededCancelWorkflowCall++
				}

				logger.Error("Failed to cancel workflow execution", zap.Error(err))
				return err
			}, nil); err == nil {
				workflowCancelled++
			}

			time.Sleep(time.Duration(waitDurationInMilliSeconds+rand.Intn(waitDurationJitterInMilliSeconds)) * time.Millisecond)
		}
	}
}

func validationActivity(
	ctx context.Context,
	config *lib.CancellationTestConfig,
	testStartTimeNanos int64,
) error {
	cc := ctx.Value(lib.CtxKeyCadenceClient).(lib.CadenceClient)

	domain := activity.GetInfo(ctx).WorkflowDomain

	query := fmt.Sprintf("WorkflowType = '%s' and StartTime > %v", sleepWorkflowName, testStartTimeNanos)
	request := &shared.CountWorkflowExecutionsRequest{
		Domain: &domain,
		Query:  &query,
	}

	// 1. check if enough workflows are started
	resp, err := cc.CountWorkflow(ctx, request)
	if err != nil {
		return err
	}
	totalLaunchCount := resp.GetCount()
	if totalLaunchCount < int64(config.TotalLaunchCount) {
		return cadence.NewCustomError(common.ErrReasonValidationFailed, fmt.Sprintf("Expected to start %v workflow, actual started: %v", config.TotalLaunchCount, totalLaunchCount))
	}

	// 2. check if all workflows are closed
	query = fmt.Sprintf("WorkflowType = '%s' and StartTime > %v and CloseTime != missing", sleepWorkflowName, testStartTimeNanos)
	request.Query = &query
	resp, err = cc.CountWorkflow(ctx, request)
	if err != nil {
		return err
	}
	if resp.GetCount() != totalLaunchCount {
		return cadence.NewCustomError(common.ErrReasonValidationFailed, fmt.Sprintf("Not all workflows are closed, started: %v, closed: %v", totalLaunchCount, resp.GetCount()))
	}

	// TODO: uncomment the following check after cancellation test is rewritten.
	// currently if the launcherActivity failed in the middle (e.g. heartbeat timeout),
	// all the workflowID in memory will be lost and those workflows can't be cancelled.

	// 3. check if enough workflows are cancelled
	// query = fmt.Sprintf("WorkflowType = '%s' and StartTime > %v and CloseStatus = 2", sleepWorkflowName, startTimeNanos)
	// request.Query = &query
	// resp, err = cc.CountWorkflow(ctx, request)
	// if err != nil {
	// 	return err
	// }
	// if resp.GetCount() < int64(float64(totalLaunchCount)*common.DefaultAvailabilityThreshold) {
	// 	return cadence.NewCustomError(common.ErrReasonValidationFailed, fmt.Sprintf("Cancelled workflow count too low, started: %v, cancelled: %v", totalLaunchCount, resp.GetCount()))
	// }

	return nil
}

func sleepWorkflow(
	ctx workflow.Context,
	sleepDuration time.Duration,
) error {
	return workflow.Sleep(ctx, sleepDuration)
}
