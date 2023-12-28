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

package metered

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/exp/maps"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

var _staticMethods = map[string]bool{
	"Close":      true,
	"GetName":    true,
	"GetShardID": true,
}

func TestWrappersAgainstPreviousImplementation(t *testing.T) {
	for _, tc := range []struct {
		name        string
		prepareMock func(t *testing.T, ctrl *gomock.Controller, oldMetricsClient metrics.Client, oldLogger log.Logger, newMetricsClient metrics.Client, newLogger log.Logger) (oldManager any, newManager any, mocked any)
	}{
		{
			name: "ConfigStoreManager",
			prepareMock: func(t *testing.T, ctrl *gomock.Controller, oldMetricsClient metrics.Client, oldLogger log.Logger, newMetricsClient metrics.Client, newLogger log.Logger) (oldManager any, newManager any, mocked any) {
				wrapped := persistence.NewMockConfigStoreManager(ctrl)

				oldObj := persistence.NewConfigStorePersistenceMetricsClient(wrapped, oldMetricsClient, oldLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})
				newObj := NewConfigStoreManager(wrapped, newMetricsClient, newLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})

				return oldObj, newObj, wrapped
			},
		},
		{
			name: "DomainManager",
			prepareMock: func(t *testing.T, ctrl *gomock.Controller, oldMetricsClient metrics.Client, oldLogger log.Logger, newMetricsClient metrics.Client, newLogger log.Logger) (oldManager any, newManager any, mocked any) {
				wrapped := persistence.NewMockDomainManager(ctrl)

				oldObj := persistence.NewDomainPersistenceMetricsClient(wrapped, oldMetricsClient, oldLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})
				newObj := NewDomainManager(wrapped, newMetricsClient, newLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})

				return oldObj, newObj, wrapped
			},
		},
		{
			name: "HistoryManager",
			prepareMock: func(t *testing.T, ctrl *gomock.Controller, oldMetricsClient metrics.Client, oldLogger log.Logger, newMetricsClient metrics.Client, newLogger log.Logger) (oldManager any, newManager any, mocked any) {
				wrapped := persistence.NewMockHistoryManager(ctrl)

				oldObj := persistence.NewHistoryPersistenceMetricsClient(wrapped, oldMetricsClient, oldLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})
				newObj := NewHistoryManager(wrapped, newMetricsClient, newLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})

				return oldObj, newObj, wrapped
			},
		},
		{
			name: "QueueManager",
			prepareMock: func(t *testing.T, ctrl *gomock.Controller, oldMetricsClient metrics.Client, oldLogger log.Logger, newMetricsClient metrics.Client, newLogger log.Logger) (oldManager any, newManager any, mocked any) {
				wrapped := persistence.NewMockQueueManager(ctrl)

				oldObj := persistence.NewQueuePersistenceMetricsClient(wrapped, oldMetricsClient, oldLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})
				newObj := NewQueueManager(wrapped, newMetricsClient, newLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})

				return oldObj, newObj, wrapped
			},
		},
		{
			name: "ShardManager",
			prepareMock: func(t *testing.T, ctrl *gomock.Controller, oldMetricsClient metrics.Client, oldLogger log.Logger, newMetricsClient metrics.Client, newLogger log.Logger) (oldManager any, newManager any, mocked any) {
				wrapped := persistence.NewMockShardManager(ctrl)

				oldObj := persistence.NewShardPersistenceMetricsClient(wrapped, oldMetricsClient, oldLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})
				newObj := NewShardManager(wrapped, newMetricsClient, newLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})

				return oldObj, newObj, wrapped
			},
		},
		{
			name: "TaskManager",
			prepareMock: func(t *testing.T, ctrl *gomock.Controller, oldMetricsClient metrics.Client, oldLogger log.Logger, newMetricsClient metrics.Client, newLogger log.Logger) (oldManager any, newManager any, mocked any) {
				wrapped := persistence.NewMockTaskManager(ctrl)

				oldObj := persistence.NewTaskPersistenceMetricsClient(wrapped, oldMetricsClient, oldLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})
				newObj := NewTaskManager(wrapped, newMetricsClient, newLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})

				return oldObj, newObj, wrapped
			},
		},
		{
			name: "VisibilityManager",
			prepareMock: func(t *testing.T, ctrl *gomock.Controller, oldMetricsClient metrics.Client, oldLogger log.Logger, newMetricsClient metrics.Client, newLogger log.Logger) (oldManager any, newManager any, mocked any) {
				wrapped := persistence.NewMockVisibilityManager(ctrl)

				oldObj := persistence.NewVisibilityPersistenceMetricsClient(wrapped, oldMetricsClient, oldLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})
				newObj := NewVisibilityManager(wrapped, newMetricsClient, newLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true})

				return oldObj, newObj, wrapped
			},
		},
		{
			name: "ExecutionManager",
			prepareMock: func(t *testing.T, ctrl *gomock.Controller, oldMetricsClient metrics.Client, oldLogger log.Logger, newMetricsClient metrics.Client, newLogger log.Logger) (oldManager any, newManager any, mocked any) {
				wrapped := persistence.NewMockExecutionManager(ctrl)

				wrapped.EXPECT().GetShardID().Return(0).AnyTimes()

				oldObj := persistence.NewWorkflowExecutionPersistenceMetricsClient(wrapped, oldMetricsClient, oldLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true},
					dynamicconfig.GetIntPropertyFn(1), dynamicconfig.GetBoolPropertyFn(true))
				newObj := NewExecutionManager(wrapped, newMetricsClient, newLogger, &config.Persistence{EnablePersistenceLatencyHistogramMetrics: true},
					dynamicconfig.GetIntPropertyFn(1), dynamicconfig.GetBoolPropertyFn(true))

				return oldObj, newObj, wrapped
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("without error", func(t *testing.T) {
				ctrl := gomock.NewController(t)

				oldZapLogger, oldLogs := setupLogsCapture()
				oldMetrics := tally.NewTestScope("", nil)
				oldMetricClient := metrics.NewClient(oldMetrics, metrics.ServiceIdx(0))
				oldLogger := loggerimpl.NewLogger(oldZapLogger)

				newZapLogger, newLogs := setupLogsCapture()
				newMetrics := tally.NewTestScope("", nil)
				newMetricsClient := metrics.NewClient(newMetrics, metrics.ServiceIdx(0))
				newLogger := loggerimpl.NewLogger(newZapLogger)

				oldObj, newObj, mocked := tc.prepareMock(t, ctrl, oldMetricClient, oldLogger, newMetricsClient, newLogger)
				prepareMockForTest(t, mocked, nil)

				runScenario(t, oldObj, newObj, oldLogs, newLogs, oldMetrics, newMetrics)
			})
			t.Run("with error", func(t *testing.T) {
				ctrl := gomock.NewController(t)

				oldZapLogger, oldLogs := setupLogsCapture()
				oldMetrics := tally.NewTestScope("", nil)
				oldMetricClient := metrics.NewClient(oldMetrics, metrics.ServiceIdx(0))
				oldLogger := loggerimpl.NewLogger(oldZapLogger)

				newZapLogger, newLogs := setupLogsCapture()
				newMetrics := tally.NewTestScope("", nil)
				newMetricsClient := metrics.NewClient(newMetrics, metrics.ServiceIdx(0))
				newLogger := loggerimpl.NewLogger(newZapLogger)

				oldObj, newObj, mocked := tc.prepareMock(t, ctrl, oldMetricClient, oldLogger, newMetricsClient, newLogger)
				prepareMockForTest(t, mocked, errors.New("persistence error"))

				runScenario(t, oldObj, newObj, oldLogs, newLogs, oldMetrics, newMetrics)
			})
		})
	}
}

