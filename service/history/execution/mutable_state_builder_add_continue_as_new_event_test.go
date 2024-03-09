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
	// "errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	// "github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
	// "github.com/stretchr/testify/suite"
	// "github.com/uber-go/tally"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	// "github.com/uber/cadence/common/checksum"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	// "github.com/uber/cadence/common/definition"
	// "github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	// "github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/events"
	// "github.com/uber/cadence/service/history/shard"
	shardCtx "github.com/uber/cadence/service/history/shard"
)

func TestAddContinueAsNewEvent(t *testing.T) {

	firstEventID := int64(15)
	decisionCompletedEventID := int64(15)

	var (
		domainID = "5391dbea-5b30-4323-82ca-e1c95339bb3e"
		ts1      = int64(1709872131923568000)
		shardID  = 123
	)

	startingExecutionInfo := &persistence.WorkflowExecutionInfo{
		DomainID:                           "5391dbea-5b30-4323-82ca-e1c95339bb3e",
		WorkflowID:                         "helloworld_b4db8bd0-74b7-4250-ade7-ac72a1efb171",
		RunID:                              "5adce5c5-b7b2-4418-9bf0-4207303f6343",
		FirstExecutionRunID:                "5adce5c5-b7b2-4418-9bf0-4207303f6343",
		InitiatedID:                        -7,
		TaskList:                           "helloWorldGroup",
		WorkflowTypeName:                   "helloWorldWorkflow",
		WorkflowTimeout:                    60,
		DecisionStartToCloseTimeout:        60,
		State:                              1,
		LastFirstEventID:                   14,
		LastEventTaskID:                    15728673,
		NextEventID:                        16,
		LastProcessedEvent:                 14,
		StartTimestamp:                     time.Date(2024, 3, 8, 4, 28, 41, 415000000, time.UTC),
		LastUpdatedTimestamp:               time.Date(2024, 3, 7, 20, 28, 51, 563480000, time.Local),
		CreateRequestID:                    "b086d62c-dd2b-4bbc-9143-5940516acbfe",
		DecisionVersion:                    -24,
		DecisionScheduleID:                 -23,
		DecisionStartedID:                  -23,
		DecisionRequestID:                  "emptyUuid",
		DecisionOriginalScheduledTimestamp: 1709872131542474000,
		StickyTaskList:                     "david-porter-DVFG73D710:04be47fa-2381-469f-b2ea-1253271ad116",
		StickyScheduleToStartTimeout:       5,
		ClientLibraryVersion:               "0.18.4",
		ClientFeatureVersion:               "1.7.0",
		ClientImpl:                         "uber-go",
		AutoResetPoints: &types.ResetPoints{
			Points: []*types.ResetPointInfo{{
				BinaryChecksum:           "6df03bf5110d681667852a8456519536",
				RunID:                    "5adce5c5-b7b2-4418-9bf0-4207303f6343",
				FirstDecisionCompletedID: 4,
				CreatedTimeNano:          common.Ptr(int64(1709872121495904000)),
				Resettable:               true,
			}},
		},
		SearchAttributes: map[string][]uint8{"BinaryChecksums": {91, 34, 54, 100, 102, 48, 51, 98, 102, 53, 49, 49, 48, 100, 54, 56, 49, 54, 54, 55, 56, 53, 50, 97, 56, 52, 53, 54, 53, 49, 57, 53, 51, 54, 34, 93}}}

	startingHistory := []*types.HistoryEvent{{
		ID:        15,
		Timestamp: common.Ptr(int64(1709872131580456000)),
		EventType: common.Ptr(types.EventTypeDecisionTaskCompleted),
		Version:   1,
		TaskID:    -1234,
		DecisionTaskCompletedEventAttributes: &types.DecisionTaskCompletedEventAttributes{
			ScheduledEventID: 13,
			StartedEventID:   14,
			Identity:         "27368@david-porter-DVFG73D710@helloWorldGroup@6027e9ee-048e-4f67-8d88-27883c496901",
			BinaryChecksum:   "6df03bf5110d681667852a8456519536",
		},
	}}

	expectedEndingBuilderExecutionInfo := &persistence.WorkflowExecutionInfo{
		DomainID:                           "5391dbea-5b30-4323-82ca-e1c95339bb3e",
		WorkflowID:                         "helloworld_b4db8bd0-74b7-4250-ade7-ac72a1efb171",
		RunID:                              "5adce5c5-b7b2-4418-9bf0-4207303f6343",
		FirstExecutionRunID:                "5adce5c5-b7b2-4418-9bf0-4207303f6343",
		InitiatedID:                        -7,
		CompletionEventBatchID:             15,
		TaskList:                           "helloWorldGroup",
		WorkflowTypeName:                   "helloWorldWorkflow",
		WorkflowTimeout:                    60,
		DecisionStartToCloseTimeout:        60,
		State:                              2,
		CloseStatus:                        5,
		LastFirstEventID:                   14,
		LastEventTaskID:                    15728673,
		NextEventID:                        17,
		LastProcessedEvent:                 14,
		StartTimestamp:                     time.Date(2024, 3, 8, 4, 28, 41, 415000000, time.UTC),
		LastUpdatedTimestamp:               time.Date(2024, 3, 7, 20, 28, 51, 563480000, time.Local),
		CreateRequestID:                    "b086d62c-dd2b-4bbc-9143-5940516acbfe",
		DecisionVersion:                    -24,
		DecisionScheduleID:                 -23,
		DecisionStartedID:                  -23,
		DecisionRequestID:                  "emptyUuid",
		DecisionOriginalScheduledTimestamp: 1709872131542474000,
		AutoResetPoints: &types.ResetPoints{
			Points: []*types.ResetPointInfo{{
				BinaryChecksum:           "6df03bf5110d681667852a8456519536",
				RunID:                    "5adce5c5-b7b2-4418-9bf0-4207303f6343",
				FirstDecisionCompletedID: 4,
				CreatedTimeNano:          common.Ptr(int64(1709872121495904000)),
				ExpiringTimeNano:         common.Ptr(int64(ts1)),
				Resettable:               true,
			}},
		},
		SearchAttributes: map[string][]uint8{"BinaryChecksums": {91, 34, 54, 100, 102, 48, 51, 98, 102, 53, 49, 49, 48, 100, 54, 56, 49, 54, 54, 55, 56, 53, 50, 97, 56, 52, 53, 54, 53, 49, 57, 53, 51, 54, 34, 93}},
	}

	expectedEndingBuilderHistoryState := []*types.HistoryEvent{
		{
			ID:        15,
			Timestamp: common.Ptr(int64(1709872131580456000)),
			EventType: common.Ptr(types.EventTypeDecisionTaskCompleted),
			Version:   1,
			TaskID:    -1234,
			DecisionTaskCompletedEventAttributes: &types.DecisionTaskCompletedEventAttributes{
				ScheduledEventID: 13,
				StartedEventID:   14,
				Identity:         "27368@david-porter-DVFG73D710@helloWorldGroup@6027e9ee-048e-4f67-8d88-27883c496901",
				BinaryChecksum:   "6df03bf5110d681667852a8456519536",
			},
		},
		{
			ID:        16,
			Timestamp: common.Ptr(int64(1709872131923473000)),
			EventType: common.Ptr(types.EventTypeWorkflowExecutionContinuedAsNew),
			Version:   1,
			TaskID:    -1234,
			WorkflowExecutionContinuedAsNewEventAttributes: &types.WorkflowExecutionContinuedAsNewEventAttributes{
				NewExecutionRunID: "db647c4b-8759-47f1-8db3-57baff104b76",
				WorkflowType: &types.WorkflowType{
					Name: "helloWorldWorkflow",
				},
				TaskList:                            &types.TaskList{Name: "helloWorldGroup"},
				Input:                               []uint8{110, 117, 108, 108, 10},
				ExecutionStartToCloseTimeoutSeconds: common.Ptr(int32(60)),
				TaskStartToCloseTimeoutSeconds:      common.Ptr(int32(60)),
				DecisionTaskCompletedEventID:        15,
				BackoffStartIntervalInSeconds:       common.Ptr(int32(0)),
				Initiator:                           common.Ptr(types.ContinueAsNewInitiatorDecider),
				Header:                              &types.Header{},
			},
		},
	}

	expectedEndingReturnExecutionState := &persistence.WorkflowExecutionInfo{
		DomainID:                           "5391dbea-5b30-4323-82ca-e1c95339bb3e",
		WorkflowID:                         "helloworld_b4db8bd0-74b7-4250-ade7-ac72a1efb171",
		RunID:                              "db647c4b-8759-47f1-8db3-57baff104b76",
		FirstExecutionRunID:                "5adce5c5-b7b2-4418-9bf0-4207303f6343",
		InitiatedID:                        -23,
		TaskList:                           "helloWorldGroup",
		WorkflowTypeName:                   "helloWorldWorkflow",
		WorkflowTimeout:                    60,
		DecisionStartToCloseTimeout:        60,
		State:                              1,
		LastFirstEventID:                   1,
		NextEventID:                        3,
		LastProcessedEvent:                 -23,
		StartTimestamp:                     time.Date(2024, 3, 7, 20, 28, 51, 923568000, time.Local),
		CreateRequestID:                    "4630bf04-5c64-41bf-92d9-576db2d535cb",
		DecisionVersion:                    1,
		DecisionScheduleID:                 2,
		DecisionStartedID:                  -23,
		DecisionRequestID:                  "emptyUuid",
		DecisionTimeout:                    60,
		DecisionScheduledTimestamp:         ts1,
		DecisionOriginalScheduledTimestamp: ts1,
		AutoResetPoints: &types.ResetPoints{
			Points: []*types.ResetPointInfo{{
				BinaryChecksum:           "6df03bf5110d681667852a8456519536",
				RunID:                    "5adce5c5-b7b2-4418-9bf0-4207303f6343",
				FirstDecisionCompletedID: 4,
				CreatedTimeNano:          common.Ptr(int64(1709872121495904000)),
				ExpiringTimeNano:         common.Ptr(int64(ts1)),
				Resettable:               true,
			}},
		},
	}
	expectedEndingReturnHistoryState := []*types.HistoryEvent{
		{
			ID:        1,
			Timestamp: common.Ptr(int64(1709872131923540000)),
			EventType: common.Ptr(types.EventTypeWorkflowExecutionStarted),
			Version:   1,
			TaskID:    -1234,
			WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &types.WorkflowType{
					Name: "helloWorldWorkflow",
				},
				TaskList:                            &types.TaskList{Name: "helloWorldGroup"},
				Input:                               []uint8{110, 117, 108, 108, 10},
				ExecutionStartToCloseTimeoutSeconds: common.Ptr(int32(60)),
				TaskStartToCloseTimeoutSeconds:      common.Ptr(int32(60)),
				ContinuedExecutionRunID:             "5adce5c5-b7b2-4418-9bf0-4207303f6343",
				OriginalExecutionRunID:              "db647c4b-8759-47f1-8db3-57baff104b76",
				FirstExecutionRunID:                 "5adce5c5-b7b2-4418-9bf0-4207303f6343",
				PrevAutoResetPoints: &types.ResetPoints{Points: []*types.ResetPointInfo{{
					BinaryChecksum:           "6df03bf5110d681667852a8456519536",
					RunID:                    "5adce5c5-b7b2-4418-9bf0-4207303f6343",
					FirstDecisionCompletedID: 4,
					CreatedTimeNano:          common.Ptr(int64(1709872121495904000)),
					ExpiringTimeNano:         common.Ptr(int64(1709937512898751000)),
					Resettable:               true,
				}}},
				Header: &types.Header{},
			},
		},
		{
			ID:        2,
			Timestamp: common.Ptr(int64(ts1)),
			EventType: common.Ptr(types.EventTypeDecisionTaskScheduled),
			Version:   1,
			TaskID:    -1234,
			DecisionTaskScheduledEventAttributes: &types.DecisionTaskScheduledEventAttributes{
				TaskList:                   &types.TaskList{Name: "helloWorldGroup"},
				StartToCloseTimeoutSeconds: common.Ptr(int32(60)),
			},
		},
	}

	fetchedHistoryEvent1 := &types.HistoryEvent{
		ID: 1, Timestamp: common.Ptr(int64(1709938156435726000)),
		EventType: common.Ptr(types.EventTypeWorkflowExecutionStarted),
		Version:   1,
		TaskID:    17826364,
		WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
			WorkflowType:                        &types.WorkflowType{Name: "helloWorldWorkflow"},
			TaskList:                            &types.TaskList{Name: "helloWorldGroup"},
			Input:                               []uint8{110, 117, 108, 108, 10},
			ExecutionStartToCloseTimeoutSeconds: common.Ptr(int32(60)),
			TaskStartToCloseTimeoutSeconds:      common.Ptr(int32(60)),
			ContinuedExecutionRunID:             "96892ca6-975a-44b1-9726-cdb63acd8cda",
			OriginalExecutionRunID:              "befc5b41-fb06-4a99-bec2-91c3e98b17d7",
			FirstExecutionRunID:                 "bcdee7e4-cb21-4bbb-a8d1-43da79e3d252",
			PrevAutoResetPoints: &types.ResetPoints{Points: []*types.ResetPointInfo{
				{
					BinaryChecksum:           "6df03bf5110d681667852a8456519536",
					RunID:                    "bcdee7e4-cb21-4bbb-a8d1-43da79e3d252",
					FirstDecisionCompletedID: 4,
					CreatedTimeNano:          common.Ptr(int64(1709938002170829000)),
					ExpiringTimeNano:         common.Ptr(int64(1710197212347858000)),
					Resettable:               true,
				},
			}},
			Header: &types.Header{},
		},
	}

	tests := map[string]struct {
		startingState *persistence.WorkflowExecutionInfo
		// history is a substruct of current state, but because they're both
		// pointing to each other, they're assembled at the test start
		startingHistory []*types.HistoryEvent

		// expectations
		historyManagerAffordance func(historyManager *persistence.MockHistoryManager)

		// this is a somewhat confusing API, both returning a new cloned state and mutating the existing
		// current state so this will be comparing both
		expectedBuilderEndState *persistence.WorkflowExecutionInfo // this is what the MSB should end up as
		expectedReturnedState   *persistence.WorkflowExecutionInfo // this is returned
		expectedBuilderHistory  []*types.HistoryEvent
		expectedReturnedHistory []*types.HistoryEvent
		expectedErr             error
	}{
		"a continue-as-new event with no errors": {
			startingState:   startingExecutionInfo,
			startingHistory: startingHistory,

			// when it goes to fetch the starting event
			historyManagerAffordance: func(historyManager *persistence.MockHistoryManager) {
				historyManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ReadHistoryBranchResponse{
					HistoryEvents: []*types.HistoryEvent{
						fetchedHistoryEvent1,
					},
				}, nil)
			},

			expectedBuilderEndState: expectedEndingBuilderExecutionInfo,
			expectedBuilderHistory:  expectedEndingBuilderHistoryState,

			expectedReturnedState:   expectedEndingReturnExecutionState,
			expectedReturnedHistory: expectedEndingReturnHistoryState,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {

			ctrl := gomock.NewController(t)

			historyManager := persistence.NewMockHistoryManager(ctrl)
			td.historyManagerAffordance(historyManager)

			domainCache := cache.NewMockDomainCache(ctrl)
			domainCache.EXPECT().GetDomainName(gomock.Any()).Return("domain", nil)

			msb := &mutableStateBuilder{
				domainEntry: cache.NewDomainCacheEntryForTest(
					&persistence.DomainInfo{ID: domainID},
					&persistence.DomainConfig{},
					true,
					&persistence.DomainReplicationConfig{},
					1,
					nil,
				),
				executionInfo: startingExecutionInfo,
				logger:        log.NewNoop(),
				config:        config.NewForTest(),
			}
			msb.timeSource = clock.NewMockedTimeSourceAt(time.Unix(0, ts1))

			msb.eventsCache = events.NewCache(shardID,
				historyManager,
				config.NewForTest(),
				log.NewNoop(),
				metrics.NewNoopMetricsClient(),
				domainCache)

			shardContext := shardCtx.NewMockContext(ctrl)
			shardContext.EXPECT().GetShardID().Return(123)
			shardContext.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(2)
			shardContext.EXPECT().GetEventsCache().Return(msb.eventsCache)
			shardContext.EXPECT().GetConfig().Return(msb.config)
			shardContext.EXPECT().GetTimeSource().Return(msb.timeSource)
			shardContext.EXPECT().GetMetricsClient().Return(metrics.NewNoopMetricsClient())
			shardContext.EXPECT().GetDomainCache().Return(domainCache)

			taskGenerator := NewMockMutableStateTaskGenerator(ctrl)

			taskGenerator.EXPECT().GenerateWorkflowCloseTasks(gomock.Any(), msb.config.WorkflowDeletionJitterRange("domain"))

			msb.shard = shardContext
			msb.executionInfo = td.startingState
			msb.hBuilder = &HistoryBuilder{
				history:   td.startingHistory,
				msBuilder: msb,
			}

			msb.taskGenerator = taskGenerator

			_, returnedBuilder, err := msb.AddContinueAsNewEvent(context.Background(), firstEventID, decisionCompletedEventID, "", &types.ContinueAsNewWorkflowExecutionDecisionAttributes{
				WorkflowType: &types.WorkflowType{
					Name: "helloWorldWorkflow",
				},
				TaskList: &types.TaskList{
					Name: "helloWorldGroup",
				},
				Input: []uint8{110, 117, 108, 108, 10},
			})

			resultExecutionInfo := returnedBuilder.GetExecutionInfo()

			// assert.Equal(t, td.expectedBuilderEndState, msb.executionInfo)
			// assert.Equal(t, td.expectedReturnedState, returnedBuilder.GetExecutionInfo())

			// assert.Equal(t, td.expectedBuilderHistory, msb.hBuilder.history)
			if td.expectedErr != nil {
				assert.ErrorAs(t, err, &td.expectedErr)
			}

			// these are generated nondeterministically, with a plain guid generator
			// todo(david): make this mockable
			td.expectedReturnedState.RunID = "a run id"
			resultExecutionInfo.RunID = "a run id"
			td.expectedReturnedState.CreateRequestID = "a request id"
			resultExecutionInfo.CreateRequestID = "a request id"

			assert.Equal(t, td.expectedReturnedState, resultExecutionInfo)
		})
	}
}

