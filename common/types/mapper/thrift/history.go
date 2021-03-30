// Copyright (c) 2017-2020 Uber Technologies Inc.
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

package thrift

import (
	"github.com/uber/cadence/common/types"

	"github.com/uber/cadence/.gen/go/history"
)

// FromDescribeMutableStateRequest converts internal DescribeMutableStateRequest type to thrift
func FromDescribeMutableStateRequest(t *types.DescribeMutableStateRequest) *history.DescribeMutableStateRequest {
	if t == nil {
		return nil
	}
	return &history.DescribeMutableStateRequest{
		DomainUUID: &t.DomainUUID,
		Execution:  FromWorkflowExecution(t.Execution),
	}
}

// ToDescribeMutableStateRequest converts thrift DescribeMutableStateRequest type to internal
func ToDescribeMutableStateRequest(t *history.DescribeMutableStateRequest) *types.DescribeMutableStateRequest {
	if t == nil {
		return nil
	}
	return &types.DescribeMutableStateRequest{
		DomainUUID: t.GetDomainUUID(),
		Execution:  ToWorkflowExecution(t.Execution),
	}
}

// FromDescribeMutableStateResponse converts internal DescribeMutableStateResponse type to thrift
func FromDescribeMutableStateResponse(t *types.DescribeMutableStateResponse) *history.DescribeMutableStateResponse {
	if t == nil {
		return nil
	}
	return &history.DescribeMutableStateResponse{
		MutableStateInCache:    &t.MutableStateInCache,
		MutableStateInDatabase: &t.MutableStateInDatabase,
	}
}

// ToDescribeMutableStateResponse converts thrift DescribeMutableStateResponse type to internal
func ToDescribeMutableStateResponse(t *history.DescribeMutableStateResponse) *types.DescribeMutableStateResponse {
	if t == nil {
		return nil
	}
	return &types.DescribeMutableStateResponse{
		MutableStateInCache:    t.GetMutableStateInCache(),
		MutableStateInDatabase: t.GetMutableStateInDatabase(),
	}
}

// FromHistoryDescribeWorkflowExecutionRequest converts internal DescribeWorkflowExecutionRequest type to thrift
func FromHistoryDescribeWorkflowExecutionRequest(t *types.HistoryDescribeWorkflowExecutionRequest) *history.DescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.DescribeWorkflowExecutionRequest{
		DomainUUID: &t.DomainUUID,
		Request:    FromDescribeWorkflowExecutionRequest(t.Request),
	}
}

// ToHistoryDescribeWorkflowExecutionRequest converts thrift DescribeWorkflowExecutionRequest type to internal
func ToHistoryDescribeWorkflowExecutionRequest(t *history.DescribeWorkflowExecutionRequest) *types.HistoryDescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryDescribeWorkflowExecutionRequest{
		DomainUUID: t.GetDomainUUID(),
		Request:    ToDescribeWorkflowExecutionRequest(t.Request),
	}
}

// FromDomainFilter converts internal DomainFilter type to thrift
func FromDomainFilter(t *types.DomainFilter) *history.DomainFilter {
	if t == nil {
		return nil
	}
	return &history.DomainFilter{
		DomainIDs:    t.DomainIDs,
		ReverseMatch: &t.ReverseMatch,
	}
}

// ToDomainFilter converts thrift DomainFilter type to internal
func ToDomainFilter(t *history.DomainFilter) *types.DomainFilter {
	if t == nil {
		return nil
	}
	return &types.DomainFilter{
		DomainIDs:    t.DomainIDs,
		ReverseMatch: t.GetReverseMatch(),
	}
}

// FromEventAlreadyStartedError converts internal EventAlreadyStartedError type to thrift
func FromEventAlreadyStartedError(t *types.EventAlreadyStartedError) *history.EventAlreadyStartedError {
	if t == nil {
		return nil
	}
	return &history.EventAlreadyStartedError{
		Message: t.Message,
	}
}

// ToEventAlreadyStartedError converts thrift EventAlreadyStartedError type to internal
func ToEventAlreadyStartedError(t *history.EventAlreadyStartedError) *types.EventAlreadyStartedError {
	if t == nil {
		return nil
	}
	return &types.EventAlreadyStartedError{
		Message: t.Message,
	}
}

// FromFailoverMarkerToken converts internal FailoverMarkerToken type to thrift
func FromFailoverMarkerToken(t *types.FailoverMarkerToken) *history.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	return &history.FailoverMarkerToken{
		ShardIDs:       t.ShardIDs,
		FailoverMarker: FromFailoverMarkerAttributes(t.FailoverMarker),
	}
}

// ToFailoverMarkerToken converts thrift FailoverMarkerToken type to internal
func ToFailoverMarkerToken(t *history.FailoverMarkerToken) *types.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	return &types.FailoverMarkerToken{
		ShardIDs:       t.ShardIDs,
		FailoverMarker: ToFailoverMarkerAttributes(t.FailoverMarker),
	}
}

// FromGetMutableStateRequest converts internal GetMutableStateRequest type to thrift
func FromGetMutableStateRequest(t *types.GetMutableStateRequest) *history.GetMutableStateRequest {
	if t == nil {
		return nil
	}
	return &history.GetMutableStateRequest{
		DomainUUID:          &t.DomainUUID,
		Execution:           FromWorkflowExecution(t.Execution),
		ExpectedNextEventId: &t.ExpectedNextEventID,
		CurrentBranchToken:  t.CurrentBranchToken,
	}
}

// ToGetMutableStateRequest converts thrift GetMutableStateRequest type to internal
func ToGetMutableStateRequest(t *history.GetMutableStateRequest) *types.GetMutableStateRequest {
	if t == nil {
		return nil
	}
	return &types.GetMutableStateRequest{
		DomainUUID:          t.GetDomainUUID(),
		Execution:           ToWorkflowExecution(t.Execution),
		ExpectedNextEventID: t.GetExpectedNextEventId(),
		CurrentBranchToken:  t.CurrentBranchToken,
	}
}

