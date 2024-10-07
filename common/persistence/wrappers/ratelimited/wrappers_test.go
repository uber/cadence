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

package ratelimited

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/types"
)

var _staticMethods = map[string]bool{
	"Close":      true,
	"GetName":    true,
	"GetShardID": true,
}

var wrappers = []any{
	&ratelimitedConfigStoreManager{},
	&ratelimitedDomainManager{},
	&ratelimitedHistoryManager{},
	&ratelimitedQueueManager{},
	&ratelimitedShardManager{},
	&ratelimitedTaskManager{},
	&ratelimitedVisibilityManager{},
	&ratelimitedExecutionManager{},
}

func TestClientsRateLimitAlwaysAllow(t *testing.T) {
	for _, injector := range wrappers {
		name := reflect.TypeOf(injector).String()
		t.Run(name, func(t *testing.T) {
			object := builderForPassThrough(t, injector, &limiterAlwaysAllow{}, true, nil)
			v := reflect.ValueOf(object)
			infoT := reflect.TypeOf(v.Interface())
			for i := 0; i < infoT.NumMethod(); i++ {
				method := infoT.Method(i)
				if _staticMethods[method.Name] {
					// Skip methods that do not use error injection.
					continue
				}
				t.Run(method.Name, func(t *testing.T) {
					vals := make([]reflect.Value, 0, method.Type.NumIn()-1)
					// First argument is always context.Context
					vals = append(vals, reflect.ValueOf(context.Background()))
					for i := 2; i < method.Type.NumIn(); i++ {
						vals = append(vals, reflect.Zero(method.Type.In(i)))
					}

					callRes := v.MethodByName(method.Name).Call(vals)
					resultErr := callRes[len(callRes)-1].Interface()

					assert.Nil(t, resultErr, "method %v returned error %v", method.Name, resultErr)
				})
			}
		})
	}
}

func TestClientsAlwaysRateLimited(t *testing.T) {
	for _, injector := range wrappers {
		name := reflect.TypeOf(injector).String()
		t.Run(name, func(t *testing.T) {
			object := builderForPassThrough(t, injector, &limiterNeverAllow{}, false, nil)
			v := reflect.ValueOf(object)
			infoT := reflect.TypeOf(v.Interface())
			for i := 0; i < infoT.NumMethod(); i++ {
				method := infoT.Method(i)
				if _staticMethods[method.Name] {
					// Skip methods that do not use error injection.
					continue
				}
				t.Run(method.Name, func(t *testing.T) {
					vals := make([]reflect.Value, 0, method.Type.NumIn()-1)
					// First argument is always context.Context
					vals = append(vals, reflect.ValueOf(context.Background()))
					for i := 2; i < method.Type.NumIn(); i++ {
						vals = append(vals, reflect.Zero(method.Type.In(i)))
					}

					callRes := v.MethodByName(method.Name).Call(vals)
					resultErr := callRes[len(callRes)-1].Interface()

					err, ok := resultErr.(error)
					require.True(t, ok, "method %v must return error")
					var expectedErr *types.ServiceBusyError
					assert.True(t, errors.As(err, &expectedErr), "method %v must return error of type *types.ServiceBusyError", method.Name)
					assert.Equal(t, ErrPersistenceLimitExceeded.Message, expectedErr.Message, "method %v returned different error, expected %v, got %v", method.Name, ErrPersistenceLimitExceeded.Message, expectedErr.Message)
				})
			}
		})
	}
}

func TestInjectorsWithUnderlyingErrors(t *testing.T) {
	for _, injector := range wrappers {
		name := reflect.TypeOf(injector).String()
		t.Run(name, func(t *testing.T) {
			expectedMethodErr := fmt.Errorf("%s: injected error", name)
			object := builderForPassThrough(t, injector, &limiterAlwaysAllow{}, true, expectedMethodErr)
			v := reflect.ValueOf(object)
			infoT := reflect.TypeOf(v.Interface())
			for i := 0; i < infoT.NumMethod(); i++ {
				method := infoT.Method(i)
				if _staticMethods[method.Name] {
					// Skip methods that do not use error injection.
					continue
				}
				t.Run(method.Name, func(t *testing.T) {
					vals := make([]reflect.Value, 0, method.Type.NumIn()-1)
					// First argument is always context.Context
					vals = append(vals, reflect.ValueOf(context.Background()))
					for i := 2; i < method.Type.NumIn(); i++ {
						vals = append(vals, reflect.Zero(method.Type.In(i)))
					}

					callRes := v.MethodByName(method.Name).Call(vals)
					resultErr := callRes[len(callRes)-1].Interface()
					err, ok := resultErr.(error)
					require.True(t, ok, "method %v must return error")
					assert.Equal(t, expectedMethodErr, err, "method %v returned different error, expected %v, got %v", method.Name, expectedMethodErr, err)
				})
			}
		})
	}
}

