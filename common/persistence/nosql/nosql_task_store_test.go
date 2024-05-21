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
	ctx "context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/types"
)

const (
	TestDomainID     = "test-domain-id"
	TestDomainName   = "test-domain"
	TestTaskListName = "test-tasklist"
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

	size, err := store.GetTaskListSize(ctx.Background(), &persistence.GetTaskListSizeRequest{
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
