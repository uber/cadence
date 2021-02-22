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

package basic

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
	c "github.com/uber/cadence/common"
)

const (
	// TestName is the test name for basic test
	TestName = "basic"

	// LauncherWorkflowName is the workflow name for launching basic load test
	LauncherWorkflowName = "basic-load-test-workflow"
	// LauncherLaunchActivityName is the activity name for launching stress load test
	LauncherLaunchActivityName = "basic-load-test-launch-activity"
	// LauncherVerifyActivityName is the verification activity name
	LauncherVerifyActivityName = "basic-load-test-verify-activity"

	stressWorkflowStartToCloseTimeout = 5 * time.Minute
	failedWorkflowQuery               = "WorkflowType='%v' and CloseStatus=1 and StartTime > %v and CloseTime < %v"
	workflowWaitTimeBuffer            = 5 * time.Minute
	maxLauncherActivityRetryCount     = 5

	errReasonValidationFailed = "validation failed"
)

var (
	maxPageSize = int32(10)
)

type (
	verifyActivityParams struct {
		FailedWorkflowCount int64
		WorkflowStartTime   int64
	}

	launcherActivityParams struct {
		RoutineID int
		Count     int
		StartID   int
		Config    lib.BasicTestConfig
	}

	activityResult struct {
		SuccessCount int64
		FailedCount  int64
	}
)

func (r *activityResult) IncSuccess() {
	r.SuccessCount++
}

func (r *activityResult) IncFailed() {
	r.FailedCount++
}

func (r *activityResult) String() string {
	return fmt.Sprintf("SuccessCount: %v, FailedCount: %v", r.SuccessCount, r.FailedCount)
}

// RegisterLauncher registers workflows and activities for basic load launching
func RegisterLauncher(w worker.Worker) {
	w.RegisterWorkflowWithOptions(launcherWorkflow, workflow.RegisterOptions{Name: LauncherWorkflowName})
	w.RegisterActivityWithOptions(launcherActivity, activity.RegisterOptions{Name: LauncherLaunchActivityName})
	w.RegisterActivityWithOptions(verifyResultActivity, activity.RegisterOptions{Name: LauncherVerifyActivityName})
}

func launcherWorkflow(ctx workflow.Context, config lib.BasicTestConfig) (string, error) {
	logger := workflow.GetLogger(ctx).With(zap.String("Test", TestName))
	workflowPerActivity := config.TotalLaunchCount / config.RoutineCount
	workflowTimeout := workflow.GetInfo(ctx).ExecutionStartToCloseTimeoutSeconds
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Duration(workflowTimeout) * time.Second,
		HeartbeatTimeout:       20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	doneCh := workflow.NewChannel(ctx)
	startTime := workflow.Now(ctx)

	var result activityResult
	for i := 0; i < config.RoutineCount; i++ {
		var batchStartID int
		workflow.Go(ctx, func(ctx workflow.Context) {
			for j := 0; j < maxLauncherActivityRetryCount; j++ {
				var runResult activityResult
				actInput := launcherActivityParams{
					RoutineID: i,
					Count:     workflowPerActivity,
					StartID:   batchStartID,
					Config:    config,
				}
				f := workflow.ExecuteActivity(
					ctx,
					LauncherLaunchActivityName,
					actInput,
				)
				err := f.Get(ctx, &runResult)
				if err == nil {
					result.SuccessCount += runResult.SuccessCount
					result.FailedCount += runResult.FailedCount
					break
				}

				logger.Error("basic test launcher activity execution error", zap.Error(err))
				switch err := err.(type) {
				case *workflow.TimeoutError:
					if err.TimeoutType() == shared.TimeoutTypeHeartbeat && err.HasDetails() {
						if err := err.Details(&batchStartID); err == nil {
							batchStartID++
						}
					}
				}
				result.SuccessCount += runResult.SuccessCount
				result.FailedCount += runResult.FailedCount
			}
			doneCh.Send(ctx, "done")
		})
	}
	for i := 0; i < config.RoutineCount; i++ {
		doneCh.Receive(ctx, nil)
	}

	workflowWaitTime := stressWorkflowStartToCloseTimeout
	if config.ExecutionStartToCloseTimeoutInSeconds > 0 {
		workflowWaitTime = time.Duration(config.ExecutionStartToCloseTimeoutInSeconds) * time.Second
	}
	if err := workflow.Sleep(ctx, workflowWaitTime+workflowWaitTimeBuffer); err != nil {
		return "", fmt.Errorf("launcher workflow sleep failed: %v", err)
	}

	actInput := verifyActivityParams{
		FailedWorkflowCount: int64(float64(config.TotalLaunchCount) * config.FailureThreshold),
		WorkflowStartTime:   startTime.UnixNano(),
	}
	actOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2,
			MaximumAttempts:          5,
			NonRetriableErrorReasons: []string{errReasonValidationFailed},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, actOptions)
	if err := workflow.ExecuteActivity(
		ctx,
		LauncherVerifyActivityName,
		actInput,
	).Get(ctx, nil); err != nil {
		return "", err
	}
	return result.String(), nil
}

