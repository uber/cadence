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

package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
	"github.com/uber/cadence/common/service/config"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

type (
	sqlVisibilityStore struct {
		sqlStore
		serializer p.PayloadSerializer
	}

	visibilityPageToken struct {
		Time  time.Time
		RunID string
	}
)

// NewSQLVisibilityStore creates an instance of ExecutionStore
func NewSQLVisibilityStore(cfg config.SQL, logger log.Logger) (p.VisibilityStore, error) {
	db, err := NewSQLDB(&cfg)
	if err != nil {
		return nil, err
	}
	return &sqlVisibilityStore{
		sqlStore: sqlStore{
			db:     db,
			logger: logger,
		},
		serializer: p.NewPayloadSerializer(),
	}, nil
}

func (s *sqlVisibilityStore) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionStartedRequest,
) error {
	memo := s.serializeMemo(request.Memo, request.DomainUUID, request.WorkflowID, request.RunID)
	_, err := s.db.InsertIntoVisibility(ctx, &sqlplugin.VisibilityRow{
		DomainID:         request.DomainUUID,
		WorkflowID:       request.WorkflowID,
		RunID:            request.RunID,
		StartTime:        time.Unix(0, request.StartTimestamp),
		ExecutionTime:    time.Unix(0, request.ExecutionTimestamp),
		WorkflowTypeName: request.WorkflowTypeName,
		Memo:             memo.Data,
		Encoding:         string(memo.GetEncoding()),
	})

	return err
}

func (s *sqlVisibilityStore) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionClosedRequest,
) error {
	memo := s.serializeMemo(request.Memo, request.DomainUUID, request.WorkflowID, request.RunID)
	closeTime := time.Unix(0, request.CloseTimestamp)
	result, err := s.db.ReplaceIntoVisibility(ctx, &sqlplugin.VisibilityRow{
		DomainID:         request.DomainUUID,
		WorkflowID:       request.WorkflowID,
		RunID:            request.RunID,
		StartTime:        time.Unix(0, request.StartTimestamp),
		ExecutionTime:    time.Unix(0, request.ExecutionTimestamp),
		WorkflowTypeName: request.WorkflowTypeName,
		CloseTime:        &closeTime,
		CloseStatus:      common.Int32Ptr(int32(*thrift.FromWorkflowExecutionCloseStatus(&request.Status))),
		HistoryLength:    &request.HistoryLength,
		Memo:             memo.Data,
		Encoding:         string(memo.GetEncoding()),
	})
	if err != nil {
		return err
	}
	noRowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("RecordWorkflowExecutionClosed rowsAffected error: %v", err)
	}
	if noRowsAffected > 2 { // either adds a new row or deletes old row and adds new row
		return fmt.Errorf("RecordWorkflowExecutionClosed unexpected numRows (%v) updated", noRowsAffected)
	}
	return nil
}

func (s *sqlVisibilityStore) UpsertWorkflowExecution(
	ctx context.Context,
	request *p.InternalUpsertWorkflowExecutionRequest,
) error {
	if p.IsNopUpsertWorkflowRequest(request) {
		return nil
	}
	return p.NewOperationNotSupportErrorForVis()
}

func (s *sqlVisibilityStore) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return s.listWorkflowExecutions("ListOpenWorkflowExecutions", request.NextPageToken, request.EarliestTime, request.LatestTime,
		func(readLevel *visibilityPageToken) ([]sqlplugin.VisibilityRow, error) {
			minStartTime := time.Unix(0, request.EarliestTime)
			return s.db.SelectFromVisibility(ctx, &sqlplugin.VisibilityFilter{
				DomainID:     request.DomainUUID,
				MinStartTime: &minStartTime,
				MaxStartTime: &readLevel.Time,
				RunID:        &readLevel.RunID,
				PageSize:     &request.PageSize,
			})
		})
}

func (s *sqlVisibilityStore) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return s.listWorkflowExecutions("ListClosedWorkflowExecutions", request.NextPageToken, request.EarliestTime, request.LatestTime,
		func(readLevel *visibilityPageToken) ([]sqlplugin.VisibilityRow, error) {
			minStartTime := time.Unix(0, request.EarliestTime)
			return s.db.SelectFromVisibility(ctx, &sqlplugin.VisibilityFilter{
				DomainID:     request.DomainUUID,
				MinStartTime: &minStartTime,
				MaxStartTime: &readLevel.Time,
				Closed:       true,
				RunID:        &readLevel.RunID,
				PageSize:     &request.PageSize,
			})
		})
}

