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

package sqlshared

import (
	"github.com/jmoiron/sqlx"
	"github.com/uber/cadence/common/service/config"
)

type (
	// Driver is the driver interface that each SQL database needs to implement
	Driver interface {
		GetDriverName() string
		CreateDBConnection(cfg *config.SQL) (*sqlx.DB, error)
		IsDupEntryError(err error) bool

		//domain
		CreateDomainQuery() string
		UpdateDomainQuery() string
		GetDomainByIDQuery() string
		GetDomainByNameQuery() string
		ListDomainsQuery() string
		ListDomainsRangeQuery() string
		DeleteDomainByIDQuery() string
		DeleteDomainByNameQuery() string
		GetDomainMetadataQuery() string
		LockDomainMetadataQuery() string
		UpdateDomainMetadataQuery() string

		//events
		AddHistoryNodesQuery() string
		GetHistoryNodesQuery() string
		DeleteHistoryNodesQuery() string
		AddHistoryTreeQuery() string
		GetHistoryTreeQuery() string
		DeleteHistoryTreeQuery() string

		//execution
		CreateExecutionQuery() string
		UpdateExecutionQuery() string
		GetExecutionQuery() string
		DeleteExecutionQuery() string
		WriteLockExecutionQuery() string
		ReadLockExecutionQuery() string
		CreateCurrentExecutionQuery() string
		DeleteCurrentExecutionQuery() string
		GetCurrentExecutionQuery() string
		LockCurrentExecutionJoinExecutionsQuery() string
		LockCurrentExecutionQuery() string
		UpdateCurrentExecutionsQuery() string
		GetTransferTasksQuery() string
		CreateTransferTasksQuery() string
		DeleteTransferTaskQuery() string
		RangeDeleteTransferTaskQuery() string
		CreateTimerTasksQuery() string
		GetTimerTasksQuery() string
		DeleteTimerTaskQuery() string
		RangeDeleteTimerTaskQuery() string
		CreateReplicationTasksQuery() string
		GetReplicationTasksQuery() string
		DeleteReplicationTaskQuery() string
		GetReplicationTasksDLQQuery() string
		CreateBufferedEventsQuery() string
		DeleteBufferedEventsQuery() string
		GetBufferedEventsQuery() string
		InsertReplicationTaskDLQQuery() string

		//execution_map
		DeleteMapQueryTemplate() string
		SetKeyInMapQueryTemplate() string
		DeleteKeyInMapQueryTemplate() string
		GetMapQueryTemplate() string
		DeleteAllSignalsRequestedSetQuery() string
		CreateSignalsRequestedSetQuery() string
		DeleteSignalsRequestedSetQuery() string
		GetSignalsRequestedSetQuery() string

		//queue
		EnqueueMessageQuery() string
		GetLastMessageIDQuery() string
		GetMessagesQuery() string
		DeleteMessagesQuery() string
		GetQueueMetadataQuery() string
		GetQueueMetadataForUpdateQuery() string
		InsertQueueMetadataQuery() string
		UpdateQueueMetadataQuery() string

		//shard
		CreateShardQuery() string
		GetShardQuery() string
		UpdateShardQuery() string
		LockShardQuery() string
		ReadLockShardQuery() string

		//task
		CreateTaskListQuery() string
		ReplaceTaskListQuery() string
		UpdateTaskListQuery() string
		ListTaskListQuery() string
		GetTaskListQuery() string
		DeleteTaskListQuery() string
		LockTaskListQuery() string
		GetTaskMinMaxQuery() string
		GetTaskMinQuery() string
		CreateTaskQuery() string
		DeleteTaskQuery() string
		RangeDeleteTaskQuery() string

		//visibility
		CreateWorkflowExecutionStartedQuery() string
		CreateWorkflowExecutionClosedQuery() string
		GetOpenWorkflowExecutionsQuery() string
		GetClosedWorkflowExecutionsQuery() string
		GetOpenWorkflowExecutionsByTypeQuery() string
		GetClosedWorkflowExecutionsByTypeQuery() string
		GetOpenWorkflowExecutionsByIDQuery() string
		GetClosedWorkflowExecutionsByIDQuery() string
		GetClosedWorkflowExecutionsByStatusQuery() string
		GetClosedWorkflowExecutionQuery() string
		DeleteWorkflowExecutionQuery() string

		//admin: for CLI and testing
		ReadSchemaVersionQuery() string
		WriteSchemaVersionQuery() string
		WriteSchemaUpdateHistoryQuery() string
		CreateSchemaVersionTableQuery() string
		CreateSchemaUpdateHistoryTableQuery() string
		CreateDatabaseQuery() string
		DropDatabaseQuery() string
		ListTablesQuery() string
		DropTableQuery() string
	}
)
