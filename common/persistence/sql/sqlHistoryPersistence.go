// Copyright (c) 2018 Uber Technologies, Inc.
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
	"fmt"
	"github.com/iancoleman/strcase"

	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"

	"github.com/hmgle/sqlx"
	"strings"
)

type (
	sqlHistoryManager struct {
		db      *sqlx.DB
		shardID int
		logger  bark.Logger
	}

	eventsRow struct {
		DomainID     string
		WorkflowID   string
		RunID        string
		FirstEventID int64
		Data         *[]byte
		DataEncoding string
		DataVersion  int64

		RangeID int64
		TxID    int64
	}
)

const (
	appendHistorySQLQuery = `INSERT INTO events (
domain_id,workflow_id,run_id,first_event_id,data,data_encoding,data_version
) VALUES (
:domain_id,:workflow_id,:run_id,:first_event_id,:data,:data_encoding,:data_version
);`

	overwriteHistorySQLQuery = `UPDATE events
SET
domain_id = :domain_id,
workflow_id = :workflow_id,
run_id = :run_id,
first_event_id = :first_event_id,
data = :data,
data_encoding = :data_encoding,
data_version = :data_version
WHERE
domain_id = :domain_id AND 
workflow_id = :workflow_id AND 
run_id = :run_id AND 
first_event_id = :first_event_id`

	pollHistorySQLQuery = `SELECT 1 FROM events WHERE domain_id = :domain_id AND 
workflow_id= :workflow_id AND run_id= :run_id AND first_event_id= :first_event_id`

	getWorkflowExecutionHistorySQLQuery = `SELECT first_event_id, data, data_encoding, data_version
FROM events
WHERE
domain_id = ? AND
workflow_id = ? AND
run_id = ? AND
first_event_id >= ? AND
first_event_id < ?`

	deleteWorkflowExecutionHistorySQLQuery = `DELETE FROM events WHERE
domain_id = ? AND workflow_id = ? AND run_id = ?`

	lockRangeIDAndTxIDSQLQuery = `SELECT range_id, tx_id FROM events WHERE
domain_id = ? AND workflow_id = ? AND run_id = ? AND first_event_id = ?`
)

func (m *sqlHistoryManager) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func NewHistoryPersistence(username, password, host, port, dbName string, logger bark.Logger) (persistence.HistoryManager, error) {
	var db, err = sqlx.Connect("mysql",
		fmt.Sprintf(Dsn, username, password, host, port, dbName))
	if err != nil {
		return nil, err
	}
	db.MapperFunc(strcase.ToSnake)
	return &sqlHistoryManager{
		db:     db,
		logger: logger,
	}, nil
}