func (s *sqlVisibilityStore) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return s.listWorkflowExecutions("ListOpenWorkflowExecutionsByType", request.NextPageToken, request.EarliestTime, request.LatestTime,
		func(readLevel *visibilityPageToken) ([]sqlplugin.VisibilityRow, error) {
			minStartTime := time.Unix(0, request.EarliestTime)
			return s.db.SelectFromVisibility(ctx, &sqlplugin.VisibilityFilter{
				DomainID:         request.DomainUUID,
				MinStartTime:     &minStartTime,
				MaxStartTime:     &readLevel.Time,
				RunID:            &readLevel.RunID,
				WorkflowTypeName: &request.WorkflowTypeName,
				PageSize:         &request.PageSize,
			})
		})
}

func (s *sqlVisibilityStore) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return s.listWorkflowExecutions("ListClosedWorkflowExecutionsByType", request.NextPageToken, request.EarliestTime, request.LatestTime,
		func(readLevel *visibilityPageToken) ([]sqlplugin.VisibilityRow, error) {
			minStartTime := time.Unix(0, request.EarliestTime)
			return s.db.SelectFromVisibility(ctx, &sqlplugin.VisibilityFilter{
				DomainID:         request.DomainUUID,
				MinStartTime:     &minStartTime,
				MaxStartTime:     &readLevel.Time,
				Closed:           true,
				RunID:            &readLevel.RunID,
				WorkflowTypeName: &request.WorkflowTypeName,
				PageSize:         &request.PageSize,
			})
		})
}

func (s *sqlVisibilityStore) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return s.listWorkflowExecutions("ListOpenWorkflowExecutionsByWorkflowID", request.NextPageToken, request.EarliestTime, request.LatestTime,
		func(readLevel *visibilityPageToken) ([]sqlplugin.VisibilityRow, error) {
			minStartTime := time.Unix(0, request.EarliestTime)
			return s.db.SelectFromVisibility(ctx, &sqlplugin.VisibilityFilter{
				DomainID:     request.DomainUUID,
				MinStartTime: &minStartTime,
				MaxStartTime: &readLevel.Time,
				RunID:        &readLevel.RunID,
				WorkflowID:   &request.WorkflowID,
				PageSize:     &request.PageSize,
			})
		})
}

func (s *sqlVisibilityStore) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return s.listWorkflowExecutions("ListClosedWorkflowExecutionsByWorkflowID", request.NextPageToken, request.EarliestTime, request.LatestTime,
		func(readLevel *visibilityPageToken) ([]sqlplugin.VisibilityRow, error) {
			minStartTime := time.Unix(0, request.EarliestTime)
			return s.db.SelectFromVisibility(ctx, &sqlplugin.VisibilityFilter{
				DomainID:     request.DomainUUID,
				MinStartTime: &minStartTime,
				MaxStartTime: &readLevel.Time,
				Closed:       true,
				RunID:        &readLevel.RunID,
				WorkflowID:   &request.WorkflowID,
				PageSize:     &request.PageSize,
			})
		})
}

func (s *sqlVisibilityStore) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *p.InternalListClosedWorkflowExecutionsByStatusRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return s.listWorkflowExecutions("ListClosedWorkflowExecutionsByStatus", request.NextPageToken, request.EarliestTime, request.LatestTime,
		func(readLevel *visibilityPageToken) ([]sqlplugin.VisibilityRow, error) {
			minStartTime := time.Unix(0, request.EarliestTime)
			return s.db.SelectFromVisibility(ctx, &sqlplugin.VisibilityFilter{
				DomainID:     request.DomainUUID,
				MinStartTime: &minStartTime,
				MaxStartTime: &readLevel.Time,
				Closed:       true,
				RunID:        &readLevel.RunID,
				CloseStatus:  common.Int32Ptr(int32(*thrift.FromWorkflowExecutionCloseStatus(&request.Status))),
				PageSize:     &request.PageSize,
			})
		})
}

func (s *sqlVisibilityStore) GetClosedWorkflowExecution(
	ctx context.Context,
	request *p.InternalGetClosedWorkflowExecutionRequest,
) (*p.InternalGetClosedWorkflowExecutionResponse, error) {
	execution := request.Execution
	rows, err := s.db.SelectFromVisibility(ctx, &sqlplugin.VisibilityFilter{
		DomainID: request.DomainUUID,
		Closed:   true,
		RunID:    execution.RunID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &workflow.EntityNotExistsError{
				Message: fmt.Sprintf("Workflow execution not found.  WorkflowId: %v, RunId: %v",
					execution.GetWorkflowID(), execution.GetRunID()),
			}
		}
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetClosedWorkflowExecution operation failed. Select failed: %v", err),
		}
	}
	rows[0].DomainID = request.DomainUUID
	rows[0].RunID = execution.GetRunID()
	rows[0].WorkflowID = execution.GetWorkflowID()
	return &p.InternalGetClosedWorkflowExecutionResponse{Execution: s.rowToInfo(&rows[0])}, nil
}

