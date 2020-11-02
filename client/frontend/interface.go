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

package frontend

import (
	"context"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/.gen/go/shared"
)

// Client is the interface exposed by frontend service client
type Client interface {
	CountWorkflowExecutions(context.Context, *shared.CountWorkflowExecutionsRequest, ...yarpc.CallOption) (*shared.CountWorkflowExecutionsResponse, error)
	DeprecateDomain(context.Context, *shared.DeprecateDomainRequest, ...yarpc.CallOption) error
	DescribeDomain(context.Context, *shared.DescribeDomainRequest, ...yarpc.CallOption) (*shared.DescribeDomainResponse, error)
	DescribeTaskList(context.Context, *shared.DescribeTaskListRequest, ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error)
	DescribeWorkflowExecution(context.Context, *shared.DescribeWorkflowExecutionRequest, ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error)
	GetClusterInfo(context.Context, ...yarpc.CallOption) (*shared.ClusterInfo, error)
	GetSearchAttributes(context.Context, ...yarpc.CallOption) (*shared.GetSearchAttributesResponse, error)
	GetWorkflowExecutionHistory(context.Context, *shared.GetWorkflowExecutionHistoryRequest, ...yarpc.CallOption) (*shared.GetWorkflowExecutionHistoryResponse, error)
	ListArchivedWorkflowExecutions(context.Context, *shared.ListArchivedWorkflowExecutionsRequest, ...yarpc.CallOption) (*shared.ListArchivedWorkflowExecutionsResponse, error)
	ListClosedWorkflowExecutions(context.Context, *shared.ListClosedWorkflowExecutionsRequest, ...yarpc.CallOption) (*shared.ListClosedWorkflowExecutionsResponse, error)
	ListDomains(context.Context, *shared.ListDomainsRequest, ...yarpc.CallOption) (*shared.ListDomainsResponse, error)
	ListOpenWorkflowExecutions(context.Context, *shared.ListOpenWorkflowExecutionsRequest, ...yarpc.CallOption) (*shared.ListOpenWorkflowExecutionsResponse, error)
	ListTaskListPartitions(context.Context, *shared.ListTaskListPartitionsRequest, ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error)
	ListWorkflowExecutions(context.Context, *shared.ListWorkflowExecutionsRequest, ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error)
	PollForActivityTask(context.Context, *shared.PollForActivityTaskRequest, ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error)
	PollForDecisionTask(context.Context, *shared.PollForDecisionTaskRequest, ...yarpc.CallOption) (*shared.PollForDecisionTaskResponse, error)
	QueryWorkflow(context.Context, *shared.QueryWorkflowRequest, ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error)
	RecordActivityTaskHeartbeat(context.Context, *shared.RecordActivityTaskHeartbeatRequest, ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error)
	RecordActivityTaskHeartbeatByID(context.Context, *shared.RecordActivityTaskHeartbeatByIDRequest, ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error)
	RegisterDomain(context.Context, *shared.RegisterDomainRequest, ...yarpc.CallOption) error
	RequestCancelWorkflowExecution(context.Context, *shared.RequestCancelWorkflowExecutionRequest, ...yarpc.CallOption) error
	ResetStickyTaskList(context.Context, *shared.ResetStickyTaskListRequest, ...yarpc.CallOption) (*shared.ResetStickyTaskListResponse, error)
	ResetWorkflowExecution(context.Context, *shared.ResetWorkflowExecutionRequest, ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error)
	RespondActivityTaskCanceled(context.Context, *shared.RespondActivityTaskCanceledRequest, ...yarpc.CallOption) error
	RespondActivityTaskCanceledByID(context.Context, *shared.RespondActivityTaskCanceledByIDRequest, ...yarpc.CallOption) error
	RespondActivityTaskCompleted(context.Context, *shared.RespondActivityTaskCompletedRequest, ...yarpc.CallOption) error
	RespondActivityTaskCompletedByID(context.Context, *shared.RespondActivityTaskCompletedByIDRequest, ...yarpc.CallOption) error
	RespondActivityTaskFailed(context.Context, *shared.RespondActivityTaskFailedRequest, ...yarpc.CallOption) error
	RespondActivityTaskFailedByID(context.Context, *shared.RespondActivityTaskFailedByIDRequest, ...yarpc.CallOption) error
	RespondDecisionTaskCompleted(context.Context, *shared.RespondDecisionTaskCompletedRequest, ...yarpc.CallOption) (*shared.RespondDecisionTaskCompletedResponse, error)
	RespondDecisionTaskFailed(context.Context, *shared.RespondDecisionTaskFailedRequest, ...yarpc.CallOption) error
	RespondQueryTaskCompleted(context.Context, *shared.RespondQueryTaskCompletedRequest, ...yarpc.CallOption) error
	ScanWorkflowExecutions(context.Context, *shared.ListWorkflowExecutionsRequest, ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error)
	SignalWithStartWorkflowExecution(context.Context, *shared.SignalWithStartWorkflowExecutionRequest, ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error)
	SignalWorkflowExecution(context.Context, *shared.SignalWorkflowExecutionRequest, ...yarpc.CallOption) error
	StartWorkflowExecution(context.Context, *shared.StartWorkflowExecutionRequest, ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error)
	TerminateWorkflowExecution(context.Context, *shared.TerminateWorkflowExecutionRequest, ...yarpc.CallOption) error
	UpdateDomain(context.Context, *shared.UpdateDomainRequest, ...yarpc.CallOption) (*shared.UpdateDomainResponse, error)
}
