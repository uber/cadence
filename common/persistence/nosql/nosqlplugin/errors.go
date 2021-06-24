// Copyright (c) 2021 Uber Technologies, Inc.
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

package nosqlplugin

import "fmt"

// Condition Errors for NoSQL interfaces
type (
	// Only one of the fields must be non-nil
	WorkflowOperationConditionFailure struct {
		UnknownConditionFailureDetails   *string // return some info for logging
		ShardRangeIDNotMatch             *int64  // return the previous shardRangeID
		WorkflowExecutionAlreadyExists   *WorkflowExecutionAlreadyExists
		CurrentWorkflowConditionFailInfo *string // return the logging info if fail on condition of CurrentWorkflow
	}

	WorkflowExecutionAlreadyExists struct {
		RunID            string
		CreateRequestID  string
		State            int
		CloseStatus      int
		LastWriteVersion int64
		OtherInfo        string
	}

	TaskOperationConditionFailure struct {
		RangeID int64
		Details string // detail info for logging
	}

	ShardOperationConditionFailure struct {
		RangeID int64
		Details string // detail info for logging
	}

	ConditionFailure struct {
		componentName string
	}
)

var _ error = (*WorkflowOperationConditionFailure)(nil)

func (e *WorkflowOperationConditionFailure) Error() string {
	return "workflow operation condition failure"
}

var _ error = (*TaskOperationConditionFailure)(nil)

func (e *TaskOperationConditionFailure) Error() string {
	return "task operation condition failure"
}

var _ error = (*ShardOperationConditionFailure)(nil)

func (e *ShardOperationConditionFailure) Error() string {
	return "shard operation condition failure"
}

var _ error = (*ConditionFailure)(nil)

func NewConditionFailure(name string) error {
	return &ConditionFailure{
		componentName: name,
	}
}
func (e *ConditionFailure) Error() string {
	return fmt.Sprintf("%s operation condition failure", e.componentName)
}
