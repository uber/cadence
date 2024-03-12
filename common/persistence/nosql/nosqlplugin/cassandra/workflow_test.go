// Copyright (c) 2021 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

package cassandra

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/checksum"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
)

func TestInsertWorkflowExecutionWithTasks(t *testing.T) {
	tests := []struct {
		name                  string
		request               *nosqlplugin.CurrentWorkflowWriteRequest
		execution             *nosqlplugin.WorkflowExecutionRequest
		transferTasks         []*nosqlplugin.TransferTask
		crossClusterTasks     []*nosqlplugin.CrossClusterTask
		replicationTasks      []*nosqlplugin.ReplicationTask
		timerTasks            []*nosqlplugin.TimerTask
		shardCondition        *nosqlplugin.ShardCondition
		mapExecuteBatchCASErr error
		wantErr               bool
	}{
		{
			name: "success",
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			execution: &nosqlplugin.WorkflowExecutionRequest{
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
			},
		},
		{
			name: "createOrUpdateCurrentWorkflow step fails",
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteMode(-999), // unknown mode will cause failure
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			execution: &nosqlplugin.WorkflowExecutionRequest{
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
			},
			wantErr: true,
		},
		{
			name: "createWorkflowExecutionWithMergeMaps step fails",
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			execution: &nosqlplugin.WorkflowExecutionRequest{
				EventBufferWriteMode: nosqlplugin.EventBufferWriteModeAppend, // this will cause failure
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
			},
			wantErr: true,
		},
		{
			name:                  "executeCreateWorkflowBatchTransaction step fails",
			mapExecuteBatchCASErr: errors.New("some random error"), // this will cause failure
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			execution: &nosqlplugin.WorkflowExecutionRequest{
				EventBufferWriteMode: nosqlplugin.EventBufferWriteModeNone,
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			session := &fakeSession{
				iter:                      &fakeIter{},
				mapExecuteBatchCASApplied: true,
				mapExecuteBatchCASErr:     tc.mapExecuteBatchCASErr,
			}
			client := gocql.NewMockClient(ctrl)
			cfg := &config.NoSQL{}
			logger := testlogger.New(t)
			dc := &persistence.DynamicConfiguration{}

			db := newCassandraDBFromSession(cfg, session, logger, dc, dbWithClient(client))

			err := db.InsertWorkflowExecutionWithTasks(
				context.Background(),
				tc.request,
				tc.execution,
				tc.transferTasks,
				tc.crossClusterTasks,
				tc.replicationTasks,
				tc.timerTasks,
				tc.shardCondition,
			)

			if (err != nil) != tc.wantErr {
				t.Errorf("InsertWorkflowExecutionWithTasks() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestSelectCurrentWorkflow(t *testing.T) {
	tests := []struct {
		name             string
		shardID          int
		domainID         string
		workflowID       string
		currentRunID     string
		lastWriteVersion int64
		createReqID      string
		wantErr          bool
		wantRow          *nosqlplugin.CurrentWorkflowRow
	}{
		{
			name:         "success",
			shardID:      1,
			domainID:     "test-domain-id",
			workflowID:   "test-workflow-id",
			currentRunID: "test-run-id",
			wantErr:      false,
			wantRow: &nosqlplugin.CurrentWorkflowRow{
				ShardID:          1,
				DomainID:         "test-domain-id",
				WorkflowID:       "test-workflow-id",
				RunID:            "test-run-id",
				LastWriteVersion: -24,
			},
		},
		{
			name:         "mapscan failure",
			shardID:      1,
			domainID:     "test-domain-id",
			workflowID:   "test-workflow-id",
			currentRunID: "test-run-id",
			wantErr:      true,
		},
		{
			name:             "lastwriteversion populated",
			shardID:          1,
			domainID:         "test-domain-id",
			workflowID:       "test-workflow-id",
			currentRunID:     "test-run-id",
			lastWriteVersion: 123,
			wantErr:          false,
			wantRow: &nosqlplugin.CurrentWorkflowRow{
				ShardID:          1,
				DomainID:         "test-domain-id",
				WorkflowID:       "test-workflow-id",
				RunID:            "test-run-id",
				LastWriteVersion: 123,
			},
		},
		{
			name:         "create request id populated",
			shardID:      1,
			domainID:     "test-domain-id",
			workflowID:   "test-workflow-id",
			currentRunID: "test-run-id",
			createReqID:  "test-create-request-id",
			wantErr:      false,
			wantRow: &nosqlplugin.CurrentWorkflowRow{
				ShardID:          1,
				DomainID:         "test-domain-id",
				WorkflowID:       "test-workflow-id",
				RunID:            "test-run-id",
				LastWriteVersion: -24,
				CreateRequestID:  "test-create-request-id",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
			query.EXPECT().MapScan(gomock.Any()).DoAndReturn(func(m map[string]interface{}) error {
				mockCurrentRunID := gocql.NewMockUUID(ctrl)
				m["current_run_id"] = mockCurrentRunID
				if !tc.wantErr {
					mockCurrentRunID.EXPECT().String().Return(tc.currentRunID).Times(1)
				}

				execMap := map[string]interface{}{}
				if tc.createReqID != "" {
					mockReqID := gocql.NewMockUUID(ctrl)
					mockReqID.EXPECT().String().Return(tc.createReqID).Times(1)
					execMap["create_request_id"] = mockReqID
				}
				m["execution"] = execMap

				if tc.lastWriteVersion != 0 {
					m["workflow_last_write_version"] = tc.lastWriteVersion
				}

				if tc.wantErr {
					return errors.New("some random error")
				}

				return nil
			}).Times(1)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			row, err := db.SelectCurrentWorkflow(context.Background(), tc.shardID, tc.domainID, tc.workflowID)

			if (err != nil) != tc.wantErr {
				t.Errorf("SelectCurrentWorkflow() error: %v, wantErr?: %v", err, tc.wantErr)
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(tc.wantRow, row); diff != "" {
				t.Fatalf("Row mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUpdateWorkflowExecutionWithTasks(t *testing.T) {
	tests := []struct {
		name                  string
		request               *nosqlplugin.CurrentWorkflowWriteRequest
		mutatedExecution      *nosqlplugin.WorkflowExecutionRequest
		insertedExecution     *nosqlplugin.WorkflowExecutionRequest
		resetExecution        *nosqlplugin.WorkflowExecutionRequest
		transferTasks         []*nosqlplugin.TransferTask
		crossClusterTasks     []*nosqlplugin.CrossClusterTask
		replicationTasks      []*nosqlplugin.ReplicationTask
		timerTasks            []*nosqlplugin.TimerTask
		shardCondition        *nosqlplugin.ShardCondition
		mapExecuteBatchCASErr error
		wantErr               bool
	}{
		{
			name: "both mutatedExecution and resetExecution not provided",
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeUpdate,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			wantErr: true,
		},
		{
			name: "mutatedExecution provided - success",
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			mutatedExecution: &nosqlplugin.WorkflowExecutionRequest{
				MapsWriteMode: nosqlplugin.WorkflowExecutionMapsWriteModeUpdate,
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
				PreviousNextEventIDCondition: common.Int64Ptr(123),
			},
		},
		{
			name:    "mutatedExecution provided - update fails",
			wantErr: true,
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			mutatedExecution: &nosqlplugin.WorkflowExecutionRequest{
				MapsWriteMode: nosqlplugin.WorkflowExecutionMapsWriteModeCreate, // this will cause failure
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
				PreviousNextEventIDCondition: common.Int64Ptr(123),
			},
		},
		{
			name: "resetExecution provided - success",
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			resetExecution: &nosqlplugin.WorkflowExecutionRequest{
				EventBufferWriteMode: nosqlplugin.EventBufferWriteModeClear,
				MapsWriteMode:        nosqlplugin.WorkflowExecutionMapsWriteModeReset,
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
				PreviousNextEventIDCondition: common.Int64Ptr(123),
			},
		},
		{
			name:    "resetExecution provided - reset fails",
			wantErr: true,
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			resetExecution: &nosqlplugin.WorkflowExecutionRequest{
				EventBufferWriteMode: nosqlplugin.EventBufferWriteModeNone, // this will cause failure
				MapsWriteMode:        nosqlplugin.WorkflowExecutionMapsWriteModeReset,
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
				PreviousNextEventIDCondition: common.Int64Ptr(123),
			},
		},
		{
			name: "resetExecution and insertedExecution provided - success",
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			resetExecution: &nosqlplugin.WorkflowExecutionRequest{
				EventBufferWriteMode: nosqlplugin.EventBufferWriteModeClear,
				MapsWriteMode:        nosqlplugin.WorkflowExecutionMapsWriteModeReset,
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
				PreviousNextEventIDCondition: common.Int64Ptr(123),
			},
			insertedExecution: &nosqlplugin.WorkflowExecutionRequest{
				EventBufferWriteMode: nosqlplugin.EventBufferWriteModeNone,
				MapsWriteMode:        nosqlplugin.WorkflowExecutionMapsWriteModeCreate,
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
				PreviousNextEventIDCondition: common.Int64Ptr(123),
			},
		},
		{
			name:    "resetExecution and insertedExecution provided - insert fails",
			wantErr: true,
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			resetExecution: &nosqlplugin.WorkflowExecutionRequest{
				EventBufferWriteMode: nosqlplugin.EventBufferWriteModeClear,
				MapsWriteMode:        nosqlplugin.WorkflowExecutionMapsWriteModeReset,
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
				PreviousNextEventIDCondition: common.Int64Ptr(123),
			},
			insertedExecution: &nosqlplugin.WorkflowExecutionRequest{
				EventBufferWriteMode: nosqlplugin.EventBufferWriteModeClear, // this will cause failure
				MapsWriteMode:        nosqlplugin.WorkflowExecutionMapsWriteModeCreate,
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
				PreviousNextEventIDCondition: common.Int64Ptr(123),
			},
		},
		{
			name:    "createOrUpdateCurrentWorkflow step fails",
			wantErr: true,
			request: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeUpdate,
				Condition: nil, // this will cause failure because Condition must be non-nil for update mode
			},
			shardCondition: &nosqlplugin.ShardCondition{
				ShardID: 1,
			},
			resetExecution: &nosqlplugin.WorkflowExecutionRequest{
				EventBufferWriteMode: nosqlplugin.EventBufferWriteModeClear,
				MapsWriteMode:        nosqlplugin.WorkflowExecutionMapsWriteModeReset,
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					CompletionEvent: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-completion-event"),
					},
					AutoResetPoints: &persistence.DataBlob{
						Encoding: common.EncodingTypeThriftRW,
						Data:     []byte("test-auto-reset-points"),
					},
				},
				VersionHistories: &persistence.DataBlob{
					Encoding: common.EncodingTypeThriftRW,
					Data:     []byte("test-version-histories"),
				},
				Checksums: &checksum.Checksum{
					Version: 1,
					Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
					Value:   []byte("test-checksum"),
				},
				PreviousNextEventIDCondition: common.Int64Ptr(123),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			session := &fakeSession{
				iter:                      &fakeIter{},
				mapExecuteBatchCASApplied: true,
			}
			client := gocql.NewMockClient(ctrl)
			cfg := &config.NoSQL{}
			logger := testlogger.New(t)
			dc := &persistence.DynamicConfiguration{}

			db := newCassandraDBFromSession(cfg, session, logger, dc, dbWithClient(client))

			err := db.UpdateWorkflowExecutionWithTasks(
				context.Background(),
				tc.request,
				tc.mutatedExecution,
				tc.insertedExecution,
				tc.resetExecution,
				tc.transferTasks,
				tc.crossClusterTasks,
				tc.replicationTasks,
				tc.timerTasks,
				tc.shardCondition,
			)

			if (err != nil) != tc.wantErr {
				t.Errorf("UpdateWorkflowExecutionWithTasks() error: %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestDeleteCurrentWorkflow(t *testing.T) {
	tests := []struct {
		name                  string
		shardID               int
		domainID              string
		workflowID            string
		currentRunIDCondition string
		queryMockFn           func(query *gocql.MockQuery)
		wantErr               bool
	}{
		{
			name:                  "success",
			shardID:               1,
			domainID:              "test-domain-id",
			workflowID:            "test-workflow-id",
			currentRunIDCondition: "test-run-id",
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:                  "query exec fails",
			shardID:               1,
			domainID:              "test-domain-id",
			workflowID:            "test-workflow-id",
			currentRunIDCondition: "test-run-id",
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.DeleteCurrentWorkflow(context.Background(), tc.shardID, tc.domainID, tc.workflowID, tc.currentRunIDCondition)

			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteCurrentWorkflow() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestDeleteWorkflowExecution(t *testing.T) {
	tests := []struct {
		name        string
		shardID     int
		domainID    string
		workflowID  string
		runID       string
		queryMockFn func(query *gocql.MockQuery)
		wantErr     bool
	}{
		{
			name:       "success",
			shardID:    1,
			domainID:   "test-domain-id",
			workflowID: "test-workflow-id",
			runID:      "test-run-id",
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:       "query exec fails",
			shardID:    1,
			domainID:   "test-domain-id",
			workflowID: "test-workflow-id",
			runID:      "test-run-id",
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.DeleteWorkflowExecution(context.Background(), tc.shardID, tc.domainID, tc.workflowID, tc.runID)

			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteWorkflowExecution() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestDeleteTransferTask(t *testing.T) {
	tests := []struct {
		name        string
		shardID     int
		taskID      int64
		queryMockFn func(query *gocql.MockQuery)
		wantErr     bool
	}{
		{
			name:    "success",
			shardID: 1,
			taskID:  123,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:    "query exec fails",
			shardID: 1,
			taskID:  123,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.DeleteTransferTask(context.Background(), tc.shardID, tc.taskID)

			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteTransferTask() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestRangeDeleteTransferTasks(t *testing.T) {
	tests := []struct {
		name                 string
		shardID              int
		exclusiveBeginTaskID int64
		inclusiveEndTaskID   int64
		queryMockFn          func(query *gocql.MockQuery)
		wantErr              bool
	}{
		{
			name:                 "success",
			shardID:              1,
			exclusiveBeginTaskID: 123,
			inclusiveEndTaskID:   456,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:                 "query exec fails",
			shardID:              1,
			exclusiveBeginTaskID: 123,
			inclusiveEndTaskID:   456,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.RangeDeleteTransferTasks(context.Background(), tc.shardID, tc.exclusiveBeginTaskID, tc.inclusiveEndTaskID)

			if (err != nil) != tc.wantErr {
				t.Errorf("RangeDeleteTransferTasks() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestDeleteTimerTask(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name                string
		shardID             int
		taskID              int64
		visibilityTimestamp time.Time
		queryMockFn         func(query *gocql.MockQuery)
		wantErr             bool
	}{
		{
			name:                "success",
			shardID:             1,
			taskID:              123,
			visibilityTimestamp: now,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:                "query exec fails",
			shardID:             1,
			taskID:              123,
			visibilityTimestamp: now,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.DeleteTimerTask(context.Background(), tc.shardID, tc.taskID, tc.visibilityTimestamp)

			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteTimerTask() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestRangeDeleteTimerTasks(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name             string
		shardID          int
		inclusiveMinTime time.Time
		exclusiveMaxTime time.Time
		queryMockFn      func(query *gocql.MockQuery)
		wantErr          bool
	}{
		{
			name:             "success",
			shardID:          1,
			inclusiveMinTime: now,
			exclusiveMaxTime: now.Add(time.Hour),
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:             "query exec fails",
			shardID:          1,
			inclusiveMinTime: now,
			exclusiveMaxTime: now.Add(time.Hour),
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.RangeDeleteTimerTasks(context.Background(), tc.shardID, tc.inclusiveMinTime, tc.exclusiveMaxTime)

			if (err != nil) != tc.wantErr {
				t.Errorf("RangeDeleteTimerTasks() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestDeleteReplicationTask(t *testing.T) {
	tests := []struct {
		name        string
		shardID     int
		taskID      int64
		queryMockFn func(query *gocql.MockQuery)
		wantErr     bool
	}{
		{
			name:    "success",
			shardID: 1,
			taskID:  123,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:    "query exec fails",
			shardID: 1,
			taskID:  123,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.DeleteReplicationTask(context.Background(), tc.shardID, tc.taskID)

			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteReplicationTask() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestRangeDeleteReplicationTasks(t *testing.T) {
	tests := []struct {
		name               string
		shardID            int
		inclusiveEndTaskID int64
		queryMockFn        func(query *gocql.MockQuery)
		wantErr            bool
	}{
		{
			name:               "success",
			shardID:            1,
			inclusiveEndTaskID: 123,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:               "query exec fails",
			shardID:            1,
			inclusiveEndTaskID: 123,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.RangeDeleteReplicationTasks(context.Background(), tc.shardID, tc.inclusiveEndTaskID)

			if (err != nil) != tc.wantErr {
				t.Errorf("RangeDeleteReplicationTasks() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestDeleteCrossClusterTask(t *testing.T) {
	tests := []struct {
		name          string
		shardID       int
		targetCluster string
		taskID        int64
		queryMockFn   func(query *gocql.MockQuery)
		wantErr       bool
	}{
		{
			name:          "success",
			shardID:       1,
			targetCluster: "test-target-cluster",
			taskID:        123,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:          "query exec fails",
			shardID:       1,
			targetCluster: "test-target-cluster",
			taskID:        123,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.DeleteCrossClusterTask(context.Background(), tc.shardID, tc.targetCluster, tc.taskID)

			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteCrossClusterTask() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestRangeDeleteCrossClusterTasks(t *testing.T) {
	tests := []struct {
		name                 string
		shardID              int
		targetCluster        string
		exclusiveBeginTaskID int64
		inclusiveEndTaskID   int64
		queryMockFn          func(query *gocql.MockQuery)
		wantErr              bool
	}{
		{
			name:                 "success",
			shardID:              1,
			targetCluster:        "test-target-cluster",
			exclusiveBeginTaskID: 123,
			inclusiveEndTaskID:   456,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:                 "query exec fails",
			shardID:              1,
			targetCluster:        "test-target-cluster",
			exclusiveBeginTaskID: 123,
			inclusiveEndTaskID:   456,
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.RangeDeleteCrossClusterTasks(context.Background(), tc.shardID, tc.targetCluster, tc.exclusiveBeginTaskID, tc.inclusiveEndTaskID)

			if (err != nil) != tc.wantErr {
				t.Errorf("RangeDeleteCrossClusterTasks() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestInsertReplicationDLQTask(t *testing.T) {
	tests := []struct {
		name          string
		shardID       int
		sourceCluster string
		taskID        int64
		task          nosqlplugin.ReplicationTask
		queryMockFn   func(query *gocql.MockQuery)
		wantErr       bool
	}{
		{
			name:          "success",
			shardID:       1,
			sourceCluster: "test-source-cluster",
			taskID:        123,
			task: nosqlplugin.ReplicationTask{
				TaskID: 123,
			},
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:          "query exec fails",
			shardID:       1,
			sourceCluster: "test-source-cluster",
			taskID:        123,
			task: nosqlplugin.ReplicationTask{
				TaskID: 123,
			},
			queryMockFn: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(1)
				query.EXPECT().Exec().Return(errors.New("failed to exec")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			query := gocql.NewMockQuery(ctrl)
			tc.queryMockFn(query)

			session := &fakeSession{
				query: query,
			}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(nil, session, logger, nil, dbWithClient(gocql.NewMockClient(ctrl)))

			err := db.InsertReplicationDLQTask(context.Background(), tc.shardID, tc.sourceCluster, tc.task)

			if (err != nil) != tc.wantErr {
				t.Errorf("RangeDeleteCrossClusterTasks() error: %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}