func assertLogs(t *testing.T, oldLogs *observer.ObservedLogs, newLogs *observer.ObservedLogs) {
	for _, oldLog := range oldLogs.All() {
		assert.NotEmpty(t, newLogs.FilterMessage(oldLog.Message), "log message %v is not found", oldLog.Message)
	}
}

func assertMetrics(t *testing.T, old tally.Snapshot, new tally.Snapshot) {
	assert.Equal(t, old.Counters(), new.Counters(), "counter should be the same")
	assert.Equal(t, old.Gauges(), new.Gauges(), "gauge should be the same")

	oldTimerNames := maps.Keys(old.Timers())
	sort.Strings(oldTimerNames)
	newTimerNames := maps.Keys(new.Timers())
	sort.Strings(newTimerNames)

	assert.Equal(t, oldTimerNames, newTimerNames, "timer names should be the same")

	oldHistogramNames := maps.Keys(old.Histograms())
	sort.Strings(oldHistogramNames)
	newHistogramNames := maps.Keys(new.Histograms())
	sort.Strings(newHistogramNames)

	assert.Equal(t, oldHistogramNames, newHistogramNames, "histogram names should be the same")
}

func prepareMockForTest(t *testing.T, input interface{}, expectedErr error) {
	switch mocked := input.(type) {
	case *persistence.MockConfigStoreManager:
		mocked.EXPECT().UpdateDynamicConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().FetchDynamicConfig(gomock.Any(), gomock.Any()).Return(&persistence.FetchDynamicConfigResponse{}, expectedErr).Times(2)
	case *persistence.MockDomainManager:
		mocked.EXPECT().CreateDomain(gomock.Any(), gomock.Any()).Return(&persistence.CreateDomainResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(&persistence.GetDomainResponse{}, expectedErr).Times(2)
		mocked.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().DeleteDomain(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().DeleteDomainByName(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().ListDomains(gomock.Any(), gomock.Any()).Return(&persistence.ListDomainsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{}, expectedErr).Times(2)
	case *persistence.MockHistoryManager:
		mocked.EXPECT().AppendHistoryNodes(gomock.Any(), gomock.Any()).Return(&persistence.AppendHistoryNodesResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ReadHistoryBranchResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ReadHistoryBranchByBatch(gomock.Any(), gomock.Any()).Return(&persistence.ReadHistoryBranchByBatchResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ReadRawHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ReadRawHistoryBranchResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ForkHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ForkHistoryBranchResponse{}, expectedErr).Times(2)
		mocked.EXPECT().DeleteHistoryBranch(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().GetHistoryTree(gomock.Any(), gomock.Any()).Return(&persistence.GetHistoryTreeResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetAllHistoryTreeBranches(gomock.Any(), gomock.Any()).Return(&persistence.GetAllHistoryTreeBranchesResponse{}, expectedErr).Times(2)
	case *persistence.MockQueueManager:
		mocked.EXPECT().EnqueueMessage(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().ReadMessages(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*persistence.QueueMessage{}, expectedErr).Times(2)
		mocked.EXPECT().UpdateAckLevel(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().GetAckLevels(gomock.Any()).Return(map[string]int64{}, expectedErr).Times(2)
		mocked.EXPECT().DeleteMessagesBefore(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().DeleteMessageFromDLQ(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().EnqueueMessageToDLQ(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().GetDLQAckLevels(gomock.Any()).Return(map[string]int64{}, expectedErr).Times(2)
		mocked.EXPECT().GetDLQSize(gomock.Any()).Return(int64(0), expectedErr).Times(2)
		mocked.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().ReadMessagesFromDLQ(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*persistence.QueueMessage{}, nil, expectedErr).Times(2)
		mocked.EXPECT().UpdateDLQAckLevel(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
	case *persistence.MockShardManager:
		mocked.EXPECT().GetShard(gomock.Any(), gomock.Any()).Return(&persistence.GetShardResponse{}, expectedErr).Times(2)
		mocked.EXPECT().UpdateShard(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().CreateShard(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
	case *persistence.MockTaskManager:
		mocked.EXPECT().CompleteTasksLessThan(gomock.Any(), gomock.Any()).Return(&persistence.CompleteTasksLessThanResponse{}, expectedErr).Times(2)
		mocked.EXPECT().CompleteTask(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().CreateTasks(gomock.Any(), gomock.Any()).Return(&persistence.CreateTasksResponse{}, expectedErr).Times(2)
		mocked.EXPECT().DeleteTaskList(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().GetTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetTasksResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetOrphanTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetOrphanTasksResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetTaskListSize(gomock.Any(), gomock.Any()).Return(&persistence.GetTaskListSizeResponse{}, expectedErr).Times(2)
		mocked.EXPECT().LeaseTaskList(gomock.Any(), gomock.Any()).Return(&persistence.LeaseTaskListResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListTaskList(gomock.Any(), gomock.Any()).Return(&persistence.ListTaskListResponse{}, expectedErr).Times(2)
		mocked.EXPECT().UpdateTaskList(gomock.Any(), gomock.Any()).Return(&persistence.UpdateTaskListResponse{}, expectedErr).Times(2)
	case *persistence.MockVisibilityManager:
		mocked.EXPECT().DeleteUninitializedWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.CountWorkflowExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetClosedWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.GetClosedWorkflowExecutionResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListClosedWorkflowExecutionsByStatus(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListClosedWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListClosedWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListOpenWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListOpenWorkflowExecutionsByType(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListOpenWorkflowExecutionsByWorkflowID(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().RecordWorkflowExecutionStarted(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().RecordWorkflowExecutionClosed(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().RecordWorkflowExecutionUninitialized(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().UpsertWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListWorkflowExecutionsResponse{}, expectedErr).Times(2)
	case *persistence.MockExecutionManager:
		mocked.EXPECT().CompleteTimerTask(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().CompleteTransferTask(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().CreateWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.CreateWorkflowExecutionResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.GetWorkflowExecutionResponse{}, expectedErr).Times(2)
		mocked.EXPECT().UpdateWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.UpdateWorkflowExecutionResponse{}, expectedErr).Times(2)
		mocked.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().DeleteCurrentWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().GetCurrentExecution(gomock.Any(), gomock.Any()).Return(&persistence.GetCurrentExecutionResponse{}, expectedErr).Times(2)
		mocked.EXPECT().CompleteCrossClusterTask(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().RangeCompleteCrossClusterTask(gomock.Any(), gomock.Any()).Return(&persistence.RangeCompleteCrossClusterTaskResponse{}, expectedErr).Times(2)
		mocked.EXPECT().CompleteReplicationTask(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().ConflictResolveWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.ConflictResolveWorkflowExecutionResponse{}, expectedErr).Times(2)
		mocked.EXPECT().CreateFailoverMarkerTasks(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().DeleteReplicationTaskFromDLQ(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().GetCrossClusterTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetCrossClusterTasksResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetReplicationDLQSize(gomock.Any(), gomock.Any()).Return(&persistence.GetReplicationDLQSizeResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetReplicationTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetReplicationTasksResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetReplicationTasksFromDLQ(gomock.Any(), gomock.Any()).Return(&persistence.GetReplicationTasksFromDLQResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetTimerIndexTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetTimerIndexTasksResponse{}, expectedErr).Times(2)
		mocked.EXPECT().GetTransferTasks(gomock.Any(), gomock.Any()).Return(&persistence.GetTransferTasksResponse{}, expectedErr).Times(2)
		mocked.EXPECT().IsWorkflowExecutionExists(gomock.Any(), gomock.Any()).Return(&persistence.IsWorkflowExecutionExistsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListConcreteExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListConcreteExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().ListCurrentExecutions(gomock.Any(), gomock.Any()).Return(&persistence.ListCurrentExecutionsResponse{}, expectedErr).Times(2)
		mocked.EXPECT().PutReplicationTaskToDLQ(gomock.Any(), gomock.Any()).Return(expectedErr).Times(2)
		mocked.EXPECT().RangeCompleteReplicationTask(gomock.Any(), gomock.Any()).Return(&persistence.RangeCompleteReplicationTaskResponse{}, expectedErr).Times(2)
		mocked.EXPECT().RangeCompleteTimerTask(gomock.Any(), gomock.Any()).Return(&persistence.RangeCompleteTimerTaskResponse{}, expectedErr).Times(2)
		mocked.EXPECT().RangeCompleteTransferTask(gomock.Any(), gomock.Any()).Return(&persistence.RangeCompleteTransferTaskResponse{}, expectedErr).Times(2)
		mocked.EXPECT().RangeDeleteReplicationTaskFromDLQ(gomock.Any(), gomock.Any()).Return(&persistence.RangeDeleteReplicationTaskFromDLQResponse{}, expectedErr).Times(2)
	default:
		t.Errorf("unsupported type %v", reflect.TypeOf(input))
		t.FailNow()
	}
	return
}

func setupLogsCapture() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

func runScenario(t *testing.T, oldObj, newObj any, oldLogs *observer.ObservedLogs, newLogs *observer.ObservedLogs, oldMetrics tally.TestScope, newMetrics tally.TestScope) {
	oldV := reflect.ValueOf(oldObj)
	newV := reflect.ValueOf(newObj)
	infoT := reflect.TypeOf(oldV.Interface())
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
				if method.Type.In(i).Kind() == reflect.Ptr {
					vals = append(vals, reflect.New(method.Type.In(i).Elem()))
					continue
				}
				vals = append(vals, reflect.Zero(method.Type.In(i)))
			}

			var oldRes []reflect.Value
			assert.NotPanicsf(t, func() {
				oldRes = oldV.MethodByName(method.Name).Call(vals)
			}, "method does not have tag defined")

			if len(oldRes) == 0 {
				// Empty result means that method panicked.
				return
			}

			var newRes []reflect.Value
			assert.NotPanicsf(t, func() {
				newRes = newV.MethodByName(method.Name).Call(vals)
			}, "method does not have tag defined")

			if len(newRes) == 0 {
				// Empty result means that method panicked.
				return
			}
			assertLogs(t, oldLogs, newLogs)
			assertMetrics(t, oldMetrics.Snapshot(), newMetrics.Snapshot())
		})
	}
}
