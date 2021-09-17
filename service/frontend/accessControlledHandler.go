// Copyright (c) 2019 Uber Technologies, Inc.
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

	"github.com/uber/cadence/common/authorization"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/types"
)

var errUnauthorized = &types.BadRequestError{Message: "Request unauthorized."}

// AccessControlledWorkflowHandler frontend handler wrapper for authentication and authorization
type AccessControlledWorkflowHandler struct {
	resource.Resource

	frontendHandler Handler
	authorizer      authorization.Authorizer
}

var _ Handler = (*AccessControlledWorkflowHandler)(nil)

// NewAccessControlledHandlerImpl creates frontend handler with authentication support
func NewAccessControlledHandlerImpl(wfHandler Handler, resource resource.Resource, authorizer authorization.Authorizer, cfg config.Authorization) *AccessControlledWorkflowHandler {
	if authorizer == nil {
		var err error
		authorizer, err = authorization.NewAuthorizer(cfg, resource.GetLogger(), resource.GetDomainCache())
		if err != nil {
			resource.GetLogger().Fatal("Error when initiating the Authorizer", tag.Error(err))
		}
	}
	return &AccessControlledWorkflowHandler{
		Resource:        resource,
		frontendHandler: wfHandler,
		authorizer:      authorizer,
	}
}

// Health callback for for health check
func (a *AccessControlledWorkflowHandler) Health(ctx context.Context) (*types.HealthStatus, error) {
	return a.frontendHandler.Health(ctx)
}

// CountWorkflowExecutions API call
func (a *AccessControlledWorkflowHandler) CountWorkflowExecutions(
	ctx context.Context,
	request *types.CountWorkflowExecutionsRequest,
) (*types.CountWorkflowExecutionsResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendCountWorkflowExecutionsScope, request)

	attr := &authorization.Attributes{
		APIName:    "CountWorkflowExecutions",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.CountWorkflowExecutions(ctx, request)
}

// DeprecateDomain API call
func (a *AccessControlledWorkflowHandler) DeprecateDomain(
	ctx context.Context,
	request *types.DeprecateDomainRequest,
) error {

	scope := a.getMetricsScopeWithDomainName(metrics.FrontendDeprecateDomainScope, request.GetName())

	attr := &authorization.Attributes{
		APIName:    "DeprecateDomain",
		DomainName: request.GetName(),
		Permission: authorization.PermissionAdmin,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}

	return a.frontendHandler.DeprecateDomain(ctx, request)
}

// DescribeDomain API call
func (a *AccessControlledWorkflowHandler) DescribeDomain(
	ctx context.Context,
	request *types.DescribeDomainRequest,
) (*types.DescribeDomainResponse, error) {

	scope := a.getMetricsScopeWithDomainName(metrics.FrontendDescribeDomainScope, request.GetName())

	attr := &authorization.Attributes{
		APIName:    "DescribeDomain",
		DomainName: request.GetName(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.DescribeDomain(ctx, request)
}

// DescribeTaskList API call
func (a *AccessControlledWorkflowHandler) DescribeTaskList(
	ctx context.Context,
	request *types.DescribeTaskListRequest,
) (*types.DescribeTaskListResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendDescribeTaskListScope, request)

	attr := &authorization.Attributes{
		APIName:    "DescribeTaskList",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.DescribeTaskList(ctx, request)
}

// DescribeWorkflowExecution API call
func (a *AccessControlledWorkflowHandler) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.DescribeWorkflowExecutionRequest,
) (*types.DescribeWorkflowExecutionResponse, error) {
	scope := a.getMetricsScopeWithDomain(metrics.FrontendDescribeWorkflowExecutionScope, request)

	attr := &authorization.Attributes{
		APIName:    "DescribeWorkflowExecution",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.DescribeWorkflowExecution(ctx, request)
}

// GetSearchAttributes API call
func (a *AccessControlledWorkflowHandler) GetSearchAttributes(
	ctx context.Context,
) (*types.GetSearchAttributesResponse, error) {
	return a.frontendHandler.GetSearchAttributes(ctx)
}

// GetWorkflowExecutionHistory API call
func (a *AccessControlledWorkflowHandler) GetWorkflowExecutionHistory(
	ctx context.Context,
	request *types.GetWorkflowExecutionHistoryRequest,
) (*types.GetWorkflowExecutionHistoryResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendGetWorkflowExecutionHistoryScope, request)

	attr := &authorization.Attributes{
		APIName:    "GetWorkflowExecutionHistory",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.GetWorkflowExecutionHistory(ctx, request)
}

// ListArchivedWorkflowExecutions API call
func (a *AccessControlledWorkflowHandler) ListArchivedWorkflowExecutions(
	ctx context.Context,
	request *types.ListArchivedWorkflowExecutionsRequest,
) (*types.ListArchivedWorkflowExecutionsResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendListArchivedWorkflowExecutionsScope, request)

	attr := &authorization.Attributes{
		APIName:    "ListArchivedWorkflowExecutions",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListArchivedWorkflowExecutions(ctx, request)
}

// ListClosedWorkflowExecutions API call
func (a *AccessControlledWorkflowHandler) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *types.ListClosedWorkflowExecutionsRequest,
) (*types.ListClosedWorkflowExecutionsResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendListClosedWorkflowExecutionsScope, request)

	attr := &authorization.Attributes{
		APIName:    "ListClosedWorkflowExecutions",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListClosedWorkflowExecutions(ctx, request)
}

// ListDomains API call
func (a *AccessControlledWorkflowHandler) ListDomains(
	ctx context.Context,
	request *types.ListDomainsRequest,
) (*types.ListDomainsResponse, error) {

	scope := a.GetMetricsClient().Scope(metrics.FrontendListDomainsScope)

	attr := &authorization.Attributes{
		APIName:    "ListDomains",
		Permission: authorization.PermissionAdmin,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListDomains(ctx, request)
}

// ListOpenWorkflowExecutions API call
func (a *AccessControlledWorkflowHandler) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *types.ListOpenWorkflowExecutionsRequest,
) (*types.ListOpenWorkflowExecutionsResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendListOpenWorkflowExecutionsScope, request)

	attr := &authorization.Attributes{
		APIName:    "ListOpenWorkflowExecutions",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListOpenWorkflowExecutions(ctx, request)
}

// ListWorkflowExecutions API call
func (a *AccessControlledWorkflowHandler) ListWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
) (*types.ListWorkflowExecutionsResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendListWorkflowExecutionsScope, request)

	attr := &authorization.Attributes{
		APIName:    "ListWorkflowExecutions",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListWorkflowExecutions(ctx, request)
}

