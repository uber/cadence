// Copyright (c) 2017-2021 Uber Technologies Ins.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package queue

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

type (
	crossClusterQueueProcessorBaseSuite struct {
		suite.Suite
		*require.Assertions

		controller        *gomock.Controller
		mockShard         *shard.TestContext
		mockTaskProcessor *task.MockProcessor

		logger        log.Logger
		metricsClient metrics.Client
		metricsScope  metrics.Scope
	}
)

func TestCrossClusterQueueProcessorBaseSuite(t *testing.T) {
	s := new(crossClusterQueueProcessorBaseSuite)
	suite.Run(t, s)
}

func (s *crossClusterQueueProcessorBaseSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockShard = shard.NewTestContext(
		s.controller,
		&persistence.ShardInfo{
			ShardID:          10,
			RangeID:          1,
			TransferAckLevel: 0,
		},
		config.NewForTest(),
	)
	s.mockShard.Resource.DomainCache.EXPECT().GetDomainName(gomock.Any()).Return(constants.TestDomainName, nil).AnyTimes()
	s.mockTaskProcessor = task.NewMockProcessor(s.controller)

	s.logger = loggerimpl.NewLoggerForTest(s.Suite)
	s.metricsClient = metrics.NewClient(tally.NoopScope, metrics.History)
	s.metricsScope = s.metricsClient.Scope(metrics.CrossClusterQueueProcessorScope)
}

func (s *crossClusterQueueProcessorBaseSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
}

func (s *crossClusterQueueProcessorBaseSuite) TestReadTasks_Success() {
	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		testKey{ID: 1},
		testKey{ID: 1},
		testKey{ID: 10},
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)
	taskInfos := []*persistence.CrossClusterTaskInfo{
		{
			TaskID:   1,
			DomainID: "testDomain1",
		},
		{
			TaskID:   10,
			DomainID: "testDomain2",
		},
		{
			TaskID:   100,
			DomainID: "testDomain1",
		},
	}
	readLevel := newCrossClusterTaskKey(1)
	maxReadLevel := newCrossClusterTaskKey(10)
	mockExecutionManager := s.mockShard.Resource.ExecutionMgr
	mockExecutionManager.On("GetCrossClusterTasks", mock.Anything, &persistence.GetCrossClusterTasksRequest{
		TargetCluster: clusterName,
		ReadLevel:     readLevel.(crossClusterTaskKey).taskID,
		MaxReadLevel:  maxReadLevel.(crossClusterTaskKey).taskID,
		BatchSize:     s.mockShard.GetConfig().CrossClusterTaskBatchSize(),
	}).Return(&persistence.GetCrossClusterTasksResponse{
		Tasks:         taskInfos,
		NextPageToken: nil,
	}, nil).Once()

	tasks, more, err := processorBase.readTasks(readLevel, maxReadLevel)
	s.NoError(err)
	s.Equal(taskInfos, tasks)
	s.False(more)
}

func (s *crossClusterQueueProcessorBaseSuite) TestReadTasks_Error() {
	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		testKey{ID: 1},
		testKey{ID: 1},
		testKey{ID: 10},
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)
	readLevel := newCrossClusterTaskKey(1)
	maxReadLevel := newCrossClusterTaskKey(10)
	mockExecutionManager := s.mockShard.Resource.ExecutionMgr
	mockExecutionManager.On("GetCrossClusterTasks", mock.Anything, &persistence.GetCrossClusterTasksRequest{
		TargetCluster: clusterName,
		ReadLevel:     readLevel.(crossClusterTaskKey).taskID,
		MaxReadLevel:  maxReadLevel.(crossClusterTaskKey).taskID,
		BatchSize:     s.mockShard.GetConfig().CrossClusterTaskBatchSize(),
	}).Return(nil, errors.New("test")).Once()

	tasks, more, err := processorBase.readTasks(readLevel, maxReadLevel)
	s.Error(err)
	s.Equal(0, len(tasks))
	s.False(more)
}

