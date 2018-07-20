// Copyright (c) 2018 Uber Technologies, Inc.
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

package sql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/uber/cadence/common/persistence"
)

type (
	sqlTaskManager struct {
		db *sqlx.DB
	}
)

func NewTaskPersistence(username, password, host, port, dbName string) (persistence.TaskManager, error) {
	var db, err = sqlx.Connect("mysql",
		fmt.Sprintf(Dsn, username, password, host, port, dbName))
	if err != nil {
		return nil, err
	}

	return &sqlTaskManager{
		db: db,
	}, nil
}

func (m *sqlTaskManager) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m *sqlTaskManager) LeaseTaskList(request *persistence.LeaseTaskListRequest) (*persistence.LeaseTaskListResponse, error) {
	return &persistence.LeaseTaskListResponse{&persistence.TaskListInfo{}}, nil
}

func (m *sqlTaskManager) UpdateTaskList(request *persistence.UpdateTaskListRequest) (*persistence.UpdateTaskListResponse, error) {
	panic("implement me")
}

func (m *sqlTaskManager) CreateTasks(request *persistence.CreateTasksRequest) (*persistence.CreateTasksResponse, error) {
	return nil, nil
}

func (m *sqlTaskManager) GetTasks(request *persistence.GetTasksRequest) (*persistence.GetTasksResponse, error) {
	return &persistence.GetTasksResponse{[](*persistence.TaskInfo){}}, nil
}

func (m *sqlTaskManager) CompleteTask(request *persistence.CompleteTaskRequest) error {
	panic("implement me")
}
