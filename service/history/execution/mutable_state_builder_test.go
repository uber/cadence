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

package execution

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/checksum"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/shard"
	shardCtx "github.com/uber/cadence/service/history/shard"
)

type (
	mutableStateSuite struct {
		suite.Suite
		*require.Assertions

		controller      *gomock.Controller
		mockShard       *shard.TestContext
		mockEventsCache *events.MockCache

		msBuilder *mutableStateBuilder
		logger    log.Logger
		testScope tally.TestScope
	}
)

func TestMutableStateSuite(t *testing.T) {
	s := new(mutableStateSuite)
	suite.Run(t, s)
}

func (s *mutableStateSuite) SetupSuite() {

}

func (s *mutableStateSuite) TearDownSuite() {

}

func (s *mutableStateSuite) SetupTest() {
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
		config.NewForTest(),
	)
	// set the checksum probabilities to 100% for exercising during test
	s.mockShard.GetConfig().MutableStateChecksumGenProbability = func(domain string) int { return 100 }
	s.mockShard.GetConfig().MutableStateChecksumVerifyProbability = func(domain string) int { return 100 }
	s.mockShard.GetConfig().EnableRetryForChecksumFailure = func(domain string) bool { return true }

	s.mockEventsCache = s.mockShard.GetEventsCache().(*events.MockCache)

	s.testScope = s.mockShard.Resource.MetricsScope
	s.logger = s.mockShard.GetLogger()

	s.mockShard.Resource.DomainCache.EXPECT().GetDomainID(constants.TestDomainName).Return(constants.TestDomainID, nil).AnyTimes()

	s.msBuilder = newMutableStateBuilder(s.mockShard, s.logger, constants.TestLocalDomainEntry)
}

func (s *mutableStateSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
}

func (s *mutableStateSuite) TestErrorReturnedWhenSchedulingTooManyPendingActivities() {
	for i := 0; i < s.msBuilder.config.PendingActivitiesCountLimitError(); i++ {
		s.msBuilder.pendingActivityInfoIDs[int64(i)] = &persistence.ActivityInfo{}
	}

	_, _, _, _, _, err := s.msBuilder.AddActivityTaskScheduledEvent(nil, 1, &types.ScheduleActivityTaskDecisionAttributes{}, false)
	assert.Equal(s.T(), "Too many pending activities", err.Error())
}

func (s *mutableStateSuite) TestTransientDecisionCompletionFirstBatchReplicated_ReplicateDecisionCompleted() {
	version := int64(12)
	runID := uuid.New()
	s.msBuilder = NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		version,
		runID,
		constants.TestGlobalDomainEntry,
	).(*mutableStateBuilder)

	newDecisionScheduleEvent, newDecisionStartedEvent := s.prepareTransientDecisionCompletionFirstBatchReplicated(version, runID)

	newDecisionCompletedEvent := &types.HistoryEvent{
		Version:   version,
		ID:        newDecisionStartedEvent.ID + 1,
		Timestamp: common.Int64Ptr(time.Now().UnixNano()),
		EventType: types.EventTypeDecisionTaskCompleted.Ptr(),
		DecisionTaskCompletedEventAttributes: &types.DecisionTaskCompletedEventAttributes{
			ScheduledEventID: newDecisionScheduleEvent.ID,
			StartedEventID:   newDecisionStartedEvent.ID,
			Identity:         "some random identity",
		},
	}
	err := s.msBuilder.ReplicateDecisionTaskCompletedEvent(newDecisionCompletedEvent)
	s.NoError(err)
	s.Equal(0, len(s.msBuilder.GetHistoryBuilder().transientHistory))
	s.Equal(0, len(s.msBuilder.GetHistoryBuilder().history))
}

func (s *mutableStateSuite) TestTransientDecisionCompletionFirstBatchReplicated_FailoverDecisionTimeout() {
	version := int64(12)
	runID := uuid.New()
	s.msBuilder = NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		version,
		runID,
		constants.TestGlobalDomainEntry,
	).(*mutableStateBuilder)

	newDecisionScheduleEvent, newDecisionStartedEvent := s.prepareTransientDecisionCompletionFirstBatchReplicated(version, runID)

	s.NotNil(s.msBuilder.AddDecisionTaskTimedOutEvent(newDecisionScheduleEvent.ID, newDecisionStartedEvent.ID))
	s.Equal(0, len(s.msBuilder.GetHistoryBuilder().transientHistory))
	s.Equal(1, len(s.msBuilder.GetHistoryBuilder().history))
}

func (s *mutableStateSuite) TestTransientDecisionCompletionFirstBatchReplicated_FailoverDecisionFailed() {
	version := int64(12)
	runID := uuid.New()
	s.msBuilder = NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		version,
		runID,
		constants.TestGlobalDomainEntry,
	).(*mutableStateBuilder)

	newDecisionScheduleEvent, newDecisionStartedEvent := s.prepareTransientDecisionCompletionFirstBatchReplicated(version, runID)

	s.NotNil(s.msBuilder.AddDecisionTaskFailedEvent(
		newDecisionScheduleEvent.ID,
		newDecisionStartedEvent.ID,
		types.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure,
		[]byte("some random decision failure details"),
		"some random decision failure identity",
		"", "", "", "", 0,
	))
	s.Equal(0, len(s.msBuilder.GetHistoryBuilder().transientHistory))
	s.Equal(1, len(s.msBuilder.GetHistoryBuilder().history))
}

func (s *mutableStateSuite) TestShouldBufferEvent() {
	// workflow status events will be assign event ID immediately
	workflowEvents := map[types.EventType]bool{
		types.EventTypeWorkflowExecutionStarted:        true,
		types.EventTypeWorkflowExecutionCompleted:      true,
		types.EventTypeWorkflowExecutionFailed:         true,
		types.EventTypeWorkflowExecutionTimedOut:       true,
		types.EventTypeWorkflowExecutionTerminated:     true,
		types.EventTypeWorkflowExecutionContinuedAsNew: true,
		types.EventTypeWorkflowExecutionCanceled:       true,
	}

	// decision events will be assign event ID immediately
	decisionTaskEvents := map[types.EventType]bool{
		types.EventTypeDecisionTaskScheduled: true,
		types.EventTypeDecisionTaskStarted:   true,
		types.EventTypeDecisionTaskCompleted: true,
		types.EventTypeDecisionTaskFailed:    true,
		types.EventTypeDecisionTaskTimedOut:  true,
	}

	// events corresponding to decisions from client will be assign event ID immediately
	decisionEvents := map[types.EventType]bool{
		types.EventTypeWorkflowExecutionCompleted:                      true,
		types.EventTypeWorkflowExecutionFailed:                         true,
		types.EventTypeWorkflowExecutionCanceled:                       true,
		types.EventTypeWorkflowExecutionContinuedAsNew:                 true,
		types.EventTypeActivityTaskScheduled:                           true,
		types.EventTypeActivityTaskCancelRequested:                     true,
		types.EventTypeTimerStarted:                                    true,
		types.EventTypeTimerCanceled:                                   true,
		types.EventTypeCancelTimerFailed:                               true,
		types.EventTypeRequestCancelExternalWorkflowExecutionInitiated: true,
		types.EventTypeMarkerRecorded:                                  true,
		types.EventTypeStartChildWorkflowExecutionInitiated:            true,
		types.EventTypeSignalExternalWorkflowExecutionInitiated:        true,
		types.EventTypeUpsertWorkflowSearchAttributes:                  true,
	}

	// other events will not be assign event ID immediately
	otherEvents := map[types.EventType]bool{}
OtherEventsLoop:
	for _, eventType := range types.EventTypeValues() {
		if _, ok := workflowEvents[eventType]; ok {
			continue OtherEventsLoop
		}
		if _, ok := decisionTaskEvents[eventType]; ok {
			continue OtherEventsLoop
		}
		if _, ok := decisionEvents[eventType]; ok {
			continue OtherEventsLoop
		}
		otherEvents[eventType] = true
	}

	// test workflowEvents, decisionTaskEvents, decisionEvents will return true
	for eventType := range workflowEvents {
		s.False(s.msBuilder.shouldBufferEvent(eventType))
	}
	for eventType := range decisionTaskEvents {
		s.False(s.msBuilder.shouldBufferEvent(eventType))
	}
	for eventType := range decisionEvents {
		s.False(s.msBuilder.shouldBufferEvent(eventType))
	}
	// other events will return false
	for eventType := range otherEvents {
		s.True(s.msBuilder.shouldBufferEvent(eventType))
	}

	// +1 is because DecisionTypeCancelTimer will be mapped
	// to either types.EventTypeTimerCanceled, or types.EventTypeCancelTimerFailed.
	s.Equal(len(types.DecisionTypeValues())+1, len(decisionEvents),
		"This assertaion will be broken a new decision is added and no corresponding logic added to shouldBufferEvent()")
}

