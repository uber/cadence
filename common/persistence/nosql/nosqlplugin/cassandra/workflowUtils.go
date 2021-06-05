package cassandra

import (
	"fmt"
	"strings"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
)


func (db *cdb) executeCreateWorkflowBatchTransaction(
	batch gocql.Batch,
	currentWorkflowRequest *nosqlplugin.CurrentWorkflowWriteRequest,
	execution *nosqlplugin.WorkflowExecutionRow,
	shardCondition *nosqlplugin.ShardCondition,
) (*nosqlplugin.ConditionFailureReason, error){
	previous := make(map[string]interface{})
	applied, iter, err := db.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			_ = iter.Close()
		}
	}()

	if err != nil {
		return nil, err
	}

	if !applied {
		// There can be two reasons why the query does not get applied. Either the RangeID has changed, or
		// the workflow is already started. Check the row info returned by Cassandra to figure out which one it is.
	GetFailureReasonLoop:
		for {
			rowType, ok := previous["type"].(int)
			if !ok {
				// This should never happen, as all our rows have the type field.
				break GetFailureReasonLoop
			}
			runID := previous["run_id"].(gocql.UUID).String()

			if rowType == rowTypeShard {
				if rangeID, ok := previous["range_id"].(int64); ok && rangeID != shardCondition.RangeID {
					// CreateWorkflowExecution failed because rangeID was modified
					return &nosqlplugin.ConditionFailureReason{
						ShardRangeIDNotMatch: common.Int64Ptr(rangeID),
					}, errConditionFailed
				}

			} else if rowType == rowTypeExecution && runID == permanentRunID {
				var columns []string
				for k, v := range previous {
					columns = append(columns, fmt.Sprintf("%s=%v", k, v))
				}

				if execution, ok := previous["execution"].(map[string]interface{}); ok {
					// CreateWorkflowExecution failed because it already exists
					executionInfo := createWorkflowExecutionInfo(execution)
					lastWriteVersion := common.EmptyVersion
					if previous["workflow_last_write_version"] != nil {
						lastWriteVersion = previous["workflow_last_write_version"].(int64)
					}

					msg := fmt.Sprintf("Workflow execution already running. WorkflowId: %v, RunId: %v, rangeID: %v, columns: (%v)",
						executionInfo.WorkflowID, executionInfo.RunID, shardCondition.RangeID, strings.Join(columns, ","))

					if currentWorkflowRequest.WriteMode == nosqlplugin.CurrentWorkflowWriteModeInsert {
						return &nosqlplugin.ConditionFailureReason{
							WorkflowExecutionAlreadyExists: &nosqlplugin.WorkflowExecutionAlreadyExists{
								OtherInfo:              msg,
								CreateRequestID:   executionInfo.CreateRequestID,
								RunID:            executionInfo.RunID,
								State:            executionInfo.State,
								CloseStatus:      executionInfo.CloseStatus,
								LastWriteVersion: lastWriteVersion,
							},
						}, errConditionFailed
					}
					return &nosqlplugin.ConditionFailureReason{
						CurrentWorkflowConditionFailInfo: &msg,
					}, errConditionFailed

				}

				if prevRunID := previous["current_run_id"].(gocql.UUID).String(); prevRunID != currentWorkflowRequest.Condition.GetCurrentRunID() {
					// currentRunID on previous run has been changed, return to caller to handle
					msg := fmt.Sprintf("Workflow execution creation condition failed by mismatch runID. WorkflowId: %v, Expected Current RunID: %v, Actual Current RunID: %v",
						execution.WorkflowID, currentWorkflowRequest.Condition.GetCurrentRunID(), prevRunID)
					return &nosqlplugin.ConditionFailureReason{
						CurrentWorkflowConditionFailInfo: &msg,
					}, errConditionFailed
				}

				msg := fmt.Sprintf("Workflow execution creation condition failed. WorkflowId: %v, CurrentRunID: %v, columns: (%v)",
					execution.WorkflowID, execution.RunID, strings.Join(columns, ","))
				return &nosqlplugin.ConditionFailureReason{
					CurrentWorkflowConditionFailInfo: &msg,
				}, errConditionFailed
			} else if rowType == rowTypeExecution && execution.RunID == runID {
				msg := fmt.Sprintf("Workflow execution already running. WorkflowId: %v, RunId: %v, rangeID: %v",
					execution.WorkflowID, execution.RunID, shardCondition.RangeID)
				lastWriteVersion := common.EmptyVersion
				if previous["workflow_last_write_version"] != nil {
					lastWriteVersion = previous["workflow_last_write_version"].(int64)
				}
				return &nosqlplugin.ConditionFailureReason{
					WorkflowExecutionAlreadyExists: &nosqlplugin.WorkflowExecutionAlreadyExists{
						OtherInfo:              msg,
						CreateRequestID:   execution.CreateRequestID,
						RunID:            execution.RunID,
						State:            execution.State,
						CloseStatus:      execution.CloseStatus,
						LastWriteVersion: lastWriteVersion,
					},
				}, errConditionFailed
			}

			previous = make(map[string]interface{})
			if !iter.MapScan(previous) {
				// Cassandra returns the actual row that caused a condition failure, so we should always return
				// from the checks above, but just in case.
				break GetFailureReasonLoop
			}
		}

		return newUnknownConditionFailureReason(shardCondition.RangeID, previous), errConditionFailed
	}

	return nil, nil
}


