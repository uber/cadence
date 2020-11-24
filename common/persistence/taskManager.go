// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package persistence

import (
	"context"

	"github.com/uber/cadence/common"
)

type (
	taskManager struct {
		persistence TaskStore
	}
)

var _ TaskManager = (*taskManager)(nil)

// NewTaskManager returns a new TaskManager
func NewTaskManager(
	persistence TaskStore,
) TaskManager {
	return &taskManager{
		persistence: persistence,
	}
}

func (t *taskManager) GetName() string {
	return t.persistence.GetName()
}

func (t *taskManager) Close() {
	t.persistence.Close()
}

func (t *taskManager) LeaseTaskList(ctx context.Context, request *LeaseTaskListRequest) (*LeaseTaskListResponse, error) {
	return t.persistence.LeaseTaskList(ctx, request)
}

func (t *taskManager) UpdateTaskList(ctx context.Context, request *UpdateTaskListRequest) (*UpdateTaskListResponse, error) {
	return t.persistence.UpdateTaskList(ctx, request)
}

func (t *taskManager) ListTaskList(ctx context.Context, request *ListTaskListRequest) (*ListTaskListResponse, error) {
	return t.persistence.ListTaskList(ctx, request)
}

func (t *taskManager) DeleteTaskList(ctx context.Context, request *DeleteTaskListRequest) error {
	return t.persistence.DeleteTaskList(ctx, request)
}

func (t *taskManager) CreateTasks(ctx context.Context, request *CreateTasksRequest) (*CreateTasksResponse, error) {
	var internalCreateTasks []*InternalCreateTasksInfo
	for _, task := range request.Tasks {
		internalCreateTasks = append(internalCreateTasks, t.toInternalCreateTaskInfo(task))
	}
	internalRequest := &InternalCreateTasksRequest{
		TaskListInfo: request.TaskListInfo,
		Tasks:        internalCreateTasks,
	}
	_, err := t.persistence.CreateTasks(ctx, internalRequest)
	if err != nil {
		return nil, err
	}
	return &CreateTasksResponse{}, err
}

func (t *taskManager) GetTasks(ctx context.Context, request *GetTasksRequest) (*GetTasksResponse, error) {
	internalResult, err := t.persistence.GetTasks(ctx, request)
	if err != nil {
		return nil, err
	}
	var taskInfo []*TaskInfo
	for _, task := range internalResult.Tasks {
		taskInfo = append(taskInfo, t.fromInternalTaskInfo(task))
	}
	return &GetTasksResponse{Tasks: taskInfo}, nil
}

func (t *taskManager) CompleteTask(ctx context.Context, request *CompleteTaskRequest) error {
	return t.persistence.CompleteTask(ctx, request)
}

func (t *taskManager) CompleteTasksLessThan(ctx context.Context, request *CompleteTasksLessThanRequest) (int, error) {
	return t.persistence.CompleteTasksLessThan(ctx, request)
}

func (t *taskManager) toInternalCreateTaskInfo(createTaskInfo *CreateTaskInfo) *InternalCreateTasksInfo {
	if createTaskInfo == nil {
		return nil
	}
	return &InternalCreateTasksInfo{
		Execution: createTaskInfo.Execution,
		Data:      t.toInternalTaskInfo(createTaskInfo.Data),
		TaskID:    createTaskInfo.TaskID,
	}
}

func (t *taskManager) toInternalTaskInfo(taskInfo *TaskInfo) *InternalTaskInfo {
	if taskInfo == nil {
		return nil
	}
	return &InternalTaskInfo{
		DomainID:               taskInfo.DomainID,
		WorkflowID:             taskInfo.WorkflowID,
		RunID:                  taskInfo.RunID,
		TaskID:                 taskInfo.TaskID,
		ScheduleID:             taskInfo.ScheduleID,
		ScheduleToStartTimeout: common.SecondsToDuration(int64(taskInfo.ScheduleToStartTimeout)),
		Expiry:                 taskInfo.Expiry,
		CreatedTime:            taskInfo.CreatedTime,
	}
}
func (t *taskManager) fromInternalTaskInfo(internalTaskInfo *InternalTaskInfo) *TaskInfo {
	if internalTaskInfo == nil {
		return nil
	}
	return &TaskInfo{
		DomainID:               internalTaskInfo.DomainID,
		WorkflowID:             internalTaskInfo.WorkflowID,
		RunID:                  internalTaskInfo.RunID,
		TaskID:                 internalTaskInfo.TaskID,
		ScheduleID:             internalTaskInfo.ScheduleID,
		ScheduleToStartTimeout: int32(internalTaskInfo.ScheduleToStartTimeout.Seconds()),
		Expiry:                 internalTaskInfo.Expiry,
		CreatedTime:            internalTaskInfo.CreatedTime,
	}
}
