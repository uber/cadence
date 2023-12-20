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

package ratelimited

import (
	"context"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
)

type visibilityClient struct {
	rateLimiter quotas.Limiter
	persistence persistence.VisibilityManager
	logger      log.Logger
}

// NewVisibilityClient creates a client to manage visibility
func NewVisibilityClient(
	persistence persistence.VisibilityManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence.VisibilityManager {
	return &visibilityClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (p *visibilityClient) GetName() string {
	return p.persistence.GetName()
}

func (p *visibilityClient) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *persistence.RecordWorkflowExecutionStartedRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RecordWorkflowExecutionStarted(ctx, request)
	return err
}

func (p *visibilityClient) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *persistence.RecordWorkflowExecutionClosedRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RecordWorkflowExecutionClosed(ctx, request)
	return err
}

func (p *visibilityClient) RecordWorkflowExecutionUninitialized(
	ctx context.Context,
	request *persistence.RecordWorkflowExecutionUninitializedRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RecordWorkflowExecutionUninitialized(ctx, request)
	return err
}

func (p *visibilityClient) UpsertWorkflowExecution(
	ctx context.Context,
	request *persistence.UpsertWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.UpsertWorkflowExecution(ctx, request)
	return err
}

func (p *visibilityClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListOpenWorkflowExecutions(ctx, request)
	return response, err
}

func (p *visibilityClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutions(ctx, request)
	return response, err
}

func (p *visibilityClient) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByTypeRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListOpenWorkflowExecutionsByType(ctx, request)
	return response, err
}

func (p *visibilityClient) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByTypeRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutionsByType(ctx, request)
	return response, err
}

func (p *visibilityClient) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByWorkflowIDRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
	return response, err
}

func (p *visibilityClient) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByWorkflowIDRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
	return response, err
}

func (p *visibilityClient) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *persistence.ListClosedWorkflowExecutionsByStatusRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutionsByStatus(ctx, request)
	return response, err
}

func (p *visibilityClient) GetClosedWorkflowExecution(
	ctx context.Context,
	request *persistence.GetClosedWorkflowExecutionRequest,
) (*persistence.GetClosedWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetClosedWorkflowExecution(ctx, request)
	return response, err
}

func (p *visibilityClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *persistence.VisibilityDeleteWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}
	return p.persistence.DeleteWorkflowExecution(ctx, request)
}

func (p *visibilityClient) DeleteUninitializedWorkflowExecution(
	ctx context.Context,
	request *persistence.VisibilityDeleteWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}
	return p.persistence.DeleteUninitializedWorkflowExecution(ctx, request)
}

func (p *visibilityClient) ListWorkflowExecutions(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByQueryRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.ListWorkflowExecutions(ctx, request)
}

func (p *visibilityClient) ScanWorkflowExecutions(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByQueryRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.ScanWorkflowExecutions(ctx, request)
}

func (p *visibilityClient) CountWorkflowExecutions(
	ctx context.Context,
	request *persistence.CountWorkflowExecutionsRequest,
) (*persistence.CountWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.CountWorkflowExecutions(ctx, request)
}

func (p *visibilityClient) Close() {
	p.persistence.Close()
}
