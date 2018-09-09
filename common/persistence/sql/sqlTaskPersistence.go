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

	"database/sql"
	"github.com/hmgle/sqlx"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/persistence"
	"time"
)

type (
	sqlTaskManager struct {
		db *sqlx.DB
	}

	tasksRow struct {
		DomainID     string    `db:"domain_id"`
		WorkflowID   string    `db:"workflow_id"`
		RunID        string    `db:"run_id"`
		ScheduleID   int64     `db:"schedule_id"`
		TaskID       int64     `db:"task_id"`
		TaskListName string    `db:"task_list_name"`
		TaskListType int64     `db:"task_list_type"`
		ExpiryTs     time.Time `db:"expiry_ts"`
	}

	tasksListsRow struct {
		DomainID string    `db:"domain_id"`
		RangeID  int64     `db:"range_id"`
		Name     string    `db:"name"`
		Type     int64     `db:"type"`
		AckLevel int64     `db:"ack_level"`
		Kind     int64     `db:"kind"`
		ExpiryTs time.Time `db:"expiry_ts"`
	}

	updateTaskListsRow struct {
		tasksListsRow
		OldRangeID int64 `db:"old_range_id"`
	}
)

const (
	taskListCreatePart = `INTO task_lists 
(domain_id, range_id, name, type, ack_level, kind, expiry_ts)
VALUES
(:domain_id, :range_id, :name, :type, :ack_level, :kind, :expiry_ts)`

	// (default range ID: initialRangeID == 1)
	createTaskListSQLQuery = `INSERT ` + taskListCreatePart

	updateTaskListWithTTLSQLQuery = `REPLACE ` + taskListCreatePart

	updateTaskListSQLQuery = `UPDATE task_lists SET
domain_id = :domain_id,
range_id = :range_id,
name = :name,
type = :type,
ack_level = :ack_level,
kind = :kind,
expiry_ts = :expiry_ts
WHERE
domain_id = :domain_id AND
name = :name AND
type = :type
`

	getTaskListSQLQuery = `SELECT domain_id, range_id, name, type, ack_level, kind, expiry_ts FROM task_lists WHERE
domain_id = ? AND 
name = ? AND 
type = ?`

	lockTaskListSQLQuery = `SELECT range_id FROM task_lists WHERE
domain_id = ? AND
name = ? AND
type = ?
FOR UPDATE
`

	getTaskSQLQuery = `SELECT
domain_id, workflow_id, run_id, schedule_id, task_list_name, task_list_type, task_id, expiry_ts
FROM tasks
WHERE
domain_id = ? AND
task_list_name = ? AND 
task_list_type = ? AND
task_id > ? AND
task_id <= ?
`

	createTaskSQLQuery = `INSERT INTO tasks
(domain_id, workflow_id, run_id, schedule_id, task_list_name, task_list_type, task_id, expiry_ts)
VALUES
(:domain_id, :workflow_id, :run_id, :schedule_id, :task_list_name, :task_list_type, :task_id, :expiry_ts)`
)

func NewTaskPersistence(username, password, host, port, dbName string) (persistence.TaskManager, error) {
	var db, err = sqlx.Connect("mysql",
		fmt.Sprintf(Dsn, username, password, host, port, dbName))
	if err != nil {
		return nil, err
	}

	return &sqlTaskManager{
		db: db,
	}, nil
}