// FromGetMutableStateResponse converts internal GetMutableStateResponse type to thrift
func FromGetMutableStateResponse(t *types.GetMutableStateResponse) *history.GetMutableStateResponse {
	if t == nil {
		return nil
	}
	return &history.GetMutableStateResponse{
		Execution:                            FromWorkflowExecution(t.Execution),
		WorkflowType:                         FromWorkflowType(t.WorkflowType),
		NextEventId:                          &t.NextEventID,
		PreviousStartedEventId:               t.PreviousStartedEventID,
		LastFirstEventId:                     &t.LastFirstEventID,
		TaskList:                             FromTaskList(t.TaskList),
		StickyTaskList:                       FromTaskList(t.StickyTaskList),
		ClientLibraryVersion:                 &t.ClientLibraryVersion,
		ClientFeatureVersion:                 &t.ClientFeatureVersion,
		ClientImpl:                           &t.ClientImpl,
		IsWorkflowRunning:                    &t.IsWorkflowRunning,
		StickyTaskListScheduleToStartTimeout: t.StickyTaskListScheduleToStartTimeout,
		EventStoreVersion:                    &t.EventStoreVersion,
		CurrentBranchToken:                   t.CurrentBranchToken,
		WorkflowState:                        t.WorkflowState,
		WorkflowCloseState:                   t.WorkflowCloseState,
		VersionHistories:                     FromVersionHistories(t.VersionHistories),
		IsStickyTaskListEnabled:              &t.IsStickyTaskListEnabled,
	}
}

// ToGetMutableStateResponse converts thrift GetMutableStateResponse type to internal
func ToGetMutableStateResponse(t *history.GetMutableStateResponse) *types.GetMutableStateResponse {
	if t == nil {
		return nil
	}
	return &types.GetMutableStateResponse{
		Execution:                            ToWorkflowExecution(t.Execution),
		WorkflowType:                         ToWorkflowType(t.WorkflowType),
		NextEventID:                          t.GetNextEventId(),
		PreviousStartedEventID:               t.PreviousStartedEventId,
		LastFirstEventID:                     t.GetLastFirstEventId(),
		TaskList:                             ToTaskList(t.TaskList),
		StickyTaskList:                       ToTaskList(t.StickyTaskList),
		ClientLibraryVersion:                 t.GetClientLibraryVersion(),
		ClientFeatureVersion:                 t.GetClientFeatureVersion(),
		ClientImpl:                           t.GetClientImpl(),
		IsWorkflowRunning:                    t.GetIsWorkflowRunning(),
		StickyTaskListScheduleToStartTimeout: t.StickyTaskListScheduleToStartTimeout,
		EventStoreVersion:                    t.GetEventStoreVersion(),
		CurrentBranchToken:                   t.CurrentBranchToken,
		WorkflowState:                        t.WorkflowState,
		WorkflowCloseState:                   t.WorkflowCloseState,
		VersionHistories:                     ToVersionHistories(t.VersionHistories),
		IsStickyTaskListEnabled:              t.GetIsStickyTaskListEnabled(),
	}
}

// FromNotifyFailoverMarkersRequest converts internal NotifyFailoverMarkersRequest type to thrift
func FromNotifyFailoverMarkersRequest(t *types.NotifyFailoverMarkersRequest) *history.NotifyFailoverMarkersRequest {
	if t == nil {
		return nil
	}
	return &history.NotifyFailoverMarkersRequest{
		FailoverMarkerTokens: FromFailoverMarkerTokenArray(t.FailoverMarkerTokens),
	}
}

// ToNotifyFailoverMarkersRequest converts thrift NotifyFailoverMarkersRequest type to internal
func ToNotifyFailoverMarkersRequest(t *history.NotifyFailoverMarkersRequest) *types.NotifyFailoverMarkersRequest {
	if t == nil {
		return nil
	}
	return &types.NotifyFailoverMarkersRequest{
		FailoverMarkerTokens: ToFailoverMarkerTokenArray(t.FailoverMarkerTokens),
	}
}

// FromParentExecutionInfo converts internal ParentExecutionInfo type to thrift
func FromParentExecutionInfo(t *types.ParentExecutionInfo) *history.ParentExecutionInfo {
	if t == nil {
		return nil
	}
	return &history.ParentExecutionInfo{
		DomainUUID:  &t.DomainUUID,
		Domain:      &t.Domain,
		Execution:   FromWorkflowExecution(t.Execution),
		InitiatedId: &t.InitiatedID,
	}
}

// ToParentExecutionInfo converts thrift ParentExecutionInfo type to internal
func ToParentExecutionInfo(t *history.ParentExecutionInfo) *types.ParentExecutionInfo {
	if t == nil {
		return nil
	}
	return &types.ParentExecutionInfo{
		DomainUUID:  t.GetDomainUUID(),
		Domain:      t.GetDomain(),
		Execution:   ToWorkflowExecution(t.Execution),
		InitiatedID: t.GetInitiatedId(),
	}
}

// FromPollMutableStateRequest converts internal PollMutableStateRequest type to thrift
func FromPollMutableStateRequest(t *types.PollMutableStateRequest) *history.PollMutableStateRequest {
	if t == nil {
		return nil
	}
	return &history.PollMutableStateRequest{
		DomainUUID:          &t.DomainUUID,
		Execution:           FromWorkflowExecution(t.Execution),
		ExpectedNextEventId: &t.ExpectedNextEventID,
		CurrentBranchToken:  t.CurrentBranchToken,
	}
}

// ToPollMutableStateRequest converts thrift PollMutableStateRequest type to internal
func ToPollMutableStateRequest(t *history.PollMutableStateRequest) *types.PollMutableStateRequest {
	if t == nil {
		return nil
	}
	return &types.PollMutableStateRequest{
		DomainUUID:          t.GetDomainUUID(),
		Execution:           ToWorkflowExecution(t.Execution),
		ExpectedNextEventID: t.GetExpectedNextEventId(),
		CurrentBranchToken:  t.CurrentBranchToken,
	}
}

// FromPollMutableStateResponse converts internal PollMutableStateResponse type to thrift
func FromPollMutableStateResponse(t *types.PollMutableStateResponse) *history.PollMutableStateResponse {
	if t == nil {
		return nil
	}
	return &history.PollMutableStateResponse{
		Execution:                            FromWorkflowExecution(t.Execution),
		WorkflowType:                         FromWorkflowType(t.WorkflowType),
		NextEventId:                          &t.NextEventID,
		PreviousStartedEventId:               t.PreviousStartedEventID,
		LastFirstEventId:                     &t.LastFirstEventID,
		TaskList:                             FromTaskList(t.TaskList),
		StickyTaskList:                       FromTaskList(t.StickyTaskList),
		ClientLibraryVersion:                 &t.ClientLibraryVersion,
		ClientFeatureVersion:                 &t.ClientFeatureVersion,
		ClientImpl:                           &t.ClientImpl,
		StickyTaskListScheduleToStartTimeout: t.StickyTaskListScheduleToStartTimeout,
		CurrentBranchToken:                   t.CurrentBranchToken,
		VersionHistories:                     FromVersionHistories(t.VersionHistories),
		WorkflowState:                        t.WorkflowState,
		WorkflowCloseState:                   t.WorkflowCloseState,
	}
}