func builderForPassThrough(t *testing.T, injector any, limiter quotas.Limiter, expectCalls bool, expectedErr error) (object any) {
	ctrl := gomock.NewController(t)
	switch injector.(type) {
	case *ratelimitedConfigStoreManager:
		mocked := persistence.NewMockConfigStoreManager(ctrl)
		object = NewConfigStoreManager(mocked, limiter)
		if expectCalls {
			mocked.EXPECT().UpdateDynamicConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().FetchDynamicConfig(gomock.Any(), gomock.Any()).Return(&persistence.FetchDynamicConfigResponse{}, expectedErr)
		}
	case *ratelimitedDomainManager:
		mocked := persistence.NewMockDomainManager(ctrl)
		object = NewDomainManager(mocked, limiter)
		if expectCalls {
			mocked.EXPECT().CreateDomain(gomock.Any(), gomock.Any()).Return(&persistence.CreateDomainResponse{}, expectedErr)
			mocked.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(&persistence.GetDomainResponse{}, expectedErr)
			mocked.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().DeleteDomain(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().DeleteDomainByName(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(&persistence.ListDomainsResponse{}, expectedErr)
			mocked.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{}, expectedErr)
		}
	case *ratelimitedHistoryManager:
		mocked := persistence.NewMockHistoryManager(ctrl)
		object = NewHistoryManager(mocked, limiter)
		if expectCalls {
			mocked.EXPECT().AppendHistoryNodes(gomock.Any(), gomock.Any()).Return(&persistence.AppendHistoryNodesResponse{}, expectedErr)
			mocked.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ReadHistoryBranchResponse{}, expectedErr)
			mocked.EXPECT().ReadHistoryBranchByBatch(gomock.Any(), gomock.Any()).Return(&persistence.ReadHistoryBranchByBatchResponse{}, expectedErr)
			mocked.EXPECT().ReadRawHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ReadRawHistoryBranchResponse{}, expectedErr)
			mocked.EXPECT().ForkHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ForkHistoryBranchResponse{}, expectedErr)
			mocked.EXPECT().DeleteHistoryBranch(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().GetHistoryTree(gomock.Any(), gomock.Any()).Return(&persistence.GetHistoryTreeResponse{}, expectedErr)
			mocked.EXPECT().GetAllHistoryTreeBranches(gomock.Any(), gomock.Any()).Return(&persistence.GetAllHistoryTreeBranchesResponse{}, expectedErr)
		}
	case *ratelimitedQueueManager:
		mocked := persistence.NewMockQueueManager(ctrl)
		object = NewQueueManager(mocked, limiter)
		if expectCalls {
			mocked.EXPECT().EnqueueMessage(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().ReadMessages(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*persistence.QueueMessage{}, expectedErr)
			mocked.EXPECT().UpdateAckLevel(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().GetAckLevels(gomock.Any()).Return(map[string]int64{}, expectedErr)
			mocked.EXPECT().DeleteMessagesBefore(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().DeleteMessageFromDLQ(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().EnqueueMessageToDLQ(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().GetDLQAckLevels(gomock.Any()).Return(map[string]int64{}, expectedErr)
			mocked.EXPECT().GetDLQSize(gomock.Any()).Return(int64(0), expectedErr)
			mocked.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().ReadMessagesFromDLQ(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*persistence.QueueMessage{}, nil, expectedErr)
			mocked.EXPECT().UpdateDLQAckLevel(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)
		}
	case *ratelimitedShardManager:
		mocked := persistence.NewMockShardManager(ctrl)
		object = NewShardManager(mocked, limiter)
		if expectCalls {
			mocked.EXPECT().GetShard(gomock.Any(), gomock.Any()).Return(&persistence.GetShardResponse{}, expectedErr)
			mocked.EXPECT().UpdateShard(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().CreateShard(gomock.Any(), gomock.Any()).Return(expectedErr)
		}
	case *ratelimitedTaskManager:
		mocked := persistence.NewMockTaskManager(ctrl)
		object = NewTaskManager(mocked, limiter)
		if expectCalls {
			mocked.EXPECT().CompleteTasksLessThan(gomock.Any(), gomock.Any()).Return(&persistence.CompleteTasksLessThanResponse{}, expectedErr)
			mocked.EXPECT().CompleteTask(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().CreateTasks(gomock.Any(), gomock.Any()).Return(&persistence.CreateTasksResponse{}, expectedErr)
			mocked.EXPECT().DeleteTaskList(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().GetTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetTasksResponse{}, expectedErr)
			mocked.EXPECT().GetOrphanTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetOrphanTasksResponse{}, expectedErr)
			mocked.EXPECT().GetTaskListSize(gomock.Any(), gomock.Any()).Return(&persistence.GetTaskListSizeResponse{}, expectedErr)
			mocked.EXPECT().LeaseTaskList(gomock.Any(), gomock.Any()).Return(&persistence.LeaseTaskListResponse{}, expectedErr)
			mocked.EXPECT().GetTaskList(gomock.Any(), gomock.Any()).Return(&persistence.GetTaskListResponse{}, expectedErr)
			mocked.EXPECT().ListTaskList(gomock.Any(), gomock.Any()).Return(&persistence.ListTaskListResponse{}, expectedErr)
			mocked.EXPECT().UpdateTaskList(gomock.Any(), gomock.Any()).Return(&persistence.UpdateTaskListResponse{}, expectedErr)
		}
	case *ratelimitedVisibilityManager:
		mocked := persistence.NewMockVisibilityManager(ctrl)
		object = NewVisibilityManager(mocked, limiter)
		if expectCalls {
			mocked.EXPECT().DeleteUninitializedWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.CountWorkflowExecutionsResponse{}, expectedErr)
			mocked.EXPECT().GetClosedWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.GetClosedWorkflowExecutionResponse{}, expectedErr)
			mocked.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr)
			mocked.EXPECT().ListClosedWorkflowExecutionsByStatus(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr)
			mocked.EXPECT().ListClosedWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr)
			mocked.EXPECT().ListClosedWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr)
			mocked.EXPECT().ListOpenWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr)
			mocked.EXPECT().ListOpenWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr)
			mocked.EXPECT().ListOpenWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr)
			mocked.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr)
			mocked.EXPECT().RecordWorkflowExecutionStarted(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().RecordWorkflowExecutionUninitialized(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().UpsertWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr)
		}
	case *ratelimitedExecutionManager:
		mocked := persistence.NewMockExecutionManager(ctrl)
		object = NewExecutionManager(mocked, limiter)
		if expectCalls {
			mocked.EXPECT().CompleteTimerTask(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().CompleteTransferTask(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().CreateWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.CreateWorkflowExecutionResponse{}, expectedErr)
			mocked.EXPECT().GetWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.GetWorkflowExecutionResponse{}, expectedErr)
			mocked.EXPECT().UpdateWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.UpdateWorkflowExecutionResponse{}, expectedErr)
			mocked.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().DeleteCurrentWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().GetCurrentExecution(gomock.Any(), gomock.Any()).Return(&persistence.GetCurrentExecutionResponse{}, expectedErr)
			mocked.EXPECT().CompleteReplicationTask(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().ConflictResolveWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.ConflictResolveWorkflowExecutionResponse{}, expectedErr)
			mocked.EXPECT().CreateFailoverMarkerTasks(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().DeleteReplicationTaskFromDLQ(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().GetReplicationDLQSize(gomock.Any(), gomock.Any()).Return(&persistence.GetReplicationDLQSizeResponse{}, expectedErr)
			mocked.EXPECT().GetReplicationTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetReplicationTasksResponse{}, expectedErr)
			mocked.EXPECT().GetReplicationTasksFromDLQ(gomock.Any(), gomock.Any()).Return(&persistence.GetReplicationTasksFromDLQResponse{}, expectedErr)
			mocked.EXPECT().GetTimerIndexTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetTimerIndexTasksResponse{}, expectedErr)
			mocked.EXPECT().GetTransferTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetTransferTasksResponse{}, expectedErr)
			mocked.EXPECT().IsWorkflowExecutionExists(gomock.Any(), gomock.Any()).Return(&persistence.IsWorkflowExecutionExistsResponse{}, expectedErr)
			mocked.EXPECT().ListConcreteExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListConcreteExecutionsResponse{}, expectedErr)
			mocked.EXPECT().ListCurrentExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListCurrentExecutionsResponse{}, expectedErr)
			mocked.EXPECT().PutReplicationTaskToDLQ(gomock.Any(), gomock.Any()).Return(expectedErr)
			mocked.EXPECT().RangeCompleteReplicationTask(gomock.Any(), gomock.Any()).Return(&persistence.RangeCompleteReplicationTaskResponse{}, expectedErr)
			mocked.EXPECT().RangeCompleteTimerTask(gomock.Any(), gomock.Any()).Return(&persistence.RangeCompleteTimerTaskResponse{}, expectedErr)
			mocked.EXPECT().RangeCompleteTransferTask(gomock.Any(), gomock.Any()).Return(&persistence.RangeCompleteTransferTaskResponse{}, expectedErr)
			mocked.EXPECT().RangeDeleteReplicationTaskFromDLQ(gomock.Any(), gomock.Any()).Return(&persistence.RangeDeleteReplicationTaskFromDLQResponse{}, expectedErr)
		}
	default:
		t.Errorf("unsupported type %v", reflect.TypeOf(injector))
		t.FailNow()
	}
	return
}
