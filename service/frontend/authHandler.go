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

	"github.com/uber/cadence/.gen/go/cadence/workflowserviceserver"
	"github.com/uber/cadence/.gen/go/health"
	"github.com/uber/cadence/.gen/go/health/metaserver"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/auth"
	"github.com/uber/cadence/common/resource"
)

// TODO(vancexu): add metrics

const (
	// ActionCommon is for authority implementation to handle auth for common operations
	ActionCommon = "common"
	// ActionAdmin is for authority implementation to handle auth for admin only operations
	ActionAdmin = "admin"
)

// AuthHandlerImpl frontend handler wrapper for authentication and authorization
type AuthHandlerImpl struct {
	resource.Resource

	frontendHandler workflowserviceserver.Interface
	authority       auth.Authority

	startFn func()
	stopFn  func()
}

var _ workflowserviceserver.Interface = (*AuthHandlerImpl)(nil)

// NewAuthHandlerImpl creates frontend handler with authentication support
func NewAuthHandlerImpl(wfHandler *DCRedirectionHandlerImpl, authority auth.Authority) *AuthHandlerImpl {
	if authority == nil {
		authority = auth.NewNopAuthority()
	}

	return &AuthHandlerImpl{
		Resource:        wfHandler.Resource,
		frontendHandler: wfHandler,
		authority:       authority,
		startFn:         func() { wfHandler.Start() },
		stopFn:          func() { wfHandler.Stop() },
	}
}

// TODO(vancexu): refactor frontend handler

// RegisterHandler register this handler, must be called before Start()
func (a *AuthHandlerImpl) RegisterHandler() {
	a.GetDispatcher().Register(workflowserviceserver.New(a))
	a.GetDispatcher().Register(metaserver.New(a))
}

// Health callback for for health check
func (a *AuthHandlerImpl) Health(ctx context.Context) (*health.HealthStatus, error) {
	hs := &health.HealthStatus{Ok: true, Msg: common.StringPtr("auth is good")}
	return hs, nil
}

// Start starts the handler
func (a *AuthHandlerImpl) Start() {
	a.startFn()
}

// Stop stops the handler
func (a *AuthHandlerImpl) Stop() {
	a.stopFn()
}

// CountWorkflowExecutions API call
func (a *AuthHandlerImpl) CountWorkflowExecutions(
	ctx context.Context,
	request *shared.CountWorkflowExecutionsRequest,
) (*shared.CountWorkflowExecutionsResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}
	return a.frontendHandler.CountWorkflowExecutions(ctx, request)
}

// DeprecateDomain API call
func (a *AuthHandlerImpl) DeprecateDomain(
	ctx context.Context,
	request *shared.DeprecateDomainRequest,
) error {

	isAllowed, err := a.isAdmin(ctx)
	if err != nil {
		return err
	}
	if !isAllowed {
		return errUnauthorized
	}

	return a.frontendHandler.DeprecateDomain(ctx, request)
}

// DescribeDomain API call
func (a *AuthHandlerImpl) DescribeDomain(
	ctx context.Context,
	request *shared.DescribeDomainRequest,
) (*shared.DescribeDomainResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetName())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.DescribeDomain(ctx, request)
}

// DescribeTaskList API call
func (a *AuthHandlerImpl) DescribeTaskList(
	ctx context.Context,
	request *shared.DescribeTaskListRequest,
) (*shared.DescribeTaskListResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.DescribeTaskList(ctx, request)
}

// DescribeWorkflowExecution API call
func (a *AuthHandlerImpl) DescribeWorkflowExecution(
	ctx context.Context,
	request *shared.DescribeWorkflowExecutionRequest,
) (*shared.DescribeWorkflowExecutionResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.DescribeWorkflowExecution(ctx, request)
}

// GetDomainReplicationMessages API call
func (a *AuthHandlerImpl) GetDomainReplicationMessages(
	ctx context.Context,
	request *replicator.GetDomainReplicationMessagesRequest,
) (*replicator.GetDomainReplicationMessagesResponse, error) {
	return a.frontendHandler.GetDomainReplicationMessages(ctx, request)
}

// GetReplicationMessages API call
func (a *AuthHandlerImpl) GetReplicationMessages(
	ctx context.Context,
	request *replicator.GetReplicationMessagesRequest,
) (*replicator.GetReplicationMessagesResponse, error) {
	return a.frontendHandler.GetReplicationMessages(ctx, request)
}

// GetSearchAttributes API call
func (a *AuthHandlerImpl) GetSearchAttributes(
	ctx context.Context,
) (*shared.GetSearchAttributesResponse, error) {
	return a.frontendHandler.GetSearchAttributes(ctx)
}

// GetWorkflowExecutionHistory API call
func (a *AuthHandlerImpl) GetWorkflowExecutionHistory(
	ctx context.Context,
	request *shared.GetWorkflowExecutionHistoryRequest,
) (*shared.GetWorkflowExecutionHistoryResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.GetWorkflowExecutionHistory(ctx, request)
}

