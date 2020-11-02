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

	"go.uber.org/yarpc"

	"github.com/uber/cadence/.gen/go/matching"
	"github.com/uber/cadence/.gen/go/shared"
)

// Client is the interface exposed by matching service client
type Client interface {
	AddActivityTask(context.Context, *matching.AddActivityTaskRequest, ...yarpc.CallOption) error
	AddDecisionTask(context.Context, *matching.AddDecisionTaskRequest, ...yarpc.CallOption) error
	CancelOutstandingPoll(context.Context, *matching.CancelOutstandingPollRequest, ...yarpc.CallOption) error
	DescribeTaskList(context.Context, *matching.DescribeTaskListRequest, ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error)
	ListTaskListPartitions(context.Context, *matching.ListTaskListPartitionsRequest, ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error)
	PollForActivityTask(context.Context, *matching.PollForActivityTaskRequest, ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error)
	PollForDecisionTask(context.Context, *matching.PollForDecisionTaskRequest, ...yarpc.CallOption) (*matching.PollForDecisionTaskResponse, error)
	QueryWorkflow(context.Context, *matching.QueryWorkflowRequest, ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error)
	RespondQueryTaskCompleted(context.Context, *matching.RespondQueryTaskCompletedRequest, ...yarpc.CallOption) error
}