// ToPollMutableStateResponse converts thrift PollMutableStateResponse type to internal
func ToPollMutableStateResponse(t *history.PollMutableStateResponse) *types.PollMutableStateResponse {
	if t == nil {
		return nil
	}
	return &types.PollMutableStateResponse{
		Execution:                            ToWorkflowExecution(t.Execution),
		WorkflowType:                         ToWorkflowType(t.WorkflowType),
		NextEventID:                          t.GetNextEventId(),
		PreviousStartedEventID:               t.PreviousStartedEventId,
		LastFirstEventID:                     t.GetLastFirstEventId(),
		TaskList:                             ToTaskList(t.TaskList),
		StickyTaskList:                       ToTaskList(t.StickyTaskList),
		ClientLibraryVersion:                 t.GetClientLibraryVersion(),
		ClientFeatureVersion:                 t.GetClientFeatureVersion(),
		ClientImpl:                           t.GetClientImpl(),
		StickyTaskListScheduleToStartTimeout: t.StickyTaskListScheduleToStartTimeout,
		CurrentBranchToken:                   t.CurrentBranchToken,
		VersionHistories:                     ToVersionHistories(t.VersionHistories),
		WorkflowState:                        t.WorkflowState,
		WorkflowCloseState:                   t.WorkflowCloseState,
	}
}

// FromProcessingQueueState converts internal ProcessingQueueState type to thrift
func FromProcessingQueueState(t *types.ProcessingQueueState) *history.ProcessingQueueState {
	if t == nil {
		return nil
	}
	return &history.ProcessingQueueState{
		Level:        t.Level,
		AckLevel:     t.AckLevel,
		MaxLevel:     t.MaxLevel,
		DomainFilter: FromDomainFilter(t.DomainFilter),
	}
}

// ToProcessingQueueState converts thrift ProcessingQueueState type to internal
func ToProcessingQueueState(t *history.ProcessingQueueState) *types.ProcessingQueueState {
	if t == nil {
		return nil
	}
	return &types.ProcessingQueueState{
		Level:        t.Level,
		AckLevel:     t.AckLevel,
		MaxLevel:     t.MaxLevel,
		DomainFilter: ToDomainFilter(t.DomainFilter),
	}
}

// FromProcessingQueueStates converts internal ProcessingQueueStates type to thrift
func FromProcessingQueueStates(t *types.ProcessingQueueStates) *history.ProcessingQueueStates {
	if t == nil {
		return nil
	}
	return &history.ProcessingQueueStates{
		StatesByCluster: FromProcessingQueueStateArrayMap(t.StatesByCluster),
	}
}

// ToProcessingQueueStates converts thrift ProcessingQueueStates type to internal
func ToProcessingQueueStates(t *history.ProcessingQueueStates) *types.ProcessingQueueStates {
	if t == nil {
		return nil
	}
	return &types.ProcessingQueueStates{
		StatesByCluster: ToProcessingQueueStateArrayMap(t.StatesByCluster),
	}
}

// FromHistoryQueryWorkflowRequest converts internal QueryWorkflowRequest type to thrift
func FromHistoryQueryWorkflowRequest(t *types.HistoryQueryWorkflowRequest) *history.QueryWorkflowRequest {
	if t == nil {
		return nil
	}
	return &history.QueryWorkflowRequest{
		DomainUUID: &t.DomainUUID,
		Request:    FromQueryWorkflowRequest(t.Request),
	}
}

// ToHistoryQueryWorkflowRequest converts thrift QueryWorkflowRequest type to internal
func ToHistoryQueryWorkflowRequest(t *history.QueryWorkflowRequest) *types.HistoryQueryWorkflowRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryQueryWorkflowRequest{
		DomainUUID: t.GetDomainUUID(),
		Request:    ToQueryWorkflowRequest(t.Request),
	}
}

// FromHistoryQueryWorkflowResponse converts internal QueryWorkflowResponse type to thrift
func FromHistoryQueryWorkflowResponse(t *types.HistoryQueryWorkflowResponse) *history.QueryWorkflowResponse {
	if t == nil {
		return nil
	}
	return &history.QueryWorkflowResponse{
		Response: FromQueryWorkflowResponse(t.Response),
	}
}

// ToHistoryQueryWorkflowResponse converts thrift QueryWorkflowResponse type to internal
func ToHistoryQueryWorkflowResponse(t *history.QueryWorkflowResponse) *types.HistoryQueryWorkflowResponse {
	if t == nil {
		return nil
	}
	return &types.HistoryQueryWorkflowResponse{
		Response: ToQueryWorkflowResponse(t.Response),
	}
}

// FromHistoryReapplyEventsRequest converts internal ReapplyEventsRequest type to thrift
func FromHistoryReapplyEventsRequest(t *types.HistoryReapplyEventsRequest) *history.ReapplyEventsRequest {
	if t == nil {
		return nil
	}
	return &history.ReapplyEventsRequest{
		DomainUUID: &t.DomainUUID,
		Request:    FromReapplyEventsRequest(t.Request),
	}
}

// ToHistoryReapplyEventsRequest converts thrift ReapplyEventsRequest type to internal
func ToHistoryReapplyEventsRequest(t *history.ReapplyEventsRequest) *types.HistoryReapplyEventsRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryReapplyEventsRequest{
		DomainUUID: t.GetDomainUUID(),
		Request:    ToReapplyEventsRequest(t.Request),
	}
}

// FromHistoryRecordActivityTaskHeartbeatRequest converts internal RecordActivityTaskHeartbeatRequest type to thrift
func FromHistoryRecordActivityTaskHeartbeatRequest(t *types.HistoryRecordActivityTaskHeartbeatRequest) *history.RecordActivityTaskHeartbeatRequest {
	if t == nil {
		return nil
	}
	return &history.RecordActivityTaskHeartbeatRequest{
		DomainUUID:       &t.DomainUUID,
		HeartbeatRequest: FromRecordActivityTaskHeartbeatRequest(t.HeartbeatRequest),
	}
}

// ToHistoryRecordActivityTaskHeartbeatRequest converts thrift RecordActivityTaskHeartbeatRequest type to internal
func ToHistoryRecordActivityTaskHeartbeatRequest(t *history.RecordActivityTaskHeartbeatRequest) *types.HistoryRecordActivityTaskHeartbeatRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRecordActivityTaskHeartbeatRequest{
		DomainUUID:       t.GetDomainUUID(),
		HeartbeatRequest: ToRecordActivityTaskHeartbeatRequest(t.HeartbeatRequest),
	}
}

// FromRecordActivityTaskStartedRequest converts internal RecordActivityTaskStartedRequest type to thrift
func FromRecordActivityTaskStartedRequest(t *types.RecordActivityTaskStartedRequest) *history.RecordActivityTaskStartedRequest {
	if t == nil {
		return nil
	}
	return &history.RecordActivityTaskStartedRequest{
		DomainUUID:        &t.DomainUUID,
		WorkflowExecution: FromWorkflowExecution(t.WorkflowExecution),
		ScheduleId:        &t.ScheduleID,
		TaskId:            &t.TaskID,
		RequestId:         &t.RequestID,
		PollRequest:       FromPollForActivityTaskRequest(t.PollRequest),
	}
}

