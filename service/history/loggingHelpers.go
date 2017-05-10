package history

import (
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/logging"
)

func logInvalidHistoryActionEvent(logger bark.Logger, action string, eventID int64, state string) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID:      logging.InvalidHistoryActionEventID,
		logging.TagHistoryBuilderAction: action,
	}).Warnf("Invalid history builder state for action: EventID: %v, State: %v", eventID, state)
}

func logHistorySerializationErrorEvent(logger bark.Logger, err error, msg string) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.HistorySerializationErrorEventID,
		logging.TagWorkflowErr:     err,
	}).Errorf("Error serializing workflow execution history.  Msg: %v", msg)
}

func logHistoryEngineStartingEvent(logger bark.Logger) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.HistoryEngineStarting,
	}).Info("HistoryEngine starting.")
}

func logHistoryEngineStartedEvent(logger bark.Logger) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.HistoryEngineStarted,
	}).Info("HistoryEngine started.")
}

func logHistoryEngineShuttingDownEvent(logger bark.Logger) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.HistoryEngineShuttingDown,
	}).Info("HistoryEngine shutting down.")
}

func logHistoryEngineShutdownEvent(logger bark.Logger) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.HistoryEngineShutdown,
	}).Info("HistoryEngine shutdown.")
}

func logDuplicateTaskEvent(lg bark.Logger, taskType int, taskID int64, requestID string, scheduleID, startedID int64,
	isRunning bool) {
	lg.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.DuplicateTaskEventID,
	}).Debugf("Potentially duplicate task.  TaskID: %v, TaskType: %v, RequestID: %v, scheduleID: %v, startedID: %v, isRunning: %v",
		taskID, taskType, requestID, scheduleID, startedID, isRunning)
}

func logTransferQueueProcesorStartingEvent(logger bark.Logger) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.TransferQueueProcessorStarting,
	}).Info("Transfer queue processor starting.")
}

func logTransferQueueProcesorStartedEvent(logger bark.Logger) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.TransferQueueProcessorStarted,
	}).Info("Transfer queue processor started.")
}

func logTransferQueueProcesorShuttingDownEvent(logger bark.Logger) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.TransferQueueProcessorShuttingDown,
	}).Info("Transfer queue processor shutting down.")
}

func logTransferQueueProcesorShutdownEvent(logger bark.Logger) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.TransferQueueProcessorShutdown,
	}).Info("Transfer queue processor shutdown.")
}

func logTransferQueueProcesorShutdownTimedoutEvent(logger bark.Logger) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.TransferQueueProcessorShutdownTimedout,
	}).Warn("Transfer queue processor timedout on shutdown.")
}

func logShardRangeUpdatedEvent(logger bark.Logger, shardID int, rangeID, startSequence, endSequence int64) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardRangeUpdatedEventID,
	}).Infof("Range updated for shardID '%v'.  RangeID: %v, StartSequence: %v, EndSequence: %v", shardID,
		rangeID, startSequence, endSequence)
}

func logShardControllerStartedEvent(logger bark.Logger, host string) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardControllerStarted,
	}).Infof("ShardController started on host: %v", host)
}

func logShardControllerShutdownEvent(logger bark.Logger, host string) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardControllerShutdown,
	}).Infof("ShardController stopped on host: %v", host)
}

func logShardControllerShuttingDownEvent(logger bark.Logger, host string) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardControllerShuttingDown,
	}).Infof("ShardController stopping on host: %v", host)
}

func logShardControllerShutdownTimedoutEvent(logger bark.Logger, host string) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardControllerShutdownTimedout,
	}).Warnf("ShardController timed out during shutdown on host: %v", host)
}

func logRingMembershipChangedEvent(logger bark.Logger, host string, added, removed, updated int) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.RingMembershipChangedEvent,
	}).Infof("ShardController on host '%v' received ring membership changed event: {Added: %v, Removed: %v, Updated: %v}",
		host, added, removed, updated)
}

func logShardClosedEvent(logger bark.Logger, host string, shardID int) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardClosedEvent,
		logging.TagHistoryShardID:  shardID,
	}).Infof("ShardController on host '%v' received shard closed event for shardID: %v", host, shardID)
}

func logShardItemCreatedEvent(logger bark.Logger, host string, shardID int) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardItemCreated,
	}).Infof("ShardController on host '%v' created a shard item for shardID '%v'.", host, shardID)
}

func logShardItemRemovedEvent(logger bark.Logger, host string, shardID int, remainingShards int) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardItemRemoved,
	}).Infof("ShardController on host '%v' removed shard item for shardID '%v'.  Remaining number of shards: %v",
		host, shardID, remainingShards)
}

func logShardEngineCreatingEvent(logger bark.Logger, host string, shardID int) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardEngineCreating,
	}).Infof("ShardController on host '%v' creating engine for shardID '%v'.", host, shardID)
}

func logShardEngineCreatedEvent(logger bark.Logger, host string, shardID int) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardEngineCreated,
	}).Infof("ShardController on host '%v' created engine for shardID '%v'.", host, shardID)
}

func logShardEngineStoppingEvent(logger bark.Logger, host string, shardID int) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardEngineStopping,
	}).Infof("ShardController on host '%v' stopping engine for shardID '%v'.", host, shardID)
}

func logShardEngineStoppedEvent(logger bark.Logger, host string, shardID int) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.ShardEngineStopped,
	}).Infof("ShardController on host '%v' stopped engine for shardID '%v'.", host, shardID)
}

func logMutableStateInvalidAction(logger bark.Logger, errorMsg string) {
	logger.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.InvalidMutableStateActionEventID,
	}).Errorf("%v.  ", errorMsg)
}

func logMultipleCompletionDecisionsEvent(lg bark.Logger, decisionType shared.DecisionType) {
	lg.WithFields(bark.Fields{
		logging.TagWorkflowEventID: logging.MultipleCompletionDecisionsEventID,
		logging.TagDecisionType:    decisionType,
	}).Warnf("Multiple completion decisions.  DecisionType: %v", decisionType)
}