// ListArchivedWorkflowExecutions API call
func (a *AuthHandlerImpl) ListArchivedWorkflowExecutions(
	ctx context.Context,
	request *shared.ListArchivedWorkflowExecutionsRequest,
) (*shared.ListArchivedWorkflowExecutionsResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListArchivedWorkflowExecutions(ctx, request)
}

// ListClosedWorkflowExecutions API call
func (a *AuthHandlerImpl) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *shared.ListClosedWorkflowExecutionsRequest,
) (*shared.ListClosedWorkflowExecutionsResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListClosedWorkflowExecutions(ctx, request)
}

// ListDomains API call
func (a *AuthHandlerImpl) ListDomains(
	ctx context.Context,
	request *shared.ListDomainsRequest,
) (*shared.ListDomainsResponse, error) {

	isAllowed, err := a.isAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListDomains(ctx, request)
}

// ListOpenWorkflowExecutions API call
func (a *AuthHandlerImpl) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *shared.ListOpenWorkflowExecutionsRequest,
) (*shared.ListOpenWorkflowExecutionsResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListOpenWorkflowExecutions(ctx, request)
}

// ListWorkflowExecutions API call
func (a *AuthHandlerImpl) ListWorkflowExecutions(
	ctx context.Context,
	request *shared.ListWorkflowExecutionsRequest,
) (*shared.ListWorkflowExecutionsResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ListWorkflowExecutions(ctx, request)
}

// PollForActivityTask API call
func (a *AuthHandlerImpl) PollForActivityTask(
	ctx context.Context,
	request *shared.PollForActivityTaskRequest,
) (*shared.PollForActivityTaskResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.PollForActivityTask(ctx, request)
}

// PollForDecisionTask API call
func (a *AuthHandlerImpl) PollForDecisionTask(
	ctx context.Context,
	request *shared.PollForDecisionTaskRequest,
) (*shared.PollForDecisionTaskResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.PollForDecisionTask(ctx, request)
}

// QueryWorkflow API call
func (a *AuthHandlerImpl) QueryWorkflow(
	ctx context.Context,
	request *shared.QueryWorkflowRequest,
) (*shared.QueryWorkflowResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.QueryWorkflow(ctx, request)
}

// ReapplyEvents API call
func (a *AuthHandlerImpl) ReapplyEvents(
	ctx context.Context,
	request *shared.ReapplyEventsRequest,
) error {
	return a.frontendHandler.ReapplyEvents(ctx, request)
}

// RecordActivityTaskHeartbeat API call
func (a *AuthHandlerImpl) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *shared.RecordActivityTaskHeartbeatRequest,
) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	// TODO(vancexu): add auth check for service API
	return a.frontendHandler.RecordActivityTaskHeartbeat(ctx, request)
}

// RecordActivityTaskHeartbeatByID API call
func (a *AuthHandlerImpl) RecordActivityTaskHeartbeatByID(
	ctx context.Context,
	request *shared.RecordActivityTaskHeartbeatByIDRequest,
) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	return a.frontendHandler.RecordActivityTaskHeartbeatByID(ctx, request)
}

// RegisterDomain API call
func (a *AuthHandlerImpl) RegisterDomain(
	ctx context.Context,
	request *shared.RegisterDomainRequest,
) error {

	isAllowed, err := a.isAdmin(ctx)
	if err != nil {
		return err
	}
	if !isAllowed {
		return errUnauthorized
	}

	return a.frontendHandler.RegisterDomain(ctx, request)
}

// RequestCancelWorkflowExecution API call
func (a *AuthHandlerImpl) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *shared.RequestCancelWorkflowExecutionRequest,
) error {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return err
	}
	if !isAllowed {
		return errUnauthorized
	}

	return a.frontendHandler.RequestCancelWorkflowExecution(ctx, request)
}

// ResetStickyTaskList API call
func (a *AuthHandlerImpl) ResetStickyTaskList(
	ctx context.Context,
	request *shared.ResetStickyTaskListRequest,
) (*shared.ResetStickyTaskListResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ResetStickyTaskList(ctx, request)
}

// ResetWorkflowExecution API call
func (a *AuthHandlerImpl) ResetWorkflowExecution(
	ctx context.Context,
	request *shared.ResetWorkflowExecutionRequest,
) (*shared.ResetWorkflowExecutionResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ResetWorkflowExecution(ctx, request)
}

// RespondActivityTaskCanceled API call
func (a *AuthHandlerImpl) RespondActivityTaskCanceled(
	ctx context.Context,
	request *shared.RespondActivityTaskCanceledRequest,
) error {
	return a.frontendHandler.RespondActivityTaskCanceled(ctx, request)
}

// RespondActivityTaskCanceledByID API call
func (a *AuthHandlerImpl) RespondActivityTaskCanceledByID(
	ctx context.Context,
	request *shared.RespondActivityTaskCanceledByIDRequest,
) error {
	return a.frontendHandler.RespondActivityTaskCanceledByID(ctx, request)
}

