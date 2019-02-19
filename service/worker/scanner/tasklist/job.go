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

package tasklist

import (
	"container/list"
	"sync"

	p "github.com/uber/cadence/common/persistence"
)

type (
	// threadSafeList is a wrapper around list.List which
	// guarantees thread safety
	threadSafeList struct {
		sync.Mutex
		list *list.List
	}

	jobInput struct {
		taskListKey
		taskListState
	}

	job struct {
		input jobInput
		wg    *sync.WaitGroup // wait group that's set to done after the job is processed
	}

	jobQueue struct {
		jobs  chan *job
		stopC chan struct{}
	}
)

// newJobQueue returns an instance of task lists worker job queue
func newJobQueue(size int, stopC chan struct{}) *jobQueue {
	return &jobQueue{
		jobs:  make(chan *job, size),
		stopC: stopC,
	}
}

func newJob(info *p.TaskListInfo, wg *sync.WaitGroup) *job {
	return &job{
		input: jobInput{
			taskListKey: taskListKey{
				DomainID: info.DomainID,
				Name:     info.Name,
				TaskType: info.TaskType,
			},
			taskListState: taskListState{
				rangeID:     info.RangeID,
				lastUpdated: info.LastUpdated,
			},
		},
		wg: wg,
	}
}

func (jq *jobQueue) add(job *job) bool {
	select {
	case jq.jobs <- job:
		return true
	case <-jq.stopC:
		return false
	}
}

func (jq *jobQueue) remove() (*job, bool) {
	select {
	case job, ok := <-jq.jobs:
		if !ok {
			return nil, false
		}
		return job, true
	case <-jq.stopC:
		return nil, false
	}
}

// newThreadSafeList returns a new thread safe linked list
func newThreadSafeList() *threadSafeList {
	return &threadSafeList{
		list: list.New(),
	}
}

func (tl *threadSafeList) add(elem interface{}) {
	tl.Lock()
	defer tl.Unlock()
	tl.list.PushBack(elem)
}

func (tl *threadSafeList) len() int {
	tl.Lock()
	defer tl.Unlock()
	return tl.list.Len()
}

func (tl *threadSafeList) remove() *list.Element {
	tl.Lock()
	defer tl.Unlock()
	e := tl.list.Front()
	if e != nil {
		tl.list.Remove(e)
	}
	return e
}
