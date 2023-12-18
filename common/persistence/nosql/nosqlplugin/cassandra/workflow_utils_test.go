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
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"

	"github.com/google/go-cmp/cmp"
	"github.com/pborman/uuid"
)

// fakeSession is fake implementation of gocql.Session
type fakeSession struct {
	// inputs
	iter                      *fakeIter
	mapExecuteBatchCASApplied bool
	mapExecuteBatchCASPrev    map[string]any
	mapExecuteBatchCASErr     error
}

func (s *fakeSession) Query(string, ...interface{}) gocql.Query {
	return nil
}

func (s *fakeSession) NewBatch(gocql.BatchType) gocql.Batch {
	return nil
}

func (s *fakeSession) ExecuteBatch(gocql.Batch) error {
	return nil
}

func (s *fakeSession) MapExecuteBatchCAS(batch gocql.Batch, prev map[string]interface{}) (bool, gocql.Iter, error) {
	for k, v := range s.mapExecuteBatchCASPrev {
		prev[k] = v
	}
	return s.mapExecuteBatchCASApplied, s.iter, s.mapExecuteBatchCASErr
}

// fakeBatch is fake implementation of gocql.Batch
type fakeBatch struct {
	// outputs
	queries []string
}

// Query is fake implementation of gocql.Batch.Query
func (b *fakeBatch) Query(queryTmpl string, args ...interface{}) {
	queryTmpl = strings.ReplaceAll(queryTmpl, "?", "%v")
	b.queries = append(b.queries, fmt.Sprintf(queryTmpl, args...))
}

// WithContext is fake implementation of gocql.Batch.WithContext
func (b *fakeBatch) WithContext(context.Context) gocql.Batch {
	return nil
}

// WithTimestamp is fake implementation of gocql.Batch.WithTimestamp
func (b *fakeBatch) WithTimestamp(int64) gocql.Batch {
	return nil
}

// Consistency is fake implementation of gocql.Batch.Consistency
func (b *fakeBatch) Consistency(gocql.Consistency) gocql.Batch {
	return nil
}

// fakeQuery is fake implementation of gocql.Query
func (s *fakeSession) Close() {
}

// fakeIter is fake implementation of gocql.Iter
type fakeIter struct {
	// output parameters
	closed bool
}

// Scan is fake implementation of gocql.Iter.Scan
func (i *fakeIter) Scan(...interface{}) bool {
	return false
}

// MapScan is fake implementation of gocql.Iter.MapScan
func (i *fakeIter) MapScan(map[string]interface{}) bool {
	return false
}

// PageState is fake implementation of gocql.Iter.PageState
func (i *fakeIter) PageState() []byte {
	return nil
}

// Close is fake implementation of gocql.Iter.Close
func (i *fakeIter) Close() error {
	i.closed = true
	return nil
}

