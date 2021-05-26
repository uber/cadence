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
	"math"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/types"
)

type (
	// Implements TaskManager
	nosqlTaskManager struct {
		db     nosqlplugin.DB
		logger log.Logger
	}
)

const (
	initialRangeID  = 1 // Id of the first range of a new task list
	initialAckLevel = 0
)

var _ p.TaskStore = (*nosqlTaskManager)(nil)

// newTaskPersistence is used to create an instance of TaskManager implementation
func newTaskPersistence(
	cfg config.Cassandra,
	logger log.Logger,
) (p.TaskStore, error) {
	// TODO hardcoding to Cassandra for now, will switch to dynamically loading later
	db, err := cassandra.NewCassandraDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &nosqlTaskManager{
		db:     db,
		logger: logger,
	}, nil
}

func (t *nosqlTaskManager) GetName() string {
	return t.db.PluginName()
}

// Close releases the underlying resources held by this object
func (t *nosqlTaskManager) Close() {
	t.db.Close()
}

func (t *nosqlTaskManager) GetOrphanTasks(ctx context.Context, request *p.GetOrphanTasksRequest) (*p.GetOrphanTasksResponse, error) {
	// TODO: It's unclear if this's necessary or possible for NoSQL
	return nil, &types.InternalServiceError{
		Message: "Unimplemented call to GetOrphanTasks for NoSQL",
	}
}

// From TaskManager interface
func (t *nosqlTaskManager) LeaseTaskList(
	ctx context.Context,
	request *p.LeaseTaskListRequest,
) (*p.LeaseTaskListResponse, error) {
	if len(request.TaskList) == 0 {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("LeaseTaskList requires non empty task list"),
		}
	}
	now := time.Now()
	currTL, err := t.db.SelectTaskList(ctx, &nosqlplugin.TaskListFilter{
		DomainID:     request.DomainID,
		TaskListName: request.TaskList,
		TaskListType: request.TaskType,
	})

	var previous *nosqlplugin.TaskListRow
	if err != nil {
		if t.db.IsNotFoundError(err) { // First time task list is used
			currTL = &nosqlplugin.TaskListRow{
				DomainID:        request.DomainID,
				TaskListName:    request.TaskList,
				TaskListType:    request.TaskType,
				RangeID:         initialRangeID,
				TaskListKind:    request.TaskListKind,
				AckLevel:        initialAckLevel,
				LastUpdatedTime: now,
			}
			previous, err = t.db.InsertTaskList(ctx, currTL)
		} else {
			return nil, convertCommonErrors(t.db, "LeaseTaskList", err)
		}
	} else {
		// if request.RangeID is > 0, we are trying to renew an already existing
		// lease on the task list. If request.RangeID=0, we are trying to steal
		// the currTL from its current owner
		if request.RangeID > 0 && request.RangeID != currTL.RangeID {
			return nil, &p.ConditionFailedError{
				Msg: fmt.Sprintf("leaseTaskList:renew failed: taskList:%v, taskListType:%v, haveRangeID:%v, gotRangeID:%v",
					request.TaskList, request.TaskType, request.RangeID, currTL.RangeID),
			}
		}

		// Update the rangeID as this is an ownership change
		currTL.RangeID += 1

		previous, err = t.db.UpdateTaskList(ctx, &nosqlplugin.TaskListRow{
			DomainID:        request.DomainID,
			TaskListName:    request.TaskList,
			TaskListType:    request.TaskType,
			RangeID:         currTL.RangeID,
			TaskListKind:    currTL.TaskListKind,
			AckLevel:        currTL.AckLevel,
			LastUpdatedTime: now,
		}, currTL.RangeID)
	}
	if err != nil {
		if t.db.IsConditionFailedError(err) {
			return nil, &p.ConditionFailedError{
				Msg: fmt.Sprintf("leaseTaskList: taskList:%v, taskListType:%v, haveRangeID:%v, gotRangeID:%v",
					request.TaskList, request.TaskType, currTL.RangeID, previous.RangeID),
			}
		}
		return nil, convertCommonErrors(t.db, "LeaseTaskList", err)
	}
	tli := &p.TaskListInfo{
		DomainID:    request.DomainID,
		Name:        request.TaskList,
		TaskType:    request.TaskType,
		RangeID:     currTL.RangeID,
		AckLevel:    currTL.AckLevel,
		Kind:        request.TaskListKind,
		LastUpdated: now,
	}
	return &p.LeaseTaskListResponse{TaskListInfo: tli}, nil
}

