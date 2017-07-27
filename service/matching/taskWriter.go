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

package matching

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/uber-common/bark"
	s "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/logging"
	"github.com/uber/cadence/common/persistence"
	"sync"
)

const (
	outstandingTaskAppendsThreshold = 250
	maxTaskBatchSize                = 100
)

type (
	writeTaskResponse struct {
		err                 error
		persistenceResponse *persistence.CreateTasksResponse
	}

	writeTaskRequest struct {
		execution  *s.WorkflowExecution
		taskInfo   *persistence.TaskInfo
		rangeID    int64
		responseCh chan<- *writeTaskResponse
	}

	// taskWriter writes tasks sequentially to persistence
	taskWriter struct {
		tlMgr        *taskListManagerImpl
		taskListID   *taskListID
		taskManager  persistence.TaskManager
		appendCh     chan *writeTaskRequest
		maxReadLevel int64
		stopped      int64 // set to 1 if the writer is stopped or is shutting down
		logger       bark.Logger
		stopCh       chan struct{}  // shutdown signal for all routines in this class
		writerStopCh chan struct{}  // shutdown signal for the writerLoop()
		writerWG     sync.WaitGroup // done signal from the writerLoop()
	}
)

// errShutdown indicates that the task list is shutting down
var errShutdown = errors.New("task list shutting down")

func newTaskWriter(tlMgr *taskListManagerImpl) *taskWriter {
	return &taskWriter{
		tlMgr:        tlMgr,
		taskListID:   tlMgr.taskListID,
		taskManager:  tlMgr.engine.taskManager,
		stopCh:       make(chan struct{}),
		writerStopCh: make(chan struct{}),
		appendCh:     make(chan *writeTaskRequest, outstandingTaskAppendsThreshold),
		logger:       tlMgr.logger,
	}
}

func (w *taskWriter) Start() {
	w.maxReadLevel = w.tlMgr.getTaskSequenceNumber() - 1
	w.writerWG.Add(1)
	go w.taskWriterLoop()
}

// Stop stops the taskWriter
func (w *taskWriter) Stop() {

	// Shutdown needs to happen in a specific order to
	// avoid duplicate tasks in the task list. Even if
	// there is a dup, its not the end of the world, but
	// its best to avoid it

	// step 1: set stopped=1 and stop accepting new tasks in appendTask()
	// step 2: stop the background routine that writes tasks to cassandra
	// step 3: wait for the background routine to exit
	// step 4: stop everything and return
	if atomic.CompareAndSwapInt64(&w.stopped, 0, 1) { // step1
		close(w.writerStopCh) // step2
		w.writerWG.Wait()     // step3
		close(w.stopCh)       // step4
	}
}

func (w *taskWriter) isStopped() bool {
	return atomic.LoadInt64(&w.stopped) == 1
}

func (w *taskWriter) appendTask(execution *s.WorkflowExecution,
	taskInfo *persistence.TaskInfo, rangeID int64) (*persistence.CreateTasksResponse, error) {

	if w.isStopped() {
		return nil, errShutdown
	}

	ch := make(chan *writeTaskResponse)
	req := &writeTaskRequest{
		execution:  execution,
		taskInfo:   taskInfo,
		rangeID:    rangeID,
		responseCh: ch,
	}

	select {
	case w.appendCh <- req:
		select {
		case r := <-ch:
			return r.persistenceResponse, r.err
		case <-w.stopCh:
			// if we are shutting down, this request will never make
			// it to cassandra, just bail out and fail this request
			return nil, errShutdown
		}
	default: // channel is full, throttle
		return nil, createServiceBusyError()
	}
}

func (w *taskWriter) GetMaxReadLevel() int64 {
	return atomic.LoadInt64(&w.maxReadLevel)
}

func (w *taskWriter) taskWriterLoop() {

	defer w.writerWG.Done()

writerLoop:
	for {

		select {
		case <-w.writerStopCh:
			// we don't close the appendCh here
			// because that can cause on a send on closed
			// channel panic on the appendTask()
			break writerLoop
		default:
			select {
			case request := <-w.appendCh:
				{
					// read a batch of requests from the channel
					reqs := []*writeTaskRequest{request}
					reqs = w.getWriteBatch(reqs)
					batchSize := len(reqs)

					maxReadLevel := int64(0)

					taskIDs, err := w.tlMgr.newTaskIDs(batchSize)
					if err != nil {
						w.sendWriteResponse(reqs, err, nil)
						continue writerLoop
					}

					tasks := []*persistence.CreateTaskInfo{}
					rangeID := int64(0)
					for i, req := range reqs {
						tasks = append(tasks, &persistence.CreateTaskInfo{
							TaskID:    taskIDs[i],
							Execution: *req.execution,
							Data:      req.taskInfo,
						})
						if req.rangeID > rangeID {
							rangeID = req.rangeID // use the maximum rangeID provided for the write operation
						}
						maxReadLevel = taskIDs[i]
					}

					w.tlMgr.persistenceLock.Lock()
					r, err := w.taskManager.CreateTasks(&persistence.CreateTasksRequest{
						DomainID:     w.taskListID.domainID,
						TaskList:     w.taskListID.taskListName,
						TaskListType: w.taskListID.taskType,
						Tasks:        tasks,
						// Note that newTaskID could increment range, so rangeID parameter
						// might be out of sync. This is OK as caller can just retry.
						RangeID: rangeID,
					})
					w.tlMgr.persistenceLock.Unlock()

					if err != nil {
						logging.LogPersistantStoreErrorEvent(w.logger, logging.TagValueStoreOperationCreateTask, err,
							fmt.Sprintf("{taskID: [%v, %v], taskType: %v, taskList: %v}",
								taskIDs[0], taskIDs[batchSize-1], w.taskListID.taskType, w.taskListID.taskListName))
					}

					// Update the maxReadLevel after the writes are completed.
					if maxReadLevel > 0 {
						atomic.StoreInt64(&w.maxReadLevel, maxReadLevel)
					}

					w.sendWriteResponse(reqs, err, r)
				}
			case <-w.writerStopCh:
				break writerLoop
			}
		}
	}
}

func (w *taskWriter) getWriteBatch(reqs []*writeTaskRequest) []*writeTaskRequest {
readLoop:
	for i := 0; i < maxTaskBatchSize; i++ {
		select {
		case req := <-w.appendCh:
			reqs = append(reqs, req)
		default: // channel is empty, don't block
			break readLoop
		}
	}
	return reqs
}

func (w *taskWriter) sendWriteResponse(reqs []*writeTaskRequest,
	err error, persistenceResponse *persistence.CreateTasksResponse) {
	for _, req := range reqs {
		resp := &writeTaskResponse{
			err:                 err,
			persistenceResponse: persistenceResponse,
		}

		req.responseCh <- resp
	}
}
