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
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/serialization"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
	"github.com/uber/cadence/common/types"
)

func TestDeleteCurrentWorkflowExecution(t *testing.T) {
	shardID := int64(100)
	testCases := []struct {
		name      string
		req       *persistence.DeleteCurrentWorkflowExecutionRequest
		mockSetup func(*sqlplugin.MockDB)
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.DeleteCurrentWorkflowExecutionRequest{
				DomainID:   "abdcea69-61d5-44c3-9d55-afe23505a542",
				WorkflowID: "aaaa",
				RunID:      "fd65967f-777d-45de-8dee-be49dfda6716",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().DeleteFromCurrentExecutions(gomock.Any(), &sqlplugin.CurrentExecutionsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "aaaa",
					RunID:      serialization.MustParseUUID("fd65967f-777d-45de-8dee-be49dfda6716"),
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.DeleteCurrentWorkflowExecutionRequest{
				DomainID:   "abdcea69-61d5-44c3-9d55-afe23505a542",
				WorkflowID: "aaaa",
				RunID:      "fd65967f-777d-45de-8dee-be49dfda6716",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().DeleteFromCurrentExecutions(gomock.Any(), &sqlplugin.CurrentExecutionsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "aaaa",
					RunID:      serialization.MustParseUUID("fd65967f-777d-45de-8dee-be49dfda6716"),
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			err = store.DeleteCurrentWorkflowExecution(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
			}
		})
	}
}

func TestGetCurrentExecution(t *testing.T) {
	shardID := int64(100)
	testCases := []struct {
		name      string
		req       *persistence.GetCurrentExecutionRequest
		mockSetup func(*sqlplugin.MockDB)
		want      *persistence.GetCurrentExecutionResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.GetCurrentExecutionRequest{
				DomainID:   "abdcea69-61d5-44c3-9d55-afe23505a542",
				WorkflowID: "aaaa",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().SelectFromCurrentExecutions(gomock.Any(), &sqlplugin.CurrentExecutionsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "aaaa",
				}).Return(&sqlplugin.CurrentExecutionsRow{
					ShardID:          shardID,
					DomainID:         serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID:       "aaaa",
					RunID:            serialization.MustParseUUID("fd65967f-777d-45de-8dee-be49dfda6716"),
					CreateRequestID:  "create",
					State:            2,
					CloseStatus:      3,
					LastWriteVersion: 9,
				}, nil)
			},
			want: &persistence.GetCurrentExecutionResponse{
				StartRequestID:   "create",
				RunID:            "fd65967f-777d-45de-8dee-be49dfda6716",
				State:            2,
				CloseStatus:      3,
				LastWriteVersion: 9,
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.GetCurrentExecutionRequest{
				DomainID:   "abdcea69-61d5-44c3-9d55-afe23505a542",
				WorkflowID: "aaaa",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().SelectFromCurrentExecutions(gomock.Any(), &sqlplugin.CurrentExecutionsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "aaaa",
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			got, err := store.GetCurrentExecution(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestGetTransferTasks(t *testing.T) {
	shardID := 1
	testCases := []struct {
		name      string
		req       *persistence.GetTransferTasksRequest
		mockSetup func(*sqlplugin.MockDB, *serialization.MockParser)
		want      *persistence.GetTransferTasksResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.GetTransferTasksRequest{
				ReadLevel:     1,
				MaxReadLevel:  99,
				BatchSize:     1,
				NextPageToken: serializePageToken(11),
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromTransferTasks(gomock.Any(), &sqlplugin.TransferTasksFilter{
					ShardID:   shardID,
					MinTaskID: 11,
					MaxTaskID: 99,
					PageSize:  1,
				}).Return([]sqlplugin.TransferTasksRow{
					{
						ShardID:      shardID,
						TaskID:       12,
						Data:         []byte(`transfer`),
						DataEncoding: "transfer",
					},
				}, nil)
				mockParser.EXPECT().TransferTaskInfoFromBlob([]byte(`transfer`), "transfer").Return(&serialization.TransferTaskInfo{
					DomainID:                serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID:              "test",
					RunID:                   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a54a"),
					VisibilityTimestamp:     time.Unix(1, 1),
					TargetDomainID:          serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
					TargetDomainIDs:         []serialization.UUID{serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a54a")},
					TargetWorkflowID:        "xyz",
					TargetRunID:             serialization.MustParseUUID("dbdcea69-61d5-44c3-9d55-afe23505a54a"),
					TargetChildWorkflowOnly: true,
					TaskList:                "tl",
					TaskType:                1,
					ScheduleID:              19,
					Version:                 202,
				}, nil)
			},
			want: &persistence.GetTransferTasksResponse{
				Tasks: []*persistence.TransferTaskInfo{
					{
						TaskID:                  12,
						DomainID:                "abdcea69-61d5-44c3-9d55-afe23505a542",
						WorkflowID:              "test",
						RunID:                   "abdcea69-61d5-44c3-9d55-afe23505a54a",
						VisibilityTimestamp:     time.Unix(1, 1),
						TargetDomainID:          "bbdcea69-61d5-44c3-9d55-afe23505a542",
						TargetWorkflowID:        "xyz",
						TargetDomainIDs:         map[string]struct{}{"bbdcea69-61d5-44c3-9d55-afe23505a54a": struct{}{}},
						TargetRunID:             "dbdcea69-61d5-44c3-9d55-afe23505a54a",
						TargetChildWorkflowOnly: true,
						TaskList:                "tl",
						TaskType:                1,
						ScheduleID:              19,
						Version:                 202,
					},
				},
				NextPageToken: serializePageToken(12),
			},
			wantErr: false,
		},
		{
			name: "Error case - failed to get from database",
			req: &persistence.GetTransferTasksRequest{
				ReadLevel:     1,
				MaxReadLevel:  99,
				BatchSize:     1,
				NextPageToken: serializePageToken(11),
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				err := errors.New("some error")
				mockDB.EXPECT().SelectFromTransferTasks(gomock.Any(), &sqlplugin.TransferTasksFilter{
					ShardID:   shardID,
					MinTaskID: 11,
					MaxTaskID: 99,
					PageSize:  1,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
		{
			name: "Error case - failed to decode data",
			req: &persistence.GetTransferTasksRequest{
				ReadLevel:     1,
				MaxReadLevel:  99,
				BatchSize:     1,
				NextPageToken: serializePageToken(11),
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromTransferTasks(gomock.Any(), &sqlplugin.TransferTasksFilter{
					ShardID:   shardID,
					MinTaskID: 11,
					MaxTaskID: 99,
					PageSize:  1,
				}).Return([]sqlplugin.TransferTasksRow{
					{
						ShardID:      shardID,
						TaskID:       12,
						Data:         []byte(`transfer`),
						DataEncoding: "transfer",
					},
				}, nil)
				mockParser.EXPECT().TransferTaskInfoFromBlob([]byte(`transfer`), "transfer").Return(nil, errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			mockParser := serialization.NewMockParser(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), mockParser, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB, mockParser)

			got, err := store.GetTransferTasks(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestCompleteTransferTask(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.CompleteTransferTaskRequest
		mockSetup func(*sqlplugin.MockDB)
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.CompleteTransferTaskRequest{
				TaskID: 123,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().DeleteFromTransferTasks(gomock.Any(), &sqlplugin.TransferTasksFilter{
					ShardID: shardID,
					TaskID:  123,
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.CompleteTransferTaskRequest{
				TaskID: 123,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().DeleteFromTransferTasks(gomock.Any(), &sqlplugin.TransferTasksFilter{
					ShardID: shardID,
					TaskID:  123,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			err = store.CompleteTransferTask(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
			}
		})
	}
}

func TestRangeCompleteTransferTask(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.RangeCompleteTransferTaskRequest
		mockSetup func(*sqlplugin.MockDB)
		want      *persistence.RangeCompleteTransferTaskResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.RangeCompleteTransferTaskRequest{
				ExclusiveBeginTaskID: 123,
				InclusiveEndTaskID:   345,
				PageSize:             1000,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().RangeDeleteFromTransferTasks(gomock.Any(), &sqlplugin.TransferTasksFilter{
					ShardID:   shardID,
					MinTaskID: 123,
					MaxTaskID: 345,
					PageSize:  1000,
				}).Return(&sqlResult{rowsAffected: 1000}, nil)
			},
			want: &persistence.RangeCompleteTransferTaskResponse{
				TasksCompleted: 1000,
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.RangeCompleteTransferTaskRequest{
				ExclusiveBeginTaskID: 123,
				InclusiveEndTaskID:   345,
				PageSize:             1000,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().RangeDeleteFromTransferTasks(gomock.Any(), &sqlplugin.TransferTasksFilter{
					ShardID:   shardID,
					MinTaskID: 123,
					MaxTaskID: 345,
					PageSize:  1000,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			got, err := store.RangeCompleteTransferTask(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestGetCrossClusterTasks(t *testing.T) {
	shardID := 1
	testCases := []struct {
		name      string
		req       *persistence.GetCrossClusterTasksRequest
		mockSetup func(*sqlplugin.MockDB, *serialization.MockParser)
		want      *persistence.GetCrossClusterTasksResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.GetCrossClusterTasksRequest{
				NextPageToken: serializePageToken(100),
				TargetCluster: "target",
				MaxReadLevel:  199,
				BatchSize:     10,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromCrossClusterTasks(gomock.Any(), &sqlplugin.CrossClusterTasksFilter{
					TargetCluster: "target",
					ShardID:       shardID,
					MinTaskID:     100,
					MaxTaskID:     199,
					PageSize:      10,
				}).Return([]sqlplugin.CrossClusterTasksRow{
					{
						TargetCluster: "target",
						ShardID:       shardID,
						TaskID:        101,
						Data:          []byte(`cross`),
						DataEncoding:  "cross",
					},
				}, nil)
				mockParser.EXPECT().CrossClusterTaskInfoFromBlob([]byte(`cross`), "cross").Return(&serialization.CrossClusterTaskInfo{
					DomainID:                serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID:              "test",
					RunID:                   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a54a"),
					VisibilityTimestamp:     time.Unix(1, 1),
					TargetDomainID:          serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
					TargetDomainIDs:         []serialization.UUID{serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a54a")},
					TargetWorkflowID:        "xyz",
					TargetRunID:             serialization.MustParseUUID("dbdcea69-61d5-44c3-9d55-afe23505a54a"),
					TargetChildWorkflowOnly: true,
					TaskList:                "tl",
					TaskType:                1,
					ScheduleID:              19,
					Version:                 202,
				}, nil)
			},
			want: &persistence.GetCrossClusterTasksResponse{
				Tasks: []*persistence.CrossClusterTaskInfo{
					{
						TaskID:                  101,
						DomainID:                "abdcea69-61d5-44c3-9d55-afe23505a542",
						WorkflowID:              "test",
						RunID:                   "abdcea69-61d5-44c3-9d55-afe23505a54a",
						VisibilityTimestamp:     time.Unix(1, 1),
						TargetDomainID:          "bbdcea69-61d5-44c3-9d55-afe23505a542",
						TargetWorkflowID:        "xyz",
						TargetDomainIDs:         map[string]struct{}{"bbdcea69-61d5-44c3-9d55-afe23505a54a": struct{}{}},
						TargetRunID:             "dbdcea69-61d5-44c3-9d55-afe23505a54a",
						TargetChildWorkflowOnly: true,
						TaskList:                "tl",
						TaskType:                1,
						ScheduleID:              19,
						Version:                 202,
					},
				},
				NextPageToken: serializePageToken(101),
			},
		},
		{
			name: "Error case - failed to get from database",
			req: &persistence.GetCrossClusterTasksRequest{
				ReadLevel:     1,
				MaxReadLevel:  99,
				BatchSize:     1,
				NextPageToken: serializePageToken(11),
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				err := errors.New("some error")
				mockDB.EXPECT().SelectFromCrossClusterTasks(gomock.Any(), gomock.Any()).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
		{
			name: "Error case - failed to decode data",
			req: &persistence.GetCrossClusterTasksRequest{
				ReadLevel:     1,
				MaxReadLevel:  99,
				BatchSize:     1,
				NextPageToken: serializePageToken(11),
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromCrossClusterTasks(gomock.Any(), gomock.Any()).Return([]sqlplugin.CrossClusterTasksRow{
					{
						ShardID:      shardID,
						TaskID:       12,
						Data:         []byte(`cross`),
						DataEncoding: "cross",
					},
				}, nil)
				mockParser.EXPECT().CrossClusterTaskInfoFromBlob([]byte(`cross`), "cross").Return(nil, errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			mockParser := serialization.NewMockParser(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), mockParser, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB, mockParser)

			got, err := store.GetCrossClusterTasks(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestCompleteCrossClusterTask(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.CompleteCrossClusterTaskRequest
		mockSetup func(*sqlplugin.MockDB)
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.CompleteCrossClusterTaskRequest{
				TaskID:        123,
				TargetCluster: "test",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().DeleteFromCrossClusterTasks(gomock.Any(), &sqlplugin.CrossClusterTasksFilter{
					TargetCluster: "test",
					ShardID:       shardID,
					TaskID:        123,
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.CompleteCrossClusterTaskRequest{
				TaskID:        123,
				TargetCluster: "test",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().DeleteFromCrossClusterTasks(gomock.Any(), &sqlplugin.CrossClusterTasksFilter{
					TargetCluster: "test",
					ShardID:       shardID,
					TaskID:        123,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			err = store.CompleteCrossClusterTask(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
			}
		})
	}
}

func TestRangeCompleteCrossClusterTask(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.RangeCompleteCrossClusterTaskRequest
		mockSetup func(*sqlplugin.MockDB)
		want      *persistence.RangeCompleteCrossClusterTaskResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.RangeCompleteCrossClusterTaskRequest{
				ExclusiveBeginTaskID: 123,
				InclusiveEndTaskID:   345,
				TargetCluster:        "test",
				PageSize:             10,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().RangeDeleteFromCrossClusterTasks(gomock.Any(), &sqlplugin.CrossClusterTasksFilter{
					TargetCluster: "test",
					ShardID:       shardID,
					MinTaskID:     123,
					MaxTaskID:     345,
					PageSize:      10,
				}).Return(&sqlResult{rowsAffected: 10}, nil)
			},
			want: &persistence.RangeCompleteCrossClusterTaskResponse{
				TasksCompleted: 10,
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.RangeCompleteCrossClusterTaskRequest{
				ExclusiveBeginTaskID: 123,
				InclusiveEndTaskID:   345,
				TargetCluster:        "test",
				PageSize:             10,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().RangeDeleteFromCrossClusterTasks(gomock.Any(), &sqlplugin.CrossClusterTasksFilter{
					TargetCluster: "test",
					ShardID:       shardID,
					MinTaskID:     123,
					MaxTaskID:     345,
					PageSize:      10,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			got, err := store.RangeCompleteCrossClusterTask(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestGetReplicationTasks(t *testing.T) {
	shardID := 0
	testCases := []struct {
		name      string
		req       *persistence.GetReplicationTasksRequest
		mockSetup func(*sqlplugin.MockDB, *serialization.MockParser)
		want      *persistence.InternalGetReplicationTasksResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.GetReplicationTasksRequest{
				NextPageToken: serializePageToken(100),
				MaxReadLevel:  199,
				BatchSize:     1000,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromReplicationTasks(gomock.Any(), &sqlplugin.ReplicationTasksFilter{
					ShardID:   shardID,
					MinTaskID: 100,
					MaxTaskID: 1100,
					PageSize:  1000,
				}).Return([]sqlplugin.ReplicationTasksRow{
					{
						ShardID:      shardID,
						TaskID:       101,
						Data:         []byte(`replication`),
						DataEncoding: "replication",
					},
				}, nil)
				mockParser.EXPECT().ReplicationTaskInfoFromBlob([]byte(`replication`), "replication").Return(&serialization.ReplicationTaskInfo{
					DomainID:          serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID:        "test",
					RunID:             serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a54a"),
					TaskType:          1,
					FirstEventID:      10,
					NextEventID:       101,
					Version:           202,
					ScheduledID:       19,
					BranchToken:       []byte(`bt`),
					NewRunBranchToken: []byte(`nbt`),
					CreationTimestamp: time.Unix(1, 1),
				}, nil)
			},
			want: &persistence.InternalGetReplicationTasksResponse{
				Tasks: []*persistence.InternalReplicationTaskInfo{
					{
						TaskID:            101,
						DomainID:          "abdcea69-61d5-44c3-9d55-afe23505a542",
						WorkflowID:        "test",
						RunID:             "abdcea69-61d5-44c3-9d55-afe23505a54a",
						TaskType:          1,
						FirstEventID:      10,
						NextEventID:       101,
						Version:           202,
						ScheduledID:       19,
						BranchToken:       []byte(`bt`),
						NewRunBranchToken: []byte(`nbt`),
						CreationTime:      time.Unix(1, 1),
					},
				},
				NextPageToken: serializePageToken(101),
			},
			wantErr: false,
		},
		{
			name: "Error case - failed to load from database",
			req: &persistence.GetReplicationTasksRequest{
				NextPageToken: serializePageToken(100),
				MaxReadLevel:  199,
				BatchSize:     1000,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				err := errors.New("some error")
				mockDB.EXPECT().SelectFromReplicationTasks(gomock.Any(), &sqlplugin.ReplicationTasksFilter{
					ShardID:   shardID,
					MinTaskID: 100,
					MaxTaskID: 1100,
					PageSize:  1000,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
		{
			name: "Error case - failed to decode data",
			req: &persistence.GetReplicationTasksRequest{
				NextPageToken: serializePageToken(100),
				MaxReadLevel:  199,
				BatchSize:     1000,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromReplicationTasks(gomock.Any(), &sqlplugin.ReplicationTasksFilter{
					ShardID:   shardID,
					MinTaskID: 100,
					MaxTaskID: 1100,
					PageSize:  1000,
				}).Return([]sqlplugin.ReplicationTasksRow{
					{
						ShardID:      shardID,
						TaskID:       101,
						Data:         []byte(`replication`),
						DataEncoding: "replication",
					},
				}, nil)
				mockParser.EXPECT().ReplicationTaskInfoFromBlob([]byte(`replication`), "replication").Return(nil, errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			mockParser := serialization.NewMockParser(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), mockParser, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB, mockParser)

			got, err := store.GetReplicationTasks(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestCompleteReplicationTask(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.CompleteReplicationTaskRequest
		mockSetup func(*sqlplugin.MockDB)
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.CompleteReplicationTaskRequest{
				TaskID: 123,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().DeleteFromReplicationTasks(gomock.Any(), &sqlplugin.ReplicationTasksFilter{
					ShardID: shardID,
					TaskID:  123,
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.CompleteReplicationTaskRequest{
				TaskID: 123,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().DeleteFromReplicationTasks(gomock.Any(), &sqlplugin.ReplicationTasksFilter{
					ShardID: shardID,
					TaskID:  123,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			err = store.CompleteReplicationTask(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
			}
		})
	}
}

func TestRangeCompleteReplicationTask(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.RangeCompleteReplicationTaskRequest
		mockSetup func(*sqlplugin.MockDB)
		want      *persistence.RangeCompleteReplicationTaskResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.RangeCompleteReplicationTaskRequest{
				InclusiveEndTaskID: 345,
				PageSize:           10,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().RangeDeleteFromReplicationTasks(gomock.Any(), &sqlplugin.ReplicationTasksFilter{
					ShardID:            shardID,
					InclusiveEndTaskID: 345,
					PageSize:           10,
				}).Return(&sqlResult{rowsAffected: 10}, nil)
			},
			want: &persistence.RangeCompleteReplicationTaskResponse{
				TasksCompleted: 10,
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.RangeCompleteReplicationTaskRequest{
				InclusiveEndTaskID: 345,
				PageSize:           10,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().RangeDeleteFromReplicationTasks(gomock.Any(), &sqlplugin.ReplicationTasksFilter{
					ShardID:            shardID,
					InclusiveEndTaskID: 345,
					PageSize:           10,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			got, err := store.RangeCompleteReplicationTask(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestGetReplicationTasksFromDLQ(t *testing.T) {
	shardID := 0
	testCases := []struct {
		name      string
		req       *persistence.GetReplicationTasksFromDLQRequest
		mockSetup func(*sqlplugin.MockDB, *serialization.MockParser)
		want      *persistence.InternalGetReplicationTasksFromDLQResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.GetReplicationTasksFromDLQRequest{
				SourceClusterName: "source",
				GetReplicationTasksRequest: persistence.GetReplicationTasksRequest{
					NextPageToken: serializePageToken(100),
					MaxReadLevel:  199,
					BatchSize:     1000,
				},
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromReplicationTasksDLQ(gomock.Any(), &sqlplugin.ReplicationTasksDLQFilter{
					ReplicationTasksFilter: sqlplugin.ReplicationTasksFilter{
						ShardID:   shardID,
						MinTaskID: 100,
						MaxTaskID: 1100,
						PageSize:  1000,
					},
					SourceClusterName: "source",
				}).Return([]sqlplugin.ReplicationTasksRow{
					{
						ShardID:      shardID,
						TaskID:       101,
						Data:         []byte(`replication`),
						DataEncoding: "replication",
					},
				}, nil)
				mockParser.EXPECT().ReplicationTaskInfoFromBlob([]byte(`replication`), "replication").Return(&serialization.ReplicationTaskInfo{
					DomainID:          serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID:        "test",
					RunID:             serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a54a"),
					TaskType:          1,
					FirstEventID:      10,
					NextEventID:       101,
					Version:           202,
					ScheduledID:       19,
					BranchToken:       []byte(`bt`),
					NewRunBranchToken: []byte(`nbt`),
					CreationTimestamp: time.Unix(1, 1),
				}, nil)
			},
			want: &persistence.InternalGetReplicationTasksFromDLQResponse{
				Tasks: []*persistence.InternalReplicationTaskInfo{
					{
						TaskID:            101,
						DomainID:          "abdcea69-61d5-44c3-9d55-afe23505a542",
						WorkflowID:        "test",
						RunID:             "abdcea69-61d5-44c3-9d55-afe23505a54a",
						TaskType:          1,
						FirstEventID:      10,
						NextEventID:       101,
						Version:           202,
						ScheduledID:       19,
						BranchToken:       []byte(`bt`),
						NewRunBranchToken: []byte(`nbt`),
						CreationTime:      time.Unix(1, 1),
					},
				},
				NextPageToken: serializePageToken(101),
			},
			wantErr: false,
		},
		{
			name: "Error case - failed to load from database",
			req: &persistence.GetReplicationTasksFromDLQRequest{
				SourceClusterName: "source",
				GetReplicationTasksRequest: persistence.GetReplicationTasksRequest{
					NextPageToken: serializePageToken(100),
					MaxReadLevel:  199,
					BatchSize:     1000,
				},
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				err := errors.New("some error")
				mockDB.EXPECT().SelectFromReplicationTasksDLQ(gomock.Any(), &sqlplugin.ReplicationTasksDLQFilter{
					ReplicationTasksFilter: sqlplugin.ReplicationTasksFilter{
						ShardID:   shardID,
						MinTaskID: 100,
						MaxTaskID: 1100,
						PageSize:  1000,
					},
					SourceClusterName: "source",
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
		{
			name: "Error case - failed to decode data",
			req: &persistence.GetReplicationTasksFromDLQRequest{
				SourceClusterName: "source",
				GetReplicationTasksRequest: persistence.GetReplicationTasksRequest{
					NextPageToken: serializePageToken(100),
					MaxReadLevel:  199,
					BatchSize:     1000,
				},
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromReplicationTasksDLQ(gomock.Any(), &sqlplugin.ReplicationTasksDLQFilter{
					ReplicationTasksFilter: sqlplugin.ReplicationTasksFilter{
						ShardID:   shardID,
						MinTaskID: 100,
						MaxTaskID: 1100,
						PageSize:  1000,
					},
					SourceClusterName: "source",
				}).Return([]sqlplugin.ReplicationTasksRow{
					{
						ShardID:      shardID,
						TaskID:       101,
						Data:         []byte(`replication`),
						DataEncoding: "replication",
					},
				}, nil)
				mockParser.EXPECT().ReplicationTaskInfoFromBlob([]byte(`replication`), "replication").Return(nil, errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			mockParser := serialization.NewMockParser(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), mockParser, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB, mockParser)

			got, err := store.GetReplicationTasksFromDLQ(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestGetReplicationDLQSize(t *testing.T) {
	shardID := 9
	testCases := []struct {
		name      string
		req       *persistence.GetReplicationDLQSizeRequest
		mockSetup func(*sqlplugin.MockDB)
		want      *persistence.GetReplicationDLQSizeResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.GetReplicationDLQSizeRequest{
				SourceClusterName: "source",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().SelectFromReplicationDLQ(gomock.Any(), &sqlplugin.ReplicationTaskDLQFilter{
					SourceClusterName: "source",
					ShardID:           shardID,
				}).Return(int64(1), nil)
			},
			want: &persistence.GetReplicationDLQSizeResponse{
				Size: 1,
			},
			wantErr: false,
		},
		{
			name: "Success case - no row",
			req: &persistence.GetReplicationDLQSizeRequest{
				SourceClusterName: "source",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().SelectFromReplicationDLQ(gomock.Any(), &sqlplugin.ReplicationTaskDLQFilter{
					SourceClusterName: "source",
					ShardID:           shardID,
				}).Return(int64(0), sql.ErrNoRows)
			},
			want: &persistence.GetReplicationDLQSizeResponse{
				Size: 0,
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.GetReplicationDLQSizeRequest{
				SourceClusterName: "source",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().SelectFromReplicationDLQ(gomock.Any(), &sqlplugin.ReplicationTaskDLQFilter{
					SourceClusterName: "source",
					ShardID:           shardID,
				}).Return(int64(0), err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			got, err := store.GetReplicationDLQSize(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestDeleteReplicationTaskFromDLQ(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.DeleteReplicationTaskFromDLQRequest
		mockSetup func(*sqlplugin.MockDB)
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.DeleteReplicationTaskFromDLQRequest{
				TaskID:            123,
				SourceClusterName: "source",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().DeleteMessageFromReplicationTasksDLQ(gomock.Any(), &sqlplugin.ReplicationTasksDLQFilter{
					ReplicationTasksFilter: sqlplugin.ReplicationTasksFilter{
						ShardID: shardID,
						TaskID:  123,
					},
					SourceClusterName: "source",
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.DeleteReplicationTaskFromDLQRequest{
				TaskID:            123,
				SourceClusterName: "source",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().DeleteMessageFromReplicationTasksDLQ(gomock.Any(), &sqlplugin.ReplicationTasksDLQFilter{
					ReplicationTasksFilter: sqlplugin.ReplicationTasksFilter{
						ShardID: shardID,
						TaskID:  123,
					},
					SourceClusterName: "source",
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			err = store.DeleteReplicationTaskFromDLQ(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
			}
		})
	}
}

func TestRangeDeleteReplicationTaskFromDLQ(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.RangeDeleteReplicationTaskFromDLQRequest
		mockSetup func(*sqlplugin.MockDB)
		want      *persistence.RangeDeleteReplicationTaskFromDLQResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.RangeDeleteReplicationTaskFromDLQRequest{
				ExclusiveBeginTaskID: 123,
				InclusiveEndTaskID:   345,
				PageSize:             10,
				SourceClusterName:    "source",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().RangeDeleteMessageFromReplicationTasksDLQ(gomock.Any(), &sqlplugin.ReplicationTasksDLQFilter{
					ReplicationTasksFilter: sqlplugin.ReplicationTasksFilter{
						ShardID:            shardID,
						TaskID:             123,
						InclusiveEndTaskID: 345,
						PageSize:           10,
					},
					SourceClusterName: "source",
				}).Return(&sqlResult{rowsAffected: 10}, nil)
			},
			want: &persistence.RangeDeleteReplicationTaskFromDLQResponse{
				TasksCompleted: 10,
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.RangeDeleteReplicationTaskFromDLQRequest{
				ExclusiveBeginTaskID: 123,
				InclusiveEndTaskID:   345,
				PageSize:             10,
				SourceClusterName:    "source",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().RangeDeleteMessageFromReplicationTasksDLQ(gomock.Any(), &sqlplugin.ReplicationTasksDLQFilter{
					ReplicationTasksFilter: sqlplugin.ReplicationTasksFilter{
						ShardID:            shardID,
						TaskID:             123,
						PageSize:           10,
						InclusiveEndTaskID: 345,
					},
					SourceClusterName: "source",
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			got, err := store.RangeDeleteReplicationTaskFromDLQ(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestGetTimerIndexTasks(t *testing.T) {
	shardID := 1
	testCases := []struct {
		name      string
		req       *persistence.GetTimerIndexTasksRequest
		mockSetup func(*sqlplugin.MockDB, *serialization.MockParser)
		want      *persistence.GetTimerIndexTasksResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.GetTimerIndexTasksRequest{
				NextPageToken: func() []byte {
					ti := &timerTaskPageToken{TaskID: 101, Timestamp: time.Unix(1000, 1).UTC()}
					token, err := ti.serialize()
					require.NoError(t, err, "failed to serialize timer task page token")
					return token
				}(),
				MaxTimestamp: time.Unix(2000, 0).UTC(),
				BatchSize:    1,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromTimerTasks(gomock.Any(), &sqlplugin.TimerTasksFilter{
					ShardID:                shardID,
					MinVisibilityTimestamp: time.Unix(1000, 1).UTC(),
					TaskID:                 101,
					MaxVisibilityTimestamp: time.Unix(2000, 0).UTC(),
					PageSize:               2,
				}).Return([]sqlplugin.TimerTasksRow{
					{
						ShardID:             shardID,
						VisibilityTimestamp: time.Unix(1001, 0).UTC(),
						TaskID:              11,
						Data:                []byte(`timer`),
						DataEncoding:        "timer",
					},
					{
						ShardID:             shardID,
						VisibilityTimestamp: time.Unix(1003, 0).UTC(),
						TaskID:              22,
						Data:                []byte(`timer`),
						DataEncoding:        "timer",
					},
				}, nil)
				mockParser.EXPECT().TimerTaskInfoFromBlob([]byte(`timer`), "timer").Return(&serialization.TimerTaskInfo{
					DomainID:        serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID:      "test",
					RunID:           serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a54a"),
					TaskType:        1,
					Version:         202,
					EventID:         10,
					ScheduleAttempt: 9,
					TimeoutType:     common.Int16Ptr(12),
				}, nil).Times(2)
			},
			want: &persistence.GetTimerIndexTasksResponse{
				Timers: []*persistence.TimerTaskInfo{
					{
						VisibilityTimestamp: time.Unix(1001, 0).UTC(),
						TaskID:              11,
						DomainID:            "abdcea69-61d5-44c3-9d55-afe23505a542",
						WorkflowID:          "test",
						RunID:               "abdcea69-61d5-44c3-9d55-afe23505a54a",
						TaskType:            1,
						Version:             202,
						EventID:             10,
						ScheduleAttempt:     9,
						TimeoutType:         12,
					},
				},
				NextPageToken: func() []byte {
					ti := &timerTaskPageToken{TaskID: 22, Timestamp: time.Unix(1003, 0).UTC()}
					token, err := ti.serialize()
					require.NoError(t, err, "failed to serialize timer page token")
					return token
				}(),
			},
			wantErr: false,
		},
		{
			name: "Error case - failed to load from database",
			req: &persistence.GetTimerIndexTasksRequest{
				NextPageToken: func() []byte {
					ti := &timerTaskPageToken{TaskID: 101, Timestamp: time.Unix(1000, 1).UTC()}
					token, err := ti.serialize()
					require.NoError(t, err, "failed to serialize timer task page token")
					return token
				}(),
				MaxTimestamp: time.Unix(2000, 0).UTC(),
				BatchSize:    10,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				err := errors.New("some error")
				mockDB.EXPECT().SelectFromTimerTasks(gomock.Any(), &sqlplugin.TimerTasksFilter{
					ShardID:                shardID,
					MinVisibilityTimestamp: time.Unix(1000, 1).UTC(),
					TaskID:                 101,
					MaxVisibilityTimestamp: time.Unix(2000, 0).UTC(),
					PageSize:               11,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
		{
			name: "Error case - failed to decode data",
			req: &persistence.GetTimerIndexTasksRequest{
				NextPageToken: func() []byte {
					ti := &timerTaskPageToken{TaskID: 101, Timestamp: time.Unix(1000, 1).UTC()}
					token, err := ti.serialize()
					require.NoError(t, err, "failed to serialize timer task page token")
					return token
				}(),
				MaxTimestamp: time.Unix(2000, 0).UTC(),
				BatchSize:    10,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockDB.EXPECT().SelectFromTimerTasks(gomock.Any(), &sqlplugin.TimerTasksFilter{
					ShardID:                shardID,
					MinVisibilityTimestamp: time.Unix(1000, 1).UTC(),
					TaskID:                 101,
					MaxVisibilityTimestamp: time.Unix(2000, 0).UTC(),
					PageSize:               11,
				}).Return([]sqlplugin.TimerTasksRow{
					{
						ShardID:             shardID,
						VisibilityTimestamp: time.Unix(1001, 0).UTC(),
						TaskID:              11,
						Data:                []byte(`timer`),
						DataEncoding:        "timer",
					},
				}, nil)
				mockParser.EXPECT().TimerTaskInfoFromBlob([]byte(`timer`), "timer").Return(nil, errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			mockParser := serialization.NewMockParser(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), mockParser, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB, mockParser)

			got, err := store.GetTimerIndexTasks(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestCompleteTimerTask(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.CompleteTimerTaskRequest
		mockSetup func(*sqlplugin.MockDB)
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.CompleteTimerTaskRequest{
				TaskID:              123,
				VisibilityTimestamp: time.Unix(9, 1),
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().DeleteFromTimerTasks(gomock.Any(), &sqlplugin.TimerTasksFilter{
					ShardID:             shardID,
					TaskID:              123,
					VisibilityTimestamp: time.Unix(9, 1),
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.CompleteTimerTaskRequest{
				TaskID:              123,
				VisibilityTimestamp: time.Unix(9, 1),
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().DeleteFromTimerTasks(gomock.Any(), &sqlplugin.TimerTasksFilter{
					ShardID:             shardID,
					TaskID:              123,
					VisibilityTimestamp: time.Unix(9, 1),
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			err = store.CompleteTimerTask(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
			}
		})
	}
}

func TestRangeCompleteTimerTask(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.RangeCompleteTimerTaskRequest
		mockSetup func(*sqlplugin.MockDB)
		want      *persistence.RangeCompleteTimerTaskResponse
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.RangeCompleteTimerTaskRequest{
				InclusiveBeginTimestamp: time.Unix(9, 1),
				ExclusiveEndTimestamp:   time.Unix(99, 1),
				PageSize:                12,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().RangeDeleteFromTimerTasks(gomock.Any(), &sqlplugin.TimerTasksFilter{
					ShardID:                shardID,
					MinVisibilityTimestamp: time.Unix(9, 1),
					MaxVisibilityTimestamp: time.Unix(99, 1),
					PageSize:               12,
				}).Return(&sqlResult{rowsAffected: 11}, nil)
			},
			want: &persistence.RangeCompleteTimerTaskResponse{
				TasksCompleted: 11,
			},
			wantErr: false,
		},
		{
			name: "Error case",
			req: &persistence.RangeCompleteTimerTaskRequest{
				InclusiveBeginTimestamp: time.Unix(9, 1),
				ExclusiveEndTimestamp:   time.Unix(99, 1),
				PageSize:                12,
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				err := errors.New("some error")
				mockDB.EXPECT().RangeDeleteFromTimerTasks(gomock.Any(), &sqlplugin.TimerTasksFilter{
					ShardID:                shardID,
					MinVisibilityTimestamp: time.Unix(9, 1),
					MaxVisibilityTimestamp: time.Unix(99, 1),
					PageSize:               12,
				}).Return(nil, err)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			got, err := store.RangeCompleteTimerTask(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
				assert.Equal(t, tc.want, got, "Unexpected result for test case")
			}
		})
	}
}

func TestPutReplicationTaskToDLQ(t *testing.T) {
	shardID := 100
	testCases := []struct {
		name      string
		req       *persistence.InternalPutReplicationTaskToDLQRequest
		mockSetup func(*sqlplugin.MockDB, *serialization.MockParser)
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.InternalPutReplicationTaskToDLQRequest{
				SourceClusterName: "source",
				TaskInfo: &persistence.InternalReplicationTaskInfo{
					DomainID:          "abdcea69-61d5-44c3-9d55-afe23505a542",
					WorkflowID:        "test",
					RunID:             "abdcea69-61d5-44c3-9d55-afe23505a54a",
					TaskType:          1,
					TaskID:            101,
					Version:           202,
					FirstEventID:      10,
					NextEventID:       101,
					ScheduledID:       19,
					BranchToken:       []byte(`bt`),
					NewRunBranchToken: []byte(`nbt`),
					CreationTime:      time.Unix(1, 1),
				},
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockParser.EXPECT().ReplicationTaskInfoToBlob(&serialization.ReplicationTaskInfo{
					DomainID:                serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID:              "test",
					RunID:                   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a54a"),
					TaskType:                1,
					Version:                 202,
					FirstEventID:            10,
					NextEventID:             101,
					ScheduledID:             19,
					EventStoreVersion:       persistence.EventStoreVersion,
					NewRunEventStoreVersion: persistence.EventStoreVersion,
					BranchToken:             []byte(`bt`),
					NewRunBranchToken:       []byte(`nbt`),
					CreationTimestamp:       time.Unix(1, 1),
				}).Return(persistence.DataBlob{Data: []byte(`replication`), Encoding: "replication"}, nil)
				mockDB.EXPECT().InsertIntoReplicationTasksDLQ(gomock.Any(), &sqlplugin.ReplicationTaskDLQRow{
					SourceClusterName: "source",
					ShardID:           shardID,
					TaskID:            101,
					Data:              []byte(`replication`),
					DataEncoding:      "replication",
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "Error case - failed to encode data",
			req: &persistence.InternalPutReplicationTaskToDLQRequest{
				SourceClusterName: "source",
				TaskInfo: &persistence.InternalReplicationTaskInfo{
					DomainID:          "abdcea69-61d5-44c3-9d55-afe23505a542",
					WorkflowID:        "test",
					RunID:             "abdcea69-61d5-44c3-9d55-afe23505a54a",
					TaskType:          1,
					TaskID:            101,
					Version:           202,
					FirstEventID:      10,
					NextEventID:       101,
					ScheduledID:       19,
					BranchToken:       []byte(`bt`),
					NewRunBranchToken: []byte(`nbt`),
					CreationTime:      time.Unix(1, 1),
				},
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockParser.EXPECT().ReplicationTaskInfoToBlob(gomock.Any()).Return(persistence.DataBlob{}, errors.New("some error"))
			},
			wantErr: true,
		},
		{
			name: "Error case - failed to insert into database",
			req: &persistence.InternalPutReplicationTaskToDLQRequest{
				SourceClusterName: "source",
				TaskInfo: &persistence.InternalReplicationTaskInfo{
					DomainID:          "abdcea69-61d5-44c3-9d55-afe23505a542",
					WorkflowID:        "test",
					RunID:             "abdcea69-61d5-44c3-9d55-afe23505a54a",
					TaskType:          1,
					TaskID:            101,
					Version:           202,
					FirstEventID:      10,
					NextEventID:       101,
					ScheduledID:       19,
					BranchToken:       []byte(`bt`),
					NewRunBranchToken: []byte(`nbt`),
					CreationTime:      time.Unix(1, 1),
				},
			},
			mockSetup: func(mockDB *sqlplugin.MockDB, mockParser *serialization.MockParser) {
				mockParser.EXPECT().ReplicationTaskInfoToBlob(gomock.Any()).Return(persistence.DataBlob{Data: []byte(`replication`), Encoding: "replication"}, nil)
				err := errors.New("some error")
				mockDB.EXPECT().InsertIntoReplicationTasksDLQ(gomock.Any(), gomock.Any()).Return(nil, err)
				mockDB.EXPECT().IsDupEntryError(err).Return(false)
				mockDB.EXPECT().IsNotFoundError(err).Return(true)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			mockParser := serialization.NewMockParser(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), mockParser, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB, mockParser)

			err = store.PutReplicationTaskToDLQ(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
			}
		})
	}
}

func TestDeleteWorkflowExecution(t *testing.T) {
	shardID := int64(100)
	testCases := []struct {
		name      string
		req       *persistence.DeleteWorkflowExecutionRequest
		mockSetup func(*sqlplugin.MockDB)
		wantErr   bool
	}{
		{
			name: "Success case",
			req: &persistence.DeleteWorkflowExecutionRequest{
				DomainID:   "abdcea69-61d5-44c3-9d55-afe23505a542",
				WorkflowID: "wid",
				RunID:      "bbdcea69-61d5-44c3-9d55-afe23505a542",
			},
			mockSetup: func(mockDB *sqlplugin.MockDB) {
				mockDB.EXPECT().DeleteFromExecutions(gomock.Any(), &sqlplugin.ExecutionsFilter{
					ShardID:    int(shardID),
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "wid",
					RunID:      serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
				}).Return(nil, nil)
				mockDB.EXPECT().DeleteFromActivityInfoMaps(gomock.Any(), &sqlplugin.ActivityInfoMapsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "wid",
					RunID:      serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
				}).Return(nil, nil)
				mockDB.EXPECT().DeleteFromTimerInfoMaps(gomock.Any(), &sqlplugin.TimerInfoMapsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "wid",
					RunID:      serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
				}).Return(nil, nil)
				mockDB.EXPECT().DeleteFromChildExecutionInfoMaps(gomock.Any(), &sqlplugin.ChildExecutionInfoMapsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "wid",
					RunID:      serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
				}).Return(nil, nil)
				mockDB.EXPECT().DeleteFromRequestCancelInfoMaps(gomock.Any(), &sqlplugin.RequestCancelInfoMapsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "wid",
					RunID:      serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
				}).Return(nil, nil)
				mockDB.EXPECT().DeleteFromSignalInfoMaps(gomock.Any(), &sqlplugin.SignalInfoMapsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "wid",
					RunID:      serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
				}).Return(nil, nil)
				mockDB.EXPECT().DeleteFromBufferedEvents(gomock.Any(), &sqlplugin.BufferedEventsFilter{
					ShardID:    int(shardID),
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "wid",
					RunID:      serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
				}).Return(nil, nil)
				mockDB.EXPECT().DeleteFromSignalsRequestedSets(gomock.Any(), &sqlplugin.SignalsRequestedSetsFilter{
					ShardID:    shardID,
					DomainID:   serialization.MustParseUUID("abdcea69-61d5-44c3-9d55-afe23505a542"),
					WorkflowID: "wid",
					RunID:      serialization.MustParseUUID("bbdcea69-61d5-44c3-9d55-afe23505a542"),
				}).Return(nil, nil)

			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := sqlplugin.NewMockDB(ctrl)
			store, err := NewSQLExecutionStore(mockDB, nil, int(shardID), nil, nil)
			require.NoError(t, err, "failed to create execution store")

			tc.mockSetup(mockDB)

			err = store.DeleteWorkflowExecution(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err, "Expected an error for test case")
			} else {
				assert.NoError(t, err, "Did not expect an error for test case")
			}
		})
	}
}

func TestTxExecuteShardLocked(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*sqlplugin.MockDB, *sqlplugin.MockTx)
		operation string
		rangeID   int64
		fn        func(sqlplugin.Tx) error
		wantError error
	}{
		{
			name: "Success",
			mockSetup: func(mockDB *sqlplugin.MockDB, mockTx *sqlplugin.MockTx) {
				mockDB.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(mockTx, nil)
				mockTx.EXPECT().ReadLockShards(gomock.Any(), gomock.Any()).Return(11, nil)
				mockTx.EXPECT().Commit().Return(nil)
			},
			operation: "Insert",
			rangeID:   11,
			fn:        func(sqlplugin.Tx) error { return nil },
			wantError: nil,
		},
		{
			name: "Error",
			mockSetup: func(mockDB *sqlplugin.MockDB, mockTx *sqlplugin.MockTx) {
				mockDB.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(mockTx, nil)
				mockTx.EXPECT().ReadLockShards(gomock.Any(), gomock.Any()).Return(11, nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB.EXPECT().IsNotFoundError(gomock.Any()).Return(false)
				mockDB.EXPECT().IsTimeoutError(gomock.Any()).Return(false)
				mockDB.EXPECT().IsThrottlingError(gomock.Any()).Return(false)
			},
			operation: "Insert",
			rangeID:   11,
			fn:        func(sqlplugin.Tx) error { return errors.New("error") },
			wantError: &types.InternalServiceError{Message: "Insert operation failed.  Error: error"},
		},
		{
			name: "Error - shard ownership lost",
			mockSetup: func(mockDB *sqlplugin.MockDB, mockTx *sqlplugin.MockTx) {
				mockDB.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(mockTx, nil)
				mockTx.EXPECT().ReadLockShards(gomock.Any(), gomock.Any()).Return(12, nil)
				mockTx.EXPECT().Rollback().Return(nil)
			},
			operation: "Insert",
			rangeID:   11,
			fn:        func(sqlplugin.Tx) error { return errors.New("error") },
			wantError: &persistence.ShardOwnershipLostError{ShardID: 0, Msg: "Failed to lock shard. Previous range ID: 11; new range ID: 12"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockDB := sqlplugin.NewMockDB(ctrl)
			mockTx := sqlplugin.NewMockTx(ctrl)
			tt.mockSetup(mockDB, mockTx)

			s := &sqlExecutionStore{
				shardID: 0,
				sqlStore: sqlStore{
					db:     mockDB,
					logger: testlogger.New(t),
				},
			}

			gotError := s.txExecuteShardLocked(context.Background(), 0, tt.operation, tt.rangeID, tt.fn)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}
