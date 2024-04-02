// Copyright (c) 2017 Uber Technologies, Inc.
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

package persistence

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common/types"
)

func TestClusterReplicationConfigGetCopy(t *testing.T) {
	config := &ClusterReplicationConfig{
		ClusterName: "test",
	}
	assert.Equal(t, config, config.GetCopy()) // deep equal
	assert.Equal(t, true, config != config.GetCopy())
}

func TestIsTransientError(t *testing.T) {
	transientErrors := []error{
		&types.ServiceBusyError{},
		&types.InternalServiceError{},
		&TimeoutError{},
	}
	for _, err := range transientErrors {
		require.True(t, IsTransientError(err))
	}

	nonRetryableErrors := []error{
		&types.EntityNotExistsError{},
		&types.DomainAlreadyExistsError{},
		&WorkflowExecutionAlreadyStartedError{},
		errors.New("some unknown error"),
	}
	for _, err := range nonRetryableErrors {
		require.False(t, IsTransientError(err))
	}
}

func TestIsTimeoutError(t *testing.T) {
	notTimeoutError := fmt.Errorf("not timeout error")
	assert.False(t, IsTimeoutError(notTimeoutError))
	assert.True(t, IsTimeoutError(&TimeoutError{}))
}

func TestIsBackgroundTransientError(t *testing.T) {
	transientErrors := map[error]bool{
		&types.ServiceBusyError{}:     false,
		&types.InternalServiceError{}: true,
		&TimeoutError{}:               true,
	}
	for error, result := range transientErrors {
		assert.Equal(t, result, IsBackgroundTransientError(error))
	}
}

func TestNewHistoryBranch(t *testing.T) {
	newTreeID := "newTreeID"
	generalToken, err := NewHistoryBranchToken(newTreeID)
	assert.NoError(t, err)
	assert.NotEmpty(t, generalToken)

	newBranchID := "newBranchID"
	branchToken, err := NewHistoryBranchTokenByBranchID(newTreeID, newBranchID)
	assert.NoError(t, err)
	assert.NotEmpty(t, branchToken)

	anotherBranchToken, err := NewHistoryBranchTokenFromAnother(newBranchID, generalToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, anotherBranchToken)
}

func TestHistoryGarbageCleanupInfo(t *testing.T) {
	testDomainID := "testDomainID"
	testWorkflowID := "testWorkflowID"
	testRunID := "testRunID"
	actualInfo := BuildHistoryGarbageCleanupInfo(testDomainID, testWorkflowID, testRunID)
	splitDomainID, splitWorkflowID, splitRunID, splitError := SplitHistoryGarbageCleanupInfo(actualInfo)
	assert.Equal(t, testDomainID, splitDomainID)
	assert.Equal(t, testWorkflowID, splitWorkflowID)
	assert.Equal(t, testRunID, splitRunID)
	assert.NoError(t, splitError)
	_, _, _, splitError = SplitHistoryGarbageCleanupInfo("invalidInfo")
	assert.Error(t, splitError)
}

func TestNewGetReplicationTasksFromDLQRequest(t *testing.T) {
	sourceCluster := "testSourceCluster"
	readLevel := int64(1)
	maxReadLevel := int64(2)
	batchSize := 10
	tasksRequest := NewGetReplicationTasksFromDLQRequest(sourceCluster, readLevel, maxReadLevel, batchSize, nil)
	assert.Equal(t, sourceCluster, tasksRequest.SourceClusterName)
}

func TestHasMoreRowsToDelete(t *testing.T) {
	assert.True(t, HasMoreRowsToDelete(10, 10))
	assert.False(t, HasMoreRowsToDelete(11, 10))
	assert.False(t, HasMoreRowsToDelete(9, 10))
	assert.False(t, HasMoreRowsToDelete(-1, 10))

}
func TestTransferTaskInfo(t *testing.T) {
	timeNow := time.Now()
	task := &TransferTaskInfo{
		DomainID:            "test-domain-id",
		WorkflowID:          "test-workflow-id",
		RunID:               "test-run-id",
		TargetDomainIDs:     map[string]struct{}{"test-target-domain-id": {}},
		TaskType:            TransferTaskTypeActivityTask,
		TaskID:              1,
		ScheduleID:          1,
		Version:             1,
		VisibilityTimestamp: timeNow,
	}
	expectedString := fmt.Sprintf("%#v", task)

	assert.Equal(t, "test-domain-id", task.GetDomainID())
	assert.Equal(t, "test-workflow-id", task.GetWorkflowID())
	assert.Equal(t, "test-run-id", task.GetRunID())
	assert.Equal(t, TransferTaskTypeActivityTask, task.GetTaskType())
	assert.Equal(t, int64(1), task.GetTaskID())
	assert.Equal(t, int64(1), task.GetVersion())
	assert.Equal(t, timeNow, task.GetVisibilityTimestamp())
	assert.Equal(t, map[string]struct{}{"test-target-domain-id": {}}, task.GetTargetDomainIDs())
	assert.Equal(t, expectedString, task.String())
}

