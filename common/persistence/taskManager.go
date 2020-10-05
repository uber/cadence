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
	internalRequest := &InternalLeaseTaskListRequest{
		DomainID:     request.DomainID,
		TaskList:     request.TaskList,
		TaskType:     request.TaskType,
		TaskListKind: request.TaskListKind,
		RangeID:      request.RangeID,
	}
	internalResult, err := t.persistence.LeaseTaskList(ctx, internalRequest)
	if err != nil {
		return nil, err
	}
	result := &LeaseTaskListResponse{
		TaskListInfo: t.fromInternalTaskListInfo(internalResult.TaskListInfo),
	}
	return result, nil
}

func (t *taskManager) UpdateTaskList(ctx context.Context, request *UpdateTaskListRequest) (*UpdateTaskListResponse, error) {
	internalRequest := &InternalUpdateTaskListRequest{
		TaskListInfo: t.toInternalTaskListInfo(request.TaskListInfo),
	}
	_, err := t.persistence.UpdateTaskList(ctx, internalRequest)
	if err != nil {
		return nil, err
	}
	return &UpdateTaskListResponse{}, nil
}

func (t *taskManager) ListTaskList(ctx context.Context, request *ListTaskListRequest) (*ListTaskListResponse, error) {
	internalRequest := &InternalListTaskListRequest{
		PageSize:  request.PageSize,
		PageToken: request.PageToken,
	}
	internalResult, err := t.persistence.ListTaskList(ctx, internalRequest)
	if err != nil {
		return nil, err
	}
	var taskListInfoList []TaskListInfo
	for _, tlInfo := range internalResult.Items {
		taskListInfoList = append(taskListInfoList, *t.fromInternalTaskListInfo(&tlInfo))
	}
	result := &ListTaskListResponse{
		Items:         taskListInfoList,
		NextPageToken: internalResult.NextPageToken,
	}
	return result, nil
}

func (t *taskManager) DeleteTaskList(ctx context.Context, request *DeleteTaskListRequest) error {
	internalRequest := &InternalDeleteTaskListRequest{
		DomainID:     request.DomainID,
		TaskListName: request.TaskListName,
		TaskListType: request.TaskListType,
		RangeID:      request.RangeID,
	}
	return t.persistence.DeleteTaskList(ctx, internalRequest)
}

func (t *taskManager) CreateTasks(ctx context.Context, request *CreateTasksRequest) (*CreateTasksResponse, error) {
	var internalCreateTasks []*InternalCreateTasksInfo
	for _, task := range request.Tasks {
		internalCreateTasks = append(internalCreateTasks, t.toInternalCreateTaskInfo(task))
	}
	internalRequest := &InternalCreateTasksRequest{
		TaskListInfo: t.toInternalTaskListInfo(request.TaskListInfo),
		Tasks:        internalCreateTasks,
	}
	_, err := t.persistence.CreateTasks(ctx, internalRequest)
	if err != nil {
		return nil, err
	}
	return &CreateTasksResponse{}, err
}

func (t *taskManager) GetTasks(ctx context.Context, request *GetTasksRequest) (*GetTasksResponse, error) {
	internalRequest := &InternalGetTasksRequest{
		DomainID:     request.DomainID,
		TaskList:     request.TaskList,
		TaskType:     request.TaskType,
		ReadLevel:    request.ReadLevel,
		MaxReadLevel: request.MaxReadLevel,
		BatchSize:    request.BatchSize,
	}
	internalResult, err := t.persistence.GetTasks(ctx, internalRequest)
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
	internalRequest := &InternalCompleteTaskRequest{
		TaskList: t.toInternalTaskListInfo(request.TaskList),
		TaskID:   request.TaskID,
	}
	return t.persistence.CompleteTask(ctx, internalRequest)
}

func (t *taskManager) CompleteTasksLessThan(ctx context.Context, request *CompleteTasksLessThanRequest) (int, error) {
	internalRequest := &InternalCompleteTasksLessThanRequest{
		DomainID:     request.DomainID,
		TaskListName: request.TaskListName,
		TaskType:     request.TaskType,
		TaskID:       request.TaskID,
		Limit:        request.Limit,
	}
	return t.persistence.CompleteTasksLessThan(ctx, internalRequest)
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
		ScheduleToStartTimeout: taskInfo.ScheduleToStartTimeout,
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
		ScheduleToStartTimeout: internalTaskInfo.ScheduleToStartTimeout,
		Expiry:                 internalTaskInfo.Expiry,
		CreatedTime:            internalTaskInfo.CreatedTime,
	}
}

func (t *taskManager) toInternalTaskListInfo(taskListInfo *TaskListInfo) *InternalTaskListInfo {
	if taskListInfo == nil {
		return nil
	}
	internalTaskListInfo := &InternalTaskListInfo{
		DomainID:    taskListInfo.DomainID,
		Name:        taskListInfo.Name,
		TaskType:    taskListInfo.TaskType,
		RangeID:     taskListInfo.RangeID,
		AckLevel:    taskListInfo.AckLevel,
		Kind:        taskListInfo.Kind,
		Expiry:      taskListInfo.Expiry,
		LastUpdated: taskListInfo.LastUpdated,
	}
	return internalTaskListInfo
}

func (t *taskManager) fromInternalTaskListInfo(internalTaskListInfo *InternalTaskListInfo) *TaskListInfo {
	if internalTaskListInfo == nil {
		return nil
	}
	taskListInfo := &TaskListInfo{
		DomainID:    internalTaskListInfo.DomainID,
		Name:        internalTaskListInfo.Name,
		TaskType:    internalTaskListInfo.TaskType,
		RangeID:     internalTaskListInfo.RangeID,
		AckLevel:    internalTaskListInfo.AckLevel,
		Kind:        internalTaskListInfo.Kind,
		Expiry:      internalTaskListInfo.Expiry,
		LastUpdated: internalTaskListInfo.LastUpdated,
	}
	return taskListInfo
}