func (s *mutableStateSuite) TestReorderEvents() {
	domainID := constants.TestDomainID
	we := types.WorkflowExecution{
		WorkflowID: "wId",
		RunID:      constants.TestRunID,
	}
	tl := "testTaskList"
	activityID := "activity_id"
	activityResult := []byte("activity_result")

	info := &persistence.WorkflowExecutionInfo{
		DomainID:                    domainID,
		WorkflowID:                  we.GetWorkflowID(),
		RunID:                       we.GetRunID(),
		TaskList:                    tl,
		WorkflowTypeName:            "wType",
		WorkflowTimeout:             200,
		DecisionStartToCloseTimeout: 100,
		State:                       persistence.WorkflowStateRunning,
		CloseStatus:                 persistence.WorkflowCloseStatusNone,
		NextEventID:                 int64(8),
		LastProcessedEvent:          int64(3),
		LastUpdatedTimestamp:        time.Now(),
		DecisionVersion:             common.EmptyVersion,
		DecisionScheduleID:          common.EmptyEventID,
		DecisionStartedID:           common.EmptyEventID,
		DecisionTimeout:             100,
	}

	activityInfos := map[int64]*persistence.ActivityInfo{
		5: {
			Version:                int64(1),
			ScheduleID:             int64(5),
			ScheduledTime:          time.Now(),
			StartedID:              common.EmptyEventID,
			StartedTime:            time.Now(),
			ActivityID:             activityID,
			ScheduleToStartTimeout: 100,
			ScheduleToCloseTimeout: 200,
			StartToCloseTimeout:    300,
			HeartbeatTimeout:       50,
		},
	}

	bufferedEvents := []*types.HistoryEvent{
		{
			ID:        common.BufferedEventID,
			EventType: types.EventTypeActivityTaskCompleted.Ptr(),
			Version:   1,
			ActivityTaskCompletedEventAttributes: &types.ActivityTaskCompletedEventAttributes{
				Result:           []byte(activityResult),
				ScheduledEventID: 5,
				StartedEventID:   common.BufferedEventID,
			},
		},
		{
			ID:        common.BufferedEventID,
			EventType: types.EventTypeActivityTaskStarted.Ptr(),
			Version:   1,
			ActivityTaskStartedEventAttributes: &types.ActivityTaskStartedEventAttributes{
				ScheduledEventID: 5,
			},
		},
	}

	dbState := &persistence.WorkflowMutableState{
		ExecutionInfo:  info,
		ActivityInfos:  activityInfos,
		BufferedEvents: bufferedEvents,
	}

	s.msBuilder.Load(dbState)
	s.Equal(types.EventTypeActivityTaskCompleted, s.msBuilder.bufferedEvents[0].GetEventType())
	s.Equal(types.EventTypeActivityTaskStarted, s.msBuilder.bufferedEvents[1].GetEventType())

	err := s.msBuilder.FlushBufferedEvents()
	s.Nil(err)
	s.Equal(types.EventTypeActivityTaskStarted, s.msBuilder.hBuilder.history[0].GetEventType())
	s.Equal(int64(8), s.msBuilder.hBuilder.history[0].ID)
	s.Equal(int64(5), s.msBuilder.hBuilder.history[0].ActivityTaskStartedEventAttributes.GetScheduledEventID())
	s.Equal(types.EventTypeActivityTaskCompleted, s.msBuilder.hBuilder.history[1].GetEventType())
	s.Equal(int64(9), s.msBuilder.hBuilder.history[1].ID)
	s.Equal(int64(8), s.msBuilder.hBuilder.history[1].ActivityTaskCompletedEventAttributes.GetStartedEventID())
	s.Equal(int64(5), s.msBuilder.hBuilder.history[1].ActivityTaskCompletedEventAttributes.GetScheduledEventID())
}