// startExecutionInfo := &persistence.WorkflowExecutionInfo{
// 	DomainID:                           "5391dbea-5b30-4323-82ca-e1c95339bb3e",
// 	WorkflowID:                         "helloworld_bfa410d9-9f49-4bcb-a943-0f3ceb252da2",
// 	RunID:                              "c6702a46-a1f0-42d2-9de7-8aca6ed6795f",
// 	FirstExecutionRunID:                "c6702a46-a1f0-42d2-9de7-8aca6ed6795f",
// 	InitiatedID:                        -7,
// 	CompletionEventBatchID:             0,
// 	TaskList:                           "tasklist-name",
// 	WorkflowTypeName:                   "test-workflow",
// 	WorkflowTimeout:                    60,
// 	DecisionStartToCloseTimeout:        60,
// 	State:                              1,
// 	CloseStatus:                        0,
// 	LastFirstEventID:                   14,
// 	LastEventTaskID:                    11534369,
// 	NextEventID:                        16,
// 	LastProcessedEvent:                 14,
// 	CreateRequestID:                    "5c0be655-1efc-4dfe-8f69-1f59e59c13ef",
// 	DecisionVersion:                    -24,
// 	DecisionScheduleID:                 -23,
// 	DecisionStartedID:                  -23,
// 	DecisionRequestID:                  "emptyUuid",
// 	DecisionTimeout:                    0,
// 	DecisionAttempt:                    0,
// 	DecisionStartedTimestamp:           0,
// 	DecisionScheduledTimestamp:         0,
// 	DecisionOriginalScheduledTimestamp: 1709790036041553000,
// 	CancelRequested:                    false,
// 	StickyTaskList:                     "david-porter-DVFG73D710:04be47fa-2381-469f-b2ea-1253271ad116",
// 	StickyScheduleToStartTimeout:       5,
// 	ClientLibraryVersion:               "0.18.4",
// 	ClientFeatureVersion:               "1.7.0",
// 	ClientImpl:                         "uber-go",
// 	Attempt:                            0,
// }

