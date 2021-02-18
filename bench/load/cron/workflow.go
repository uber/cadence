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

package cron

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/uber/cadence/bench/lib"
	"github.com/uber/cadence/bench/load/basic"
	"github.com/uber/cadence/bench/load/cancellation"
	"github.com/uber/cadence/bench/load/common"
	"github.com/uber/cadence/bench/load/concurrentexec"
	"github.com/uber/cadence/bench/load/signal"
	"github.com/uber/cadence/bench/load/timer"
)

const (
	// TestName is the test name for cron bench test
	TestName = "cron"
)

const (
	cronWorkflowName         = "cron-test-workflow"
	cronLauncherWorkflowName = "cron-launcher-workflow"

	queryTypeTestResults = "test-results"

	testResultSignalName = "test-result"

	testPassedBoolSearchAttribute = "Passed"
)

type (
	workflowExecution struct {
		WorkflowID string
		RunID      string
	}

	testResult struct {
		Name        string
		Description string
		TestStatus  string
		Details     string
	}
)

const (
	testStatusPassed = "Passed"
	testStatusFailed = "Failed"
)

// RegisterLauncher registers workflows for cron load launching
func RegisterLauncher(w worker.Worker) {
	w.RegisterWorkflowWithOptions(cronWorkflow, workflow.RegisterOptions{Name: cronWorkflowName})
	w.RegisterWorkflowWithOptions(launcherWorkflow, workflow.RegisterOptions{Name: cronLauncherWorkflowName})
	w.RegisterActivity(launcherActivity)
	w.RegisterActivity(signalResultActivity)
}

func cronWorkflow(
	ctx workflow.Context,
	config lib.CronTestConfig,
) error {
	now := workflow.Now(ctx).UnixNano()
	profile, err := lib.BeginWorkflow(ctx, cronWorkflowName, now)
	if err != nil {
		return err
	}

	workflowTimeoutInSeconds := workflow.GetInfo(ctx).ExecutionStartToCloseTimeoutSeconds

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			ExpirationInterval: time.Duration(workflowTimeoutInSeconds) * time.Second, // retry the activity until workflow timeout
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	totalTests := 0
	futures := make([]workflow.Future, 0, len(config.TestSuites))
	for _, testSuite := range config.TestSuites {
		totalTests += len(testSuite.Configs)
		testSuiteName := testSuite.Name
		future := workflow.ExecuteActivity(ctx, launcherActivity, testSuite.Domain, testSuiteName, testSuite.Configs, workflowTimeoutInSeconds)
		futures = append(futures, future)
	}

	for _, future := range futures {
		if err := future.Get(ctx, nil); err != nil {
			return profile.End(err)
		}
	}

	testResults := make([]testResult, 0, totalTests)
	_ = workflow.SetQueryHandler(ctx, queryTypeTestResults, func() ([]testResult, error) {
		return testResults, nil
	})

	testPassed := true
	testResultCh := workflow.GetSignalChannel(ctx, testResultSignalName)
	for len(testResults) != totalTests {
		var result testResult
		testResultCh.Receive(ctx, &result)
		testResults = append(testResults, result)

		if result.TestStatus == testStatusFailed {
			testPassed = false
		}
	}

	err = workflow.UpsertSearchAttributes(ctx, map[string]interface{}{testPassedBoolSearchAttribute: testPassed})

	return profile.End(err)
}

func launcherActivity(
	ctx context.Context,
	domain string,
	testSuiteName string,
	testConfigs []lib.AggregatedTestConfig,
	timeoutInSeconds int32,
) error {
	logger := activity.GetLogger(ctx)

	rc := ctx.Value(lib.CtxKeyRuntimeContext).(*lib.RuntimeContext)
	cc, err := lib.NewCadenceClientForDomain(rc, domain)
	if err != nil {
		logger.Error("Failed to create cadence client", zap.String("domain", domain))
		return err
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:                              "cron-launcher-" + testSuiteName,
		TaskList:                        common.GetTaskListName(0), // default to use the tasklist with ID 0
		ExecutionStartToCloseTimeout:    time.Duration(timeoutInSeconds) * time.Second,
		DecisionTaskStartToCloseTimeout: time.Minute,
		WorkflowIDReusePolicy:           client.WorkflowIDReusePolicyAllowDuplicate,
	}

	activityInfo := activity.GetInfo(ctx)

	_ = common.RetryOp(func() error {
		_, err = cc.StartWorkflow(
			ctx,
			workflowOptions,
			cronLauncherWorkflowName,
			activityInfo.WorkflowDomain,
			workflowExecution{
				WorkflowID: activityInfo.WorkflowExecution.ID,
				RunID:      activityInfo.WorkflowExecution.RunID,
			},
			testSuiteName,
			testConfigs,
			time.Now().UnixNano(),
		)
		if err == nil || cadence.IsWorkflowExecutionAlreadyStartedError(err) {
			// TODO: it's possible that the launcher workflow is started by another cron
			logger.Info("Started test suite", zap.String("test-suite", testSuiteName))
			return nil
		}

		logger.Error("Failed to start test suite", zap.String("test-suite", testSuiteName))
		return err
	}, nil)

	return err
}

