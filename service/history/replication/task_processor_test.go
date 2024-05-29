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

package replication

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/goleak"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/reconciliation"
	"github.com/uber/cadence/common/reconciliation/entity"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/shard"
)

type (
	taskProcessorSuite struct {
		suite.Suite
		*require.Assertions
		controller *gomock.Controller

		mockShard          *shard.TestContext
		mockEngine         *engine.MockEngine
		config             *config.Config
		mockDomainCache    *cache.MockDomainCache
		mockClientBean     *client.MockBean
		mockFrontendClient *frontend.MockClient
		adminClient        *admin.MockClient
		executionManager   *mocks.ExecutionManager
		requestChan        chan *request
		taskFetcher        *fakeTaskFetcher
		taskExecutor       *MockTaskExecutor
		taskProcessor      *taskProcessorImpl
	}
)

func TestTaskProcessorSuite(t *testing.T) {
	s := new(taskProcessorSuite)
	suite.Run(t, s)
}

func (s *taskProcessorSuite) SetupSuite() {

}

func (s *taskProcessorSuite) TearDownSuite() {

}

func (s *taskProcessorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockShard = shard.NewTestContext(
		s.T(),
		s.controller,
		&persistence.ShardInfo{
			ShardID:          0,
			RangeID:          1,
			TransferAckLevel: 0,
		},
		s.config,
	)

	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockClientBean = s.mockShard.Resource.ClientBean
	s.mockFrontendClient = s.mockShard.Resource.RemoteFrontendClient
	s.adminClient = s.mockShard.Resource.RemoteAdminClient
	s.executionManager = s.mockShard.Resource.ExecutionMgr

	s.mockEngine = engine.NewMockEngine(s.controller)
	s.config = config.NewForTest()
	s.config.ReplicationTaskProcessorNoTaskRetryWait = dynamicconfig.GetDurationPropertyFnFilteredByShardID(1 * time.Millisecond)
	metricsClient := metrics.NewClient(tally.NoopScope, metrics.History)
	s.requestChan = make(chan *request, 10)

	s.taskFetcher = &fakeTaskFetcher{
		sourceCluster: "standby",
		requestChan:   s.requestChan,
		rateLimiter:   quotas.NewDynamicRateLimiter(func() float64 { return 100 }),
	}

	s.taskExecutor = NewMockTaskExecutor(s.controller)

	s.taskProcessor = NewTaskProcessor(
		s.mockShard,
		s.mockEngine,
		s.config,
		metricsClient,
		s.taskFetcher,
		s.taskExecutor,
	).(*taskProcessorImpl)
}

func (s *taskProcessorSuite) TearDownTest() {
	s.mockShard.Finish(s.T())
	goleak.VerifyNone(s.T())
}

func (s *taskProcessorSuite) TestStartStop() {
	s.taskProcessor.Start()
	s.taskProcessor.Stop()
}

func (s *taskProcessorSuite) TestProcessResponse_NoTask() {
	response := &types.ReplicationMessages{
		LastRetrievedMessageID: 100,
	}

	s.taskProcessor.processResponse(response)
	s.Equal(int64(100), s.taskProcessor.lastProcessedMessageID)
	s.Equal(int64(100), s.taskProcessor.lastRetrievedMessageID)
}

func (s *taskProcessorSuite) TestProcessorLoop_RequestChanPopulated() {
	// start the process loop so it poppulates requestChan
	s.taskProcessor.wg.Add(1)
	go s.taskProcessor.processorLoop()

	// wait a bit and terminate the loop
	time.Sleep(50 * time.Millisecond)
	close(s.taskProcessor.done)

	// check the request
	requestMessage := <-s.requestChan

	s.Equal(int32(0), requestMessage.token.GetShardID())
	s.Equal(int64(-1), requestMessage.token.GetLastProcessedMessageID())
	s.Equal(int64(-1), requestMessage.token.GetLastRetrievedMessageID())
	s.NotNil(requestMessage.respChan)
}

func (s *taskProcessorSuite) TestProcessorLoop_RespChanClosed() {
	// start the process loop
	s.taskProcessor.wg.Add(1)
	go s.taskProcessor.processorLoop()
	defer close(s.taskProcessor.done)

	// act like taskFetcher here and populate respChan of the request
	requestMessage := <-s.requestChan
	close(requestMessage.respChan)

	// loop should have continued by now. validate by checking the new request
	select {
	case <-s.requestChan:
	// expected
	case <-time.After(50 * time.Millisecond):
		s.Fail("new request not sent to requestChan")
	}
}