// PollForActivityTask API call
func (a *AccessControlledWorkflowHandler) PollForActivityTask(
	ctx context.Context,
	request *types.PollForActivityTaskRequest,
) (*types.PollForActivityTaskResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendPollForActivityTaskScope, request)

	attr := &authorization.Attributes{
		APIName:    "PollForActivityTask",
		DomainName: request.GetDomain(),
		TaskList:   request.TaskList,
		Permission: authorization.PermissionWrite,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.PollForActivityTask(ctx, request)
}

// PollForDecisionTask API call
func (a *AccessControlledWorkflowHandler) PollForDecisionTask(
	ctx context.Context,
	request *types.PollForDecisionTaskRequest,
) (*types.PollForDecisionTaskResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendPollForDecisionTaskScope, request)

	attr := &authorization.Attributes{
		APIName:    "PollForDecisionTask",
		DomainName: request.GetDomain(),
		TaskList:   request.TaskList,
		Permission: authorization.PermissionWrite,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.PollForDecisionTask(ctx, request)
}

// QueryWorkflow API call
func (a *AccessControlledWorkflowHandler) QueryWorkflow(
	ctx context.Context,
	request *types.QueryWorkflowRequest,
) (*types.QueryWorkflowResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendQueryWorkflowScope, request)

	attr := &authorization.Attributes{
		APIName:    "QueryWorkflow",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.QueryWorkflow(ctx, request)
}

// GetClusterInfo API call
func (a *AccessControlledWorkflowHandler) GetClusterInfo(
	ctx context.Context,
) (*types.ClusterInfo, error) {
	return a.frontendHandler.GetClusterInfo(ctx)
}

// RecordActivityTaskHeartbeat API call
func (a *AccessControlledWorkflowHandler) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatRequest,
) (*types.RecordActivityTaskHeartbeatResponse, error) {
	// TODO(vancexu): add auth check for service API
	return a.frontendHandler.RecordActivityTaskHeartbeat(ctx, request)
}

// RecordActivityTaskHeartbeatByID API call
func (a *AccessControlledWorkflowHandler) RecordActivityTaskHeartbeatByID(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatByIDRequest,
) (*types.RecordActivityTaskHeartbeatResponse, error) {
	return a.frontendHandler.RecordActivityTaskHeartbeatByID(ctx, request)
}

// RegisterDomain API call
func (a *AccessControlledWorkflowHandler) RegisterDomain(
	ctx context.Context,
	request *types.RegisterDomainRequest,
) error {

	scope := a.getMetricsScopeWithDomainName(metrics.FrontendRegisterDomainScope, request.GetName())

	attr := &authorization.Attributes{
		APIName:    "RegisterDomain",
		DomainName: request.GetName(),
		Permission: authorization.PermissionAdmin,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}

	return a.frontendHandler.RegisterDomain(ctx, request)
}