func (s *crossClusterQueueProcessorBaseSuite) TestPollTasks() {
	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		testKey{ID: 1},
		testKey{ID: 1},
		testKey{ID: 10},
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)

	// case 1: normal case, task need to be returned
	taskID1 := int64(2)
	task1 := task.NewMockCrossClusterTask(s.controller)
	task1.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	task1.EXPECT().GetTaskID().Return(taskID1).AnyTimes()
	task1.EXPECT().GetCrossClusterRequest().Return(&types.CrossClusterTaskRequest{}, nil).Times(1)
	// case 2: task is no-op, doesn't need to be returned, should be removed from readyForPoll
	taskID2 := int64(3)
	task2 := task.NewMockCrossClusterTask(s.controller)
	task2.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	task2.EXPECT().GetTaskID().Return(taskID2).AnyTimes()
	task2.EXPECT().GetCrossClusterRequest().Return(nil, nil).Times(1)
	// case 3: failed to get request, should retry on next poll
	taskID3 := int64(4)
	task3 := task.NewMockCrossClusterTask(s.controller)
	task3.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	task3.EXPECT().GetTaskID().Return(taskID3).AnyTimes()
	task3.EXPECT().GetCrossClusterRequest().Return(nil, errors.New("some random error")).Times(1)
	task3.EXPECT().IsValid().Return(true).Times(1)
	// case 4: invalid task
	taskID4 := int64(5)
	task4 := task.NewMockCrossClusterTask(s.controller)
	task4.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	task4.EXPECT().GetTaskID().Return(taskID4).AnyTimes()
	task4.EXPECT().GetCrossClusterRequest().Return(nil, errors.New("task is invalid")).Times(1)
	task4.EXPECT().IsValid().Return(false).Times(1)
	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(true, nil).Times(1)
	newTaskMap := map[task.Key]task.Task{
		testKey{ID: int(taskID1)}: task1,
		testKey{ID: int(taskID2)}: task2,
		testKey{ID: int(taskID3)}: task3,
		testKey{ID: int(taskID4)}: task4,
	}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, testKey{ID: 10})
	for _, task := range newTaskMap {
		processorBase.readyForPollTasks.Put(task.GetTaskID(), task)
	}

	tasks := processorBase.pollTasks(context.Background())
	s.Equal(1, len(tasks))                            // only task1 needs to be returned
	s.Equal(2, processorBase.readyForPollTasks.Len()) // task 1 and 3
}

func (s *crossClusterQueueProcessorBaseSuite) TestPollTasks_ContextCancelled() {
	processingQueueState := newProcessingQueueState(
		0,
		testKey{ID: 1},
		testKey{ID: 1},
		testKey{ID: 100},
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		"test",
		nil,
		nil,
		nil,
	)

	pollCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	numPolledTasks := 5
	totalTasks := 2 * numPolledTasks
	taskMap := make(map[task.Key]task.Task)
	for i := 0; i != totalTasks; i++ {
		currentIdx := i
		taskID := int64(i + 2)
		task := task.NewMockCrossClusterTask(s.controller)
		task.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
		task.EXPECT().GetTaskID().Return(taskID).AnyTimes()
		task.EXPECT().GetCrossClusterRequest().DoAndReturn(
			func() (*types.CrossClusterTaskRequest, error) {
				if currentIdx == numPolledTasks-1 {
					cancel()
				}
				return &types.CrossClusterTaskRequest{}, nil
			},
		).AnyTimes()
		taskMap[testKey{ID: int(taskID)}] = task
		processorBase.readyForPollTasks.Put(taskID, task)
	}
	processorBase.processingQueueCollections[0].AddTasks(taskMap, testKey{ID: 100})

	tasks := processorBase.pollTasks(pollCtx)
	s.Equal(numPolledTasks, len(tasks))
	s.Equal(totalTasks, processorBase.readyForPollTasks.Len())
}