func newUnknownConditionFailureReason(
	rangeID int64,
	row map[string]interface{},
) *nosqlplugin.ConditionFailureReason {
	// At this point we only know that the write was not applied.
	// It's much safer to return ShardOwnershipLostError as the default to force the application to reload
	// shard to recover from such errors
	var columns []string
	for k, v := range row {
		columns = append(columns, fmt.Sprintf("%s=%v", k, v))
	}

	msg := fmt.Sprintf("Failed to operate on workflow execution.  Request RangeID: %v, columns: (%v)",
		rangeID, strings.Join(columns, ","))

	return &nosqlplugin.ConditionFailureReason{
		UnknownConditionFailureDetails: &msg,
	}
}

func (db *cdb) assertShardRangeID(batch gocql.Batch, shardID int, rangeID int64) error{
	batch.Query(templateUpdateLeaseQuery,
		rangeID,
		shardID,
		rowTypeShard,
		rowTypeShardDomainID,
		rowTypeShardWorkflowID,
		rowTypeShardRunID,
		defaultVisibilityTimestamp,
		rowTypeShardTaskID,
		rangeID,
	)
	return nil
}

func (db *cdb) createTimerTasks(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	transferTasks []*nosqlplugin.TimerTask,
) error {
	for _, task := range transferTasks {
		// Ignoring possible type cast errors.
		ts := persistence.UnixNanoToDBTimestamp(task.VisibilityTimestamp.UnixNano())

		batch.Query(templateCreateTimerTaskQuery,
			shardID,
			rowTypeTimerTask,
			rowTypeTimerDomainID,
			rowTypeTimerWorkflowID,
			rowTypeTimerRunID,
			domainID,
			workflowID,
			runID,
			ts,
			task.TaskID,
			task.Type,
			task.TimeoutType,
			task.EventID,
			task.Attempt,
			task.Version,
			ts,
			task.TaskID)
	}
	return nil
}

func (db *cdb) createReplicationTasks(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	transferTasks []*nosqlplugin.ReplicationTask,
) error {
	for _, task := range transferTasks {
		batch.Query(templateCreateReplicationTaskQuery,
			shardID,
			rowTypeReplicationTask,
			rowTypeReplicationDomainID,
			rowTypeReplicationWorkflowID,
			rowTypeReplicationRunID,
			domainID,
			workflowID,
			runID,
			task.TaskID,
			task.Type,
			task.FirstEventID,
			task.NextEventID,
			task.Version,
			task.ActivityScheduleID,
			persistence.EventStoreVersion,
			task.BranchToken,
			persistence.EventStoreVersion,
			task.NewRunBranchToken,
			task.VisibilityTimestamp.UnixNano(),
			defaultVisibilityTimestamp, // ?? TODO?? should we use task.VisibilityTimestamp?
			task.TaskID)
	}
	return nil
}

func (db *cdb) createTransferTasks(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	transferTasks []*nosqlplugin.TransferTask,
) error {
	for _, task := range transferTasks {
		batch.Query(templateCreateTransferTaskQuery,
			shardID,
			rowTypeTransferTask,
			rowTypeTransferDomainID,
			rowTypeTransferWorkflowID,
			rowTypeTransferRunID,
			domainID,
			workflowID,
			runID,
			task.VisibilityTimestamp,
			task.TaskID,
			task.TargetDomainID,
			task.TargetWorkflowID,
			task.TargetRunID,
			task.TargetChildWorkflowOnly,
			task.TaskList,
			task.Type,
			task.ScheduleID,
			task.RecordVisibility,
			task.Version,
			defaultVisibilityTimestamp, // ?? TODO?? should we use task.VisibilityTimestamp?
			task.TaskID)
	}
	return nil
}

