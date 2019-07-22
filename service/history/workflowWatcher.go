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

package history

import (
	"sync"
	"sync/atomic"

	"github.com/pborman/uuid"
	"github.com/uber/cadence/common/persistence"
)

type (
	// WorkflowWatcher is used to get updates on mutable state changes.
	WorkflowWatcher interface {
		Publish(*persistence.WorkflowMutableState)
		Subscribe() (string, <-chan struct{})
		Unsubscribe(string)
		LatestMutableState() *persistence.WorkflowMutableState
	}

	workflowWatcher struct {
		// An atomic is used to keep track of mutable state rather than pushing updates on a channel.
		// If a channel (buffered or unbuffered) is used, then at some point either blocking will occur or a mutable state update will be dropped.
		// Since clients only care about getting the latest mutable state, simplify notifying clients
		// of updates and providing the ability to get the latest satisfies all current use cases.
		latestMS atomic.Value
		lockableSubscribers
	}

	lockableSubscribers struct {
		sync.RWMutex
		subscribers map[string]chan struct{}
	}
)

func (ls *lockableSubscribers) add() (string, chan struct{}) {
	ls.Lock()
	defer ls.Unlock()
	id := uuid.New()
	ch := make(chan struct{}, 1)
	ls.subscribers[id] = ch
	return id, ch
}

func (ls *lockableSubscribers) remove(id string) {
	ls.Lock()
	defer ls.Unlock()
	delete(ls.subscribers, id)
}

func (ls *lockableSubscribers) notifyAll() {
	ls.RLock()
	defer ls.RUnlock()
	for _, ch := range ls.subscribers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// NewWorkflowWatcher returns new WorkflowWatcher.
// This should only be called from workflowExecutionContext.go.
func NewWorkflowWatcher() WorkflowWatcher {
	return &workflowWatcher{
		lockableSubscribers: lockableSubscribers{
			subscribers: make(map[string]chan struct{}),
		},
	}
}

// Publish is used to indicate a mutable state change has been successfully persisted.
func (w *workflowWatcher) Publish(latestMS *persistence.WorkflowMutableState) {
	if latestMS == nil {
		return
	}
	w.latestMS.Store(latestMS)
	w.notifyAll()
}

// Subscribe to updates to mutable state.
// Every time returned channel is received on it is guaranteed that at least one new update is available.
func (w *workflowWatcher) Subscribe() (string, <-chan struct{}) {
	return w.add()
}

// Unsubscribe from updates to mutable state.
func (w *workflowWatcher) Unsubscribe(id string) {
	w.remove(id)
}

// LatestMutableState returns the latest mutable state update which has been persisted.
func (w *workflowWatcher) LatestMutableState() *persistence.WorkflowMutableState {
	v := w.latestMS.Load()
	if v == nil {
		return nil
	}
	return v.(*persistence.WorkflowMutableState)
}
