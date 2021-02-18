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
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/bench/load/common"
)

const (
	timerWorkflowName = "timerWorkflow"
)

type (
	// WorkflowParams inputs to workflow.
	WorkflowParams struct {
		TimerCount           int
		EarliesTimerFireTime time.Time
		LatestTimerFireTime  time.Time
		MaxTimerLatency      time.Duration
	}
)

// RegisterWorker registers workflows and activities for timer load
func RegisterWorker(w worker.Worker) {
	w.RegisterWorkflowWithOptions(timerWorkflow, workflow.RegisterOptions{Name: timerWorkflowName})
}

func timerWorkflow(ctx workflow.Context, workflowInput WorkflowParams) error {
	now := workflow.Now(ctx)
	shortestTimerDuration := workflowInput.EarliesTimerFireTime.Sub(now)
	timerFireWindowInSeconds := int64(workflowInput.LatestTimerFireTime.Sub(workflowInput.EarliesTimerFireTime).Seconds())

	selector := workflow.NewSelector(ctx)
	expectedFireTime := time.Unix(0, math.MaxInt64)
	for i := 0; i != workflowInput.TimerCount; i++ {
		timerDuration := shortestTimerDuration
		if timerFireWindowInSeconds > 0 {
			var randDurationNano int64
			workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
				return rand.Int63n(timerFireWindowInSeconds)
			}).Get(&randDurationNano)

			timerDuration += time.Duration(randDurationNano) * time.Second
		}

		if timerDuration < 0 {
			timerDuration = time.Second
		}

		fireTime := now.Add(timerDuration)
		if fireTime.Before(expectedFireTime) {
			expectedFireTime = fireTime
		}

		f := workflow.NewTimer(ctx, timerDuration)
		selector.AddFuture(f, func(_ workflow.Future) {})
	}
	selector.Select(ctx)

	timerLatency := workflow.Now(ctx).Sub(expectedFireTime)
	if timerLatency > workflowInput.MaxTimerLatency {
		return cadence.NewCustomError("timer latency too high", fmt.Sprintf("expectedLatency: %v, actual latency: %v", workflowInput.MaxTimerLatency, timerLatency))
	}

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Hour,
		StartToCloseTimeout:    time.Hour,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	return workflow.ExecuteActivity(ctx, common.EchoActivityName, common.EchoActivityParams{}).Get(ctx, nil)
}
