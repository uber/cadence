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
	pinotVisibilityTripleManager struct {
		logger                 log.Logger
		dbVisibilityManager    VisibilityManager
		pinotVisibilityManager VisibilityManager
		esVisibilityManager    VisibilityManager
		readModeIsFromPinot    dynamicconfig.BoolPropertyFnWithDomainFilter
		readModeIsFromES       dynamicconfig.BoolPropertyFnWithDomainFilter
		readModeIsBoth         dynamicconfig.BoolPropertyFnWithDomainFilter
		writeMode              dynamicconfig.StringPropertyFn
	}
)

var _ VisibilityManager = (*pinotVisibilityTripleManager)(nil)

// NewPinotVisibilityTripleManager create a visibility manager that operate on DB or Pinot based on dynamic config.
func NewPinotVisibilityTripleManager(
	dbVisibilityManager VisibilityManager, // one of the VisibilityManager can be nil
	pinotVisibilityManager VisibilityManager, // one of the VisibilityManager can be nil
	esVisibilityManager VisibilityManager,
	readModeIsFromPinot dynamicconfig.BoolPropertyFnWithDomainFilter,
	visWritingMode dynamicconfig.StringPropertyFn,
	logger log.Logger,
) VisibilityManager {
	if dbVisibilityManager == nil && pinotVisibilityManager == nil && esVisibilityManager == nil {
		logger.Fatal("require one of dbVisibilityManager or pinotVisibilityManager or esVisibilityManager")
		return nil
	}
	return &pinotVisibilityTripleManager{
		dbVisibilityManager:    dbVisibilityManager,
		pinotVisibilityManager: pinotVisibilityManager,
		esVisibilityManager:    esVisibilityManager,
		readModeIsFromPinot:    readModeIsFromPinot,
		writeMode:              visWritingMode,
		logger:                 logger,
	}
}

func (v *pinotVisibilityTripleManager) Close() {
	if v.dbVisibilityManager != nil {
		v.dbVisibilityManager.Close()
	}
	if v.pinotVisibilityManager != nil {
		v.pinotVisibilityManager.Close()
	}
	if v.esVisibilityManager != nil {
		v.esVisibilityManager.Close()
	}
}

func (v *pinotVisibilityTripleManager) GetName() string {
	if v.pinotVisibilityManager != nil {
		return v.pinotVisibilityManager.GetName()
	} else if v.esVisibilityManager != nil {
		return v.esVisibilityManager.GetName()
	}
	return v.dbVisibilityManager.GetName()
}

func (v *pinotVisibilityTripleManager) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *RecordWorkflowExecutionStartedRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		ctx,
		func() error {
			return v.dbVisibilityManager.RecordWorkflowExecutionStarted(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.RecordWorkflowExecutionStarted(ctx, request)
		},
		func() error {
			return v.pinotVisibilityManager.RecordWorkflowExecutionStarted(ctx, request)
		},
	)
}

func (v *pinotVisibilityTripleManager) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *RecordWorkflowExecutionClosedRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		ctx,
		func() error {
			return v.dbVisibilityManager.RecordWorkflowExecutionClosed(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.RecordWorkflowExecutionClosed(ctx, request)
		},
		func() error {
			return v.pinotVisibilityManager.RecordWorkflowExecutionClosed(ctx, request)
		},
	)
}

func (v *pinotVisibilityTripleManager) RecordWorkflowExecutionUninitialized(
	ctx context.Context,
	request *RecordWorkflowExecutionUninitializedRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		ctx,
		func() error {
			return v.dbVisibilityManager.RecordWorkflowExecutionUninitialized(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.RecordWorkflowExecutionUninitialized(ctx, request)
		},
		func() error {
			return v.pinotVisibilityManager.RecordWorkflowExecutionUninitialized(ctx, request)
		},
	)
}

func (v *pinotVisibilityTripleManager) DeleteWorkflowExecution(
	ctx context.Context,
	request *VisibilityDeleteWorkflowExecutionRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		ctx,
		func() error {
			return v.dbVisibilityManager.DeleteWorkflowExecution(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.DeleteWorkflowExecution(ctx, request)
		},
		func() error {
			return v.pinotVisibilityManager.DeleteWorkflowExecution(ctx, request)
		},
	)
}

func (v *pinotVisibilityTripleManager) DeleteUninitializedWorkflowExecution(
	ctx context.Context,
	request *VisibilityDeleteWorkflowExecutionRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		ctx,
		func() error {
			return v.dbVisibilityManager.DeleteUninitializedWorkflowExecution(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.DeleteUninitializedWorkflowExecution(ctx, request)
		},
		func() error {
			return v.pinotVisibilityManager.DeleteUninitializedWorkflowExecution(ctx, request)
		},
	)
}

func (v *pinotVisibilityTripleManager) UpsertWorkflowExecution(
	ctx context.Context,
	request *UpsertWorkflowExecutionRequest,
) error {
	return v.chooseVisibilityManagerForWrite(
		ctx,
		func() error {
			return v.dbVisibilityManager.UpsertWorkflowExecution(ctx, request)
		},
		func() error {
			return v.esVisibilityManager.UpsertWorkflowExecution(ctx, request)
		},
		func() error {
			return v.pinotVisibilityManager.UpsertWorkflowExecution(ctx, request)
		},
	)
}

