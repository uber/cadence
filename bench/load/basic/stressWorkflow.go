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
	"fmt"
	"time"

	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/uber/cadence/bench/load/common"
)

const (
	stressWorkflowName = "basicStressWorkflow"
)

type (
	// WorkflowParams inputs to workflow.
	WorkflowParams struct {
		ChainSequence    int
		ConcurrentCount  int
		TaskListNumber   int
		PayloadSizeBytes int
		CadenceSleep     time.Duration
		PanicWorkflow    bool
	}
)

// RegisterWorker registers workflows and activities for basic load
func RegisterWorker(w worker.Worker) {
	w.RegisterWorkflowWithOptions(stressWorkflowExecute, workflow.RegisterOptions{Name: stressWorkflowName})
}

func stressWorkflowExecute(ctx workflow.Context, workflowInput WorkflowParams) error {

	if workflowInput.PanicWorkflow {
		panic("panic workflow load test.")
	}

	activityParams := common.EchoActivityParams{
		Payload: make([]byte, workflowInput.PayloadSizeBytes),
	}

	ao := workflow.ActivityOptions{
		TaskList:               common.GetTaskListName(workflowInput.TaskListNumber),
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 0; i < workflowInput.ChainSequence; i++ {
		selector := workflow.NewSelector(ctx)
		var activityErr error
		for j := 0; j < workflowInput.ConcurrentCount; j++ {
			selector.AddFuture(workflow.ExecuteActivity(ctx, common.EchoActivityName, activityParams), func(f workflow.Future) {
				err := f.Get(ctx, nil)
				if err != nil {
					workflow.GetLogger(ctx).Error("basic test stress workflow echo activity execution failed", zap.Error(err))
					activityErr = err
				}
			})
		}

		for i := 0; i < workflowInput.ConcurrentCount; i++ {
			selector.Select(ctx) // this will wait for one branch
			if activityErr != nil {
				return fmt.Errorf("echo activity execution failed: %v", activityErr)
			}
		}

		if workflowInput.CadenceSleep > 0 {
			if err := workflow.Sleep(ctx, workflowInput.CadenceSleep); err != nil {
				workflow.GetLogger(ctx).Error("cadence.Sleep() returned error", zap.Error(err))
				return fmt.Errorf("stress workflow sleep failed: %v", err)
			}
		}
	}
	return nil
}
