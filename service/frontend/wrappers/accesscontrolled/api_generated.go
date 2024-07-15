// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

// Code generated by gowrap. DO NOT EDIT.
// template: ../../templates/accesscontrolled.tmpl
// gowrap: http://github.com/hexdigest/gowrap

package accesscontrolled

import (
	"context"

	"github.com/uber/cadence/common/authorization"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/frontend/api"
)

// apiHandler frontend handler wrapper for authentication and authorization
type apiHandler struct {
	handler    api.Handler
	authorizer authorization.Authorizer
	resource.Resource
}

// NewAPIHandler creates frontend handler with authentication support
func NewAPIHandler(handler api.Handler, resource resource.Resource, authorizer authorization.Authorizer, cfg config.Authorization) api.Handler {
	if authorizer == nil {
		var err error
		authorizer, err = authorization.NewAuthorizer(cfg, resource.GetLogger(), resource.GetDomainCache())
		if err != nil {
			resource.GetLogger().Fatal("Error when initiating the Authorizer", tag.Error(err))
		}
	}
	return &apiHandler{
		handler:    handler,
		authorizer: authorizer,
		Resource:   resource,
	}
}

func (a *apiHandler) CountWorkflowExecutions(ctx context.Context, cp1 *types.CountWorkflowExecutionsRequest) (cp2 *types.CountWorkflowExecutionsResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendCountWorkflowExecutionsScope, cp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "CountWorkflowExecutions",
		Permission:  authorization.PermissionRead,
		RequestBody: cp1,
		DomainName:  cp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.CountWorkflowExecutions(ctx, cp1)
}

func (a *apiHandler) DeprecateDomain(ctx context.Context, dp1 *types.DeprecateDomainRequest) (err error) {
	scope := a.GetMetricsClient().Scope(metrics.FrontendDeprecateDomainScope)
	attr := &authorization.Attributes{
		APIName:     "DeprecateDomain",
		Permission:  authorization.PermissionAdmin,
		RequestBody: dp1,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}
	return a.handler.DeprecateDomain(ctx, dp1)
}