func (v *pinotVisibilityTripleManager) chooseVisibilityModeForAdmin() string {
	switch {
	case v.dbVisibilityManager != nil && v.pinotVisibilityManager != nil:
		return common.AdvancedVisibilityWritingModeDual
	case v.pinotVisibilityManager != nil:
		return common.AdvancedVisibilityWritingModeOn
	case v.dbVisibilityManager != nil:
		return common.AdvancedVisibilityWritingModeOff
	default:
		return "INVALID_ADMIN_MODE"
	}
}

func (v *pinotVisibilityTripleManager) chooseVisibilityManagerForWrite(ctx context.Context, dbVisFunc, esVisFunc, pinotVisFunc func() error) error {
	var writeMode string
	if v.writeMode != nil {
		writeMode = v.writeMode()
	} else {
		key := VisibilityAdminDeletionKey("visibilityAdminDelete")
		if value := ctx.Value(key); value != nil && value.(bool) {
			writeMode = v.chooseVisibilityModeForAdmin()
		}
	}

	switch writeMode {
	case common.AdvancedVisibilityWritingModeOff:
		if v.dbVisibilityManager != nil {
			return dbVisFunc()
		}
		v.logger.Warn("basic visibility is not available to write, fall back to advanced visibility")
		return pinotVisFunc()
	case common.AdvancedVisibilityWritingModeOn:
		if v.pinotVisibilityManager != nil {
			return pinotVisFunc()
		}
		v.logger.Warn("advanced visibility is not available to write, fall back to basic visibility")
		return dbVisFunc()
	case common.AdvancedVisibilityWritingModeDual:
		if v.pinotVisibilityManager != nil {
			if err := pinotVisFunc(); err != nil {
				return err
			}
			if v.dbVisibilityManager != nil {
				return dbVisFunc()
			}
			v.logger.Warn("basic visibility is not available to write")
			return nil
		}
		v.logger.Warn("advanced visibility is not available to write")
		return dbVisFunc()
	case common.AdvacnedVisibilityWritingModeTriple:
		if v.pinotVisibilityManager != nil && v.esVisibilityManager != nil {
			if err := pinotVisFunc(); err != nil {
				return err
			}
			if err := esVisFunc(); err != nil {
				return err
			}
			if v.dbVisibilityManager != nil {
				return dbVisFunc()
			}
			v.logger.Warn("basic visibility is not available to write")
			return nil
		}
		v.logger.Warn("advanced visibility is not available to write")
		return dbVisFunc()
	default:
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Unknown visibility writing mode: %s", writeMode),
		}
	}
}

func (v *pinotVisibilityTripleManager) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESListOpenResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.ListOpenWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESListClosedResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.ListClosedWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESListOpenByTypeResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.ListOpenWorkflowExecutionsByType(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESListClosedByTypeResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.ListClosedWorkflowExecutionsByType(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESListOpenByWorkflowIDResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESListClosedByWorkflowIDResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *ListClosedWorkflowExecutionsByStatusRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESListClosedByStatusResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.ListClosedWorkflowExecutionsByStatus(ctx, request)
}

func (v *pinotVisibilityTripleManager) GetClosedWorkflowExecution(
	ctx context.Context,
	request *GetClosedWorkflowExecutionRequest,
) (*GetClosedWorkflowExecutionResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESGetClosedByStatusResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.GetClosedWorkflowExecution(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESListByQueryResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.ListWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) ScanWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESScanResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.ScanWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) CountWorkflowExecutions(
	ctx context.Context,
	request *CountWorkflowExecutionsRequest,
) (*CountWorkflowExecutionsResponse, error) {
	manager, manager2 := v.chooseVisibilityManagerForRead(request.Domain)
	if manager2 != nil {
		return comparePinotESCountResponse(ctx, manager, manager2, request, v.logger)
	}
	return manager.CountWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) chooseVisibilityManagerForRead(domain string) (VisibilityManager, VisibilityManager) {
	var visibilityMgr VisibilityManager
	var visibilityMgr2 VisibilityManager
	if v.readModeIsBoth(domain) {
		if v.esVisibilityManager != nil {
			visibilityMgr = v.esVisibilityManager
		} else {
			visibilityMgr = v.dbVisibilityManager
			v.logger.Warn("domain is configured to read from advanced visibility(Pinot based) but it's not available, fall back to basic visibility",
				tag.WorkflowDomainName(domain))
		}

		if v.pinotVisibilityManager != nil {
			visibilityMgr2 = v.pinotVisibilityManager
		} else {
			visibilityMgr2 = v.dbVisibilityManager
			v.logger.Warn("domain is configured to read from advanced visibility(ElasticSearch based) but it's not available, fall back to basic visibility",
				tag.WorkflowDomainName(domain))
		}
		// mgr1 = esManager, mgr2 = pinotManager
		return visibilityMgr, visibilityMgr2
	}

	if v.readModeIsFromPinot(domain) {
		if v.pinotVisibilityManager != nil {
			visibilityMgr = v.pinotVisibilityManager
		} else {
			visibilityMgr = v.dbVisibilityManager
			v.logger.Warn("domain is configured to read from advanced visibility(Pinot based) but it's not available, fall back to basic visibility",
				tag.WorkflowDomainName(domain))
		}
	} else if v.readModeIsFromES(domain) {
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
			visibilityMgr = v.pinotVisibilityManager
			v.logger.Warn("domain is configured to read from basic visibility but it's not available, fall back to advanced visibility",
				tag.WorkflowDomainName(domain))
		}
	}
	return visibilityMgr, nil
}