func (s *mutableStateSuite) TestChecksum() {
	testCases := []struct {
		name                 string
		enableBufferedEvents bool
		closeTxFunc          func(ms *mutableStateBuilder) (checksum.Checksum, error)
	}{
		{
			name: "closeTransactionAsSnapshot",
			closeTxFunc: func(ms *mutableStateBuilder) (checksum.Checksum, error) {
				snapshot, _, err := ms.CloseTransactionAsSnapshot(time.Now(), TransactionPolicyPassive)
				if err != nil {
					return checksum.Checksum{}, err
				}
				return snapshot.Checksum, err
			},
		},
		{
			name:                 "closeTransactionAsMutation",
			enableBufferedEvents: true,
			closeTxFunc: func(ms *mutableStateBuilder) (checksum.Checksum, error) {
				mutation, _, err := ms.CloseTransactionAsMutation(time.Now(), TransactionPolicyPassive)
				if err != nil {
					return checksum.Checksum{}, err
				}
				return mutation.Checksum, err
			},
		},
	}

	loadErrorsFunc := func() int64 {
		counter := s.testScope.Snapshot().Counters()["test.mutable_state_checksum_mismatch+operation=WorkflowContext"]
		if counter != nil {
			return counter.Value()
		}
		return 0
	}

	var loadErrors int64

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			dbState := s.buildWorkflowMutableState()
			if !tc.enableBufferedEvents {
				dbState.BufferedEvents = nil
			}

			// create mutable state and verify checksum is generated on close
			loadErrors = loadErrorsFunc()
			s.msBuilder.Load(dbState)
			s.Equal(loadErrors, loadErrorsFunc()) // no errors expected
			s.EqualValues(dbState.Checksum, s.msBuilder.checksum)
			s.msBuilder.domainEntry = s.newDomainCacheEntry()
			csum, err := tc.closeTxFunc(s.msBuilder)
			s.Nil(err)
			s.NotNil(csum.Value)
			s.Equal(checksum.FlavorIEEECRC32OverThriftBinary, csum.Flavor)
			s.Equal(mutableStateChecksumPayloadV1, csum.Version)
			s.EqualValues(csum, s.msBuilder.checksum)

			// verify checksum is verified on Load
			dbState.Checksum = csum
			err = s.msBuilder.Load(dbState)
			s.NoError(err)
			s.Equal(loadErrors, loadErrorsFunc())

			// generate checksum again and verify its the same
			csum, err = tc.closeTxFunc(s.msBuilder)
			s.Nil(err)
			s.NotNil(csum.Value)
			s.Equal(dbState.Checksum.Value, csum.Value)

			// modify checksum and verify Load fails
			dbState.Checksum.Value[0]++
			err = s.msBuilder.Load(dbState)
			s.Error(err)
			s.Equal(loadErrors+1, loadErrorsFunc())
			s.EqualValues(dbState.Checksum, s.msBuilder.checksum)

			// test checksum is invalidated
			loadErrors = loadErrorsFunc()
			s.mockShard.GetConfig().MutableStateChecksumInvalidateBefore = func(...dynamicconfig.FilterOption) float64 {
				return float64((s.msBuilder.executionInfo.LastUpdatedTimestamp.UnixNano() / int64(time.Second)) + 1)
			}
			err = s.msBuilder.Load(dbState)
			s.NoError(err)
			s.Equal(loadErrors, loadErrorsFunc())
			s.EqualValues(checksum.Checksum{}, s.msBuilder.checksum)

			// revert the config value for the next test case
			s.mockShard.GetConfig().MutableStateChecksumInvalidateBefore = func(...dynamicconfig.FilterOption) float64 {
				return float64(0)
			}
		})
	}
}

func (s *mutableStateSuite) TestChecksumProbabilities() {
	for _, prob := range []int{0, 100} {
		s.mockShard.GetConfig().MutableStateChecksumGenProbability = func(domain string) int { return prob }
		s.mockShard.GetConfig().MutableStateChecksumVerifyProbability = func(domain string) int { return prob }
		for i := 0; i < 100; i++ {
			shouldGenerate := s.msBuilder.shouldGenerateChecksum()
			shouldVerify := s.msBuilder.shouldVerifyChecksum()
			s.Equal(prob == 100, shouldGenerate)
			s.Equal(prob == 100, shouldVerify)
		}
	}
}

func (s *mutableStateSuite) TestChecksumShouldInvalidate() {
	s.mockShard.GetConfig().MutableStateChecksumInvalidateBefore = func(...dynamicconfig.FilterOption) float64 { return 0 }
	s.False(s.msBuilder.shouldInvalidateChecksum())
	s.msBuilder.executionInfo.LastUpdatedTimestamp = time.Now()
	s.mockShard.GetConfig().MutableStateChecksumInvalidateBefore = func(...dynamicconfig.FilterOption) float64 {
		return float64((s.msBuilder.executionInfo.LastUpdatedTimestamp.UnixNano() / int64(time.Second)) + 1)
	}
	s.True(s.msBuilder.shouldInvalidateChecksum())
	s.mockShard.GetConfig().MutableStateChecksumInvalidateBefore = func(...dynamicconfig.FilterOption) float64 {
		return float64((s.msBuilder.executionInfo.LastUpdatedTimestamp.UnixNano() / int64(time.Second)) - 1)
	}
	s.False(s.msBuilder.shouldInvalidateChecksum())
}

func (s *mutableStateSuite) TestTrimEvents() {
	var input []*types.HistoryEvent
	output := s.msBuilder.trimEventsAfterWorkflowClose(input)
	s.Equal(input, output)

	input = []*types.HistoryEvent{}
	output = s.msBuilder.trimEventsAfterWorkflowClose(input)
	s.Equal(input, output)

	input = []*types.HistoryEvent{
		{
			EventType: types.EventTypeActivityTaskCanceled.Ptr(),
		},
		{
			EventType: types.EventTypeWorkflowExecutionSignaled.Ptr(),
		},
	}
	output = s.msBuilder.trimEventsAfterWorkflowClose(input)
	s.Equal(input, output)

	input = []*types.HistoryEvent{
		{
			EventType: types.EventTypeActivityTaskCanceled.Ptr(),
		},
		{
			EventType: types.EventTypeWorkflowExecutionCompleted.Ptr(),
		},
	}
	output = s.msBuilder.trimEventsAfterWorkflowClose(input)
	s.Equal(input, output)

	input = []*types.HistoryEvent{
		{
			EventType: types.EventTypeWorkflowExecutionCompleted.Ptr(),
		},
		{
			EventType: types.EventTypeActivityTaskCanceled.Ptr(),
		},
	}
	output = s.msBuilder.trimEventsAfterWorkflowClose(input)
	s.Equal([]*types.HistoryEvent{
		{
			EventType: types.EventTypeWorkflowExecutionCompleted.Ptr(),
		},
	}, output)
}

func (s *mutableStateSuite) TestMergeMapOfByteArray() {
	var currentMap map[string][]byte
	var newMap map[string][]byte
	resultMap := mergeMapOfByteArray(currentMap, newMap)
	s.Equal(make(map[string][]byte), resultMap)

	newMap = map[string][]byte{"key": []byte("val")}
	resultMap = mergeMapOfByteArray(currentMap, newMap)
	s.Equal(newMap, resultMap)

	currentMap = map[string][]byte{"number": []byte("1")}
	resultMap = mergeMapOfByteArray(currentMap, newMap)
	s.Equal(2, len(resultMap))
}

func (s *mutableStateSuite) TestEventReapplied() {
	runID := uuid.New()
	eventID := int64(1)
	version := int64(2)
	dedupResource := definition.NewEventReappliedID(runID, eventID, version)
	isReapplied := s.msBuilder.IsResourceDuplicated(dedupResource)
	s.False(isReapplied)
	s.msBuilder.UpdateDuplicatedResource(dedupResource)
	isReapplied = s.msBuilder.IsResourceDuplicated(dedupResource)
	s.True(isReapplied)
}

func (s *mutableStateSuite) TestTransientDecisionTaskSchedule_CurrentVersionChanged() {
	version := int64(2000)
	runID := uuid.New()
	s.msBuilder = NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		version,
		runID,
		constants.TestGlobalDomainEntry,
	).(*mutableStateBuilder)
	_, _ = s.prepareTransientDecisionCompletionFirstBatchReplicated(version, runID)
	err := s.msBuilder.ReplicateDecisionTaskFailedEvent()
	s.NoError(err)

	err = s.msBuilder.UpdateCurrentVersion(version+1, true)
	s.NoError(err)
	versionHistories := s.msBuilder.GetVersionHistories()
	versionHistory, err := versionHistories.GetCurrentVersionHistory()
	s.NoError(err)
	versionHistory.AddOrUpdateItem(&persistence.VersionHistoryItem{
		EventID: 3,
		Version: version,
	})

	now := time.Now()
	di, err := s.msBuilder.AddDecisionTaskScheduledEventAsHeartbeat(true, now.UnixNano())
	s.NoError(err)
	s.NotNil(di)

	s.Equal(int64(0), s.msBuilder.GetExecutionInfo().DecisionAttempt)
	s.Equal(0, len(s.msBuilder.GetHistoryBuilder().transientHistory))
	s.Equal(1, len(s.msBuilder.GetHistoryBuilder().history))
}