func (s *crossClusterQueueProcessorBaseSuite) TestPollTasks_FetchBatchSizeLimit() {
	fetchBatchSize := 10
	s.mockShard.GetConfig().CrossClusterTaskFetchBatchSize = dynamicconfig.GetIntPropertyFilteredByShardID(fetchBatchSize)
	numReadyTasks := fetchBatchSize * 2

	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		testKey{ID: 0},
		testKey{ID: 0},
		testKey{ID: numReadyTasks},
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)

	readyTasksMap := make(map[task.Key]task.Task)
	for i := 0; i != numReadyTasks; i++ {
		taskID := i + 1
		readyTask := task.NewMockCrossClusterTask(s.controller)
		readyTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
		readyTask.EXPECT().GetTaskID().Return(int64(taskID)).AnyTimes()
		readyTask.EXPECT().IsReadyForPoll().Return(true).AnyTimes()
		readyTask.EXPECT().GetCrossClusterRequest().Return(&types.CrossClusterTaskRequest{}, nil).AnyTimes()
		readyTasksMap[testKey{ID: taskID}] = readyTask
	}
	processorBase.processingQueueCollections[0].AddTasks(readyTasksMap, testKey{ID: numReadyTasks})
	for _, task := range readyTasksMap {
		processorBase.readyForPollTasks.Put(task.GetTaskID(), task)
	}

	tasks := processorBase.pollTasks(context.Background())
	s.Len(tasks, fetchBatchSize)
	s.Equal(numReadyTasks, processorBase.readyForPollTasks.Len()) // fetched tasks are still available for poll
}

func (s *crossClusterQueueProcessorBaseSuite) TestUpdateTask_Success() {
	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(10),
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)

	taskID := int64(2)
	crossClusterTask := task.NewMockCrossClusterTask(s.controller)
	crossClusterTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	crossClusterTask.EXPECT().GetTaskID().Return(taskID).AnyTimes()
	crossClusterTask.EXPECT().RecordResponse(gomock.Any()).Return(nil).Times(1)
	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(true, nil).Times(1)
	newTaskMap := map[task.Key]task.Task{newCrossClusterTaskKey(taskID): crossClusterTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(taskID+1))
	for _, task := range newTaskMap {
		processorBase.readyForPollTasks.Put(task.GetTaskID(), task)
	}

	err := processorBase.recordResponse(&types.CrossClusterTaskResponse{
		TaskID: taskID,
	})
	s.NoError(err)
	s.False(processorBase.readyForPollTasks.Contains(taskID))
	s.Equal(0, processorBase.redispatcher.Size())
}

func (s *crossClusterQueueProcessorBaseSuite) TestUpdateTask_UpdateTask_Fail() {
	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(10),
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)

	taskID := int64(2)
	crossClusterTask := task.NewMockCrossClusterTask(s.controller)
	crossClusterTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	crossClusterTask.EXPECT().GetTaskID().Return(taskID).AnyTimes()
	crossClusterTask.EXPECT().GetWorkflowID().Return("workflowID").AnyTimes()
	crossClusterTask.EXPECT().GetRunID().Return("runID").AnyTimes()
	crossClusterTask.EXPECT().RecordResponse(gomock.Any()).Return(errors.New("test")).Times(1)
	newTaskMap := map[task.Key]task.Task{newCrossClusterTaskKey(taskID): crossClusterTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(taskID+1))
	for _, task := range newTaskMap {
		processorBase.readyForPollTasks.Put(task.GetTaskID(), task)
	}

	err := processorBase.recordResponse(&types.CrossClusterTaskResponse{
		TaskID: taskID,
	})
	s.Error(err)
	s.True(processorBase.readyForPollTasks.Contains(taskID))
	s.Equal(0, processorBase.redispatcher.Size())
}