// endExecutionInfo := &persistence.WorkflowExecutionInfo{
// 	DomainID:                           "5391dbea-5b30-4323-82ca-e1c95339bb3e",
// 	WorkflowID:                         "helloworld_bfa410d9-9f49-4bcb-a943-0f3ceb252da2",
// 	RunID:                              "c6702a46-a1f0-42d2-9de7-8aca6ed6795f",
// 	FirstExecutionRunID:                "c6702a46-a1f0-42d2-9de7-8aca6ed6795f",
// 	InitiatedID:                        -7,
// 	CompletionEventBatchID:             15,
// 	TaskList:                           "tasklist-name",
// 	WorkflowTypeName:                   "test-workflow",
// 	WorkflowTimeout:                    60,
// 	DecisionStartToCloseTimeout:        60,
// 	State:                              2,
// 	CloseStatus:                        5,
// 	LastFirstEventID:                   14,
// 	LastEventTaskID:                    11534369,
// 	NextEventID:                        17,
// 	LastProcessedEvent:                 14,
// 	CreateRequestID:                    "5c0be655-1efc-4dfe-8f69-1f59e59c13ef",
// 	DecisionVersion:                    -24,
// 	DecisionScheduleID:                 -23,
// 	DecisionStartedID:                  -23,
// 	DecisionRequestID:                  "emptyUuid",
// 	DecisionTimeout:                    0,
// 	DecisionAttempt:                    0,
// 	DecisionOriginalScheduledTimestamp: 1709790036041553000,
// }

