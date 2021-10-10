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
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/types"
)

var _ Handler = (*ClusterRedirectionHandlerImpl)(nil)

type (
	// ClusterRedirectionHandlerImpl is simple wrapper over frontend service, doing redirection based on policy for global domains not being active in current cluster
	ClusterRedirectionHandlerImpl struct {
		resource.Resource

		currentClusterName string
		redirectionPolicy  ClusterRedirectionPolicy
		tokenSerializer    common.TaskTokenSerializer
		frontendHandler    Handler
	}
)

// NewClusterRedirectionHandler creates a frontend handler to handle cluster redirection for global domains not being active in current cluster
func NewClusterRedirectionHandler(
	wfHandler Handler,
	resource resource.Resource,
	config *Config,
	policy config.ClusterRedirectionPolicy,
) *ClusterRedirectionHandlerImpl {
	dcRedirectionPolicy := RedirectionPolicyGenerator(
		resource.GetClusterMetadata(),
		config,
		resource.GetDomainCache(),
		policy,
	)

	return &ClusterRedirectionHandlerImpl{
		Resource:           resource,
		currentClusterName: resource.GetClusterMetadata().GetCurrentClusterName(),
		redirectionPolicy:  dcRedirectionPolicy,
		tokenSerializer:    common.NewJSONTaskTokenSerializer(),
		frontendHandler:    wfHandler,
	}
}

// Health is for health check
func (handler *ClusterRedirectionHandlerImpl) Health(ctx context.Context) (*types.HealthStatus, error) {
	return handler.frontendHandler.Health(ctx)
}

// Domain APIs, domain APIs does not require redirection

// DeprecateDomain API call
func (handler *ClusterRedirectionHandlerImpl) DeprecateDomain(
	ctx context.Context,
	request *types.DeprecateDomainRequest,
) (retError error) {

	var cluster = handler.currentClusterName

	scope, startTime := handler.beforeCall(metrics.DCRedirectionDeprecateDomainScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	return handler.frontendHandler.DeprecateDomain(ctx, request)
}

// DescribeDomain API call
func (handler *ClusterRedirectionHandlerImpl) DescribeDomain(
	ctx context.Context,
	request *types.DescribeDomainRequest,
) (resp *types.DescribeDomainResponse, retError error) {

	var cluster = handler.currentClusterName

	scope, startTime := handler.beforeCall(metrics.DCRedirectionDescribeDomainScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	return handler.frontendHandler.DescribeDomain(ctx, request)
}

// ListDomains API call
func (handler *ClusterRedirectionHandlerImpl) ListDomains(
	ctx context.Context,
	request *types.ListDomainsRequest,
) (resp *types.ListDomainsResponse, retError error) {

	var cluster = handler.currentClusterName

	scope, startTime := handler.beforeCall(metrics.DCRedirectionListDomainsScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	return handler.frontendHandler.ListDomains(ctx, request)
}

// RegisterDomain API call
func (handler *ClusterRedirectionHandlerImpl) RegisterDomain(
	ctx context.Context,
	request *types.RegisterDomainRequest,
) (retError error) {

	var cluster = handler.currentClusterName

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRegisterDomainScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	return handler.frontendHandler.RegisterDomain(ctx, request)
}

// UpdateDomain API call
func (handler *ClusterRedirectionHandlerImpl) UpdateDomain(
	ctx context.Context,
	request *types.UpdateDomainRequest,
) (resp *types.UpdateDomainResponse, retError error) {

	var cluster = handler.currentClusterName

	scope, startTime := handler.beforeCall(metrics.DCRedirectionUpdateDomainScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	return handler.frontendHandler.UpdateDomain(ctx, request)
}

// Other APIs

// DescribeTaskList API call
func (handler *ClusterRedirectionHandlerImpl) DescribeTaskList(
	ctx context.Context,
	request *types.DescribeTaskListRequest,
) (resp *types.DescribeTaskListResponse, retError error) {

	var apiName = "DescribeTaskList"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionDescribeTaskListScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.DescribeTaskList(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.DescribeTaskList(ctx, request)
		}
		return err
	})

	return resp, err
}

// DescribeWorkflowExecution API call
func (handler *ClusterRedirectionHandlerImpl) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.DescribeWorkflowExecutionRequest,
) (resp *types.DescribeWorkflowExecutionResponse, retError error) {

	var apiName = "DescribeWorkflowExecution"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionDescribeWorkflowExecutionScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.DescribeWorkflowExecution(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.DescribeWorkflowExecution(ctx, request)
		}
		return err
	})

	return resp, err
}

