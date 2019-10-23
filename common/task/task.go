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

package task

import (
	"time"
)

type (
	// Info is the interface for describing a task
	Info interface {
		GetVersion() int64
		GetTaskID() int64
		GetTaskType() int
		GetVisibilityTimestamp() time.Time
		GetWorkflowID() string
		GetRunID() string
		GetDomainID() string
	}

	// SequenceIDs is a list of task SequenceID
	SequenceIDs []SequenceID

	// SequenceID determines the total order of tasks
	SequenceID struct {
		VisibilityTimestamp time.Time
		TaskID              int64
	}

	// SequentialTask is the interface for tasks which should be executed sequentially
	// TODO: rename this interface to Task
	SequentialTask interface {
		// Execute process this task
		Execute() error
		// HandleErr handle the error returned by Execute
		HandleErr(err error) error
		// RetryErr check whether to retry after HandleErr(Execute())
		RetryErr(err error) bool
		// Ack marks the task as successful completed
		Ack()
		// Nack marks the task as unsuccessful completed
		Nack()
	}
)

// ToSequenceID generates SequenceID from Info
func ToSequenceID(taskInfo Info) SequenceID {
	return SequenceID{
		VisibilityTimestamp: taskInfo.GetVisibilityTimestamp(),
		TaskID:              taskInfo.GetTaskID(),
	}
}

// Len implements sort.Interace
func (t SequenceIDs) Len() int {
	return len(t)
}

// Swap implements sort.Interface.
func (t SequenceIDs) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Less implements sort.Interface
func (t SequenceIDs) Less(i, j int) bool {
	return compareTaskSequenceIDLess(&t[i], &t[j])
}

func compareTaskSequenceIDLess(
	first *SequenceID,
	second *SequenceID,
) bool {
	if first.VisibilityTimestamp.Before(second.VisibilityTimestamp) {
		return true
	}
	if first.VisibilityTimestamp.Equal(second.VisibilityTimestamp) {
		return first.TaskID < second.TaskID
	}
	return false
}