func (s *mutableStateSuite) TestTransientDecisionTaskStart_CurrentVersionChanged() {
	version := int64(2000)
	runID := uuid.New()
	s.msBuilder = NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		version,
		runID,
		constants.TestGlobalDomainEntry,
	).(*mutableStateBuilder)
	_, _ = s.prepareTransientDecisionCompletionFirstBatchReplicated(version, runID)
	err := s.msBuilder.ReplicateDecisionTaskFailedEvent()
	s.NoError(err)

	decisionScheduleID := int64(4)
	now := time.Now()
	tasklist := "some random tasklist"
	decisionTimeoutSecond := int32(11)
	decisionAttempt := int64(2)
	newDecisionScheduleEvent := &types.HistoryEvent{
		Version:   version,
		ID:        decisionScheduleID,
		Timestamp: common.Int64Ptr(now.UnixNano()),
		EventType: types.EventTypeDecisionTaskScheduled.Ptr(),
		DecisionTaskScheduledEventAttributes: &types.DecisionTaskScheduledEventAttributes{
			TaskList:                   &types.TaskList{Name: tasklist},
			StartToCloseTimeoutSeconds: common.Int32Ptr(decisionTimeoutSecond),
			Attempt:                    decisionAttempt,
		},
	}
	di, err := s.msBuilder.ReplicateDecisionTaskScheduledEvent(
		newDecisionScheduleEvent.Version,
		newDecisionScheduleEvent.ID,
		newDecisionScheduleEvent.DecisionTaskScheduledEventAttributes.TaskList.GetName(),
		newDecisionScheduleEvent.DecisionTaskScheduledEventAttributes.GetStartToCloseTimeoutSeconds(),
		newDecisionScheduleEvent.DecisionTaskScheduledEventAttributes.GetAttempt(),
		0,
		0,
		false,
	)
	s.NoError(err)
	s.NotNil(di)

	err = s.msBuilder.UpdateCurrentVersion(version+1, true)
	s.NoError(err)
	versionHistories := s.msBuilder.GetVersionHistories()
	versionHistory, err := versionHistories.GetCurrentVersionHistory()
	s.NoError(err)
	versionHistory.AddOrUpdateItem(&persistence.VersionHistoryItem{
		EventID: 3,
		Version: version,
	})

	_, _, err = s.msBuilder.AddDecisionTaskStartedEvent(
		decisionScheduleID,
		uuid.New(),
		&types.PollForDecisionTaskRequest{
			Identity: IdentityHistoryService,
		},
	)
	s.NoError(err)

	s.Equal(0, len(s.msBuilder.GetHistoryBuilder().transientHistory))
	s.Equal(2, len(s.msBuilder.GetHistoryBuilder().history))
}

func (s *mutableStateSuite) prepareTransientDecisionCompletionFirstBatchReplicated(version int64, runID string) (*types.HistoryEvent, *types.HistoryEvent) {
	domainID := constants.TestDomainID
	execution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      runID,
	}

	now := time.Now()
	workflowType := "some random workflow type"
	tasklist := "some random tasklist"
	workflowTimeoutSecond := int32(222)
	decisionTimeoutSecond := int32(11)
	decisionAttempt := int64(0)
	partitionConfig := map[string]string{
		"zone": "dca",
	}

	eventID := int64(1)
	workflowStartEvent := &types.HistoryEvent{
		Version:   version,
		ID:        eventID,
		Timestamp: common.Int64Ptr(now.UnixNano()),
		EventType: types.EventTypeWorkflowExecutionStarted.Ptr(),
		WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
			WorkflowType:                        &types.WorkflowType{Name: workflowType},
			TaskList:                            &types.TaskList{Name: tasklist},
			Input:                               nil,
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(workflowTimeoutSecond),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(decisionTimeoutSecond),
			PartitionConfig:                     partitionConfig,
		},
	}
	eventID++

	decisionScheduleEvent := &types.HistoryEvent{
		Version:   version,
		ID:        eventID,
		Timestamp: common.Int64Ptr(now.UnixNano()),
		EventType: types.EventTypeDecisionTaskScheduled.Ptr(),
		DecisionTaskScheduledEventAttributes: &types.DecisionTaskScheduledEventAttributes{
			TaskList:                   &types.TaskList{Name: tasklist},
			StartToCloseTimeoutSeconds: common.Int32Ptr(decisionTimeoutSecond),
			Attempt:                    decisionAttempt,
		},
	}
	eventID++

	decisionStartedEvent := &types.HistoryEvent{
		Version:   version,
		ID:        eventID,
		Timestamp: common.Int64Ptr(now.UnixNano()),
		EventType: types.EventTypeDecisionTaskStarted.Ptr(),
		DecisionTaskStartedEventAttributes: &types.DecisionTaskStartedEventAttributes{
			ScheduledEventID: decisionScheduleEvent.ID,
			RequestID:        uuid.New(),
		},
	}
	eventID++

	_ = &types.HistoryEvent{
		Version:   version,
		ID:        eventID,
		Timestamp: common.Int64Ptr(now.UnixNano()),
		EventType: types.EventTypeDecisionTaskFailed.Ptr(),
		DecisionTaskFailedEventAttributes: &types.DecisionTaskFailedEventAttributes{
			ScheduledEventID: decisionScheduleEvent.ID,
			StartedEventID:   decisionStartedEvent.ID,
		},
	}
	eventID++

	s.mockEventsCache.EXPECT().PutEvent(
		domainID, execution.GetWorkflowID(), execution.GetRunID(),
		workflowStartEvent.ID, workflowStartEvent,
	).Times(1)
	err := s.msBuilder.ReplicateWorkflowExecutionStartedEvent(
		nil,
		execution,
		uuid.New(),
		workflowStartEvent,
		false,
	)
	s.Nil(err)

	// setup transient decision
	di, err := s.msBuilder.ReplicateDecisionTaskScheduledEvent(
		decisionScheduleEvent.Version,
		decisionScheduleEvent.ID,
		decisionScheduleEvent.DecisionTaskScheduledEventAttributes.TaskList.GetName(),
		decisionScheduleEvent.DecisionTaskScheduledEventAttributes.GetStartToCloseTimeoutSeconds(),
		decisionScheduleEvent.DecisionTaskScheduledEventAttributes.GetAttempt(),
		0,
		0,
		false,
	)
	s.Nil(err)
	s.NotNil(di)

	di, err = s.msBuilder.ReplicateDecisionTaskStartedEvent(nil,
		decisionStartedEvent.Version,
		decisionScheduleEvent.ID,
		decisionStartedEvent.ID,
		decisionStartedEvent.DecisionTaskStartedEventAttributes.GetRequestID(),
		decisionStartedEvent.GetTimestamp(),
	)
	s.Nil(err)
	s.NotNil(di)

	err = s.msBuilder.ReplicateDecisionTaskFailedEvent()
	s.Nil(err)

	decisionAttempt = int64(123)
	newDecisionScheduleEvent := &types.HistoryEvent{
		Version:   version,
		ID:        eventID,
		Timestamp: common.Int64Ptr(now.UnixNano()),
		EventType: types.EventTypeDecisionTaskScheduled.Ptr(),
		DecisionTaskScheduledEventAttributes: &types.DecisionTaskScheduledEventAttributes{
			TaskList:                   &types.TaskList{Name: tasklist},
			StartToCloseTimeoutSeconds: common.Int32Ptr(decisionTimeoutSecond),
			Attempt:                    decisionAttempt,
		},
	}
	eventID++

	newDecisionStartedEvent := &types.HistoryEvent{
		Version:   version,
		ID:        eventID,
		Timestamp: common.Int64Ptr(now.UnixNano()),
		EventType: types.EventTypeDecisionTaskStarted.Ptr(),
		DecisionTaskStartedEventAttributes: &types.DecisionTaskStartedEventAttributes{
			ScheduledEventID: decisionScheduleEvent.ID,
			RequestID:        uuid.New(),
		},
	}
	eventID++ //nolint:ineffassign

	di, err = s.msBuilder.ReplicateDecisionTaskScheduledEvent(
		newDecisionScheduleEvent.Version,
		newDecisionScheduleEvent.ID,
		newDecisionScheduleEvent.DecisionTaskScheduledEventAttributes.TaskList.GetName(),
		newDecisionScheduleEvent.DecisionTaskScheduledEventAttributes.GetStartToCloseTimeoutSeconds(),
		newDecisionScheduleEvent.DecisionTaskScheduledEventAttributes.GetAttempt(),
		0,
		0,
		false,
	)
	s.Nil(err)
	s.NotNil(di)

	di, err = s.msBuilder.ReplicateDecisionTaskStartedEvent(nil,
		newDecisionStartedEvent.Version,
		newDecisionScheduleEvent.ID,
		newDecisionStartedEvent.ID,
		newDecisionStartedEvent.DecisionTaskStartedEventAttributes.GetRequestID(),
		newDecisionStartedEvent.GetTimestamp(),
	)
	s.Nil(err)
	s.NotNil(di)

	return newDecisionScheduleEvent, newDecisionStartedEvent
}