// RespondActivityTaskCompleted API call
func (a *AuthHandlerImpl) RespondActivityTaskCompleted(
	ctx context.Context,
	request *shared.RespondActivityTaskCompletedRequest,
) error {
	return a.frontendHandler.RespondActivityTaskCompleted(ctx, request)
}

// RespondActivityTaskCompletedByID API call
func (a *AuthHandlerImpl) RespondActivityTaskCompletedByID(
	ctx context.Context,
	request *shared.RespondActivityTaskCompletedByIDRequest,
) error {
	return a.frontendHandler.RespondActivityTaskCompletedByID(ctx, request)
}

// RespondActivityTaskFailed API call
func (a *AuthHandlerImpl) RespondActivityTaskFailed(
	ctx context.Context,
	request *shared.RespondActivityTaskFailedRequest,
) error {
	return a.frontendHandler.RespondActivityTaskFailed(ctx, request)
}

// RespondActivityTaskFailedByID API call
func (a *AuthHandlerImpl) RespondActivityTaskFailedByID(
	ctx context.Context,
	request *shared.RespondActivityTaskFailedByIDRequest,
) error {
	return a.frontendHandler.RespondActivityTaskFailedByID(ctx, request)
}

// RespondDecisionTaskCompleted API call
func (a *AuthHandlerImpl) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *shared.RespondDecisionTaskCompletedRequest,
) (*shared.RespondDecisionTaskCompletedResponse, error) {
	return a.frontendHandler.RespondDecisionTaskCompleted(ctx, request)
}

// RespondDecisionTaskFailed API call
func (a *AuthHandlerImpl) RespondDecisionTaskFailed(
	ctx context.Context,
	request *shared.RespondDecisionTaskFailedRequest,
) error {
	return a.frontendHandler.RespondDecisionTaskFailed(ctx, request)
}

// RespondQueryTaskCompleted API call
func (a *AuthHandlerImpl) RespondQueryTaskCompleted(
	ctx context.Context,
	request *shared.RespondQueryTaskCompletedRequest,
) error {
	return a.frontendHandler.RespondQueryTaskCompleted(ctx, request)
}

// ScanWorkflowExecutions API call
func (a *AuthHandlerImpl) ScanWorkflowExecutions(
	ctx context.Context,
	request *shared.ListWorkflowExecutionsRequest,
) (*shared.ListWorkflowExecutionsResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.ScanWorkflowExecutions(ctx, request)
}

// SignalWithStartWorkflowExecution API call
func (a *AuthHandlerImpl) SignalWithStartWorkflowExecution(
	ctx context.Context,
	request *shared.SignalWithStartWorkflowExecutionRequest,
) (*shared.StartWorkflowExecutionResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.SignalWithStartWorkflowExecution(ctx, request)
}

// SignalWorkflowExecution API call
func (a *AuthHandlerImpl) SignalWorkflowExecution(
	ctx context.Context,
	request *shared.SignalWorkflowExecutionRequest,
) error {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return err
	}
	if !isAllowed {
		return errUnauthorized
	}

	return a.frontendHandler.SignalWorkflowExecution(ctx, request)
}

// StartWorkflowExecution API call
func (a *AuthHandlerImpl) StartWorkflowExecution(
	ctx context.Context,
	request *shared.StartWorkflowExecutionRequest,
) (*shared.StartWorkflowExecutionResponse, error) {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.StartWorkflowExecution(ctx, request)
}

// TerminateWorkflowExecution API call
func (a *AuthHandlerImpl) TerminateWorkflowExecution(
	ctx context.Context,
	request *shared.TerminateWorkflowExecutionRequest,
) error {

	isAllowed, err := a.isAllowedForDomain(ctx, request.GetDomain())
	if err != nil {
		return err
	}
	if !isAllowed {
		return errUnauthorized
	}

	return a.frontendHandler.TerminateWorkflowExecution(ctx, request)
}

// UpdateDomain API call
func (a *AuthHandlerImpl) UpdateDomain(
	ctx context.Context,
	request *shared.UpdateDomainRequest,
) (*shared.UpdateDomainResponse, error) {

	isAllowed, err := a.isAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !isAllowed {
		return nil, errUnauthorized
	}

	return a.frontendHandler.UpdateDomain(ctx, request)
}

func (a *AuthHandlerImpl) isAllowedForDomain(
	ctx context.Context,
	domain string,
) (bool, error) {
	authParams := &auth.AuthorizationParams{
		Action:     ActionCommon,
		DomainName: domain,
	}
	result, err := a.authority.IsAuthorized(ctx, authParams)
	if err != nil {
		return false, err
	}
	return result.AuthorizationDecision == auth.DecisionAllow, nil
}

func (a *AuthHandlerImpl) isAdmin(ctx context.Context) (bool, error) {
	authParams := &auth.AuthorizationParams{
		Action: ActionAdmin,
	}
	result, err := a.authority.IsAuthorized(ctx, authParams)
	if err != nil {
		return false, err
	}
	return result.AuthorizationDecision == auth.DecisionAllow, nil
}
