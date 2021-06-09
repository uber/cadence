// Copyright (c) 2021 Uber Technologies, Inc.
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

package persistence

import (
	"context"
	"fmt"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
)

type (
	visibilityDualManager struct {
		logger              log.Logger
		dbVisibilityManager VisibilityManager
		esVisibilityManager VisibilityManager
		readModeIsFromES    dynamicconfig.BoolPropertyFnWithDomainFilter
		writeMode           dynamicconfig.StringPropertyFn
	}
)

var _ VisibilityManager = (*visibilityDualManager)(nil)

// NewVisibilityDualManager create a visibility manager that operate on DB or ElasticSearch based on dynamic config.
func NewVisibilityDualManager(
	dbVisibilityManager VisibilityManager, // one of the VisibilityManager can be nil
	esVisibilityManager VisibilityManager, // one of the VisibilityManager can be nil
	readModeIsFromES dynamicconfig.BoolPropertyFnWithDomainFilter,
	visWritingMode dynamicconfig.StringPropertyFn,
	logger log.Logger,
) VisibilityManager {
	if dbVisibilityManager == nil && esVisibilityManager == nil {
		logger.Fatal("require one of dbVisibilityManager or esVisibilityManager")
		return nil
	}
	return &visibilityDualManager{
		dbVisibilityManager: dbVisibilityManager,
		esVisibilityManager: esVisibilityManager,
		readModeIsFromES:    readModeIsFromES,
		writeMode:           visWritingMode,
		logger:              logger,
	}
}

func (v *visibilityDualManager) Close() {
	if v.dbVisibilityManager != nil {
		v.dbVisibilityManager.Close()
	}
	if v.esVisibilityManager != nil {
		v.esVisibilityManager.Close()
	}
}

func (v *visibilityDualManager) GetName() string {
	if v.esVisibilityManager != nil {
		return v.esVisibilityManager.GetName()
	}
	return v.dbVisibilityManager.GetName()
}

func (v *visibilityDualManager) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *RecordWorkflowExecutionStartedRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		func() error {
			return v.dbVisibilityManager.RecordWorkflowExecutionStarted(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.RecordWorkflowExecutionStarted(ctx, request)
		},
	)
}

func (v *visibilityDualManager) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *RecordWorkflowExecutionClosedRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		func() error {
			return v.dbVisibilityManager.RecordWorkflowExecutionClosed(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.RecordWorkflowExecutionClosed(ctx, request)
		},
	)
}

func (v *visibilityDualManager) DeleteWorkflowExecution(
	ctx context.Context,
	request *VisibilityDeleteWorkflowExecutionRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		func() error {
			return v.dbVisibilityManager.DeleteWorkflowExecution(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.DeleteWorkflowExecution(ctx, request)
		},
	)
}

func (v *visibilityDualManager) UpsertWorkflowExecution(
	ctx context.Context,
	request *UpsertWorkflowExecutionRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		func() error {
			return v.dbVisibilityManager.UpsertWorkflowExecution(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.UpsertWorkflowExecution(ctx, request)
		},
	)
}

func (v *visibilityDualManager) chooseVisibilityManagerForWrite(dbVisFunc, esVisFunc func() error) error {
	switch v.writeMode() {
	case common.AdvancedVisibilityWritingModeOff:
		if v.dbVisibilityManager == nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("visibility writing mode: %s is misconfigured", v.writeMode()),
			}
		}
		return dbVisFunc()
	case common.AdvancedVisibilityWritingModeOn:
		if v.esVisibilityManager == nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("visibility writing mode: %s is misconfigured", v.writeMode()),
			}
		}

		return esVisFunc()
	case common.AdvancedVisibilityWritingModeDual:
		if v.dbVisibilityManager == nil || v.esVisibilityManager == nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("visibility writing mode: %s is misconfigured", v.writeMode()),
			}
		}

		if err := esVisFunc(); err != nil {
			return err
		}
		return dbVisFunc()
	default:
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Unknown visibility writing mode: %s", v.writeMode()),
		}
	}
}

func (v *visibilityDualManager) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.ListOpenWorkflowExecutions(ctx, request)
}

func (v *visibilityDualManager) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.ListClosedWorkflowExecutions(ctx, request)
}

func (v *visibilityDualManager) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.ListOpenWorkflowExecutionsByType(ctx, request)
}

func (v *visibilityDualManager) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.ListClosedWorkflowExecutionsByType(ctx, request)
}

func (v *visibilityDualManager) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
}

func (v *visibilityDualManager) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
}

func (v *visibilityDualManager) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *ListClosedWorkflowExecutionsByStatusRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.ListClosedWorkflowExecutionsByStatus(ctx, request)
}

func (v *visibilityDualManager) GetClosedWorkflowExecution(
	ctx context.Context,
	request *GetClosedWorkflowExecutionRequest,
) (*GetClosedWorkflowExecutionResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.GetClosedWorkflowExecution(ctx, request)
}

func (v *visibilityDualManager) ListWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.ListWorkflowExecutions(ctx, request)
}

func (v *visibilityDualManager) ScanWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.ScanWorkflowExecutions(ctx, request)
}

func (v *visibilityDualManager) CountWorkflowExecutions(
	ctx context.Context,
	request *CountWorkflowExecutionsRequest,
) (*CountWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForRead(request.Domain)
	return manager.CountWorkflowExecutions(ctx, request)
}

func (v *visibilityDualManager) chooseVisibilityManagerForRead(domain string) VisibilityManager {
	var visibilityMgr VisibilityManager
	if v.readModeIsFromES(domain) {
		if v.esVisibilityManager != nil {
			visibilityMgr = v.esVisibilityManager
		} else {
			visibilityMgr = v.dbVisibilityManager
			v.logger.Warn("domain is configured to read from advanced visibility(ElasticSearch based) but it's not available, fall back to basic visibility",
				tag.WorkflowDomainName(domain))
		}
	} else {
		if v.dbVisibilityManager != nil {
			visibilityMgr = v.dbVisibilityManager
		} else {
			visibilityMgr = v.esVisibilityManager
			v.logger.Warn("domain is configured to read from basic visibility but it's not available, fall back to advanced visibility",
				tag.WorkflowDomainName(domain))
		}
	}
	return visibilityMgr
}
