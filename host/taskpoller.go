// Copyright (c) 2017 Uber Technologies, Inc.
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

package host

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/client"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/matching"
)

type (
	decisionTaskHandler func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error)
	activityTaskHandler func(execution *types.WorkflowExecution, activityType *types.ActivityType,
		ActivityID string, input []byte, takeToken []byte) ([]byte, bool, error)

	queryHandler func(task *types.PollForDecisionTaskResponse) ([]byte, error)

	// TaskPoller is used in integration tests to poll decision or activity tasks
	TaskPoller struct {
		Engine                              FrontendClient
		Domain                              string
		TaskList                            *types.TaskList
		StickyTaskList                      *types.TaskList
		StickyScheduleToStartTimeoutSeconds *int32
		Identity                            string
		DecisionHandler                     decisionTaskHandler
		ActivityHandler                     activityTaskHandler
		QueryHandler                        queryHandler
		Logger                              log.Logger
		T                                   *testing.T
	}
)

// PollAndProcessDecisionTask for decision tasks
func (p *TaskPoller) PollAndProcessDecisionTask(dumpHistory bool, dropTask bool) (isQueryTask bool, err error) {
	return p.PollAndProcessDecisionTaskWithAttempt(dumpHistory, dropTask, false, false, int64(0))
}

// PollAndProcessDecisionTaskWithSticky for decision tasks
func (p *TaskPoller) PollAndProcessDecisionTaskWithSticky(dumpHistory bool, dropTask bool) (isQueryTask bool, err error) {
	return p.PollAndProcessDecisionTaskWithAttempt(dumpHistory, dropTask, true, true, int64(0))
}

// PollAndProcessDecisionTaskWithoutRetry for decision tasks
func (p *TaskPoller) PollAndProcessDecisionTaskWithoutRetry(dumpHistory bool, dropTask bool) (isQueryTask bool, err error) {
	return p.PollAndProcessDecisionTaskWithAttemptAndRetry(dumpHistory, dropTask, false, false, int64(0), 1)
}

// PollAndProcessDecisionTaskWithAttempt for decision tasks
func (p *TaskPoller) PollAndProcessDecisionTaskWithAttempt(
	dumpHistory bool,
	dropTask bool,
	pollStickyTaskList bool,
	respondStickyTaskList bool,
	decisionAttempt int64,
) (isQueryTask bool, err error) {

	return p.PollAndProcessDecisionTaskWithAttemptAndRetry(
		dumpHistory,
		dropTask,
		pollStickyTaskList,
		respondStickyTaskList,
		decisionAttempt,
		5)
}

// PollAndProcessDecisionTaskWithAttemptAndRetry for decision tasks
func (p *TaskPoller) PollAndProcessDecisionTaskWithAttemptAndRetry(
	dumpHistory bool,
	dropTask bool,
	pollStickyTaskList bool,
	respondStickyTaskList bool,
	decisionAttempt int64,
	retryCount int,
) (isQueryTask bool, err error) {

	isQueryTask, _, err = p.PollAndProcessDecisionTaskWithAttemptAndRetryAndForceNewDecision(
		dumpHistory,
		dropTask,
		pollStickyTaskList,
		respondStickyTaskList,
		decisionAttempt,
		retryCount,
		false,
		nil)
	return isQueryTask, err
}