// ToRecordActivityTaskStartedRequest converts thrift RecordActivityTaskStartedRequest type to internal
func ToRecordActivityTaskStartedRequest(t *history.RecordActivityTaskStartedRequest) *types.RecordActivityTaskStartedRequest {
	if t == nil {
		return nil
	}
	return &types.RecordActivityTaskStartedRequest{
		DomainUUID:        t.GetDomainUUID(),
		WorkflowExecution: ToWorkflowExecution(t.WorkflowExecution),
		ScheduleID:        t.GetScheduleId(),
		TaskID:            t.GetTaskId(),
		RequestID:         t.GetRequestId(),
		PollRequest:       ToPollForActivityTaskRequest(t.PollRequest),
	}
}

// FromRecordActivityTaskStartedResponse converts internal RecordActivityTaskStartedResponse type to thrift
func FromRecordActivityTaskStartedResponse(t *types.RecordActivityTaskStartedResponse) *history.RecordActivityTaskStartedResponse {
	if t == nil {
		return nil
	}
	return &history.RecordActivityTaskStartedResponse{
		ScheduledEvent:                  FromHistoryEvent(t.ScheduledEvent),
		StartedTimestamp:                t.StartedTimestamp,
		Attempt:                         &t.Attempt,
		ScheduledTimestampOfThisAttempt: t.ScheduledTimestampOfThisAttempt,
		HeartbeatDetails:                t.HeartbeatDetails,
		WorkflowType:                    FromWorkflowType(t.WorkflowType),
		WorkflowDomain:                  &t.WorkflowDomain,
	}
}

// ToRecordActivityTaskStartedResponse converts thrift RecordActivityTaskStartedResponse type to internal
func ToRecordActivityTaskStartedResponse(t *history.RecordActivityTaskStartedResponse) *types.RecordActivityTaskStartedResponse {
	if t == nil {
		return nil
	}
	return &types.RecordActivityTaskStartedResponse{
		ScheduledEvent:                  ToHistoryEvent(t.ScheduledEvent),
		StartedTimestamp:                t.StartedTimestamp,
		Attempt:                         t.GetAttempt(),
		ScheduledTimestampOfThisAttempt: t.ScheduledTimestampOfThisAttempt,
		HeartbeatDetails:                t.HeartbeatDetails,
		WorkflowType:                    ToWorkflowType(t.WorkflowType),
		WorkflowDomain:                  t.GetWorkflowDomain(),
	}
}

// FromRecordChildExecutionCompletedRequest converts internal RecordChildExecutionCompletedRequest type to thrift
func FromRecordChildExecutionCompletedRequest(t *types.RecordChildExecutionCompletedRequest) *history.RecordChildExecutionCompletedRequest {
	if t == nil {
		return nil
	}
	return &history.RecordChildExecutionCompletedRequest{
		DomainUUID:         &t.DomainUUID,
		WorkflowExecution:  FromWorkflowExecution(t.WorkflowExecution),
		InitiatedId:        &t.InitiatedID,
		CompletedExecution: FromWorkflowExecution(t.CompletedExecution),
		CompletionEvent:    FromHistoryEvent(t.CompletionEvent),
	}
}

// ToRecordChildExecutionCompletedRequest converts thrift RecordChildExecutionCompletedRequest type to internal
func ToRecordChildExecutionCompletedRequest(t *history.RecordChildExecutionCompletedRequest) *types.RecordChildExecutionCompletedRequest {
	if t == nil {
		return nil
	}
	return &types.RecordChildExecutionCompletedRequest{
		DomainUUID:         t.GetDomainUUID(),
		WorkflowExecution:  ToWorkflowExecution(t.WorkflowExecution),
		InitiatedID:        t.GetInitiatedId(),
		CompletedExecution: ToWorkflowExecution(t.CompletedExecution),
		CompletionEvent:    ToHistoryEvent(t.CompletionEvent),
	}
}

// FromRecordDecisionTaskStartedRequest converts internal RecordDecisionTaskStartedRequest type to thrift
func FromRecordDecisionTaskStartedRequest(t *types.RecordDecisionTaskStartedRequest) *history.RecordDecisionTaskStartedRequest {
	if t == nil {
		return nil
	}
	return &history.RecordDecisionTaskStartedRequest{
		DomainUUID:        &t.DomainUUID,
		WorkflowExecution: FromWorkflowExecution(t.WorkflowExecution),
		ScheduleId:        &t.ScheduleID,
		TaskId:            &t.TaskID,
		RequestId:         &t.RequestID,
		PollRequest:       FromPollForDecisionTaskRequest(t.PollRequest),
	}
}

// ToRecordDecisionTaskStartedRequest converts thrift RecordDecisionTaskStartedRequest type to internal
func ToRecordDecisionTaskStartedRequest(t *history.RecordDecisionTaskStartedRequest) *types.RecordDecisionTaskStartedRequest {
	if t == nil {
		return nil
	}
	return &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        t.GetDomainUUID(),
		WorkflowExecution: ToWorkflowExecution(t.WorkflowExecution),
		ScheduleID:        t.GetScheduleId(),
		TaskID:            t.GetTaskId(),
		RequestID:         t.GetRequestId(),
		PollRequest:       ToPollForDecisionTaskRequest(t.PollRequest),
	}
}

// FromRecordDecisionTaskStartedResponse converts internal RecordDecisionTaskStartedResponse type to thrift
func FromRecordDecisionTaskStartedResponse(t *types.RecordDecisionTaskStartedResponse) *history.RecordDecisionTaskStartedResponse {
	if t == nil {
		return nil
	}
	return &history.RecordDecisionTaskStartedResponse{
		WorkflowType:              FromWorkflowType(t.WorkflowType),
		PreviousStartedEventId:    t.PreviousStartedEventID,
		ScheduledEventId:          &t.ScheduledEventID,
		StartedEventId:            &t.StartedEventID,
		NextEventId:               &t.NextEventID,
		Attempt:                   &t.Attempt,
		StickyExecutionEnabled:    &t.StickyExecutionEnabled,
		DecisionInfo:              FromTransientDecisionInfo(t.DecisionInfo),
		WorkflowExecutionTaskList: FromTaskList(t.WorkflowExecutionTaskList),
		EventStoreVersion:         &t.EventStoreVersion,
		BranchToken:               t.BranchToken,
		ScheduledTimestamp:        t.ScheduledTimestamp,
		StartedTimestamp:          t.StartedTimestamp,
		Queries:                   FromWorkflowQueryMap(t.Queries),
	}
}

