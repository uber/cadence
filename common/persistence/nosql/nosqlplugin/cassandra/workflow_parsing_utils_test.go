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

package cassandra

import (
	"testing"
	"time"

	cql "github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
)

type mockUUID struct {
	uuid string
}

func (m mockUUID) String() string {
	return m.uuid
}

func newMockUUID(s string) mockUUID {
	return mockUUID{s}
}

func Test_parseWorkflowExecutionInfo(t *testing.T) {

	completionEventData := []byte("completion event data")
	autoResetPointsData := []byte("auto reset points data")
	searchAttributes := map[string][]byte{"AttributeKey": []byte("AttributeValue")}
	memo := map[string][]byte{"MemoKey": []byte("MemoValue")}
	partitionConfig := map[string]string{"PartitionKey": "PartitionValue"}
	timeNow := time.Now()

	tests := []struct {
		args map[string]interface{}
		want *persistence.InternalWorkflowExecutionInfo
	}{
		{
			args: map[string]interface{}{
				"domain_id":                             newMockUUID("domain_id"),
				"workflow_id":                           "workflow_id",
				"run_id":                                newMockUUID("run_id"),
				"parent_workflow_id":                    "parent_workflow_id",
				"initiated_id":                          int64(1),
				"completion_event_batch_id":             int64(2),
				"task_list":                             "task_list",
				"workflow_type_name":                    "workflow_type_name",
				"workflow_timeout":                      10,
				"decision_task_timeout":                 5,
				"execution_context":                     []byte("execution context"),
				"state":                                 1,
				"close_status":                          2,
				"last_first_event_id":                   int64(3),
				"last_event_task_id":                    int64(4),
				"next_event_id":                         int64(5),
				"last_processed_event":                  int64(6),
				"start_time":                            timeNow,
				"last_updated_time":                     timeNow,
				"create_request_id":                     newMockUUID("create_request_id"),
				"signal_count":                          7,
				"history_size":                          int64(8),
				"decision_version":                      int64(9),
				"decision_schedule_id":                  int64(10),
				"decision_started_id":                   int64(11),
				"decision_request_id":                   "decision_request_id",
				"decision_timeout":                      8,
				"decision_timestamp":                    int64(200),
				"decision_scheduled_timestamp":          int64(201),
				"decision_original_scheduled_timestamp": int64(202),
				"decision_attempt":                      int64(203),
				"cancel_requested":                      true,
				"cancel_request_id":                     "cancel_request_id",
				"sticky_task_list":                      "sticky_task_list",
				"sticky_schedule_to_start_timeout":      9,
				"client_library_version":                "client_lib_version",
				"client_feature_version":                "client_feature_version",
				"client_impl":                           "client_impl",
				"attempt":                               12,
				"has_retry_policy":                      true,
				"init_interval":                         10,
				"backoff_coefficient":                   1.5,
				"max_interval":                          20,
				"max_attempts":                          13,
				"expiration_time":                       timeNow,
				"non_retriable_errors":                  []string{"error1", "error2"},
				"branch_token":                          []byte("branch token"),
				"cron_schedule":                         "cron_schedule",
				"expiration_seconds":                    14,
				"search_attributes":                     searchAttributes,
				"memo":                                  memo,
				"partition_config":                      partitionConfig,
				"completion_event":                      completionEventData,
				"completion_event_data_encoding":        "Proto3",
				"auto_reset_points":                     autoResetPointsData,
				"auto_reset_points_encoding":            "Proto3",
			},
			want: &persistence.InternalWorkflowExecutionInfo{
				DomainID:                           "domain_id",
				WorkflowID:                         "workflow_id",
				RunID:                              "run_id",
				ParentWorkflowID:                   "parent_workflow_id",
				InitiatedID:                        int64(1),
				CompletionEventBatchID:             int64(2),
				TaskList:                           "task_list",
				WorkflowTypeName:                   "workflow_type_name",
				WorkflowTimeout:                    common.SecondsToDuration(int64(10)),
				DecisionStartToCloseTimeout:        common.SecondsToDuration(int64(5)),
				ExecutionContext:                   []byte("execution context"),
				State:                              1,
				CloseStatus:                        2,
				LastFirstEventID:                   int64(3),
				LastEventTaskID:                    int64(4),
				NextEventID:                        int64(5),
				LastProcessedEvent:                 int64(6),
				StartTimestamp:                     timeNow,
				LastUpdatedTimestamp:               timeNow,
				CreateRequestID:                    "create_request_id",
				SignalCount:                        int32(7),
				HistorySize:                        int64(8),
				DecisionVersion:                    int64(9),
				DecisionScheduleID:                 int64(10),
				DecisionStartedID:                  int64(11),
				DecisionRequestID:                  "decision_request_id",
				DecisionTimeout:                    common.SecondsToDuration(int64(8)),
				DecisionStartedTimestamp:           time.Unix(0, int64(200)),
				DecisionScheduledTimestamp:         time.Unix(0, int64(201)),
				DecisionOriginalScheduledTimestamp: time.Unix(0, int64(202)),
				DecisionAttempt:                    int64(203),
				CancelRequested:                    true,
				CancelRequestID:                    "cancel_request_id",
				StickyTaskList:                     "sticky_task_list",
				StickyScheduleToStartTimeout:       common.SecondsToDuration(int64(9)),
				ClientLibraryVersion:               "client_lib_version",
				ClientFeatureVersion:               "client_feature_version",
				ClientImpl:                         "client_impl",
				Attempt:                            int32(12),
				HasRetryPolicy:                     true,
				InitialInterval:                    common.SecondsToDuration(int64(10)),
				BackoffCoefficient:                 1.5,
				MaximumInterval:                    common.SecondsToDuration(int64(20)),
				MaximumAttempts:                    int32(13),
				ExpirationTime:                     timeNow,
				NonRetriableErrors:                 []string{"error1", "error2"},
				Memo:                               memo,
				PartitionConfig:                    partitionConfig,
			},
		},
		{
			args: map[string]interface{}{
				"first_run_id":     newMockUUID("first_run_id"),
				"parent_domain_id": newMockUUID("parent_domain_id"),
				"parent_run_id":    newMockUUID("parent_run_id"),
			},
			want: &persistence.InternalWorkflowExecutionInfo{
				FirstExecutionRunID: "first_run_id",
				ParentDomainID:      "parent_domain_id",
				ParentRunID:         "parent_run_id",
			},
		},
		{
			args: map[string]interface{}{
				"first_run_id":     newMockUUID(emptyRunID),
				"parent_domain_id": newMockUUID(emptyDomainID),
				"parent_run_id":    newMockUUID(emptyRunID),
			},
			want: &persistence.InternalWorkflowExecutionInfo{},
		},
		{
			args: map[string]interface{}{
				"first_run_id": newMockUUID(cql.UUID{}.String()),
			},
			want: &persistence.InternalWorkflowExecutionInfo{
				FirstExecutionRunID: "",
			},
		},
	}
	for _, tt := range tests {
		result := parseWorkflowExecutionInfo(tt.args)
		assert.Equal(t, result.FirstExecutionRunID, tt.want.FirstExecutionRunID)
		assert.Equal(t, result.DomainID, tt.want.DomainID)
		assert.Equal(t, result.WorkflowID, tt.want.WorkflowID)
		assert.Equal(t, result.RunID, tt.want.RunID)
		assert.Equal(t, result.ParentWorkflowID, tt.want.ParentWorkflowID)
		assert.Equal(t, result.InitiatedID, tt.want.InitiatedID)
		assert.Equal(t, result.CompletionEventBatchID, tt.want.CompletionEventBatchID)
		assert.Equal(t, result.TaskList, tt.want.TaskList)
		assert.Equal(t, result.WorkflowTypeName, tt.want.WorkflowTypeName)
		assert.Equal(t, result.WorkflowTimeout, tt.want.WorkflowTimeout)
		assert.Equal(t, result.DecisionStartToCloseTimeout, tt.want.DecisionStartToCloseTimeout)
		assert.Equal(t, result.ExecutionContext, tt.want.ExecutionContext)
		assert.Equal(t, result.State, tt.want.State)
		assert.Equal(t, result.CloseStatus, tt.want.CloseStatus)
		assert.Equal(t, result.LastFirstEventID, tt.want.LastFirstEventID)
		assert.Equal(t, result.LastEventTaskID, tt.want.LastEventTaskID)
		assert.Equal(t, result.NextEventID, tt.want.NextEventID)
		assert.Equal(t, result.LastProcessedEvent, tt.want.LastProcessedEvent)
		assert.Equal(t, result.StartTimestamp, tt.want.StartTimestamp)
		assert.Equal(t, result.LastUpdatedTimestamp, tt.want.LastUpdatedTimestamp)
		assert.Equal(t, result.CreateRequestID, tt.want.CreateRequestID)
		assert.Equal(t, result.SignalCount, tt.want.SignalCount)
		assert.Equal(t, result.HistorySize, tt.want.HistorySize)
		assert.Equal(t, result.DecisionVersion, tt.want.DecisionVersion)
		assert.Equal(t, result.DecisionScheduleID, tt.want.DecisionScheduleID)
		assert.Equal(t, result.DecisionStartedID, tt.want.DecisionStartedID)
		assert.Equal(t, result.DecisionRequestID, tt.want.DecisionRequestID)
		assert.Equal(t, result.DecisionTimeout, tt.want.DecisionTimeout)
		assert.Equal(t, result.CancelRequested, tt.want.CancelRequested)
		assert.Equal(t, result.DecisionStartedTimestamp, tt.want.DecisionStartedTimestamp)
		assert.Equal(t, result.DecisionScheduledTimestamp, tt.want.DecisionScheduledTimestamp)
		assert.Equal(t, result.DecisionOriginalScheduledTimestamp, tt.want.DecisionOriginalScheduledTimestamp)
		assert.Equal(t, result.DecisionAttempt, tt.want.DecisionAttempt)
		assert.Equal(t, result.ParentDomainID, tt.want.ParentDomainID)
	}
}
