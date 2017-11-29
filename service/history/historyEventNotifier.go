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
	"time"

	"github.com/pborman/uuid"
	gen "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/metrics"
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
		metrics metrics.Client
		// internal status indicator
		status int32
		// stop signal channel
		closeChan chan bool
		// this channel will never close
		eventsChan chan *historyEventNotification
		// function which calculate the shard ID from given workflow ID
		workflowIDToShardID func(string) int

		metadataLock       sync.RWMutex
		shardLocks         map[int]*sync.RWMutex
		shardEventsPubsubs map[int]map[workflowIdentifier]map[string]chan *historyEventNotification
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

func newHistoryEventNotifier(metrics metrics.Client, workflowIDToShardID func(string) int) *historyEventNotifierImpl {
	return &historyEventNotifierImpl{
		metrics:    metrics,
		status:     statusIdle,
		closeChan:  make(chan bool),
		eventsChan: make(chan *historyEventNotification, eventsChanSize),

		workflowIDToShardID: workflowIDToShardID,

		shardLocks:         make(map[int]*sync.RWMutex),
		shardEventsPubsubs: make(map[int]map[workflowIdentifier]map[string]chan *historyEventNotification),
	}
}

func (notifier *historyEventNotifierImpl) WatchHistoryEvent(
	identifier *workflowIdentifier) (string, chan *historyEventNotification, error) {

	shardID := notifier.workflowIDToShardID(identifier.workflowID)
	lock, eventsPubsubs := notifier.createOrGetShards(shardID)
	lock.Lock()
	defer lock.Unlock()

	subscribers, ok := eventsPubsubs[*identifier]
	if !ok {
		subscribers = make(map[string]chan *historyEventNotification)
		eventsPubsubs[*identifier] = subscribers
	}

	subscriberID := uuid.NewUUID().String()
	if _, ok := subscribers[subscriberID]; ok {
		// UUID collision
		return "", nil, &gen.InternalServiceError{
			Message: "Unable to watch on workflow execution.",
		}
	}

	channel := make(chan *historyEventNotification, 1)
	subscribers[subscriberID] = channel
	return subscriberID, channel, nil
}

func (notifier *historyEventNotifierImpl) UnwatchHistoryEvent(
	identifier *workflowIdentifier, subscriberID string) error {

	shardID := notifier.workflowIDToShardID(identifier.workflowID)
	lock, eventsPubsubs := notifier.createOrGetShards(shardID)
	lock.Lock()
	defer lock.Unlock()

	subscribers, ok := eventsPubsubs[*identifier]
	if !ok {
		return &gen.InternalServiceError{
			Message: "Unable to unwatch on workflow execution.",
		}
	}

	if _, ok := subscribers[subscriberID]; !ok {
		// cannot find the subscribe ID, which means there is a bug
		return &gen.EntityNotExistsError{
			Message: "Unable to unwatch on workflow execution.",
		}
	}

	delete(subscribers, subscriberID)

	if len(subscribers) == 0 {
		delete(eventsPubsubs, *identifier)
	}

	return nil
}

func (notifier *historyEventNotifierImpl) dispatchHistoryEventNotification(event *historyEventNotification) {
	identifier := &event.workflowIdentifier

	shardID := notifier.workflowIDToShardID(identifier.workflowID)
	lock, eventsPubsubs := notifier.createOrGetShards(shardID)
	lock.RLock()
	defer lock.RUnlock()

	subscribers, ok := eventsPubsubs[*identifier]
	if !ok {
		return
	}

	timer := notifier.metrics.StartTimer(metrics.HistoryEventNotificationFanoutLatency, metrics.HistoryEventNotificationFanoutLatency)
	defer timer.Stop()
	for _, channel := range subscribers {
		select {
		case channel <- event:
		default:
			// in case the channel is already filled with message
			// this should NOT happen, unless there is a bug or high load
		}
	}
}

func (notifier *historyEventNotifierImpl) enqueueHistoryEventNotification(event *historyEventNotification) {
	// set the timestamp just before enqueuing the event
	event.timestamp = time.Now()
	select {
	case notifier.eventsChan <- event:
	default:
		// in case the channel is already filled with message
		// this can be caused by high load
	}
}

func (notifier *historyEventNotifierImpl) dequeueHistoryEventNotifications() {
	for {
		// send out metrics about the current number of messages in flight
		notifier.metrics.UpdateGauge(metrics.HistoryEventNotificationScope,
			metrics.HistoryEventNotificationInFlightMessageGauge, float64(len(notifier.eventsChan)))
		select {
		case event := <-notifier.eventsChan:
			// send out metrics about message processing delay
			timeelapsed := time.Since(event.timestamp)
			notifier.metrics.RecordTimer(metrics.HistoryEventNotificationScope,
				metrics.HistoryEventNotificationQueuingLatency, timeelapsed)

			notifier.dispatchHistoryEventNotification(event)
		case <-notifier.closeChan:
			// shutdown
			return
		}
	}
}

func (notifier *historyEventNotifierImpl) createOrGetShards(
	shardID int) (*sync.RWMutex, map[workflowIdentifier]map[string]chan *historyEventNotification) {

	notifier.metadataLock.RLock()
	lock, ok := notifier.shardLocks[shardID]
	if ok {
		eventsPubsubs := notifier.shardEventsPubsubs[shardID]
		notifier.metadataLock.RUnlock()
		return lock, eventsPubsubs
	}
	notifier.metadataLock.RUnlock()

	notifier.metadataLock.Lock()
	defer notifier.metadataLock.Unlock()

	lock, ok = notifier.shardLocks[shardID]
	if !ok {
		lock = &sync.RWMutex{}
		notifier.shardLocks[shardID] = lock
	}

	eventsPubsubs, ok := notifier.shardEventsPubsubs[shardID]
	if !ok {
		eventsPubsubs = make(map[workflowIdentifier]map[string]chan *historyEventNotification)
		notifier.shardEventsPubsubs[shardID] = eventsPubsubs
	}
	return lock, eventsPubsubs
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
