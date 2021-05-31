// Copyright (c) 2019 Uber Technologies, Inc.
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

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	s "github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
)

const (
	templateCreateWorkflowExecutionStarted = `INSERT INTO executions_visibility (` +
		`domain_id, workflow_id, run_id, start_time, execution_time, workflow_type_name, memo, encoding, is_cron) ` +
		`VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
         ON CONFLICT (domain_id, run_id) DO NOTHING`

	templateCreateWorkflowExecutionClosed = `INSERT INTO executions_visibility (` +
		`domain_id, workflow_id, run_id, start_time, execution_time, workflow_type_name, close_time, close_status, history_length, memo, encoding, is_cron) ` +
		`VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (domain_id, run_id) DO UPDATE
		  SET workflow_id = excluded.workflow_id,
		      start_time = excluded.start_time,
		      execution_time = excluded.execution_time,
              workflow_type_name = excluded.workflow_type_name,
			  close_time = excluded.close_time,
			  close_status = excluded.close_status,
			  history_length = excluded.history_length,
			  memo = excluded.memo,
			  encoding = excluded.encoding,
				is_cron = excluded.is_cron`

	templateOpenFieldNames = `workflow_id, run_id, start_time, execution_time, workflow_type_name, memo, encoding, is_cron`
	templateOpenSelect     = `SELECT ` + templateOpenFieldNames + ` FROM executions_visibility WHERE close_status IS NULL `

	templateClosedSelect = `SELECT ` + templateOpenFieldNames + `, close_time, close_status, history_length
		 FROM executions_visibility WHERE close_status IS NOT NULL `

	templateDeleteWorkflowExecution = "DELETE FROM executions_visibility WHERE domain_id=$1 AND run_id=$2"
)

var errCloseParams = errors.New("missing one of {closeStatus, closeTime, historyLength} params")

// buildTemplate creates the Postgres query based on given
func (pdb *db) buildTemplate(filter *sqlplugin.VisibilityFilter, selectType s.SelectType) string {
	numParams := 0
	dynamicFilter := ""
	if filter.IsCron != nil {
		dynamicFilter = ` AND is_cron = $%d`
		numParams++
	}

	numParams += 6
	// RunID condition is needed for correct pagination
	templateConditions := ` AND domain_id = $%d
		 AND start_time >= $%d
		 AND start_time <= $%d
 		 AND (run_id > $%d OR start_time < $%d)
         ORDER BY start_time DESC, run_id
         LIMIT $%d`

	// This function will help creating Postgres template with params like $1 $2 $3 ...
	// with correct number of parameters
	rangeSlice := func(numElems int) []interface{} {
		out := make([]interface{}, numElems)
		for i := 0; i < numElems; i++ {
			out[i] = i + 1
		}
		return out
	}

	templateStr := func(prefix string, numParams int) string {
		return fmt.Sprintf(dynamicFilter+prefix+templateConditions, rangeSlice(numParams)...)
	}

	switch selectType {
	case s.OpenWorkflowExecutions:
		return templateOpenSelect + templateStr("", numParams)
	case s.ClosedWorkflowExecutions:
		return templateClosedSelect + templateStr("", numParams)
	case s.OpenWorkflowExecutionsByType:
		return templateOpenSelect + templateStr(` AND workflow_type_name = $%d`, numParams+1)
	case s.ClosedWorkflowExecutionsByType:
		return templateClosedSelect + templateStr(` AND workflow_type_name = $%d`, numParams+1)
	case s.OpenWorkflowExecutionsByID:
		return templateOpenSelect + templateStr(` AND workflow_id = $%d`, numParams+1)
	case s.ClosedWorkflowExecutionsByID:
		return templateClosedSelect + templateStr(` AND workflow_id = $%d`, numParams+1)
	case s.ClosedWorkflowExecutionsByStatus:
		return templateClosedSelect + templateStr(` AND close_status = $%d`, numParams+1)
	case s.ClosedWorkflowExecution:
		return `SELECT workflow_id, run_id, start_time, execution_time, memo, encoding, close_time, workflow_type_name, close_status, history_length, is_cron
		FROM executions_visibility
		WHERE domain_id = $1 AND close_status IS NOT NULL
		AND run_id = $2`
	}
	return ""
}

func (pdb *db) SelectContext(
	ctx context.Context,
	filter *sqlplugin.VisibilityFilter,
	dest interface{},
	query string,
	args ...interface{}) error {
	return pdb.conn.SelectContext(ctx, dest, query, s.DynamicSelectArgs(filter, args)...)
}

// InsertIntoVisibility inserts a row into visibility table. If an row already exist,
// its left as such and no update will be made
func (pdb *db) InsertIntoVisibility(ctx context.Context, row *sqlplugin.VisibilityRow) (sql.Result, error) {
	row.StartTime = pdb.converter.ToPostgresDateTime(row.StartTime)
	return pdb.conn.ExecContext(ctx, templateCreateWorkflowExecutionStarted,
		row.DomainID,
		row.WorkflowID,
		row.RunID,
		row.StartTime,
		row.ExecutionTime,
		row.WorkflowTypeName,
		row.Memo,
		row.Encoding,
		row.IsCron)
}

