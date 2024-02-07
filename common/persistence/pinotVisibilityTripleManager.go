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
	"math/rand"
	"strings"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
)

type (
	pinotVisibilityTripleManager struct {
		logger                    log.Logger
		dbVisibilityManager       VisibilityManager
		pinotVisibilityManager    VisibilityManager
		esVisibilityManager       VisibilityManager
		readModeIsFromPinot       dynamicconfig.BoolPropertyFnWithDomainFilter
		readModeIsFromES          dynamicconfig.BoolPropertyFnWithDomainFilter
		writeMode                 dynamicconfig.StringPropertyFn
		logCustomerQueryParameter dynamicconfig.BoolPropertyFnWithDomainFilter
	}
)

const (
	VisibilityOverridePrimary   = "Primary"
	VisibilityOverrideSecondary = "Secondary"
	ContextKey                  = ResponseComparatorContextKey("visibility-override")
)

// ResponseComparatorContextKey is for Pinot/ES response comparator. This struct will be passed into ctx as a key.
type ResponseComparatorContextKey string

type OperationType string

var Operation = struct {
	LIST  OperationType
	COUNT OperationType
}{
	LIST:  "list",
	COUNT: "count",
}

var _ VisibilityManager = (*pinotVisibilityTripleManager)(nil)

// NewPinotVisibilityTripleManager create a visibility manager that operate on DB or Pinot based on dynamic config.
func NewPinotVisibilityTripleManager(
	dbVisibilityManager VisibilityManager, // one of the VisibilityManager can be nil
	pinotVisibilityManager VisibilityManager,
	esVisibilityManager VisibilityManager,
	readModeIsFromPinot dynamicconfig.BoolPropertyFnWithDomainFilter,
	readModeIsFromES dynamicconfig.BoolPropertyFnWithDomainFilter,
	visWritingMode dynamicconfig.StringPropertyFn,
	logCustomerQueryParameter dynamicconfig.BoolPropertyFnWithDomainFilter,
	logger log.Logger,
) VisibilityManager {
	if dbVisibilityManager == nil && pinotVisibilityManager == nil && esVisibilityManager == nil {
		logger.Fatal("require one of dbVisibilityManager or pinotVisibilityManager or esVisibilityManager")
		return nil
	}
	return &pinotVisibilityTripleManager{
		dbVisibilityManager:       dbVisibilityManager,
		pinotVisibilityManager:    pinotVisibilityManager,
		esVisibilityManager:       esVisibilityManager,
		readModeIsFromPinot:       readModeIsFromPinot,
		readModeIsFromES:          readModeIsFromES,
		writeMode:                 visWritingMode,
		logger:                    logger,
		logCustomerQueryParameter: logCustomerQueryParameter,
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
	case v.dbVisibilityManager != nil && v.esVisibilityManager != nil && v.pinotVisibilityManager != nil:
		return common.AdvancedVisibilityWritingModeTriple
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
	//only perform as triple manager during migration by setting write mode to triple,
	//other time perform as a dual visibility manager of pinot and db
	case common.AdvancedVisibilityWritingModeOff:
		if v.dbVisibilityManager != nil {
			return dbVisFunc()
		}
		v.logger.Warn("basic visibility is not available to write, fall back to advanced visibility")
		return pinotVisFunc()
	case common.AdvancedVisibilityWritingModeOn:
		// this is the way to make it work for migration, will clean up after migration is done
		// by default the AdvancedVisibilityWritingMode is set to ON for ES
		// if we change this dynamic config before deployment, ES will stop working and block task processing
		// we have to change it after deployment. But need to make sure double writes are working, so the only way is changing the behavior of this function
		if v.pinotVisibilityManager != nil && v.esVisibilityManager != nil {
			if err := esVisFunc(); err != nil {
				return err
			}
			return pinotVisFunc()
		} else if v.pinotVisibilityManager != nil {
			v.logger.Warn("ES visibility is not available to write, fall back to pinot visibility")
			return pinotVisFunc()
		} else if v.esVisibilityManager != nil {
			v.logger.Warn("Pinot visibility is not available to write, fall back to es visibility")
			return esVisFunc()
		} else {
			v.logger.Warn("advanced visibility is not available to write, fall back to basic visibility")
			return dbVisFunc()
		}
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
	case common.AdvancedVisibilityWritingModeTriple:
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

// For Pinot Migration uses. It will be a temporary usage
type userParameters struct {
	operation    string
	domainName   string
	workflowType string
	workflowID   string
	closeStatus  int // if it is -1, then will have --open flag in comparator workflow
	customQuery  string
	earliestTime int64
	latestTime   int64
}

// For Pinot Migration uses. It will be a temporary usage
// logUserQueryParameters will log user queries' parameters so that a comparator workflow can consume
func (v *pinotVisibilityTripleManager) logUserQueryParameters(userParam userParameters, domain string) {
	// Don't log if it is not enabled
	if !v.logCustomerQueryParameter(domain) {
		return
	}
	randNum := rand.Intn(10)
	if randNum != 5 { // Intentionally to have 1/10 chance to log custom query parameters
		return
	}

	v.logger.Info("Logging user query parameters for Pinot/ES response comparator...",
		tag.OperationName(userParam.operation),
		tag.WorkflowDomainName(userParam.domainName),
		tag.WorkflowType(userParam.workflowType),
		tag.WorkflowID(userParam.workflowID),
		tag.WorkflowCloseStatus(userParam.closeStatus),
		tag.VisibilityQuery(filterAttrPrefix(userParam.customQuery)),
		tag.EarliestTime(userParam.earliestTime),
		tag.LatestTime(userParam.latestTime))

}

// This is for only logUserQueryParameters (for Pinot Response comparator) usage.
// Be careful because there's a low possibility that there'll be false positive cases (shown in unit tests)
func filterAttrPrefix(str string) string {
	str = strings.Replace(str, "`Attr.", "", -1)
	return strings.Replace(str, "`", "", -1)
}

func (v *pinotVisibilityTripleManager) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		closeStatus:  -1, // is open. Will have --open flag in comparator workflow
		earliestTime: request.EarliestTime,
		latestTime:   request.LatestTime,
	}, request.Domain)

	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.ListOpenWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		closeStatus:  6, // 6 means not set closeStatus.
		earliestTime: request.EarliestTime,
		latestTime:   request.LatestTime,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.ListClosedWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		workflowType: request.WorkflowTypeName,
		closeStatus:  -1, // is open. Will have --open flag in comparator workflow
		earliestTime: request.EarliestTime,
		latestTime:   request.LatestTime,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.ListOpenWorkflowExecutionsByType(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		workflowType: request.WorkflowTypeName,
		closeStatus:  6, // 6 means not set closeStatus.
		earliestTime: request.EarliestTime,
		latestTime:   request.LatestTime,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.ListClosedWorkflowExecutionsByType(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		workflowID:   request.WorkflowID,
		closeStatus:  -1,
		earliestTime: request.EarliestTime,
		latestTime:   request.LatestTime,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		workflowID:   request.WorkflowID,
		closeStatus:  6, // 6 means not set closeStatus.
		earliestTime: request.EarliestTime,
		latestTime:   request.LatestTime,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *ListClosedWorkflowExecutionsByStatusRequest,
) (*ListWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		closeStatus:  int(request.Status),
		earliestTime: request.EarliestTime,
		latestTime:   request.LatestTime,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.ListClosedWorkflowExecutionsByStatus(ctx, request)
}

func (v *pinotVisibilityTripleManager) GetClosedWorkflowExecution(
	ctx context.Context,
	request *GetClosedWorkflowExecutionRequest,
) (*GetClosedWorkflowExecutionResponse, error) {
	earlistTime := int64(0) // this is to get all closed workflow execution
	latestTime := time.Now().UnixNano()

	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		closeStatus:  6, // 6 means not set closeStatus.
		earliestTime: earlistTime,
		latestTime:   latestTime,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.GetClosedWorkflowExecution(ctx, request)
}

func (v *pinotVisibilityTripleManager) ListWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		closeStatus:  6, // 6 means not set closeStatus.
		customQuery:  request.Query,
		earliestTime: -1,
		latestTime:   -1,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.ListWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) ScanWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.LIST),
		domainName:   request.Domain,
		closeStatus:  6, // 6 means not set closeStatus.
		customQuery:  request.Query,
		earliestTime: -1,
		latestTime:   -1,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.ScanWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) CountWorkflowExecutions(
	ctx context.Context,
	request *CountWorkflowExecutionsRequest,
) (*CountWorkflowExecutionsResponse, error) {
	v.logUserQueryParameters(userParameters{
		operation:    string(Operation.COUNT),
		domainName:   request.Domain,
		closeStatus:  6, // 6 means not set closeStatus.
		customQuery:  request.Query,
		earliestTime: -1,
		latestTime:   -1,
	}, request.Domain)
	manager := v.chooseVisibilityManagerForRead(ctx, request.Domain)
	return manager.CountWorkflowExecutions(ctx, request)
}