// From TaskManager interface
func (t *nosqlTaskManager) UpdateTaskList(
	ctx context.Context,
	request *p.UpdateTaskListRequest,
) (*p.UpdateTaskListResponse, error) {
	tli := request.TaskListInfo
	var err error
	var previous *nosqlplugin.TaskListRow
	taskListToUpdate := &nosqlplugin.TaskListRow{
		DomainID:        tli.DomainID,
		TaskListName:    tli.Name,
		TaskListType:    tli.TaskType,
		RangeID:         tli.RangeID,
		TaskListKind:    tli.Kind,
		AckLevel:        tli.AckLevel,
		LastUpdatedTime: time.Now(),
	}

	if tli.Kind == p.TaskListKindSticky { // if task_list is sticky, then update with TTL
		previous, err = t.db.UpdateTaskListWithTTL(ctx, stickyTaskListTTL, taskListToUpdate, tli.RangeID)
	} else {
		previous, err = t.db.UpdateTaskList(ctx, taskListToUpdate, tli.RangeID)
	}

	if err != nil {
		if t.db.IsConditionFailedError(err) {
			return nil, &p.ConditionFailedError{
				Msg: fmt.Sprintf("Failed to update task list. name: %v, type: %v, rangeID: %v, columns: (%v)",
					tli.Name, tli.TaskType, tli.RangeID, previous),
			}
		}
		return nil, convertCommonErrors(t.db, "UpdateTaskList", err)
	}

	return &p.UpdateTaskListResponse{}, nil
}

func (t *nosqlTaskManager) ListTaskList(
	_ context.Context,
	_ *p.ListTaskListRequest,
) (*p.ListTaskListResponse, error) {
	return nil, &types.InternalServiceError{
		Message: fmt.Sprintf("unsupported operation"),
	}
}

func (t *nosqlTaskManager) DeleteTaskList(
	ctx context.Context,
	request *p.DeleteTaskListRequest,
) error {
	previous, err := t.db.DeleteTaskList(ctx, &nosqlplugin.TaskListFilter{
		DomainID:     request.DomainID,
		TaskListName: request.TaskListName,
		TaskListType: request.TaskListType,
	}, request.RangeID)

	if err != nil {
		if t.db.IsConditionFailedError(err) {
			return &p.ConditionFailedError{
				Msg: fmt.Sprintf("Failed to delete task list. name: %v, type: %v, rangeID: %v, columns: (%v)",
					request.TaskListName, request.TaskListType, request.RangeID, previous),
			}
		}
		return convertCommonErrors(t.db, "DeleteTaskList", err)
	}

	return nil
}

// From TaskManager interface
func (t *nosqlTaskManager) CreateTasks(
	ctx context.Context,
	request *p.InternalCreateTasksRequest,
) (*p.CreateTasksResponse, error) {
	now := time.Now()
	var tasks []*nosqlplugin.TaskRowForInsert
	for _, t := range request.Tasks {
		task := &nosqlplugin.TaskRow{
			DomainID:     request.TaskListInfo.DomainID,
			TaskListName: request.TaskListInfo.Name,
			TaskListType: request.TaskListInfo.TaskType,
			TaskID:       t.TaskID,
			WorkflowID:   t.Execution.GetWorkflowID(),
			RunID:        t.Execution.GetRunID(),
			ScheduledID:  t.Data.ScheduleID,
			CreatedTime:  now,
		}
		ttl := int(t.Data.ScheduleToStartTimeout.Seconds())
		tasks = append(tasks, &nosqlplugin.TaskRowForInsert{
			TaskRow:    *task,
			TTLSeconds: ttl,
		})
	}

	previous, err := t.db.InsertTasks(ctx, tasks, toTaskListRow(request.TaskListInfo))

	if err != nil {
		if t.db.IsConditionFailedError(err) {
			return nil, &p.ConditionFailedError{
				Msg: fmt.Sprintf("Failed to insert tasks. name: %v, type: %v, rangeID: %v, columns: (%v)",
					request.TaskListInfo.Name, request.TaskListInfo.TaskType, request.TaskListInfo.RangeID, previous),
			}
		}
		return nil, convertCommonErrors(t.db, "CreateTasks", err)
	}

	return &p.CreateTasksResponse{}, nil
}