func (s *taskProcessorSuite) TestProcessorLoop_TaskExecuteSuccess() {
	// taskExecutor will fail to execute the task
	// returning a non-retriable task to keep mocking simpler
	s.taskExecutor.EXPECT().execute(gomock.Any(), false).Return(0, nil).Times(1)

	// domain name will be fetched
	s.mockDomainCache.EXPECT().GetDomainName(testDomainID).Return(testDomainName, nil).AnyTimes()

	// start the process loop
	s.taskProcessor.wg.Add(1)
	go s.taskProcessor.processorLoop()

	// act like taskFetcher here and populate respChan of the request
	requestMessage := <-s.requestChan
	requestMessage.respChan <- &types.ReplicationMessages{
		LastRetrievedMessageID: 100,
		ReplicationTasks: []*types.ReplicationTask{
			{
				TaskType: types.ReplicationTaskTypeSyncActivity.Ptr(),
				SyncActivityTaskAttributes: &types.SyncActivityTaskAttributes{
					DomainID:    testDomainID,
					WorkflowID:  testWorkflowID,
					RunID:       testRunID,
					ScheduledID: testScheduleID,
				},
				SourceTaskID: testTaskID,
			},
		},
	}

	// wait a bit and terminate the loop
	time.Sleep(50 * time.Millisecond)
	close(s.taskProcessor.done)
}

func (s *taskProcessorSuite) TestProcessorLoop_TaskExecuteFailed_PutDLQSuccess() {
	// taskExecutor will fail to execute the task
	// returning a non-retriable task to keep mocking simpler
	s.taskExecutor.EXPECT().execute(gomock.Any(), false).Return(0, &types.BadRequestError{}).Times(1)

	// domain name will be fetched
	s.mockDomainCache.EXPECT().GetDomainName(testDomainID).Return(testDomainName, nil).AnyTimes()

	// task will be put into dlq
	dlqReq := &persistence.PutReplicationTaskToDLQRequest{
		SourceClusterName: "standby", // TODO move to a constant
		TaskInfo: &persistence.ReplicationTaskInfo{
			DomainID:    testDomainID,
			WorkflowID:  testWorkflowID,
			RunID:       testRunID,
			TaskID:      testTaskID,
			TaskType:    persistence.ReplicationTaskTypeSyncActivity,
			ScheduledID: testScheduleID,
		},
		DomainName: testDomainName,
	}
	s.mockShard.Resource.ExecutionMgr.On("PutReplicationTaskToDLQ", mock.Anything, dlqReq).Return(nil).Times(1)

	// start the process loop
	s.taskProcessor.wg.Add(1)
	go s.taskProcessor.processorLoop()

	// act like taskFetcher here and populate respChan of the request
	requestMessage := <-s.requestChan
	requestMessage.respChan <- &types.ReplicationMessages{
		LastRetrievedMessageID: 100,
		ReplicationTasks: []*types.ReplicationTask{
			{
				TaskType: types.ReplicationTaskTypeSyncActivity.Ptr(),
				SyncActivityTaskAttributes: &types.SyncActivityTaskAttributes{
					DomainID:    testDomainID,
					WorkflowID:  testWorkflowID,
					RunID:       testRunID,
					ScheduledID: testScheduleID,
				},
				SourceTaskID: testTaskID,
			},
		},
	}

	// wait a bit and terminate the loop
	time.Sleep(50 * time.Millisecond)
	close(s.taskProcessor.done)
}

