// Copyright (c) 2020 Uber Technologies, Inc.
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

//go:generate mockgen -copyright_file ../../../LICENSE -package $GOPACKAGE -source $GOFILE -destination interface_mock.go -self_package github.com/uber/cadence/service/history/processing

package processing

import (
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/task"
)

type (
	// DomainFilter filters domain
	DomainFilter struct {
		Domains map[string]struct{}
		// by default, a DomainFilter matches domains listed in the Domains field
		// if reverseMatch is true then the DomainFilter matches domains that are
		// not in the Domains field.
		ReverseMatch bool
	}

	// JobInfo indicates the scope of a Job and its progress
	JobInfo interface {
		QueueID() int
		MinLevel() task.Key
		MaxLevel() task.Key
		AckLevel() task.Key
		ReadLevel() task.Key
		DomainFilter() DomainFilter
	}

	// Job is responsible for keeping track of the state of tasks
	// within the scope defined by its JobInfo; it can also be split
	// into multiple Jobs with non-overlapping scope or be merged with
	// another Job
	Job interface {
		Info() JobInfo
		Split(JobSplitPolicy) []Job
		Merge(Job) []Job
		AddTasks(map[task.Key]task.Task)
		UpdateAckLevel()
		// TODO: add Offload() method
	}

	// JobSplitPolicy determins if a Job should be split
	// into multiple Jobs
	JobSplitPolicy interface {
		Evaluate(Job) []JobInfo
	}

	// JobQueue manages a list of non-overlapping Jobs
	// and keep track of the current active Job
	JobQueue interface {
		Info() []JobInfo
		Split(JobSplitPolicy) []Job
		Merge([]Job)
		AddTasks(map[task.Key]task.Task)
		ActiveJob() Job
		// TODO: add Offload() method
	}

	// JobQueueManager manages a set of JobQueue and
	// controls the event loop for loading tasks, updating
	// and persisting JobInfo, spliting/merging Jobs, etc.
	JobQueueManager interface {
		common.Daemon

		NotifyNewTasks([]persistence.Task)
	}
)
