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

package persistence

import (
	"context"
	"fmt"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"github.com/uber/cadence/common/types"
)

type (
	visibilityManagerWrapper struct {
		visibilityManager          VisibilityManager
		esVisibilityManager        VisibilityManager
		enableReadVisibilityFromES dynamicconfig.BoolPropertyFnWithDomainFilter
		advancedVisWritingMode     dynamicconfig.StringPropertyFn
	}
)

var _ VisibilityManager = (*visibilityManagerWrapper)(nil)

// NewVisibilityManagerWrapper create a visibility manager that operate on DB or ElasticSearch based on dynamic config.
func NewVisibilityManagerWrapper(
	visibilityManager VisibilityManager,
	esVisibilityManager VisibilityManager,
	enableReadVisibilityFromES dynamicconfig.BoolPropertyFnWithDomainFilter,
	advancedVisWritingMode dynamicconfig.StringPropertyFn,
) VisibilityManager {
	return &visibilityManagerWrapper{
		visibilityManager:          visibilityManager,
		esVisibilityManager:        esVisibilityManager,
		enableReadVisibilityFromES: enableReadVisibilityFromES,
		advancedVisWritingMode:     advancedVisWritingMode,
	}
}

func (v *visibilityManagerWrapper) Close() {
	if v.visibilityManager != nil {
		v.visibilityManager.Close()
	}
	if v.esVisibilityManager != nil {
		v.esVisibilityManager.Close()
	}
}

func (v *visibilityManagerWrapper) GetName() string {
	return "visibilityManagerWrapper"
}

func (v *visibilityManagerWrapper) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *RecordWorkflowExecutionStartedRequest,
) error {
	switch v.advancedVisWritingMode() {
	case common.AdvancedVisibilityWritingModeOff:
		return v.visibilityManager.RecordWorkflowExecutionStarted(ctx, request)
	case common.AdvancedVisibilityWritingModeOn:
		return v.esVisibilityManager.RecordWorkflowExecutionStarted(ctx, request)
	case common.AdvancedVisibilityWritingModeDual:
		if err := v.esVisibilityManager.RecordWorkflowExecutionStarted(ctx, request); err != nil {
			return err
		}
		return v.visibilityManager.RecordWorkflowExecutionStarted(ctx, request)
	default:
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Unknown advanced visibility writing mode: %s", v.advancedVisWritingMode()),
		}
	}
}

func (v *visibilityManagerWrapper) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *RecordWorkflowExecutionClosedRequest,
) error {
	switch v.advancedVisWritingMode() {
	case common.AdvancedVisibilityWritingModeOff:
		return v.visibilityManager.RecordWorkflowExecutionClosed(ctx, request)
	case common.AdvancedVisibilityWritingModeOn:
		return v.esVisibilityManager.RecordWorkflowExecutionClosed(ctx, request)
	case common.AdvancedVisibilityWritingModeDual:
		if err := v.esVisibilityManager.RecordWorkflowExecutionClosed(ctx, request); err != nil {
			return err
		}
		return v.visibilityManager.RecordWorkflowExecutionClosed(ctx, request)
	default:
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Unknown advanced visibility writing mode: %s", v.advancedVisWritingMode()),
		}
	}
}

func (v *visibilityManagerWrapper) UpsertWorkflowExecution(
	ctx context.Context,
	request *UpsertWorkflowExecutionRequest,
) error {
	if v.esVisibilityManager == nil { // return operation not support
		return v.visibilityManager.UpsertWorkflowExecution(ctx, request)
	}

	return v.esVisibilityManager.UpsertWorkflowExecution(ctx, request)
}

func (v *visibilityManagerWrapper) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.ListOpenWorkflowExecutions(ctx, request)
}

func (v *visibilityManagerWrapper) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.ListClosedWorkflowExecutions(ctx, request)
}

func (v *visibilityManagerWrapper) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.ListOpenWorkflowExecutionsByType(ctx, request)
}

func (v *visibilityManagerWrapper) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.ListClosedWorkflowExecutionsByType(ctx, request)
}

func (v *visibilityManagerWrapper) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
}

func (v *visibilityManagerWrapper) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
}

func (v *visibilityManagerWrapper) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *ListClosedWorkflowExecutionsByStatusRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.ListClosedWorkflowExecutionsByStatus(ctx, request)
}

func (v *visibilityManagerWrapper) GetClosedWorkflowExecution(
	ctx context.Context,
	request *GetClosedWorkflowExecutionRequest,
) (*GetClosedWorkflowExecutionResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.GetClosedWorkflowExecution(ctx, request)
}

func (v *visibilityManagerWrapper) DeleteWorkflowExecution(
	ctx context.Context,
	request *VisibilityDeleteWorkflowExecutionRequest,
) error {
	if v.esVisibilityManager != nil {
		if err := v.esVisibilityManager.DeleteWorkflowExecution(ctx, request); err != nil {
			return err
		}
	}
	return v.visibilityManager.DeleteWorkflowExecution(ctx, request)
}

func (v *visibilityManagerWrapper) ListWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.ListWorkflowExecutions(ctx, request)
}

func (v *visibilityManagerWrapper) ScanWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.ScanWorkflowExecutions(ctx, request)
}

func (v *visibilityManagerWrapper) CountWorkflowExecutions(
	ctx context.Context,
	request *CountWorkflowExecutionsRequest,
) (*CountWorkflowExecutionsResponse, error) {
	manager := v.chooseVisibilityManagerForDomain(request.Domain)
	return manager.CountWorkflowExecutions(ctx, request)
}

func (v *visibilityManagerWrapper) chooseVisibilityManagerForDomain(domain string) VisibilityManager {
	var visibilityMgr VisibilityManager
	if v.enableReadVisibilityFromES(domain) && v.esVisibilityManager != nil {
		visibilityMgr = v.esVisibilityManager
	} else {
		visibilityMgr = v.visibilityManager
	}
	return visibilityMgr
}
