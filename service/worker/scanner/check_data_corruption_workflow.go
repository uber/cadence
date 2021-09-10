// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package scanner

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	c "github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation"
	"github.com/uber/cadence/common/reconciliation/entity"
	"github.com/uber/cadence/common/reconciliation/fetcher"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/resource"
)

func init() {
	workflow.RegisterWithOptions(
		CheckDataCorruptionWorkflow,
		workflow.RegisterOptions{Name: reconciliation.CheckDataCorruptionWorkflowType},
	)
	activity.Register(ExecutionFixerActivity)
}

const (
	workflowTimer   = 5 * time.Minute
	maxSignalNumber = 1000
)

// CheckDataCorruptionWorkflow is invoked by remote cluster via signals
func CheckDataCorruptionWorkflow(ctx workflow.Context) error {

	logger := workflow.GetLogger(ctx)
	signalCh := workflow.GetSignalChannel(ctx, reconciliation.CheckDataCorruptionWorkflowSignalName)
	signalCount := 0

	for {
		selector := workflow.NewSelector(ctx)
		// timer
		timerCtx, timerCancel := workflow.WithCancel(ctx)
		waitTimer := workflow.NewTimer(timerCtx, workflowTimer)
		timerFire := false
		selector.AddFuture(waitTimer, func(f workflow.Future) {
			if f.Get(timerCtx, nil) != workflow.ErrCanceled {
				timerFire = true
			}
		})

		// signal
		var fixList []entity.Execution
		selector.AddReceive(signalCh, func(c workflow.Channel, more bool) {
			var fixExecution entity.Execution
			for c.Receive(ctx, &fixExecution) {
				signalCount++
				fixList = append(fixList, fixExecution)
			}
		})

		selector.Select(ctx)
		if timerFire {
			return nil
		}
		timerCancel()

		activityOptions = workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    60 * time.Minute,
			HeartbeatTimeout:       5 * time.Minute,
			RetryPolicy:            &activityRetryPolicy,
		}
		activityCtx := workflow.WithActivityOptions(ctx, activityOptions)
		err := workflow.ExecuteActivity(activityCtx, ExecutionFixerActivity, fixList).Get(activityCtx, nil)
		if err != nil {
			logger.Error("failed to run execution fixer activity", zap.Error(err))
			return err
		}
		if signalCount > maxSignalNumber {
			return workflow.NewContinueAsNewError(ctx, reconciliation.CheckDataCorruptionWorkflowType)
		}
	}
}

func ExecutionFixerActivity(ctx context.Context, fixList []entity.Execution) ([]invariant.FixResult, error) {
	var result []invariant.FixResult
	index := 0
	if activity.HasHeartbeatDetails(ctx) {
		activity.GetHeartbeatDetails(ctx, &index)
	}

	for index < len(fixList) {
		execution := fixList[index]
		pr, err := getDefaultDAO(ctx, execution.ShardID)
		if err != nil {
			return nil, err
		}
		request := fetcher.ExecutionRequest{
			DomainID:   execution.DomainID,
			WorkflowID: execution.WorkflowID,
			RunID:      execution.RunID,
		}
		concreteExecution, err := fetcher.ConcreteExecution(ctx, pr, request)
		if err != nil {
			return nil, err
		}

		concreteExecutionInvariant := invariant.NewConcreteExecutionExists(pr)
		fixResult := concreteExecutionInvariant.Fix(ctx, concreteExecution)
		result = append(result, fixResult)
		historyInvariant := invariant.NewHistoryExists(pr)
		fixResult = historyInvariant.Fix(ctx, concreteExecution)
		result = append(result, fixResult)
		activity.RecordHeartbeat(ctx, index)
		index++
	}
	return result, nil
}

func getDefaultDAO(
	ctx context.Context,
	shardID int,
) (persistence.Retryer, error) {
	res, ok := ctx.Value(reconciliation.CheckDataCorruptionWorkflowType).(resource.Resource)
	if !ok {
		return nil, fmt.Errorf("cannot find key %v in context", reconciliation.CheckDataCorruptionWorkflowType)
	}

	execManager, err := res.GetExecutionManager(shardID)
	if err != nil {
		return nil, err
	}
	pr := persistence.NewPersistenceRetryer(execManager, res.GetHistoryManager(), c.CreatePersistenceRetryPolicy())
	return pr, nil
}
