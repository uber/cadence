// Copyright (c) 2020 Uber Technologies, Inc.
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

package cassandra

import (
	"context"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
)

const (
	domainPartition        = 0
)

const (
	///////////////// Open Executions /////////////////
	openExecutionsColumnsForSelect = " workflow_id, run_id, start_time, execution_time, workflow_type_name, memo, encoding, task_list, is_cron "

	openExecutionsColumnsForInsert = "(domain_id, domain_partition, " + openExecutionsColumnsForSelect + ")"

	templateCreateWorkflowExecutionStartedWithTTL = `INSERT INTO open_executions ` +
		openExecutionsColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) using TTL ?`

	templateCreateWorkflowExecutionStarted = `INSERT INTO open_executions` +
		openExecutionsColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	templateDeleteWorkflowExecutionStarted = `DELETE FROM open_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time = ? ` +
		`AND run_id = ?`

	templateGetOpenWorkflowExecutions = `SELECT ` + openExecutionsColumnsForSelect +
		`FROM open_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition IN (?) ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? `

	templateGetOpenWorkflowExecutionsByType = `SELECT ` + openExecutionsColumnsForSelect +
		`FROM open_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND workflow_type_name = ? `

	templateGetOpenWorkflowExecutionsByID = `SELECT ` + openExecutionsColumnsForSelect +
		`FROM open_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND workflow_id = ? `

	///////////////// Closed Executions /////////////////
	closedExecutionColumnsForSelect = " workflow_id, run_id, start_time, execution_time, close_time, workflow_type_name, status, history_length, memo, encoding, task_list, is_cron "

	closedExecutionColumnsForInsert = "(domain_id, domain_partition, " + closedExecutionColumnsForSelect + ")"

	templateCreateWorkflowExecutionClosedWithTTL = `INSERT INTO closed_executions ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) using TTL ?`

	templateCreateWorkflowExecutionClosed = `INSERT INTO closed_executions ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	templateCreateWorkflowExecutionClosedWithTTLV2 = `INSERT INTO closed_executions_v2 ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) using TTL ?`

	templateCreateWorkflowExecutionClosedV2 = `INSERT INTO closed_executions_v2 ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	templateGetClosedWorkflowExecutions = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition IN (?) ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? `

	templateGetClosedWorkflowExecutionsByType = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND workflow_type_name = ? `

	templateGetClosedWorkflowExecutionsByID = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND workflow_id = ? `

	templateGetClosedWorkflowExecutionsByStatus = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND status = ? `

	templateGetClosedWorkflowExecution = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND workflow_id = ? ` +
		`AND run_id = ? ALLOW FILTERING `

	templateGetClosedWorkflowExecutionsV2 = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition IN (?) ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? `

	templateGetClosedWorkflowExecutionsByTypeV2 = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? ` +
		`AND workflow_type_name = ? `

	templateGetClosedWorkflowExecutionsByIDV2 = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? ` +
		`AND workflow_id = ? `

	templateGetClosedWorkflowExecutionsByStatusV2 = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? ` +
		`AND status = ? `
)

// InsertShard creates a new shard, return error is there is any.
// When error is nil, return applied=true if there is a conflict, and return the conflicted row as previous
func (db *cdb) InsertSha2rd(ctx context.Context, row *persistence.InternalShardInfo) (error, *nosqlplugin.ConflictedShardRow) {
}