func TestReplicationTaskInfo(t *testing.T) {
	task := &ReplicationTaskInfo{
		DomainID:     "test-domain-id",
		WorkflowID:   "test-workflow-id",
		RunID:        "test-run-id",
		TaskType:     ReplicationTaskTypeHistory,
		TaskID:       1,
		Version:      1,
		FirstEventID: 1,
		NextEventID:  2,
		ScheduledID:  3,
	}
	assert.Equal(t, "test-domain-id", task.GetDomainID())
	assert.Equal(t, "test-workflow-id", task.GetWorkflowID())
	assert.Equal(t, "test-run-id", task.GetRunID())
	assert.Equal(t, ReplicationTaskTypeHistory, task.GetTaskType())
	assert.Equal(t, int64(1), task.GetTaskID())
	assert.Equal(t, int64(1), task.GetVersion())
	assert.Equal(t, time.Time{}, task.GetVisibilityTimestamp())
}

func TestTimeTaskInfo(t *testing.T) {
	timeNow := time.Now()
	task := &TimerTaskInfo{
		VisibilityTimestamp: timeNow,
		TaskType:            4,
		TaskID:              3,
		Version:             2,
		RunID:               "test-run-id",
		DomainID:            "test-domain-id",
		WorkflowID:          "test-workflow-id",
	}
	expectedString := fmt.Sprintf(
		"{DomainID: %v, WorkflowID: %v, RunID: %v, VisibilityTimestamp: %v, TaskID: %v, TaskType: %v, TimeoutType: %v, EventID: %v, ScheduleAttempt: %v, Version: %v.}",
		task.DomainID, task.WorkflowID, task.RunID, task.VisibilityTimestamp, task.TaskID, task.TaskType, task.TimeoutType, task.EventID, task.ScheduleAttempt, task.Version,
	)
	assert.Equal(t, 4, task.GetTaskType())
	assert.Equal(t, int64(3), task.GetTaskID())
	assert.Equal(t, int64(2), task.GetVersion())
	assert.Equal(t, timeNow, task.GetVisibilityTimestamp())
	assert.Equal(t, "test-run-id", task.GetRunID())
	assert.Equal(t, "test-domain-id", task.GetDomainID())
	assert.Equal(t, "test-workflow-id", task.GetWorkflowID())
	assert.Equal(t, expectedString, task.String())
}

func TestShardInfoCopy(t *testing.T) {
	info := &ShardInfo{
		ShardID:                 1,
		RangeID:                 2,
		ClusterTransferAckLevel: map[string]int64{"test-cluster": 3},
		ClusterTimerAckLevel:    map[string]time.Time{"test-cluster": time.Now()},
		ClusterReplicationLevel: map[string]int64{"test-cluster": 4},
		ReplicationDLQAckLevel:  map[string]int64{"test-cluster": 5},
	}

	infoCopy := info.Copy()
	assert.Equal(t, info, infoCopy)
}

func TestSerializeAndDeserializeClusterConfigs(t *testing.T) {
	configs := []*ClusterReplicationConfig{
		{
			ClusterName: "test-cluster1",
		},
		{
			ClusterName: "test-cluster2",
		},
	}
	serializedResult := SerializeClusterConfigs(configs)
	deserializedResult := DeserializeClusterConfigs(serializedResult)

	assert.Equal(t, configs, deserializedResult)

}

func TestTimeStampConvertion(t *testing.T) {
	timeNow := time.Now()
	milisSecond := UnixNanoToDBTimestamp(timeNow.UnixNano())
	unixNanoTime := DBTimestampToUnixNano(milisSecond)
	assert.Equal(t, timeNow.UnixNano()/(1000*1000), unixNanoTime/(1000*1000)) // unixNano to milisSecond will result in info loss
}