func (m *sqlTaskManager) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m *sqlTaskManager) LeaseTaskList(request *persistence.LeaseTaskListRequest) (*persistence.LeaseTaskListResponse, error) {
	var row tasksListsRow
	var rangeID int64 = 0
	var ackLevel int64 = 0
	if err := m.db.Get(&row, getTaskListSQLQuery, request.DomainID, request.TaskList, request.TaskType); err != nil {
		if err == sql.ErrNoRows {
			// The task list does not exist. Create it.
			if _, err := m.db.NamedExec(createTaskListSQLQuery,
				&tasksListsRow{
					DomainID: request.DomainID,
					RangeID:  rangeID + 1,
					Name:     request.TaskList,
					Type:     int64(request.TaskType),
					AckLevel: ackLevel,
					Kind:     int64(request.TaskListKind),
					ExpiryTs: MaximumExpiryTs,
				}); err != nil {
				return nil, &workflow.InternalServiceError{
					Message: fmt.Sprintf("LeaseTaskList operation failed. Failed to make task list %v of type %v. Error: %v", request.TaskList, request.TaskType, err),
				}
			}
		} else {
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("LeaseTaskList operation failed. Failed to check if task list existed. Error: %v", err),
			}
		}
	} else {
		// The task list exists.
		rangeID = row.RangeID
		ackLevel = row.AckLevel

		// We need to separately check the condition and do the
		// update because we want to throw different error codes.
		// Since we need to do things separately (in a transaction), we need to take a lock.
		tx, err := m.db.Beginx()
		if err != nil {
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("LeaseTaskList operation failed. Failed to begin transaction. Error: %v", err),
			}
		}
		defer tx.Rollback()

		if err := lockAndCheckTaskListRangeID(tx, request.DomainID, request.TaskList, request.TaskType, row.RangeID); err != nil {
			switch err.(type) {
			case *persistence.ConditionFailedError:
				return nil, err
			default:
				return nil, &workflow.InternalServiceError{
					Message: fmt.Sprintf("LeaseTaskList operation failed. Error: %v", err),
				}
			}
		}

		if result, err := tx.NamedExec(updateTaskListSQLQuery,
			&updateTaskListsRow{
				tasksListsRow{
					DomainID: row.DomainID,
					RangeID:  row.RangeID + 1,
					Name:     row.Name,
					Type:     row.Type,
					AckLevel: row.AckLevel,
					Kind:     row.Kind,
					ExpiryTs: row.ExpiryTs,
				},
				row.RangeID,
			}); err != nil {
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("LeaseTaskList operation failed. Failed to lease task list %v of type %v. Error: %v", request.TaskList, request.TaskType, err),
			}
		} else {
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return nil, &workflow.InternalServiceError{
					Message: fmt.Sprintf("LeaseTaskList operation failed. Failed to check if lease was successful. Error: %v", err),
				}
			}
			if rowsAffected == 0 {
				return nil, &workflow.InternalServiceError{
					Message: fmt.Sprintf("LeaseTaskList operation failed. Updated %v rows instead of 1", rowsAffected),
				}
			}
		}

		if err := tx.Commit(); err != nil {
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("LeaseTaskList operation failed. Failed to commit transaction. Error: %v", err),
			}
		}

	}

	return &persistence.LeaseTaskListResponse{&persistence.TaskListInfo{
		DomainID: request.DomainID,
		Name:     request.TaskList,
		TaskType: request.TaskType,
		RangeID:  rangeID + 1,
		AckLevel: ackLevel,
		Kind:     request.TaskListKind,
	}}, nil
}

func (m *sqlTaskManager) UpdateTaskList(request *persistence.UpdateTaskListRequest) (*persistence.UpdateTaskListResponse, error) {
	if request.TaskListInfo.Kind == persistence.TaskListKindSticky {
		// If sticky, update with TTL
		if _, err := m.db.NamedExec(updateTaskListWithTTLSQLQuery, &tasksListsRow{
			DomainID: request.TaskListInfo.DomainID,
			RangeID:  request.TaskListInfo.RangeID,
			Name:     request.TaskListInfo.Name,
			Type:     int64(request.TaskListInfo.TaskType),
			AckLevel: request.TaskListInfo.AckLevel,
			Kind:     int64(request.TaskListInfo.Kind),
			ExpiryTs: stickyTaskListTTL(),
		}); err != nil {
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("UpdateTaskList operation failed. Failed to make sticky task list. Error: %v", err),
			}
		}
	} else {
		tx, err := m.db.Beginx()
		if err != nil {
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("UpdateTaskList operation failed. Failed to begin transaction. Error: %v", err),
			}
		}
		defer tx.Rollback()

		if err := lockAndCheckTaskListRangeID(tx, request.TaskListInfo.DomainID, request.TaskListInfo.Name, request.TaskListInfo.TaskType, request.TaskListInfo.RangeID); err != nil {
			switch err.(type) {
			case *persistence.ConditionFailedError:
				return nil, err
			default:
				return nil, &workflow.InternalServiceError{
					Message: fmt.Sprintf("UpdateTaskList operation failed. Error: %v", err),
				}
			}
		}
		if result, err := tx.NamedExec(updateTaskListSQLQuery,
			&updateTaskListsRow{
				tasksListsRow{
					request.TaskListInfo.DomainID,
					request.TaskListInfo.RangeID,
					request.TaskListInfo.Name,
					int64(request.TaskListInfo.TaskType),
					request.TaskListInfo.AckLevel,
					int64(request.TaskListInfo.Kind),
					MaximumExpiryTs,
				},
				request.TaskListInfo.RangeID,
			}); err != nil {
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("UpdateTaskList operation failed. Failed to update task list. Error: %v", err),
			}
		} else {
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return nil, &workflow.InternalServiceError{
					Message: fmt.Sprintf("UpdateTaskList operation failed. Failed to verify how many rows were affected. Error: %v", err),
				}
			}
			if rowsAffected != 1 {
				return nil, &workflow.InternalServiceError{
					Message: fmt.Sprintf("UpdateTaskList operation failed. %v rows were affected instead of 1.", rowsAffected),
				}
			}
		}
		if err := tx.Commit(); err != nil {
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("UpdateTaskList operation failed. Failed to commit transaction. Error: %v", err),
			}
		}
	}

	return &persistence.UpdateTaskListResponse{}, nil
}