func (s *sqlVisibilityStore) DeleteWorkflowExecution(
	ctx context.Context,
	request *p.VisibilityDeleteWorkflowExecutionRequest,
) error {
	_, err := s.db.DeleteFromVisibility(ctx, &sqlplugin.VisibilityFilter{
		DomainID: request.DomainID,
		RunID:    &request.RunID,
	})
	if err != nil {
		return &workflow.InternalServiceError{Message: err.Error()}
	}
	return nil
}

func (s *sqlVisibilityStore) ListWorkflowExecutions(
	ctx context.Context,
	request *p.ListWorkflowExecutionsByQueryRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.NewOperationNotSupportErrorForVis()
}

func (s *sqlVisibilityStore) ScanWorkflowExecutions(
	ctx context.Context,
	request *p.ListWorkflowExecutionsByQueryRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.NewOperationNotSupportErrorForVis()
}

func (s *sqlVisibilityStore) CountWorkflowExecutions(
	ctx context.Context,
	request *p.CountWorkflowExecutionsRequest,
) (*p.CountWorkflowExecutionsResponse, error) {
	return nil, p.NewOperationNotSupportErrorForVis()
}

func (s *sqlVisibilityStore) rowToInfo(row *sqlplugin.VisibilityRow) *p.InternalVisibilityWorkflowExecutionInfo {
	if row.ExecutionTime.UnixNano() == 0 {
		row.ExecutionTime = row.StartTime
	}
	memo, err := s.serializer.DeserializeVisibilityMemo(p.NewDataBlob(row.Memo, common.EncodingType(row.Encoding)))
	if err != nil {
		s.logger.Error("failed to deserialize memo",
			tag.WorkflowID(row.WorkflowID),
			tag.WorkflowRunID(row.RunID),
			tag.Error(err))
	}
	info := &p.InternalVisibilityWorkflowExecutionInfo{
		WorkflowID:    row.WorkflowID,
		RunID:         row.RunID,
		TypeName:      row.WorkflowTypeName,
		StartTime:     row.StartTime,
		ExecutionTime: row.ExecutionTime,
		Memo:          thrift.ToMemo(memo),
	}
	if row.CloseStatus != nil {
		status := workflow.WorkflowExecutionCloseStatus(*row.CloseStatus)
		info.Status = thrift.ToWorkflowExecutionCloseStatus(&status)
		info.CloseTime = *row.CloseTime
		info.HistoryLength = *row.HistoryLength
	}
	return info
}

func (s *sqlVisibilityStore) listWorkflowExecutions(opName string, pageToken []byte, earliestTime int64, latestTime int64, selectOp func(readLevel *visibilityPageToken) ([]sqlplugin.VisibilityRow, error)) (*p.InternalListWorkflowExecutionsResponse, error) {
	var readLevel *visibilityPageToken
	var err error
	if len(pageToken) > 0 {
		readLevel, err = s.deserializePageToken(pageToken)
		if err != nil {
			return nil, err
		}
	} else {
		readLevel = &visibilityPageToken{Time: time.Unix(0, latestTime), RunID: ""}
	}
	rows, err := selectOp(readLevel)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("%v operation failed. Select failed: %v", opName, err),
		}
	}
	if len(rows) == 0 {
		return &p.InternalListWorkflowExecutionsResponse{}, nil
	}

	var infos = make([]*p.InternalVisibilityWorkflowExecutionInfo, len(rows))
	for i, row := range rows {
		infos[i] = s.rowToInfo(&row)
	}
	var nextPageToken []byte
	lastRow := rows[len(rows)-1]
	lastStartTime := lastRow.StartTime
	if lastStartTime.Sub(time.Unix(0, earliestTime)).Nanoseconds() > 0 {
		nextPageToken, err = s.serializePageToken(&visibilityPageToken{
			Time:  lastStartTime,
			RunID: lastRow.RunID,
		})
		if err != nil {
			return nil, err
		}
	}
	return &p.InternalListWorkflowExecutionsResponse{
		Executions:    infos,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *sqlVisibilityStore) deserializePageToken(data []byte) (*visibilityPageToken, error) {
	var token visibilityPageToken
	err := json.Unmarshal(data, &token)
	return &token, err
}

func (s *sqlVisibilityStore) serializePageToken(token *visibilityPageToken) ([]byte, error) {
	data, err := json.Marshal(token)
	return data, err
}

func (s *sqlVisibilityStore) serializeMemo(visibilityMemo *types.Memo, domainID, wID, rID string) *p.DataBlob {
	memo, err := s.serializer.SerializeVisibilityMemo(thrift.FromMemo(visibilityMemo), common.EncodingTypeThriftRW)
	if err != nil {
		s.logger.WithTags(
			tag.WorkflowDomainID(domainID),
			tag.WorkflowID(wID),
			tag.WorkflowRunID(rID),
			tag.Error(err)).
			Error("Unable to encode visibility memo")
	}
	if memo == nil {
		return &p.DataBlob{}
	}
	return memo
}