func (m *sqlHistoryManager) AppendHistoryEvents(request *persistence.AppendHistoryEventsRequest) error {
	arg := &eventsRow{
		DomainID:     request.DomainID,
		WorkflowID:   *request.Execution.WorkflowId,
		RunID:        *request.Execution.RunId,
		FirstEventID: request.FirstEventID,
		Data:         takeAddressIfNotNil(request.Events.Data),
		DataEncoding: string(request.Events.EncodingType),
		DataVersion:  int64(request.Events.Version),
		RangeID:      request.RangeID,
		TxID:         request.TransactionID,
	}

	if request.Overwrite {
		tx, err := m.db.Beginx()
		if err != nil {
			return &workflow.InternalServiceError{
				Message: fmt.Sprintf("AppendHistoryEvents operation failed. Failed to begin transaction for overwrite. Error: %v", err),
			}
		}
		defer tx.Rollback()

		if err := lockAndCheckRangeIDAndTxID(tx,
			request.RangeID,
			request.TransactionID,
			request.DomainID,
			*request.Execution.WorkflowId,
			*request.Execution.RunId,
			request.FirstEventID); err != nil {
			switch err.(type) {
			case *persistence.ConditionFailedError:
				return &persistence.ConditionFailedError{
					Msg: fmt.Sprintf("AppendHistoryEvents operation failed. Overwrite failed. Error: %v", err),
				}
			default:
				return &workflow.InternalServiceError{
					Message: fmt.Sprintf("AppendHistoryEvents operation failed. Failed to lock row for overwrite. Error: %v", err),
				}
			}
		}

		if result, err := tx.NamedExec(overwriteHistorySQLQuery, arg); err != nil {
			return &workflow.InternalServiceError{
				Message: fmt.Sprintf("AppendHistoryEvents operation failed. Update failed. Error: %v", err),
			}
		} else {
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return &workflow.InternalServiceError{
					Message: fmt.Sprintf("AppendHistoryEvents operation failed. Failed to check number of rows updated. Error: %v", err),
				}
			}
			if rowsAffected != 1 {
				return &workflow.InternalServiceError{
					Message: fmt.Sprintf("AppendHistoryEvents operation failed. Updated %v rows instead of one.", rowsAffected),
				}
			}
		}

		if err := tx.Commit(); err != nil {
			return &workflow.InternalServiceError{
				Message: fmt.Sprintf("AppendHistoryEvents operation failed. Failed to lock row. Error: %v", err),
			}
		}
	} else {
		if _, err := m.db.NamedExec(appendHistorySQLQuery, arg); err != nil {
			// TODO Find another way to do this without inspecting the error message (?)
			// Error 1062 indicates a duplicate primary key i.e. the row already exists,
			// so we don't do the insert and return a ConditionalUpdate error.
			if strings.HasPrefix(err.Error(), "Error 1062") {
				return &persistence.ConditionFailedError{
					Msg: fmt.Sprintf("AppendHistoryEvents operaiton failed. Couldn't insert since row already existed. Erorr: %v", err),
				}
			}
			return &workflow.InternalServiceError{
				Message: fmt.Sprintf("AppendHistoryEvents operation failed. Insert failed. Error: %v", err),
			}
		}
	}

	return nil
}

func (m *sqlHistoryManager) GetWorkflowExecutionHistory(request *persistence.GetWorkflowExecutionHistoryRequest) (*persistence.GetWorkflowExecutionHistoryResponse,
	error) {
	var rows []eventsRow
	if err := m.db.Select(&rows,
		getWorkflowExecutionHistorySQLQuery,
		request.DomainID,
		request.Execution.WorkflowId,
		request.Execution.RunId,
		request.FirstEventID,
		request.NextEventID); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetWorkflowExecutionHistory operation failed. Select failed. Error: %v", err),
		}
	}

	if len(rows) == 0 {
		return nil, &workflow.EntityNotExistsError{
			Message: fmt.Sprintf("Workflow execution history not found.  WorkflowId: %v, RunId: %v",
				*request.Execution.WorkflowId, *request.Execution.RunId),
		}
	}

	events := make([]persistence.SerializedHistoryEventBatch, len(rows))
	for i, v := range rows {
		events[i].EncodingType = common.EncodingType(v.DataEncoding)
		events[i].Version = int(v.DataVersion)
		if v.Data != nil {
			events[i].Data = *v.Data
		}
	}

	return &persistence.GetWorkflowExecutionHistoryResponse{
		History:        events,
		NextPageToken: []byte{},
	}, nil
}

func (m *sqlHistoryManager) DeleteWorkflowExecutionHistory(request *persistence.DeleteWorkflowExecutionHistoryRequest) error {
	if _, err := m.db.Exec(deleteWorkflowExecutionHistorySQLQuery, request.DomainID, request.Execution.WorkflowId, request.Execution.RunId); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("DeleteWorkflowExecutionHistory operation failed. Error: %v", err),
		}
	}
	return nil
}

func lockAndCheckRangeIDAndTxID(tx *sqlx.Tx,
	maxRangeID int64,
	maxTxIDPlusOne int64,
	domainID string,
	workflowID string,
	runID string,
	firstEventID int64) error {
	var row eventsRow
	if err := tx.Get(&row,
		lockRangeIDAndTxIDSQLQuery,
		domainID,
		workflowID,
		runID,
		firstEventID); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to lock range ID and tx ID. Get failed. Error: %v", err),
		}
	}
	if !(row.RangeID <= maxRangeID) {
		return &persistence.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to lock range ID and tx ID. %v should've been at most %v.", row.RangeID, maxRangeID),
		}
	} else if !(row.TxID < maxTxIDPlusOne) {
		return &persistence.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to lock range ID and tx ID. %v should've been strictly less than %v.", row.TxID, maxTxIDPlusOne),
		}
	}
	return nil
}