// GetWorkflowExecutionHistory API call
func (handler *ClusterRedirectionHandlerImpl) GetWorkflowExecutionHistory(
	ctx context.Context,
	request *types.GetWorkflowExecutionHistoryRequest,
) (resp *types.GetWorkflowExecutionHistoryResponse, retError error) {

	var apiName = "GetWorkflowExecutionHistory"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionGetWorkflowExecutionHistoryScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.GetWorkflowExecutionHistory(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.GetWorkflowExecutionHistory(ctx, request)
		}
		return err
	})

	return resp, err
}

// ListArchivedWorkflowExecutions API call
func (handler *ClusterRedirectionHandlerImpl) ListArchivedWorkflowExecutions(
	ctx context.Context,
	request *types.ListArchivedWorkflowExecutionsRequest,
) (resp *types.ListArchivedWorkflowExecutionsResponse, retError error) {

	var apiName = "ListArchivedWorkflowExecutions"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionListArchivedWorkflowExecutionsScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.ListArchivedWorkflowExecutions(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.ListArchivedWorkflowExecutions(ctx, request)
		}
		return err
	})

	return resp, err
}

// ListClosedWorkflowExecutions API call
func (handler *ClusterRedirectionHandlerImpl) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *types.ListClosedWorkflowExecutionsRequest,
) (resp *types.ListClosedWorkflowExecutionsResponse, retError error) {

	var apiName = "ListClosedWorkflowExecutions"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionListClosedWorkflowExecutionsScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.ListClosedWorkflowExecutions(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.ListClosedWorkflowExecutions(ctx, request)
		}
		return err
	})

	return resp, err
}

// ListOpenWorkflowExecutions API call
func (handler *ClusterRedirectionHandlerImpl) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *types.ListOpenWorkflowExecutionsRequest,
) (resp *types.ListOpenWorkflowExecutionsResponse, retError error) {

	var apiName = "ListOpenWorkflowExecutions"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionListOpenWorkflowExecutionsScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.ListOpenWorkflowExecutions(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.ListOpenWorkflowExecutions(ctx, request)
		}
		return err
	})

	return resp, err
}

// ListWorkflowExecutions API call
func (handler *ClusterRedirectionHandlerImpl) ListWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
) (resp *types.ListWorkflowExecutionsResponse, retError error) {

	var apiName = "ListWorkflowExecutions"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionListWorkflowExecutionsScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.ListWorkflowExecutions(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.ListWorkflowExecutions(ctx, request)
		}
		return err
	})

	return resp, err
}

// ScanWorkflowExecutions API call
func (handler *ClusterRedirectionHandlerImpl) ScanWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
) (resp *types.ListWorkflowExecutionsResponse, retError error) {

	var apiName = "ScanWorkflowExecutions"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionScanWorkflowExecutionsScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()
	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.ScanWorkflowExecutions(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.ScanWorkflowExecutions(ctx, request)
		}
		return err
	})

	return resp, err
}

// CountWorkflowExecutions API call
func (handler *ClusterRedirectionHandlerImpl) CountWorkflowExecutions(
	ctx context.Context,
	request *types.CountWorkflowExecutionsRequest,
) (resp *types.CountWorkflowExecutionsResponse, retError error) {

	var apiName = "CountWorkflowExecutions"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionCountWorkflowExecutionsScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.CountWorkflowExecutions(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.CountWorkflowExecutions(ctx, request)
		}
		return err
	})

	return resp, err
}

// GetSearchAttributes API call
func (handler *ClusterRedirectionHandlerImpl) GetSearchAttributes(
	ctx context.Context,
) (resp *types.GetSearchAttributesResponse, retError error) {

	var cluster = handler.currentClusterName

	scope, startTime := handler.beforeCall(metrics.DCRedirectionGetSearchAttributesScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	return handler.frontendHandler.GetSearchAttributes(ctx)
}

// PollForActivityTask API call
func (handler *ClusterRedirectionHandlerImpl) PollForActivityTask(
	ctx context.Context,
	request *types.PollForActivityTaskRequest,
) (resp *types.PollForActivityTaskResponse, retError error) {

	var apiName = "PollForActivityTask"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionPollForActivityTaskScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.PollForActivityTask(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.PollForActivityTask(ctx, request)
		}
		return err
	})

	return resp, err
}

// PollForDecisionTask API call
func (handler *ClusterRedirectionHandlerImpl) PollForDecisionTask(
	ctx context.Context,
	request *types.PollForDecisionTaskRequest,
) (resp *types.PollForDecisionTaskResponse, retError error) {

	var apiName = "PollForDecisionTask"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionPollForDecisionTaskScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.PollForDecisionTask(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.PollForDecisionTask(ctx, request)
		}
		return err
	})

	return resp, err
}