func TestExecuteCreateWorkflowBatchTransaction(t *testing.T) {
	tests := []struct {
		// fake setup parameters
		desc    string
		session *fakeSession

		// executeCreateWorkflowBatchTransaction args
		batch        *fakeBatch
		currentWFReq *nosqlplugin.CurrentWorkflowWriteRequest
		execution    *nosqlplugin.WorkflowExecutionRequest
		shardCond    *nosqlplugin.ShardCondition

		// expectations
		wantErr error
	}{
		{
			desc:  "applied",
			batch: &fakeBatch{},
			session: &fakeSession{
				mapExecuteBatchCASApplied: true,
				iter:                      &fakeIter{},
			},
		},
		{
			desc:  "CAS error",
			batch: &fakeBatch{},
			session: &fakeSession{
				mapExecuteBatchCASErr: fmt.Errorf("db operation failed for some reason"),
				iter:                  &fakeIter{},
			},
			wantErr: fmt.Errorf("db operation failed for some reason"),
		},
		{
			desc: "shard range id mismatch",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":     rowTypeShard,
					"run_id":   uuid.Parse("bda9cd9c-32fb-4267-b120-346e5351fc46"),
					"range_id": int64(200),
				},
			},
			batch:        &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{},
			shardCond: &nosqlplugin.ShardCondition{
				RangeID: 100,
			},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				ShardRangeIDNotMatch: common.Int64Ptr(200),
			},
		},
		{
			desc: "execution already exists",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":                        rowTypeExecution,
					"run_id":                      uuid.Parse(permanentRunID),
					"range_id":                    int64(100),
					"workflow_last_write_version": int64(3),
					"execution": map[string]any{
						"workflow_id": "test-workflow-id",
						"run_id":      uuid.Parse("bda9cd9c-32fb-4267-b120-346e5351fc46"),
						"state":       1,
					},
				},
			},
			batch:        &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{},
			shardCond: &nosqlplugin.ShardCondition{
				RangeID: 100,
			},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				CurrentWorkflowConditionFailInfo: common.StringPtr(
					"Workflow execution already running. WorkflowId: test-workflow-id, " +
						"RunId: bda9cd9c-32fb-4267-b120-346e5351fc46, rangeID: 100"),
			},
		},
		{
			desc: "execution already exists and write mode is CurrentWorkflowWriteModeInsert",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":                        rowTypeExecution,
					"run_id":                      uuid.Parse(permanentRunID),
					"range_id":                    int64(100),
					"workflow_last_write_version": int64(3),
					"execution": map[string]any{
						"workflow_id": "test-workflow-id",
						"run_id":      uuid.Parse("bda9cd9c-32fb-4267-b120-346e5351fc46"),
						"state":       1,
					},
				},
			},
			batch: &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeInsert,
			},
			shardCond: &nosqlplugin.ShardCondition{
				RangeID: 100,
			},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				WorkflowExecutionAlreadyExists: &nosqlplugin.WorkflowExecutionAlreadyExists{
					RunID:            "bda9cd9c-32fb-4267-b120-346e5351fc46",
					State:            1,
					LastWriteVersion: 3,
					OtherInfo:        "Workflow execution already running. WorkflowId: test-workflow-id, RunId: bda9cd9c-32fb-4267-b120-346e5351fc46, rangeID: 100",
				},
			},
		},
		{
			desc: "creation condition failed by mismatch runID",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":                        rowTypeExecution,
					"run_id":                      uuid.Parse(permanentRunID),
					"range_id":                    int64(100),
					"workflow_last_write_version": int64(3),
					"current_run_id":              uuid.Parse("bda9cd9c-32fb-4267-b120-346e5351fc46"),
				},
			},
			batch: &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{
				Condition: &nosqlplugin.CurrentWorkflowWriteCondition{
					CurrentRunID: common.StringPtr("fd88863f-bb32-4daa-8878-49e08b91545e"), // not matching current_run_id above on purpose
				},
			},
			execution: &nosqlplugin.WorkflowExecutionRequest{
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					WorkflowID: "wfid",
				},
			},
			shardCond: &nosqlplugin.ShardCondition{
				RangeID: 100,
			},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				CurrentWorkflowConditionFailInfo: common.StringPtr(
					"Workflow execution creation condition failed by mismatch runID. " +
						"WorkflowId: wfid, Expected Current RunID: fd88863f-bb32-4daa-8878-49e08b91545e, " +
						"Actual Current RunID: bda9cd9c-32fb-4267-b120-346e5351fc46"),
			},
		},
		{
			desc: "creation condition failed",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":                        rowTypeExecution,
					"run_id":                      uuid.Parse(permanentRunID),
					"range_id":                    int64(100),
					"workflow_last_write_version": int64(3),
					"current_run_id":              uuid.Parse("bda9cd9c-32fb-4267-b120-346e5351fc46"),
				},
			},
			batch:        &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{},
			execution: &nosqlplugin.WorkflowExecutionRequest{
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					WorkflowID: "wfid",
					RunID:      "bda9cd9c-32fb-4267-b120-346e5351fc46",
				},
			},
			shardCond: &nosqlplugin.ShardCondition{
				RangeID: 100,
			},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				CurrentWorkflowConditionFailInfo: common.StringPtr(
					"Workflow execution creation condition failed. WorkflowId: wfid, " +
						"CurrentRunID: bda9cd9c-32fb-4267-b120-346e5351fc46"),
			},
		},
		{
			desc: "execution already running",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":                        rowTypeExecution,
					"run_id":                      uuid.Parse("bda9cd9c-32fb-4267-b120-346e5351fc46"),
					"range_id":                    int64(100),
					"workflow_last_write_version": int64(3),
				},
			},
			batch:        &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{},
			execution: &nosqlplugin.WorkflowExecutionRequest{
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					WorkflowID:      "wfid",
					RunID:           "bda9cd9c-32fb-4267-b120-346e5351fc46",
					CreateRequestID: "reqid_123",
					State:           2,
				},
			},
			shardCond: &nosqlplugin.ShardCondition{
				RangeID: 100,
			},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				WorkflowExecutionAlreadyExists: &nosqlplugin.WorkflowExecutionAlreadyExists{
					OtherInfo: "Workflow execution already running. WorkflowId: wfid, " +
						"RunId: bda9cd9c-32fb-4267-b120-346e5351fc46, rangeID: 100",
					CreateRequestID:  "reqid_123",
					RunID:            "bda9cd9c-32fb-4267-b120-346e5351fc46",
					State:            2,
					LastWriteVersion: 3,
				},
			},
		},
		{
			desc: "unknown condition failure",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":                        rowTypeExecution,
					"run_id":                      uuid.Parse("bda9cd9c-32fb-4267-b120-346e5351fc46"),
					"range_id":                    int64(100),
					"workflow_last_write_version": int64(3),
				},
			},
			batch:        &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{},
			execution: &nosqlplugin.WorkflowExecutionRequest{
				InternalWorkflowExecutionInfo: persistence.InternalWorkflowExecutionInfo{
					RunID: "something else",
				},
			},
			shardCond: &nosqlplugin.ShardCondition{
				RangeID: 100,
			},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				UnknownConditionFailureDetails: common.StringPtr("Failed to operate on workflow execution.  Request RangeID: 100"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := executeCreateWorkflowBatchTransaction(tc.session, tc.batch, tc.currentWFReq, tc.execution, tc.shardCond)
			if diff := errDiff(tc.wantErr, err); diff != "" {
				t.Fatalf("Error mismatch (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if !tc.session.iter.closed {
				t.Error("iterator not closed")
			}
		})
	}
}

