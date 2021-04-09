// Copyright (c) 2020 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"fmt"
	"strings"
	"time"

	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/service/config"
	"github.com/uber/cadence/common/types"
)

type (
	// Implements TaskManager
	cassandraTaskPersistence struct {
		cassandraStore
		shardID            int
		currentClusterName string
	}
)

const (
	// Row types for table tasks
	rowTypeTask = iota
	rowTypeTaskList
)

const (
	taskListTaskID = -12345
	initialRangeID = 1 // Id of the first range of a new task list
)

const (
	templateTaskListType = `{` +
		`domain_id: ?, ` +
		`name: ?, ` +
		`type: ?, ` +
		`ack_level: ?, ` +
		`kind: ?, ` +
		`last_updated: ? ` +
		`}`

	templateTaskType = `{` +
		`domain_id: ?, ` +
		`workflow_id: ?, ` +
		`run_id: ?, ` +
		`schedule_id: ?,` +
		`created_time: ? ` +
		`}`

	templateCreateTaskQuery = `INSERT INTO tasks (` +
		`domain_id, task_list_name, task_list_type, type, task_id, task) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateTaskType + `)`

	templateCreateTaskWithTTLQuery = `INSERT INTO tasks (` +
		`domain_id, task_list_name, task_list_type, type, task_id, task) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateTaskType + `) USING TTL ?`

	templateGetTasksQuery = `SELECT task_id, task ` +
		`FROM tasks ` +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id > ? ` +
		`and task_id <= ?`

	templateCompleteTaskQuery = `DELETE FROM tasks ` +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id = ?`

	templateCompleteTasksLessThanQuery = `DELETE FROM tasks ` +
		`WHERE domain_id = ? ` +
		`AND task_list_name = ? ` +
		`AND task_list_type = ? ` +
		`AND type = ? ` +
		`AND task_id <= ? `

	templateGetTaskList = `SELECT ` +
		`range_id, ` +
		`task_list ` +
		`FROM tasks ` +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id = ?`

	templateInsertTaskListQuery = `INSERT INTO tasks (` +
		`domain_id, ` +
		`task_list_name, ` +
		`task_list_type, ` +
		`type, ` +
		`task_id, ` +
		`range_id, ` +
		`task_list ` +
		`) VALUES (?, ?, ?, ?, ?, ?, ` + templateTaskListType + `) IF NOT EXISTS`

	templateUpdateTaskListQuery = `UPDATE tasks SET ` +
		`range_id = ?, ` +
		`task_list = ` + templateTaskListType + " " +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id = ? ` +
		`IF range_id = ?`

	templateUpdateTaskListQueryWithTTLPart1 = ` INSERT INTO tasks (` +
		`domain_id, ` +
		`task_list_name, ` +
		`task_list_type, ` +
		`type, ` +
		`task_id ` +
		`) VALUES (?, ?, ?, ?, ?) USING TTL ?`

	templateUpdateTaskListQueryWithTTLPart2 = `UPDATE tasks USING TTL ? SET ` +
		`range_id = ?, ` +
		`task_list = ` + templateTaskListType + " " +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id = ? ` +
		`IF range_id = ?`

	templateDeleteTaskListQuery = `DELETE FROM tasks ` +
		`WHERE domain_id = ? ` +
		`AND task_list_name = ? ` +
		`AND task_list_type = ? ` +
		`AND type = ? ` +
		`AND task_id = ? ` +
		`IF range_id = ?`
)

var _ p.TaskStore = (*cassandraTaskPersistence)(nil)

// newTaskPersistence is used to create an instance of TaskManager implementation
func newTaskPersistence(
	cfg config.Cassandra,
	logger log.Logger,
) (p.TaskStore, error) {
	session, err := cassandra.CreateSession(cfg)
	if err != nil {
		return nil, err
	}

	return &cassandraTaskPersistence{
		cassandraStore: cassandraStore{
			client:  cfg.CQLClient,
			session: session,
			logger:  logger,
		},
		shardID: -1,
	}, nil
}

// From TaskManager interface
func (d *cassandraTaskPersistence) LeaseTaskList(
	ctx context.Context,
	request *p.LeaseTaskListRequest,
) (*p.LeaseTaskListResponse, error) {
	if len(request.TaskList) == 0 {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("LeaseTaskList requires non empty task list"),
		}
	}
	now := time.Now()
	query := d.session.Query(templateGetTaskList,
		request.DomainID,
		request.TaskList,
		request.TaskType,
		rowTypeTaskList,
		taskListTaskID,
	).WithContext(ctx)
	var rangeID, ackLevel int64
	var tlDB map[string]interface{}
	err := query.Scan(&rangeID, &tlDB)
	if err != nil {
		if d.client.IsNotFoundError(err) { // First time task list is used
			query = d.session.Query(templateInsertTaskListQuery,
				request.DomainID,
				request.TaskList,
				request.TaskType,
				rowTypeTaskList,
				taskListTaskID,
				initialRangeID,
				request.DomainID,
				request.TaskList,
				request.TaskType,
				0,
				request.TaskListKind,
				now,
			).WithContext(ctx)
		} else {
			return nil, convertCommonErrors(d.client, "LeaseTaskList", err)
		}
	} else {
		// if request.RangeID is > 0, we are trying to renew an already existing
		// lease on the task list. If request.RangeID=0, we are trying to steal
		// the tasklist from its current owner
		if request.RangeID > 0 && request.RangeID != rangeID {
			return nil, &p.ConditionFailedError{
				Msg: fmt.Sprintf("leaseTaskList:renew failed: taskList:%v, taskListType:%v, haveRangeID:%v, gotRangeID:%v",
					request.TaskList, request.TaskType, request.RangeID, rangeID),
			}
		}
		ackLevel = tlDB["ack_level"].(int64)
		taskListKind := tlDB["kind"].(int)
		query = d.session.Query(templateUpdateTaskListQuery,
			rangeID+1,
			request.DomainID,
			&request.TaskList,
			request.TaskType,
			ackLevel,
			taskListKind,
			now,
			request.DomainID,
			&request.TaskList,
			request.TaskType,
			rowTypeTaskList,
			taskListTaskID,
			rangeID,
		).WithContext(ctx)
	}
	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return nil, convertCommonErrors(d.client, "LeaseTaskList", err)
	}
	if !applied {
		previousRangeID := previous["range_id"]
		return nil, &p.ConditionFailedError{
			Msg: fmt.Sprintf("leaseTaskList: taskList:%v, taskListType:%v, haveRangeID:%v, gotRangeID:%v",
				request.TaskList, request.TaskType, rangeID, previousRangeID),
		}
	}
	tli := &p.TaskListInfo{
		DomainID:    request.DomainID,
		Name:        request.TaskList,
		TaskType:    request.TaskType,
		RangeID:     rangeID + 1,
		AckLevel:    ackLevel,
		Kind:        request.TaskListKind,
		LastUpdated: now,
	}
	return &p.LeaseTaskListResponse{TaskListInfo: tli}, nil
}

// From TaskManager interface
func (d *cassandraTaskPersistence) UpdateTaskList(
	ctx context.Context,
	request *p.UpdateTaskListRequest,
) (*p.UpdateTaskListResponse, error) {
	tli := request.TaskListInfo

	var applied bool
	var err error
	previous := make(map[string]interface{})
	if tli.Kind == p.TaskListKindSticky { // if task_list is sticky, then update with TTL
		batch := d.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
		// part 1 is used to set TTL on primary key as UPDATE can't set TTL for primary key
		batch.Query(templateUpdateTaskListQueryWithTTLPart1,
			tli.DomainID,
			&tli.Name,
			tli.TaskType,
			rowTypeTaskList,
			taskListTaskID,
			stickyTaskListTTL,
		)
		// part 2 is for CAS and setting TTL for the rest of the columns
		batch.Query(templateUpdateTaskListQueryWithTTLPart2,
			stickyTaskListTTL,
			tli.RangeID,
			tli.DomainID,
			&tli.Name,
			tli.TaskType,
			tli.AckLevel,
			tli.Kind,
			time.Now(),
			tli.DomainID,
			&tli.Name,
			tli.TaskType,
			rowTypeTaskList,
			taskListTaskID,
			tli.RangeID,
		)
		applied, _, err = d.session.MapExecuteBatchCAS(batch, previous)
	} else {
		query := d.session.Query(templateUpdateTaskListQuery,
			tli.RangeID,
			tli.DomainID,
			&tli.Name,
			tli.TaskType,
			tli.AckLevel,
			tli.Kind,
			time.Now(),
			tli.DomainID,
			&tli.Name,
			tli.TaskType,
			rowTypeTaskList,
			taskListTaskID,
			tli.RangeID,
		).WithContext(ctx)
		applied, err = query.MapScanCAS(previous)
	}

	if err != nil {
		return nil, convertCommonErrors(d.client, "UpdateTaskList", err)
	}

	if !applied {
		var columns []string
		for k, v := range previous {
			columns = append(columns, fmt.Sprintf("%s=%v", k, v))
		}

		return nil, &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to update task list. name: %v, type: %v, rangeID: %v, columns: (%v)",
				tli.Name, tli.TaskType, tli.RangeID, strings.Join(columns, ",")),
		}
	}

	return &p.UpdateTaskListResponse{}, nil
}

func (d *cassandraTaskPersistence) ListTaskList(
	ctx context.Context,
	request *p.ListTaskListRequest,
) (*p.ListTaskListResponse, error) {
	return nil, &types.InternalServiceError{
		Message: fmt.Sprintf("unsupported operation"),
	}
}

func (d *cassandraTaskPersistence) DeleteTaskList(
	ctx context.Context,
	request *p.DeleteTaskListRequest,
) error {
	query := d.session.Query(templateDeleteTaskListQuery,
		request.DomainID,
		request.TaskListName,
		request.TaskListType,
		rowTypeTaskList,
		taskListTaskID,
		request.RangeID,
	).WithContext(ctx)
	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return convertCommonErrors(d.client, "DeleteTaskList", err)
	}
	if !applied {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("DeleteTaskList operation failed: expected_range_id=%v but found %+v", request.RangeID, previous),
		}
	}
	return nil
}

// From TaskManager interface
func (d *cassandraTaskPersistence) CreateTasks(
	ctx context.Context,
	request *p.InternalCreateTasksRequest,
) (*p.CreateTasksResponse, error) {
	batch := d.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	domainID := request.TaskListInfo.DomainID
	taskList := request.TaskListInfo.Name
	taskListType := request.TaskListInfo.TaskType
	taskListKind := request.TaskListInfo.Kind
	ackLevel := request.TaskListInfo.AckLevel
	cqlNowTimestamp := p.UnixNanoToDBTimestamp(time.Now().UnixNano())

	for _, task := range request.Tasks {
		scheduleID := task.Data.ScheduleID
		ttl := int64(task.Data.ScheduleToStartTimeout.Seconds())
		if ttl <= 0 {
			batch.Query(templateCreateTaskQuery,
				domainID,
				taskList,
				taskListType,
				rowTypeTask,
				task.TaskID,
				domainID,
				task.Execution.GetWorkflowID(),
				task.Execution.GetRunID(),
				scheduleID,
				cqlNowTimestamp)
		} else {
			if ttl > maxCassandraTTL {
				ttl = maxCassandraTTL
			}
			batch.Query(templateCreateTaskWithTTLQuery,
				domainID,
				taskList,
				taskListType,
				rowTypeTask,
				task.TaskID,
				domainID,
				task.Execution.GetWorkflowID(),
				task.Execution.GetRunID(),
				scheduleID,
				cqlNowTimestamp,
				ttl)
		}
	}

	// The following query is used to ensure that range_id didn't change
	batch.Query(templateUpdateTaskListQuery,
		request.TaskListInfo.RangeID,
		domainID,
		taskList,
		taskListType,
		ackLevel,
		taskListKind,
		time.Now(),
		domainID,
		taskList,
		taskListType,
		rowTypeTaskList,
		taskListTaskID,
		request.TaskListInfo.RangeID,
	)

	previous := make(map[string]interface{})
	applied, _, err := d.session.MapExecuteBatchCAS(batch, previous)
	if err != nil {
		return nil, convertCommonErrors(d.client, "CreateTasks", err)
	}
	if !applied {
		rangeID := previous["range_id"]
		return nil, &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to create task. TaskList: %v, taskListType: %v, rangeID: %v, db rangeID: %v",
				taskList, taskListType, request.TaskListInfo.RangeID, rangeID),
		}
	}

	return &p.CreateTasksResponse{}, nil
}

// From TaskManager interface
func (d *cassandraTaskPersistence) GetTasks(
	ctx context.Context,
	request *p.GetTasksRequest,
) (*p.InternalGetTasksResponse, error) {
	if request.MaxReadLevel == nil {
		return nil, &types.InternalServiceError{
			Message: "getTasks: both readLevel and maxReadLevel MUST be specified for cassandra persistence",
		}
	}
	if request.ReadLevel > *request.MaxReadLevel {
		return &p.InternalGetTasksResponse{}, nil
	}

	// Reading tasklist tasks need to be quorum level consistent, otherwise we could loose task
	query := d.session.Query(templateGetTasksQuery,
		request.DomainID,
		request.TaskList,
		request.TaskType,
		rowTypeTask,
		request.ReadLevel,
		*request.MaxReadLevel,
	).PageSize(request.BatchSize).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, &types.InternalServiceError{
			Message: "GetTasks operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalGetTasksResponse{}
	task := make(map[string]interface{})
PopulateTasks:
	for iter.MapScan(task) {
		taskID, ok := task["task_id"]
		if !ok { // no tasks, but static column record returned
			continue
		}
		t := createTaskInfo(task["task"].(map[string]interface{}))
		t.TaskID = taskID.(int64)
		response.Tasks = append(response.Tasks, t)
		if len(response.Tasks) == request.BatchSize {
			break PopulateTasks
		}
		task = make(map[string]interface{}) // Reinitialize map as initialized fails on unmarshalling
	}

	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(d.client, "GetTasks", err)
	}

	return response, nil
}

// From TaskManager interface
func (d *cassandraTaskPersistence) CompleteTask(
	ctx context.Context,
	request *p.CompleteTaskRequest,
) error {
	tli := request.TaskList
	query := d.session.Query(templateCompleteTaskQuery,
		tli.DomainID,
		tli.Name,
		tli.TaskType,
		rowTypeTask,
		request.TaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "CompleteTask", err)
	}

	return nil
}

// CompleteTasksLessThan deletes all tasks less than or equal to the given task id. This API ignores the
// Limit request parameter i.e. either all tasks leq the task_id will be deleted or an error will
// be returned to the caller
func (d *cassandraTaskPersistence) CompleteTasksLessThan(
	ctx context.Context,
	request *p.CompleteTasksLessThanRequest,
) (int, error) {
	query := d.session.Query(templateCompleteTasksLessThanQuery,
		request.DomainID,
		request.TaskListName,
		request.TaskType,
		rowTypeTask,
		request.TaskID,
	).WithContext(ctx)
	err := query.Exec()
	if err != nil {
		return 0, convertCommonErrors(d.client, "CompleteTasksLessThan", err)
	}
	return p.UnknownNumRowsAffected, nil
}