// PollAndProcessDecisionTaskWithAttemptAndRetryAndForceNewDecision for decision tasks
func (p *TaskPoller) PollAndProcessDecisionTaskWithAttemptAndRetryAndForceNewDecision(
	dumpHistory bool,
	dropTask bool,
	pollStickyTaskList bool,
	respondStickyTaskList bool,
	decisionAttempt int64,
	retryCount int,
	forceCreateNewDecision bool,
	queryResult *types.WorkflowQueryResult,
) (isQueryTask bool, newTask *types.RespondDecisionTaskCompletedResponse, err error) {
Loop:
	for attempt := 0; attempt < retryCount; attempt++ {

		taskList := p.TaskList
		if pollStickyTaskList {
			taskList = p.StickyTaskList
		}
		response, err1 := p.Engine.PollForDecisionTask(createContext(), &types.PollForDecisionTaskRequest{
			Domain:   p.Domain,
			TaskList: taskList,
			Identity: p.Identity,
		})

		if err1 != nil {
			return false, nil, err1
		}

		if response == nil || len(response.TaskToken) == 0 {
			p.Logger.Info("Empty Decision task: Polling again.")
			continue Loop
		}

		if response.GetNextEventID() == 0 {
			p.Logger.Fatal("NextEventID is not set for decision or query task")
		}

		var events []*types.HistoryEvent
		if response.Query == nil || !pollStickyTaskList {
			// if not query task, should have some history events
			// for non sticky query, there should be events returned
			history := response.History
			if history == nil {
				p.Logger.Fatal("History is nil")
			}

			events = history.Events
			if events == nil || len(events) == 0 {
				p.Logger.Fatal("History Events are empty")
			}

			nextPageToken := response.NextPageToken
			for nextPageToken != nil {
				resp, err2 := p.Engine.GetWorkflowExecutionHistory(createContext(), &types.GetWorkflowExecutionHistoryRequest{
					Domain:        p.Domain,
					Execution:     response.WorkflowExecution,
					NextPageToken: nextPageToken,
				})

				if err2 != nil {
					return false, nil, err2
				}

				events = append(events, resp.History.Events...)
				nextPageToken = resp.NextPageToken
			}
		} else {
			// for sticky query, there should be NO events returned
			// since worker side already has the state machine and we do not intend to update that.
			history := response.History
			nextPageToken := response.NextPageToken
			if !(history == nil || (len(history.Events) == 0 && nextPageToken == nil)) {
				// if history is not nil, and contains events or next token
				p.Logger.Fatal("History is not empty for sticky query")
			}
		}

		if dropTask {
			p.Logger.Info("Dropping Decision task: ")
			return false, nil, nil
		}

		if dumpHistory {
			common.PrettyPrintHistory(response.History, p.Logger)
		}

		// handle query task response
		if response.Query != nil {
			blob, err := p.QueryHandler(response)

			completeRequest := &types.RespondQueryTaskCompletedRequest{TaskToken: response.TaskToken}
			if err != nil {
				completeType := types.QueryTaskCompletedTypeFailed
				completeRequest.CompletedType = &completeType
				completeRequest.ErrorMessage = err.Error()
			} else {
				completeType := types.QueryTaskCompletedTypeCompleted
				completeRequest.CompletedType = &completeType
				completeRequest.QueryResult = blob
			}

			return true, nil, p.Engine.RespondQueryTaskCompleted(createContext(), completeRequest)
		}

		// handle normal decision task / non query task response
		var lastDecisionScheduleEvent *types.HistoryEvent
		for _, e := range events {
			if e.GetEventType() == types.EventTypeDecisionTaskScheduled {
				lastDecisionScheduleEvent = e
			}
		}
		if lastDecisionScheduleEvent != nil && decisionAttempt > 0 {
			require.Equal(p.T, decisionAttempt, lastDecisionScheduleEvent.DecisionTaskScheduledEventAttributes.GetAttempt())
		}

		executionCtx, decisions, err := p.DecisionHandler(response.WorkflowExecution, response.WorkflowType,
			common.Int64Default(response.PreviousStartedEventID), response.StartedEventID, response.History)
		if err != nil {
			p.Logger.Info("Failing Decision. Decision handler failed with error", tag.Error(err))
			return isQueryTask, nil, p.Engine.RespondDecisionTaskFailed(createContext(), &types.RespondDecisionTaskFailedRequest{
				TaskToken: response.TaskToken,
				Cause:     types.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure.Ptr(),
				Details:   []byte(err.Error()),
				Identity:  p.Identity,
			})
		}

		p.Logger.Info("Completing Decision.  Decisions", tag.Value(decisions))
		if !respondStickyTaskList {
			// non sticky tasklist
			newTask, err := p.Engine.RespondDecisionTaskCompleted(createContext(), &types.RespondDecisionTaskCompletedRequest{
				TaskToken:                  response.TaskToken,
				Identity:                   p.Identity,
				ExecutionContext:           executionCtx,
				Decisions:                  decisions,
				ReturnNewDecisionTask:      forceCreateNewDecision,
				ForceCreateNewDecisionTask: forceCreateNewDecision,
				QueryResults:               getQueryResults(response.GetQueries(), queryResult),
			})
			return false, newTask, err
		}
		// sticky tasklist
		newTask, err := p.Engine.RespondDecisionTaskCompleted(
			createContext(),
			&types.RespondDecisionTaskCompletedRequest{
				TaskToken:        response.TaskToken,
				Identity:         p.Identity,
				ExecutionContext: executionCtx,
				Decisions:        decisions,
				StickyAttributes: &types.StickyExecutionAttributes{
					WorkerTaskList:                p.StickyTaskList,
					ScheduleToStartTimeoutSeconds: p.StickyScheduleToStartTimeoutSeconds,
				},
				ReturnNewDecisionTask:      forceCreateNewDecision,
				ForceCreateNewDecisionTask: forceCreateNewDecision,
				QueryResults:               getQueryResults(response.GetQueries(), queryResult),
			},
			yarpc.WithHeader(common.LibraryVersionHeaderName, "0.0.1"),
			yarpc.WithHeader(common.FeatureVersionHeaderName, client.GoWorkerConsistentQueryVersion),
			yarpc.WithHeader(common.ClientImplHeaderName, client.GoSDK),
		)

		return false, newTask, err
	}

	return false, nil, matching.ErrNoTasks
}