// startHistory := []*types.HistoryEvent{
// 	{
// 		ID:        15,
// 		Timestamp: common.Ptr(int64(1709791556528026000)),
// 		EventType: types.EventTypeDecisionTaskCompleted.Ptr(),
// 		Version:   1,
// 		TaskID:    -1234,
// 		DecisionTaskCompletedEventAttributes: &types.DecisionTaskCompletedEventAttributes{
// 			ScheduledEventID: 13,
// 			StartedEventID:   14,
// 			Identity:         "27368@david-porter-DVFG73D710@helloWorldGroup@6027e9ee-048e-4f67-8d88-27883c496901",
// 			BinaryChecksum:   "6df03bf5110d681667852a8456519536",
// 		},
// 	},
// }

// endHistory := []*types.HistoryEvent{
// 	{
// 		ID:        15,
// 		Timestamp: common.Ptr(int64(1709791556528026000)),
// 		EventType: types.EventTypeDecisionTaskCompleted.Ptr(),
// 		Version:   1,
// 		TaskID:    -1234,
// 		DecisionTaskCompletedEventAttributes: &types.DecisionTaskCompletedEventAttributes{
// 			ScheduledEventID: 13,
// 			StartedEventID:   14,
// 			Identity:         "27368@david-porter-DVFG73D710@helloWorldGroup@6027e9ee-048e-4f67-8d88-27883c496901",
// 			BinaryChecksum:   "6df03bf5110d681667852a8456519536",
// 		},
// 	},
// 	{
// 		ID:        16,
// 		Timestamp: common.Ptr(int64(1709791556529788000)),
// 		EventType: common.Ptr(types.EventTypeWorkflowExecutionContinuedAsNew),
// 		Version:   1,
// 		TaskID:    -1234,
// 		WorkflowExecutionContinuedAsNewEventAttributes: &types.WorkflowExecutionContinuedAsNewEventAttributes{
// 			NewExecutionRunID: "1b094f71-9c23-4177-8cf9-7f723cc52955",
// 			WorkflowType: &types.WorkflowType{
// 				Name: "helloWorldWorkflow",
// 			},
// 			TaskList: &types.TaskList{
// 				Name: "helloWorldGroup",
// 			},
// 			Input:                               []byte("some-input"),
// 			ExecutionStartToCloseTimeoutSeconds: common.Ptr(int32(60)),
// 			TaskStartToCloseTimeoutSeconds:      common.Ptr(int32(60)),
// 			DecisionTaskCompletedEventID:        15,
// 			Initiator:                           common.Ptr(types.ContinueAsNewInitiatorDecider),
// 		},
// 	},
// }
