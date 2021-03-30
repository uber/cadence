// Copyright (c) 2021 Uber Technologies Inc.
//
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

package testdata

import "github.com/uber/cadence/common/types"

var (
	ArchivalStatus                             = types.ArchivalStatusEnabled
	CancelExternalWorkflowExecutionFailedCause = types.CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution
	ChildWorkflowExecutionFailedCause          = types.ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning
	ContinueAsNewInitiator                     = types.ContinueAsNewInitiatorRetryPolicy
	DecisionTaskFailedCause                    = types.DecisionTaskFailedCauseBadCancelWorkflowExecutionAttributes
	DecisionTaskTimedOutCause                  = types.DecisionTaskTimedOutCauseReset
	DecisionType                               = types.DecisionTypeCancelTimer
	DomainStatus                               = types.DomainStatusRegistered
	EncodingType                               = types.EncodingTypeJSON
	EventType                                  = types.EventTypeWorkflowExecutionStarted
	HistoryEventFilterType                     = types.HistoryEventFilterTypeCloseEvent
	IndexedValueType                           = types.IndexedValueTypeInt
	ParentClosePolicy                          = types.ParentClosePolicyTerminate
	PendingActivityState                       = types.PendingActivityStateCancelRequested
	PendingDecisionState                       = types.PendingDecisionStateStarted
	QueryConsistencyLevel                      = types.QueryConsistencyLevelStrong
	QueryRejectCondition                       = types.QueryRejectConditionNotCompletedCleanly
	QueryResultType                            = types.QueryResultTypeFailed
	QueryTaskCompletedType                     = types.QueryTaskCompletedTypeFailed
	SignalExternalWorkflowExecutionFailedCause = types.SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution
	TaskListKind                               = types.TaskListKindSticky
	TaskListType                               = types.TaskListTypeActivity
	TimeoutType                                = types.TimeoutTypeScheduleToStart
	WorkflowExecutionCloseStatus               = types.WorkflowExecutionCloseStatusContinuedAsNew
	WorkflowIDReusePolicy                      = types.WorkflowIDReusePolicyTerminateIfRunning
)