// ToRecordDecisionTaskStartedResponse converts thrift RecordDecisionTaskStartedResponse type to internal
func ToRecordDecisionTaskStartedResponse(t *history.RecordDecisionTaskStartedResponse) *types.RecordDecisionTaskStartedResponse {
	if t == nil {
		return nil
	}
	return &types.RecordDecisionTaskStartedResponse{
		WorkflowType:              ToWorkflowType(t.WorkflowType),
		PreviousStartedEventID:    t.PreviousStartedEventId,
		ScheduledEventID:          t.GetScheduledEventId(),
		StartedEventID:            t.GetStartedEventId(),
		NextEventID:               t.GetNextEventId(),
		Attempt:                   t.GetAttempt(),
		StickyExecutionEnabled:    t.GetStickyExecutionEnabled(),
		DecisionInfo:              ToTransientDecisionInfo(t.DecisionInfo),
		WorkflowExecutionTaskList: ToTaskList(t.WorkflowExecutionTaskList),
		EventStoreVersion:         t.GetEventStoreVersion(),
		BranchToken:               t.BranchToken,
		ScheduledTimestamp:        t.ScheduledTimestamp,
		StartedTimestamp:          t.StartedTimestamp,
		Queries:                   ToWorkflowQueryMap(t.Queries),
	}
}

// FromHistoryRefreshWorkflowTasksRequest converts internal RefreshWorkflowTasksRequest type to thrift
func FromHistoryRefreshWorkflowTasksRequest(t *types.HistoryRefreshWorkflowTasksRequest) *history.RefreshWorkflowTasksRequest {
	if t == nil {
		return nil
	}
	return &history.RefreshWorkflowTasksRequest{
		DomainUIID: &t.DomainUIID,
		Request:    FromRefreshWorkflowTasksRequest(t.Request),
	}
}

// ToHistoryRefreshWorkflowTasksRequest converts thrift RefreshWorkflowTasksRequest type to internal
func ToHistoryRefreshWorkflowTasksRequest(t *history.RefreshWorkflowTasksRequest) *types.HistoryRefreshWorkflowTasksRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRefreshWorkflowTasksRequest{
		DomainUIID: t.GetDomainUIID(),
		Request:    ToRefreshWorkflowTasksRequest(t.Request),
	}
}

// FromRemoveSignalMutableStateRequest converts internal RemoveSignalMutableStateRequest type to thrift
func FromRemoveSignalMutableStateRequest(t *types.RemoveSignalMutableStateRequest) *history.RemoveSignalMutableStateRequest {
	if t == nil {
		return nil
	}
	return &history.RemoveSignalMutableStateRequest{
		DomainUUID:        &t.DomainUUID,
		WorkflowExecution: FromWorkflowExecution(t.WorkflowExecution),
		RequestId:         &t.RequestID,
	}
}

// ToRemoveSignalMutableStateRequest converts thrift RemoveSignalMutableStateRequest type to internal
func ToRemoveSignalMutableStateRequest(t *history.RemoveSignalMutableStateRequest) *types.RemoveSignalMutableStateRequest {
	if t == nil {
		return nil
	}
	return &types.RemoveSignalMutableStateRequest{
		DomainUUID:        t.GetDomainUUID(),
		WorkflowExecution: ToWorkflowExecution(t.WorkflowExecution),
		RequestID:         t.GetRequestId(),
	}
}

// FromReplicateEventsV2Request converts internal ReplicateEventsV2Request type to thrift
func FromReplicateEventsV2Request(t *types.ReplicateEventsV2Request) *history.ReplicateEventsV2Request {
	if t == nil {
		return nil
	}
	return &history.ReplicateEventsV2Request{
		DomainUUID:          &t.DomainUUID,
		WorkflowExecution:   FromWorkflowExecution(t.WorkflowExecution),
		VersionHistoryItems: FromVersionHistoryItemArray(t.VersionHistoryItems),
		Events:              FromDataBlob(t.Events),
		NewRunEvents:        FromDataBlob(t.NewRunEvents),
	}
}

// ToReplicateEventsV2Request converts thrift ReplicateEventsV2Request type to internal
func ToReplicateEventsV2Request(t *history.ReplicateEventsV2Request) *types.ReplicateEventsV2Request {
	if t == nil {
		return nil
	}
	return &types.ReplicateEventsV2Request{
		DomainUUID:          t.GetDomainUUID(),
		WorkflowExecution:   ToWorkflowExecution(t.WorkflowExecution),
		VersionHistoryItems: ToVersionHistoryItemArray(t.VersionHistoryItems),
		Events:              ToDataBlob(t.Events),
		NewRunEvents:        ToDataBlob(t.NewRunEvents),
	}
}

// FromHistoryRequestCancelWorkflowExecutionRequest converts internal RequestCancelWorkflowExecutionRequest type to thrift
func FromHistoryRequestCancelWorkflowExecutionRequest(t *types.HistoryRequestCancelWorkflowExecutionRequest) *history.RequestCancelWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.RequestCancelWorkflowExecutionRequest{
		DomainUUID:                &t.DomainUUID,
		CancelRequest:             FromRequestCancelWorkflowExecutionRequest(t.CancelRequest),
		ExternalInitiatedEventId:  t.ExternalInitiatedEventID,
		ExternalWorkflowExecution: FromWorkflowExecution(t.ExternalWorkflowExecution),
		ChildWorkflowOnly:         &t.ChildWorkflowOnly,
	}
}

// ToHistoryRequestCancelWorkflowExecutionRequest converts thrift RequestCancelWorkflowExecutionRequest type to internal
func ToHistoryRequestCancelWorkflowExecutionRequest(t *history.RequestCancelWorkflowExecutionRequest) *types.HistoryRequestCancelWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRequestCancelWorkflowExecutionRequest{
		DomainUUID:                t.GetDomainUUID(),
		CancelRequest:             ToRequestCancelWorkflowExecutionRequest(t.CancelRequest),
		ExternalInitiatedEventID:  t.ExternalInitiatedEventId,
		ExternalWorkflowExecution: ToWorkflowExecution(t.ExternalWorkflowExecution),
		ChildWorkflowOnly:         t.GetChildWorkflowOnly(),
	}
}

// FromHistoryResetStickyTaskListRequest converts internal ResetStickyTaskListRequest type to thrift
func FromHistoryResetStickyTaskListRequest(t *types.HistoryResetStickyTaskListRequest) *history.ResetStickyTaskListRequest {
	if t == nil {
		return nil
	}
	return &history.ResetStickyTaskListRequest{
		DomainUUID: &t.DomainUUID,
		Execution:  FromWorkflowExecution(t.Execution),
	}
}

