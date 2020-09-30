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
	"encoding/json"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

type (
	visibilityManagerImpl struct {
		serializer  PayloadSerializer
		persistence VisibilityStore
		logger      log.Logger
	}
)

// VisibilityEncoding is default encoding for visibility data
const VisibilityEncoding = common.EncodingTypeThriftRW

var _ VisibilityManager = (*visibilityManagerImpl)(nil)

// NewVisibilityManagerImpl returns new VisibilityManager
func NewVisibilityManagerImpl(persistence VisibilityStore, logger log.Logger) VisibilityManager {
	return &visibilityManagerImpl{
		serializer:  NewPayloadSerializer(),
		persistence: persistence,
		logger:      logger,
	}
}

func (v *visibilityManagerImpl) Close() {
	v.persistence.Close()
}

func (v *visibilityManagerImpl) GetName() string {
	return v.persistence.GetName()
}

func (v *visibilityManagerImpl) RecordWorkflowExecutionStarted(request *RecordWorkflowExecutionStartedRequest) error {
	req := &InternalRecordWorkflowExecutionStartedRequest{
		DomainUUID:         request.DomainUUID,
		WorkflowID:         request.Execution.GetWorkflowId(),
		RunID:              request.Execution.GetRunId(),
		WorkflowTypeName:   request.WorkflowTypeName,
		StartTimestamp:     request.StartTimestamp,
		ExecutionTimestamp: request.ExecutionTimestamp,
		WorkflowTimeout:    request.WorkflowTimeout,
		TaskID:             request.TaskID,
		TaskList:           request.TaskList,
		Memo:               v.serializeMemo(request.Memo, request.DomainUUID, request.Execution.GetWorkflowId(), request.Execution.GetRunId()),
		SearchAttributes:   request.SearchAttributes,
	}
	return v.persistence.RecordWorkflowExecutionStarted(context.TODO(), req)
}

func (v *visibilityManagerImpl) RecordWorkflowExecutionClosed(request *RecordWorkflowExecutionClosedRequest) error {
	req := &InternalRecordWorkflowExecutionClosedRequest{
		DomainUUID:         request.DomainUUID,
		WorkflowID:         request.Execution.GetWorkflowId(),
		RunID:              request.Execution.GetRunId(),
		WorkflowTypeName:   request.WorkflowTypeName,
		StartTimestamp:     request.StartTimestamp,
		ExecutionTimestamp: request.ExecutionTimestamp,
		TaskID:             request.TaskID,
		Memo:               v.serializeMemo(request.Memo, request.DomainUUID, request.Execution.GetWorkflowId(), request.Execution.GetRunId()),
		TaskList:           request.TaskList,
		SearchAttributes:   request.SearchAttributes,
		CloseTimestamp:     request.CloseTimestamp,
		Status:             request.Status,
		HistoryLength:      request.HistoryLength,
		RetentionSeconds:   request.RetentionSeconds,
	}
	return v.persistence.RecordWorkflowExecutionClosed(context.TODO(), req)
}

func (v *visibilityManagerImpl) UpsertWorkflowExecution(request *UpsertWorkflowExecutionRequest) error {
	req := &InternalUpsertWorkflowExecutionRequest{
		DomainUUID:         request.DomainUUID,
		WorkflowID:         request.Execution.GetWorkflowId(),
		RunID:              request.Execution.GetRunId(),
		WorkflowTypeName:   request.WorkflowTypeName,
		StartTimestamp:     request.StartTimestamp,
		ExecutionTimestamp: request.ExecutionTimestamp,
		TaskID:             request.TaskID,
		Memo:               v.serializeMemo(request.Memo, request.DomainUUID, request.Execution.GetWorkflowId(), request.Execution.GetRunId()),
		TaskList:           request.TaskList,
		SearchAttributes:   request.SearchAttributes,
	}
	return v.persistence.UpsertWorkflowExecution(context.TODO(), req)
}