func (s *taskProcessorSuite) TestProcessorLoop_TaskExecuteFailed_PutDLQFailed() {
	// taskExecutor will fail to execute the task
	// returning a non-retriable task to keep mocking simpler
	s.taskExecutor.EXPECT().execute(gomock.Any(), false).Return(0, &types.BadRequestError{}).Times(1)

	// domain name will be fetched
	s.mockDomainCache.EXPECT().GetDomainName(testDomainID).Return(testDomainName, nil).AnyTimes()

	// task will be put into dlq and will fail. It will be attempted 3 times. (first call + 2 retries based on policy overriden below)
	dqlRetryPolicy := backoff.NewExponentialRetryPolicy(time.Millisecond)
	dqlRetryPolicy.SetMaximumAttempts(2)
	s.taskProcessor.dlqRetryPolicy = dqlRetryPolicy
	dlqReq := &persistence.PutReplicationTaskToDLQRequest{
		SourceClusterName: "standby", // TODO move to a constant
		TaskInfo: &persistence.ReplicationTaskInfo{
			DomainID:    testDomainID,
			WorkflowID:  testWorkflowID,
			RunID:       testRunID,
			TaskID:      testTaskID,
			TaskType:    persistence.ReplicationTaskTypeSyncActivity,
			ScheduledID: testScheduleID,
		},
		DomainName: testDomainName,
	}
	s.mockShard.Resource.ExecutionMgr.
		On("PutReplicationTaskToDLQ", mock.Anything, dlqReq).
		Return(errors.New("failed to put to dlq")).
		Times(3)

	// start the process loop
	s.taskProcessor.wg.Add(1)
	go s.taskProcessor.processorLoop()

	// act like taskFetcher here and populate respChan of the request
	requestMessage := <-s.requestChan
	requestMessage.respChan <- &types.ReplicationMessages{
		LastRetrievedMessageID: 100,
		ReplicationTasks: []*types.ReplicationTask{
			{
				TaskType: types.ReplicationTaskTypeSyncActivity.Ptr(),
				SyncActivityTaskAttributes: &types.SyncActivityTaskAttributes{
					DomainID:    testDomainID,
					WorkflowID:  testWorkflowID,
					RunID:       testRunID,
					ScheduledID: testScheduleID,
				},
				SourceTaskID: testTaskID,
			},
		},
	}

	// wait a bit and terminate the loop
	time.Sleep(50 * time.Millisecond)
	close(s.taskProcessor.done)
}

func (s *taskProcessorSuite) TestHandleSyncShardStatus() {
	now := time.Now()
	s.mockEngine.EXPECT().SyncShardStatus(gomock.Any(), &types.SyncShardStatusRequest{
		SourceCluster: "standby",
		ShardID:       0,
		Timestamp:     common.Int64Ptr(now.UnixNano()),
	}).Return(nil).Times(1)

	err := s.taskProcessor.handleSyncShardStatus(&types.SyncShardStatus{
		Timestamp: common.Int64Ptr(now.UnixNano()),
	})
	s.NoError(err)
}

func (s *taskProcessorSuite) TestPutReplicationTaskToDLQ_SyncActivityReplicationTask() {
	request := &persistence.PutReplicationTaskToDLQRequest{
		SourceClusterName: "standby",
		TaskInfo: &persistence.ReplicationTaskInfo{
			DomainID:   uuid.New(),
			WorkflowID: uuid.New(),
			RunID:      uuid.New(),
			TaskType:   persistence.ReplicationTaskTypeSyncActivity,
		},
		DomainName: uuid.New(),
	}
	s.executionManager.On("PutReplicationTaskToDLQ", mock.Anything, request).Return(nil)
	err := s.taskProcessor.putReplicationTaskToDLQ(request)
	s.NoError(err)
}

func (s *taskProcessorSuite) TestPutReplicationTaskToDLQ_HistoryV2ReplicationTask() {
	request := &persistence.PutReplicationTaskToDLQRequest{
		SourceClusterName: "standby",
		TaskInfo: &persistence.ReplicationTaskInfo{
			DomainID:     uuid.New(),
			WorkflowID:   uuid.New(),
			RunID:        uuid.New(),
			TaskType:     persistence.ReplicationTaskTypeHistory,
			FirstEventID: 1,
			NextEventID:  2,
			Version:      1,
		},
		DomainName: uuid.New(),
	}
	s.executionManager.On("PutReplicationTaskToDLQ", mock.Anything, request).Return(nil)
	err := s.taskProcessor.putReplicationTaskToDLQ(request)
	s.NoError(err)
}