// ToHistoryResetStickyTaskListRequest converts thrift ResetStickyTaskListRequest type to internal
func ToHistoryResetStickyTaskListRequest(t *history.ResetStickyTaskListRequest) *types.HistoryResetStickyTaskListRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryResetStickyTaskListRequest{
		DomainUUID: t.GetDomainUUID(),
		Execution:  ToWorkflowExecution(t.Execution),
	}
}

// FromHistoryResetStickyTaskListResponse converts internal ResetStickyTaskListResponse type to thrift
func FromHistoryResetStickyTaskListResponse(t *types.HistoryResetStickyTaskListResponse) *history.ResetStickyTaskListResponse {
	if t == nil {
		return nil
	}
	return &history.ResetStickyTaskListResponse{}
}

// ToHistoryResetStickyTaskListResponse converts thrift ResetStickyTaskListResponse type to internal
func ToHistoryResetStickyTaskListResponse(t *history.ResetStickyTaskListResponse) *types.HistoryResetStickyTaskListResponse {
	if t == nil {
		return nil
	}
	return &types.HistoryResetStickyTaskListResponse{}
}

// FromHistoryResetWorkflowExecutionRequest converts internal ResetWorkflowExecutionRequest type to thrift
func FromHistoryResetWorkflowExecutionRequest(t *types.HistoryResetWorkflowExecutionRequest) *history.ResetWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.ResetWorkflowExecutionRequest{
		DomainUUID:   &t.DomainUUID,
		ResetRequest: FromResetWorkflowExecutionRequest(t.ResetRequest),
	}
}

// ToHistoryResetWorkflowExecutionRequest converts thrift ResetWorkflowExecutionRequest type to internal
func ToHistoryResetWorkflowExecutionRequest(t *history.ResetWorkflowExecutionRequest) *types.HistoryResetWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryResetWorkflowExecutionRequest{
		DomainUUID:   t.GetDomainUUID(),
		ResetRequest: ToResetWorkflowExecutionRequest(t.ResetRequest),
	}
}

// FromHistoryRespondActivityTaskCanceledRequest converts internal RespondActivityTaskCanceledRequest type to thrift
func FromHistoryRespondActivityTaskCanceledRequest(t *types.HistoryRespondActivityTaskCanceledRequest) *history.RespondActivityTaskCanceledRequest {
	if t == nil {
		return nil
	}
	return &history.RespondActivityTaskCanceledRequest{
		DomainUUID:    &t.DomainUUID,
		CancelRequest: FromRespondActivityTaskCanceledRequest(t.CancelRequest),
	}
}

// ToHistoryRespondActivityTaskCanceledRequest converts thrift RespondActivityTaskCanceledRequest type to internal
func ToHistoryRespondActivityTaskCanceledRequest(t *history.RespondActivityTaskCanceledRequest) *types.HistoryRespondActivityTaskCanceledRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondActivityTaskCanceledRequest{
		DomainUUID:    t.GetDomainUUID(),
		CancelRequest: ToRespondActivityTaskCanceledRequest(t.CancelRequest),
	}
}

// FromHistoryRespondActivityTaskCompletedRequest converts internal RespondActivityTaskCompletedRequest type to thrift
func FromHistoryRespondActivityTaskCompletedRequest(t *types.HistoryRespondActivityTaskCompletedRequest) *history.RespondActivityTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &history.RespondActivityTaskCompletedRequest{
		DomainUUID:      &t.DomainUUID,
		CompleteRequest: FromRespondActivityTaskCompletedRequest(t.CompleteRequest),
	}
}

// ToHistoryRespondActivityTaskCompletedRequest converts thrift RespondActivityTaskCompletedRequest type to internal
func ToHistoryRespondActivityTaskCompletedRequest(t *history.RespondActivityTaskCompletedRequest) *types.HistoryRespondActivityTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondActivityTaskCompletedRequest{
		DomainUUID:      t.GetDomainUUID(),
		CompleteRequest: ToRespondActivityTaskCompletedRequest(t.CompleteRequest),
	}
}

// FromHistoryRespondActivityTaskFailedRequest converts internal RespondActivityTaskFailedRequest type to thrift
func FromHistoryRespondActivityTaskFailedRequest(t *types.HistoryRespondActivityTaskFailedRequest) *history.RespondActivityTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &history.RespondActivityTaskFailedRequest{
		DomainUUID:    &t.DomainUUID,
		FailedRequest: FromRespondActivityTaskFailedRequest(t.FailedRequest),
	}
}

// ToHistoryRespondActivityTaskFailedRequest converts thrift RespondActivityTaskFailedRequest type to internal
func ToHistoryRespondActivityTaskFailedRequest(t *history.RespondActivityTaskFailedRequest) *types.HistoryRespondActivityTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondActivityTaskFailedRequest{
		DomainUUID:    t.GetDomainUUID(),
		FailedRequest: ToRespondActivityTaskFailedRequest(t.FailedRequest),
	}
}

// FromHistoryRespondDecisionTaskCompletedRequest converts internal RespondDecisionTaskCompletedRequest type to thrift
func FromHistoryRespondDecisionTaskCompletedRequest(t *types.HistoryRespondDecisionTaskCompletedRequest) *history.RespondDecisionTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &history.RespondDecisionTaskCompletedRequest{
		DomainUUID:      &t.DomainUUID,
		CompleteRequest: FromRespondDecisionTaskCompletedRequest(t.CompleteRequest),
	}
}

// ToHistoryRespondDecisionTaskCompletedRequest converts thrift RespondDecisionTaskCompletedRequest type to internal
func ToHistoryRespondDecisionTaskCompletedRequest(t *history.RespondDecisionTaskCompletedRequest) *types.HistoryRespondDecisionTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondDecisionTaskCompletedRequest{
		DomainUUID:      t.GetDomainUUID(),
		CompleteRequest: ToRespondDecisionTaskCompletedRequest(t.CompleteRequest),
	}
}

// FromHistoryRespondDecisionTaskCompletedResponse converts internal RespondDecisionTaskCompletedResponse type to thrift
func FromHistoryRespondDecisionTaskCompletedResponse(t *types.HistoryRespondDecisionTaskCompletedResponse) *history.RespondDecisionTaskCompletedResponse {
	if t == nil {
		return nil
	}
	return &history.RespondDecisionTaskCompletedResponse{
		StartedResponse:             FromRecordDecisionTaskStartedResponse(t.StartedResponse),
		ActivitiesToDispatchLocally: FromActivityLocalDispatchInfoMap(t.ActivitiesToDispatchLocally),
	}
}

