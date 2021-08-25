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
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common/types"

	ctask "github.com/uber/cadence/common/task"

	"github.com/uber/cadence/common/dynamicconfig"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
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

func (s *crossClusterQueueProcessorBaseSuite) TestGetTask_TaskExist_Success() {
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
	mockTask := task.NewMockTask(s.controller)
	mockTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	taskKey := testKey{ID: 10}
	newTaskMap := map[task.Key]task.Task{taskKey: mockTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, testKey{ID: 11})
	task, err := processorBase.getTask(taskKey)
	s.NoError(err)
	s.Equal(mockTask, task)
}

func (s *crossClusterQueueProcessorBaseSuite) TestGetTask_NoTask_Fail() {
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
	taskKey := testKey{ID: 10}

	_, err := processorBase.getTask(taskKey)
	s.Error(err)
}

func (s *crossClusterQueueProcessorBaseSuite) TestGetTasks_NoTask_Success() {
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

	tasks := processorBase.getTasks()
	s.Equal(0, len(tasks))
}

func (s *crossClusterQueueProcessorBaseSuite) TestGetTasks_Success() {
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
	taskKey := testKey{ID: 10}
	newTaskMap := map[task.Key]task.Task{taskKey: mockTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, testKey{ID: 11})
	tasks := processorBase.getTasks()
	s.Equal(1, len(tasks))
	s.Equal(mockTask, tasks[0])
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

func (s *crossClusterQueueProcessorBaseSuite) TestPollTasks_OneTaskReady_ReturnOneTask() {
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

	readyTask := task.NewMockCrossClusterTask(s.controller)
	readyTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	readyTask.EXPECT().IsReadyForPoll().Return(true)
	readyTask.EXPECT().GetVisibilityTimestamp().Return(time.Time{}).AnyTimes()
	readyTask.EXPECT().GetWorkflowID().Return(uuid.New()).AnyTimes()
	readyTask.EXPECT().GetRunID().Return(uuid.New()).AnyTimes()
	readyTask.EXPECT().GetTaskType().Return(1).AnyTimes()
	readyTask.EXPECT().State().Return(ctask.State(0)).AnyTimes()
	readyTask.EXPECT().GetTaskID().Return(int64(0)).AnyTimes()
	readyTask.EXPECT().GetCrossClusterRequest().Return(nil).AnyTimes()
	notReadyTask := task.NewMockCrossClusterTask(s.controller)
	notReadyTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	notReadyTask.EXPECT().IsReadyForPoll().Return(false)
	newTaskMap := map[task.Key]task.Task{testKey{ID: 2}: readyTask, testKey{3}: notReadyTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, testKey{ID: 10})
	tasks := processorBase.pollTasks()
	s.Equal(1, len(tasks))
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

	crossClusterTask := task.NewMockCrossClusterTask(s.controller)
	crossClusterTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	crossClusterTask.EXPECT().Update(gomock.Any()).Return(nil).Times(1)
	crossClusterTask.EXPECT().GetTaskType().Return(1).AnyTimes()
	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(true, nil).Times(1)
	newTaskMap := map[task.Key]task.Task{newCrossClusterTaskKey(2): crossClusterTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(10))
	err := processorBase.updateTask(&types.CrossClusterTaskResponse{
		TaskID: int64(2),
	})
	s.NoError(err)
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

	crossClusterTask := task.NewMockCrossClusterTask(s.controller)
	crossClusterTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	crossClusterTask.EXPECT().GetTaskID().Return(int64(2)).Times(1)
	crossClusterTask.EXPECT().GetWorkflowID().Return("workflowID").Times(1)
	crossClusterTask.EXPECT().GetRunID().Return("runID").Times(1)
	crossClusterTask.EXPECT().Update(gomock.Any()).Return(errors.New("test")).Times(1)
	crossClusterTask.EXPECT().GetTaskType().Return(1).AnyTimes()
	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(true, nil).Times(0)
	newTaskMap := map[task.Key]task.Task{newCrossClusterTaskKey(2): crossClusterTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(10))
	err := processorBase.updateTask(&types.CrossClusterTaskResponse{
		TaskID: int64(2),
	})
	s.Error(err)
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

	crossClusterTask := task.NewMockCrossClusterTask(s.controller)
	crossClusterTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	crossClusterTask.EXPECT().Update(gomock.Any()).Return(nil).Times(1)
	crossClusterTask.EXPECT().Priority().Return(0).AnyTimes()
	crossClusterTask.EXPECT().GetAttempt().Return(0).Times(1)
	crossClusterTask.EXPECT().GetTaskType().Return(1).AnyTimes()
	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(false, errors.New("test")).Times(1)
	newTaskMap := map[task.Key]task.Task{newCrossClusterTaskKey(2): crossClusterTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(10))
	err := processorBase.updateTask(&types.CrossClusterTaskResponse{
		TaskID: int64(2),
	})
	s.NoError(err)
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

	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(true, nil).AnyTimes()
	processorBase.options.MaxPendingTaskSize = dynamicconfig.GetIntPropertyFn(100)
	processorBase.processQueueCollections()

	queueCollection := processorBase.processingQueueCollections[0]
	s.NotNil(queueCollection.ActiveQueue())
	s.True(taskKeyEquals(maxLevel, queueCollection.Queues()[0].State().ReadLevel()))

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

	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(true, nil).AnyTimes()

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
	mockTask.EXPECT().IsReadyForPoll().Return(true).AnyTimes()
	mockTask.EXPECT().GetVisibilityTimestamp().Return(time.Time{}).AnyTimes()
	mockTask.EXPECT().GetWorkflowID().Return(uuid.New()).AnyTimes()
	mockTask.EXPECT().GetRunID().Return(uuid.New()).AnyTimes()
	mockTask.EXPECT().GetTaskType().Return(1).AnyTimes()
	mockTask.EXPECT().State().Return(ctask.State(0)).AnyTimes()
	mockTask.EXPECT().GetTaskID().Return(int64(0)).AnyTimes()
	mockTask.EXPECT().GetCrossClusterRequest().Return(nil).AnyTimes()
	taskKey := testKey{ID: 10}
	newTaskMap := map[task.Key]task.Task{taskKey: mockTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, testKey{ID: 11})
	notification := actionNotification{
		action: &Action{
			ActionType:         ActionTypeGetTasks,
			GetTasksAttributes: &GetTasksAttributes{},
		},
		resultNotificationCh: make(chan actionResultNotification, 1),
	}
	processorBase.handleActionNotification(notification)
	select {
	case res := <-notification.resultNotificationCh:
		s.NotNil(res.result.GetTasksResult.TaskRequests)
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
	mockTask := task.NewMockCrossClusterTask(s.controller)
	mockTask.EXPECT().GetDomainID().Return(uuid.New()).AnyTimes()
	mockTask.EXPECT().GetTaskID().Return(int64(10)).AnyTimes()
	mockTask.EXPECT().Update(gomock.Any()).Return(nil).AnyTimes()
	mockTask.EXPECT().GetTaskType().Return(1).AnyTimes()
	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(true, nil).AnyTimes()
	taskKey := newCrossClusterTaskKey(10)
	newTaskMap := map[task.Key]task.Task{taskKey: mockTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(11))
	notification := actionNotification{
		action: &Action{
			ActionType: ActionTypeUpdateTask,
			UpdateTaskAttributes: &UpdateTasksAttributes{
				TaskResponses: []*types.CrossClusterTaskResponse{
					{
						TaskID:   int64(10),
						TaskType: types.CrossClusterTaskType(1).Ptr(),
					},
				},
			},
		},
		resultNotificationCh: make(chan actionResultNotification, 1),
	}
	processorBase.handleActionNotification(notification)
	select {
	case res := <-notification.resultNotificationCh:
		s.NotNil(res.result.UpdateTaskResult)
	default:
		s.Fail("fail to receive result from notification channel")
	}
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
	mockTask.EXPECT().Update(gomock.Any()).Return(errors.New("test")).AnyTimes()
	mockTask.EXPECT().GetWorkflowID().Return("workflowID").AnyTimes()
	mockTask.EXPECT().GetRunID().Return("runID").AnyTimes()
	mockTask.EXPECT().GetTaskType().Return(1).AnyTimes()
	s.mockTaskProcessor.EXPECT().TrySubmit(gomock.Any()).Return(true, nil).AnyTimes()
	taskKey := newCrossClusterTaskKey(10)
	newTaskMap := map[task.Key]task.Task{taskKey: mockTask}
	processorBase.processingQueueCollections[0].AddTasks(newTaskMap, newCrossClusterTaskKey(11))
	notification := actionNotification{
		action: &Action{
			ActionType: ActionTypeUpdateTask,
			UpdateTaskAttributes: &UpdateTasksAttributes{
				TaskResponses: []*types.CrossClusterTaskResponse{
					{
						TaskID: int64(10),
					},
				},
			},
		},
		resultNotificationCh: make(chan actionResultNotification, 1),
	}
	processorBase.handleActionNotification(notification)
	select {
	case res := <-notification.resultNotificationCh:
		s.Error(res.err)
	default:
		s.Fail("fail to receive result from notification channel")
	}
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
