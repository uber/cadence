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
