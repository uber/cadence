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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination interface_mock.go -self_package github.com/uber/cadence/service/frontend

package frontend

import (
	"context"

	"github.com/uber/cadence/.gen/go/health"
	"github.com/uber/cadence/.gen/go/shared"
)

type (
	// Handler is interface wrapping frontend handler
	Handler interface {
		Health(context.Context) (*health.HealthStatus, error)
		CountWorkflowExecutions(context.Context, *shared.CountWorkflowExecutionsRequest) (*shared.CountWorkflowExecutionsResponse, error)
		DeprecateDomain(context.Context, *shared.DeprecateDomainRequest) error
		DescribeDomain(context.Context, *shared.DescribeDomainRequest) (*shared.DescribeDomainResponse, error)
		DescribeTaskList(context.Context, *shared.DescribeTaskListRequest) (*shared.DescribeTaskListResponse, error)
		DescribeWorkflowExecution(context.Context, *shared.DescribeWorkflowExecutionRequest) (*shared.DescribeWorkflowExecutionResponse, error)
		GetClusterInfo(context.Context) (*shared.ClusterInfo, error)
		GetSearchAttributes(context.Context) (*shared.GetSearchAttributesResponse, error)
		GetWorkflowExecutionHistory(context.Context, *shared.GetWorkflowExecutionHistoryRequest) (*shared.GetWorkflowExecutionHistoryResponse, error)
		ListArchivedWorkflowExecutions(context.Context, *shared.ListArchivedWorkflowExecutionsRequest) (*shared.ListArchivedWorkflowExecutionsResponse, error)
		ListClosedWorkflowExecutions(context.Context, *shared.ListClosedWorkflowExecutionsRequest) (*shared.ListClosedWorkflowExecutionsResponse, error)
		ListDomains(context.Context, *shared.ListDomainsRequest) (*shared.ListDomainsResponse, error)
		ListOpenWorkflowExecutions(context.Context, *shared.ListOpenWorkflowExecutionsRequest) (*shared.ListOpenWorkflowExecutionsResponse, error)
		ListTaskListPartitions(context.Context, *shared.ListTaskListPartitionsRequest) (*shared.ListTaskListPartitionsResponse, error)
		ListWorkflowExecutions(context.Context, *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error)
		PollForActivityTask(context.Context, *shared.PollForActivityTaskRequest) (*shared.PollForActivityTaskResponse, error)
		PollForDecisionTask(context.Context, *shared.PollForDecisionTaskRequest) (*shared.PollForDecisionTaskResponse, error)
		QueryWorkflow(context.Context, *shared.QueryWorkflowRequest) (*shared.QueryWorkflowResponse, error)
		RecordActivityTaskHeartbeat(context.Context, *shared.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse, error)
		RecordActivityTaskHeartbeatByID(context.Context, *shared.RecordActivityTaskHeartbeatByIDRequest) (*shared.RecordActivityTaskHeartbeatResponse, error)
		RegisterDomain(context.Context, *shared.RegisterDomainRequest) error
		RequestCancelWorkflowExecution(context.Context, *shared.RequestCancelWorkflowExecutionRequest) error
		ResetStickyTaskList(context.Context, *shared.ResetStickyTaskListRequest) (*shared.ResetStickyTaskListResponse, error)
		ResetWorkflowExecution(context.Context, *shared.ResetWorkflowExecutionRequest) (*shared.ResetWorkflowExecutionResponse, error)
		RespondActivityTaskCanceled(context.Context, *shared.RespondActivityTaskCanceledRequest) error
		RespondActivityTaskCanceledByID(context.Context, *shared.RespondActivityTaskCanceledByIDRequest) error
		RespondActivityTaskCompleted(context.Context, *shared.RespondActivityTaskCompletedRequest) error
		RespondActivityTaskCompletedByID(context.Context, *shared.RespondActivityTaskCompletedByIDRequest) error
		RespondActivityTaskFailed(context.Context, *shared.RespondActivityTaskFailedRequest) error
		RespondActivityTaskFailedByID(context.Context, *shared.RespondActivityTaskFailedByIDRequest) error
		RespondDecisionTaskCompleted(context.Context, *shared.RespondDecisionTaskCompletedRequest) (*shared.RespondDecisionTaskCompletedResponse, error)
		RespondDecisionTaskFailed(context.Context, *shared.RespondDecisionTaskFailedRequest) error
		RespondQueryTaskCompleted(context.Context, *shared.RespondQueryTaskCompletedRequest) error
		ScanWorkflowExecutions(context.Context, *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error)
		SignalWithStartWorkflowExecution(context.Context, *shared.SignalWithStartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error)
		SignalWorkflowExecution(context.Context, *shared.SignalWorkflowExecutionRequest) error
		StartWorkflowExecution(context.Context, *shared.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error)
		TerminateWorkflowExecution(context.Context, *shared.TerminateWorkflowExecutionRequest) error
		UpdateDomain(context.Context, *shared.UpdateDomainRequest) (*shared.UpdateDomainResponse, error)
	}
)