func (a *apiHandler) DescribeDomain(ctx context.Context, dp1 *types.DescribeDomainRequest) (dp2 *types.DescribeDomainResponse, err error) {
	scope := a.GetMetricsClient().Scope(metrics.FrontendDescribeDomainScope)
	attr := &authorization.Attributes{
		APIName:     "DescribeDomain",
		Permission:  authorization.PermissionRead,
		RequestBody: dp1,
		DomainName:  dp1.GetName(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.DescribeDomain(ctx, dp1)
}

func (a *apiHandler) DescribeTaskList(ctx context.Context, dp1 *types.DescribeTaskListRequest) (dp2 *types.DescribeTaskListResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendDescribeTaskListScope, dp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "DescribeTaskList",
		Permission:  authorization.PermissionRead,
		RequestBody: dp1,
		DomainName:  dp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.DescribeTaskList(ctx, dp1)
}

func (a *apiHandler) DescribeWorkflowExecution(ctx context.Context, dp1 *types.DescribeWorkflowExecutionRequest) (dp2 *types.DescribeWorkflowExecutionResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendDescribeWorkflowExecutionScope, dp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "DescribeWorkflowExecution",
		Permission:  authorization.PermissionRead,
		RequestBody: dp1,
		DomainName:  dp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.DescribeWorkflowExecution(ctx, dp1)
}

func (a *apiHandler) GetClusterInfo(ctx context.Context) (cp1 *types.ClusterInfo, err error) {
	return a.handler.GetClusterInfo(ctx)
}

func (a *apiHandler) GetSearchAttributes(ctx context.Context) (gp1 *types.GetSearchAttributesResponse, err error) {
	return a.handler.GetSearchAttributes(ctx)
}

func (a *apiHandler) GetTaskListsByDomain(ctx context.Context, gp1 *types.GetTaskListsByDomainRequest) (gp2 *types.GetTaskListsByDomainResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendGetTaskListsByDomainScope, gp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "GetTaskListsByDomain",
		Permission:  authorization.PermissionRead,
		RequestBody: gp1,
		DomainName:  gp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.GetTaskListsByDomain(ctx, gp1)
}

func (a *apiHandler) GetWorkflowExecutionHistory(ctx context.Context, gp1 *types.GetWorkflowExecutionHistoryRequest) (gp2 *types.GetWorkflowExecutionHistoryResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendGetWorkflowExecutionHistoryScope, gp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "GetWorkflowExecutionHistory",
		Permission:  authorization.PermissionRead,
		RequestBody: gp1,
		DomainName:  gp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.GetWorkflowExecutionHistory(ctx, gp1)
}

func (a *apiHandler) Health(ctx context.Context) (hp1 *types.HealthStatus, err error) {
	return a.handler.Health(ctx)
}

func (a *apiHandler) ListAllWorkflowExecutions(ctx context.Context, lp1 *types.ListAllWorkflowExecutionsRequest) (lp2 *types.ListAllWorkflowExecutionsResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendListAllWorkflowExecutionsScope, lp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "ListAllWorkflowExecutions",
		Permission:  authorization.PermissionRead,
		RequestBody: lp1,
		DomainName:  lp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ListAllWorkflowExecutions(ctx, lp1)
}

func (a *apiHandler) ListArchivedWorkflowExecutions(ctx context.Context, lp1 *types.ListArchivedWorkflowExecutionsRequest) (lp2 *types.ListArchivedWorkflowExecutionsResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendListArchivedWorkflowExecutionsScope, lp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "ListArchivedWorkflowExecutions",
		Permission:  authorization.PermissionRead,
		RequestBody: lp1,
		DomainName:  lp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ListArchivedWorkflowExecutions(ctx, lp1)
}

func (a *apiHandler) ListClosedWorkflowExecutions(ctx context.Context, lp1 *types.ListClosedWorkflowExecutionsRequest) (lp2 *types.ListClosedWorkflowExecutionsResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendListClosedWorkflowExecutionsScope, lp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "ListClosedWorkflowExecutions",
		Permission:  authorization.PermissionRead,
		RequestBody: lp1,
		DomainName:  lp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ListClosedWorkflowExecutions(ctx, lp1)
}

func (a *apiHandler) ListDomains(ctx context.Context, lp1 *types.ListDomainsRequest) (lp2 *types.ListDomainsResponse, err error) {
	scope := a.GetMetricsClient().Scope(metrics.FrontendListDomainsScope)
	attr := &authorization.Attributes{
		APIName:     "ListDomains",
		Permission:  authorization.PermissionAdmin,
		RequestBody: lp1,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ListDomains(ctx, lp1)
}

func (a *apiHandler) ListOpenWorkflowExecutions(ctx context.Context, lp1 *types.ListOpenWorkflowExecutionsRequest) (lp2 *types.ListOpenWorkflowExecutionsResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendListOpenWorkflowExecutionsScope, lp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "ListOpenWorkflowExecutions",
		Permission:  authorization.PermissionRead,
		RequestBody: lp1,
		DomainName:  lp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ListOpenWorkflowExecutions(ctx, lp1)
}

func (a *apiHandler) ListTaskListPartitions(ctx context.Context, lp1 *types.ListTaskListPartitionsRequest) (lp2 *types.ListTaskListPartitionsResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendListTaskListPartitionsScope, lp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "ListTaskListPartitions",
		Permission:  authorization.PermissionRead,
		RequestBody: lp1,
		DomainName:  lp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ListTaskListPartitions(ctx, lp1)
}

func (a *apiHandler) ListWorkflowExecutions(ctx context.Context, lp1 *types.ListWorkflowExecutionsRequest) (lp2 *types.ListWorkflowExecutionsResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendListWorkflowExecutionsScope, lp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "ListWorkflowExecutions",
		Permission:  authorization.PermissionRead,
		RequestBody: lp1,
		DomainName:  lp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ListWorkflowExecutions(ctx, lp1)
}

func (a *apiHandler) PollForActivityTask(ctx context.Context, pp1 *types.PollForActivityTaskRequest) (pp2 *types.PollForActivityTaskResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendPollForActivityTaskScope, pp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "PollForActivityTask",
		Permission:  authorization.PermissionWrite,
		RequestBody: pp1,
		DomainName:  pp1.GetDomain(),
		TaskList:    pp1.TaskList,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.PollForActivityTask(ctx, pp1)
}

func (a *apiHandler) PollForDecisionTask(ctx context.Context, pp1 *types.PollForDecisionTaskRequest) (pp2 *types.PollForDecisionTaskResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendPollForDecisionTaskScope, pp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "PollForDecisionTask",
		Permission:  authorization.PermissionWrite,
		RequestBody: pp1,
		DomainName:  pp1.GetDomain(),
		TaskList:    pp1.TaskList,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.PollForDecisionTask(ctx, pp1)
}

func (a *apiHandler) QueryWorkflow(ctx context.Context, qp1 *types.QueryWorkflowRequest) (qp2 *types.QueryWorkflowResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendQueryWorkflowScope, qp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "QueryWorkflow",
		Permission:  authorization.PermissionRead,
		RequestBody: qp1,
		DomainName:  qp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.QueryWorkflow(ctx, qp1)
}

func (a *apiHandler) RecordActivityTaskHeartbeat(ctx context.Context, rp1 *types.RecordActivityTaskHeartbeatRequest) (rp2 *types.RecordActivityTaskHeartbeatResponse, err error) {
	return a.handler.RecordActivityTaskHeartbeat(ctx, rp1)
}

func (a *apiHandler) RecordActivityTaskHeartbeatByID(ctx context.Context, rp1 *types.RecordActivityTaskHeartbeatByIDRequest) (rp2 *types.RecordActivityTaskHeartbeatResponse, err error) {
	return a.handler.RecordActivityTaskHeartbeatByID(ctx, rp1)
}

func (a *apiHandler) RefreshWorkflowTasks(ctx context.Context, rp1 *types.RefreshWorkflowTasksRequest) (err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendRefreshWorkflowTasksScope, rp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "RefreshWorkflowTasks",
		Permission:  authorization.PermissionWrite,
		RequestBody: rp1,
		DomainName:  rp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}
	return a.handler.RefreshWorkflowTasks(ctx, rp1)
}

func (a *apiHandler) RegisterDomain(ctx context.Context, rp1 *types.RegisterDomainRequest) (err error) {
	scope := a.GetMetricsClient().Scope(metrics.FrontendRegisterDomainScope)
	attr := &authorization.Attributes{
		APIName:     "RegisterDomain",
		Permission:  authorization.PermissionAdmin,
		RequestBody: rp1,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}
	return a.handler.RegisterDomain(ctx, rp1)
}

func (a *apiHandler) RequestCancelWorkflowExecution(ctx context.Context, rp1 *types.RequestCancelWorkflowExecutionRequest) (err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendRequestCancelWorkflowExecutionScope, rp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "RequestCancelWorkflowExecution",
		Permission:  authorization.PermissionWrite,
		RequestBody: rp1,
		DomainName:  rp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}
	return a.handler.RequestCancelWorkflowExecution(ctx, rp1)
}

func (a *apiHandler) ResetStickyTaskList(ctx context.Context, rp1 *types.ResetStickyTaskListRequest) (rp2 *types.ResetStickyTaskListResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendResetStickyTaskListScope, rp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "ResetStickyTaskList",
		Permission:  authorization.PermissionWrite,
		RequestBody: rp1,
		DomainName:  rp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ResetStickyTaskList(ctx, rp1)
}

func (a *apiHandler) ResetWorkflowExecution(ctx context.Context, rp1 *types.ResetWorkflowExecutionRequest) (rp2 *types.ResetWorkflowExecutionResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendResetWorkflowExecutionScope, rp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "ResetWorkflowExecution",
		Permission:  authorization.PermissionWrite,
		RequestBody: rp1,
		DomainName:  rp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ResetWorkflowExecution(ctx, rp1)
}

func (a *apiHandler) RespondActivityTaskCanceled(ctx context.Context, rp1 *types.RespondActivityTaskCanceledRequest) (err error) {
	return a.handler.RespondActivityTaskCanceled(ctx, rp1)
}

func (a *apiHandler) RespondActivityTaskCanceledByID(ctx context.Context, rp1 *types.RespondActivityTaskCanceledByIDRequest) (err error) {
	return a.handler.RespondActivityTaskCanceledByID(ctx, rp1)
}

func (a *apiHandler) RespondActivityTaskCompleted(ctx context.Context, rp1 *types.RespondActivityTaskCompletedRequest) (err error) {
	return a.handler.RespondActivityTaskCompleted(ctx, rp1)
}

func (a *apiHandler) RespondActivityTaskCompletedByID(ctx context.Context, rp1 *types.RespondActivityTaskCompletedByIDRequest) (err error) {
	return a.handler.RespondActivityTaskCompletedByID(ctx, rp1)
}

func (a *apiHandler) RespondActivityTaskFailed(ctx context.Context, rp1 *types.RespondActivityTaskFailedRequest) (err error) {
	return a.handler.RespondActivityTaskFailed(ctx, rp1)
}

func (a *apiHandler) RespondActivityTaskFailedByID(ctx context.Context, rp1 *types.RespondActivityTaskFailedByIDRequest) (err error) {
	return a.handler.RespondActivityTaskFailedByID(ctx, rp1)
}

func (a *apiHandler) RespondDecisionTaskCompleted(ctx context.Context, rp1 *types.RespondDecisionTaskCompletedRequest) (rp2 *types.RespondDecisionTaskCompletedResponse, err error) {
	return a.handler.RespondDecisionTaskCompleted(ctx, rp1)
}

func (a *apiHandler) RespondDecisionTaskFailed(ctx context.Context, rp1 *types.RespondDecisionTaskFailedRequest) (err error) {
	return a.handler.RespondDecisionTaskFailed(ctx, rp1)
}

func (a *apiHandler) RespondQueryTaskCompleted(ctx context.Context, rp1 *types.RespondQueryTaskCompletedRequest) (err error) {
	return a.handler.RespondQueryTaskCompleted(ctx, rp1)
}

func (a *apiHandler) RestartWorkflowExecution(ctx context.Context, rp1 *types.RestartWorkflowExecutionRequest) (rp2 *types.RestartWorkflowExecutionResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendRestartWorkflowExecutionScope, rp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "RestartWorkflowExecution",
		Permission:  authorization.PermissionWrite,
		RequestBody: rp1,
		DomainName:  rp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.RestartWorkflowExecution(ctx, rp1)
}

func (a *apiHandler) ScanWorkflowExecutions(ctx context.Context, lp1 *types.ListWorkflowExecutionsRequest) (lp2 *types.ListWorkflowExecutionsResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendScanWorkflowExecutionsScope, lp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "ScanWorkflowExecutions",
		Permission:  authorization.PermissionRead,
		RequestBody: lp1,
		DomainName:  lp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.ScanWorkflowExecutions(ctx, lp1)
}

func (a *apiHandler) SignalWithStartWorkflowExecution(ctx context.Context, sp1 *types.SignalWithStartWorkflowExecutionRequest) (sp2 *types.StartWorkflowExecutionResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendSignalWithStartWorkflowExecutionScope, sp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:      "SignalWithStartWorkflowExecution",
		Permission:   authorization.PermissionWrite,
		RequestBody:  sp1,
		DomainName:   sp1.GetDomain(),
		WorkflowType: sp1.WorkflowType,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.SignalWithStartWorkflowExecution(ctx, sp1)
}

func (a *apiHandler) SignalWithStartWorkflowExecutionAsync(ctx context.Context, sp1 *types.SignalWithStartWorkflowExecutionAsyncRequest) (sp2 *types.SignalWithStartWorkflowExecutionAsyncResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendSignalWithStartWorkflowExecutionAsyncScope, sp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "SignalWithStartWorkflowExecutionAsync",
		Permission:  authorization.PermissionWrite,
		RequestBody: sp1,
		DomainName:  sp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.SignalWithStartWorkflowExecutionAsync(ctx, sp1)
}

func (a *apiHandler) SignalWorkflowExecution(ctx context.Context, sp1 *types.SignalWorkflowExecutionRequest) (err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendSignalWorkflowExecutionScope, sp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "SignalWorkflowExecution",
		Permission:  authorization.PermissionWrite,
		RequestBody: sp1,
		DomainName:  sp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}
	return a.handler.SignalWorkflowExecution(ctx, sp1)
}