func (s *crossClusterQueueProcessorBaseSuite) TestUpdateTask_SubmitTask_Redispatch() {
	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(10),
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)

	taskID := int64(2)
	crossClusterTask := task.NewMockCrossClusterTask(s.controller)
	crossClusterTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	crossClusterTask.EXPECT().GetTaskID().Return(taskID).AnyTimes()
	crossClusterTask.EXPECT().RecordResponse(gomock.Any()).Return(nil).Times(1)
	crossClusterTask.EXPECT().Priority().Return(0).AnyTimes()
	crossClusterTask.EXPECT().GetAttempt().Return(0).Times(1)
	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(false, errors.New("test")).Times(1)
	newTaskMap := map[task.Key]task.Task{newCrossClusterTaskKey(taskID): crossClusterTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(taskID+1))
	for _, task := range newTaskMap {
		processorBase.readyForPollTasks.Put(task.GetTaskID(), task)
	}

	err := processorBase.recordResponse(&types.CrossClusterTaskResponse{
		TaskID: taskID,
	})
	s.NoError(err)
	s.False(processorBase.readyForPollTasks.Contains(taskID))
	s.Equal(1, processorBase.redispatcher.Size())
}

func (s *crossClusterQueueProcessorBaseSuite) TestProcessQueueCollections_NoNextPage_WithNextQueue() {
	clusterName := "test"
	queueLevel := 0
	ackLevel := newCrossClusterTaskKey(0)
	maxLevel := newCrossClusterTaskKey(1000)
	processingQueueStates := []ProcessingQueueState{
		NewProcessingQueueState(
			queueLevel,
			ackLevel,
			maxLevel,
			NewDomainFilter(map[string]struct{}{"testDomain1": {}}, false),
		),
		NewProcessingQueueState(
			queueLevel,
			newCrossClusterTaskKey(1000),
			newCrossClusterTaskKey(10000),
			NewDomainFilter(map[string]struct{}{"testDomain1": {}}, false),
		),
	}
	updateMaxReadLevel := func() task.Key {
		return newCrossClusterTaskKey(10000)
	}

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		processingQueueStates,
		clusterName,
		updateMaxReadLevel,
		nil,
		nil,
	)

	taskInfos := []*persistence.CrossClusterTaskInfo{
		{
			TaskID:   1,
			TaskType: 1,
			DomainID: "testDomain1",
		},
		{
			TaskID:   10,
			TaskType: 1,
			DomainID: "testDomain2",
		},
		{
			TaskID:   100,
			TaskType: 1,
			DomainID: "testDomain1",
		},
	}
	mockExecutionManager := s.mockShard.Resource.ExecutionMgr
	mockExecutionManager.On("GetCrossClusterTasks", mock.Anything, mock.Anything).Return(
		&persistence.GetCrossClusterTasksResponse{
			Tasks:         taskInfos,
			NextPageToken: nil,
		}, nil).Once()

	processorBase.options.MaxPendingTaskSize = dynamicconfig.GetIntPropertyFn(100)
	processorBase.processQueueCollections()

	queueCollection := processorBase.processingQueueCollections[0]
	s.NotNil(queueCollection.ActiveQueue())
	s.True(taskKeyEquals(maxLevel, queueCollection.Queues()[0].State().ReadLevel()))
	s.Equal(2, processorBase.readyForPollTasks.Len()) // only 2 tasks belong to domain1

	s.True(processorBase.shouldProcess[queueLevel])
	select {
	case <-processorBase.processCh:
	default:
		s.Fail("processCh should be unblocked")
	}
}