func TestExecuteUpdateWorkflowBatchTransaction(t *testing.T) {
	tests := []struct {
		// fake setup parameters
		desc    string
		session *fakeSession

		// executeUpdateWorkflowBatchTransaction args
		batch               *fakeBatch
		currentWFReq        *nosqlplugin.CurrentWorkflowWriteRequest
		prevNextEventIDCond int64
		shardCond           *nosqlplugin.ShardCondition

		// expectations
		wantErr error
	}{
		{
			desc:  "applied",
			batch: &fakeBatch{},
			session: &fakeSession{
				mapExecuteBatchCASApplied: true,
				iter:                      &fakeIter{},
			},
		},
		{
			desc:  "CAS error",
			batch: &fakeBatch{},
			session: &fakeSession{
				mapExecuteBatchCASErr: fmt.Errorf("db operation failed for some reason"),
				iter:                  &fakeIter{},
			},
			wantErr: fmt.Errorf("db operation failed for some reason"),
		},
		{
			desc: "range id mismatch for shard row",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":     rowTypeShard,
					"run_id":   uuid.Parse("bda9cd9c-32fb-4267-b120-346e5351fc46"),
					"range_id": int64(200),
				},
			},
			batch:        &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{},
			shardCond: &nosqlplugin.ShardCondition{
				RangeID: 100,
			},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				ShardRangeIDNotMatch: common.Int64Ptr(200),
			},
		},
		{
			desc: "nextEventID mismatch for execution row",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":          rowTypeExecution,
					"run_id":        uuid.Parse("0875863e-dcef-496a-b8a2-3210b2958e25"),
					"next_event_id": int64(10),
				},
			},
			batch:               &fakeBatch{},
			prevNextEventIDCond: 11, // not matching next_event_id above on purpose
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{
				Row: nosqlplugin.CurrentWorkflowRow{
					RunID: "0875863e-dcef-496a-b8a2-3210b2958e25",
				},
			},
			shardCond: &nosqlplugin.ShardCondition{
				RangeID: 200,
			},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				UnknownConditionFailureDetails: common.StringPtr(
					"Failed to update mutable state. " +
						"previousNextEventIDCondition: 11, actualNextEventID: 10, Request Current RunID: 0875863e-dcef-496a-b8a2-3210b2958e25"),
			},
		},
		{
			desc: "runID mismatch for current execution row",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":           rowTypeExecution,
					"run_id":         uuid.Parse(permanentRunID),
					"current_run_id": uuid.Parse("0875863e-dcef-496a-b8a2-3210b2958e25"),
				},
			},
			batch: &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{
				Condition: &nosqlplugin.CurrentWorkflowWriteCondition{
					CurrentRunID: common.StringPtr("fd88863f-bb32-4daa-8878-49e08b91545e"), // not matching current_run_id above on purpose
				},
			},
			shardCond: &nosqlplugin.ShardCondition{},
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				CurrentWorkflowConditionFailInfo: common.StringPtr(
					"Failed to update mutable state. requestConditionalRunID: fd88863f-bb32-4daa-8878-49e08b91545e, " +
						"Actual Value: 0875863e-dcef-496a-b8a2-3210b2958e25"),
			},
		},
		{
			desc: "unknown condition failure",
			session: &fakeSession{
				mapExecuteBatchCASApplied: false,
				iter:                      &fakeIter{},
				mapExecuteBatchCASPrev: map[string]any{
					"type":           rowTypeExecution,
					"run_id":         uuid.Parse(permanentRunID),
					"current_run_id": uuid.Parse("0875863e-dcef-496a-b8a2-3210b2958e25"),
				},
			},
			batch: &fakeBatch{},
			currentWFReq: &nosqlplugin.CurrentWorkflowWriteRequest{
				Condition: &nosqlplugin.CurrentWorkflowWriteCondition{
					CurrentRunID: common.StringPtr("0875863e-dcef-496a-b8a2-3210b2958e25"), // not matching current_run_id above on purpose
				},
			},
			shardCond: &nosqlplugin.ShardCondition{
				ShardID: 345,
				RangeID: 200,
			},
			prevNextEventIDCond: 11,
			wantErr: &nosqlplugin.WorkflowOperationConditionFailure{
				UnknownConditionFailureDetails: common.StringPtr(
					"Failed to update mutable state. ShardID: 345, RangeID: 200, previousNextEventIDCondition: 11, requestConditionalRunID: 0875863e-dcef-496a-b8a2-3210b2958e25"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := executeUpdateWorkflowBatchTransaction(tc.session, tc.batch, tc.currentWFReq, tc.prevNextEventIDCond, tc.shardCond)
			if diff := errDiff(tc.wantErr, err); diff != "" {
				t.Fatalf("Error mismatch (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if !tc.session.iter.closed {
				t.Error("iterator not closed")
			}
		})
	}
}