// ToHistoryRespondDecisionTaskCompletedResponse converts thrift RespondDecisionTaskCompletedResponse type to internal
func ToHistoryRespondDecisionTaskCompletedResponse(t *history.RespondDecisionTaskCompletedResponse) *types.HistoryRespondDecisionTaskCompletedResponse {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondDecisionTaskCompletedResponse{
		StartedResponse:             ToRecordDecisionTaskStartedResponse(t.StartedResponse),
		ActivitiesToDispatchLocally: ToActivityLocalDispatchInfoMap(t.ActivitiesToDispatchLocally),
	}
}

// FromHistoryRespondDecisionTaskFailedRequest converts internal RespondDecisionTaskFailedRequest type to thrift
func FromHistoryRespondDecisionTaskFailedRequest(t *types.HistoryRespondDecisionTaskFailedRequest) *history.RespondDecisionTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &history.RespondDecisionTaskFailedRequest{
		DomainUUID:    &t.DomainUUID,
		FailedRequest: FromRespondDecisionTaskFailedRequest(t.FailedRequest),
	}
}

// ToHistoryRespondDecisionTaskFailedRequest converts thrift RespondDecisionTaskFailedRequest type to internal
func ToHistoryRespondDecisionTaskFailedRequest(t *history.RespondDecisionTaskFailedRequest) *types.HistoryRespondDecisionTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondDecisionTaskFailedRequest{
		DomainUUID:    t.GetDomainUUID(),
		FailedRequest: ToRespondDecisionTaskFailedRequest(t.FailedRequest),
	}
}

// FromScheduleDecisionTaskRequest converts internal ScheduleDecisionTaskRequest type to thrift
func FromScheduleDecisionTaskRequest(t *types.ScheduleDecisionTaskRequest) *history.ScheduleDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &history.ScheduleDecisionTaskRequest{
		DomainUUID:        &t.DomainUUID,
		WorkflowExecution: FromWorkflowExecution(t.WorkflowExecution),
		IsFirstDecision:   &t.IsFirstDecision,
	}
}

// ToScheduleDecisionTaskRequest converts thrift ScheduleDecisionTaskRequest type to internal
func ToScheduleDecisionTaskRequest(t *history.ScheduleDecisionTaskRequest) *types.ScheduleDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &types.ScheduleDecisionTaskRequest{
		DomainUUID:        t.GetDomainUUID(),
		WorkflowExecution: ToWorkflowExecution(t.WorkflowExecution),
		IsFirstDecision:   t.GetIsFirstDecision(),
	}
}

// FromShardOwnershipLostError converts internal ShardOwnershipLostError type to thrift
func FromShardOwnershipLostError(t *types.ShardOwnershipLostError) *history.ShardOwnershipLostError {
	if t == nil {
		return nil
	}
	return &history.ShardOwnershipLostError{
		Message: &t.Message,
		Owner:   &t.Owner,
	}
}

// ToShardOwnershipLostError converts thrift ShardOwnershipLostError type to internal
func ToShardOwnershipLostError(t *history.ShardOwnershipLostError) *types.ShardOwnershipLostError {
	if t == nil {
		return nil
	}
	return &types.ShardOwnershipLostError{
		Message: t.GetMessage(),
		Owner:   t.GetOwner(),
	}
}

// FromHistorySignalWithStartWorkflowExecutionRequest converts internal SignalWithStartWorkflowExecutionRequest type to thrift
func FromHistorySignalWithStartWorkflowExecutionRequest(t *types.HistorySignalWithStartWorkflowExecutionRequest) *history.SignalWithStartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.SignalWithStartWorkflowExecutionRequest{
		DomainUUID:             &t.DomainUUID,
		SignalWithStartRequest: FromSignalWithStartWorkflowExecutionRequest(t.SignalWithStartRequest),
	}
}

// ToHistorySignalWithStartWorkflowExecutionRequest converts thrift SignalWithStartWorkflowExecutionRequest type to internal
func ToHistorySignalWithStartWorkflowExecutionRequest(t *history.SignalWithStartWorkflowExecutionRequest) *types.HistorySignalWithStartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistorySignalWithStartWorkflowExecutionRequest{
		DomainUUID:             t.GetDomainUUID(),
		SignalWithStartRequest: ToSignalWithStartWorkflowExecutionRequest(t.SignalWithStartRequest),
	}
}

// FromHistorySignalWorkflowExecutionRequest converts internal SignalWorkflowExecutionRequest type to thrift
func FromHistorySignalWorkflowExecutionRequest(t *types.HistorySignalWorkflowExecutionRequest) *history.SignalWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.SignalWorkflowExecutionRequest{
		DomainUUID:                &t.DomainUUID,
		SignalRequest:             FromSignalWorkflowExecutionRequest(t.SignalRequest),
		ExternalWorkflowExecution: FromWorkflowExecution(t.ExternalWorkflowExecution),
		ChildWorkflowOnly:         &t.ChildWorkflowOnly,
	}
}

// ToHistorySignalWorkflowExecutionRequest converts thrift SignalWorkflowExecutionRequest type to internal
func ToHistorySignalWorkflowExecutionRequest(t *history.SignalWorkflowExecutionRequest) *types.HistorySignalWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistorySignalWorkflowExecutionRequest{
		DomainUUID:                t.GetDomainUUID(),
		SignalRequest:             ToSignalWorkflowExecutionRequest(t.SignalRequest),
		ExternalWorkflowExecution: ToWorkflowExecution(t.ExternalWorkflowExecution),
		ChildWorkflowOnly:         t.GetChildWorkflowOnly(),
	}
}

// FromHistoryStartWorkflowExecutionRequest converts internal StartWorkflowExecutionRequest type to thrift
func FromHistoryStartWorkflowExecutionRequest(t *types.HistoryStartWorkflowExecutionRequest) *history.StartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.StartWorkflowExecutionRequest{
		DomainUUID:                      &t.DomainUUID,
		StartRequest:                    FromStartWorkflowExecutionRequest(t.StartRequest),
		ParentExecutionInfo:             FromParentExecutionInfo(t.ParentExecutionInfo),
		Attempt:                         &t.Attempt,
		ExpirationTimestamp:             t.ExpirationTimestamp,
		ContinueAsNewInitiator:          FromContinueAsNewInitiator(t.ContinueAsNewInitiator),
		ContinuedFailureReason:          t.ContinuedFailureReason,
		ContinuedFailureDetails:         t.ContinuedFailureDetails,
		LastCompletionResult:            t.LastCompletionResult,
		FirstDecisionTaskBackoffSeconds: t.FirstDecisionTaskBackoffSeconds,
	}
}

