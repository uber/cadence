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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

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
				mockDB.EXPECT().IsTimeoutError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsThrottlingError(gomock.Any()).Return(false).AnyTimes()
				mockDB.EXPECT().IsDBUnavailableError(gomock.Any()).Return(false).AnyTimes()
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
				request.Mode = persistence.UpdateWorkflowMode(-1) // Invalid mode
				return store.UpdateWorkflowExecution(ctx, request)
			},
			expectedError: &types.InternalServiceError{},
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
