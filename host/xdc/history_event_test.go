// Copyright (c) 2018 Uber Technologies, Inc.
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

package xdc

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/.gen/go/shared"
	"testing"
)

var (
	notPendingDecisionTask = func(pastEvents []Vertex) bool {
		count := 0
		for _, e := range pastEvents {
			switch {
			case e.GetName() == shared.EventTypeDecisionTaskScheduled.String():
				count++
			case e.GetName() == shared.EventTypeDecisionTaskCompleted.String(),
				e.GetName() == shared.EventTypeDecisionTaskFailed.String(),
				e.GetName() == shared.EventTypeDecisionTaskTimedOut.String():
				count--
			}
		}
		return count <= 0
	}

	containActivityComplete = func(pastEvents []Vertex) bool {
		for _, e := range pastEvents {
			if e.GetName() == shared.EventTypeActivityTaskCompleted.String() {
				return true
			}
		}
		return false
	}
)

type (
	historyEventTestSuit struct {
		suite.Suite
		generator Generator
	}
)

func TestHistoryEventTestSuite(t *testing.T) {
	suite.Run(t, new(historyEventTestSuit))
}

func (s *historyEventTestSuit) SetupSuite() {
	dtModel := NewHistoryEventModel()
	dtSchedule := NewHistoryEvent(shared.EventTypeDecisionTaskScheduled.String())
	dtStart := NewHistoryEvent(shared.EventTypeDecisionTaskStarted.String())
	dtStart.SetIsStrictOnNextVertex(true)
	dtFail := NewHistoryEvent(shared.EventTypeDecisionTaskFailed.String())
	dtTimedOut := NewHistoryEvent(shared.EventTypeDecisionTaskTimedOut.String())
	dtComplete := NewHistoryEvent(shared.EventTypeDecisionTaskCompleted.String())
	dtComplete.SetIsStrictOnNextVertex(true)
	dtComplete.SetMaxNextVertex(2)

	dtScheduleToStart := NewConnection(dtSchedule, dtStart)
	dtStartToComplete := NewConnection(dtStart, dtComplete)
	dtStartToFail := NewConnection(dtStart, dtFail)
	dtStartToTimedOut := NewConnection(dtStart, dtTimedOut)
	dtFailToSchedule := NewConnection(dtFail, dtSchedule)
	dtFailToSchedule.SetCondition(notPendingDecisionTask)
	dtTimedOutToSchedule := NewConnection(dtTimedOut, dtSchedule)
	dtTimedOutToSchedule.SetCondition(notPendingDecisionTask)
	dtModel.AddEdge(dtScheduleToStart, dtStartToComplete, dtStartToFail, dtStartToTimedOut, dtFailToSchedule, dtTimedOutToSchedule)

	wfModel := NewHistoryEventModel()
	wfStart := NewHistoryEvent(shared.EventTypeWorkflowExecutionStarted.String())
	wfSignal := NewHistoryEvent(shared.EventTypeWorkflowExecutionSignaled.String())
	wfComplete := NewHistoryEvent(shared.EventTypeWorkflowExecutionCompleted.String())
	continueAsNew := NewHistoryEvent(shared.EventTypeWorkflowExecutionContinuedAsNew.String())
	wfFail := NewHistoryEvent(shared.EventTypeWorkflowExecutionFailed.String())
	wfCancel := NewHistoryEvent(shared.EventTypeWorkflowExecutionCanceled.String())
	wfCancelRequest := NewHistoryEvent(shared.EventTypeWorkflowExecutionCancelRequested.String()) //?
	wfTerminate := NewHistoryEvent(shared.EventTypeWorkflowExecutionTerminated.String())
	wfTimedOut := NewHistoryEvent(shared.EventTypeWorkflowExecutionTimedOut.String())

	wfStartToSignal := NewConnection(wfStart, wfSignal)
	wfStartToDTSchedule := NewConnection(wfStart, dtSchedule)
	wfStartToDTSchedule.SetCondition(notPendingDecisionTask)
	wfSignalToDTSchedule := NewConnection(wfSignal, dtSchedule)
	wfSignalToDTSchedule.SetCondition(notPendingDecisionTask)
	dtCompleteToWFComplete := NewConnection(dtComplete, wfComplete)
	dtCompleteToWFComplete.SetCondition(containActivityComplete)
	dtCompleteToWFFailed := NewConnection(dtComplete, wfFail)
	dtCompleteToWFFailed.SetCondition(containActivityComplete)
	dtCompleteToCAN := NewConnection(dtComplete, continueAsNew)
	dtCompleteToCAN.SetCondition(containActivityComplete)
	wfCancelRequestToCancel := NewConnection(wfCancelRequest, wfCancel)
	wfModel.AddEdge(wfStartToSignal, wfStartToDTSchedule, wfSignalToDTSchedule, dtCompleteToCAN, dtCompleteToWFComplete, dtCompleteToWFFailed, wfCancelRequestToCancel)

	atModel := NewHistoryEventModel()
	atSchedule := NewHistoryEvent(shared.EventTypeActivityTaskScheduled.String())
	atStart := NewHistoryEvent(shared.EventTypeActivityTaskStarted.String())
	atComplete := NewHistoryEvent(shared.EventTypeActivityTaskCompleted.String())
	atFail := NewHistoryEvent(shared.EventTypeActivityTaskFailed.String())
	atTimedOut := NewHistoryEvent(shared.EventTypeActivityTaskTimedOut.String())
	atCancelRequest := NewHistoryEvent(shared.EventTypeActivityTaskCancelRequested.String()) //?
	atCancel := NewHistoryEvent(shared.EventTypeActivityTaskCanceled.String())
	atCancelRequestFail := NewHistoryEvent(shared.EventTypeRequestCancelActivityTaskFailed.String())

	dtCompleteToATSchedule := NewConnection(dtComplete, atSchedule)
	atScheduleToStart := NewConnection(atSchedule, atStart)
	atStartToComplete := NewConnection(atStart, atComplete)
	atStartToFail := NewConnection(atStart, atFail)
	atStartToTimedOut := NewConnection(atStart, atTimedOut)
	atCompleteToDTSchedule := NewConnection(atComplete, dtSchedule)
	atCompleteToDTSchedule.SetCondition(notPendingDecisionTask)
	atFailToDTSchedule := NewConnection(atFail, dtSchedule)
	atFailToDTSchedule.SetCondition(notPendingDecisionTask)
	atTimedOutToDTSchedule := NewConnection(atTimedOut, dtSchedule)
	atTimedOutToDTSchedule.SetCondition(notPendingDecisionTask)
	atCancelReqToCancel := NewConnection(atCancelRequest, atCancel)
	atCancelReqToCancelFail := NewConnection(atCancelRequest, atCancelRequestFail)
	atCancelToDTSchedule := NewConnection(atCancel, dtSchedule)
	atCancelToDTSchedule.SetCondition(notPendingDecisionTask)
	atCancelRequestFailToDTSchedule := NewConnection(atCancelRequestFail, dtSchedule)
	atCancelRequestFailToDTSchedule.SetCondition(notPendingDecisionTask)
	//atScheduleToTimedOut := NewConnection(atSchedule, atTimedOut)
	//atScheduleToCancelReq := NewConnection(atSchedule, atCancelRequest)
	atStartToCancalReq := NewConnection(atStart, atCancelRequest)
	atModel.AddEdge(dtCompleteToATSchedule, atScheduleToStart, atStartToComplete, atStartToFail, atStartToTimedOut,
		dtCompleteToATSchedule, atCompleteToDTSchedule, atFailToDTSchedule, atTimedOutToDTSchedule, atCancelReqToCancel,
		atCancelReqToCancelFail, atCancelToDTSchedule, atStartToCancalReq,
		atCancelRequestFailToDTSchedule)

	timerModel := NewHistoryEventModel()
	timerStart := NewHistoryEvent(shared.EventTypeTimerStarted.String())
	timerFired := NewHistoryEvent(shared.EventTypeTimerFired.String())
	timerCancel := NewHistoryEvent(shared.EventTypeTimerCanceled.String())

	timerStartToFire := NewConnection(timerStart, timerFired)
	timerStartToCancel := NewConnection(timerStart, timerCancel)
	dtCompleteToTimerStart := NewConnection(dtComplete, timerStart)
	timerFiredToDTSchedule := NewConnection(timerFired, dtSchedule)
	timerFiredToDTSchedule.SetCondition(notPendingDecisionTask)
	timerCancelToDTSchedule := NewConnection(timerCancel, dtSchedule)
	timerCancelToDTSchedule.SetCondition(notPendingDecisionTask)
	timerModel.AddEdge(timerStartToFire, timerStartToCancel, dtCompleteToTimerStart, timerFiredToDTSchedule, timerCancelToDTSchedule)

	cwfModel := NewHistoryEventModel()
	cwfInitial := NewHistoryEvent(shared.EventTypeStartChildWorkflowExecutionInitiated.String())
	cwfInitialFail := NewHistoryEvent(shared.EventTypeStartChildWorkflowExecutionFailed.String())
	cwfStart := NewHistoryEvent(shared.EventTypeChildWorkflowExecutionStarted.String())
	cwfCancel := NewHistoryEvent(shared.EventTypeChildWorkflowExecutionCanceled.String())
	cwfComplete := NewHistoryEvent(shared.EventTypeChildWorkflowExecutionCompleted.String())
	cwfFail := NewHistoryEvent(shared.EventTypeChildWorkflowExecutionFailed.String())
	cwfTerminate := NewHistoryEvent(shared.EventTypeChildWorkflowExecutionTerminated.String())
	cwfTimedOut := NewHistoryEvent(shared.EventTypeChildWorkflowExecutionTimedOut.String())

	dtCompleteToCWFInitial := NewConnection(dtComplete, cwfInitial)
	cwfInitialToFail := NewConnection(cwfInitial, cwfInitialFail)
	cwfInitialToStart := NewConnection(cwfInitial, cwfStart)
	cwfStartToCancel := NewConnection(cwfStart, cwfCancel)
	cwfStartToFail := NewConnection(cwfStart, cwfFail)
	cwfStartToComplete := NewConnection(cwfStart, cwfComplete)
	cwfStartToTerminate := NewConnection(cwfStart, cwfTerminate)
	cwfStartToTimedOut := NewConnection(cwfStart, cwfTimedOut)
	cwfCancelToDTSchedule := NewConnection(cwfCancel, dtSchedule)
	cwfCancelToDTSchedule.SetCondition(notPendingDecisionTask)
	cwfFailToDTSchedule := NewConnection(cwfFail, dtSchedule)
	cwfFailToDTSchedule.SetCondition(notPendingDecisionTask)
	cwfCompleteToDTSchedule := NewConnection(cwfComplete, dtSchedule)
	cwfCompleteToDTSchedule.SetCondition(notPendingDecisionTask)
	cwfTerminateToDTSchedule := NewConnection(cwfTerminate, dtSchedule)
	cwfTerminateToDTSchedule.SetCondition(notPendingDecisionTask)
	cwfTimedOutToDTSchedule := NewConnection(cwfTimedOut, dtSchedule)
	cwfTimedOutToDTSchedule.SetCondition(notPendingDecisionTask)
	cwfInitialFailToDTSchedule := NewConnection(cwfInitialFail, dtSchedule)
	cwfInitialFailToDTSchedule.SetCondition(notPendingDecisionTask)
	cwfModel.AddEdge(dtCompleteToCWFInitial, cwfInitialToFail, cwfInitialToStart, cwfStartToCancel, cwfStartToFail,
		cwfStartToComplete, cwfStartToTerminate, cwfStartToTimedOut, cwfCancelToDTSchedule, cwfFailToDTSchedule,
		cwfCompleteToDTSchedule, cwfTerminateToDTSchedule, cwfTimedOutToDTSchedule, cwfInitialFailToDTSchedule)

	ewfModel := NewHistoryEventModel()
	ewfSignal := NewHistoryEvent(shared.EventTypeSignalExternalWorkflowExecutionInitiated.String())
	ewfSignalFailed := NewHistoryEvent(shared.EventTypeSignalExternalWorkflowExecutionFailed.String())
	ewfSignaled := NewHistoryEvent(shared.EventTypeExternalWorkflowExecutionSignaled.String())
	ewfCancel := NewHistoryEvent(shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated.String())
	ewfCancelFail := NewHistoryEvent(shared.EventTypeRequestCancelExternalWorkflowExecutionFailed.String())
	ewfCanceled := NewHistoryEvent(shared.EventTypeExternalWorkflowExecutionCancelRequested.String())

	dtCompleteToEWFSignal := NewConnection(dtComplete, ewfSignal)
	dtCompleteToEWFCancel := NewConnection(dtComplete, ewfCancel)
	ewfSignalToFail := NewConnection(ewfSignal, ewfSignalFailed)
	ewfSignalToSignaled := NewConnection(ewfSignal, ewfSignaled)
	ewfCancelToFail := NewConnection(ewfCancel, ewfCancelFail)
	ewfCancelToCanceled := NewConnection(ewfCancel, ewfCanceled)
	ewfSignaledToDTSchedule := NewConnection(ewfSignaled, dtSchedule)
	ewfSignaledToDTSchedule.SetCondition(notPendingDecisionTask)
	ewfSignalFailedToDTSchedule := NewConnection(ewfSignalFailed, dtSchedule)
	ewfSignalFailedToDTSchedule.SetCondition(notPendingDecisionTask)
	ewfCanceledToDTSchedule := NewConnection(ewfCanceled, dtSchedule)
	ewfCanceledToDTSchedule.SetCondition(notPendingDecisionTask)
	ewfCancelFailToDTSchedule := NewConnection(ewfCancelFail, dtSchedule)
	ewfCancelFailToDTSchedule.SetCondition(notPendingDecisionTask)
	ewfModel.AddEdge(dtCompleteToEWFSignal, dtCompleteToEWFCancel, ewfSignalToFail, ewfSignalToSignaled, ewfCancelToFail,
		ewfCancelToCanceled, ewfSignaledToDTSchedule, ewfSignalFailedToDTSchedule, ewfCanceledToDTSchedule, ewfCancelFailToDTSchedule)

	generator := NewEventGenerator()
	generator.AddInitialEntryVertex(wfStart)
	generator.AddExitVertex(wfComplete, wfFail, continueAsNew, wfTerminate, wfTimedOut)
	generator.AddRandomEntryVertex(wfSignal, wfTerminate, wfTimedOut)

	generator.AddModel(dtModel)
	generator.AddModel(wfModel)
	generator.AddModel(atModel)
	generator.AddModel(timerModel)
	generator.AddModel(cwfModel)
	generator.AddModel(ewfModel)
	s.generator = generator
}

func (s *historyEventTestSuit) SetupTest() {
	s.generator.Reset()
}

func (s *historyEventTestSuit) Test_HistoryEvent_Generator() {
	for s.generator.HasNextVertex() {
		v := s.generator.GetNextVertex()
		for _, e := range v {
			fmt.Println(e.GetName())
		}
	}
	s.NotEmpty(s.generator.ListGeneratedVertex())
}
