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

package mysql

const (
    templateCreateWorkflowExecutionStarted = `INSERT IGNORE INTO executions_visibility (` +
        `domain_id, workflow_id, run_id, start_time, execution_time, workflow_type_name, memo, encoding) ` +
        `VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

    templateCreateWorkflowExecutionClosed = `REPLACE INTO executions_visibility (` +
        `domain_id, workflow_id, run_id, start_time, execution_time, workflow_type_name, close_time, close_status, history_length, memo, encoding) ` +
        `VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

    // RunID condition is needed for correct pagination
    templateConditions = ` AND domain_id = ?
		 AND start_time >= ?
		 AND start_time <= ?
 		 AND (run_id > ? OR start_time < ?)
         ORDER BY start_time DESC, run_id
         LIMIT ?`

    templateOpenFieldNames = `workflow_id, run_id, start_time, execution_time, workflow_type_name, memo, encoding`
    templateOpenSelect     = `SELECT ` + templateOpenFieldNames + ` FROM executions_visibility WHERE close_status IS NULL `

    templateClosedSelect = `SELECT ` + templateOpenFieldNames + `, close_time, close_status, history_length
		 FROM executions_visibility WHERE close_status IS NOT NULL `

    templateGetOpenWorkflowExecutions = templateOpenSelect + templateConditions

    templateGetClosedWorkflowExecutions = templateClosedSelect + templateConditions

    templateGetOpenWorkflowExecutionsByType = templateOpenSelect + `AND workflow_type_name = ?` + templateConditions

    templateGetClosedWorkflowExecutionsByType = templateClosedSelect + `AND workflow_type_name = ?` + templateConditions

    templateGetOpenWorkflowExecutionsByID = templateOpenSelect + `AND workflow_id = ?` + templateConditions

    templateGetClosedWorkflowExecutionsByID = templateClosedSelect + `AND workflow_id = ?` + templateConditions

    templateGetClosedWorkflowExecutionsByStatus = templateClosedSelect + `AND close_status = ?` + templateConditions

    templateGetClosedWorkflowExecution = `SELECT workflow_id, run_id, start_time, execution_time, memo, encoding, close_time, workflow_type_name, close_status, history_length 
		 FROM executions_visibility
		 WHERE domain_id = ? AND close_status IS NOT NULL
		 AND run_id = ?`

    templateDeleteWorkflowExecution = "DELETE FROM executions_visibility WHERE domain_id=? AND run_id=?"
)


func (d *driver) CreateWorkflowExecutionStartedQuery() string {
    return templateCreateWorkflowExecutionStarted
}
func (d *driver) CreateWorkflowExecutionClosedQuery() string {
    return templateCreateWorkflowExecutionClosed
}
func (d *driver) GetOpenWorkflowExecutionsQuery() string {
    return templateGetOpenWorkflowExecutions
}
func (d *driver) GetClosedWorkflowExecutionsQuery() string {
    return templateGetClosedWorkflowExecutions
}
func (d *driver) GetOpenWorkflowExecutionsByTypeQuery() string {
    return templateGetOpenWorkflowExecutionsByType
}
func (d *driver) GetClosedWorkflowExecutionsByTypeQuery() string {
    return templateGetClosedWorkflowExecutionsByType
}
func (d *driver) GetOpenWorkflowExecutionsByIDQuery() string {
    return templateGetOpenWorkflowExecutionsByID
}
func (d *driver) GetClosedWorkflowExecutionsByIDQuery() string {
    return templateGetClosedWorkflowExecutionsByID
}
func (d *driver) GetClosedWorkflowExecutionsByStatusQuery() string {
    return templateGetClosedWorkflowExecutionsByStatus
}
func (d *driver) GetClosedWorkflowExecutionQuery() string {
    return templateGetClosedWorkflowExecution
}
func (d *driver) DeleteWorkflowExecutionQuery() string {
    return templateDeleteWorkflowExecution
}