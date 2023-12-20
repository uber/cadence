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

package error_injectors

import (
	"context"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type visibilityClient struct {
	persistence persistence.VisibilityManager
	errorRate   float64
	logger      log.Logger
}

// NewVisibilityClient creates an error injection client to manage visibility
func NewVisibilityClient(
	persistence persistence.VisibilityManager,
	errorRate float64,
	logger log.Logger,
) persistence.VisibilityManager {
	return &visibilityClient{
		persistence: persistence,
		errorRate:   errorRate,
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
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RecordWorkflowExecutionStarted(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRecordWorkflowExecutionStarted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityClient) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *persistence.RecordWorkflowExecutionClosedRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RecordWorkflowExecutionClosed(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRecordWorkflowExecutionClosed,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityClient) RecordWorkflowExecutionUninitialized(
	ctx context.Context,
	request *persistence.RecordWorkflowExecutionUninitializedRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RecordWorkflowExecutionUninitialized(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRecordWorkflowExecutionStarted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityClient) UpsertWorkflowExecution(
	ctx context.Context,
	request *persistence.UpsertWorkflowExecutionRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpsertWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpsertWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListOpenWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListOpenWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListClosedWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListClosedWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByTypeRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListOpenWorkflowExecutionsByType(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListOpenWorkflowExecutionsByType,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByTypeRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListClosedWorkflowExecutionsByType(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListClosedWorkflowExecutionsByType,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByWorkflowIDRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListOpenWorkflowExecutionsByWorkflowID,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByWorkflowIDRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListClosedWorkflowExecutionsByWorkflowID,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *persistence.ListClosedWorkflowExecutionsByStatusRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListClosedWorkflowExecutionsByStatus(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListClosedWorkflowExecutionsByStatus,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) GetClosedWorkflowExecution(
	ctx context.Context,
	request *persistence.GetClosedWorkflowExecutionRequest,
) (*persistence.GetClosedWorkflowExecutionResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetClosedWorkflowExecutionResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetClosedWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetClosedWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *persistence.VisibilityDeleteWorkflowExecutionRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationVisibilityDeleteWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityClient) DeleteUninitializedWorkflowExecution(
	ctx context.Context,
	request *persistence.VisibilityDeleteWorkflowExecutionRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteUninitializedWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationVisibilityDeleteWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityClient) ListWorkflowExecutions(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByQueryRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) ScanWorkflowExecutions(
	ctx context.Context,
	request *persistence.ListWorkflowExecutionsByQueryRequest,
) (*persistence.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ScanWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationScanWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) CountWorkflowExecutions(
	ctx context.Context,
	request *persistence.CountWorkflowExecutionsRequest,
) (*persistence.CountWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.CountWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.CountWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCountWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityClient) Close() {
	p.persistence.Close()
}
