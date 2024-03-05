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

package queueconfigapi

import (
	"context"

	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
)

type handlerImpl struct {
	logger        log.Logger
	domainHandler domain.Handler
}

func New(logger log.Logger, dh domain.Handler) Handler {
	return &handlerImpl{
		logger:        logger,
		domainHandler: dh,
	}
}

func (h *handlerImpl) GetConfiguraton(ctx context.Context, req *types.GetDomainAsyncWorkflowConfiguratonRequest) (*types.GetDomainAsyncWorkflowConfiguratonResponse, error) {
	resp, err := h.domainHandler.DescribeDomain(ctx, &types.DescribeDomainRequest{
		Name: &req.Domain,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Configuration == nil || resp.Configuration.AsyncWorkflowConfig == nil {
		return &types.GetDomainAsyncWorkflowConfiguratonResponse{}, nil
	}

	return &types.GetDomainAsyncWorkflowConfiguratonResponse{
		Configuration: resp.Configuration.AsyncWorkflowConfig,
	}, nil
}

func (h *handlerImpl) UpdateConfiguration(ctx context.Context, req *types.UpdateDomainAsyncWorkflowConfiguratonRequest) (*types.UpdateDomainAsyncWorkflowConfiguratonResponse, error) {
	if req == nil {
		return nil, &types.BadRequestError{Message: "Request is nil."}
	}

	err := h.domainHandler.UpdateAsyncWorkflowConfiguraton(ctx, *req)
	if err != nil {
		return nil, err
	}

	return &types.UpdateDomainAsyncWorkflowConfiguratonResponse{}, nil
}
