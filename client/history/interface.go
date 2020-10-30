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

package history

import (
	"context"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
)

// Client is the interface exposed by history service client
type Client interface {
	CloseShard(context.Context, *shared.CloseShardRequest, ...yarpc.CallOption) error
	DescribeHistoryHost(context.Context, *shared.DescribeHistoryHostRequest, ...yarpc.CallOption) (*shared.DescribeHistoryHostResponse, error)
	DescribeMutableState(context.Context, *history.DescribeMutableStateRequest, ...yarpc.CallOption) (*history.DescribeMutableStateResponse, error)
	DescribeQueue(context.Context, *shared.DescribeQueueRequest, ...yarpc.CallOption) (*shared.DescribeQueueResponse, error)
	DescribeWorkflowExecution(context.Context, *history.DescribeWorkflowExecutionRequest, ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error)
	GetDLQReplicationMessages(context.Context, *replicator.GetDLQReplicationMessagesRequest, ...yarpc.CallOption) (*replicator.GetDLQReplicationMessagesResponse, error)
	GetMutableState(context.Context, *history.GetMutableStateRequest, ...yarpc.CallOption) (*history.GetMutableStateResponse, error)
	GetReplicationMessages(context.Context, *replicator.GetReplicationMessagesRequest, ...yarpc.CallOption) (*replicator.GetReplicationMessagesResponse, error)
	MergeDLQMessages(context.Context, *replicator.MergeDLQMessagesRequest, ...yarpc.CallOption) (*replicator.MergeDLQMessagesResponse, error)
	NotifyFailoverMarkers(context.Context, *history.NotifyFailoverMarkersRequest, ...yarpc.CallOption) error
	PollMutableState(context.Context, *history.PollMutableStateRequest, ...yarpc.CallOption) (*history.PollMutableStateResponse, error)
	PurgeDLQMessages(context.Context, *replicator.PurgeDLQMessagesRequest, ...yarpc.CallOption) error
	QueryWorkflow(context.Context, *history.QueryWorkflowRequest, ...yarpc.CallOption) (*history.QueryWorkflowResponse, error)
	ReadDLQMessages(context.Context, *replicator.ReadDLQMessagesRequest, ...yarpc.CallOption) (*replicator.ReadDLQMessagesResponse, error)
	ReapplyEvents(context.Context, *history.ReapplyEventsRequest, ...yarpc.CallOption) error
	RecordActivityTaskHeartbeat(context.Context, *history.RecordActivityTaskHeartbeatRequest, ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error)
	RecordActivityTaskStarted(context.Context, *history.RecordActivityTaskStartedRequest, ...yarpc.CallOption) (*history.RecordActivityTaskStartedResponse, error)
	RecordChildExecutionCompleted(context.Context, *history.RecordChildExecutionCompletedRequest, ...yarpc.CallOption) error
	RecordDecisionTaskStarted(context.Context, *history.RecordDecisionTaskStartedRequest, ...yarpc.CallOption) (*history.RecordDecisionTaskStartedResponse, error)
	RefreshWorkflowTasks(context.Context, *history.RefreshWorkflowTasksRequest, ...yarpc.CallOption) error
	RemoveSignalMutableState(context.Context, *history.RemoveSignalMutableStateRequest, ...yarpc.CallOption) error
	RemoveTask(context.Context, *shared.RemoveTaskRequest, ...yarpc.CallOption) error
	ReplicateEventsV2(context.Context, *history.ReplicateEventsV2Request, ...yarpc.CallOption) error
	RequestCancelWorkflowExecution(context.Context, *history.RequestCancelWorkflowExecutionRequest, ...yarpc.CallOption) error
	ResetQueue(context.Context, *shared.ResetQueueRequest, ...yarpc.CallOption) error
	ResetStickyTaskList(context.Context, *history.ResetStickyTaskListRequest, ...yarpc.CallOption) (*history.ResetStickyTaskListResponse, error)
	ResetWorkflowExecution(context.Context, *history.ResetWorkflowExecutionRequest, ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error)
	RespondActivityTaskCanceled(context.Context, *history.RespondActivityTaskCanceledRequest, ...yarpc.CallOption) error
	RespondActivityTaskCompleted(context.Context, *history.RespondActivityTaskCompletedRequest, ...yarpc.CallOption) error
	RespondActivityTaskFailed(context.Context, *history.RespondActivityTaskFailedRequest, ...yarpc.CallOption) error
	RespondDecisionTaskCompleted(context.Context, *history.RespondDecisionTaskCompletedRequest, ...yarpc.CallOption) (*history.RespondDecisionTaskCompletedResponse, error)
	RespondDecisionTaskFailed(context.Context, *history.RespondDecisionTaskFailedRequest, ...yarpc.CallOption) error
	ScheduleDecisionTask(context.Context, *history.ScheduleDecisionTaskRequest, ...yarpc.CallOption) error
	SignalWithStartWorkflowExecution(context.Context, *history.SignalWithStartWorkflowExecutionRequest, ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error)
	SignalWorkflowExecution(context.Context, *history.SignalWorkflowExecutionRequest, ...yarpc.CallOption) error
	StartWorkflowExecution(context.Context, *history.StartWorkflowExecutionRequest, ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error)
	SyncActivity(context.Context, *history.SyncActivityRequest, ...yarpc.CallOption) error
	SyncShardStatus(context.Context, *history.SyncShardStatusRequest, ...yarpc.CallOption) error
	TerminateWorkflowExecution(context.Context, *history.TerminateWorkflowExecutionRequest, ...yarpc.CallOption) error
}
