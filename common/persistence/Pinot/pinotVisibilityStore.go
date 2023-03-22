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

package pinotVisibility

import (
	"context"

	"github.com/startreedata/pinot-client-go/pinot"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service"
)

type (
	pinotVisibilityStore struct {
		pinotClinent *pinot.Connection
		logger       log.Logger
		producer     messaging.Producer
		config       *service.Config
	}
)

// NewPinotVisibilityStore create a visibility store connecting to Pinot
func NewPinotVisibilityStore(
	pinotClinent *pinot.Connection,
	config *service.Config,
	producer messaging.Producer,
	logger log.Logger,
) p.VisibilityStore {
	return &pinotVisibilityStore{
		pinotClinent: pinotClinent,
		logger:       logger.WithTags(tag.ComponentPinotVisibilityManager),
		config:       config,
		producer:     producer,
	}
}

func (v *pinotVisibilityStore) Close() {}

func (v *pinotVisibilityStore) GetName() string {
	return "pinotVisibility"
}
func (v *pinotVisibilityStore) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionStartedRequest,
) error {
	return p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionClosedRequest,
) error {
	return p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) RecordWorkflowExecutionUninitialized(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionUninitializedRequest,
) error {
	return p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) UpsertWorkflowExecution(
	ctx context.Context,
	request *p.InternalUpsertWorkflowExecutionRequest,
) error {
	return p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *p.InternalListClosedWorkflowExecutionsByStatusRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) GetClosedWorkflowExecution(
	ctx context.Context,
	request *p.InternalGetClosedWorkflowExecutionRequest,
) (*p.InternalGetClosedWorkflowExecutionResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) DeleteWorkflowExecution(
	ctx context.Context,
	request *p.VisibilityDeleteWorkflowExecutionRequest,
) error {
	return p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) DeleteUninitializedWorkflowExecution(
	ctx context.Context,
	request *p.VisibilityDeleteWorkflowExecutionRequest,
) error {
	// temporary: not implemented, only implemented for ES
	return p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) ListWorkflowExecutions(
	_ context.Context,
	_ *p.ListWorkflowExecutionsByQueryRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) ScanWorkflowExecutions(
	_ context.Context,
	_ *p.ListWorkflowExecutionsByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}

func (v *pinotVisibilityStore) CountWorkflowExecutions(
	_ context.Context,
	_ *p.CountWorkflowExecutionsRequest,
) (*p.CountWorkflowExecutionsResponse, error) {
	return nil, p.ErrVisibilityOperationNotSupported
}