func (s *mutableStateSuite) TestLoad_BackwardsCompatibility() {
	mutableState := s.buildWorkflowMutableState()

	s.msBuilder.Load(mutableState)

	s.Equal(constants.TestDomainID, s.msBuilder.pendingChildExecutionInfoIDs[81].DomainID)
}

func (s *mutableStateSuite) TestUpdateCurrentVersion_WorkflowOpen() {
	mutableState := s.buildWorkflowMutableState()

	s.msBuilder.Load(mutableState)
	s.Equal(common.EmptyVersion, s.msBuilder.GetCurrentVersion())

	version := int64(2000)
	s.msBuilder.UpdateCurrentVersion(version, false)
	s.Equal(version, s.msBuilder.GetCurrentVersion())
}

func (s *mutableStateSuite) TestUpdateCurrentVersion_WorkflowClosed() {
	mutableState := s.buildWorkflowMutableState()
	mutableState.ExecutionInfo.State = persistence.WorkflowStateCompleted
	mutableState.ExecutionInfo.CloseStatus = persistence.WorkflowCloseStatusCompleted

	s.msBuilder.Load(mutableState)
	s.Equal(common.EmptyVersion, s.msBuilder.GetCurrentVersion())

	versionHistory, err := mutableState.VersionHistories.GetCurrentVersionHistory()
	s.NoError(err)
	lastItem, err := versionHistory.GetLastItem()
	s.NoError(err)
	lastWriteVersion := lastItem.Version

	version := int64(2000)
	s.msBuilder.UpdateCurrentVersion(version, false)
	s.Equal(lastWriteVersion, s.msBuilder.GetCurrentVersion())
}

func (s *mutableStateSuite) newDomainCacheEntry() *cache.DomainCacheEntry {
	return cache.NewDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: "mutableStateTest"},
		&persistence.DomainConfig{},
		true,
		&persistence.DomainReplicationConfig{},
		1,
		nil,
	)
}

func (s *mutableStateSuite) buildWorkflowMutableState() *persistence.WorkflowMutableState {
	domainID := constants.TestDomainID
	we := types.WorkflowExecution{
		WorkflowID: "wId",
		RunID:      constants.TestRunID,
	}
	tl := "testTaskList"
	failoverVersion := int64(300)
	partitionConfig := map[string]string{
		"zone": "phx",
	}

	info := &persistence.WorkflowExecutionInfo{
		DomainID:                    domainID,
		WorkflowID:                  we.GetWorkflowID(),
		RunID:                       we.GetRunID(),
		TaskList:                    tl,
		WorkflowTypeName:            "wType",
		WorkflowTimeout:             200,
		DecisionStartToCloseTimeout: 100,
		State:                       persistence.WorkflowStateRunning,
		CloseStatus:                 persistence.WorkflowCloseStatusNone,
		NextEventID:                 int64(101),
		LastProcessedEvent:          int64(99),
		LastUpdatedTimestamp:        time.Now(),
		DecisionVersion:             failoverVersion,
		DecisionScheduleID:          common.EmptyEventID,
		DecisionStartedID:           common.EmptyEventID,
		DecisionTimeout:             100,
		PartitionConfig:             partitionConfig,
	}

	activityInfos := map[int64]*persistence.ActivityInfo{
		5: {
			Version:                failoverVersion,
			ScheduleID:             int64(5),
			ScheduledTime:          time.Now(),
			StartedID:              common.EmptyEventID,
			StartedTime:            time.Now(),
			ActivityID:             "activityID_5",
			ScheduleToStartTimeout: 100,
			ScheduleToCloseTimeout: 200,
			StartToCloseTimeout:    300,
			HeartbeatTimeout:       50,
		},
	}

	timerInfos := map[string]*persistence.TimerInfo{
		"25": {
			Version:    failoverVersion,
			TimerID:    "25",
			StartedID:  85,
			ExpiryTime: time.Now().Add(time.Hour),
		},
	}

	childInfos := map[int64]*persistence.ChildExecutionInfo{
		80: {
			Version:               failoverVersion,
			InitiatedID:           80,
			InitiatedEventBatchID: 20,
			InitiatedEvent:        &types.HistoryEvent{},
			StartedID:             common.EmptyEventID,
			CreateRequestID:       uuid.New(),
			DomainID:              constants.TestDomainID,
			WorkflowTypeName:      "code.uber.internal/test/foobar",
		},
		81: {
			Version:               failoverVersion,
			InitiatedID:           80,
			InitiatedEventBatchID: 20,
			InitiatedEvent:        &types.HistoryEvent{},
			StartedID:             common.EmptyEventID,
			CreateRequestID:       uuid.New(),
			DomainNameDEPRECATED:  constants.TestDomainName,
			WorkflowTypeName:      "code.uber.internal/test/foobar",
		},
	}

	signalInfos := map[int64]*persistence.SignalInfo{
		75: {
			Version:               failoverVersion,
			InitiatedID:           75,
			InitiatedEventBatchID: 17,
			SignalRequestID:       uuid.New(),
			SignalName:            "test-signal-75",
			Input:                 []byte("signal-input-75"),
		},
	}

	signalRequestIDs := map[string]struct{}{
		uuid.New(): {},
	}

	bufferedEvents := []*types.HistoryEvent{
		{
			ID:        common.BufferedEventID,
			EventType: types.EventTypeWorkflowExecutionSignaled.Ptr(),
			Version:   failoverVersion,
			WorkflowExecutionSignaledEventAttributes: &types.WorkflowExecutionSignaledEventAttributes{
				SignalName: "test-signal-buffered",
				Input:      []byte("test-signal-buffered-input"),
			},
		},
	}

	versionHistories := &persistence.VersionHistories{
		CurrentVersionHistoryIndex: 0,
		Histories: []*persistence.VersionHistory{
			{
				BranchToken: []byte("token#1"),
				Items: []*persistence.VersionHistoryItem{
					{EventID: 1, Version: 300},
				},
			},
		},
	}

	return &persistence.WorkflowMutableState{
		ExecutionInfo:       info,
		ActivityInfos:       activityInfos,
		TimerInfos:          timerInfos,
		ChildExecutionInfos: childInfos,
		SignalInfos:         signalInfos,
		SignalRequestedIDs:  signalRequestIDs,
		BufferedEvents:      bufferedEvents,
		VersionHistories:    versionHistories,
	}
}