// ReplaceIntoVisibility replaces an existing row if it exist or creates a new row in visibility table
func (pdb *db) ReplaceIntoVisibility(ctx context.Context, row *sqlplugin.VisibilityRow) (sql.Result, error) {
	switch {
	case row.CloseStatus != nil && row.CloseTime != nil && row.HistoryLength != nil:
		row.StartTime = pdb.converter.ToPostgresDateTime(row.StartTime)
		closeTime := pdb.converter.ToPostgresDateTime(*row.CloseTime)
		return pdb.conn.ExecContext(ctx, templateCreateWorkflowExecutionClosed,
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
			row.Encoding,
			row.IsCron)
	default:
		return nil, errCloseParams
	}
}

// DeleteFromVisibility deletes a row from visibility table if it exist
func (pdb *db) DeleteFromVisibility(ctx context.Context, filter *sqlplugin.VisibilityFilter) (sql.Result, error) {
	return pdb.conn.ExecContext(ctx, templateDeleteWorkflowExecution, filter.DomainID, filter.RunID)
}

// SelectFromVisibility reads one or more rows from visibility table
func (pdb *db) SelectFromVisibility(ctx context.Context, filter *sqlplugin.VisibilityFilter) ([]sqlplugin.VisibilityRow, error) {
	var err error
	var rows []sqlplugin.VisibilityRow
	if filter.MinStartTime != nil {
		*filter.MinStartTime = pdb.converter.ToPostgresDateTime(*filter.MinStartTime)
	}
	if filter.MaxStartTime != nil {
		*filter.MaxStartTime = pdb.converter.ToPostgresDateTime(*filter.MaxStartTime)
	}
	switch {
	case filter.MinStartTime == nil && filter.RunID != nil && filter.Closed:
		var row sqlplugin.VisibilityRow
		err = pdb.conn.GetContext(ctx, &row, pdb.buildTemplate(filter, s.ClosedWorkflowExecution), filter.DomainID, *filter.RunID)
		if err == nil {
			rows = append(rows, row)
		}
	case filter.MinStartTime != nil && filter.WorkflowID != nil:
		qry := pdb.buildTemplate(filter, s.OpenWorkflowExecutionsByID)
		if filter.Closed {
			qry = pdb.buildTemplate(filter, s.ClosedWorkflowExecutionsByID)
		}
		err = pdb.SelectContext(ctx,
			filter,
			&rows,
			qry,
			*filter.WorkflowID,
			filter.DomainID,
			pdb.converter.ToPostgresDateTime(*filter.MinStartTime),
			pdb.converter.ToPostgresDateTime(*filter.MaxStartTime),
			*filter.RunID,
			*filter.MinStartTime,
			*filter.PageSize)
	case filter.MinStartTime != nil && filter.WorkflowTypeName != nil:
		qry := pdb.buildTemplate(filter, s.OpenWorkflowExecutionsByType)
		if filter.Closed {
			qry = pdb.buildTemplate(filter, s.ClosedWorkflowExecutionsByType)
		}
		err = pdb.SelectContext(ctx,
			filter,
			&rows,
			qry,
			*filter.WorkflowTypeName,
			filter.DomainID,
			pdb.converter.ToPostgresDateTime(*filter.MinStartTime),
			pdb.converter.ToPostgresDateTime(*filter.MaxStartTime),
			*filter.RunID,
			*filter.MaxStartTime,
			*filter.PageSize)
	case filter.MinStartTime != nil && filter.CloseStatus != nil:
		err = pdb.SelectContext(ctx,
			filter,
			&rows,
			pdb.buildTemplate(filter, s.ClosedWorkflowExecutionsByStatus),
			*filter.CloseStatus,
			filter.DomainID,
			pdb.converter.ToPostgresDateTime(*filter.MinStartTime),
			pdb.converter.ToPostgresDateTime(*filter.MaxStartTime),
			*filter.RunID,
			pdb.converter.ToPostgresDateTime(*filter.MaxStartTime),
			*filter.PageSize)
	case filter.MinStartTime != nil:
		qry := pdb.buildTemplate(filter, s.OpenWorkflowExecutions)
		if filter.Closed {
			qry = pdb.buildTemplate(filter, s.ClosedWorkflowExecutions)
		}
		minSt := pdb.converter.ToPostgresDateTime(*filter.MinStartTime)
		maxSt := pdb.converter.ToPostgresDateTime(*filter.MaxStartTime)
		err = pdb.SelectContext(ctx,
			filter,
			&rows,
			qry,
			filter.DomainID,
			minSt,
			maxSt,
			*filter.RunID,
			maxSt,
			*filter.PageSize)
	default:
		return nil, fmt.Errorf("invalid query filter")
	}
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].StartTime = pdb.converter.FromPostgresDateTime(rows[i].StartTime)
		rows[i].ExecutionTime = pdb.converter.FromPostgresDateTime(rows[i].ExecutionTime)
		if rows[i].CloseTime != nil {
			closeTime := pdb.converter.FromPostgresDateTime(*rows[i].CloseTime)
			rows[i].CloseTime = &closeTime
		}
		rows[i].RunID = strings.TrimSpace(rows[i].RunID)
		rows[i].WorkflowID = strings.TrimSpace(rows[i].WorkflowID)
	}
	return rows, err
}
