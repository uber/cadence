package execution

import (
	"context"
	"github.com/aws/smithy-go/ptr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/shard"
	"testing"
	"time"
)

func testMutableStateBuilder(t *testing.T) *mutableStateBuilder {
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
	// set the checksum probabilities to 100% for exercising during test
	mockShard.GetConfig().MutableStateChecksumGenProbability = func(domain string) int { return 100 }
	mockShard.GetConfig().MutableStateChecksumVerifyProbability = func(domain string) int { return 100 }
	mockShard.GetConfig().EnableRetryForChecksumFailure = func(domain string) bool { return true }
	logger := log.NewNoop()

	mockShard.Resource.DomainCache.EXPECT().GetDomainID(constants.TestDomainName).Return(constants.TestDomainID, nil).AnyTimes()
	return newMutableStateBuilder(mockShard, logger, constants.TestLocalDomainEntry)
}

func Test__UpdateActivityProgress(t *testing.T) {
	mb := testMutableStateBuilder(t)
	ai := &persistence.ActivityInfo{
		Version:    1,
		ScheduleID: 1,
	}
	request := &types.RecordActivityTaskHeartbeatRequest{
		TaskToken: nil,
		Details:   []byte{10, 0},
		Identity:  "",
	}
	assert.Equal(t, int64(1), ai.Version)
	mb.UpdateActivityProgress(ai, request)
	assert.Equal(t, common.EmptyVersion, ai.Version)
	assert.Equal(t, request.Details, ai.Details)
	assert.Equal(t, ai, mb.updateActivityInfos[ai.ScheduleID])
	assert.NotNil(t, mb.syncActivityTasks[ai.ScheduleID])
}

func Test__ReplicateActivityInfo(t *testing.T) {
	mb := testMutableStateBuilder(t)
	now := time.Now()
	request := &types.SyncActivityRequest{
		ScheduledID:       1,
		Version:           1,
		ScheduledTime:     ptr.Int64(now.UnixNano()),
		LastHeartbeatTime: ptr.Int64(now.UnixNano()),
	}
	ai := &persistence.ActivityInfo{}

	err := mb.ReplicateActivityInfo(request, true)
	assert.Error(t, err)
	assert.Equal(t, ErrMissingActivityInfo, err)

	mb.pendingActivityInfoIDs[request.ScheduledID] = ai
	err = mb.ReplicateActivityInfo(request, true)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), ai.Version)
	assert.Equal(t, now.UTC(), ai.ScheduledTime.UTC())
	assert.Equal(t, request.StartedID, ai.StartedID)
	assert.Equal(t, now.UTC(), ai.LastHeartBeatUpdatedTime.UTC())
}

func Test__UpdateActivity(t *testing.T) {
	mb := testMutableStateBuilder(t)
	ai := &persistence.ActivityInfo{ScheduleID: 1}
	t.Run("error missing activity info", func(t *testing.T) {
		err := mb.UpdateActivity(ai)
		assert.Error(t, err)
		assert.Equal(t, ErrMissingActivityInfo, err)
	})
	t.Run("update success", func(t *testing.T) {
		mb.pendingActivityInfoIDs[1] = ai
		err := mb.UpdateActivity(ai)
		assert.NoError(t, err)
	})
}

func Test__GetActivityScheduledEvent(t *testing.T) {
	mb := testMutableStateBuilder(t)
	ai := &persistence.ActivityInfo{
		ScheduleID:     1,
		ScheduledEvent: &types.HistoryEvent{},
	}
	t.Run("error missing activity info", func(t *testing.T) {
		_, err := mb.GetActivityScheduledEvent(context.Background(), ai.ScheduleID)
		assert.Error(t, err)
		assert.Equal(t, ErrMissingActivityInfo, err)
	})
	t.Run("scheduled event from activity info", func(t *testing.T) {
		mb.pendingActivityInfoIDs[1] = ai
		result, err := mb.GetActivityScheduledEvent(context.Background(), ai.ScheduleID)
		assert.NoError(t, err)
		assert.Equal(t, ai.ScheduledEvent, result)
	})
	t.Run("scheduled event from events cache", func(t *testing.T) {
		mockEventsCache := mb.shard.GetEventsCache().(*events.MockCache)
		mockEventsCache.EXPECT().GetEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&types.HistoryEvent{}, nil)
		mb.pendingActivityInfoIDs[1] = &persistence.ActivityInfo{
			ScheduleID: 1,
		}
		result, err := mb.GetActivityScheduledEvent(context.Background(), ai.ScheduleID)
		assert.NoError(t, err)
		assert.Equal(t, &types.HistoryEvent{}, result)
	})
	t.Run("error missing scheduled event", func(t *testing.T) {
		mockEventsCache := mb.shard.GetEventsCache().(*events.MockCache)
		mockEventsCache.EXPECT().GetEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
		mb.pendingActivityInfoIDs[1] = &persistence.ActivityInfo{
			ScheduleID: 1,
		}
		_, err := mb.GetActivityScheduledEvent(context.Background(), ai.ScheduleID)
		assert.Error(t, err)
		assert.Equal(t, ErrMissingActivityScheduledEvent, err)
	})
}

func Test__AddActivityTaskCompletedEvent(t *testing.T) {
	mb := testMutableStateBuilder(t)
	request := &types.RespondActivityTaskCompletedRequest{
		TaskToken: nil,
		Result:    nil,
		Identity:  "",
	}
	t.Run("error workflow finished", func(t *testing.T) {
		mbCompleted := testMutableStateBuilder(t)
		mbCompleted.executionInfo.State = persistence.WorkflowStateCompleted
		_, err := mbCompleted.AddActivityTaskCompletedEvent(1, 1, request)
		assert.Error(t, err)
		assert.Equal(t, ErrWorkflowFinished, err)
	})
	t.Run("error getting activity info", func(t *testing.T) {
		_, err := mb.AddActivityTaskCompletedEvent(1, 1, request)
		assert.Error(t, err)
		assert.Equal(t, "add-activitytask-completed-event operation failed", err.Error())
	})
	t.Run("success", func(t *testing.T) {
		ai := &persistence.ActivityInfo{
			ScheduleID:     1,
			ActivityID:     "1",
			ScheduledEvent: &types.HistoryEvent{},
			StartedID:      1,
		}
		mb.pendingActivityInfoIDs[1] = ai
		mb.pendingActivityIDToEventID["1"] = 1
		mb.updateActivityInfos[1] = ai
		mb.hBuilder = NewHistoryBuilder(mb)
		_, err := mb.AddActivityTaskCompletedEvent(1, 1, request)
		assert.NoError(t, err)
	})
}
