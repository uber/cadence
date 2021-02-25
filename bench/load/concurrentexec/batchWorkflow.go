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

package concurrentexec

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/bench/lib"
	"github.com/uber/cadence/bench/load/common"
)

const (
	batchWorkflowName = "concurrent-execution-batch-workflow"
)

const (
	batchTypeActivity      = "activity"
	batchTypeChildWorkflow = "childworkflow"

	maxSleepTimeInSeconds = 10
)

// RegisterWorker registers workflows and activities for concurrent execution test
func RegisterWorker(w worker.Worker) {
	w.RegisterWorkflowWithOptions(batchWorkflow, workflow.RegisterOptions{Name: batchWorkflowName})
	w.RegisterWorkflow(concurrentChildWorkflow)
	w.RegisterActivity(concurrentActivity)
}

func batchWorkflow(
	ctx workflow.Context,
	config lib.ConcurrentExecTestConfig,
	scheduledTimeNanos int64,
) (float64, error) {
	profile, err := lib.BeginWorkflow(ctx, batchWorkflowName, scheduledTimeNanos)
	if err != nil {
		return 0, err
	}

	now := workflow.Now(ctx).UnixNano()

	batchType := strings.ToLower(strings.TrimSpace(config.BatchType))

	var numTaskList int
	if numTaskList, err = getTaskListNum(ctx); err != nil {
		return 0, profile.End(fmt.Errorf("failed to getTaskListNum, error: %v", err))
	}

	futures := make([]workflow.Future, 0, config.BatchSize)
	for i := 0; i != config.BatchSize; i++ {
		switch batchType {
		case batchTypeActivity:
			futures = append(futures, scheduleActivity(ctx, numTaskList, config.BatchTimeoutInSeconds, now))
		case batchTypeChildWorkflow:
			futures = append(futures, scheduleChildWorkflow(ctx, numTaskList, config.BatchTimeoutInSeconds, now))
		default:
			return 0, profile.End(fmt.Errorf("unknown batch type: %v", batchType))
		}
	}

	scheduledInTime := 0
	maxScheduleLatency := time.Duration(config.BatchMaxLatencyInSeconds) * time.Second
	for _, future := range futures {
		var scheduleLatency time.Duration
		if err := future.Get(ctx, &scheduleLatency); err != nil {
			return 0, profile.End(fmt.Errorf("childworkflow/activity failed, error: %v", err))
		}
		if scheduleLatency < maxScheduleLatency {
			scheduledInTime++
		}
	}

	return float64(scheduledInTime) / float64(config.BatchSize), profile.End(nil)
}

func scheduleActivity(
	ctx workflow.Context,
	numTaskList int,
	timeoutInSeconds int,
	scheduledTimeNanos int64,
) workflow.Future {
	ao := workflow.ActivityOptions{
		TaskList:               common.GetTaskListName(rand.Intn(numTaskList)),
		ScheduleToStartTimeout: time.Duration(timeoutInSeconds) * time.Second,
		StartToCloseTimeout:    2 * maxSleepTimeInSeconds * time.Second,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			ExpirationInterval: time.Duration(timeoutInSeconds) * time.Second,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)
	return workflow.ExecuteActivity(activityCtx, concurrentActivity, scheduledTimeNanos, maxSleepTimeInSeconds)
}

func scheduleChildWorkflow(
	ctx workflow.Context,
	numTaskList int,
	timeoutInSeconds int,
	scheduledTimeNanos int64,
) workflow.Future {
	cwo := workflow.ChildWorkflowOptions{
		TaskList:                     common.GetTaskListName(rand.Intn(numTaskList)),
		ExecutionStartToCloseTimeout: time.Duration(timeoutInSeconds) * time.Second,
		TaskStartToCloseTimeout:      time.Second * 10,
		ParentClosePolicy:            client.ParentClosePolicyTerminate,
	}
	childCtx := workflow.WithChildOptions(ctx, cwo)
	return workflow.ExecuteChildWorkflow(childCtx, concurrentChildWorkflow, scheduledTimeNanos, maxSleepTimeInSeconds)
}

func concurrentActivity(
	ctx context.Context,
	scheduledTimeNanos int64,
	maxSleepTimeInSeconds int,
) (time.Duration, error) {
	var latency time.Duration
	if activity.GetInfo(ctx).Attempt == 0 {
		latency = time.Now().Sub(time.Unix(0, scheduledTimeNanos))
	}

	time.Sleep(time.Duration(rand.Intn(maxSleepTimeInSeconds)) * time.Second)
	return latency, nil
}

func concurrentChildWorkflow(
	ctx workflow.Context,
	scheduledTimeNanos int64,
	maxSleepTimeInSeconds int,
) (time.Duration, error) {
	latency := workflow.Now(ctx).Sub(time.Unix(0, scheduledTimeNanos))

	var sleepTimeInSeconds int
	workflow.SideEffect(ctx, func(_ workflow.Context) interface{} {
		return rand.Intn(maxSleepTimeInSeconds)
	}).Get(&sleepTimeInSeconds)

	if err := workflow.Sleep(ctx, time.Duration(sleepTimeInSeconds)*time.Second); err != nil {
		return 0, err
	}
	return latency, nil
}

func getTaskListNum(ctx workflow.Context) (int, error) {
	var numTaskList int
	lao := workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
	}
	laCtx := workflow.WithLocalActivityOptions(ctx, lao)
	if err := workflow.ExecuteLocalActivity(laCtx, func(ctx context.Context) (int, error) {
		return ctx.Value(lib.CtxKeyRuntimeContext).(*lib.RuntimeContext).Bench.NumTaskLists, nil
	}).Get(laCtx, &numTaskList); err != nil {
		return 0, err
	}
	return numTaskList, nil
}