func (s *taskProcessorSuite) TestGenerateDLQRequest_ReplicationTaskTypeHistoryV2() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	events := []*types.HistoryEvent{
		{
			ID:      1,
			Version: 1,
		},
	}
	serializer := s.mockShard.GetPayloadSerializer()
	data, err := serializer.SerializeBatchEvents(events, common.EncodingTypeThriftRW)
	s.NoError(err)
	task := &types.ReplicationTask{
		TaskType: types.ReplicationTaskTypeHistoryV2.Ptr(),
		HistoryTaskV2Attributes: &types.HistoryTaskV2Attributes{
			DomainID:   domainID,
			WorkflowID: workflowID,
			RunID:      runID,
			Events: &types.DataBlob{
				EncodingType: types.EncodingTypeThriftRW.Ptr(),
				Data:         data.Data,
			},
		},
	}
	s.mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("test_domain_name", nil).AnyTimes()
	request, err := s.taskProcessor.generateDLQRequest(task)
	s.NoError(err)
	s.Equal("standby", request.SourceClusterName)
	s.Equal(int64(1), request.TaskInfo.FirstEventID)
	s.Equal(int64(2), request.TaskInfo.NextEventID)
	s.Equal(int64(1), request.TaskInfo.GetVersion())
	s.Equal(domainID, request.TaskInfo.GetDomainID())
	s.Equal(workflowID, request.TaskInfo.GetWorkflowID())
	s.Equal(runID, request.TaskInfo.GetRunID())
	s.Equal(persistence.ReplicationTaskTypeHistory, request.TaskInfo.GetTaskType())
}

func (s *taskProcessorSuite) TestGenerateDLQRequest_ReplicationTaskTypeSyncActivity() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	domainName := uuid.New()
	task := &types.ReplicationTask{
		TaskType: types.ReplicationTaskTypeSyncActivity.Ptr(),
		SyncActivityTaskAttributes: &types.SyncActivityTaskAttributes{
			DomainID:    domainID,
			WorkflowID:  workflowID,
			RunID:       runID,
			ScheduledID: 1,
		},
	}
	s.mockDomainCache.EXPECT().GetDomainName(domainID).Return(domainName, nil).AnyTimes()
	request, err := s.taskProcessor.generateDLQRequest(task)
	s.NoError(err)
	s.Equal("standby", request.SourceClusterName)
	s.Equal(int64(1), request.TaskInfo.ScheduledID)
	s.Equal(domainID, request.TaskInfo.GetDomainID())
	s.Equal(workflowID, request.TaskInfo.GetWorkflowID())
	s.Equal(runID, request.TaskInfo.GetRunID())
	s.Equal(persistence.ReplicationTaskTypeSyncActivity, request.TaskInfo.GetTaskType())
}

func (s *taskProcessorSuite) TestTriggerDataInconsistencyScan_Success() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	task := &types.ReplicationTask{
		TaskType: types.ReplicationTaskTypeSyncActivity.Ptr(),
		SyncActivityTaskAttributes: &types.SyncActivityTaskAttributes{
			DomainID:    domainID,
			WorkflowID:  workflowID,
			RunID:       runID,
			ScheduledID: 1,
			Version:     100,
		},
	}
	fixExecution := entity.Execution{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
		ShardID:    s.mockShard.GetShardID(),
	}
	jsArray, err := json.Marshal(fixExecution)
	s.NoError(err)
	s.mockFrontendClient.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, request *types.SignalWithStartWorkflowExecutionRequest, option ...yarpc.CallOption) {
			s.Equal(common.SystemLocalDomainName, request.GetDomain())
			s.Equal(reconciliation.CheckDataCorruptionWorkflowID, request.GetWorkflowID())
			s.Equal(reconciliation.CheckDataCorruptionWorkflowType, request.GetWorkflowType().GetName())
			s.Equal(reconciliation.CheckDataCorruptionWorkflowTaskList, request.GetTaskList().GetName())
			s.Equal(types.WorkflowIDReusePolicyAllowDuplicate.String(), request.GetWorkflowIDReusePolicy().String())
			s.Equal(reconciliation.CheckDataCorruptionWorkflowSignalName, request.GetSignalName())
			s.Equal(jsArray, request.GetSignalInput())
		}).Return(&types.StartWorkflowExecutionResponse{}, nil)

	err = s.taskProcessor.triggerDataInconsistencyScan(task)
	s.NoError(err)
}

type fakeTaskFetcher struct {
	sourceCluster string
	requestChan   chan<- *request
	rateLimiter   *quotas.DynamicRateLimiter
}

func (f fakeTaskFetcher) Start() {}
func (f fakeTaskFetcher) Stop()  {}
func (f fakeTaskFetcher) GetSourceCluster() string {
	return f.sourceCluster
}
func (f fakeTaskFetcher) GetRequestChan() chan<- *request {
	return f.requestChan
}
func (f fakeTaskFetcher) GetRateLimiter() *quotas.DynamicRateLimiter {
	return f.rateLimiter
}