// QueryWorkflow API call
func (handler *ClusterRedirectionHandlerImpl) QueryWorkflow(
	ctx context.Context,
	request *types.QueryWorkflowRequest,
) (resp *types.QueryWorkflowResponse, retError error) {

	var apiName = "QueryWorkflow"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionQueryWorkflowScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.QueryWorkflow(ctx, request)
		default:
			// Only autofoward consistent queries, this is done for two reasons:
			// 1. Query is meant to be fast, autoforwarding all queries will increase latency.
			// 2. If eventual consistency was requested then the results from running out of local dc will be fine.
			if request.GetQueryConsistencyLevel() == types.QueryConsistencyLevelStrong {
				remoteClient := handler.GetRemoteFrontendClient(targetDC)
				resp, err = remoteClient.QueryWorkflow(ctx, request)
			} else {
				resp, err = handler.frontendHandler.QueryWorkflow(ctx, request)
			}
		}
		return err
	})

	return resp, err
}

// RecordActivityTaskHeartbeat API call
func (handler *ClusterRedirectionHandlerImpl) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatRequest,
) (resp *types.RecordActivityTaskHeartbeatResponse, retError error) {

	var apiName = "RecordActivityTaskHeartbeat"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRecordActivityTaskHeartbeatScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	token, err := handler.tokenSerializer.Deserialize(request.TaskToken)
	if err != nil {
		return nil, err
	}

	err = handler.redirectionPolicy.WithDomainIDRedirect(ctx, token.DomainID, apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.RecordActivityTaskHeartbeat(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.RecordActivityTaskHeartbeat(ctx, request)
		}
		return err
	})

	return resp, err
}

// RecordActivityTaskHeartbeatByID API call
func (handler *ClusterRedirectionHandlerImpl) RecordActivityTaskHeartbeatByID(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatByIDRequest,
) (resp *types.RecordActivityTaskHeartbeatResponse, retError error) {

	var apiName = "RecordActivityTaskHeartbeatByID"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRecordActivityTaskHeartbeatByIDScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.RecordActivityTaskHeartbeatByID(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.RecordActivityTaskHeartbeatByID(ctx, request)
		}
		return err
	})

	return resp, err
}

// RequestCancelWorkflowExecution API call
func (handler *ClusterRedirectionHandlerImpl) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *types.RequestCancelWorkflowExecutionRequest,
) (retError error) {

	var apiName = "RequestCancelWorkflowExecution"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRequestCancelWorkflowExecutionScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.RequestCancelWorkflowExecution(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.RequestCancelWorkflowExecution(ctx, request)
		}
		return err
	})

	return err
}

// ResetStickyTaskList API call
func (handler *ClusterRedirectionHandlerImpl) ResetStickyTaskList(
	ctx context.Context,
	request *types.ResetStickyTaskListRequest,
) (resp *types.ResetStickyTaskListResponse, retError error) {

	var apiName = "ResetStickyTaskList"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionResetStickyTaskListScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.ResetStickyTaskList(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.ResetStickyTaskList(ctx, request)
		}
		return err
	})

	return resp, err
}

// ResetWorkflowExecution API call
func (handler *ClusterRedirectionHandlerImpl) ResetWorkflowExecution(
	ctx context.Context,
	request *types.ResetWorkflowExecutionRequest,
) (resp *types.ResetWorkflowExecutionResponse, retError error) {

	var apiName = "ResetWorkflowExecution"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionResetWorkflowExecutionScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.ResetWorkflowExecution(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.ResetWorkflowExecution(ctx, request)
		}
		return err
	})

	return resp, err
}

// RespondActivityTaskCanceled API call
func (handler *ClusterRedirectionHandlerImpl) RespondActivityTaskCanceled(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledRequest,
) (retError error) {

	var apiName = "RespondActivityTaskCanceled"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRespondActivityTaskCanceledScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	token, err := handler.tokenSerializer.Deserialize(request.TaskToken)
	if err != nil {
		return err
	}

	err = handler.redirectionPolicy.WithDomainIDRedirect(ctx, token.DomainID, apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.RespondActivityTaskCanceled(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.RespondActivityTaskCanceled(ctx, request)
		}
		return err
	})

	return err
}

// RespondActivityTaskCanceledByID API call
func (handler *ClusterRedirectionHandlerImpl) RespondActivityTaskCanceledByID(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledByIDRequest,
) (retError error) {

	var apiName = "RespondActivityTaskCanceledByID"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRespondActivityTaskCanceledByIDScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.RespondActivityTaskCanceledByID(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.RespondActivityTaskCanceledByID(ctx, request)
		}
		return err
	})

	return err
}

