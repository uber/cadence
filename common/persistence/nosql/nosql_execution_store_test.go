// Copyright (c) 2023 Uber Technologies, Inc.
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

package nosql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/constants"
)

func TestNosqlExecutionStore(t *testing.T) {
	ctx := context.Background()
	shardID := 1
	testCases := []struct {
		name          string
		setupMock     func(*gomock.Controller) *nosqlExecutionStore
		testFunc      func(*nosqlExecutionStore) error
		expectedError error
	}{
		{
			name: "CreateWorkflowExecution success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					InsertWorkflowExecutionWithTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.CreateWorkflowExecution(ctx, newCreateWorkflowExecutionRequest())
				return err
			},
			expectedError: nil,
		},
		{
			name: "CreateWorkflowExecution failure - workflow already exists",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					InsertWorkflowExecutionWithTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&persistence.WorkflowExecutionAlreadyStartedError{}).Times(1)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsTimeoutError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsThrottlingError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsDBUnavailableError(gomock.Any()).Return(false).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.CreateWorkflowExecution(ctx, newCreateWorkflowExecutionRequest())
				return err
			},
			expectedError: &persistence.WorkflowExecutionAlreadyStartedError{},
		},
		{
			name: "CreateWorkflowExecution failure - shard ownership lost",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					InsertWorkflowExecutionWithTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&persistence.ShardOwnershipLostError{ShardID: shardID, Msg: "shard ownership lost"}).Times(1)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsTimeoutError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsThrottlingError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsDBUnavailableError(gomock.Any()).Return(false).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.CreateWorkflowExecution(ctx, newCreateWorkflowExecutionRequest())
				return err
			},
			expectedError: &persistence.ShardOwnershipLostError{},
		},
		{
			name: "CreateWorkflowExecution failure - current workflow condition failed",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					InsertWorkflowExecutionWithTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&persistence.CurrentWorkflowConditionFailedError{Msg: "current workflow condition failed"}).Times(1)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsTimeoutError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsThrottlingError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsDBUnavailableError(gomock.Any()).Return(false).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.CreateWorkflowExecution(ctx, newCreateWorkflowExecutionRequest())
				return err
			},
			expectedError: &persistence.CurrentWorkflowConditionFailedError{},
		},
		{
			name: "CreateWorkflowExecution failure - generic internal service error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					InsertWorkflowExecutionWithTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&types.InternalServiceError{Message: "generic internal service error"}).Times(1)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsTimeoutError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsThrottlingError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsDBUnavailableError(gomock.Any()).Return(false).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.CreateWorkflowExecution(ctx, newCreateWorkflowExecutionRequest())
				return err
			},
			expectedError: &types.InternalServiceError{},
		},
		{
			name: "GetWorkflowExecution success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectWorkflowExecution(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&nosqlplugin.WorkflowExecution{}, nil).Times(1)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.GetWorkflowExecution(ctx, newGetWorkflowExecutionRequest())
				return err
			},
			expectedError: nil,
		},
		{
			name: "GetWorkflowExecution failure - not found",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectWorkflowExecution(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, &types.EntityNotExistsError{}).Times(1)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.GetWorkflowExecution(ctx, newGetWorkflowExecutionRequest())
				return err
			},
			expectedError: &types.EntityNotExistsError{},
		},
		{
			name: "UpdateWorkflowExecution success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					UpdateWorkflowExecutionWithTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				err := store.UpdateWorkflowExecution(ctx, newUpdateWorkflowExecutionRequest())
				return err
			},
			expectedError: nil,
		},
		{
			name: "UpdateWorkflowExecution failure - invalid update mode",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				// No operation expected on the DB due to invalid mode
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				request := newUpdateWorkflowExecutionRequest()
				request.Mode = persistence.UpdateWorkflowMode(-1)
				return store.UpdateWorkflowExecution(ctx, request)
			},
			expectedError: &types.InternalServiceError{},
		},
		{
			name: "UpdateWorkflowExecution failure - condition not met",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				conditionFailure := &nosqlplugin.WorkflowOperationConditionFailure{
					UnknownConditionFailureDetails: common.StringPtr("condition not met"),
				}
				mockDB.EXPECT().
					UpdateWorkflowExecutionWithTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(conditionFailure).Times(1)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.UpdateWorkflowExecution(ctx, newUpdateWorkflowExecutionRequest())
			},
			expectedError: &persistence.ConditionFailedError{},
		},
		{
			name: "UpdateWorkflowExecution failure - operational error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					UpdateWorkflowExecutionWithTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("database is unavailable")).Times(1)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.UpdateWorkflowExecution(ctx, newUpdateWorkflowExecutionRequest())
			},
			expectedError: &types.InternalServiceError{Message: "database is unavailable"},
		},
		{
			name: "DeleteWorkflowExecution success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteWorkflowExecution(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.DeleteWorkflowExecution(ctx, &persistence.DeleteWorkflowExecutionRequest{
					DomainID:   "domainID",
					WorkflowID: "workflowID",
					RunID:      "runID",
				})
			},
			expectedError: nil,
		},
		{
			name: "DeleteWorkflowExecution failure - workflow does not exist",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteWorkflowExecution(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&types.EntityNotExistsError{Message: "workflow does not exist"})
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.DeleteWorkflowExecution(ctx, &persistence.DeleteWorkflowExecutionRequest{
					DomainID:   "domainID",
					WorkflowID: "workflowID",
					RunID:      "runID",
				})
			},
			expectedError: &types.EntityNotExistsError{Message: "workflow does not exist"},
		},
		{
			name: "DeleteCurrentWorkflowExecution success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteCurrentWorkflow(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.DeleteCurrentWorkflowExecution(ctx, &persistence.DeleteCurrentWorkflowExecutionRequest{
					DomainID:   "domainID",
					WorkflowID: "workflowID",
					RunID:      "runID",
				})
			},
			expectedError: nil,
		},
		{
			name: "DeleteCurrentWorkflowExecution failure - current workflow does not exist",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteCurrentWorkflow(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&types.EntityNotExistsError{Message: "current workflow does not exist"})
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.DeleteCurrentWorkflowExecution(ctx, &persistence.DeleteCurrentWorkflowExecutionRequest{
					DomainID:   "domainID",
					WorkflowID: "workflowID",
					RunID:      "runID",
				})
			},
			expectedError: &types.EntityNotExistsError{Message: "current workflow does not exist"},
		},
		{
			name: "ListCurrentExecutions success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectAllCurrentWorkflows(ctx, shardID, gomock.Any(), gomock.Any()).
					Return([]*persistence.CurrentWorkflowExecution{}, nil, nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.ListCurrentExecutions(ctx, &persistence.ListCurrentExecutionsRequest{})
				return err
			},
			expectedError: nil,
		},
		{
			name: "ListCurrentExecutions failure - database error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectAllCurrentWorkflows(ctx, shardID, gomock.Any(), gomock.Any()).
					Return(nil, nil, errors.New("database error"))
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.ListCurrentExecutions(ctx, &persistence.ListCurrentExecutionsRequest{})
				return err
			},
			expectedError: &types.InternalServiceError{Message: "database error"},
		},
		{
			name: "ListConcreteExecutions success - has executions",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				// Corrected return value type to match expected method signature
				executions := []*persistence.InternalListConcreteExecutionsEntity{
					{
						ExecutionInfo: &persistence.InternalWorkflowExecutionInfo{
							WorkflowID: "workflowID",
							RunID:      "runID",
						},
					},
				}
				mockDB.EXPECT().
					SelectAllWorkflowExecutions(ctx, shardID, gomock.Any(), gomock.Any()).
					Return(executions, nil, nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				resp, err := store.ListConcreteExecutions(ctx, &persistence.ListConcreteExecutionsRequest{})
				if err != nil {
					return err
				}
				if len(resp.Executions) == 0 {
					return errors.New("expected to find executions")
				}
				return nil
			},
			expectedError: nil,
		},
		{
			name: "ListConcreteExecutions failure - database error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectAllWorkflowExecutions(ctx, shardID, gomock.Any(), gomock.Any()).
					Return(nil, nil, errors.New("database error"))
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.ListConcreteExecutions(ctx, &persistence.ListConcreteExecutionsRequest{})
				return err
			},
			expectedError: &types.InternalServiceError{Message: "database error"},
		},
		{
			name: "GetTransferTasks success - has tasks",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				tasks := []*nosqlplugin.TransferTask{{TaskID: 1}}
				mockDB.EXPECT().
					SelectTransferTasksOrderByTaskID(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(tasks, nil, nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				resp, err := store.GetTransferTasks(ctx, &persistence.GetTransferTasksRequest{})
				if err != nil {
					return err
				}
				if len(resp.Tasks) == 0 {
					return errors.New("expected to find transfer tasks")
				}
				return nil
			},
			expectedError: nil,
		},
		{
			name: "GetCrossClusterTasks success - has tasks",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				tasks := []*nosqlplugin.CrossClusterTask{{TransferTask: persistence.CrossClusterTaskInfo{TaskID: 1}}}
				mockDB.EXPECT().
					SelectCrossClusterTasksOrderByTaskID(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(tasks, nil, nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				resp, err := store.GetCrossClusterTasks(ctx, &persistence.GetCrossClusterTasksRequest{})
				if err != nil {
					return err
				}
				if len(resp.Tasks) == 0 {
					return errors.New("expected to find cross cluster tasks")
				}
				return nil
			},
			expectedError: nil,
		},
		{
			name: "GetReplicationTasks success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				tasks := []*nosqlplugin.ReplicationTask{{TaskID: 1}}
				mockDB.EXPECT().
					SelectReplicationTasksOrderByTaskID(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(tasks, nil, nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				resp, err := store.GetReplicationTasks(ctx, &persistence.GetReplicationTasksRequest{})
				if err != nil {
					return err
				}
				if len(resp.Tasks) == 0 {
					return errors.New("expected to find replication tasks")
				}
				return nil
			},
			expectedError: nil,
		},
		{
			name: "GetReplicationTasks success - empty task list",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectReplicationTasksOrderByTaskID(ctx, shardID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*nosqlplugin.ReplicationTask{}, nil, nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				resp, err := store.GetReplicationTasks(ctx, &persistence.GetReplicationTasksRequest{})
				if err != nil {
					return err
				}
				if len(resp.Tasks) != 0 {
					return errors.New("expected no replication tasks")
				}
				return nil
			},
			expectedError: nil,
		},
		{
			name: "CompleteTransferTask success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteTransferTask(ctx, shardID, gomock.Any()).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteTransferTask(ctx, &persistence.CompleteTransferTaskRequest{TaskID: 1})
			},
			expectedError: nil,
		},
		{
			name: "CompleteTransferTask with zero TaskID",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)

				mockDB.EXPECT().
					DeleteTransferTask(ctx, shardID, int64(0)).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteTransferTask(ctx, &persistence.CompleteTransferTaskRequest{TaskID: 0})
			},
			expectedError: nil,
		},
		{
			name: "RangeCompleteTransferTask success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					RangeDeleteTransferTasks(ctx, shardID, gomock.Any(), gomock.Any()).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.RangeCompleteTransferTask(ctx, &persistence.RangeCompleteTransferTaskRequest{
					ExclusiveBeginTaskID: 1,
					InclusiveEndTaskID:   10,
				})
				return err
			},
			expectedError: nil,
		},
		{
			name: "RangeCompleteTransferTask with inverted TaskID range proceeds",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					RangeDeleteTransferTasks(ctx, shardID, int64(10), int64(1)).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.RangeCompleteTransferTask(ctx, &persistence.RangeCompleteTransferTaskRequest{
					ExclusiveBeginTaskID: 10,
					InclusiveEndTaskID:   1,
				})
				return err
			},
			expectedError: nil,
		},
		{
			name: "CompleteCrossClusterTask success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteCrossClusterTask(ctx, shardID, gomock.Any(), gomock.Any()).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteCrossClusterTask(ctx, &persistence.CompleteCrossClusterTaskRequest{
					TargetCluster: "targetCluster",
					TaskID:        1,
				})
			},
			expectedError: nil,
		},
		{
			name: "CompleteCrossClusterTask with missing TargetCluster proceeds",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteCrossClusterTask(ctx, shardID, "", int64(1)).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteCrossClusterTask(ctx, &persistence.CompleteCrossClusterTaskRequest{
					TargetCluster: "",
					TaskID:        1,
				})
			},
			expectedError: nil,
		},
		{
			name: "RangeCompleteCrossClusterTask success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					RangeDeleteCrossClusterTasks(ctx, shardID, "targetCluster", int64(1), int64(10)).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				resp, err := store.RangeCompleteCrossClusterTask(ctx, &persistence.RangeCompleteCrossClusterTaskRequest{
					TargetCluster:        "targetCluster",
					ExclusiveBeginTaskID: 1,
					InclusiveEndTaskID:   10,
				})
				require.Nil(t, err)
				require.Equal(t, persistence.UnknownNumRowsAffected, resp.TasksCompleted)
				return nil
			},
			expectedError: nil,
		},
		{
			name: "RangeCompleteCrossClusterTask failure - database error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				mockDB.EXPECT().
					RangeDeleteCrossClusterTasks(ctx, shardID, "targetCluster", int64(1), int64(10)).
					Return(errors.New("database error"))
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.RangeCompleteCrossClusterTask(ctx, &persistence.RangeCompleteCrossClusterTaskRequest{
					TargetCluster:        "targetCluster",
					ExclusiveBeginTaskID: 1,
					InclusiveEndTaskID:   10,
				})
				return err
			},
			expectedError: &types.InternalServiceError{Message: "database error"},
		},
		{
			name: "RangeCompleteCrossClusterTask with inverted TaskID range proceeds",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					RangeDeleteCrossClusterTasks(ctx, shardID, "targetCluster", int64(10), int64(1)).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.RangeCompleteCrossClusterTask(ctx, &persistence.RangeCompleteCrossClusterTaskRequest{
					TargetCluster:        "targetCluster",
					ExclusiveBeginTaskID: 10,
					InclusiveEndTaskID:   1,
				})
				return err
			},
			expectedError: nil,
		},
		{
			name: "RangeCompleteCrossClusterTask with no tasks in range",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					RangeDeleteCrossClusterTasks(ctx, shardID, "targetCluster", int64(100), int64(200)).
					Return(nil) // Simulate no tasks found within the range but operation succeeds
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.RangeCompleteCrossClusterTask(ctx, &persistence.RangeCompleteCrossClusterTaskRequest{
					TargetCluster:        "targetCluster",
					ExclusiveBeginTaskID: 100,
					InclusiveEndTaskID:   200,
				})
				return err
			},
			expectedError: nil, // No error expected, successful operation even if no tasks are found
		},
		{
			name: "CompleteReplicationTask success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteReplicationTask(ctx, shardID, int64(1)).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteReplicationTask(ctx, &persistence.CompleteReplicationTaskRequest{TaskID: 1})
			},
			expectedError: nil,
		},
		{
			name: "CompleteReplicationTask failure - database error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteReplicationTask(ctx, shardID, int64(1)).
					Return(errors.New("database error"))
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteReplicationTask(ctx, &persistence.CompleteReplicationTaskRequest{TaskID: 1})
			},
			expectedError: &types.InternalServiceError{Message: "database error"},
		},
		{
			name: "CompleteReplicationTask with non-existent task ID",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				mockDB.EXPECT().
					DeleteReplicationTask(ctx, shardID, int64(9999)).
					Return(&types.EntityNotExistsError{Message: "replication task does not exist"}) // Simulate task not found
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteReplicationTask(ctx, &persistence.CompleteReplicationTaskRequest{TaskID: 9999})
			},
			expectedError: &types.EntityNotExistsError{Message: "replication task does not exist"},
		},
		{
			name: "RangeCompleteReplicationTask success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					RangeDeleteReplicationTasks(ctx, shardID, int64(10)).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.RangeCompleteReplicationTask(ctx, &persistence.RangeCompleteReplicationTaskRequest{
					InclusiveEndTaskID: 10,
				})
				return err
			},
			expectedError: nil,
		},
		{
			name: "RangeCompleteReplicationTask failure - database error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					RangeDeleteReplicationTasks(ctx, shardID, int64(10)).
					Return(errors.New("database error"))
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.RangeCompleteReplicationTask(ctx, &persistence.RangeCompleteReplicationTaskRequest{
					InclusiveEndTaskID: 10,
				})
				return err
			},
			expectedError: &types.InternalServiceError{Message: "database error"},
		},
		{
			name: "RangeCompleteReplicationTask with zero InclusiveEndTaskID",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				// Expect the call with InclusiveEndTaskID of 0
				mockDB.EXPECT().
					RangeDeleteReplicationTasks(ctx, shardID, int64(0)).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.RangeCompleteReplicationTask(ctx, &persistence.RangeCompleteReplicationTaskRequest{
					InclusiveEndTaskID: 0,
				})
				return err
			},
			expectedError: nil,
		},

		{
			name: "CompleteTimerTask success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteTimerTask(ctx, shardID, int64(1), gomock.Any()).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteTimerTask(ctx, &persistence.CompleteTimerTaskRequest{
					TaskID:              1,
					VisibilityTimestamp: time.Now(),
				})
			},
			expectedError: nil,
		},
		{
			name: "CompleteTimerTask failure - database error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteTimerTask(ctx, shardID, int64(1), gomock.Any()).
					Return(errors.New("database error"))
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteTimerTask(ctx, &persistence.CompleteTimerTaskRequest{
					TaskID:              1,
					VisibilityTimestamp: time.Now(),
				})
			},
			expectedError: &types.InternalServiceError{Message: "database error"},
		},
		{
			name: "CompleteTimerTask with future VisibilityTimestamp",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					DeleteTimerTask(ctx, shardID, int64(1), gomock.Any()). // Expect the call with any timestamp
					Return(nil)                                            // Assuming successful deletion
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				return store.CompleteTimerTask(ctx, &persistence.CompleteTimerTaskRequest{
					TaskID:              1,
					VisibilityTimestamp: time.Now().Add(24 * time.Hour), // Future timestamp
				})
			},
			expectedError: nil, // Adjust based on actual behavior
		},
		{
			name: "RangeCompleteTimerTask success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					RangeDeleteTimerTasks(ctx, shardID, gomock.Any(), gomock.Any()).
					Return(nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				now := time.Now()
				// Assuming you're testing with a time range starting from 'now' and ending 1 hour later.
				beginTime := now
				endTime := now.Add(time.Hour)

				_, err := store.RangeCompleteTimerTask(ctx, &persistence.RangeCompleteTimerTaskRequest{
					InclusiveBeginTimestamp: beginTime,
					ExclusiveEndTimestamp:   endTime,
				})
				return err
			},
			expectedError: nil,
		},
		{
			name: "RangeCompleteTimerTask failure - database error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					RangeDeleteTimerTasks(ctx, shardID, gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				now := time.Now()
				// Assuming you're testing with a time range starting from 'now' and ending 1 hour later.
				beginTime := now
				endTime := now.Add(time.Hour)
				_, err := store.RangeCompleteTimerTask(ctx, &persistence.RangeCompleteTimerTaskRequest{
					InclusiveBeginTimestamp: beginTime,
					ExclusiveEndTimestamp:   endTime,
				})
				return err
			},
			expectedError: &types.InternalServiceError{Message: "database error"},
		},
		{
			name: "RangeCompleteTimerTask with inverted time range proceeds",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				// Set up an expectation for the call, even with inverted time range
				mockDB.EXPECT().
					RangeDeleteTimerTasks(ctx, shardID, gomock.Any(), gomock.Any()).
					Return(nil) // Assuming the operation proceeds regardless of time range order
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.RangeCompleteTimerTask(ctx, &persistence.RangeCompleteTimerTaskRequest{
					InclusiveBeginTimestamp: time.Now().Add(time.Hour), // Future time
					ExclusiveEndTimestamp:   time.Now(),                // Present time
				})
				return err
			},
			expectedError: nil,
		},
		{
			name: "GetTimerIndexTasks success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectTimerTasksOrderByVisibilityTime(
						ctx,
						shardID,
						10,
						gomock.Nil(),
						gomock.Any(),
						gomock.Any(),
					).Return([]*persistence.TimerTaskInfo{}, nil, nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.GetTimerIndexTasks(ctx, &persistence.GetTimerIndexTasksRequest{
					BatchSize:    10,
					MinTimestamp: time.Now().Add(-time.Hour),
					MaxTimestamp: time.Now(),
				})
				return err
			},
			expectedError: nil,
		},
		{
			name: "GetTimerIndexTasks success - empty result",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectTimerTasksOrderByVisibilityTime(ctx, shardID, 10, gomock.Nil(), gomock.Any(), gomock.Any()).
					Return([]*persistence.TimerTaskInfo{}, []byte{}, nil) // Return an empty list
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				resp, err := store.GetTimerIndexTasks(ctx, &persistence.GetTimerIndexTasksRequest{
					BatchSize:    10,
					MinTimestamp: time.Now().Add(-time.Hour),
					MaxTimestamp: time.Now(),
				})
				if err != nil {
					return err
				}
				if len(resp.Timers) != 0 {
					return errors.New("expected empty result set for timers")
				}
				return nil
			},
			expectedError: nil,
		},
		{
			name: "PutReplicationTaskToDLQ success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				replicationTaskInfo := newInternalReplicationTaskInfo()

				mockDB.EXPECT().
					InsertReplicationDLQTask(ctx, shardID, "sourceCluster", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ int, _ string, taskInfo persistence.InternalReplicationTaskInfo) error {
						require.Equal(t, replicationTaskInfo, taskInfo)
						return nil
					})

				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				taskInfo := newInternalReplicationTaskInfo()
				return store.PutReplicationTaskToDLQ(ctx, &persistence.InternalPutReplicationTaskToDLQRequest{
					SourceClusterName: "sourceCluster",
					TaskInfo:          &taskInfo,
				})
			},
			expectedError: nil,
		},
		{
			name: "GetTimerIndexTasks failure - database error",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(true).AnyTimes()
				mockDB.EXPECT().
					SelectTimerTasksOrderByVisibilityTime(ctx, shardID, 10, gomock.Nil(), gomock.Any(), gomock.Any()).
					Return(nil, nil, errors.New("database error"))
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.GetTimerIndexTasks(ctx, &persistence.GetTimerIndexTasksRequest{
					BatchSize:    10,
					MinTimestamp: time.Now().Add(-time.Hour),
					MaxTimestamp: time.Now(),
				})
				return err
			},
			expectedError: &types.InternalServiceError{Message: "database error"},
		},
		{
			name: "GetReplicationTasksFromDLQ success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)

				nextPageToken := []byte("next-page-token")
				replicationTasks := []*persistence.InternalReplicationTaskInfo{}
				mockDB.EXPECT().
					SelectReplicationDLQTasksOrderByTaskID(
						ctx,
						shardID,
						"sourceCluster",
						10,
						gomock.Any(),
						int64(0),
						int64(100),
					).Return(replicationTasks, nextPageToken, nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				initialNextPageToken := []byte{}
				_, err := store.GetReplicationTasksFromDLQ(ctx, &persistence.GetReplicationTasksFromDLQRequest{
					SourceClusterName: "sourceCluster",
					GetReplicationTasksRequest: persistence.GetReplicationTasksRequest{
						BatchSize:     10,
						NextPageToken: initialNextPageToken,
						ReadLevel:     0,
						MaxReadLevel:  100,
					},
				})

				return err
			},
			expectedError: nil,
		},
		{
			name: "GetReplicationTasksFromDLQ failure - invalid read levels",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				return newTestNosqlExecutionStore(nosqlplugin.NewMockDB(ctrl), log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.GetReplicationTasksFromDLQ(ctx, &persistence.GetReplicationTasksFromDLQRequest{
					SourceClusterName: "sourceCluster",
					GetReplicationTasksRequest: persistence.GetReplicationTasksRequest{
						ReadLevel:     100,
						MaxReadLevel:  50,
						BatchSize:     10,
						NextPageToken: []byte{},
					},
				})
				return err
			},
			expectedError: &types.BadRequestError{Message: "ReadLevel cannot be higher than MaxReadLevel"},
		},
		{
			name: "GetReplicationDLQSize success",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectReplicationDLQTasksCount(ctx, shardID, "sourceCluster").
					Return(int64(42), nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				resp, err := store.GetReplicationDLQSize(ctx, &persistence.GetReplicationDLQSizeRequest{
					SourceClusterName: "sourceCluster",
				})
				if err != nil {
					return err
				}
				if resp.Size != 42 {
					return errors.New("unexpected DLQ size")
				}
				return nil
			},
			expectedError: nil,
		},
		{
			name: "GetReplicationDLQSize failure - invalid source cluster name",
			setupMock: func(ctrl *gomock.Controller) *nosqlExecutionStore {
				mockDB := nosqlplugin.NewMockDB(ctrl)
				mockDB.EXPECT().
					SelectReplicationDLQTasksCount(ctx, shardID, "").
					Return(int64(0), nil)
				return newTestNosqlExecutionStore(mockDB, log.NewNoop())
			},
			testFunc: func(store *nosqlExecutionStore) error {
				_, err := store.GetReplicationDLQSize(ctx, &persistence.GetReplicationDLQSizeRequest{
					SourceClusterName: "",
				})
				return err
			},
			expectedError: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := tc.setupMock(ctrl)
			err := tc.testFunc(store)

			if tc.expectedError != nil {
				var expectedErrType error
				require.ErrorAs(t, err, &expectedErrType, "Expected error type does not match.")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func newCreateWorkflowExecutionRequest() *persistence.InternalCreateWorkflowExecutionRequest {
	return &persistence.InternalCreateWorkflowExecutionRequest{
		RangeID:                  123,
		Mode:                     persistence.CreateWorkflowModeBrandNew,
		PreviousRunID:            "previous-run-id",
		PreviousLastWriteVersion: 456,
		NewWorkflowSnapshot:      getNewWorkflowSnapshot(),
	}
}

func newGetWorkflowExecutionRequest() *persistence.InternalGetWorkflowExecutionRequest {
	return &persistence.InternalGetWorkflowExecutionRequest{
		DomainID: constants.TestDomainID,
		Execution: types.WorkflowExecution{
			WorkflowID: constants.TestWorkflowID,
			RunID:      constants.TestRunID,
		},
	}
}

func newUpdateWorkflowExecutionRequest() *persistence.InternalUpdateWorkflowExecutionRequest {
	return &persistence.InternalUpdateWorkflowExecutionRequest{
		RangeID: 123,
		UpdateWorkflowMutation: persistence.InternalWorkflowMutation{
			ExecutionInfo: &persistence.InternalWorkflowExecutionInfo{
				DomainID:    constants.TestDomainID,
				WorkflowID:  constants.TestWorkflowID,
				RunID:       constants.TestRunID,
				State:       persistence.WorkflowStateCreated,
				CloseStatus: persistence.WorkflowCloseStatusNone,
			},
		},
	}
}

func getNewWorkflowSnapshot() persistence.InternalWorkflowSnapshot {
	return persistence.InternalWorkflowSnapshot{
		VersionHistories: &persistence.DataBlob{},
		ExecutionInfo: &persistence.InternalWorkflowExecutionInfo{
			DomainID:   constants.TestDomainID,
			WorkflowID: constants.TestWorkflowID,
			RunID:      constants.TestRunID,
		},
	}
}
func newTestNosqlExecutionStore(db nosqlplugin.DB, logger log.Logger) *nosqlExecutionStore {
	return &nosqlExecutionStore{
		shardID:    1,
		nosqlStore: nosqlStore{logger: logger, db: db},
	}
}

func newInternalReplicationTaskInfo() persistence.InternalReplicationTaskInfo {
	var fixedCreationTime = time.Date(2024, time.April, 3, 14, 35, 44, 0, time.UTC)
	return persistence.InternalReplicationTaskInfo{
		DomainID:          "testDomainID",
		WorkflowID:        "testWorkflowID",
		RunID:             "testRunID",
		TaskID:            123,
		TaskType:          persistence.ReplicationTaskTypeHistory,
		FirstEventID:      1,
		NextEventID:       2,
		Version:           1,
		ScheduledID:       3,
		BranchToken:       []byte("branchToken"),
		NewRunBranchToken: []byte("newRunBranchToken"),
		CreationTime:      fixedCreationTime,
	}
}
