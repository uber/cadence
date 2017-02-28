package history

import (
	"fmt"

	"github.com/uber-common/bark"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
)

type (
	mutableStateBuilder struct {
		pendingActivityInfoIDs          map[int64]*persistence.ActivityInfo // Schedule Event ID -> Activity Info.
		pendingActivityInfoByActivityID map[string]int64                    // Activity ID -> Schedule Event ID of the activity.
		updateActivityInfos             []*persistence.ActivityInfo
		deleteActivityInfo              *int64
		pendingTimerInfoIDs             map[string]*persistence.TimerInfo // User Timer ID -> Timer Info.
		updateTimerInfos                []*persistence.TimerInfo
		deleteTimerInfos                []string
		logger                          bark.Logger
	}

	// Represents a decision Info in the mutable state.
	decisionInfo struct {
		ScheduleID          int64
		StartedID           int64
		RequestID           string
		StartToCloseTimeout int32
	}
)

func newMutableStateBuilder(logger bark.Logger) *mutableStateBuilder {
	return &mutableStateBuilder{
		updateActivityInfos:             []*persistence.ActivityInfo{},
		pendingActivityInfoIDs:          make(map[int64]*persistence.ActivityInfo),
		pendingActivityInfoByActivityID: make(map[string]int64),
		pendingTimerInfoIDs:             make(map[string]*persistence.TimerInfo),
		updateTimerInfos:                []*persistence.TimerInfo{},
		deleteTimerInfos:                []string{},
		logger:                          logger}
}

func (e *mutableStateBuilder) Load(
	activityInfos map[int64]*persistence.ActivityInfo,
	timerInfos map[string]*persistence.TimerInfo) {
	e.pendingActivityInfoIDs = activityInfos
	e.pendingTimerInfoIDs = timerInfos
	for _, ai := range activityInfos {
		e.pendingActivityInfoByActivityID[ai.ActivityID] = ai.ScheduleID
	}
}

// GetActivity gives details about an activity that is currently in progress.
func (e *mutableStateBuilder) GetActivity(scheduleEventID int64) (bool, *persistence.ActivityInfo) {
	a, ok := e.pendingActivityInfoIDs[scheduleEventID]
	return ok, a
}

// GetActivityByActivityID gives details about an activity that is currently in progress.
func (e *mutableStateBuilder) GetActivityByActivityID(activityID string) (bool, *persistence.ActivityInfo) {
	eventID, ok := e.pendingActivityInfoByActivityID[activityID]
	if !ok {
		return ok, nil
	}
	a, ok := e.pendingActivityInfoIDs[eventID]
	return ok, a
}

// UpdateActivity updates details about an activity that is in progress.
func (e *mutableStateBuilder) UpdateActivity(scheduleEventID int64, ai *persistence.ActivityInfo) {
	e.pendingActivityInfoIDs[scheduleEventID] = ai
	e.pendingActivityInfoByActivityID[ai.ActivityID] = scheduleEventID
	e.updateActivityInfos = append(e.updateActivityInfos, ai)
}

// DeleteActivity deletes details about an activity.
func (e *mutableStateBuilder) DeleteActivity(scheduleEventID int64) {
	e.deleteActivityInfo = common.Int64Ptr(scheduleEventID)
}

// GetUserTimer gives details about a user timer.
func (e *mutableStateBuilder) GetUserTimer(timerID string) (bool, *persistence.TimerInfo) {
	a, ok := e.pendingTimerInfoIDs[timerID]
	return ok, a
}

// UpdateUserTimer updates the user timer in progress.
func (e *mutableStateBuilder) UpdateUserTimer(timerID string, ti *persistence.TimerInfo) {
	e.pendingTimerInfoIDs[timerID] = ti
	e.updateTimerInfos = append(e.updateTimerInfos, ti)
}

// DeleteUserTimer deletes an user timer.
func (e *mutableStateBuilder) DeleteUserTimer(timerID string) {
	e.deleteTimerInfos = append(e.deleteTimerInfos, timerID)
}

// GetDecision returns details about the in-progress decision task
func (e *mutableStateBuilder) GetDecision(scheduleEventID int64) (bool, *decisionInfo) {
	isRunning, ai := e.GetActivity(scheduleEventID)
	if isRunning {
		di := &decisionInfo{
			ScheduleID:          ai.ScheduleID,
			StartedID:           ai.StartedID,
			RequestID:           ai.RequestID,
			StartToCloseTimeout: ai.StartToCloseTimeout,
		}
		return isRunning, di
	}
	return isRunning, nil
}

// UpdateDecision updates a decision task.
func (e *mutableStateBuilder) UpdateDecision(scheduleEventID int64, di *decisionInfo) {
	decisionTaskName := fmt.Sprintf("DecisionTask-%d", scheduleEventID)
	e.UpdateActivity(scheduleEventID, &persistence.ActivityInfo{
		ScheduleID:          di.ScheduleID,
		StartedID:           di.StartedID,
		RequestID:           di.RequestID,
		StartToCloseTimeout: di.StartToCloseTimeout,
		ActivityID:          decisionTaskName,
	})
}

// DeleteDecision deletes a decision task.
func (e *mutableStateBuilder) DeleteDecision(scheduleEventID int64) {
	e.DeleteActivity(scheduleEventID)
}
