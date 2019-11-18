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
	"errors"
	"fmt"

	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

var errCloseParams = errors.New("missing one of {closeStatus, closeTime, historyLength} params")

// InsertIntoVisibility inserts a row into visibility table. If an row already exist,
// its left as such and no update will be made
func (mdb *DB) InsertIntoVisibility(row *sqldb.VisibilityRow) (sql.Result, error) {
	row.StartTime = mdb.converter.ToMySQLDateTime(row.StartTime)
	return mdb.conn.Exec(mdb.driver.CreateWorkflowExecutionStartedQuery(),
		row.DomainID,
		row.WorkflowID,
		row.RunID,
		row.StartTime,
		row.ExecutionTime,
		row.WorkflowTypeName,
		row.Memo,
		row.Encoding)
}

// ReplaceIntoVisibility replaces an existing row if it exist or creates a new row in visibility table
func (mdb *DB) ReplaceIntoVisibility(row *sqldb.VisibilityRow) (sql.Result, error) {
	switch {
	case row.CloseStatus != nil && row.CloseTime != nil && row.HistoryLength != nil:
		row.StartTime = mdb.converter.ToMySQLDateTime(row.StartTime)
		closeTime := mdb.converter.ToMySQLDateTime(*row.CloseTime)
		return mdb.conn.Exec(mdb.driver.CreateWorkflowExecutionClosedQuery(),
			row.DomainID,
			row.WorkflowID,
			row.RunID,
			row.StartTime,
			row.ExecutionTime,
			row.WorkflowTypeName,
			closeTime,
			*row.CloseStatus,
			*row.HistoryLength,
			row.Memo,
			row.Encoding)
	default:
		return nil, errCloseParams
	}
}

// DeleteFromVisibility deletes a row from visibility table if it exist
func (mdb *DB) DeleteFromVisibility(filter *sqldb.VisibilityFilter) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.DeleteWorkflowExecutionQuery(), filter.DomainID, filter.RunID)
}

// SelectFromVisibility reads one or more rows from visibility table
func (mdb *DB) SelectFromVisibility(filter *sqldb.VisibilityFilter) ([]sqldb.VisibilityRow, error) {
	var err error
	var rows []sqldb.VisibilityRow
	if filter.MinStartTime != nil {
		*filter.MinStartTime = mdb.converter.ToMySQLDateTime(*filter.MinStartTime)
	}
	if filter.MaxStartTime != nil {
		*filter.MaxStartTime = mdb.converter.ToMySQLDateTime(*filter.MaxStartTime)
	}
	switch {
	case filter.MinStartTime == nil && filter.RunID != nil && filter.Closed:
		var row sqldb.VisibilityRow
		err = mdb.conn.Get(&row, mdb.driver.GetClosedWorkflowExecutionQuery(), filter.DomainID, *filter.RunID)
		if err == nil {
			rows = append(rows, row)
		}
	case filter.MinStartTime != nil && filter.WorkflowID != nil:
		qry := mdb.driver.GetOpenWorkflowExecutionsByIDQuery()
		if filter.Closed {
			qry = mdb.driver.GetClosedWorkflowExecutionsByIDQuery()
		}
		err = mdb.conn.Select(&rows,
			qry,
			*filter.WorkflowID,
			filter.DomainID,
			mdb.converter.ToMySQLDateTime(*filter.MinStartTime),
			mdb.converter.ToMySQLDateTime(*filter.MaxStartTime),
			*filter.RunID,
			*filter.MinStartTime,
			*filter.PageSize)
	case filter.MinStartTime != nil && filter.WorkflowTypeName != nil:
		qry := mdb.driver.GetOpenWorkflowExecutionsByTypeQuery()
		if filter.Closed {
			qry = mdb.driver.GetClosedWorkflowExecutionsByTypeQuery()
		}
		err = mdb.conn.Select(&rows,
			qry,
			*filter.WorkflowTypeName,
			filter.DomainID,
			mdb.converter.ToMySQLDateTime(*filter.MinStartTime),
			mdb.converter.ToMySQLDateTime(*filter.MaxStartTime),
			*filter.RunID,
			*filter.MaxStartTime,
			*filter.PageSize)
	case filter.MinStartTime != nil && filter.CloseStatus != nil:
		err = mdb.conn.Select(&rows,
			mdb.driver.GetClosedWorkflowExecutionsByStatusQuery(),
			*filter.CloseStatus,
			filter.DomainID,
			mdb.converter.ToMySQLDateTime(*filter.MinStartTime),
			mdb.converter.ToMySQLDateTime(*filter.MaxStartTime),
			*filter.RunID,
			mdb.converter.ToMySQLDateTime(*filter.MaxStartTime),
			*filter.PageSize)
	case filter.MinStartTime != nil:
		qry := mdb.driver.GetOpenWorkflowExecutionsQuery()
		if filter.Closed {
			qry = mdb.driver.GetClosedWorkflowExecutionsQuery()
		}
		err = mdb.conn.Select(&rows,
			qry,
			filter.DomainID,
			mdb.converter.ToMySQLDateTime(*filter.MinStartTime),
			mdb.converter.ToMySQLDateTime(*filter.MaxStartTime),
			*filter.RunID,
			mdb.converter.ToMySQLDateTime(*filter.MaxStartTime),
			*filter.PageSize)
	default:
		return nil, fmt.Errorf("invalid query filter")
	}
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].StartTime = mdb.converter.FromMySQLDateTime(rows[i].StartTime)
		rows[i].ExecutionTime = mdb.converter.FromMySQLDateTime(rows[i].ExecutionTime)
		if rows[i].CloseTime != nil {
			closeTime := mdb.converter.FromMySQLDateTime(*rows[i].CloseTime)
			rows[i].CloseTime = &closeTime
		}
	}
	return rows, err
}