func (m *sqlTaskManager) CreateTasks(request *persistence.CreateTasksRequest) (*persistence.CreateTasksResponse, error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateTasks operation failed. Failed to begin transaction. Error: %v", err),
		}
	}
	defer tx.Rollback()

	tasksRows := make([]tasksRow, len(request.Tasks))
	for i, v := range request.Tasks {
		tasksRows[i] = tasksRow{
			DomainID:     v.Data.DomainID,
			WorkflowID:   v.Data.WorkflowID,
			RunID:        v.Data.RunID,
			ScheduleID:   v.Data.ScheduleID,
			TaskListName: request.TaskListInfo.Name,
			TaskListType: int64(request.TaskListInfo.TaskType),
			TaskID:       v.TaskID,
			ExpiryTs:     MaximumExpiryTs,
		}
	}

	query, args, err := m.db.BindNamed(createTaskSQLQuery, tasksRows)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateTasks operation failed. Failed to bind statement. Error: %v", err),
		}
	}

	if _, err := tx.Exec(query, args...); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateTasks operation failed. Failed to create tasks. Error: %v", err),
		}
	}

	// Lock task list before committing.

	if err := lockAndCheckTaskListRangeID(tx, request.TaskListInfo.DomainID, request.TaskListInfo.Name, request.TaskListInfo.TaskType, request.TaskListInfo.RangeID); err != nil {
		switch err.(type) {
		case *persistence.ConditionFailedError:
			return nil, err
		default:
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("CreateTasks operation failed. Error: %v", err),
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateTasks operation failed. Failed to commit transaction. Error: %v", err),
		}
	}

	return &persistence.CreateTasksResponse{}, nil
}

func (m *sqlTaskManager) GetTasks(request *persistence.GetTasksRequest) (*persistence.GetTasksResponse, error) {
	var rows []tasksRow
	if err := m.db.Select(&rows, getTaskSQLQuery, request.DomainID, request.TaskList, request.TaskType, request.ReadLevel, request.MaxReadLevel); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetTasks operation failed. Failed to get rows. Error: %v", err),
		}
	}

	var tasks = make([]*persistence.TaskInfo, len(rows))
	for i, v := range rows {
		tasks[i] = &persistence.TaskInfo{
			DomainID:               v.DomainID,
			WorkflowID:             v.WorkflowID,
			RunID:                  v.RunID,
			TaskID:                 v.TaskID,
			ScheduleID:             v.ScheduleID,
			ScheduleToStartTimeout: 0,
		}
	}

	return &persistence.GetTasksResponse{tasks}, nil
}

func (m *sqlTaskManager) CompleteTask(request *persistence.CompleteTaskRequest) error {
	return nil
}

func lockAndCheckTaskListRangeID(tx *sqlx.Tx, domainID, name string, taskListType int, oldRangeID int64) error {
	var rangeID int64
	if err := tx.Get(&rangeID, lockTaskListSQLQuery, domainID, name, taskListType); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to lock task list. Error: %v", err),
		}
	}

	if rangeID != oldRangeID {
		return &persistence.ConditionFailedError{
			Msg: fmt.Sprintf("Task list range ID was %v when it was should have been %v", rangeID, oldRangeID),
		}
	}

	return nil
}

func stickyTaskListTTL() time.Time {
	return time.Now().Add(24 * time.Hour)
}
