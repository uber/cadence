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

package matching

import (
	"context"
	"time"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/common/future"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

var _ Client = (*clientImpl)(nil)

const (
	// DefaultTimeout is the default timeout used to make calls
	DefaultTimeout = time.Minute
	// DefaultLongPollTimeout is the long poll default timeout used to make calls
	DefaultLongPollTimeout = time.Minute * 2
)

type clientImpl struct {
	timeout         time.Duration
	longPollTimeout time.Duration
	client          Client
	peerResolver    PeerResolver
	loadBalancer    LoadBalancer
}

// NewClient creates a new history service TChannel client
func NewClient(
	timeout time.Duration,
	longPollTimeout time.Duration,
	client Client,
	peerResolver PeerResolver,
	lb LoadBalancer,
) Client {
	return &clientImpl{
		timeout:         timeout,
		longPollTimeout: longPollTimeout,
		client:          client,
		peerResolver:    peerResolver,
		loadBalancer:    lb,
	}
}

func (c *clientImpl) AddActivityTask(
	ctx context.Context,
	request *types.AddActivityTaskRequest,
	opts ...yarpc.CallOption,
) error {
	partition := c.loadBalancer.PickWritePartition(
		request.GetDomainUUID(),
		*request.GetTaskList(),
		persistence.TaskListTypeActivity,
		request.GetForwardedFrom(),
	)
	request.TaskList.Name = partition
	peer, err := c.peerResolver.FromTaskList(partition)
	if err != nil {
		return err
	}
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.AddActivityTask(ctx, request, append(opts, yarpc.WithShardKey(peer))...)
}

func (c *clientImpl) AddDecisionTask(
	ctx context.Context,
	request *types.AddDecisionTaskRequest,
	opts ...yarpc.CallOption,
) error {
	partition := c.loadBalancer.PickWritePartition(
		request.GetDomainUUID(),
		*request.GetTaskList(),
		persistence.TaskListTypeDecision,
		request.GetForwardedFrom(),
	)
	request.TaskList.Name = partition
	peer, err := c.peerResolver.FromTaskList(request.TaskList.GetName())
	if err != nil {
		return err
	}
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.AddDecisionTask(ctx, request, append(opts, yarpc.WithShardKey(peer))...)
}

func (c *clientImpl) PollForActivityTask(
	ctx context.Context,
	request *types.MatchingPollForActivityTaskRequest,
	opts ...yarpc.CallOption,
) (*types.PollForActivityTaskResponse, error) {
	partition := c.loadBalancer.PickReadPartition(
		request.GetDomainUUID(),
		*request.PollRequest.GetTaskList(),
		persistence.TaskListTypeActivity,
		request.GetForwardedFrom(),
	)
	request.PollRequest.TaskList.Name = partition
	peer, err := c.peerResolver.FromTaskList(request.PollRequest.TaskList.GetName())
	if err != nil {
		return nil, err
	}
	ctx, cancel := c.createLongPollContext(ctx)
	defer cancel()
	return c.client.PollForActivityTask(ctx, request, append(opts, yarpc.WithShardKey(peer))...)
}

func (c *clientImpl) PollForDecisionTask(
	ctx context.Context,
	request *types.MatchingPollForDecisionTaskRequest,
	opts ...yarpc.CallOption,
) (*types.MatchingPollForDecisionTaskResponse, error) {
	partition := c.loadBalancer.PickReadPartition(
		request.GetDomainUUID(),
		*request.PollRequest.GetTaskList(),
		persistence.TaskListTypeDecision,
		request.GetForwardedFrom(),
	)
	request.PollRequest.TaskList.Name = partition
	peer, err := c.peerResolver.FromTaskList(request.PollRequest.TaskList.GetName())
	if err != nil {
		return nil, err
	}
	ctx, cancel := c.createLongPollContext(ctx)
	defer cancel()
	return c.client.PollForDecisionTask(ctx, request, append(opts, yarpc.WithShardKey(peer))...)
}

func (c *clientImpl) QueryWorkflow(
	ctx context.Context,
	request *types.MatchingQueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.QueryWorkflowResponse, error) {
	partition := c.loadBalancer.PickReadPartition(
		request.GetDomainUUID(),
		*request.GetTaskList(),
		persistence.TaskListTypeDecision,
		request.GetForwardedFrom(),
	)
	request.TaskList.Name = partition
	peer, err := c.peerResolver.FromTaskList(request.TaskList.GetName())
	if err != nil {
		return nil, err
	}
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.QueryWorkflow(ctx, request, append(opts, yarpc.WithShardKey(peer))...)
}

func (c *clientImpl) RespondQueryTaskCompleted(
	ctx context.Context,
	request *types.MatchingRespondQueryTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	peer, err := c.peerResolver.FromTaskList(request.TaskList.GetName())
	if err != nil {
		return err
	}
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondQueryTaskCompleted(ctx, request, append(opts, yarpc.WithShardKey(peer))...)
}

func (c *clientImpl) CancelOutstandingPoll(
	ctx context.Context,
	request *types.CancelOutstandingPollRequest,
	opts ...yarpc.CallOption,
) error {
	peer, err := c.peerResolver.FromTaskList(request.TaskList.GetName())
	if err != nil {
		return err
	}
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.CancelOutstandingPoll(ctx, request, append(opts, yarpc.WithShardKey(peer))...)
}

func (c *clientImpl) DescribeTaskList(
	ctx context.Context,
	request *types.MatchingDescribeTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeTaskListResponse, error) {
	peer, err := c.peerResolver.FromTaskList(request.DescRequest.TaskList.GetName())
	if err != nil {
		return nil, err
	}
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DescribeTaskList(ctx, request, append(opts, yarpc.WithShardKey(peer))...)
}

func (c *clientImpl) ListTaskListPartitions(
	ctx context.Context,
	request *types.MatchingListTaskListPartitionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListTaskListPartitionsResponse, error) {
	peer, err := c.peerResolver.FromTaskList(request.TaskList.GetName())
	if err != nil {
		return nil, err
	}
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ListTaskListPartitions(ctx, request, append(opts, yarpc.WithShardKey(peer))...)
}

func (c *clientImpl) GetTaskListsByDomain(
	ctx context.Context,
	request *types.GetTaskListsByDomainRequest,
	opts ...yarpc.CallOption,
) (*types.GetTaskListsByDomainResponse, error) {
	peers, err := c.peerResolver.GetAllPeers()
	if err != nil {
		return nil, err
	}

	var futures []future.Future
	for _, peer := range peers {
		future, settable := future.NewFuture()
		settable.Set(c.client.GetTaskListsByDomain(ctx, request, append(opts, yarpc.WithShardKey(peer))...))
		futures = append(futures, future)
	}

	decisionTaskListMap := make(map[string]*types.DescribeTaskListResponse)
	activityTaskListMap := make(map[string]*types.DescribeTaskListResponse)
	for _, future := range futures {
		var resp *types.GetTaskListsByDomainResponse
		if err = future.Get(ctx, &resp); err != nil {
			return nil, err
		}
		for name, tl := range resp.GetDecisionTaskListMap() {
			if _, ok := decisionTaskListMap[name]; !ok {
				decisionTaskListMap[name] = tl
			} else {
				decisionTaskListMap[name].Pollers = append(decisionTaskListMap[name].Pollers, tl.GetPollers()...)
			}
		}
		for name, tl := range resp.GetActivityTaskListMap() {
			if _, ok := activityTaskListMap[name]; !ok {
				activityTaskListMap[name] = tl
			} else {
				activityTaskListMap[name].Pollers = append(activityTaskListMap[name].Pollers, tl.GetPollers()...)
			}
		}
	}

	return &types.GetTaskListsByDomainResponse{
		DecisionTaskListMap: decisionTaskListMap,
		ActivityTaskListMap: activityTaskListMap,
	}, nil
}

func (c *clientImpl) createContext(
	parent context.Context,
) (context.Context, context.CancelFunc) {
	if parent == nil {
		return context.WithTimeout(context.Background(), c.timeout)
	}
	return context.WithTimeout(parent, c.timeout)
}

func (c *clientImpl) createLongPollContext(
	parent context.Context,
) (context.Context, context.CancelFunc) {
	if parent == nil {
		return context.WithTimeout(context.Background(), c.longPollTimeout)
	}
	return context.WithTimeout(parent, c.longPollTimeout)
}
