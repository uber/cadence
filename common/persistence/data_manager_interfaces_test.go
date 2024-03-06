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

func TestTaskCommonMethods(t *testing.T) {
	timeNow := time.Now()
	tasks := []interface{}{
		&ActivityTask{Version: 1, TaskID: 1, VisibilityTimestamp: timeNow},
		&DecisionTask{Version: 1, TaskID: 1, VisibilityTimestamp: timeNow},
		&RecordWorkflowStartedTask{Version: 1, TaskID: 1, VisibilityTimestamp: timeNow},
		&ResetWorkflowTask{Version: 1, TaskID: 1, VisibilityTimestamp: timeNow},
		&CloseExecutionTask{Version: 1, TaskID: 1, VisibilityTimestamp: timeNow},
		&DeleteHistoryEventTask{Version: 1, TaskID: 1, VisibilityTimestamp: timeNow},
		&DecisionTimeoutTask{Version: 1, TaskID: 1, VisibilityTimestamp: timeNow},
		&ActivityTimeoutTask{Version: 1, TaskID: 1, VisibilityTimestamp: timeNow},
		&UserTimerTask{Version: 1, TaskID: 1, VisibilityTimestamp: timeNow},
	}

	for _, task := range tasks {
		switch ty := task.(type) {
		case *ActivityTask:
			assert.Equal(t, TransferTaskTypeActivityTask, ty.GetType())
		case *DecisionTask:
			assert.Equal(t, TransferTaskTypeDecisionTask, ty.GetType())
		case *RecordWorkflowStartedTask:
			assert.Equal(t, TransferTaskTypeRecordWorkflowStarted, ty.GetType())
		case *ResetWorkflowTask:
			assert.Equal(t, TransferTaskTypeResetWorkflow, ty.GetType())
		case *CloseExecutionTask:
			assert.Equal(t, TransferTaskTypeCloseExecution, ty.GetType())
		case *DeleteHistoryEventTask:
			assert.Equal(t, TaskTypeDeleteHistoryEvent, ty.GetType())
		case *DecisionTimeoutTask:
			assert.Equal(t, TaskTypeDecisionTimeout, ty.GetType())
		case *ActivityTimeoutTask:
			assert.Equal(t, TaskTypeActivityTimeout, ty.GetType())
		case *UserTimerTask:
			assert.Equal(t, TaskTypeUserTimer, ty.GetType())
		default:
			t.Fatalf("Unhandled task type: %T", t)
		}

		// Test version methods
		assert.Equal(t, int64(1), task.(Task).GetVersion())
		task.(Task).SetVersion(2)
		assert.Equal(t, int64(2), task.(Task).GetVersion())

		// Test TaskID methods
		assert.Equal(t, int64(1), task.(Task).GetTaskID())
		task.(Task).SetTaskID(2)
		assert.Equal(t, int64(2), task.(Task).GetTaskID())

		// Test VisibilityTimestamp methods
		assert.Equal(t, timeNow, task.(Task).GetVisibilityTimestamp())
		newTime := timeNow.Add(time.Second)
		task.(Task).SetVisibilityTimestamp(newTime)
		assert.Equal(t, newTime, task.(Task).GetVisibilityTimestamp())
	}
}