// RespondActivityTaskCompleted API call
func (handler *ClusterRedirectionHandlerImpl) RespondActivityTaskCompleted(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedRequest,
) (retError error) {

	var apiName = "RespondActivityTaskCompleted"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRespondActivityTaskCompletedScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	token, err := handler.tokenSerializer.Deserialize(request.TaskToken)
	if err != nil {
		return err
	}

	err = handler.redirectionPolicy.WithDomainIDRedirect(ctx, token.DomainID, apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.RespondActivityTaskCompleted(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.RespondActivityTaskCompleted(ctx, request)
		}
		return err
	})

	return err
}

// RespondActivityTaskCompletedByID API call
func (handler *ClusterRedirectionHandlerImpl) RespondActivityTaskCompletedByID(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedByIDRequest,
) (retError error) {

	var apiName = "RespondActivityTaskCompletedByID"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRespondActivityTaskCompletedByIDScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.RespondActivityTaskCompletedByID(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.RespondActivityTaskCompletedByID(ctx, request)
		}
		return err
	})

	return err
}

// RespondActivityTaskFailed API call
func (handler *ClusterRedirectionHandlerImpl) RespondActivityTaskFailed(
	ctx context.Context,
	request *types.RespondActivityTaskFailedRequest,
) (retError error) {

	var apiName = "RespondActivityTaskFailed"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRespondActivityTaskFailedScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	token, err := handler.tokenSerializer.Deserialize(request.TaskToken)
	if err != nil {
		return err
	}

	err = handler.redirectionPolicy.WithDomainIDRedirect(ctx, token.DomainID, apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.RespondActivityTaskFailed(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.RespondActivityTaskFailed(ctx, request)
		}
		return err
	})

	return err
}

// RespondActivityTaskFailedByID API call
func (handler *ClusterRedirectionHandlerImpl) RespondActivityTaskFailedByID(
	ctx context.Context,
	request *types.RespondActivityTaskFailedByIDRequest,
) (retError error) {

	var apiName = "RespondActivityTaskFailedByID"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRespondActivityTaskFailedByIDScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.RespondActivityTaskFailedByID(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.RespondActivityTaskFailedByID(ctx, request)
		}
		return err
	})

	return err
}

// RespondDecisionTaskCompleted API call
func (handler *ClusterRedirectionHandlerImpl) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *types.RespondDecisionTaskCompletedRequest,
) (resp *types.RespondDecisionTaskCompletedResponse, retError error) {

	var apiName = "RespondDecisionTaskCompleted"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRespondDecisionTaskCompletedScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	token, err := handler.tokenSerializer.Deserialize(request.TaskToken)
	if err != nil {
		return nil, err
	}

	err = handler.redirectionPolicy.WithDomainIDRedirect(ctx, token.DomainID, apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.RespondDecisionTaskCompleted(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.RespondDecisionTaskCompleted(ctx, request)
		}
		return err
	})

	return resp, err
}

// RespondDecisionTaskFailed API call
func (handler *ClusterRedirectionHandlerImpl) RespondDecisionTaskFailed(
	ctx context.Context,
	request *types.RespondDecisionTaskFailedRequest,
) (retError error) {

	var apiName = "RespondDecisionTaskFailed"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRespondDecisionTaskFailedScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	token, err := handler.tokenSerializer.Deserialize(request.TaskToken)
	if err != nil {
		return err
	}

	err = handler.redirectionPolicy.WithDomainIDRedirect(ctx, token.DomainID, apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.RespondDecisionTaskFailed(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.RespondDecisionTaskFailed(ctx, request)
		}
		return err
	})

	return err
}

// RespondQueryTaskCompleted API call
func (handler *ClusterRedirectionHandlerImpl) RespondQueryTaskCompleted(
	ctx context.Context,
	request *types.RespondQueryTaskCompletedRequest,
) (retError error) {

	var apiName = "RespondQueryTaskCompleted"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionRespondQueryTaskCompletedScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	token, err := handler.tokenSerializer.DeserializeQueryTaskToken(request.TaskToken)
	if err != nil {
		return err
	}

	err = handler.redirectionPolicy.WithDomainIDRedirect(ctx, token.DomainID, apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.RespondQueryTaskCompleted(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.RespondQueryTaskCompleted(ctx, request)
		}
		return err
	})

	return err
}