func (v *visibilityManagerImpl) ListOpenWorkflowExecutions(request *ListWorkflowExecutionsRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListOpenWorkflowExecutions(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListClosedWorkflowExecutions(request *ListWorkflowExecutionsRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListClosedWorkflowExecutions(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListOpenWorkflowExecutionsByType(request *ListWorkflowExecutionsByTypeRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListOpenWorkflowExecutionsByType(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListClosedWorkflowExecutionsByType(request *ListWorkflowExecutionsByTypeRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListClosedWorkflowExecutionsByType(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListOpenWorkflowExecutionsByWorkflowID(request *ListWorkflowExecutionsByWorkflowIDRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListOpenWorkflowExecutionsByWorkflowID(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListClosedWorkflowExecutionsByWorkflowID(request *ListWorkflowExecutionsByWorkflowIDRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListClosedWorkflowExecutionsByWorkflowID(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ListClosedWorkflowExecutionsByStatus(request *ListClosedWorkflowExecutionsByStatusRequest) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListClosedWorkflowExecutionsByStatus(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) GetClosedWorkflowExecution(request *GetClosedWorkflowExecutionRequest) (*GetClosedWorkflowExecutionResponse, error) {
	internalResp, err := v.persistence.GetClosedWorkflowExecution(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalGetResponse(internalResp), nil
}

func (v *visibilityManagerImpl) DeleteWorkflowExecution(request *VisibilityDeleteWorkflowExecutionRequest) error {
	return v.persistence.DeleteWorkflowExecution(context.TODO(), request)
}

func (v *visibilityManagerImpl) ListWorkflowExecutions(request *ListWorkflowExecutionsRequestV2) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ListWorkflowExecutions(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) ScanWorkflowExecutions(request *ListWorkflowExecutionsRequestV2) (*ListWorkflowExecutionsResponse, error) {
	internalResp, err := v.persistence.ScanWorkflowExecutions(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	return v.convertInternalListResponse(internalResp), nil
}

func (v *visibilityManagerImpl) CountWorkflowExecutions(request *CountWorkflowExecutionsRequest) (*CountWorkflowExecutionsResponse, error) {
	return v.persistence.CountWorkflowExecutions(context.TODO(), request)
}

func (v *visibilityManagerImpl) convertInternalGetResponse(internalResp *InternalGetClosedWorkflowExecutionResponse) *GetClosedWorkflowExecutionResponse {
	if internalResp == nil {
		return nil
	}

	resp := &GetClosedWorkflowExecutionResponse{}
	resp.Execution = v.convertVisibilityWorkflowExecutionInfo(internalResp.Execution)
	return resp
}

func (v *visibilityManagerImpl) convertInternalListResponse(internalResp *InternalListWorkflowExecutionsResponse) *ListWorkflowExecutionsResponse {
	if internalResp == nil {
		return nil
	}

	resp := &ListWorkflowExecutionsResponse{}
	resp.Executions = make([]*shared.WorkflowExecutionInfo, len(internalResp.Executions))
	for i, execution := range internalResp.Executions {
		resp.Executions[i] = v.convertVisibilityWorkflowExecutionInfo(execution)
	}

	resp.NextPageToken = internalResp.NextPageToken
	return resp
}

func (v *visibilityManagerImpl) getSearchAttributes(attr map[string]interface{}) (*shared.SearchAttributes, error) {
	indexedFields := make(map[string][]byte)
	var err error
	var valBytes []byte
	for k, val := range attr {
		valBytes, err = json.Marshal(val)
		if err != nil {
			v.logger.Error("error when encode search attributes", tag.Value(val))
			continue
		}
		indexedFields[k] = valBytes
	}
	if err != nil {
		return nil, err
	}
	return &shared.SearchAttributes{
		IndexedFields: indexedFields,
	}, nil
}

func (v *visibilityManagerImpl) convertVisibilityWorkflowExecutionInfo(execution *VisibilityWorkflowExecutionInfo) *shared.WorkflowExecutionInfo {
	// special handling of ExecutionTime for cron or retry
	if execution.ExecutionTime.UnixNano() == 0 {
		execution.ExecutionTime = execution.StartTime
	}

	memo, err := v.serializer.DeserializeVisibilityMemo(execution.Memo)
	if err != nil {
		v.logger.Error("failed to deserialize memo",
			tag.WorkflowID(execution.WorkflowID),
			tag.WorkflowRunID(execution.RunID),
			tag.Error(err))
	}
	searchAttributes, err := v.getSearchAttributes(execution.SearchAttributes)
	if err != nil {
		v.logger.Error("failed to convert search attributes",
			tag.WorkflowID(execution.WorkflowID),
			tag.WorkflowRunID(execution.RunID),
			tag.Error(err))
	}

	convertedExecution := &shared.WorkflowExecutionInfo{
		Execution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(execution.WorkflowID),
			RunId:      common.StringPtr(execution.RunID),
		},
		Type: &shared.WorkflowType{
			Name: common.StringPtr(execution.TypeName),
		},
		StartTime:        common.Int64Ptr(execution.StartTime.UnixNano()),
		ExecutionTime:    common.Int64Ptr(execution.ExecutionTime.UnixNano()),
		Memo:             memo,
		SearchAttributes: searchAttributes,
		TaskList:         common.StringPtr(execution.TaskList),
	}

	// for close records
	if execution.Status != nil {
		convertedExecution.CloseTime = common.Int64Ptr(execution.CloseTime.UnixNano())
		convertedExecution.CloseStatus = execution.Status
		convertedExecution.HistoryLength = common.Int64Ptr(execution.HistoryLength)
	}

	return convertedExecution
}

func (v *visibilityManagerImpl) serializeMemo(visibilityMemo *shared.Memo, domainID, wID, rID string) *DataBlob {
	memo, err := v.serializer.SerializeVisibilityMemo(visibilityMemo, VisibilityEncoding)
	if err != nil {
		v.logger.WithTags(
			tag.WorkflowDomainID(domainID),
			tag.WorkflowID(wID),
			tag.WorkflowRunID(rID),
			tag.Error(err)).
			Error("Unable to encode visibility memo")
	}
	if memo == nil {
		return &DataBlob{}
	}
	return memo
}