func toTaskListRow(info *p.TaskListInfo) *nosqlplugin.TaskListRow {
	return &nosqlplugin.TaskListRow{
		DomainID:        info.DomainID,
		TaskListName:    info.Name,
		TaskListType:    info.TaskType,
		TaskListKind:    info.Kind,
		RangeID:         info.RangeID,
		AckLevel:        info.AckLevel,
		LastUpdatedTime: info.LastUpdated,
	}
}

// From TaskManager interface
func (t *nosqlTaskManager) GetTasks(
	ctx context.Context,
	request *p.GetTasksRequest,
) (*p.InternalGetTasksResponse, error) {
	if request.MaxReadLevel == nil {
		request.MaxReadLevel = common.Int64Ptr(math.MaxInt64 - 1)
	}

	if request.ReadLevel > *request.MaxReadLevel {
		return &p.InternalGetTasksResponse{}, nil
	}

	resp, err := t.db.SelectTasks(ctx, &nosqlplugin.TasksFilter{
		TaskListFilter: nosqlplugin.TaskListFilter{
			DomainID:     request.DomainID,
			TaskListName: request.TaskList,
			TaskListType: request.TaskType,
		},
		BatchSize: request.BatchSize,

		// Change from exclusive to inclusive
		MinTaskID: request.ReadLevel + 1,
		// Change from inclusive to exclusive
		MaxTaskID: *request.MaxReadLevel + 1,
	})

	if err != nil {
		return nil, convertCommonErrors(t.db, "GetTasks", err)
	}

	response := &p.InternalGetTasksResponse{}
	for _, t := range resp {
		response.Tasks = append(response.Tasks, toTaskInfo(t))
	}

	return response, nil
}

func toTaskInfo(t *nosqlplugin.TaskRow) *p.InternalTaskInfo {
	return &p.InternalTaskInfo{
		DomainID:    t.DomainID,
		WorkflowID:  t.WorkflowID,
		RunID:       t.RunID,
		TaskID:      t.TaskID,
		ScheduleID:  t.ScheduledID,
		CreatedTime: t.CreatedTime,
	}
}

// From TaskManager interface
func (t *nosqlTaskManager) CompleteTask(
	ctx context.Context,
	request *p.CompleteTaskRequest,
) error {
	tli := request.TaskList
	err := t.db.DeleteTask(ctx, &nosqlplugin.TaskRowPK{
		DomainID:     tli.DomainID,
		TaskListName: tli.Name,
		TaskListType: tli.TaskType,
		TaskID:       request.TaskID,
	})
	if err != nil {
		return convertCommonErrors(t.db, "CompleteTask", err)
	}

	return nil
}

// CompleteTasksLessThan deletes all tasks less than or equal to the given task id. This API ignores the
// Limit request parameter i.e. either all tasks leq the task_id will be deleted or an error will
// be returned to the caller
func (t *nosqlTaskManager) CompleteTasksLessThan(
	ctx context.Context,
	request *p.CompleteTasksLessThanRequest,
) (int, error) {
	num, err := t.db.RangeDeleteTasks(ctx, &nosqlplugin.TaskRowPK{
		DomainID:     request.DomainID,
		TaskListName: request.TaskListName,
		TaskListType: request.TaskType,
		TaskID:       request.TaskID,
	}, request.Limit)
	if err != nil {
		return 0, convertCommonErrors(t.db, "CompleteTasksLessThan", err)
	}
	return num, nil
}