func launcherWorkflow(
	ctx workflow.Context,
	parentDomain string,
	parentExecution workflowExecution,
	testSuiteName string,
	testConfigs []lib.AggregatedTestConfig,
	scheduledTimeNanos int64,
) ([]testResult, error) {
	profile, err := lib.BeginWorkflow(ctx, cronLauncherWorkflowName, scheduledTimeNanos)
	if err != nil {
		return nil, err
	}

	testTimeout := workflow.GetInfo(ctx).ExecutionStartToCloseTimeoutSeconds
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2,
			ExpirationInterval:       time.Duration(testTimeout) * time.Second, // retry the signal activity until workflow timeout
			NonRetriableErrorReasons: []string{common.ErrReasonWorkflowNotExist},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	numTests := len(testConfigs)
	testResults := make([]testResult, 0, numTests)

	// run all tests in random order
	var testOrder []int
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return rand.Perm(len(testConfigs))
	}).Get(&testOrder)
	if err != nil {
		for idx := range testConfigs {
			result := testResult{
				Name:       testConfigs[idx].Name,
				TestStatus: testStatusFailed,
				Details:    fmt.Sprintf("Failed to generate randomized test order: %v", err.Error()),
			}
			testResults = append(testResults, result)
			if err := workflow.ExecuteActivity(ctx, signalResultActivity, parentDomain, parentExecution, result).Get(ctx, nil); err != nil {
				workflow.GetLogger(ctx).Error("Failed to signal test result", zap.Error(err))
				return nil, profile.End(err)
			}
		}
		return testResults, nil
	}

	for _, idx := range testOrder {
		testConfig := testConfigs[idx]
		cwo := workflow.ChildWorkflowOptions{
			WorkflowID:                   testSuiteName + "-" + testConfig.Name + "-" + strconv.Itoa(idx),
			TaskList:                     common.GetTaskListName(0), // default to use the tasklist with ID 0
			ExecutionStartToCloseTimeout: time.Duration(testConfig.TimeoutInSeconds) * time.Second,
			TaskStartToCloseTimeout:      time.Minute,
			ParentClosePolicy:            client.ParentClosePolicyTerminate,
			WorkflowIDReusePolicy:        client.WorkflowIDReusePolicyAllowDuplicate,
		}
		childCtx := workflow.WithChildOptions(ctx, cwo)
		var childFuture workflow.Future
		switch testConfig.Name {
		case basic.TestName:
			childFuture = workflow.ExecuteChildWorkflow(childCtx, basic.LauncherWorkflowName, *testConfig.Basic)
		case signal.TestName:
			childFuture = workflow.ExecuteChildWorkflow(childCtx, signal.LauncherWorkflowName, *testConfig.Signal)
		case timer.TestName:
			childFuture = workflow.ExecuteChildWorkflow(childCtx, timer.LauncherWorkflowName, *testConfig.Timer)
		case concurrentexec.TestName:
			childFuture = workflow.ExecuteChildWorkflow(childCtx, concurrentexec.LauncherWorkflowName, *testConfig.ConcurrentExec)
		case cancellation.TestName:
			childFuture = workflow.ExecuteChildWorkflow(childCtx, cancellation.LauncherWorkflowName, *testConfig.Cancellation)
		default:
			workflow.GetLogger(ctx).Error("Unknown test name", zap.String("test-name", testConfig.Name))
		}

		result := testResult{
			Name:        testSuiteName + "::" + testConfig.Name,
			Description: testConfig.Description,
		}
		if childFuture == nil {
			result.TestStatus = testStatusFailed
			result.Details = "Unknown test"
		} else if err := childFuture.Get(childCtx, nil); err != nil {
			result.TestStatus = testStatusFailed
			result.Details = err.Error()
			if customErr, ok := err.(*cadence.CustomError); ok {
				var detailStr string
				if err := customErr.Details(&detailStr); err == nil {
					result.Details += ": " + detailStr
				}
			}
		} else {
			result.TestStatus = testStatusPassed
		}
		testResults = append(testResults, result)

		if err := workflow.ExecuteActivity(ctx, signalResultActivity, parentDomain, parentExecution, result).Get(ctx, nil); err != nil {
			workflow.GetLogger(ctx).Error("Failed to signal test result", zap.String("test-name", testConfig.Name), zap.Error(err))
			return nil, profile.End(err)
		}
	}

	return testResults, profile.End(nil)
}

func signalResultActivity(
	ctx context.Context,
	targetDomain string,
	targetWorkflowExecution workflowExecution,
	result testResult,
) error {
	logger := activity.GetLogger(ctx)

	rc := ctx.Value(lib.CtxKeyRuntimeContext).(*lib.RuntimeContext)
	cc, err := lib.NewCadenceClientForDomain(rc, targetDomain)
	if err != nil {
		logger.Error("Failed to create cadence client", zap.String("domain", targetDomain))
		return err
	}

	if err = common.RetryOp(func() error {
		return cc.SignalWorkflow(
			ctx,
			targetWorkflowExecution.WorkflowID,
			targetWorkflowExecution.RunID,
			testResultSignalName,
			result,
		)
	}, common.IsNonRetryableError); err != nil {
		logger.Error("Failed to signal test result back to workflow",
			zap.String("domain", targetDomain),
			zap.String("workflowID", targetWorkflowExecution.WorkflowID),
			zap.String("workflowID", targetWorkflowExecution.RunID),
			zap.Error(err),
		)
		return cadence.NewCustomError(common.ErrReasonWorkflowNotExist, err.Error())
	}

	return nil
}