// SignalWithStartWorkflowExecution API call
func (handler *ClusterRedirectionHandlerImpl) SignalWithStartWorkflowExecution(
	ctx context.Context,
	request *types.SignalWithStartWorkflowExecutionRequest,
) (resp *types.StartWorkflowExecutionResponse, retError error) {

	var apiName = "SignalWithStartWorkflowExecution"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionSignalWithStartWorkflowExecutionScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.SignalWithStartWorkflowExecution(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.SignalWithStartWorkflowExecution(ctx, request)
		}
		return err
	})

	return resp, err
}

// SignalWorkflowExecution API call
func (handler *ClusterRedirectionHandlerImpl) SignalWorkflowExecution(
	ctx context.Context,
	request *types.SignalWorkflowExecutionRequest,
) (retError error) {

	var apiName = "SignalWorkflowExecution"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionSignalWorkflowExecutionScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.SignalWorkflowExecution(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.SignalWorkflowExecution(ctx, request)
		}
		return err
	})
	return err
}

// StartWorkflowExecution API call
func (handler *ClusterRedirectionHandlerImpl) StartWorkflowExecution(
	ctx context.Context,
	request *types.StartWorkflowExecutionRequest,
) (resp *types.StartWorkflowExecutionResponse, retError error) {

	var apiName = "StartWorkflowExecution"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionStartWorkflowExecutionScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.StartWorkflowExecution(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.StartWorkflowExecution(ctx, request)
		}
		return err
	})

	return resp, err
}

// TerminateWorkflowExecution API call
func (handler *ClusterRedirectionHandlerImpl) TerminateWorkflowExecution(
	ctx context.Context,
	request *types.TerminateWorkflowExecutionRequest,
) (retError error) {

	var apiName = "TerminateWorkflowExecution"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionTerminateWorkflowExecutionScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			err = handler.frontendHandler.TerminateWorkflowExecution(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			err = remoteClient.TerminateWorkflowExecution(ctx, request)
		}
		return err
	})

	return err
}

// ListTaskListPartitions API call
func (handler *ClusterRedirectionHandlerImpl) ListTaskListPartitions(
	ctx context.Context,
	request *types.ListTaskListPartitionsRequest,
) (resp *types.ListTaskListPartitionsResponse, retError error) {

	var apiName = "ListTaskListPartitions"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionListTaskListPartitionsScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.ListTaskListPartitions(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.ListTaskListPartitions(ctx, request)
		}
		return err
	})

	return resp, err
}

// GetTaskListsByDomain API call
func (handler *ClusterRedirectionHandlerImpl) GetTaskListsByDomain(
	ctx context.Context,
	request *types.GetTaskListsByDomainRequest,
) (resp *types.GetTaskListsByDomainResponse, retError error) {

	var apiName = "GetTaskListsByDomain"
	var err error
	var cluster string

	scope, startTime := handler.beforeCall(metrics.DCRedirectionGetTaskListsByDomainScope)
	defer func() {
		handler.afterCall(scope, startTime, cluster, &retError)
	}()

	err = handler.redirectionPolicy.WithDomainNameRedirect(ctx, request.GetDomain(), apiName, func(targetDC string) error {
		cluster = targetDC
		switch {
		case targetDC == handler.currentClusterName:
			resp, err = handler.frontendHandler.GetTaskListsByDomain(ctx, request)
		default:
			remoteClient := handler.GetRemoteFrontendClient(targetDC)
			resp, err = remoteClient.GetTaskListsByDomain(ctx, request)
		}
		return err
	})

	return resp, err
}

// GetClusterInfo API call
func (handler *ClusterRedirectionHandlerImpl) GetClusterInfo(
	ctx context.Context,
) (*types.ClusterInfo, error) {
	return handler.frontendHandler.GetClusterInfo(ctx)
}

func (handler *ClusterRedirectionHandlerImpl) beforeCall(
	scope int,
) (metrics.Scope, time.Time) {

	return handler.GetMetricsClient().Scope(scope), handler.GetTimeSource().Now()
}

func (handler *ClusterRedirectionHandlerImpl) afterCall(
	scope metrics.Scope,
	startTime time.Time,
	cluster string,
	retError *error,
) {

	log.CapturePanic(handler.GetLogger(), retError)

	scope = scope.Tagged(metrics.TargetClusterTag(cluster))
	scope.IncCounter(metrics.CadenceDcRedirectionClientRequests)
	scope.RecordTimer(metrics.CadenceDcRedirectionClientLatency, handler.GetTimeSource().Now().Sub(startTime))
	if *retError != nil {
		scope.IncCounter(metrics.CadenceDcRedirectionClientFailures)
	}
}