func TestNewMutableStateBuilderWithEventV2(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockShard := shard.NewTestContext(
		t,
		ctrl,
		&persistence.ShardInfo{
			ShardID:          0,
			RangeID:          1,
			TransferAckLevel: 0,
		},
		config.NewForTest(),
	)
	domainCache := cache.NewDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: "mutableStateTest"},
		&persistence.DomainConfig{},
		true,
		&persistence.DomainReplicationConfig{},
		1,
		nil,
	)

	NewMutableStateBuilderWithEventV2(mockShard, log.NewNoop(), "A82146B5-7A5C-4660-9195-E80E5161EC56", domainCache)
}

var (
	domainID = "A6338800-D143-4FEF-8A49-9BBB31386C5F"
	wfID     = "879A361B-B435-491D-8A3B-ACF3BAD30F4B"
	runID    = "81DFCB6B-ACD4-46D1-89C2-804388203880"
	ts1      = int64(1234)
	shardID  = 123
)

// Guiding real data example: ie:
// `select execution from executions where run_id = <run-id> ALLOW FILTERING;`
//
// executions.execution {
// domainID: "A6338800-D143-4FEF-8A49-9BBB31386C5F",
// wfID: "879A361B-B435-491D-8A3B-ACF3BAD30F4B",
// runID: "81DFCB6B-ACD4-46D1-89C2-804388203880",
// initiated_id: -7,
// completion_event: null,
// state: 2,
// close_status: 1,
// next_event_id: 12,
// last_processed_event: 9,
// decision_schedule_id: -23,
// decision_started_id: -23,
// last_first_event_id: 10,
// decision_version: -24,
// completion_event_batch_id: 10,
// last_event_task_id: 4194328,
// }
var exampleMutableStateForClosedWF = &mutableStateBuilder{
	executionInfo: &persistence.WorkflowExecutionInfo{
		WorkflowID:             wfID,
		DomainID:               domainID,
		RunID:                  runID,
		NextEventID:            12,
		State:                  persistence.WorkflowStateCompleted,
		CompletionEventBatchID: 10,
		BranchToken:            []byte("branch-token"),
	},
}

var exampleCompletionEvent = &types.HistoryEvent{
	ID:        11,
	TaskID:    4194328,
	Version:   1,
	Timestamp: &ts1,
	WorkflowExecutionCompletedEventAttributes: &types.WorkflowExecutionCompletedEventAttributes{
		Result:                       []byte("some random workflow completion result"),
		DecisionTaskCompletedEventID: 10,
	},
}

var exampleStartEvent = &types.HistoryEvent{
	ID:        1,
	TaskID:    4194328,
	Version:   1,
	Timestamp: &ts1,
	WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
		WorkflowType:                        &types.WorkflowType{Name: "workflow-type"},
		TaskList:                            &types.TaskList{Name: "tasklist"},
		Input:                               []byte("some random workflow input"),
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(60),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		OriginalExecutionRunID:              runID,
		Identity:                            "123@some-hostname@@uuid",
	},
}

func TestGetCompletionEvent(t *testing.T) {
	tests := map[string]struct {
		currentState *mutableStateBuilder

		historyManagerAffordance func(historyManager *persistence.MockHistoryManager)

		expectedResult *types.HistoryEvent
		expectedErr    error
	}{
		"Getting a completed event from a normal, completed workflow - taken from a real example": {
			currentState: exampleMutableStateForClosedWF,
			historyManagerAffordance: func(historyManager *persistence.MockHistoryManager) {

				historyManager.EXPECT().ReadHistoryBranch(gomock.Any(),
					&persistence.ReadHistoryBranchRequest{
						BranchToken:   []byte("branch-token"),
						MinEventID:    10,
						MaxEventID:    12, // nextEventID +1
						PageSize:      1,
						NextPageToken: nil,
						ShardID:       common.IntPtr(shardID),
						DomainName:    "domain",
					}).Return(&persistence.ReadHistoryBranchResponse{
					HistoryEvents: []*types.HistoryEvent{
						exampleCompletionEvent,
					},
				}, nil)

			},

			expectedResult: exampleCompletionEvent,
		},
		"An unexpected error while fetchhing history, such as not found err": {
			currentState: &mutableStateBuilder{
				executionInfo: &persistence.WorkflowExecutionInfo{
					WorkflowID:             wfID,
					DomainID:               domainID,
					RunID:                  runID,
					NextEventID:            12,
					State:                  persistence.WorkflowStateCompleted,
					CompletionEventBatchID: 10,
					BranchToken:            []byte("branch-token"),
				},
			},
			historyManagerAffordance: func(historyManager *persistence.MockHistoryManager) {
				historyManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).Return(nil, errors.New("a transient random error"))
			},

			expectedResult: nil,
			expectedErr:    &types.InternalServiceError{Message: "unable to get workflow completion event"},
		},
		"A 'transient' internal service error, this should be returned to the caller": {
			currentState: &mutableStateBuilder{
				executionInfo: &persistence.WorkflowExecutionInfo{
					WorkflowID:             wfID,
					DomainID:               domainID,
					RunID:                  runID,
					NextEventID:            12,
					State:                  persistence.WorkflowStateCompleted,
					CompletionEventBatchID: 10,
					BranchToken:            []byte("branch-token"),
				},
			},
			historyManagerAffordance: func(historyManager *persistence.MockHistoryManager) {
				historyManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).
					Return(nil, &types.InternalServiceError{Message: "an err"})
			},

			expectedResult: nil,
			expectedErr:    &types.InternalServiceError{Message: "an err"},
		},
		"initial validation: An invalid starting mutable state should return an error": {
			currentState: &mutableStateBuilder{
				executionInfo: &persistence.WorkflowExecutionInfo{},
			},
			historyManagerAffordance: func(historyManager *persistence.MockHistoryManager) {},
			expectedResult:           nil,
			expectedErr:              &types.InternalServiceError{Message: "unable to get workflow completion event"},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {

			ctrl := gomock.NewController(t)

			shardContext := shardCtx.NewMockContext(ctrl)
			shardContext.EXPECT().GetShardID().Return(123).AnyTimes() // this isn't called on a few of the validation failures
			historyManager := persistence.NewMockHistoryManager(ctrl)
			td.historyManagerAffordance(historyManager)

			domainCache := cache.NewMockDomainCache(ctrl)
			domainCache.EXPECT().GetDomainName(gomock.Any()).Return("domain", nil).AnyTimes() // this isn't called on validation

			td.currentState.eventsCache = events.NewCache(shardID, historyManager, config.NewForTest(), log.NewNoop(), metrics.NewNoopMetricsClient(), domainCache)
			td.currentState.shard = shardContext

			res, err := td.currentState.GetCompletionEvent(context.Background())

			assert.Equal(t, td.expectedResult, res)
			if td.expectedErr != nil {
				assert.ErrorAs(t, td.expectedErr, &err)
			}
		})
	}
}