func (s *crossClusterQueueProcessorBaseSuite) TestProcessQueueCollections_NoNextPage_NoNextQueue() {
	clusterName := "test"
	queueLevel := 0
	ackLevel := newCrossClusterTaskKey(0)
	maxLevel := newCrossClusterTaskKey(1000)
	shardMaxLevel := newCrossClusterTaskKey(500)
	processingQueueStates := []ProcessingQueueState{
		NewProcessingQueueState(
			queueLevel,
			ackLevel,
			maxLevel,
			NewDomainFilter(map[string]struct{}{"testDomain1": {}}, false),
		),
	}
	updateMaxReadLevel := func() task.Key {
		return shardMaxLevel
	}
	taskInfos := []*persistence.CrossClusterTaskInfo{
		{
			TaskID:   1,
			TaskType: 1,
			DomainID: "testDomain1",
		},
		{
			TaskID:   10,
			TaskType: 1,
			DomainID: "testDomain2",
		},
		{
			TaskID:   100,
			TaskType: 1,
			DomainID: "testDomain1",
		},
	}
	mockExecutionManager := s.mockShard.Resource.ExecutionMgr
	mockExecutionManager.On("GetCrossClusterTasks", mock.Anything, mock.Anything).Return(
		&persistence.GetCrossClusterTasksResponse{
			Tasks:         taskInfos,
			NextPageToken: nil,
		}, nil).Once()

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		processingQueueStates,
		clusterName,
		updateMaxReadLevel,
		nil,
		nil,
	)
	processorBase.options.MaxPendingTaskSize = dynamicconfig.GetIntPropertyFn(100)
	processorBase.processQueueCollections()

	queueCollection := processorBase.processingQueueCollections[0]
	s.NotNil(queueCollection.ActiveQueue())
	s.True(taskKeyEquals(shardMaxLevel, queueCollection.Queues()[0].State().ReadLevel()))
	s.Equal(2, processorBase.readyForPollTasks.Len()) // only 2 tasks belong to domain1

	shouldProcess, ok := processorBase.shouldProcess[queueLevel]
	if ok {
		s.False(shouldProcess)
	}
	select {
	case <-processorBase.processCh:
		s.Fail("processCh should be blocked")
	default:
	}
}

func (s *crossClusterQueueProcessorBaseSuite) TestHandleActionNotification_GetTasks_Success() {
	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		testKey{ID: 1},
		testKey{ID: 1},
		testKey{ID: 10},
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)
	mockTask := task.NewMockCrossClusterTask(s.controller)
	mockTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	mockTask.EXPECT().GetTaskID().Return(int64(10)).AnyTimes()
	mockTask.EXPECT().GetCrossClusterRequest().Return(&types.CrossClusterTaskRequest{}, nil).Times(1)
	taskKey := testKey{ID: 10}
	newTaskMap := map[task.Key]task.Task{taskKey: mockTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, testKey{ID: 11})
	for _, task := range newTaskMap {
		processorBase.readyForPollTasks.Put(task.GetTaskID(), task)
	}
	notification := actionNotification{
		ctx: context.Background(),
		action: &Action{
			ActionType:         ActionTypeGetTasks,
			GetTasksAttributes: &GetTasksAttributes{},
		},
		resultNotificationCh: make(chan actionResultNotification, 1),
	}
	processorBase.handleActionNotification(notification)
	select {
	case res := <-notification.resultNotificationCh:
		s.Nil(res.err)
		s.NotNil(res.result.GetTasksResult.TaskRequests)
		s.Len(res.result.GetTasksResult.TaskRequests, len(newTaskMap))
	default:
		s.Fail("fail to receive result from notification channel")
	}
}

