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
	"sync"
	"time"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
)

var _ Client = (*clientImpl)(nil)

const (
	// DefaultTimeout is the default timeout used to make calls
	DefaultTimeout = time.Second * 30
)

type clientImpl struct {
	numberOfShards  int
	tokenSerializer common.TaskTokenSerializer
	timeout         time.Duration
	clients         common.ClientCache
	logger          log.Logger
}

// NewClient creates a new history service TChannel client
func NewClient(
	numberOfShards int,
	timeout time.Duration,
	clients common.ClientCache,
	logger log.Logger,
) Client {
	return &clientImpl{
		numberOfShards:  numberOfShards,
		tokenSerializer: common.NewJSONTaskTokenSerializer(),
		timeout:         timeout,
		clients:         clients,
		logger:          logger,
	}
}

func (c *clientImpl) StartWorkflowExecution(
	ctx context.Context,
	request *types.HistoryStartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {
	client, err := c.getClientForWorkflowID(*request.StartRequest.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.StartWorkflowExecutionResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.StartWorkflowExecution(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) GetMutableState(
	ctx context.Context,
	request *types.GetMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.GetMutableStateResponse, error) {
	client, err := c.getClientForWorkflowID(*request.Execution.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.GetMutableStateResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.GetMutableState(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) PollMutableState(
	ctx context.Context,
	request *types.PollMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.PollMutableStateResponse, error) {
	client, err := c.getClientForWorkflowID(*request.Execution.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.PollMutableStateResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.PollMutableState(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) DescribeHistoryHost(
	ctx context.Context,
	request *types.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeHistoryHostResponse, error) {

	var err error
	var client Client

	if request.ShardIDForHost != nil {
		client, err = c.getClientForShardID(int(request.GetShardIDForHost()))
	} else if request.ExecutionForHost != nil {
		client, err = c.getClientForWorkflowID(request.ExecutionForHost.GetWorkflowID())
	} else {
		ret, err := c.clients.GetClientForClientKey(request.GetHostAddress())
		if err != nil {
			return nil, err
		}
		client = ret.(Client)
	}
	if err != nil {
		return nil, err
	}

	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.DescribeHistoryHostResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.DescribeHistoryHost(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) RemoveTask(
	ctx context.Context,
	request *types.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) error {
	var err error
	var client Client
	if request.ShardID != nil {
		client, err = c.getClientForShardID(int(request.GetShardID()))
		if err != nil {
			return err
		}
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		err = client.RemoveTask(ctx, request, opts...)
		return err
	}

	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) CloseShard(
	ctx context.Context,
	request *types.CloseShardRequest,
	opts ...yarpc.CallOption,
) error {

	var err error
	var client Client
	if request.ShardID != nil {
		client, err = c.getClientForShardID(int(request.GetShardID()))
		if err != nil {
			return err
		}
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		err = client.CloseShard(ctx, request, opts...)
		return err
	}

	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) ResetQueue(
	ctx context.Context,
	request *types.ResetQueueRequest,
	opts ...yarpc.CallOption,
) error {

	var err error
	var client Client
	if request.ShardID != nil {
		client, err = c.getClientForShardID(int(request.GetShardID()))
		if err != nil {
			return err
		}
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		err = client.ResetQueue(ctx, request, opts...)
		return err
	}

	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return err
	}
	return nil
}

func (c *clientImpl) DescribeQueue(
	ctx context.Context,
	request *types.DescribeQueueRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeQueueResponse, error) {

	var err error
	var client Client
	if request.ShardID != nil {
		client, err = c.getClientForShardID(int(request.GetShardID()))
		if err != nil {
			return nil, err
		}
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.DescribeQueueResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.DescribeQueue(ctx, request, opts...)
		return err
	}

	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) DescribeMutableState(
	ctx context.Context,
	request *types.DescribeMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeMutableStateResponse, error) {
	client, err := c.getClientForWorkflowID(*request.Execution.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.DescribeMutableStateResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.DescribeMutableState(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) ResetStickyTaskList(
	ctx context.Context,
	request *types.HistoryResetStickyTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryResetStickyTaskListResponse, error) {
	client, err := c.getClientForWorkflowID(*request.Execution.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.HistoryResetStickyTaskListResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.ResetStickyTaskList(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.HistoryDescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeWorkflowExecutionResponse, error) {
	client, err := c.getClientForWorkflowID(*request.Request.Execution.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.DescribeWorkflowExecutionResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.DescribeWorkflowExecution(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) RecordDecisionTaskStarted(
	ctx context.Context,
	request *types.RecordDecisionTaskStartedRequest,
	opts ...yarpc.CallOption,
) (*types.RecordDecisionTaskStartedResponse, error) {
	client, err := c.getClientForWorkflowID(*request.WorkflowExecution.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.RecordDecisionTaskStartedResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.RecordDecisionTaskStarted(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) RecordActivityTaskStarted(
	ctx context.Context,
	request *types.RecordActivityTaskStartedRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskStartedResponse, error) {
	client, err := c.getClientForWorkflowID(*request.WorkflowExecution.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.RecordActivityTaskStartedResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.RecordActivityTaskStarted(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *types.HistoryRespondDecisionTaskCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryRespondDecisionTaskCompletedResponse, error) {
	taskToken, err := c.tokenSerializer.Deserialize(request.CompleteRequest.TaskToken)
	if err != nil {
		return nil, err
	}
	client, err := c.getClientForWorkflowID(taskToken.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.HistoryRespondDecisionTaskCompletedResponse
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.RespondDecisionTaskCompleted(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	return response, err
}

func (c *clientImpl) RespondDecisionTaskFailed(
	ctx context.Context,
	request *types.HistoryRespondDecisionTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {
	taskToken, err := c.tokenSerializer.Deserialize(request.FailedRequest.TaskToken)
	if err != nil {
		return err
	}
	client, err := c.getClientForWorkflowID(taskToken.WorkflowID)
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.RespondDecisionTaskFailed(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) RespondActivityTaskCompleted(
	ctx context.Context,
	request *types.HistoryRespondActivityTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	taskToken, err := c.tokenSerializer.Deserialize(request.CompleteRequest.TaskToken)
	if err != nil {
		return err
	}
	client, err := c.getClientForWorkflowID(taskToken.WorkflowID)
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.RespondActivityTaskCompleted(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) RespondActivityTaskFailed(
	ctx context.Context,
	request *types.HistoryRespondActivityTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {
	taskToken, err := c.tokenSerializer.Deserialize(request.FailedRequest.TaskToken)
	if err != nil {
		return err
	}
	client, err := c.getClientForWorkflowID(taskToken.WorkflowID)
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.RespondActivityTaskFailed(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) RespondActivityTaskCanceled(
	ctx context.Context,
	request *types.HistoryRespondActivityTaskCanceledRequest,
	opts ...yarpc.CallOption,
) error {
	taskToken, err := c.tokenSerializer.Deserialize(request.CancelRequest.TaskToken)
	if err != nil {
		return err
	}
	client, err := c.getClientForWorkflowID(taskToken.WorkflowID)
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.RespondActivityTaskCanceled(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *types.HistoryRecordActivityTaskHeartbeatRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskHeartbeatResponse, error) {
	taskToken, err := c.tokenSerializer.Deserialize(request.HeartbeatRequest.TaskToken)
	if err != nil {
		return nil, err
	}
	client, err := c.getClientForWorkflowID(taskToken.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.RecordActivityTaskHeartbeatResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.RecordActivityTaskHeartbeat(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *types.HistoryRequestCancelWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	client, err := c.getClientForWorkflowID(*request.CancelRequest.WorkflowExecution.WorkflowID)
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.RequestCancelWorkflowExecution(ctx, request, opts...)
	}
	return c.executeWithRedirect(ctx, client, op)
}

func (c *clientImpl) SignalWorkflowExecution(
	ctx context.Context,
	request *types.HistorySignalWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	client, err := c.getClientForWorkflowID(*request.SignalRequest.WorkflowExecution.WorkflowID)
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.SignalWorkflowExecution(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)

	return err
}

func (c *clientImpl) SignalWithStartWorkflowExecution(
	ctx context.Context,
	request *types.HistorySignalWithStartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {
	client, err := c.getClientForWorkflowID(*request.SignalWithStartRequest.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.StartWorkflowExecutionResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.SignalWithStartWorkflowExecution(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}

	return response, err
}

func (c *clientImpl) RemoveSignalMutableState(
	ctx context.Context,
	request *types.RemoveSignalMutableStateRequest,
	opts ...yarpc.CallOption,
) error {
	client, err := c.getClientForWorkflowID(*request.WorkflowExecution.WorkflowID)
	if err != nil {
		return err
	}
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.RemoveSignalMutableState(ctx, request)
	}
	err = c.executeWithRedirect(ctx, client, op)

	return err
}

func (c *clientImpl) TerminateWorkflowExecution(
	ctx context.Context,
	request *types.HistoryTerminateWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	client, err := c.getClientForWorkflowID(*request.TerminateRequest.WorkflowExecution.WorkflowID)
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.TerminateWorkflowExecution(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) ResetWorkflowExecution(
	ctx context.Context,
	request *types.HistoryResetWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.ResetWorkflowExecutionResponse, error) {
	client, err := c.getClientForWorkflowID(*request.ResetRequest.WorkflowExecution.WorkflowID)
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.ResetWorkflowExecutionResponse
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.ResetWorkflowExecution(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, err
}

func (c *clientImpl) ScheduleDecisionTask(
	ctx context.Context,
	request *types.ScheduleDecisionTaskRequest,
	opts ...yarpc.CallOption,
) error {
	client, err := c.getClientForWorkflowID(*request.WorkflowExecution.WorkflowID)
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.ScheduleDecisionTask(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) RecordChildExecutionCompleted(
	ctx context.Context,
	request *types.RecordChildExecutionCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	client, err := c.getClientForWorkflowID(*request.WorkflowExecution.WorkflowID)
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.RecordChildExecutionCompleted(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) ReplicateEventsV2(
	ctx context.Context,
	request *types.ReplicateEventsV2Request,
	opts ...yarpc.CallOption,
) error {
	client, err := c.getClientForWorkflowID(request.WorkflowExecution.GetWorkflowID())
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.ReplicateEventsV2(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) SyncShardStatus(
	ctx context.Context,
	request *types.SyncShardStatusRequest,
	opts ...yarpc.CallOption,
) error {

	// we do not have a workflow ID here, instead, we have something even better
	client, err := c.getClientForShardID(int(request.GetShardID()))
	if err != nil {
		return err
	}

	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.SyncShardStatus(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) SyncActivity(
	ctx context.Context,
	request *types.SyncActivityRequest,
	opts ...yarpc.CallOption,
) error {

	client, err := c.getClientForWorkflowID(request.GetWorkflowID())
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.SyncActivity(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) QueryWorkflow(
	ctx context.Context,
	request *types.HistoryQueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryQueryWorkflowResponse, error) {
	client, err := c.getClientForWorkflowID(request.GetRequest().GetExecution().GetWorkflowID())
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	var response *types.HistoryQueryWorkflowResponse
	op := func(ctx context.Context, client Client) error {
		var err error
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		response, err = client.QueryWorkflow(ctx, request, opts...)
		return err
	}
	err = c.executeWithRedirect(ctx, client, op)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *clientImpl) GetReplicationMessages(
	ctx context.Context,
	request *types.GetReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetReplicationMessagesResponse, error) {
	requestsByClient := make(map[Client]*types.GetReplicationMessagesRequest)

	for _, token := range request.Tokens {
		client, err := c.getClientForShardID(int(token.GetShardID()))
		if err != nil {
			return nil, err
		}

		if _, ok := requestsByClient[client]; !ok {
			requestsByClient[client] = &types.GetReplicationMessagesRequest{
				ClusterName: request.ClusterName,
			}
		}

		req := requestsByClient[client]
		req.Tokens = append(req.Tokens, token)
	}

	var wg sync.WaitGroup
	wg.Add(len(requestsByClient))
	respChan := make(chan *types.GetReplicationMessagesResponse, len(requestsByClient))
	errChan := make(chan error, 1)
	for client, req := range requestsByClient {
		go func(client Client, request *types.GetReplicationMessagesRequest) {
			defer wg.Done()

			ctx, cancel := c.createContext(ctx)
			defer cancel()
			resp, err := client.GetReplicationMessages(ctx, request, opts...)
			if err != nil {
				c.logger.Warn("Failed to get replication tasks from client", tag.Error(err))
				// Returns service busy error to notify replication
				if _, ok := err.(*types.ServiceBusyError); ok {
					select {
					case errChan <- err:
					default:
					}
				}
				return
			}
			respChan <- resp
		}(client, req)
	}

	wg.Wait()
	close(respChan)
	close(errChan)

	response := &types.GetReplicationMessagesResponse{MessagesByShard: make(map[int32]*types.ReplicationMessages)}
	for resp := range respChan {
		for shardID, tasks := range resp.MessagesByShard {
			response.MessagesByShard[shardID] = tasks
		}
	}
	var err error
	if len(errChan) > 0 {
		err = <-errChan
	}
	return response, err
}

func (c *clientImpl) GetDLQReplicationMessages(
	ctx context.Context,
	request *types.GetDLQReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDLQReplicationMessagesResponse, error) {
	// All workflow IDs are in the same shard per request
	workflowID := request.GetTaskInfos()[0].GetWorkflowID()
	client, err := c.getClientForWorkflowID(workflowID)
	if err != nil {
		return nil, err
	}

	return client.GetDLQReplicationMessages(
		ctx,
		request,
		opts...,
	)
}

func (c *clientImpl) ReapplyEvents(
	ctx context.Context,
	request *types.HistoryReapplyEventsRequest,
	opts ...yarpc.CallOption,
) error {
	client, err := c.getClientForWorkflowID(request.GetRequest().GetWorkflowExecution().GetWorkflowID())
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.ReapplyEvents(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) ReadDLQMessages(
	ctx context.Context,
	request *types.ReadDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.ReadDLQMessagesResponse, error) {

	client, err := c.getClientForShardID(int(request.GetShardID()))
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	return client.ReadDLQMessages(ctx, request, opts...)
}

func (c *clientImpl) PurgeDLQMessages(
	ctx context.Context,
	request *types.PurgeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) error {

	client, err := c.getClientForShardID(int(request.GetShardID()))
	if err != nil {
		return err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	return client.PurgeDLQMessages(ctx, request, opts...)
}

func (c *clientImpl) MergeDLQMessages(
	ctx context.Context,
	request *types.MergeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.MergeDLQMessagesResponse, error) {

	client, err := c.getClientForShardID(int(request.GetShardID()))
	if err != nil {
		return nil, err
	}
	opts = common.AggregateYarpcOptions(ctx, opts...)
	return client.MergeDLQMessages(ctx, request, opts...)
}

func (c *clientImpl) RefreshWorkflowTasks(
	ctx context.Context,
	request *types.HistoryRefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) error {
	client, err := c.getClientForWorkflowID(request.GetRequest().GetExecution().GetWorkflowID())
	op := func(ctx context.Context, client Client) error {
		ctx, cancel := c.createContext(ctx)
		defer cancel()
		return client.RefreshWorkflowTasks(ctx, request, opts...)
	}
	err = c.executeWithRedirect(ctx, client, op)
	return err
}

func (c *clientImpl) NotifyFailoverMarkers(
	ctx context.Context,
	request *types.NotifyFailoverMarkersRequest,
	opts ...yarpc.CallOption,
) error {
	requestsByClient := make(map[Client]*types.NotifyFailoverMarkersRequest)

	for _, token := range request.GetFailoverMarkerTokens() {
		marker := token.GetFailoverMarker()
		client, err := c.getClientForDomainID(marker.GetDomainID())
		if err != nil {
			return err
		}
		if _, ok := requestsByClient[client]; !ok {
			requestsByClient[client] = &types.NotifyFailoverMarkersRequest{
				FailoverMarkerTokens: []*types.FailoverMarkerToken{},
			}
		}

		req := requestsByClient[client]
		req.FailoverMarkerTokens = append(req.FailoverMarkerTokens, token)
	}

	var wg sync.WaitGroup
	wg.Add(len(requestsByClient))
	respChan := make(chan error, len(requestsByClient))
	for client, req := range requestsByClient {
		go func(client Client, request *types.NotifyFailoverMarkersRequest) {
			defer wg.Done()

			ctx, cancel := c.createContext(ctx)
			defer cancel()
			err := client.NotifyFailoverMarkers(
				ctx,
				request,
				opts...,
			)
			respChan <- err
		}(client, req)
	}

	wg.Wait()
	close(respChan)

	for err := range respChan {
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *clientImpl) createContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		return context.WithTimeout(context.Background(), c.timeout)
	}
	return context.WithTimeout(parent, c.timeout)
}

func (c *clientImpl) getClientForWorkflowID(workflowID string) (Client, error) {
	key := common.WorkflowIDToHistoryShard(workflowID, c.numberOfShards)
	return c.getClientForShardID(key)
}

func (c *clientImpl) getClientForDomainID(domainID string) (Client, error) {
	key := common.DomainIDToHistoryShard(domainID, c.numberOfShards)
	return c.getClientForShardID(key)
}

func (c *clientImpl) getClientForShardID(shardID int) (Client, error) {
	client, err := c.clients.GetClientForKey(string(shardID))
	if err != nil {
		return nil, err
	}
	return client.(Client), nil
}

func (c *clientImpl) executeWithRedirect(
	ctx context.Context,
	client Client,
	op func(ctx context.Context, client Client) error,
) error {
	var err error
	if ctx == nil {
		ctx = context.Background()
	}
redirectLoop:
	for {
		err = common.IsValidContext(ctx)
		if err != nil {
			break redirectLoop
		}
		err = op(ctx, client)
		if err != nil {
			if s, ok := err.(*types.ShardOwnershipLostError); ok {
				// TODO: consider emitting a metric for number of redirects
				ret, err := c.clients.GetClientForClientKey(s.GetOwner())
				if err != nil {
					return err
				}
				client = ret.(Client)
				continue redirectLoop
			}
		}
		break redirectLoop
	}
	return err
}
