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

package persistence

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/log/testlogger"
)

func TestExecutionManager_ProxyStoreMethods(t *testing.T) {
	for _, tc := range []struct {
		method       string
		prepareMocks func(*MockExecutionStore)
	}{
		{
			method: "GetShardID",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().GetShardID().Return(1).Times(1)
			},
		},
		{
			method: "GetName",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().GetName().Return("test").Times(1)
			},
		},
		{
			method: "Close",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().Close().Return().Times(1)
			},
		},
		{
			method: "GetTransferTasks",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().GetTransferTasks(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "CompleteTransferTask",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().CompleteTransferTask(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			method: "RangeCompleteTransferTask",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().RangeCompleteTransferTask(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "GetCrossClusterTasks",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().GetCrossClusterTasks(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "CompleteCrossClusterTask",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().CompleteCrossClusterTask(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			method: "RangeCompleteCrossClusterTask",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().RangeCompleteCrossClusterTask(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "DeleteReplicationTaskFromDLQ",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().DeleteReplicationTaskFromDLQ(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			method: "RangeDeleteReplicationTaskFromDLQ",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().RangeDeleteReplicationTaskFromDLQ(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "CreateFailoverMarkerTasks",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().CreateFailoverMarkerTasks(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			method: "GetTimerIndexTasks",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().GetTimerIndexTasks(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "CompleteTimerTask",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().CompleteTimerTask(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			method: "RangeCompleteTimerTask",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().RangeCompleteTimerTask(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "CompleteReplicationTask",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().CompleteReplicationTask(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			method: "RangeCompleteReplicationTask",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().RangeCompleteReplicationTask(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "DeleteWorkflowExecution",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			method: "DeleteCurrentWorkflowExecution",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().DeleteCurrentWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			method: "GetCurrentExecution",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().GetCurrentExecution(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "ListCurrentExecutions",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().ListCurrentExecutions(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "IsWorkflowExecutionExists",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().IsWorkflowExecutionExists(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			method: "GetReplicationDLQSize",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().GetReplicationDLQSize(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockedStore := NewMockExecutionStore(ctrl)
			tc.prepareMocks(mockedStore)
			manager := NewExecutionManagerImpl(mockedStore, testlogger.New(t), nil)
			v := reflect.ValueOf(manager)
			method := v.MethodByName(tc.method)
			methodType := method.Type()
			args := methodType.NumIn()
			var vals []reflect.Value
			// If a method requires arguments, we expect the first argument to be a context
			// and the rest to be zero values of the correct type.
			// For methods like Close and GetShardID, we don't expect any arguments.
			if args > 0 {
				vals = append(vals, reflect.ValueOf(context.Background()))
				for i := 1; i < args; i++ {
					vals = append(vals, reflect.Zero(methodType.In(i)))
				}
			}

			callRes := method.Call(vals)
			if callRes == nil {
				return
			}
			resultErr := callRes[len(callRes)-1].Interface()
			err, ok := resultErr.(error)
			if ok {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetReplicationTasks(t *testing.T) {
	for _, tc := range []struct {
		name         string
		prepareMocks func(*MockExecutionStore)
		checkRes     func(*testing.T, *GetReplicationTasksResponse, error)
	}{
		{
			name: "success",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().GetReplicationTasks(gomock.Any(), gomock.Any()).Return(&InternalGetReplicationTasksResponse{
					Tasks: []*InternalReplicationTaskInfo{
						{
							DomainID: "test",
							TaskID:   1,
						},
						{
							DomainID: "test",
							TaskID:   2,
						},
					},
					NextPageToken: nil,
				}, nil)
			},
			checkRes: func(t *testing.T, res *GetReplicationTasksResponse, err error) {
				assert.NoError(t, err)
				assert.Len(t, res.Tasks, 2)
				assert.Equal(t, int64(1), res.Tasks[0].TaskID)
				assert.Equal(t, int64(2), res.Tasks[1].TaskID)
				assert.Nil(t, res.NextPageToken)
			},
		},
		{
			name: "error",
			prepareMocks: func(mockedStore *MockExecutionStore) {
				mockedStore.EXPECT().GetReplicationTasks(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
			},
			checkRes: func(t *testing.T, res *GetReplicationTasksResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockedStore := NewMockExecutionStore(ctrl)
			tc.prepareMocks(mockedStore)
			manager := NewExecutionManagerImpl(mockedStore, testlogger.New(t), nil)
			res, err := manager.GetReplicationTasks(context.Background(), &GetReplicationTasksRequest{})
			tc.checkRes(t, res, err)
		})
	}
}