// HandlePartialDecision for decision task
func (p *TaskPoller) HandlePartialDecision(response *types.PollForDecisionTaskResponse) (
	*types.RespondDecisionTaskCompletedResponse, error) {
	if response == nil || len(response.TaskToken) == 0 {
		p.Logger.Info("Empty Decision task: Polling again.")
		return nil, nil
	}

	var events []*types.HistoryEvent
	history := response.History
	if history == nil {
		p.Logger.Fatal("History is nil")
	}

	events = history.Events
	if events == nil || len(events) == 0 {
		p.Logger.Fatal("History Events are empty")
	}

	executionCtx, decisions, err := p.DecisionHandler(response.WorkflowExecution, response.WorkflowType,
		common.Int64Default(response.PreviousStartedEventID), response.StartedEventID, response.History)
	if err != nil {
		p.Logger.Info("Failing Decision. Decision handler failed with error: %v", tag.Error(err))
		return nil, p.Engine.RespondDecisionTaskFailed(createContext(), &types.RespondDecisionTaskFailedRequest{
			TaskToken: response.TaskToken,
			Cause:     types.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure.Ptr(),
			Details:   []byte(err.Error()),
			Identity:  p.Identity,
		})
	}

	p.Logger.Info("Completing Decision.  Decisions: %v", tag.Value(decisions))

	// sticky tasklist
	newTask, err := p.Engine.RespondDecisionTaskCompleted(
		createContext(),
		&types.RespondDecisionTaskCompletedRequest{
			TaskToken:        response.TaskToken,
			Identity:         p.Identity,
			ExecutionContext: executionCtx,
			Decisions:        decisions,
			StickyAttributes: &types.StickyExecutionAttributes{
				WorkerTaskList:                p.StickyTaskList,
				ScheduleToStartTimeoutSeconds: p.StickyScheduleToStartTimeoutSeconds,
			},
			ReturnNewDecisionTask:      true,
			ForceCreateNewDecisionTask: true,
		},
		yarpc.WithHeader(common.LibraryVersionHeaderName, "0.0.1"),
		yarpc.WithHeader(common.FeatureVersionHeaderName, client.GoWorkerConsistentQueryVersion),
		yarpc.WithHeader(common.ClientImplHeaderName, client.GoSDK),
	)

	return newTask, err
}