func (a *apiHandler) StartWorkflowExecution(ctx context.Context, sp1 *types.StartWorkflowExecutionRequest) (sp2 *types.StartWorkflowExecutionResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendStartWorkflowExecutionScope, sp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:      "StartWorkflowExecution",
		Permission:   authorization.PermissionWrite,
		RequestBody:  sp1,
		DomainName:   sp1.GetDomain(),
		WorkflowType: sp1.WorkflowType,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.StartWorkflowExecution(ctx, sp1)
}

func (a *apiHandler) StartWorkflowExecutionAsync(ctx context.Context, sp1 *types.StartWorkflowExecutionAsyncRequest) (sp2 *types.StartWorkflowExecutionAsyncResponse, err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendStartWorkflowExecutionAsyncScope, sp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "StartWorkflowExecutionAsync",
		Permission:  authorization.PermissionWrite,
		RequestBody: sp1,
		DomainName:  sp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.StartWorkflowExecutionAsync(ctx, sp1)
}

func (a *apiHandler) TerminateWorkflowExecution(ctx context.Context, tp1 *types.TerminateWorkflowExecutionRequest) (err error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendTerminateWorkflowExecutionScope, tp1.GetDomain())
	attr := &authorization.Attributes{
		APIName:     "TerminateWorkflowExecution",
		Permission:  authorization.PermissionWrite,
		RequestBody: tp1,
		DomainName:  tp1.GetDomain(),
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}
	return a.handler.TerminateWorkflowExecution(ctx, tp1)
}

func (a *apiHandler) UpdateDomain(ctx context.Context, up1 *types.UpdateDomainRequest) (up2 *types.UpdateDomainResponse, err error) {
	scope := a.GetMetricsClient().Scope(metrics.FrontendUpdateDomainScope)
	attr := &authorization.Attributes{
		APIName:     "UpdateDomain",
		Permission:  authorization.PermissionAdmin,
		RequestBody: up1,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}
	return a.handler.UpdateDomain(ctx, up1)
}
