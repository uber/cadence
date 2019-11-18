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

package sqlshared

import (
	"database/sql"
	"fmt"

	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

// InsertIntoTasks inserts one or more rows into tasks table
func (mdb *DB) InsertIntoTasks(rows []sqldb.TasksRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.CreateTaskQuery(), rows)
}

// SelectFromTasks reads one or more rows from tasks table
func (mdb *DB) SelectFromTasks(filter *sqldb.TasksFilter) ([]sqldb.TasksRow, error) {
	var err error
	var rows []sqldb.TasksRow
	switch {
	case filter.MaxTaskID != nil:
		err = mdb.conn.Select(&rows, mdb.driver.GetTaskMinMaxQuery(), filter.DomainID,
			filter.TaskListName, filter.TaskType, *filter.MinTaskID, *filter.MaxTaskID, *filter.PageSize)
	default:
		err = mdb.conn.Select(&rows, mdb.driver.GetTaskMinQuery(), filter.DomainID,
			filter.TaskListName, filter.TaskType, *filter.MinTaskID, *filter.PageSize)
	}
	if err != nil {
		return nil, err
	}
	return rows, err
}

// DeleteFromTasks deletes one or more rows from tasks table
func (mdb *DB) DeleteFromTasks(filter *sqldb.TasksFilter) (sql.Result, error) {
	if filter.TaskIDLessThanEquals != nil {
		if filter.Limit == nil || *filter.Limit == 0 {
			return nil, fmt.Errorf("missing limit parameter")
		}
		return mdb.conn.Exec(mdb.driver.RangeDeleteTaskQuery(),
			filter.DomainID, filter.TaskListName, filter.TaskType, *filter.TaskIDLessThanEquals, *filter.Limit)
	}
	return mdb.conn.Exec(mdb.driver.DeleteTaskQuery(), filter.DomainID, filter.TaskListName, filter.TaskType, *filter.TaskID)
}

// InsertIntoTaskLists inserts one or more rows into task_lists table
func (mdb *DB) InsertIntoTaskLists(row *sqldb.TaskListsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.CreateTaskListQuery(), row)
}

// ReplaceIntoTaskLists replaces one or more rows in task_lists table
func (mdb *DB) ReplaceIntoTaskLists(row *sqldb.TaskListsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.ReplaceTaskListQuery(), row)
}

// UpdateTaskLists updates a row in task_lists table
func (mdb *DB) UpdateTaskLists(row *sqldb.TaskListsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.UpdateTaskListQuery(), row)
}

// SelectFromTaskLists reads one or more rows from task_lists table
func (mdb *DB) SelectFromTaskLists(filter *sqldb.TaskListsFilter) ([]sqldb.TaskListsRow, error) {
	switch {
	case filter.DomainID != nil && filter.Name != nil && filter.TaskType != nil:
		return mdb.selectFromTaskLists(filter)
	case filter.DomainIDGreaterThan != nil && filter.NameGreaterThan != nil && filter.TaskTypeGreaterThan != nil && filter.PageSize != nil:
		return mdb.rangeSelectFromTaskLists(filter)
	default:
		return nil, fmt.Errorf("invalid set of query filter params")
	}
}

func (mdb *DB) selectFromTaskLists(filter *sqldb.TaskListsFilter) ([]sqldb.TaskListsRow, error) {
	var err error
	var row sqldb.TaskListsRow
	err = mdb.conn.Get(&row, mdb.driver.GetTaskListQuery(), filter.ShardID, *filter.DomainID, *filter.Name, *filter.TaskType)
	if err != nil {
		return nil, err
	}
	return []sqldb.TaskListsRow{row}, err
}

func (mdb *DB) rangeSelectFromTaskLists(filter *sqldb.TaskListsFilter) ([]sqldb.TaskListsRow, error) {
	var err error
	var rows []sqldb.TaskListsRow
	err = mdb.conn.Select(&rows, mdb.driver.ListTaskListQuery(),
		filter.ShardID, *filter.DomainIDGreaterThan, *filter.NameGreaterThan, *filter.TaskTypeGreaterThan, *filter.PageSize)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].ShardID = filter.ShardID
	}
	return rows, nil
}

// DeleteFromTaskLists deletes a row from task_lists table
func (mdb *DB) DeleteFromTaskLists(filter *sqldb.TaskListsFilter) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.DeleteTaskListQuery(), filter.ShardID, *filter.DomainID, *filter.Name, *filter.TaskType, *filter.RangeID)
}

// LockTaskLists locks a row in task_lists table
func (mdb *DB) LockTaskLists(filter *sqldb.TaskListsFilter) (int64, error) {
	var rangeID int64
	err := mdb.conn.Get(&rangeID, mdb.driver.LockTaskListQuery(), filter.ShardID, *filter.DomainID, *filter.Name, *filter.TaskType)
	return rangeID, err
}
