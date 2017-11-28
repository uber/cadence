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
	gen "github.com/uber/cadence/.gen/go/shared"
)

const (
	eventsChanSize = 1000

	// used for workflow pubsub status
	statusIdle    int32 = 0
	statusStarted int32 = 1
	statusStopped int32 = 2
)

type (
	historyEventNotifierImpl struct {
		status int32
		// stop signal channel
		closeChan chan bool
		// this channel will never close
		eventsChan chan *historyEventNotification

		sync.Mutex
		// map of workflow identifier to map of subscriber ID to channel
		eventsPubsubs map[workflowIdentifier]map[string]chan *historyEventNotification
	}
)

var _ historyEventNotifier = (*historyEventNotifierImpl)(nil)

func newWorkflowIdentifier(domainID string, workflowExecution *gen.WorkflowExecution) *workflowIdentifier {
	return &workflowIdentifier{
		domainID:   domainID,
		workflowID: *workflowExecution.WorkflowId,
		runID:      *workflowExecution.RunId,
	}
}

func newHistoryEventNotification(domainID string, workflowExecution *gen.WorkflowExecution,
	nextEventID int64, isWorkflowRunning bool) *historyEventNotification {
	return &historyEventNotification{
		workflowIdentifier: workflowIdentifier{
			domainID:   domainID,
			workflowID: *workflowExecution.WorkflowId,
			runID:      *workflowExecution.RunId,
		},
		nextEventID:       nextEventID,
		isWorkflowRunning: isWorkflowRunning,
	}
}

func newHistoryEventNotifier() *historyEventNotifierImpl {
	return &historyEventNotifierImpl{
		status:        statusIdle,
		closeChan:     make(chan bool),
		eventsChan:    make(chan *historyEventNotification, eventsChanSize),
		eventsPubsubs: make(map[workflowIdentifier]map[string]chan *historyEventNotification),
	}
}

func (notifier *historyEventNotifierImpl) WatchHistoryEvent(domainID string,
	execution *gen.WorkflowExecution) (string, chan *historyEventNotification, error) {

	identifier := newWorkflowIdentifier(domainID, execution)

	notifier.Lock()
	defer notifier.Unlock()

	eventsPubsubs, ok := notifier.eventsPubsubs[*identifier]
	if !ok {
		eventsPubsubs = make(map[string]chan *historyEventNotification)
		notifier.eventsPubsubs[*identifier] = eventsPubsubs
	}

	subscriberID := uuid.NewUUID().String()
	if _, ok := eventsPubsubs[subscriberID]; ok {
		// UUID collision
		return "", nil, &gen.InternalServiceError{
			Message: "Unable to watch on workflow execution.",
		}
	}

	channel := make(chan *historyEventNotification, 1)
	eventsPubsubs[subscriberID] = channel
	return subscriberID, channel, nil
}

func (notifier *historyEventNotifierImpl) UnwatchHistoryEvent(domainID string,
	execution *gen.WorkflowExecution, subscriberID string) error {

	identifier := newWorkflowIdentifier(domainID, execution)

	notifier.Lock()
	defer notifier.Unlock()

	eventsPubsubs, ok := notifier.eventsPubsubs[*identifier]
	if !ok {
		return &gen.InternalServiceError{
			Message: "Unable to unwatch on workflow execution.",
		}
	}

	if _, ok := eventsPubsubs[subscriberID]; !ok {
		// cannot find the subscribe ID, which means there is a bug
		return &gen.EntityNotExistsError{
			Message: "Unable to unwatch on workflow execution.",
		}
	}

	delete(eventsPubsubs, subscriberID)

	if len(eventsPubsubs) == 0 {
		delete(notifier.eventsPubsubs, *identifier)
	}

	return nil
}

func (notifier *historyEventNotifierImpl) dispatchHistoryEventNotification(event *historyEventNotification) {
	identifier := &event.workflowIdentifier

	notifier.Lock()
	defer notifier.Unlock()

	eventsPubsubs, ok := notifier.eventsPubsubs[*identifier]
	if !ok {
		return
	}

	for _, channel := range eventsPubsubs {
		select {
		case channel <- event:
		default:
			// in case the channel is already filled with message
			// this should NOT happen, unless there is a bug or high load
		}
	}
}

func (notifier *historyEventNotifierImpl) enqueueHistoryEventNotification(event *historyEventNotification) {
	select {
	case notifier.eventsChan <- event:
	default:
		// in case the channel is already filled with message
		// this can be caused by high load
	}
}

func (notifier *historyEventNotifierImpl) dequeueHistoryEventNotifications() {
	for {
		select {
		case event := <-notifier.eventsChan:
			notifier.dispatchHistoryEventNotification(event)
		case <-notifier.closeChan:
			// shutdown
			return
		}
	}
}

func (notifier *historyEventNotifierImpl) Start() {
	if !atomic.CompareAndSwapInt32(&notifier.status, statusIdle, statusStarted) {
		return
	}
	go notifier.dequeueHistoryEventNotifications()
}

func (notifier *historyEventNotifierImpl) Stop() {
	if !atomic.CompareAndSwapInt32(&notifier.status, statusStarted, statusStopped) {
		return
	}
	close(notifier.closeChan)
}

func (notifier *historyEventNotifierImpl) NotifyNewHistoryEvent(event *historyEventNotification) {
	notifier.enqueueHistoryEventNotification(event)
}