func (db *cdb) createCrossClusterTasks(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	transferTasks []*nosqlplugin.CrossClusterTask,
) error {
	for _, task := range transferTasks {
		batch.Query(templateCreateCrossClusterTaskQuery,
			shardID,
			rowTypeCrossClusterTask,
			rowTypeCrossClusterDomainID,
			task.TargetCluster,
			rowTypeCrossClusterRunID,
			domainID,
			workflowID,
			runID,
			task.VisibilityTimestamp,
			task.TaskID,
			task.TargetDomainID,
			task.TargetWorkflowID,
			task.TargetRunID,
			task.TargetChildWorkflowOnly,
			task.TaskList,
			task.Type,
			task.ScheduleID,
			task.RecordVisibility,
			task.Version,
			defaultVisibilityTimestamp, // ?? TODO?? should we use task.VisibilityTimestamp?
			task.TaskID,
		)
	}
	return nil
}

func (db *cdb) updateSignalsRequested(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	signalReqIDs []string,
	deleteSignalReqIDs []string,
) error {

	if len(signalReqIDs) > 0 {
		batch.Query(templateUpdateSignalRequestedQuery,
			signalReqIDs,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}

	if len(deleteSignalReqIDs) > 0 {
		batch.Query(templateDeleteWorkflowExecutionSignalRequestedQuery,
			deleteSignalReqIDs,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}
	return nil
}

func (db *cdb) updateSignalInfos(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	signalInfos map[int64]*persistence.SignalInfo,
	deleteInfos []int64,
) error {
	for _, c := range signalInfos {
		batch.Query(templateUpdateSignalInfoQuery,
			c.InitiatedID,
			c.Version,
			c.InitiatedID,
			c.InitiatedEventBatchID,
			c.SignalRequestID,
			c.SignalName,
			c.Input,
			c.Control,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}

	// deleteInfos are the initiatedIDs for SignalInfo being deleted
	for _, deleteInfo := range deleteInfos {
		batch.Query(templateDeleteSignalInfoQuery,
			deleteInfo,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}
	return nil
}

func (db *cdb) updateRequestCancelInfos(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	requestCancelInfos map[int64]*persistence.RequestCancelInfo,
	deleteInfos []int64,
) error {

	for _, c := range requestCancelInfos {
		batch.Query(templateUpdateRequestCancelInfoQuery,
			c.InitiatedID,
			c.Version,
			c.InitiatedID,
			c.InitiatedEventBatchID,
			c.CancelRequestID,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}

	// deleteInfos are the initiatedIDs for RequestCancelInfo being deleted
	for _, deleteInfo := range deleteInfos {
		batch.Query(templateDeleteRequestCancelInfoQuery,
			deleteInfo,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}
	return nil
}

func (db *cdb) updateChildExecutionInfos(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	childExecutionInfos map[int64]*persistence.InternalChildExecutionInfo,
	deleteInfos []int64,
) error {

	for _, c := range childExecutionInfos {
		batch.Query(templateUpdateChildExecutionInfoQuery,
			c.InitiatedID,
			c.Version,
			c.InitiatedID,
			c.InitiatedEventBatchID,
			c.InitiatedEvent.Data,
			c.StartedID,
			c.StartedWorkflowID,
			c.StartedRunID,
			c.StartedEvent.Data,
			c.CreateRequestID,
			c.InitiatedEvent.GetEncodingString(),
			c.DomainName,
			c.WorkflowTypeName,
			int32(c.ParentClosePolicy),
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}

	// deleteInfos are the initiatedIDs for ChildInfo being deleted
	for _, deleteInfo := range deleteInfos {
		batch.Query(templateDeleteChildExecutionInfoQuery,
			deleteInfo,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}
	return nil
}

func (db *cdb) updateTimerInfos(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	timerInfos map[string]*persistence.TimerInfo,
	deleteInfos []string,
) error {
	for _, timerInfo := range timerInfos {
		batch.Query(templateUpdateTimerInfoQuery,
			timerInfo.TimerID,
			timerInfo.Version,
			timerInfo.TimerID,
			timerInfo.StartedID,
			timerInfo.ExpiryTime,
			timerInfo.TaskStatus,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}

	for _, deleteInfo := range deleteInfos {
		batch.Query(templateDeleteTimerInfoQuery,
			deleteInfo,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}
	return nil
}

func (db *cdb) updateActivityInfos(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	activityInfos map[int64]*persistence.InternalActivityInfo,
	deleteInfos []int64,
) error {
	for _, a := range activityInfos {
		batch.Query(templateUpdateActivityInfoQuery,
			a.ScheduleID,
			a.Version,
			a.ScheduleID,
			a.ScheduledEventBatchID,
			a.ScheduledEvent.Data,
			a.ScheduledTime,
			a.StartedID,
			a.StartedEvent.Data,
			a.StartedTime,
			a.ActivityID,
			a.RequestID,
			a.Details,
			int32(a.ScheduleToStartTimeout.Seconds()),
			int32(a.ScheduleToCloseTimeout.Seconds()),
			int32(a.StartToCloseTimeout.Seconds()),
			int32(a.HeartbeatTimeout.Seconds()),
			a.CancelRequested,
			a.CancelRequestID,
			a.LastHeartBeatUpdatedTime,
			a.TimerTaskStatus,
			a.Attempt,
			a.TaskList,
			a.StartedIdentity,
			a.HasRetryPolicy,
			int32(a.InitialInterval.Seconds()),
			a.BackoffCoefficient,
			int32(a.MaximumInterval.Seconds()),
			a.ExpirationTime,
			a.MaximumAttempts,
			a.NonRetriableErrors,
			a.LastFailureReason,
			a.LastWorkerIdentity,
			a.LastFailureDetails,
			a.ScheduledEvent.GetEncodingString(),
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}

	for _, deleteInfo := range deleteInfos {
		batch.Query(templateDeleteActivityInfoQuery,
			deleteInfo,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			runID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID)
	}
	return nil
}

// NOTE: not sure why this is needed. We keep the behavior for safe during refactoring
func (db *cdb) convertToCassandraTimestamp(in time.Time) time.Time {
	return time.Unix(0, persistence.DBTimestampToUnixNano(in.UnixNano()))
}

func (db *cdb) createWorkflowExecution(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	runID string,
	execution *nosqlplugin.WorkflowExecutionRow,
) error {
	execution.StartTimestamp = db.convertToCassandraTimestamp(execution.StartTimestamp)
	execution.LastUpdatedTimestamp = db.convertToCassandraTimestamp(execution.LastUpdatedTimestamp)

	batch.Query(templateCreateWorkflowExecutionWithVersionHistoriesQuery,
		shardID,
		domainID,
		workflowID,
		runID,
		rowTypeExecution,
		domainID,
		workflowID,
		runID,
		execution.ParentDomainID,
		execution.ParentWorkflowID,
		execution.ParentRunID,
		execution.InitiatedID,
		execution.CompletionEventBatchID,
		execution.CompletionEvent.Data,
		execution.CompletionEvent.GetEncodingString(),
		execution.TaskList,
		execution.WorkflowTypeName,
		int32(execution.WorkflowTimeout.Seconds()),
		int32(execution.DecisionStartToCloseTimeout.Seconds()),
		execution.ExecutionContext,
		execution.State,
		execution.CloseStatus,
		execution.LastFirstEventID,
		execution.LastEventTaskID,
		execution.NextEventID,
		execution.LastProcessedEvent,
		execution.StartTimestamp,
		execution.LastUpdatedTimestamp,
		execution.CreateRequestID,
		execution.SignalCount,
		execution.HistorySize,
		execution.DecisionVersion,
		execution.DecisionScheduleID,
		execution.DecisionStartedID,
		execution.DecisionRequestID,
		int32(execution.DecisionTimeout.Seconds()),
		execution.DecisionAttempt,
		execution.DecisionStartedTimestamp.UnixNano(),
		execution.DecisionScheduledTimestamp.UnixNano(),
		execution.DecisionOriginalScheduledTimestamp.UnixNano(),
		execution.CancelRequested,
		execution.CancelRequestID,
		execution.StickyTaskList,
		int32(execution.StickyScheduleToStartTimeout.Seconds()),
		execution.ClientLibraryVersion,
		execution.ClientFeatureVersion,
		execution.ClientImpl,
		execution.AutoResetPoints.Data,
		execution.AutoResetPoints.GetEncodingString(),
		execution.Attempt,
		execution.HasRetryPolicy,
		int32(execution.InitialInterval.Seconds()),
		execution.BackoffCoefficient,
		int32(execution.MaximumInterval.Seconds()),
		execution.ExpirationTime,
		execution.MaximumAttempts,
		execution.NonRetriableErrors,
		persistence.EventStoreVersion,
		execution.BranchToken,
		execution.CronSchedule,
		int32(execution.ExpirationSeconds.Seconds()),
		execution.SearchAttributes,
		execution.Memo,
		execution.NextEventID,
		defaultVisibilityTimestamp,
		rowTypeExecutionTaskID,
		execution.VersionHistories.Data,
		execution.VersionHistories.GetEncodingString(),
		execution.Checksums.Version,
		execution.Checksums.Flavor,
		execution.Checksums.Value,
		execution.LastWriteVersion,
		execution.State,
	)
	return nil
}

func (db *cdb) createOrUpdateCurrentWorkflow(
	batch gocql.Batch,
	shardID int,
	domainID string,
	workflowID string,
	request *nosqlplugin.CurrentWorkflowWriteRequest,
) error {
	switch request.WriteMode {
	case nosqlplugin.CurrentWorkflowWriteModeNoop:
		return nil
	case nosqlplugin.CurrentWorkflowWriteModeInsert:
		batch.Query(templateCreateCurrentWorkflowExecutionQuery,
			shardID,
			rowTypeExecution,
			domainID,
			workflowID,
			permanentRunID,
			defaultVisibilityTimestamp,
			rowTypeExecutionTaskID,
			request.Row.RunID,
			request.Row.RunID,
			request.Row.CreateRequestID,
			request.Row.State,
			request.Row.CloseStatus,
			request.Row.LastWriteVersion,
			request.Row.State,
		)
	case nosqlplugin.CurrentWorkflowWriteModeUpdate:
		if request.Condition == nil || request.Condition.GetCurrentRunID() == ""{
			return fmt.Errorf("CurrentWorkflowWriteModeUpdate require Condition.CurrentRunID")
		}
		if request.Condition.LastWriteVersion != nil && request.Condition.State != nil {
			batch.Query(templateUpdateCurrentWorkflowExecutionForNewQuery,
				request.Row.RunID,
				request.Row.RunID,
				request.Row.CreateRequestID,
				request.Row.State,
				request.Row.CloseStatus,
				request.Row.LastWriteVersion,
				request.Row.State,
				shardID,
				rowTypeExecution,
				domainID,
				workflowID,
				permanentRunID,
				defaultVisibilityTimestamp,
				rowTypeExecutionTaskID,
				*request.Condition.CurrentRunID,
				*request.Condition.LastWriteVersion,
				*request.Condition.State,
			)
		} else {
			batch.Query(templateUpdateCurrentWorkflowExecutionQuery,
				request.Row.RunID,
				request.Row.RunID,
				request.Row.CreateRequestID,
				request.Row.State,
				request.Row.CloseStatus,
				request.Row.LastWriteVersion,
				request.Row.State,
				shardID,
				rowTypeExecution,
				domainID,
				workflowID,
				permanentRunID,
				defaultVisibilityTimestamp,
				rowTypeExecutionTaskID,
				*request.Condition.CurrentRunID,
			)
		}
	default:
		return fmt.Errorf("unknown mode %v", request.WriteMode)
	}
	return nil
}

func createWorkflowExecutionInfo(
	result map[string]interface{},
) *persistence.InternalWorkflowExecutionInfo {

	info := &persistence.InternalWorkflowExecutionInfo{}
	var completionEventData []byte
	var completionEventEncoding common.EncodingType
	var autoResetPoints []byte
	var autoResetPointsEncoding common.EncodingType

	for k, v := range result {
		switch k {
		case "domain_id":
			info.DomainID = v.(gocql.UUID).String()
		case "workflow_id":
			info.WorkflowID = v.(string)
		case "run_id":
			info.RunID = v.(gocql.UUID).String()
		case "parent_domain_id":
			info.ParentDomainID = v.(gocql.UUID).String()
			if info.ParentDomainID == emptyDomainID {
				info.ParentDomainID = ""
			}
		case "parent_workflow_id":
			info.ParentWorkflowID = v.(string)
		case "parent_run_id":
			info.ParentRunID = v.(gocql.UUID).String()
			if info.ParentRunID == emptyRunID {
				info.ParentRunID = ""
			}
		case "initiated_id":
			info.InitiatedID = v.(int64)
		case "completion_event_batch_id":
			info.CompletionEventBatchID = v.(int64)
		case "completion_event":
			completionEventData = v.([]byte)
		case "completion_event_data_encoding":
			completionEventEncoding = common.EncodingType(v.(string))
		case "auto_reset_points":
			autoResetPoints = v.([]byte)
		case "auto_reset_points_encoding":
			autoResetPointsEncoding = common.EncodingType(v.(string))
		case "task_list":
			info.TaskList = v.(string)
		case "workflow_type_name":
			info.WorkflowTypeName = v.(string)
		case "workflow_timeout":
			info.WorkflowTimeout = common.SecondsToDuration(int64(v.(int)))
		case "decision_task_timeout":
			info.DecisionStartToCloseTimeout = common.SecondsToDuration(int64(v.(int)))
		case "execution_context":
			info.ExecutionContext = v.([]byte)
		case "state":
			info.State = v.(int)
		case "close_status":
			info.CloseStatus = v.(int)
		case "last_first_event_id":
			info.LastFirstEventID = v.(int64)
		case "last_event_task_id":
			info.LastEventTaskID = v.(int64)
		case "next_event_id":
			info.NextEventID = v.(int64)
		case "last_processed_event":
			info.LastProcessedEvent = v.(int64)
		case "start_time":
			info.StartTimestamp = v.(time.Time)
		case "last_updated_time":
			info.LastUpdatedTimestamp = v.(time.Time)
		case "create_request_id":
			info.CreateRequestID = v.(gocql.UUID).String()
		case "signal_count":
			info.SignalCount = int32(v.(int))
		case "history_size":
			info.HistorySize = v.(int64)
		case "decision_version":
			info.DecisionVersion = v.(int64)
		case "decision_schedule_id":
			info.DecisionScheduleID = v.(int64)
		case "decision_started_id":
			info.DecisionStartedID = v.(int64)
		case "decision_request_id":
			info.DecisionRequestID = v.(string)
		case "decision_timeout":
			info.DecisionTimeout = common.SecondsToDuration(int64(v.(int)))
		case "decision_attempt":
			info.DecisionAttempt = v.(int64)
		case "decision_timestamp":
			info.DecisionStartedTimestamp = time.Unix(0, v.(int64))
		case "decision_scheduled_timestamp":
			info.DecisionScheduledTimestamp = time.Unix(0, v.(int64))
		case "decision_original_scheduled_timestamp":
			info.DecisionOriginalScheduledTimestamp = time.Unix(0, v.(int64))
		case "cancel_requested":
			info.CancelRequested = v.(bool)
		case "cancel_request_id":
			info.CancelRequestID = v.(string)
		case "sticky_task_list":
			info.StickyTaskList = v.(string)
		case "sticky_schedule_to_start_timeout":
			info.StickyScheduleToStartTimeout = common.SecondsToDuration(int64(v.(int)))
		case "client_library_version":
			info.ClientLibraryVersion = v.(string)
		case "client_feature_version":
			info.ClientFeatureVersion = v.(string)
		case "client_impl":
			info.ClientImpl = v.(string)
		case "attempt":
			info.Attempt = int32(v.(int))
		case "has_retry_policy":
			info.HasRetryPolicy = v.(bool)
		case "init_interval":
			info.InitialInterval = common.SecondsToDuration(int64(v.(int)))
		case "backoff_coefficient":
			info.BackoffCoefficient = v.(float64)
		case "max_interval":
			info.MaximumInterval = common.SecondsToDuration(int64(v.(int)))
		case "max_attempts":
			info.MaximumAttempts = int32(v.(int))
		case "expiration_time":
			info.ExpirationTime = v.(time.Time)
		case "non_retriable_errors":
			info.NonRetriableErrors = v.([]string)
		case "branch_token":
			info.BranchToken = v.([]byte)
		case "cron_schedule":
			info.CronSchedule = v.(string)
		case "expiration_seconds":
			info.ExpirationSeconds = common.SecondsToDuration(int64(v.(int)))
		case "search_attributes":
			info.SearchAttributes = v.(map[string][]byte)
		case "memo":
			info.Memo = v.(map[string][]byte)
		}
	}
	info.CompletionEvent = persistence.NewDataBlob(completionEventData, completionEventEncoding)
	info.AutoResetPoints = persistence.NewDataBlob(autoResetPoints, autoResetPointsEncoding)
	return info
}
