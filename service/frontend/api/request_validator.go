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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination request_validator_mock.go -self_package github.com/uber/cadence/service/frontend/api requestValidator

package api

import (
	"context"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/frontend/config"
	"github.com/uber/cadence/service/frontend/validate"
)

type (
	RequestValidator interface {
		ValidateRefreshWorkflowTasksRequest(context.Context, *types.RefreshWorkflowTasksRequest) error
		ValidateDescribeTaskListRequest(context.Context, *types.DescribeTaskListRequest) error
		ValidateListTaskListPartitionsRequest(context.Context, *types.ListTaskListPartitionsRequest) error
		ValidateGetTaskListsByDomainRequest(context.Context, *types.GetTaskListsByDomainRequest) error
		ValidateResetStickyTaskListRequest(context.Context, *types.ResetStickyTaskListRequest) error
	}

	requestValidatorImpl struct {
		logger        log.Logger
		metricsClient metrics.Client
		config        *config.Config
	}
)

func NewRequestValidator(logger log.Logger, metricsClient metrics.Client, config *config.Config) RequestValidator {
	return &requestValidatorImpl{
		logger:        logger,
		metricsClient: metricsClient,
		config:        config,
	}
}

func (v *requestValidatorImpl) validateTaskList(t *types.TaskList, scope metrics.Scope, domain string) error {
	if t == nil || t.GetName() == "" {
		return validate.ErrTaskListNotSet
	}

	if !common.IsValidIDLength(
		t.GetName(),
		scope,
		v.config.MaxIDLengthWarnLimit(),
		v.config.TaskListNameMaxLength(domain),
		metrics.CadenceErrTaskListNameExceededWarnLimit,
		domain,
		v.logger,
		tag.IDTypeTaskListName) {
		return validate.ErrTaskListTooLong
	}
	return nil
}

func (v *requestValidatorImpl) ValidateRefreshWorkflowTasksRequest(ctx context.Context, req *types.RefreshWorkflowTasksRequest) error {
	if req == nil {
		return validate.ErrRequestNotSet
	}
	return validate.CheckExecution(req.Execution)
}

func (v *requestValidatorImpl) ValidateDescribeTaskListRequest(ctx context.Context, request *types.DescribeTaskListRequest) error {
	if request == nil {
		return validate.ErrRequestNotSet
	}
	if request.GetDomain() == "" {
		return validate.ErrDomainNotSet
	}
	if request.TaskListType == nil {
		return validate.ErrTaskListTypeNotSet
	}
	scope := getMetricsScopeWithDomain(metrics.FrontendDescribeTaskListScope, request, v.metricsClient).Tagged(metrics.GetContextTags(ctx)...)
	return v.validateTaskList(request.TaskList, scope, request.GetDomain())
}

func (v *requestValidatorImpl) ValidateListTaskListPartitionsRequest(ctx context.Context, request *types.ListTaskListPartitionsRequest) error {
	if request == nil {
		return validate.ErrRequestNotSet
	}
	if request.GetDomain() == "" {
		return validate.ErrDomainNotSet
	}
	scope := getMetricsScopeWithDomain(metrics.FrontendListTaskListPartitionsScope, request, v.metricsClient).Tagged(metrics.GetContextTags(ctx)...)
	return v.validateTaskList(request.TaskList, scope, request.GetDomain())
}

func (v *requestValidatorImpl) ValidateGetTaskListsByDomainRequest(ctx context.Context, request *types.GetTaskListsByDomainRequest) error {
	if request == nil {
		return validate.ErrRequestNotSet
	}
	if request.GetDomain() == "" {
		return validate.ErrDomainNotSet
	}
	return nil
}

func (v *requestValidatorImpl) ValidateResetStickyTaskListRequest(ctx context.Context, resetRequest *types.ResetStickyTaskListRequest) error {
	if resetRequest == nil {
		return validate.ErrRequestNotSet
	}
	domainName := resetRequest.GetDomain()
	if domainName == "" {
		return validate.ErrDomainNotSet
	}
	wfExecution := resetRequest.GetExecution()
	return validate.CheckExecution(wfExecution)
}
