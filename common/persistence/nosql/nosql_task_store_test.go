// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package nosql

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/types"
)

const (
	TestDomainID     = "test-domain-id"
	TestDomainName   = "test-domain"
	TestTaskListName = "test-tasklist"
	TestWorkflowID   = "test-workflow-id"
	TestRunID        = "test-run-id"
)

func TestNewNoSQLStore(t *testing.T) {
	registerCassandraMock(t)
	cfg := getValidShardedNoSQLConfig()

	store, err := newNoSQLTaskStore(cfg, log.NewNoop(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func setupNoSQLStoreMocks(t *testing.T) (*nosqlTaskStore, *nosqlplugin.MockDB) {
	ctrl := gomock.NewController(t)
	dbMock := nosqlplugin.NewMockDB(ctrl)

	nosqlSt := nosqlStore{
		logger: log.NewNoop(),
		db:     dbMock,
	}

	shardedNosqlStoreMock := NewMockshardedNosqlStore(ctrl)
	shardedNosqlStoreMock.EXPECT().
		GetStoreShardByTaskList(
			TestDomainID,
			TestTaskListName,
			int(types.TaskListTypeDecision)).
		Return(&nosqlSt, nil).
		AnyTimes()

	store := &nosqlTaskStore{
		shardedNosqlStore: shardedNosqlStoreMock,
	}

	return store, dbMock
}

func TestGetOrphanTasks(t *testing.T) {
	store, _ := setupNoSQLStoreMocks(t)

	// We just expect the function to return an error so we don't need to check the result
	_, err := store.GetOrphanTasks(context.Background(), nil)

	var expectedErr *types.InternalServiceError
	assert.ErrorAs(t, err, &expectedErr)
	assert.ErrorContains(t, err, "Unimplemented call to GetOrphanTasks for NoSQL")
}

func TestGetTaskListSize(t *testing.T) {
	store, db := setupNoSQLStoreMocks(t)

	db.EXPECT().GetTasksCount(
		gomock.Any(),
		&nosqlplugin.TasksFilter{
			TaskListFilter: nosqlplugin.TaskListFilter{
				DomainID:     TestDomainID,
				TaskListName: TestTaskListName,
				TaskListType: int(types.TaskListTypeDecision),
			},
			MinTaskID: 456,
		},
	).Return(int64(123), nil)

	size, err := store.GetTaskListSize(context.Background(), &persistence.GetTaskListSizeRequest{
		DomainID:     TestDomainID,
		DomainName:   TestDomainName,
		TaskListName: TestTaskListName,
		TaskListType: int(types.TaskListTypeDecision),
		AckLevel:     456,
	})

	assert.NoError(t, err)
	assert.Equal(t,
		&persistence.GetTaskListSizeResponse{Size: 123},
		size,
	)
}

func TestGetTasks(t *testing.T) {
	store, db := setupNoSQLStoreMocks(t)
	now := time.Unix(123, 456)

	taskrow1 := nosqlplugin.TaskRow{
		DomainID:        TestDomainID,
		TaskListName:    TestTaskListName,
		TaskListType:    int(types.TaskListTypeDecision),
		TaskID:          5,
		WorkflowID:      TestWorkflowID,
		RunID:           TestRunID,
		ScheduledID:     0,
		CreatedTime:     now,
		PartitionConfig: nil,
	}

	taskrow2 := taskrow1
	taskrow2.TaskID = 6

	db.EXPECT().SelectTasks(gomock.Any(), &nosqlplugin.TasksFilter{
		TaskListFilter: *getDecisionTaskListFilter(),
		BatchSize:      100,
		MinTaskID:      1,
		MaxTaskID:      15,
	}).Return([]*nosqlplugin.TaskRow{&taskrow1, &taskrow2}, nil)

	resp, err := store.GetTasks(context.Background(), getValidGetTasksRequest())

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Tasks, 2)
	taskRowEqualTaskInfo(t, taskrow1, resp.Tasks[0])
	taskRowEqualTaskInfo(t, taskrow2, resp.Tasks[1])
}

func TestGetTasks_Empty(t *testing.T) {
	store, _ := setupNoSQLStoreMocks(t)

	request := getValidGetTasksRequest()
	// Set the max read level to be less than the min read level
	request.ReadLevel = 10
	request.MaxReadLevel = common.Int64Ptr(5)
	resp, err := store.GetTasks(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Tasks)
}

func taskRowEqualTaskInfo(t *testing.T, taskrow1 nosqlplugin.TaskRow, taskInfo1 *persistence.InternalTaskInfo) {
	assert.Equal(t, taskrow1.DomainID, taskInfo1.DomainID)
	assert.Equal(t, taskrow1.WorkflowID, taskInfo1.WorkflowID)
	assert.Equal(t, taskrow1.RunID, taskInfo1.RunID)
	assert.Equal(t, taskrow1.TaskID, taskInfo1.TaskID)
	assert.Equal(t, taskrow1.ScheduledID, taskInfo1.ScheduleID)
	assert.Equal(t, taskrow1.CreatedTime, taskInfo1.CreatedTime)
	assert.Equal(t, taskrow1.PartitionConfig, taskInfo1.PartitionConfig)
}

func getValidGetTasksRequest() *persistence.GetTasksRequest {
	return &persistence.GetTasksRequest{
		DomainID:   TestDomainID,
		DomainName: TestDomainName,
		TaskList:   TestTaskListName,
		TaskType:   int(types.TaskListTypeDecision),
		// The read level is the smallest taskID that we want to read, the maxReadLevel is the largest
		ReadLevel:    1,
		MaxReadLevel: common.Int64Ptr(15),
		BatchSize:    100,
	}
}

func getDecisionTaskListFilter() *nosqlplugin.TaskListFilter {
	return &nosqlplugin.TaskListFilter{
		DomainID:     TestDomainID,
		TaskListName: TestTaskListName,
		TaskListType: int(types.TaskListTypeDecision),
	}
}