// ToHistoryStartWorkflowExecutionRequest converts thrift StartWorkflowExecutionRequest type to internal
func ToHistoryStartWorkflowExecutionRequest(t *history.StartWorkflowExecutionRequest) *types.HistoryStartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryStartWorkflowExecutionRequest{
		DomainUUID:                      t.GetDomainUUID(),
		StartRequest:                    ToStartWorkflowExecutionRequest(t.StartRequest),
		ParentExecutionInfo:             ToParentExecutionInfo(t.ParentExecutionInfo),
		Attempt:                         t.GetAttempt(),
		ExpirationTimestamp:             t.ExpirationTimestamp,
		ContinueAsNewInitiator:          ToContinueAsNewInitiator(t.ContinueAsNewInitiator),
		ContinuedFailureReason:          t.ContinuedFailureReason,
		ContinuedFailureDetails:         t.ContinuedFailureDetails,
		LastCompletionResult:            t.LastCompletionResult,
		FirstDecisionTaskBackoffSeconds: t.FirstDecisionTaskBackoffSeconds,
	}
}

// FromSyncActivityRequest converts internal SyncActivityRequest type to thrift
func FromSyncActivityRequest(t *types.SyncActivityRequest) *history.SyncActivityRequest {
	if t == nil {
		return nil
	}
	return &history.SyncActivityRequest{
		DomainId:           &t.DomainID,
		WorkflowId:         &t.WorkflowID,
		RunId:              &t.RunID,
		Version:            &t.Version,
		ScheduledId:        &t.ScheduledID,
		ScheduledTime:      t.ScheduledTime,
		StartedId:          &t.StartedID,
		StartedTime:        t.StartedTime,
		LastHeartbeatTime:  t.LastHeartbeatTime,
		Details:            t.Details,
		Attempt:            &t.Attempt,
		LastFailureReason:  t.LastFailureReason,
		LastWorkerIdentity: &t.LastWorkerIdentity,
		LastFailureDetails: t.LastFailureDetails,
		VersionHistory:     FromVersionHistory(t.VersionHistory),
	}
}

// ToSyncActivityRequest converts thrift SyncActivityRequest type to internal
func ToSyncActivityRequest(t *history.SyncActivityRequest) *types.SyncActivityRequest {
	if t == nil {
		return nil
	}
	return &types.SyncActivityRequest{
		DomainID:           t.GetDomainId(),
		WorkflowID:         t.GetWorkflowId(),
		RunID:              t.GetRunId(),
		Version:            t.GetVersion(),
		ScheduledID:        t.GetScheduledId(),
		ScheduledTime:      t.ScheduledTime,
		StartedID:          t.GetStartedId(),
		StartedTime:        t.StartedTime,
		LastHeartbeatTime:  t.LastHeartbeatTime,
		Details:            t.Details,
		Attempt:            t.GetAttempt(),
		LastFailureReason:  t.LastFailureReason,
		LastWorkerIdentity: t.GetLastWorkerIdentity(),
		LastFailureDetails: t.LastFailureDetails,
		VersionHistory:     ToVersionHistory(t.VersionHistory),
	}
}

// FromSyncShardStatusRequest converts internal SyncShardStatusRequest type to thrift
func FromSyncShardStatusRequest(t *types.SyncShardStatusRequest) *history.SyncShardStatusRequest {
	if t == nil {
		return nil
	}
	return &history.SyncShardStatusRequest{
		SourceCluster: &t.SourceCluster,
		ShardId:       &t.ShardID,
		Timestamp:     t.Timestamp,
	}
}

// ToSyncShardStatusRequest converts thrift SyncShardStatusRequest type to internal
func ToSyncShardStatusRequest(t *history.SyncShardStatusRequest) *types.SyncShardStatusRequest {
	if t == nil {
		return nil
	}
	return &types.SyncShardStatusRequest{
		SourceCluster: t.GetSourceCluster(),
		ShardID:       t.GetShardId(),
		Timestamp:     t.Timestamp,
	}
}

// FromHistoryTerminateWorkflowExecutionRequest converts internal TerminateWorkflowExecutionRequest type to thrift
func FromHistoryTerminateWorkflowExecutionRequest(t *types.HistoryTerminateWorkflowExecutionRequest) *history.TerminateWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.TerminateWorkflowExecutionRequest{
		DomainUUID:       &t.DomainUUID,
		TerminateRequest: FromTerminateWorkflowExecutionRequest(t.TerminateRequest),
	}
}

// ToHistoryTerminateWorkflowExecutionRequest converts thrift TerminateWorkflowExecutionRequest type to internal
func ToHistoryTerminateWorkflowExecutionRequest(t *history.TerminateWorkflowExecutionRequest) *types.HistoryTerminateWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryTerminateWorkflowExecutionRequest{
		DomainUUID:       t.GetDomainUUID(),
		TerminateRequest: ToTerminateWorkflowExecutionRequest(t.TerminateRequest),
	}
}

// FromFailoverMarkerTokenArray converts internal FailoverMarkerToken type array to thrift
func FromFailoverMarkerTokenArray(t []*types.FailoverMarkerToken) []*history.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	v := make([]*history.FailoverMarkerToken, len(t))
	for i := range t {
		v[i] = FromFailoverMarkerToken(t[i])
	}
	return v
}

// ToFailoverMarkerTokenArray converts thrift FailoverMarkerToken type array to internal
func ToFailoverMarkerTokenArray(t []*history.FailoverMarkerToken) []*types.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	v := make([]*types.FailoverMarkerToken, len(t))
	for i := range t {
		v[i] = ToFailoverMarkerToken(t[i])
	}
	return v
}

// FromProcessingQueueStateArray converts internal ProcessingQueueState type array to thrift
func FromProcessingQueueStateArray(t []*types.ProcessingQueueState) []*history.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make([]*history.ProcessingQueueState, len(t))
	for i := range t {
		v[i] = FromProcessingQueueState(t[i])
	}
	return v
}

// ToProcessingQueueStateArray converts thrift ProcessingQueueState type array to internal
func ToProcessingQueueStateArray(t []*history.ProcessingQueueState) []*types.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make([]*types.ProcessingQueueState, len(t))
	for i := range t {
		v[i] = ToProcessingQueueState(t[i])
	}
	return v
}

// FromProcessingQueueStateArrayMap converts internal ProcessingQueueState array map to thrift
func FromProcessingQueueStateArrayMap(t map[string][]*types.ProcessingQueueState) map[string][]*history.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make(map[string][]*history.ProcessingQueueState, len(t))
	for key := range t {
		v[key] = FromProcessingQueueStateArray(t[key])
	}
	return v
}

// ToProcessingQueueStateArrayMap converts thrift ProcessingQueueState array map to internal
func ToProcessingQueueStateArrayMap(t map[string][]*history.ProcessingQueueState) map[string][]*types.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make(map[string][]*types.ProcessingQueueState, len(t))
	for key := range t {
		v[key] = ToProcessingQueueStateArray(t[key])
	}
	return v
}
