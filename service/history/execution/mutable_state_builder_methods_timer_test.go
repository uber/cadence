package execution

import (
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"testing"
	"time"
)

func Test__checkAndClearTimerFiredEvent(t *testing.T) {
	t.Run("no timer fired event to clear", func(t *testing.T) {
		events := []*types.HistoryEvent{{
			ID:        1,
			Timestamp: nil,
			EventType: types.EventTypeActivityTaskScheduled.Ptr(),
		}}
		remainingEvents, timerEvent := checkAndClearTimerFiredEvent(events, "1")
		assert.Nil(t, timerEvent)
		assert.Equal(t, len(events), len(remainingEvents))
	})
	t.Run("timer fired event cleared", func(t *testing.T) {
		timerEvent := &types.HistoryEvent{
			ID:        2,
			Timestamp: nil,
			EventType: types.EventTypeTimerFired.Ptr(),
			TimerFiredEventAttributes: &types.TimerFiredEventAttributes{
				TimerID:        "101",
				StartedEventID: 0,
			},
		}
		events := []*types.HistoryEvent{{
			ID:        1,
			Timestamp: nil,
			EventType: types.EventTypeActivityTaskScheduled.Ptr(),
		},
			timerEvent,
		}
		remainingEvents, clearedEvent := checkAndClearTimerFiredEvent(events, timerEvent.TimerFiredEventAttributes.TimerID)
		assert.NotNil(t, timerEvent)
		assert.Equal(t, timerEvent, clearedEvent)
		assert.Equal(t, len(events)-1, len(remainingEvents))
	})
}

func Test__DeleteUserTimer(t *testing.T) {
	mb := testMutableStateBuilder(t)
	ti := &persistence.TimerInfo{
		TimerID:   "101",
		StartedID: 1,
	}
	mb.pendingTimerInfoIDs[ti.TimerID] = ti
	mb.pendingTimerEventIDToID[ti.StartedID] = ti.TimerID
	err := mb.DeleteUserTimer(ti.TimerID)
	assert.NoError(t, err)
}

func Test__UpdateUserTimer(t *testing.T) {
	mb := testMutableStateBuilder(t)
	ti := &persistence.TimerInfo{
		TimerID:   "101",
		StartedID: 1,
	}
	t.Run("missing timer info", func(t *testing.T) {
		err := mb.UpdateUserTimer(ti)
		assert.Error(t, err)
		assert.Equal(t, ErrMissingTimerInfo, err)

		mb.pendingTimerEventIDToID[ti.StartedID] = ti.TimerID
		err = mb.UpdateUserTimer(ti)
		assert.Error(t, err)
		assert.Equal(t, ErrMissingTimerInfo, err)
	})
	t.Run("success", func(t *testing.T) {
		mb.pendingTimerInfoIDs[ti.TimerID] = ti
		mb.pendingTimerEventIDToID[ti.StartedID] = ti.TimerID
		err := mb.UpdateUserTimer(ti)
		assert.NoError(t, err)
	})
}

func Test__GetUserTimerInfo(t *testing.T) {
	mb := testMutableStateBuilder(t)
	ti := &persistence.TimerInfo{
		TimerID:   "101",
		StartedID: 1,
	}
	mb.pendingTimerInfoIDs[ti.TimerID] = ti
	info, ok := mb.GetUserTimerInfo(ti.TimerID)
	assert.Equal(t, ti, info)
	assert.True(t, ok)
}

func Test__ReplicateTimerCanceledEvent(t *testing.T) {
	mb := testMutableStateBuilder(t)
	timerEvent := &types.HistoryEvent{
		ID:        1,
		Timestamp: nil,
		EventType: types.EventTypeTimerCanceled.Ptr(),
		TimerCanceledEventAttributes: &types.TimerCanceledEventAttributes{
			TimerID:        "101",
			StartedEventID: 0,
		},
	}
	ti := &persistence.TimerInfo{
		TimerID:   "101",
		StartedID: 1,
	}
	mb.pendingTimerInfoIDs[ti.TimerID] = ti
	mb.pendingTimerEventIDToID[ti.StartedID] = ti.TimerID
	err := mb.ReplicateTimerCanceledEvent(timerEvent)
	assert.NoError(t, err)
}

func Test__ReplicateTimerFiredEvent(t *testing.T) {
	mb := testMutableStateBuilder(t)
	timerEvent := &types.HistoryEvent{
		ID:        1,
		Timestamp: nil,
		EventType: types.EventTypeTimerFired.Ptr(),
		TimerFiredEventAttributes: &types.TimerFiredEventAttributes{
			TimerID:        "101",
			StartedEventID: 0,
		},
	}
	ti := &persistence.TimerInfo{
		TimerID:   "101",
		StartedID: 1,
	}
	mb.pendingTimerInfoIDs[ti.TimerID] = ti
	mb.pendingTimerEventIDToID[ti.StartedID] = ti.TimerID
	err := mb.ReplicateTimerFiredEvent(timerEvent)
	assert.NoError(t, err)
}

func Test__ReplicateTimerStartedEvent(t *testing.T) {
	mb := testMutableStateBuilder(t)
	startToFireTimeoutSeconds := int64(5)
	now := time.Now()
	nowUnix := now.UnixNano()
	timerEvent := &types.HistoryEvent{
		ID:        1,
		Version:   0,
		Timestamp: &nowUnix,
		EventType: types.EventTypeTimerStarted.Ptr(),
		TimerStartedEventAttributes: &types.TimerStartedEventAttributes{
			TimerID:                   "101",
			StartToFireTimeoutSeconds: &startToFireTimeoutSeconds,
		},
	}
	expectedTimerInfo := &persistence.TimerInfo{
		Version:    0,
		TimerID:    "101",
		StartedID:  1,
		ExpiryTime: now.Add(time.Second * time.Duration(int64(5))),
	}

	ti, err := mb.ReplicateTimerStartedEvent(timerEvent)
	assert.NoError(t, err)
	assert.Equal(t, expectedTimerInfo.ExpiryTime.UTC(), ti.ExpiryTime.UTC())
	assert.Equal(t, expectedTimerInfo.TimerID, ti.TimerID)
}

func Test__GetPendingTimerInfos(t *testing.T) {
	mb := testMutableStateBuilder(t)
	pendingTimerInfo := map[string]*persistence.TimerInfo{
		"101": {
			Version:   0,
			TimerID:   "101",
			StartedID: 1,
		},
	}
	mb.pendingTimerInfoIDs = pendingTimerInfo
	result := mb.GetPendingTimerInfos()
	assert.Equal(t, pendingTimerInfo, result)
}