func TestGetStartEvent(t *testing.T) {
	tests := map[string]struct {
		currentState *mutableStateBuilder

		historyManagerAffordance func(historyManager *persistence.MockHistoryManager)

		expectedResult *types.HistoryEvent
		expectedErr    error
	}{
		"Getting a start event from a normal, completed workflow - taken from a real example": {
			currentState: exampleMutableStateForClosedWF,
			historyManagerAffordance: func(historyManager *persistence.MockHistoryManager) {

				historyManager.EXPECT().ReadHistoryBranch(gomock.Any(),
					&persistence.ReadHistoryBranchRequest{
						BranchToken:   []byte("branch-token"),
						MinEventID:    1,
						MaxEventID:    2,
						PageSize:      1,
						NextPageToken: nil,
						ShardID:       common.IntPtr(shardID),
						DomainName:    "domain",
					}).Return(&persistence.ReadHistoryBranchResponse{
					HistoryEvents: []*types.HistoryEvent{
						exampleStartEvent,
					},
				}, nil)

			},

			expectedResult: exampleStartEvent,
		},
		"Getting a start event but hitting an error when reaching into history": {
			currentState: exampleMutableStateForClosedWF,
			historyManagerAffordance: func(historyManager *persistence.MockHistoryManager) {
				historyManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).Return(nil, errors.New("an error"))
			},
			expectedErr: types.InternalServiceError{Message: "unable to get workflow start event"},
		},
		"Getting a start event but hitting a 'transient' error when reaching into history. This should be passed back up the call stack": {
			currentState: exampleMutableStateForClosedWF,
			historyManagerAffordance: func(historyManager *persistence.MockHistoryManager) {
				historyManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).Return(nil, &types.InternalServiceError{Message: "an error"})
			},
			expectedErr: types.InternalServiceError{Message: "an error"},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {

			ctrl := gomock.NewController(t)

			shardContext := shardCtx.NewMockContext(ctrl)
			shardContext.EXPECT().GetShardID().Return(123).AnyTimes() // this isn't called on a few of the validation failures
			historyManager := persistence.NewMockHistoryManager(ctrl)
			td.historyManagerAffordance(historyManager)

			domainCache := cache.NewMockDomainCache(ctrl)
			domainCache.EXPECT().GetDomainName(gomock.Any()).Return("domain", nil).AnyTimes() // this isn't called on validation

			td.currentState.eventsCache = events.NewCache(shardID, historyManager, config.NewForTest(), log.NewNoop(), metrics.NewNoopMetricsClient(), domainCache)
			td.currentState.shard = shardContext

			res, err := td.currentState.GetStartEvent(context.Background())

			assert.Equal(t, td.expectedResult, res)
			if td.expectedErr != nil {
				assert.ErrorAs(t, err, &td.expectedErr)
			}
		})
	}
}

// func TestAddContinueAsNewEvent(t *testing.T) {

// 	firstEventID := int64(15)
// 	decisionCompletedEventID := int64(15)

// 	startExecutionInfo := &persistence.WorkflowExecutionInfo{
// 		DomainID:                           "5391dbea-5b30-4323-82ca-e1c95339bb3e",
// 		WorkflowID:                         "helloworld_bfa410d9-9f49-4bcb-a943-0f3ceb252da2",
// 		RunID:                              "c6702a46-a1f0-42d2-9de7-8aca6ed6795f",
// 		FirstExecutionRunID:                "c6702a46-a1f0-42d2-9de7-8aca6ed6795f",
// 		InitiatedID:                        -7,
// 		CompletionEventBatchID:             0,
// 		TaskList:                           "tasklist-name",
// 		WorkflowTypeName:                   "test-workflow",
// 		WorkflowTimeout:                    60,
// 		DecisionStartToCloseTimeout:        60,
// 		State:                              1,
// 		CloseStatus:                        0,
// 		LastFirstEventID:                   14,
// 		LastEventTaskID:                    11534369,
// 		NextEventID:                        16,
// 		LastProcessedEvent:                 14,
// 		CreateRequestID:                    "5c0be655-1efc-4dfe-8f69-1f59e59c13ef",
// 		DecisionVersion:                    -24,
// 		DecisionScheduleID:                 -23,
// 		DecisionStartedID:                  -23,
// 		DecisionRequestID:                  "emptyUuid",
// 		DecisionTimeout:                    0,
// 		DecisionAttempt:                    0,
// 		DecisionStartedTimestamp:           0,
// 		DecisionScheduledTimestamp:         0,
// 		DecisionOriginalScheduledTimestamp: 1709790036041553000,
// 		CancelRequested:                    false,
// 		StickyTaskList:                     "david-porter-DVFG73D710:04be47fa-2381-469f-b2ea-1253271ad116",
// 		StickyScheduleToStartTimeout:       5,
// 		ClientLibraryVersion:               "0.18.4",
// 		ClientFeatureVersion:               "1.7.0",
// 		ClientImpl:                         "uber-go",
// 		Attempt:                            0,
// 	}

// 	endExecutionInfo := &persistence.WorkflowExecutionInfo{
// 		DomainID:                           "5391dbea-5b30-4323-82ca-e1c95339bb3e",
// 		WorkflowID:                         "helloworld_bfa410d9-9f49-4bcb-a943-0f3ceb252da2",
// 		RunID:                              "c6702a46-a1f0-42d2-9de7-8aca6ed6795f",
// 		FirstExecutionRunID:                "c6702a46-a1f0-42d2-9de7-8aca6ed6795f",
// 		InitiatedID:                        -7,
// 		CompletionEventBatchID:             15,
// 		TaskList:                           "tasklist-name",
// 		WorkflowTypeName:                   "test-workflow",
// 		WorkflowTimeout:                    60,
// 		DecisionStartToCloseTimeout:        60,
// 		State:                              2,
// 		CloseStatus:                        5,
// 		LastFirstEventID:                   14,
// 		LastEventTaskID:                    11534369,
// 		NextEventID:                        17,
// 		LastProcessedEvent:                 14,
// 		CreateRequestID:                    "5c0be655-1efc-4dfe-8f69-1f59e59c13ef",
// 		DecisionVersion:                    -24,
// 		DecisionScheduleID:                 -23,
// 		DecisionStartedID:                  -23,
// 		DecisionRequestID:                  "emptyUuid",
// 		DecisionTimeout:                    0,
// 		DecisionAttempt:                    0,
// 		DecisionOriginalScheduledTimestamp: 1709790036041553000,
// 	}