// PollAndProcessActivityTask for activity tasks
func (p *TaskPoller) PollAndProcessActivityTask(dropTask bool) error {
	for attempt := 0; attempt < 5; attempt++ {
		response, err1 := p.Engine.PollForActivityTask(createContext(), &types.PollForActivityTaskRequest{
			Domain:   p.Domain,
			TaskList: p.TaskList,
			Identity: p.Identity,
		})

		if err1 != nil {
			return err1
		}

		if response == nil || len(response.TaskToken) == 0 {
			p.Logger.Info("Empty Activity task: Polling again.")
			return nil
		}

		if dropTask {
			p.Logger.Info("Dropping Activity task: ")
			return nil
		}
		p.Logger.Debug("Received Activity task", tag.Value(response))

		result, cancel, err2 := p.ActivityHandler(response.WorkflowExecution, response.ActivityType, response.ActivityID,
			response.Input, response.TaskToken)
		if cancel {
			p.Logger.Info("Executing RespondActivityTaskCanceled")
			return p.Engine.RespondActivityTaskCanceled(createContext(), &types.RespondActivityTaskCanceledRequest{
				TaskToken: response.TaskToken,
				Details:   []byte("details"),
				Identity:  p.Identity,
			})
		}

		if err2 != nil {
			return p.Engine.RespondActivityTaskFailed(createContext(), &types.RespondActivityTaskFailedRequest{
				TaskToken: response.TaskToken,
				Reason:    common.StringPtr(err2.Error()),
				Details:   []byte(err2.Error()),
				Identity:  p.Identity,
			})
		}

		return p.Engine.RespondActivityTaskCompleted(createContext(), &types.RespondActivityTaskCompletedRequest{
			TaskToken: response.TaskToken,
			Identity:  p.Identity,
			Result:    result,
		})
	}

	return matching.ErrNoTasks
}

// PollAndProcessActivityTaskWithID is similar to PollAndProcessActivityTask but using RespondActivityTask...ByID
func (p *TaskPoller) PollAndProcessActivityTaskWithID(dropTask bool) error {
	for attempt := 0; attempt < 5; attempt++ {
		response, err1 := p.Engine.PollForActivityTask(createContext(), &types.PollForActivityTaskRequest{
			Domain:   p.Domain,
			TaskList: p.TaskList,
			Identity: p.Identity,
		})

		if err1 != nil {
			return err1
		}

		if response == nil || len(response.TaskToken) == 0 {
			p.Logger.Info("Empty Activity task: Polling again.")
			return nil
		}

		if response.GetActivityID() == "" {
			p.Logger.Info("Empty ActivityID")
			return nil
		}

		if dropTask {
			p.Logger.Info("Dropping Activity task: ")
			return nil
		}
		p.Logger.Debug("Received Activity task: %v", tag.Value(response))

		result, cancel, err2 := p.ActivityHandler(response.WorkflowExecution, response.ActivityType, response.ActivityID,
			response.Input, response.TaskToken)
		if cancel {
			p.Logger.Info("Executing RespondActivityTaskCanceled")
			return p.Engine.RespondActivityTaskCanceledByID(createContext(), &types.RespondActivityTaskCanceledByIDRequest{
				Domain:     p.Domain,
				WorkflowID: response.WorkflowExecution.GetWorkflowID(),
				RunID:      response.WorkflowExecution.GetRunID(),
				ActivityID: response.GetActivityID(),
				Details:    []byte("details"),
				Identity:   p.Identity,
			})
		}

		if err2 != nil {
			return p.Engine.RespondActivityTaskFailedByID(createContext(), &types.RespondActivityTaskFailedByIDRequest{
				Domain:     p.Domain,
				WorkflowID: response.WorkflowExecution.GetWorkflowID(),
				RunID:      response.WorkflowExecution.GetRunID(),
				ActivityID: response.GetActivityID(),
				Reason:     common.StringPtr(err2.Error()),
				Details:    []byte(err2.Error()),
				Identity:   p.Identity,
			})
		}

		return p.Engine.RespondActivityTaskCompletedByID(createContext(), &types.RespondActivityTaskCompletedByIDRequest{
			Domain:     p.Domain,
			WorkflowID: response.WorkflowExecution.GetWorkflowID(),
			RunID:      response.WorkflowExecution.GetRunID(),
			ActivityID: response.GetActivityID(),
			Identity:   p.Identity,
			Result:     result,
		})
	}

	return matching.ErrNoTasks
}

func createContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 90*time.Second)
	return ctx
}

func getQueryResults(queries map[string]*types.WorkflowQuery, queryResult *types.WorkflowQueryResult) map[string]*types.WorkflowQueryResult {
	result := make(map[string]*types.WorkflowQueryResult)
	for k := range queries {
		result[k] = queryResult
	}
	return result
}