func (v *pinotVisibilityTripleManager) chooseVisibilityManagerForRead(ctx context.Context, domain string) VisibilityManager {
	if override := ctx.Value(ContextKey); override == VisibilityOverridePrimary {
		v.logger.Info("Pinot Migration log: Primary visibility manager was chosen for read.")
		return v.esVisibilityManager
	} else if override == VisibilityOverrideSecondary {
		v.logger.Info("Pinot Migration log: Secondary visibility manager was chosen for read.")
		return v.pinotVisibilityManager
	}

	var visibilityMgr VisibilityManager
	if v.readModeIsFromES(domain) {
		if v.esVisibilityManager != nil {
			visibilityMgr = v.esVisibilityManager
		} else {
			visibilityMgr = v.dbVisibilityManager
			v.logger.Warn("domain is configured to read from advanced visibility(ElasticSearch based) but it's not available, fall back to basic visibility",
				tag.WorkflowDomainName(domain))
		}
	} else if v.readModeIsFromPinot(domain) {
		if v.pinotVisibilityManager != nil {
			visibilityMgr = v.pinotVisibilityManager
		} else {
			visibilityMgr = v.dbVisibilityManager
			v.logger.Warn("domain is configured to read from advanced visibility(Pinot based) but it's not available, fall back to basic visibility",
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
	return visibilityMgr
}