// 	startHistory := []*types.HistoryEvent{
// 		{
// 			ID:        15,
// 			Timestamp: common.Ptr(int64(1709791556528026000)),
// 			EventType: types.EventTypeDecisionTaskCompleted.Ptr(),
// 			Version:   1,
// 			TaskID:    -1234,
// 			DecisionTaskCompletedEventAttributes: &types.DecisionTaskCompletedEventAttributes{
// 				ScheduledEventID: 13,
// 				StartedEventID:   14,
// 				Identity:         "27368@david-porter-DVFG73D710@helloWorldGroup@6027e9ee-048e-4f67-8d88-27883c496901",
// 				BinaryChecksum:   "6df03bf5110d681667852a8456519536",
// 			},
// 		},
// 	}

// 	endHistory := []*types.HistoryEvent{
// 		{
// 			ID:        15,
// 			Timestamp: common.Ptr(int64(1709791556528026000)),
// 			EventType: types.EventTypeDecisionTaskCompleted.Ptr(),
// 			Version:   1,
// 			TaskID:    -1234,
// 			DecisionTaskCompletedEventAttributes: &types.DecisionTaskCompletedEventAttributes{
// 				ScheduledEventID: 13,
// 				StartedEventID:   14,
// 				Identity:         "27368@david-porter-DVFG73D710@helloWorldGroup@6027e9ee-048e-4f67-8d88-27883c496901",
// 				BinaryChecksum:   "6df03bf5110d681667852a8456519536",
// 			},
// 		},
// 		{
// 			ID:        16,
// 			Timestamp: common.Ptr(int64(1709791556529788000)),
// 			EventType: common.Ptr(types.EventTypeWorkflowExecutionContinuedAsNew),
// 			Version:   1,
// 			TaskID:    -1234,
// 			WorkflowExecutionContinuedAsNewEventAttributes: &types.WorkflowExecutionContinuedAsNewEventAttributes{
// 				NewExecutionRunID: "1b094f71-9c23-4177-8cf9-7f723cc52955",
// 				WorkflowType: &types.WorkflowType{
// 					Name: "helloWorldWorkflow",
// 				},
// 				TaskList: &types.TaskList{
// 					Name: "helloWorldGroup",
// 				},
// 				Input:                               []byte("some-input"),
// 				ExecutionStartToCloseTimeoutSeconds: common.Ptr(int32(60)),
// 				TaskStartToCloseTimeoutSeconds:      common.Ptr(int32(60)),
// 				DecisionTaskCompletedEventID:        15,
// 				Initiator:                           common.Ptr(types.ContinueAsNewInitiatorDecider),
// 			},
// 		},
// 	}

// 	tests := map[string]struct {
// 		currentState *mutableStateBuilder
// 		// history is a substruct of current state, but because they're both
// 		// pointing to each other, they're assembled at the test start
// 		currentHistoryBuilder *HistoryBuilder

// 		// expectations
// 		historyManagerAffordance func(historyManager *persistence.MockHistoryManager)

// 		// this is a somewhat confusing API, both returning a new cloned state and mutating the existing
// 		// current state so this will be comparing both (they should be the same)
// 		expectedEndState *mutableStateBuilder
// 		expectedErr      error
// 	}{
// 		"a continue-as-new event with no errors": {
// 			currentHistoryBuilder: &HistoryBuilder{
// 				history: startHistory,
// 			},
// 			currentState: &mutableStateBuilder{
// 				domainEntry:   cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{ID: domainID}, &persistence.DomainConfig{}, true, &persistence.DomainReplicationConfig{}, 1, nil),
// 				executionInfo: startExecutionInfo,
// 				logger:        log.NewNoop(),
// 				config:        config.NewForTest(),
// 			},
// 			historyManagerAffordance: func(historyManager *persistence.MockHistoryManager) {
// 				// when it gets the first-run ID
// 				historyManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ReadHistoryBranchResponse{
// 					HistoryEvents: []*types.HistoryEvent{
// 						{
// 							ID: 1,
// 							WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
// 								FirstExecutionRunID: "some-first-run-id",
// 								WorkflowType: &types.WorkflowType{
// 									Name: "test-workflow",
// 								},
// 							},
// 						},
// 					},
// 				}, nil)
// 			},

// 			expectedEndState: &mutableStateBuilder{
// 				hBuilder: &HistoryBuilder{
// 					history: endHistory, // insert Fukuyama joke
// 				},
// 				executionInfo: endExecutionInfo,
// 			},
// 		},
// 	}

// 	for name, td := range tests {
// 		t.Run(name, func(t *testing.T) {

// 			ctrl := gomock.NewController(t)

// 			historyManager := persistence.NewMockHistoryManager(ctrl)
// 			td.historyManagerAffordance(historyManager)

// 			domainCache := cache.NewMockDomainCache(ctrl)
// 			domainCache.EXPECT().GetDomainName(gomock.Any()).Return("domain", nil)

// 			td.currentState.eventsCache = events.NewCache(shardID, historyManager, config.NewForTest(), log.NewNoop(), metrics.NewNoopMetricsClient(), domainCache)

// 			shardContext := shardCtx.NewMockContext(ctrl)
// 			shardContext.EXPECT().GetShardID().Return(123)
// 			shardContext.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(2)
// 			shardContext.EXPECT().GetEventsCache().Return(td.currentState.eventsCache)
// 			shardContext.EXPECT().GetConfig().Return(td.currentState.config)
// 			shardContext.EXPECT().GetTimeSource().Return(clock.NewMockedTimeSource())
// 			shardContext.EXPECT().GetMetricsClient().Return(metrics.NewNoopMetricsClient())
// 			shardContext.EXPECT().GetDomainCache().Return(domainCache)

// 			taskGenerator := NewMockMutableStateTaskGenerator(ctrl)

// 			taskGenerator.EXPECT().GenerateWorkflowCloseTasks(gomock.Any(), td.currentState.config.WorkflowDeletionJitterRange("domain"))

// 			td.currentState.shard = shardContext
// 			td.currentHistoryBuilder.msBuilder = td.currentState
// 			td.currentState.hBuilder = td.currentHistoryBuilder
// 			td.currentState.taskGenerator = taskGenerator

// 			td.currentState.timeSource = clock.NewMockedTimeSourceAt(time.Unix(0, ts1))

// 			_, ms, err := td.currentState.AddContinueAsNewEvent(context.Background(), firstEventID, decisionCompletedEventID, "", &types.ContinueAsNewWorkflowExecutionDecisionAttributes{
// 				WorkflowType: &types.WorkflowType{Name: "test-workflow"},
// 				TaskList:     &types.TaskList{Name: "tasklist-name"},
// 				Input:        []byte("some random workflow input"),
// 			})

// 			assert.Equal(t, td.expectedEndState, ms)
// 			if td.expectedErr != nil {
// 				assert.ErrorAs(t, err, &td.expectedErr)
// 			}
// 		})
// 	}
// }