func (s *crossClusterQueueProcessorBaseSuite) TestHandleActionNotification_UpdateTask_Success() {
	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(10),
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)
	taskID := int64(10)
	mockTask := task.NewMockCrossClusterTask(s.controller)
	mockTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	mockTask.EXPECT().GetTaskID().Return(taskID).AnyTimes()
	mockTask.EXPECT().RecordResponse(gomock.Any()).Return(nil).Times(1)
	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(true, nil).Times(1)
	taskKey := newCrossClusterTaskKey(taskID)
	newTaskMap := map[task.Key]task.Task{taskKey: mockTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(taskID+1))
	for _, task := range newTaskMap {
		processorBase.readyForPollTasks.Put(task.GetTaskID(), task)
	}
	notification := actionNotification{
		ctx: context.Background(),
		action: &Action{
			ActionType: ActionTypeUpdateTask,
			UpdateTaskAttributes: &UpdateTasksAttributes{
				TaskResponses: []*types.CrossClusterTaskResponse{
					{
						TaskID: taskID,
					},
				},
			},
		},
		resultNotificationCh: make(chan actionResultNotification, 1),
	}
	processorBase.handleActionNotification(notification)
	select {
	case res := <-notification.resultNotificationCh:
		s.NoError(res.err)
		s.NotNil(res.result.UpdateTaskResult)
	default:
		s.Fail("fail to receive result from notification channel")
	}
	s.False(processorBase.readyForPollTasks.Contains(mockTask.GetTaskID()))
}

func (s *crossClusterQueueProcessorBaseSuite) TestHandleActionNotification_UpdateTask_Fail() {
	clusterName := "test"
	processingQueueState := newProcessingQueueState(
		0,
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(1),
		newCrossClusterTaskKey(10),
		NewDomainFilter(map[string]struct{}{}, true),
	)

	processorBase := s.newTestCrossClusterQueueProcessorBase(
		[]ProcessingQueueState{processingQueueState},
		clusterName,
		nil,
		nil,
		nil,
	)
	mockTask := task.NewMockCrossClusterTask(s.controller)
	mockTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	mockTask.EXPECT().GetTaskID().Return(int64(10)).AnyTimes()
	mockTask.EXPECT().GetWorkflowID().Return("workflowID").AnyTimes()
	mockTask.EXPECT().GetRunID().Return("runID").AnyTimes()
	mockTask.EXPECT().RecordResponse(gomock.Any()).Return(errors.New("test")).Times(1)
	taskKey := newCrossClusterTaskKey(mockTask.GetTaskID())
	newTaskMap := map[task.Key]task.Task{taskKey: mockTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(11))
	for _, task := range newTaskMap {
		processorBase.readyForPollTasks.Put(task.GetTaskID(), task)
	}
	notification := actionNotification{
		ctx: context.Background(),
		action: &Action{
			ActionType: ActionTypeUpdateTask,
			UpdateTaskAttributes: &UpdateTasksAttributes{
				TaskResponses: []*types.CrossClusterTaskResponse{
					{
						TaskID: mockTask.GetTaskID(),
					},
				},
			},
		},
		resultNotificationCh: make(chan actionResultNotification, 1),
	}
	processorBase.handleActionNotification(notification)
	select {
	case res := <-notification.resultNotificationCh:
		// we won't return error for update task,
		// tasks will still available for polling if update fails
		s.NoError(res.err)
	default:
		s.Fail("fail to receive result from notification channel")
	}
	s.True(processorBase.readyForPollTasks.Contains(mockTask.GetTaskID()))
}

func (s *crossClusterQueueProcessorBaseSuite) newTestCrossClusterQueueProcessorBase(
	processingQueueStates []ProcessingQueueState,
	clusterName string,
	updateMaxReadLevel updateMaxReadLevelFn,
	updateProcessingQueueStates updateProcessingQueueStatesFn,
	transferQueueShutdown queueShutdownFn,
) *crossClusterQueueProcessorBase {

	processorBase := newCrossClusterQueueProcessorBaseHelper(
		s.mockShard,
		clusterName,
		nil,
		processingQueueStates,
		s.mockTaskProcessor,
		newCrossClusterQueueProcessorOptions(s.mockShard.GetConfig()),
		updateMaxReadLevel,
		updateProcessingQueueStates,
		transferQueueShutdown,
		nil,
		s.logger,
		s.metricsClient,
	)
	for _, queueCollections := range processorBase.processingQueueCollections {
		processorBase.shouldProcess[queueCollections.Level()] = true
	}
	return processorBase
}