// RequestCancelWorkflowExecution API call
func (a *AccessControlledWorkflowHandler) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *types.RequestCancelWorkflowExecutionRequest,
) error {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendRequestCancelWorkflowExecutionScope, request)

	attr := &authorization.Attributes{
		APIName:    "RequestCancelWorkflowExecution",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionWrite,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}

	return a.frontendHandler.RequestCancelWorkflowExecution(ctx, request)
}

// ResetStickyTaskList API call
func (a *AccessControlledWorkflowHandler) ResetStickyTaskList(
	ctx context.Context,
	request *types.ResetStickyTaskListRequest,
) (*types.ResetStickyTaskListResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendResetStickyTaskListScope, request)

	attr := &authorization.Attributes{
		APIName:    "ResetStickyTaskList",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionWrite,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ResetStickyTaskList(ctx, request)
}

// ResetWorkflowExecution API call
func (a *AccessControlledWorkflowHandler) ResetWorkflowExecution(
	ctx context.Context,
	request *types.ResetWorkflowExecutionRequest,
) (*types.ResetWorkflowExecutionResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendResetWorkflowExecutionScope, request)

	attr := &authorization.Attributes{
		APIName:    "ResetWorkflowExecution",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionWrite,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ResetWorkflowExecution(ctx, request)
}

// RespondActivityTaskCanceled API call
func (a *AccessControlledWorkflowHandler) RespondActivityTaskCanceled(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledRequest,
) error {
	return a.frontendHandler.RespondActivityTaskCanceled(ctx, request)
}

// RespondActivityTaskCanceledByID API call
func (a *AccessControlledWorkflowHandler) RespondActivityTaskCanceledByID(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledByIDRequest,
) error {
	return a.frontendHandler.RespondActivityTaskCanceledByID(ctx, request)
}

// RespondActivityTaskCompleted API call
func (a *AccessControlledWorkflowHandler) RespondActivityTaskCompleted(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedRequest,
) error {
	return a.frontendHandler.RespondActivityTaskCompleted(ctx, request)
}

// RespondActivityTaskCompletedByID API call
func (a *AccessControlledWorkflowHandler) RespondActivityTaskCompletedByID(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedByIDRequest,
) error {
	return a.frontendHandler.RespondActivityTaskCompletedByID(ctx, request)
}

// RespondActivityTaskFailed API call
func (a *AccessControlledWorkflowHandler) RespondActivityTaskFailed(
	ctx context.Context,
	request *types.RespondActivityTaskFailedRequest,
) error {
	return a.frontendHandler.RespondActivityTaskFailed(ctx, request)
}

// RespondActivityTaskFailedByID API call
func (a *AccessControlledWorkflowHandler) RespondActivityTaskFailedByID(
	ctx context.Context,
	request *types.RespondActivityTaskFailedByIDRequest,
) error {
	return a.frontendHandler.RespondActivityTaskFailedByID(ctx, request)
}

// RespondDecisionTaskCompleted API call
func (a *AccessControlledWorkflowHandler) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *types.RespondDecisionTaskCompletedRequest,
) (*types.RespondDecisionTaskCompletedResponse, error) {
	return a.frontendHandler.RespondDecisionTaskCompleted(ctx, request)
}

// RespondDecisionTaskFailed API call
func (a *AccessControlledWorkflowHandler) RespondDecisionTaskFailed(
	ctx context.Context,
	request *types.RespondDecisionTaskFailedRequest,
) error {
	return a.frontendHandler.RespondDecisionTaskFailed(ctx, request)
}

// RespondQueryTaskCompleted API call
func (a *AccessControlledWorkflowHandler) RespondQueryTaskCompleted(
	ctx context.Context,
	request *types.RespondQueryTaskCompletedRequest,
) error {
	return a.frontendHandler.RespondQueryTaskCompleted(ctx, request)
}

// ScanWorkflowExecutions API call
func (a *AccessControlledWorkflowHandler) ScanWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
) (*types.ListWorkflowExecutionsResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendScanWorkflowExecutionsScope, request)

	attr := &authorization.Attributes{
		APIName:    "ScanWorkflowExecutions",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ScanWorkflowExecutions(ctx, request)
}

// SignalWithStartWorkflowExecution API call
func (a *AccessControlledWorkflowHandler) SignalWithStartWorkflowExecution(
	ctx context.Context,
	request *types.SignalWithStartWorkflowExecutionRequest,
) (*types.StartWorkflowExecutionResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendSignalWithStartWorkflowExecutionScope, request)

	attr := &authorization.Attributes{
		APIName:      "SignalWithStartWorkflowExecution",
		DomainName:   request.GetDomain(),
		Permission:   authorization.PermissionWrite,
		WorkflowType: request.WorkflowType,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.SignalWithStartWorkflowExecution(ctx, request)
}