func TestCreateTimerTasks(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2023-12-12T22:08:41Z")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		desc       string
		shardID    int
		domainID   string
		workflowID string
		timerTasks []*nosqlplugin.TimerTask
		// expectations
		wantQueries []string
	}{
		{
			desc:       "ok",
			shardID:    1000,
			domainID:   "domain_xyz",
			workflowID: "workflow_xyz",
			timerTasks: []*nosqlplugin.TimerTask{
				{
					RunID:               "rundid_1",
					TaskID:              1,
					TaskType:            1,
					TimeoutType:         1,
					EventID:             10,
					ScheduleAttempt:     0,
					Version:             0,
					VisibilityTimestamp: ts,
				},
				{
					RunID:               "rundid_1",
					TaskID:              2,
					TaskType:            1,
					TimeoutType:         1,
					EventID:             11,
					ScheduleAttempt:     0,
					Version:             0,
					VisibilityTimestamp: ts.Add(time.Minute),
				},
			},
			wantQueries: []string{
				`INSERT INTO executions (shard_id, type, domain_id, workflow_id, run_id, timer, visibility_ts, task_id) ` +
					`VALUES(1000, 3, 10000000-4000-f000-f000-000000000000, 20000000-4000-f000-f000-000000000000, 30000000-4000-f000-f000-000000000000, ` +
					`{domain_id: domain_xyz, workflow_id: workflow_xyz, run_id: rundid_1, visibility_ts: 1702418921000, task_id: 1, type: 1, timeout_type: 1, event_id: 10, schedule_attempt: 0, version: 0}, ` +
					`1702418921000, 1)`,
				`INSERT INTO executions (shard_id, type, domain_id, workflow_id, run_id, timer, visibility_ts, task_id) ` +
					`VALUES(1000, 3, 10000000-4000-f000-f000-000000000000, 20000000-4000-f000-f000-000000000000, 30000000-4000-f000-f000-000000000000, ` +
					`{domain_id: domain_xyz, workflow_id: workflow_xyz, run_id: rundid_1, visibility_ts: 1702418981000, task_id: 2, type: 1, timeout_type: 1, event_id: 11, schedule_attempt: 0, version: 0}, ` +
					`1702418981000, 2)`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			batch := &fakeBatch{}
			err := createTimerTasks(batch, tc.shardID, tc.domainID, tc.workflowID, tc.timerTasks)
			if err != nil {
				t.Fatalf("createTimerTasks failed: %v", err)
			}

			if diff := cmp.Diff(tc.wantQueries, batch.queries); diff != "" {
				t.Fatalf("Query mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReplicationTasks(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2023-12-12T22:08:41Z")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		desc       string
		shardID    int
		domainID   string
		workflowID string
		replTasks  []*nosqlplugin.ReplicationTask
		// expectations
		wantQueries []string
	}{
		{
			desc:       "ok",
			shardID:    1000,
			domainID:   "domain_xyz",
			workflowID: "workflow_xyz",
			replTasks: []*nosqlplugin.ReplicationTask{
				{
					RunID:             "rundid_1",
					TaskID:            644,
					TaskType:          0,
					FirstEventID:      5,
					NextEventID:       8,
					Version:           0,
					ScheduledID:       common.EmptyEventID,
					NewRunBranchToken: []byte{'a', 'b', 'c'},
					CreationTime:      ts,
				},
				{
					RunID:             "rundid_1",
					TaskID:            645,
					TaskType:          0,
					FirstEventID:      25,
					NextEventID:       28,
					Version:           0,
					ScheduledID:       common.EmptyEventID,
					NewRunBranchToken: []byte{'a', 'b', 'c'},
					CreationTime:      ts.Add(time.Hour),
				},
			},
			wantQueries: []string{
				`INSERT INTO executions (shard_id, type, domain_id, workflow_id, run_id, replication, visibility_ts, task_id) ` +
					`VALUES(1000, 4, 10000000-5000-f000-f000-000000000000, 20000000-5000-f000-f000-000000000000, 30000000-5000-f000-f000-000000000000, ` +
					`{domain_id: domain_xyz, workflow_id: workflow_xyz, run_id: rundid_1, task_id: 644, type: 0, ` +
					`first_event_id: 5,next_event_id: 8,version: 0,scheduled_id: -23, event_store_version: 2, branch_token: [], ` +
					`new_run_event_store_version: 2, new_run_branch_token: [97 98 99], created_time: 1702418921000000000 }, 946684800000, 644)`,
				`INSERT INTO executions (shard_id, type, domain_id, workflow_id, run_id, replication, visibility_ts, task_id) ` +
					`VALUES(1000, 4, 10000000-5000-f000-f000-000000000000, 20000000-5000-f000-f000-000000000000, 30000000-5000-f000-f000-000000000000, ` +
					`{domain_id: domain_xyz, workflow_id: workflow_xyz, run_id: rundid_1, task_id: 645, type: 0, ` +
					`first_event_id: 25,next_event_id: 28,version: 0,scheduled_id: -23, event_store_version: 2, branch_token: [], ` +
					`new_run_event_store_version: 2, new_run_branch_token: [97 98 99], created_time: 1702422521000000000 }, 946684800000, 645)`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			batch := &fakeBatch{}
			err := createReplicationTasks(batch, tc.shardID, tc.domainID, tc.workflowID, tc.replTasks)
			if err != nil {
				t.Fatalf("createReplicationTasks failed: %v", err)
			}

			if diff := cmp.Diff(tc.wantQueries, batch.queries); diff != "" {
				t.Fatalf("Query mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTransferTasks(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2023-12-12T22:08:41Z")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		desc          string
		shardID       int
		domainID      string
		workflowID    string
		transferTasks []*nosqlplugin.TransferTask
		// expectations
		wantQueries []string
	}{
		{
			desc:       "ok",
			shardID:    1000,
			domainID:   "domain_xyz",
			workflowID: "workflow_xyz",
			transferTasks: []*nosqlplugin.TransferTask{
				{
					RunID:                   "rundid_1",
					TaskID:                  355,
					TaskType:                0,
					Version:                 1,
					VisibilityTimestamp:     ts,
					TargetDomainID:          "e2bf2c8f-0ddf-4451-8840-27cfe8addd62",
					TargetWorkflowID:        persistence.TransferTaskTransferTargetWorkflowID,
					TargetRunID:             persistence.TransferTaskTransferTargetRunID,
					TargetChildWorkflowOnly: true,
					TaskList:                "tasklist_1",
					ScheduleID:              14,
				},
				{
					RunID:                   "rundid_2",
					TaskID:                  220,
					TaskType:                0,
					Version:                 1,
					VisibilityTimestamp:     ts.Add(time.Minute),
					TargetDomainID:          "e2bf2c8f-0ddf-4451-8840-27cfe8addd62",
					TargetWorkflowID:        persistence.TransferTaskTransferTargetWorkflowID,
					TargetRunID:             persistence.TransferTaskTransferTargetRunID,
					TargetChildWorkflowOnly: true,
					TaskList:                "tasklist_2",
					ScheduleID:              3,
				},
			},
			wantQueries: []string{
				`INSERT INTO executions (shard_id, type, domain_id, workflow_id, run_id, transfer, visibility_ts, task_id) ` +
					`VALUES(1000, 2, 10000000-3000-f000-f000-000000000000, 20000000-3000-f000-f000-000000000000, 30000000-3000-f000-f000-000000000000, ` +
					`{domain_id: domain_xyz, workflow_id: workflow_xyz, run_id: rundid_1, visibility_ts: 2023-12-12 22:08:41 +0000 UTC, ` +
					`task_id: 355, target_domain_id: e2bf2c8f-0ddf-4451-8840-27cfe8addd62, target_domain_ids: map[],` +
					`target_workflow_id: 20000000-0000-f000-f000-000000000001, target_run_id: 30000000-0000-f000-f000-000000000002, ` +
					`target_child_workflow_only: true, task_list: tasklist_1, type: 0, schedule_id: 14, record_visibility: false, version: 1}, ` +
					`946684800000, 355)`,
				`INSERT INTO executions (shard_id, type, domain_id, workflow_id, run_id, transfer, visibility_ts, task_id) ` +
					`VALUES(1000, 2, 10000000-3000-f000-f000-000000000000, 20000000-3000-f000-f000-000000000000, 30000000-3000-f000-f000-000000000000, ` +
					`{domain_id: domain_xyz, workflow_id: workflow_xyz, run_id: rundid_2, visibility_ts: 2023-12-12 22:09:41 +0000 UTC, ` +
					`task_id: 220, target_domain_id: e2bf2c8f-0ddf-4451-8840-27cfe8addd62, target_domain_ids: map[],` +
					`target_workflow_id: 20000000-0000-f000-f000-000000000001, target_run_id: 30000000-0000-f000-f000-000000000002, ` +
					`target_child_workflow_only: true, task_list: tasklist_2, type: 0, schedule_id: 3, record_visibility: false, version: 1}, ` +
					`946684800000, 220)`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			batch := &fakeBatch{}
			err := createTransferTasks(batch, tc.shardID, tc.domainID, tc.workflowID, tc.transferTasks)
			if err != nil {
				t.Fatalf("createTransferTasks failed: %v", err)
			}

			if diff := cmp.Diff(tc.wantQueries, batch.queries); diff != "" {
				t.Fatalf("Query mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCrossClusterTasks(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2023-12-12T22:08:41Z")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		desc          string
		shardID       int
		domainID      string
		workflowID    string
		xClusterTasks []*nosqlplugin.CrossClusterTask
		// expectations
		wantQueries []string
	}{
		{
			desc:       "ok",
			shardID:    1000,
			domainID:   "domain_xyz",
			workflowID: "workflow_xyz",
			xClusterTasks: []*nosqlplugin.CrossClusterTask{
				{
					TransferTask: nosqlplugin.TransferTask{
						RunID:                   "rundid_1",
						TaskID:                  355,
						TaskType:                0,
						Version:                 1,
						VisibilityTimestamp:     ts,
						TargetDomainID:          "e2bf2c8f-0ddf-4451-8840-27cfe8addd62",
						TargetWorkflowID:        persistence.TransferTaskTransferTargetWorkflowID,
						TargetRunID:             persistence.TransferTaskTransferTargetRunID,
						TargetChildWorkflowOnly: true,
						TaskList:                "tasklist_1",
						ScheduleID:              14,
					},
				},
			},
			wantQueries: []string{
				`INSERT INTO executions (shard_id, type, domain_id, workflow_id, run_id, cross_cluster, visibility_ts, task_id) ` +
					`VALUES(1000, 6, 10000000-7000-f000-f000-000000000000, , 30000000-7000-f000-f000-000000000000, ` +
					`{domain_id: domain_xyz, workflow_id: workflow_xyz, run_id: rundid_1, visibility_ts: 2023-12-12 22:08:41 +0000 UTC, ` +
					`task_id: 355, target_domain_id: e2bf2c8f-0ddf-4451-8840-27cfe8addd62, target_domain_ids: map[],` +
					`target_workflow_id: 20000000-0000-f000-f000-000000000001, target_run_id: 30000000-0000-f000-f000-000000000002, ` +
					`target_child_workflow_only: true, task_list: tasklist_1, type: 0, schedule_id: 14, record_visibility: false, version: 1}, ` +
					`946684800000, 355)`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			batch := &fakeBatch{}
			err := createCrossClusterTasks(batch, tc.shardID, tc.domainID, tc.workflowID, tc.xClusterTasks)
			if err != nil {
				t.Fatalf("createCrossClusterTasks failed: %v", err)
			}

			if diff := cmp.Diff(tc.wantQueries, batch.queries); diff != "" {
				t.Fatalf("Query mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func errDiff(want, got error) string {
	wantCondFailure, wantOk := want.(*nosqlplugin.WorkflowOperationConditionFailure)
	gotCondFailure, gotOk := got.(*nosqlplugin.WorkflowOperationConditionFailure)
	if wantOk && gotOk {
		arg1 := trimWorkflowConditionalFailureErr(wantCondFailure)
		arg2 := trimWorkflowConditionalFailureErr(gotCondFailure)
		return cmp.Diff(arg1, arg2)
	}

	wantMsg := ""
	if want != nil {
		wantMsg = want.Error()
	}
	gotMsg := ""
	if got != nil {
		gotMsg = got.Error()
	}
	return cmp.Diff(wantMsg, gotMsg)
}

func trimWorkflowConditionalFailureErr(condFailure *nosqlplugin.WorkflowOperationConditionFailure) any {
	trimColumnsPart(condFailure.CurrentWorkflowConditionFailInfo)
	trimColumnsPart(condFailure.UnknownConditionFailureDetails)
	if condFailure.WorkflowExecutionAlreadyExists != nil {
		trimColumnsPart(&condFailure.WorkflowExecutionAlreadyExists.OtherInfo)
	}
	return condFailure
}

func trimColumnsPart(s *string) {
	if s == nil {
		return
	}
	re := regexp.MustCompile(`, columns: \(.*\)`)
	trimmed := re.ReplaceAllString(*s, "")
	*s = trimmed
}