func launcherActivity(ctx context.Context, params launcherActivityParams) (activityResult, error) {
	logger := activity.GetLogger(ctx).With(zap.String("Test", TestName))
	logger.Info("Start activity routineID", zap.Int("RoutineID", params.RoutineID))

	cc := ctx.Value(lib.CtxKeyCadenceClient).(lib.CadenceClient)
	rc := ctx.Value(lib.CtxKeyRuntimeContext).(*lib.RuntimeContext)
	numTaskList := rc.Bench.NumTaskLists
	config := params.Config
	input := WorkflowParams{
		ChainSequence:    config.ChainSequence,
		ConcurrentCount:  config.ConcurrentCount,
		PayloadSizeBytes: config.PayloadSizeBytes,
		PanicWorkflow:    config.PanicStressWorkflow,
	}
	workflowOptions := client.StartWorkflowOptions{
		ExecutionStartToCloseTimeout:    stressWorkflowStartToCloseTimeout,
		DecisionTaskStartToCloseTimeout: 10 * time.Second,
	}

	if config.ExecutionStartToCloseTimeoutInSeconds > 0 {
		// default 5m is not enough for long running workflow
		workflowOptions.ExecutionStartToCloseTimeout = time.Duration(config.ExecutionStartToCloseTimeoutInSeconds) * time.Second
	}
	if config.MinCadenceSleepInSeconds > 0 && config.MaxCadenceSleepInSeconds > 0 {
		minSleep := time.Duration(config.MinCadenceSleepInSeconds) * time.Second
		maxSleep := time.Duration(config.MaxCadenceSleepInSeconds) * time.Second
		jitter := rand.Float64() * float64(maxSleep-minSleep)
		sleepTime := minSleep + time.Duration(int64(jitter))
		input.CadenceSleep = sleepTime
	}

	res := activityResult{}
	if config.ContextTimeoutInSeconds == 0 {
		config.ContextTimeoutInSeconds = int(common.DefaultContextTimeout.Seconds())
	}
	for i := params.StartID; i < params.Count; i++ {
		input.TaskListNumber = rand.Intn(numTaskList)

		workflowOptions.ID = fmt.Sprintf("%s-%d-%d", uuid.New(), params.RoutineID, i)
		workflowOptions.TaskList = common.GetTaskListName(input.TaskListNumber)

		startWorkflowContext, cancelF := context.WithTimeout(context.Background(), time.Duration(config.ContextTimeoutInSeconds)*time.Second)
		we, err := cc.StartWorkflow(startWorkflowContext, workflowOptions, stressWorkflowExecute, input)
		cancelF()
		if err == nil {
			res.IncSuccess()
			logger.Debug("Created Workflow", zap.String("WorkflowID", we.ID), zap.String("RunID", we.RunID))
		} else {
			res.IncFailed()
			logger.Error("Failed to start workflow execution", zap.String("WorkflowID", we.ID), zap.Error(err))
		}
		activity.RecordHeartbeat(ctx, i)
		jitter := time.Duration(75 + rand.Intn(25))
		time.Sleep(jitter * time.Millisecond)
	}

	logger.Info("Result of running launcher activity", zap.String("Result", res.String()))
	return res, nil
}

func verifyResultActivity(
	ctx context.Context,
	params verifyActivityParams,
) error {

	cc := ctx.Value(lib.CtxKeyCadenceClient).(lib.CadenceClient)
	info := activity.GetInfo(ctx)

	// verify if any open workflow
	listWorkflowRequest := &shared.ListOpenWorkflowExecutionsRequest{
		Domain:          c.StringPtr(info.WorkflowDomain),
		MaximumPageSize: &maxPageSize,
		StartTimeFilter: &shared.StartTimeFilter{
			EarliestTime: c.Int64Ptr(params.WorkflowStartTime),
			LatestTime:   c.Int64Ptr(time.Now().UnixNano()),
		},
		TypeFilter: &shared.WorkflowTypeFilter{
			Name: c.StringPtr(stressWorkflowName),
		},
	}
	openWorkflow, err := cc.ListOpenWorkflow(ctx, listWorkflowRequest)
	if err != nil {
		return err
	}
	if len(openWorkflow.Executions) > 0 {
		return cadence.NewCustomError(
			errReasonValidationFailed,
			fmt.Sprintf("found open workflows after basic load test completed, workflowID: %v, runID: %v",
				openWorkflow.Executions[0].Execution.GetWorkflowId(),
				openWorkflow.Executions[0].Execution.GetRunId(),
			),
		)

	}

	// verify the number of failed workflows
	query := fmt.Sprintf(
		failedWorkflowQuery,
		stressWorkflowName,
		params.WorkflowStartTime,
		time.Now().UnixNano())
	request := &shared.CountWorkflowExecutionsRequest{
		Domain: c.StringPtr(info.WorkflowDomain),
		Query:  &query,
	}
	resp, err := cc.CountWorkflow(ctx, request)
	if err != nil {
		return err
	}
	if resp.GetCount() > params.FailedWorkflowCount {
		return cadence.NewCustomError(
			errReasonValidationFailed,
			fmt.Sprintf("found too many failed workflows(%v) after basic load test completed", params.FailedWorkflowCount),
		)
	}
	return nil
}