// SignalWorkflowExecution API call
func (a *AccessControlledWorkflowHandler) SignalWorkflowExecution(
	ctx context.Context,
	request *types.SignalWorkflowExecutionRequest,
) error {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendSignalWorkflowExecutionScope, request)

	attr := &authorization.Attributes{
		APIName:    "SignalWorkflowExecution",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionWrite,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}

	return a.frontendHandler.SignalWorkflowExecution(ctx, request)
}

// StartWorkflowExecution API call
func (a *AccessControlledWorkflowHandler) StartWorkflowExecution(
	ctx context.Context,
	request *types.StartWorkflowExecutionRequest,
) (*types.StartWorkflowExecutionResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendStartWorkflowExecutionScope, request)

	attr := &authorization.Attributes{
		APIName:      "StartWorkflowExecution",
		DomainName:   request.GetDomain(),
		Permission:   authorization.PermissionWrite,
		WorkflowType: request.WorkflowType,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.StartWorkflowExecution(ctx, request)
}

// TerminateWorkflowExecution API call
func (a *AccessControlledWorkflowHandler) TerminateWorkflowExecution(
	ctx context.Context,
	request *types.TerminateWorkflowExecutionRequest,
) error {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendTerminateWorkflowExecutionScope, request)

	attr := &authorization.Attributes{
		APIName:    "TerminateWorkflowExecution",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionWrite,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return err
	}
	if !isAuthorized {
		return errUnauthorized
	}

	return a.frontendHandler.TerminateWorkflowExecution(ctx, request)
}

// ListTaskListPartitions API call
func (a *AccessControlledWorkflowHandler) ListTaskListPartitions(
	ctx context.Context,
	request *types.ListTaskListPartitionsRequest,
) (*types.ListTaskListPartitionsResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendListTaskListPartitionsScope, request)

	attr := &authorization.Attributes{
		APIName:    "ListTaskListPartitions",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListTaskListPartitions(ctx, request)
}

// GetTaskListsByDomain API call
func (a *AccessControlledWorkflowHandler) GetTaskListsByDomain(
	ctx context.Context,
	request *types.GetTaskListsByDomainRequest,
) (*types.GetTaskListsByDomainResponse, error) {

	scope := a.getMetricsScopeWithDomain(metrics.FrontendGetTaskListsByDomainScope, request)

	attr := &authorization.Attributes{
		APIName:    "GetTaskListsByDomain",
		DomainName: request.GetDomain(),
		Permission: authorization.PermissionRead,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.GetTaskListsByDomain(ctx, request)
}

// UpdateDomain API call
func (a *AccessControlledWorkflowHandler) UpdateDomain(
	ctx context.Context,
	request *types.UpdateDomainRequest,
) (*types.UpdateDomainResponse, error) {

	scope := a.getMetricsScopeWithDomainName(metrics.FrontendUpdateDomainScope, request.GetName())

	attr := &authorization.Attributes{
		APIName:    "UpdateDomain",
		DomainName: request.GetName(),
		Permission: authorization.PermissionAdmin,
	}
	isAuthorized, err := a.isAuthorized(ctx, attr, scope)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, errUnauthorized
	}

	return a.frontendHandler.UpdateDomain(ctx, request)
}

func (a *AccessControlledWorkflowHandler) isAuthorized(
	ctx context.Context,
	attr *authorization.Attributes,
	scope metrics.Scope,
) (bool, error) {
	sw := scope.StartTimer(metrics.CadenceAuthorizationLatency)
	defer sw.Stop()

	result, err := a.authorizer.Authorize(ctx, attr)
	if err != nil {
		scope.IncCounter(metrics.CadenceErrAuthorizeFailedCounter)
		return false, err
	}
	isAuth := result.Decision == authorization.DecisionAllow
	if !isAuth {
		scope.IncCounter(metrics.CadenceErrUnauthorizedCounter)
	}
	return isAuth, nil
}

// getMetricsScopeWithDomain return metrics scope with domain tag
func (a *AccessControlledWorkflowHandler) getMetricsScopeWithDomain(
	scope int,
	d domainGetter,
) metrics.Scope {
	return getMetricsScopeWithDomain(scope, d, a.GetMetricsClient())
}

func getMetricsScopeWithDomain(
	scope int,
	d domainGetter,
	metricsClient metrics.Client,
) metrics.Scope {
	var metricsScope metrics.Scope
	if d != nil {
		metricsScope = metricsClient.Scope(scope).Tagged(metrics.DomainTag(d.GetDomain()))
	} else {
		metricsScope = metricsClient.Scope(scope).Tagged(metrics.DomainUnknownTag())
	}
	return metricsScope
}

// getMetricsScopeWithDomainName is for XXXDomain APIs, whose request is not domainGetter
func (a *AccessControlledWorkflowHandler) getMetricsScopeWithDomainName(
	scope int,
	domainName string,
) metrics.Scope {
	return a.GetMetricsClient().Scope(scope).Tagged(metrics.DomainTag(domainName))
}
