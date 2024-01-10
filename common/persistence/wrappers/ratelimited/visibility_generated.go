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

// Code generated by gowrap. DO NOT EDIT.
// template: ../templates/ratelimited.tmpl
// gowrap: http://github.com/hexdigest/gowrap

import (
	"context"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
)

// ratelimitedVisibilityManager implements persistence.VisibilityManager interface instrumented with rate limiter.
type ratelimitedVisibilityManager struct {
	wrapped     persistence.VisibilityManager
	rateLimiter quotas.Limiter
}

// NewVisibilityManager creates a new instance of VisibilityManager with ratelimiter.
func NewVisibilityManager(
	wrapped persistence.VisibilityManager,
	rateLimiter quotas.Limiter,
) persistence.VisibilityManager {
	return &ratelimitedVisibilityManager{
		wrapped:     wrapped,
		rateLimiter: rateLimiter,
	}
}

func (c *ratelimitedVisibilityManager) Close() {
	c.wrapped.Close()
	return
}

func (c *ratelimitedVisibilityManager) CountWorkflowExecutions(ctx context.Context, request *persistence.CountWorkflowExecutionsRequest) (cp1 *persistence.CountWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.CountWorkflowExecutions(ctx, request)
}

func (c *ratelimitedVisibilityManager) DeleteUninitializedWorkflowExecution(ctx context.Context, request *persistence.VisibilityDeleteWorkflowExecutionRequest) (err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.DeleteUninitializedWorkflowExecution(ctx, request)
}

func (c *ratelimitedVisibilityManager) DeleteWorkflowExecution(ctx context.Context, request *persistence.VisibilityDeleteWorkflowExecutionRequest) (err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.DeleteWorkflowExecution(ctx, request)
}

func (c *ratelimitedVisibilityManager) GetClosedWorkflowExecution(ctx context.Context, request *persistence.GetClosedWorkflowExecutionRequest) (gp1 *persistence.GetClosedWorkflowExecutionResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.GetClosedWorkflowExecution(ctx, request)
}

func (c *ratelimitedVisibilityManager) GetName() (s1 string) {
	return c.wrapped.GetName()
}

func (c *ratelimitedVisibilityManager) ListClosedWorkflowExecutions(ctx context.Context, request *persistence.ListWorkflowExecutionsRequest) (lp1 *persistence.ListWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.ListClosedWorkflowExecutions(ctx, request)
}

func (c *ratelimitedVisibilityManager) ListClosedWorkflowExecutionsByStatus(ctx context.Context, request *persistence.ListClosedWorkflowExecutionsByStatusRequest) (lp1 *persistence.ListWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.ListClosedWorkflowExecutionsByStatus(ctx, request)
}

func (c *ratelimitedVisibilityManager) ListClosedWorkflowExecutionsByType(ctx context.Context, request *persistence.ListWorkflowExecutionsByTypeRequest) (lp1 *persistence.ListWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.ListClosedWorkflowExecutionsByType(ctx, request)
}

func (c *ratelimitedVisibilityManager) ListClosedWorkflowExecutionsByWorkflowID(ctx context.Context, request *persistence.ListWorkflowExecutionsByWorkflowIDRequest) (lp1 *persistence.ListWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
}

func (c *ratelimitedVisibilityManager) ListOpenWorkflowExecutions(ctx context.Context, request *persistence.ListWorkflowExecutionsRequest) (lp1 *persistence.ListWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.ListOpenWorkflowExecutions(ctx, request)
}

func (c *ratelimitedVisibilityManager) ListOpenWorkflowExecutionsByType(ctx context.Context, request *persistence.ListWorkflowExecutionsByTypeRequest) (lp1 *persistence.ListWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.ListOpenWorkflowExecutionsByType(ctx, request)
}

func (c *ratelimitedVisibilityManager) ListOpenWorkflowExecutionsByWorkflowID(ctx context.Context, request *persistence.ListWorkflowExecutionsByWorkflowIDRequest) (lp1 *persistence.ListWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
}

func (c *ratelimitedVisibilityManager) ListWorkflowExecutions(ctx context.Context, request *persistence.ListWorkflowExecutionsByQueryRequest) (lp1 *persistence.ListWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.ListWorkflowExecutions(ctx, request)
}

func (c *ratelimitedVisibilityManager) RecordWorkflowExecutionClosed(ctx context.Context, request *persistence.RecordWorkflowExecutionClosedRequest) (err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.RecordWorkflowExecutionClosed(ctx, request)
}

func (c *ratelimitedVisibilityManager) RecordWorkflowExecutionStarted(ctx context.Context, request *persistence.RecordWorkflowExecutionStartedRequest) (err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.RecordWorkflowExecutionStarted(ctx, request)
}

func (c *ratelimitedVisibilityManager) RecordWorkflowExecutionUninitialized(ctx context.Context, request *persistence.RecordWorkflowExecutionUninitializedRequest) (err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.RecordWorkflowExecutionUninitialized(ctx, request)
}

func (c *ratelimitedVisibilityManager) ScanWorkflowExecutions(ctx context.Context, request *persistence.ListWorkflowExecutionsByQueryRequest) (lp1 *persistence.ListWorkflowExecutionsResponse, err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.ScanWorkflowExecutions(ctx, request)
}

func (c *ratelimitedVisibilityManager) UpsertWorkflowExecution(ctx context.Context, request *persistence.UpsertWorkflowExecutionRequest) (err error) {
	if ok := c.rateLimiter.Allow(); !ok {
		err = ErrPersistenceLimitExceeded
		return
	}
	return c.wrapped.UpsertWorkflowExecution(ctx, request)
}
