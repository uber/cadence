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

package proto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/testdata"
)

func TestHistoryCloseShardRequest(t *testing.T) {
	for _, item := range []*types.CloseShardRequest{nil, {}, &testdata.HistoryCloseShardRequest} {
		assert.Equal(t, item, ToHistoryCloseShardRequest(FromHistoryCloseShardRequest(item)))
	}
}
func TestHistoryDescribeHistoryHostRequest(t *testing.T) {
	for _, item := range []*types.DescribeHistoryHostRequest{nil, {}, &testdata.HistoryDescribeHistoryHostRequest} {
		assert.Equal(t, item, ToHistoryDescribeHistoryHostRequest(FromHistoryDescribeHistoryHostRequest(item)))
	}
}
func TestHistoryDescribeHistoryHostResponse(t *testing.T) {
	for _, item := range []*types.DescribeHistoryHostResponse{nil, {}, &testdata.HistoryDescribeHistoryHostResponse} {
		assert.Equal(t, item, ToHistoryDescribeHistoryHostResponse(FromHistoryDescribeHistoryHostResponse(item)))
	}
}
func TestHistoryDescribeMutableStateRequest(t *testing.T) {
	for _, item := range []*types.DescribeMutableStateRequest{nil, {}, &testdata.HistoryDescribeMutableStateRequest} {
		assert.Equal(t, item, ToHistoryDescribeMutableStateRequest(FromHistoryDescribeMutableStateRequest(item)))
	}
}
func TestHistoryDescribeMutableStateResponse(t *testing.T) {
	for _, item := range []*types.DescribeMutableStateResponse{nil, {}, &testdata.HistoryDescribeMutableStateResponse} {
		assert.Equal(t, item, ToHistoryDescribeMutableStateResponse(FromHistoryDescribeMutableStateResponse(item)))
	}
}
func TestHistoryDescribeQueueRequest(t *testing.T) {
	for _, item := range []*types.DescribeQueueRequest{nil, {}, &testdata.HistoryDescribeQueueRequest} {
		assert.Equal(t, item, ToHistoryDescribeQueueRequest(FromHistoryDescribeQueueRequest(item)))
	}
}
func TestHistoryDescribeQueueResponse(t *testing.T) {
	for _, item := range []*types.DescribeQueueResponse{nil, {}, &testdata.HistoryDescribeQueueResponse} {
		assert.Equal(t, item, ToHistoryDescribeQueueResponse(FromHistoryDescribeQueueResponse(item)))
	}
}
func TestHistoryDescribeWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*types.HistoryDescribeWorkflowExecutionRequest{nil, {}, &testdata.HistoryDescribeWorkflowExecutionRequest} {
		assert.Equal(t, item, ToHistoryDescribeWorkflowExecutionRequest(FromHistoryDescribeWorkflowExecutionRequest(item)))
	}
}
func TestHistoryDescribeWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*types.DescribeWorkflowExecutionResponse{nil, {}, &testdata.HistoryDescribeWorkflowExecutionResponse} {
		assert.Equal(t, item, ToHistoryDescribeWorkflowExecutionResponse(FromHistoryDescribeWorkflowExecutionResponse(item)))
	}
}
func TestHistoryGetDLQReplicationMessagesRequest(t *testing.T) {
	for _, item := range []*types.GetDLQReplicationMessagesRequest{nil, {}, &testdata.HistoryGetDLQReplicationMessagesRequest} {
		assert.Equal(t, item, ToHistoryGetDLQReplicationMessagesRequest(FromHistoryGetDLQReplicationMessagesRequest(item)))
	}
}
func TestHistoryGetDLQReplicationMessagesResponse(t *testing.T) {
	for _, item := range []*types.GetDLQReplicationMessagesResponse{nil, {}, &testdata.HistoryGetDLQReplicationMessagesResponse} {
		assert.Equal(t, item, ToHistoryGetDLQReplicationMessagesResponse(FromHistoryGetDLQReplicationMessagesResponse(item)))
	}
}
func TestHistoryGetMutableStateRequest(t *testing.T) {
	for _, item := range []*types.GetMutableStateRequest{nil, {}, &testdata.HistoryGetMutableStateRequest} {
		assert.Equal(t, item, ToHistoryGetMutableStateRequest(FromHistoryGetMutableStateRequest(item)))
	}
}
func TestHistoryGetMutableStateResponse(t *testing.T) {
	for _, item := range []*types.GetMutableStateResponse{nil, &testdata.HistoryGetMutableStateResponse} {
		assert.Equal(t, item, ToHistoryGetMutableStateResponse(FromHistoryGetMutableStateResponse(item)))
	}
}
func TestHistoryGetReplicationMessagesRequest(t *testing.T) {
	for _, item := range []*types.GetReplicationMessagesRequest{nil, {}, &testdata.HistoryGetReplicationMessagesRequest} {
		assert.Equal(t, item, ToHistoryGetReplicationMessagesRequest(FromHistoryGetReplicationMessagesRequest(item)))
	}
}
func TestHistoryGetReplicationMessagesResponse(t *testing.T) {
	for _, item := range []*types.GetReplicationMessagesResponse{nil, {}, &testdata.HistoryGetReplicationMessagesResponse} {
		assert.Equal(t, item, ToHistoryGetReplicationMessagesResponse(FromHistoryGetReplicationMessagesResponse(item)))
	}
}
func TestHistoryMergeDLQMessagesRequest(t *testing.T) {
	for _, item := range []*types.MergeDLQMessagesRequest{nil, {}, &testdata.HistoryMergeDLQMessagesRequest} {
		assert.Equal(t, item, ToHistoryMergeDLQMessagesRequest(FromHistoryMergeDLQMessagesRequest(item)))
	}
}
func TestHistoryMergeDLQMessagesResponse(t *testing.T) {
	for _, item := range []*types.MergeDLQMessagesResponse{nil, {}, &testdata.HistoryMergeDLQMessagesResponse} {
		assert.Equal(t, item, ToHistoryMergeDLQMessagesResponse(FromHistoryMergeDLQMessagesResponse(item)))
	}
}
func TestHistoryNotifyFailoverMarkersRequest(t *testing.T) {
	for _, item := range []*types.NotifyFailoverMarkersRequest{nil, {}, &testdata.HistoryNotifyFailoverMarkersRequest} {
		assert.Equal(t, item, ToHistoryNotifyFailoverMarkersRequest(FromHistoryNotifyFailoverMarkersRequest(item)))
	}
}
func TestHistoryPollMutableStateRequest(t *testing.T) {
	for _, item := range []*types.PollMutableStateRequest{nil, {}, &testdata.HistoryPollMutableStateRequest} {
		assert.Equal(t, item, ToHistoryPollMutableStateRequest(FromHistoryPollMutableStateRequest(item)))
	}
}
func TestHistoryPollMutableStateResponse(t *testing.T) {
	for _, item := range []*types.PollMutableStateResponse{nil, &testdata.HistoryPollMutableStateResponse} {
		assert.Equal(t, item, ToHistoryPollMutableStateResponse(FromHistoryPollMutableStateResponse(item)))
	}
}
func TestHistoryPurgeDLQMessagesRequest(t *testing.T) {
	for _, item := range []*types.PurgeDLQMessagesRequest{nil, {}, &testdata.HistoryPurgeDLQMessagesRequest} {
		assert.Equal(t, item, ToHistoryPurgeDLQMessagesRequest(FromHistoryPurgeDLQMessagesRequest(item)))
	}
}
func TestHistoryQueryWorkflowRequest(t *testing.T) {
	for _, item := range []*types.HistoryQueryWorkflowRequest{nil, {}, &testdata.HistoryQueryWorkflowRequest} {
		assert.Equal(t, item, ToHistoryQueryWorkflowRequest(FromHistoryQueryWorkflowRequest(item)))
	}
}
func TestHistoryQueryWorkflowResponse(t *testing.T) {
	for _, item := range []*types.HistoryQueryWorkflowResponse{nil, &testdata.HistoryQueryWorkflowResponse} {
		assert.Equal(t, item, ToHistoryQueryWorkflowResponse(FromHistoryQueryWorkflowResponse(item)))
	}
	assert.Nil(t, FromHistoryQueryWorkflowResponse(&types.HistoryQueryWorkflowResponse{}))
}
func TestHistoryReadDLQMessagesRequest(t *testing.T) {
	for _, item := range []*types.ReadDLQMessagesRequest{nil, {}, &testdata.HistoryReadDLQMessagesRequest} {
		assert.Equal(t, item, ToHistoryReadDLQMessagesRequest(FromHistoryReadDLQMessagesRequest(item)))
	}
}
func TestHistoryReadDLQMessagesResponse(t *testing.T) {
	for _, item := range []*types.ReadDLQMessagesResponse{nil, {}, &testdata.HistoryReadDLQMessagesResponse} {
		assert.Equal(t, item, ToHistoryReadDLQMessagesResponse(FromHistoryReadDLQMessagesResponse(item)))
	}
}
func TestHistoryReapplyEventsRequest(t *testing.T) {
	for _, item := range []*types.HistoryReapplyEventsRequest{nil, &testdata.HistoryReapplyEventsRequest} {
		assert.Equal(t, item, ToHistoryReapplyEventsRequest(FromHistoryReapplyEventsRequest(item)))
	}
	assert.Nil(t, FromHistoryReapplyEventsRequest(&types.HistoryReapplyEventsRequest{}))
}
func TestHistoryRecordActivityTaskHeartbeatRequest(t *testing.T) {
	for _, item := range []*types.HistoryRecordActivityTaskHeartbeatRequest{nil, {}, &testdata.HistoryRecordActivityTaskHeartbeatRequest} {
		assert.Equal(t, item, ToHistoryRecordActivityTaskHeartbeatRequest(FromHistoryRecordActivityTaskHeartbeatRequest(item)))
	}
}
func TestHistoryRecordActivityTaskHeartbeatResponse(t *testing.T) {
	for _, item := range []*types.RecordActivityTaskHeartbeatResponse{nil, {}, &testdata.HistoryRecordActivityTaskHeartbeatResponse} {
		assert.Equal(t, item, ToHistoryRecordActivityTaskHeartbeatResponse(FromHistoryRecordActivityTaskHeartbeatResponse(item)))
	}
}
func TestHistoryRecordActivityTaskStartedRequest(t *testing.T) {
	for _, item := range []*types.RecordActivityTaskStartedRequest{nil, {}, &testdata.HistoryRecordActivityTaskStartedRequest} {
		assert.Equal(t, item, ToHistoryRecordActivityTaskStartedRequest(FromHistoryRecordActivityTaskStartedRequest(item)))
	}
}
func TestHistoryRecordActivityTaskStartedResponse(t *testing.T) {
	for _, item := range []*types.RecordActivityTaskStartedResponse{nil, {}, &testdata.HistoryRecordActivityTaskStartedResponse} {
		assert.Equal(t, item, ToHistoryRecordActivityTaskStartedResponse(FromHistoryRecordActivityTaskStartedResponse(item)))
	}
}
func TestHistoryRecordChildExecutionCompletedRequest(t *testing.T) {
	for _, item := range []*types.RecordChildExecutionCompletedRequest{nil, {}, &testdata.HistoryRecordChildExecutionCompletedRequest} {
		assert.Equal(t, item, ToHistoryRecordChildExecutionCompletedRequest(FromHistoryRecordChildExecutionCompletedRequest(item)))
	}
}
func TestHistoryRecordDecisionTaskStartedRequest(t *testing.T) {
	for _, item := range []*types.RecordDecisionTaskStartedRequest{nil, {}, &testdata.HistoryRecordDecisionTaskStartedRequest} {
		assert.Equal(t, item, ToHistoryRecordDecisionTaskStartedRequest(FromHistoryRecordDecisionTaskStartedRequest(item)))
	}
}
func TestHistoryRecordDecisionTaskStartedResponse(t *testing.T) {
	for _, item := range []*types.RecordDecisionTaskStartedResponse{nil, {}, &testdata.HistoryRecordDecisionTaskStartedResponse} {
		assert.Equal(t, item, ToHistoryRecordDecisionTaskStartedResponse(FromHistoryRecordDecisionTaskStartedResponse(item)))
	}
}
func TestHistoryRefreshWorkflowTasksRequest(t *testing.T) {
	for _, item := range []*types.HistoryRefreshWorkflowTasksRequest{nil, &testdata.HistoryRefreshWorkflowTasksRequest} {
		assert.Equal(t, item, ToHistoryRefreshWorkflowTasksRequest(FromHistoryRefreshWorkflowTasksRequest(item)))
	}
	assert.Nil(t, FromHistoryRefreshWorkflowTasksRequest(&types.HistoryRefreshWorkflowTasksRequest{}))
}
func TestHistoryRemoveSignalMutableStateRequest(t *testing.T) {
	for _, item := range []*types.RemoveSignalMutableStateRequest{nil, {}, &testdata.HistoryRemoveSignalMutableStateRequest} {
		assert.Equal(t, item, ToHistoryRemoveSignalMutableStateRequest(FromHistoryRemoveSignalMutableStateRequest(item)))
	}
}
func TestHistoryRemoveTaskRequest(t *testing.T) {
	for _, item := range []*types.RemoveTaskRequest{nil, {}, &testdata.HistoryRemoveTaskRequest} {
		assert.Equal(t, item, ToHistoryRemoveTaskRequest(FromHistoryRemoveTaskRequest(item)))
	}
}
func TestHistoryReplicateEventsV2Request(t *testing.T) {
	for _, item := range []*types.ReplicateEventsV2Request{nil, {}, &testdata.HistoryReplicateEventsV2Request} {
		assert.Equal(t, item, ToHistoryReplicateEventsV2Request(FromHistoryReplicateEventsV2Request(item)))
	}
}
func TestHistoryRequestCancelWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*types.HistoryRequestCancelWorkflowExecutionRequest{nil, {}, &testdata.HistoryRequestCancelWorkflowExecutionRequest} {
		assert.Equal(t, item, ToHistoryRequestCancelWorkflowExecutionRequest(FromHistoryRequestCancelWorkflowExecutionRequest(item)))
	}
}
func TestHistoryResetQueueRequest(t *testing.T) {
	for _, item := range []*types.ResetQueueRequest{nil, {}, &testdata.HistoryResetQueueRequest} {
		assert.Equal(t, item, ToHistoryResetQueueRequest(FromHistoryResetQueueRequest(item)))
	}
}
func TestHistoryResetStickyTaskListRequest(t *testing.T) {
	for _, item := range []*types.HistoryResetStickyTaskListRequest{nil, {}, &testdata.HistoryResetStickyTaskListRequest} {
		assert.Equal(t, item, ToHistoryResetStickyTaskListRequest(FromHistoryResetStickyTaskListRequest(item)))
	}
}
func TestHistoryResetWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*types.HistoryResetWorkflowExecutionRequest{nil, {}, &testdata.HistoryResetWorkflowExecutionRequest} {
		assert.Equal(t, item, ToHistoryResetWorkflowExecutionRequest(FromHistoryResetWorkflowExecutionRequest(item)))
	}
}
func TestHistoryResetWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*types.ResetWorkflowExecutionResponse{nil, {}, &testdata.HistoryResetWorkflowExecutionResponse} {
		assert.Equal(t, item, ToHistoryResetWorkflowExecutionResponse(FromHistoryResetWorkflowExecutionResponse(item)))
	}
}
func TestHistoryRespondActivityTaskCanceledRequest(t *testing.T) {
	for _, item := range []*types.HistoryRespondActivityTaskCanceledRequest{nil, {}, &testdata.HistoryRespondActivityTaskCanceledRequest} {
		assert.Equal(t, item, ToHistoryRespondActivityTaskCanceledRequest(FromHistoryRespondActivityTaskCanceledRequest(item)))
	}
}
func TestHistoryRespondActivityTaskCompletedRequest(t *testing.T) {
	for _, item := range []*types.HistoryRespondActivityTaskCompletedRequest{nil, {}, &testdata.HistoryRespondActivityTaskCompletedRequest} {
		assert.Equal(t, item, ToHistoryRespondActivityTaskCompletedRequest(FromHistoryRespondActivityTaskCompletedRequest(item)))
	}
}
func TestHistoryRespondActivityTaskFailedRequest(t *testing.T) {
	for _, item := range []*types.HistoryRespondActivityTaskFailedRequest{nil, {}, &testdata.HistoryRespondActivityTaskFailedRequest} {
		assert.Equal(t, item, ToHistoryRespondActivityTaskFailedRequest(FromHistoryRespondActivityTaskFailedRequest(item)))
	}
}
func TestHistoryRespondDecisionTaskCompletedRequest(t *testing.T) {
	for _, item := range []*types.HistoryRespondDecisionTaskCompletedRequest{nil, {}, &testdata.HistoryRespondDecisionTaskCompletedRequest} {
		assert.Equal(t, item, ToHistoryRespondDecisionTaskCompletedRequest(FromHistoryRespondDecisionTaskCompletedRequest(item)))
	}
}
func TestHistoryRespondDecisionTaskCompletedResponse(t *testing.T) {
	for _, item := range []*types.HistoryRespondDecisionTaskCompletedResponse{nil, {}, &testdata.HistoryRespondDecisionTaskCompletedResponse} {
		assert.Equal(t, item, ToHistoryRespondDecisionTaskCompletedResponse(FromHistoryRespondDecisionTaskCompletedResponse(item)))
	}
}
func TestHistoryRespondDecisionTaskFailedRequest(t *testing.T) {
	for _, item := range []*types.HistoryRespondDecisionTaskFailedRequest{nil, {}, &testdata.HistoryRespondDecisionTaskFailedRequest} {
		assert.Equal(t, item, ToHistoryRespondDecisionTaskFailedRequest(FromHistoryRespondDecisionTaskFailedRequest(item)))
	}
}
func TestHistoryScheduleDecisionTaskRequest(t *testing.T) {
	for _, item := range []*types.ScheduleDecisionTaskRequest{nil, {}, &testdata.HistoryScheduleDecisionTaskRequest} {
		assert.Equal(t, item, ToHistoryScheduleDecisionTaskRequest(FromHistoryScheduleDecisionTaskRequest(item)))
	}
}
func TestHistorySignalWithStartWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*types.HistorySignalWithStartWorkflowExecutionRequest{nil, {}, &testdata.HistorySignalWithStartWorkflowExecutionRequest} {
		assert.Equal(t, item, ToHistorySignalWithStartWorkflowExecutionRequest(FromHistorySignalWithStartWorkflowExecutionRequest(item)))
	}
}
func TestHistorySignalWithStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*types.StartWorkflowExecutionResponse{nil, {}, &testdata.HistorySignalWithStartWorkflowExecutionResponse} {
		assert.Equal(t, item, ToHistorySignalWithStartWorkflowExecutionResponse(FromHistorySignalWithStartWorkflowExecutionResponse(item)))
	}
}
func TestHistorySignalWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*types.HistorySignalWorkflowExecutionRequest{nil, {}, &testdata.HistorySignalWorkflowExecutionRequest} {
		assert.Equal(t, item, ToHistorySignalWorkflowExecutionRequest(FromHistorySignalWorkflowExecutionRequest(item)))
	}
}
func TestHistoryStartWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*types.HistoryStartWorkflowExecutionRequest{nil, {}, &testdata.HistoryStartWorkflowExecutionRequest} {
		assert.Equal(t, item, ToHistoryStartWorkflowExecutionRequest(FromHistoryStartWorkflowExecutionRequest(item)))
	}
}
func TestHistoryStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*types.StartWorkflowExecutionResponse{nil, {}, &testdata.HistoryStartWorkflowExecutionResponse} {
		assert.Equal(t, item, ToHistoryStartWorkflowExecutionResponse(FromHistoryStartWorkflowExecutionResponse(item)))
	}
}
func TestHistorySyncActivityRequest(t *testing.T) {
	for _, item := range []*types.SyncActivityRequest{nil, {}, &testdata.HistorySyncActivityRequest} {
		assert.Equal(t, item, ToHistorySyncActivityRequest(FromHistorySyncActivityRequest(item)))
	}
}
func TestHistorySyncShardStatusRequest(t *testing.T) {
	for _, item := range []*types.SyncShardStatusRequest{nil, {}, &testdata.HistorySyncShardStatusRequest} {
		assert.Equal(t, item, ToHistorySyncShardStatusRequest(FromHistorySyncShardStatusRequest(item)))
	}
}
func TestHistoryTerminateWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*types.HistoryTerminateWorkflowExecutionRequest{nil, {}, &testdata.HistoryTerminateWorkflowExecutionRequest} {
		assert.Equal(t, item, ToHistoryTerminateWorkflowExecutionRequest(FromHistoryTerminateWorkflowExecutionRequest(item)))
	}
}
