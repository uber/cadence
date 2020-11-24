// Copyright (c) 2017-2020 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uber/cadence/.gen/go/sqlblobs"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/serialization"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
	"github.com/uber/cadence/common/types"
)

func updateActivityInfos(
	ctx context.Context,
	tx sqlplugin.Tx,
	activityInfos []*persistence.InternalActivityInfo,
	deleteInfos []int64,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) error {

	if len(activityInfos) > 0 {
		rows := make([]sqlplugin.ActivityInfoMapsRow, len(activityInfos))
		for i, activityInfo := range activityInfos {
			scheduledEvent, scheduledEncoding := persistence.FromDataBlob(activityInfo.ScheduledEvent)
			startEvent, startEncoding := persistence.FromDataBlob(activityInfo.StartedEvent)

			info := &sqlblobs.ActivityInfo{
				Version:                       &activityInfo.Version,
				ScheduledEventBatchID:         &activityInfo.ScheduledEventBatchID,
				ScheduledEvent:                scheduledEvent,
				ScheduledEventEncoding:        common.StringPtr(scheduledEncoding),
				ScheduledTimeNanos:            common.Int64Ptr(activityInfo.ScheduledTime.UnixNano()),
				StartedID:                     &activityInfo.StartedID,
				StartedEvent:                  startEvent,
				StartedEventEncoding:          common.StringPtr(startEncoding),
				StartedTimeNanos:              common.Int64Ptr(activityInfo.StartedTime.UnixNano()),
				ActivityID:                    &activityInfo.ActivityID,
				RequestID:                     &activityInfo.RequestID,
				ScheduleToStartTimeoutSeconds: common.Int32Ptr(int32(activityInfo.ScheduleToStartTimeout.Seconds())),
				ScheduleToCloseTimeoutSeconds: common.Int32Ptr(int32(activityInfo.ScheduleToCloseTimeout.Seconds())),
				StartToCloseTimeoutSeconds:    common.Int32Ptr(int32(activityInfo.StartToCloseTimeout.Seconds())),
				HeartbeatTimeoutSeconds:       common.Int32Ptr(int32(activityInfo.HeartbeatTimeout.Seconds())),
				CancelRequested:               &activityInfo.CancelRequested,
				CancelRequestID:               &activityInfo.CancelRequestID,
				TimerTaskStatus:               &activityInfo.TimerTaskStatus,
				Attempt:                       &activityInfo.Attempt,
				TaskList:                      &activityInfo.TaskList,
				StartedIdentity:               &activityInfo.StartedIdentity,
				HasRetryPolicy:                &activityInfo.HasRetryPolicy,
				RetryInitialIntervalSeconds:   common.Int32Ptr(int32(activityInfo.InitialInterval.Seconds())),
				RetryBackoffCoefficient:       &activityInfo.BackoffCoefficient,
				RetryMaximumIntervalSeconds:   common.Int32Ptr(int32(activityInfo.MaximumInterval.Seconds())),
				RetryExpirationTimeNanos:      common.Int64Ptr(activityInfo.ExpirationTime.UnixNano()),
				RetryMaximumAttempts:          &activityInfo.MaximumAttempts,
				RetryNonRetryableErrors:       activityInfo.NonRetriableErrors,
				RetryLastFailureReason:        &activityInfo.LastFailureReason,
				RetryLastWorkerIdentity:       &activityInfo.LastWorkerIdentity,
				RetryLastFailureDetails:       activityInfo.LastFailureDetails,
			}
			blob, err := parser.ActivityInfoToBlob(info)
			if err != nil {
				return err
			}
			rows[i] = sqlplugin.ActivityInfoMapsRow{
				ShardID:                  int64(shardID),
				DomainID:                 domainID,
				WorkflowID:               workflowID,
				RunID:                    runID,
				ScheduleID:               activityInfo.ScheduleID,
				LastHeartbeatUpdatedTime: activityInfo.LastHeartBeatUpdatedTime,
				LastHeartbeatDetails:     activityInfo.Details,
				Data:                     blob.Data,
				DataEncoding:             string(blob.Encoding),
			}
		}

		if _, err := tx.ReplaceIntoActivityInfoMaps(ctx, rows); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update activity info. Failed to execute update query. Error: %v", err),
			}
		}
	}

	for _, deleteInfo := range deleteInfos {
		if _, err := tx.DeleteFromActivityInfoMaps(ctx, &sqlplugin.ActivityInfoMapsFilter{
			ShardID:    int64(shardID),
			DomainID:   domainID,
			WorkflowID: workflowID,
			RunID:      runID,
			ScheduleID: &deleteInfo,
		}); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update activity info. Failed to execute delete query. Error: %v", err),
			}
		}
	}

	return nil
}

func getActivityInfoMap(
	ctx context.Context,
	db sqlplugin.DB,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) (map[int64]*persistence.InternalActivityInfo, error) {

	rows, err := db.SelectFromActivityInfoMaps(ctx, &sqlplugin.ActivityInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to get activity info. Error: %v", err),
		}
	}

	ret := make(map[int64]*persistence.InternalActivityInfo)
	for _, row := range rows {
		decoded, err := parser.ActivityInfoFromBlob(row.Data, row.DataEncoding)
		if err != nil {
			return nil, err
		}
		info := &persistence.InternalActivityInfo{
			DomainID:                 row.DomainID.String(),
			ScheduleID:               row.ScheduleID,
			Details:                  row.LastHeartbeatDetails,
			LastHeartBeatUpdatedTime: row.LastHeartbeatUpdatedTime,
			Version:                  decoded.GetVersion(),
			ScheduledEventBatchID:    decoded.GetScheduledEventBatchID(),
			ScheduledEvent:           persistence.NewDataBlob(decoded.ScheduledEvent, common.EncodingType(decoded.GetScheduledEventEncoding())),
			ScheduledTime:            time.Unix(0, decoded.GetScheduledTimeNanos()),
			StartedID:                decoded.GetStartedID(),
			StartedTime:              time.Unix(0, decoded.GetStartedTimeNanos()),
			ActivityID:               decoded.GetActivityID(),
			RequestID:                decoded.GetRequestID(),
			ScheduleToStartTimeout:   common.SecondsToDuration(int64(decoded.GetScheduleToStartTimeoutSeconds())),
			ScheduleToCloseTimeout:   common.SecondsToDuration(int64(decoded.GetScheduleToCloseTimeoutSeconds())),
			StartToCloseTimeout:      common.SecondsToDuration(int64(decoded.GetStartToCloseTimeoutSeconds())),
			HeartbeatTimeout:         common.SecondsToDuration(int64(decoded.GetHeartbeatTimeoutSeconds())),
			CancelRequested:          decoded.GetCancelRequested(),
			CancelRequestID:          decoded.GetCancelRequestID(),
			TimerTaskStatus:          decoded.GetTimerTaskStatus(),
			Attempt:                  decoded.GetAttempt(),
			StartedIdentity:          decoded.GetStartedIdentity(),
			TaskList:                 decoded.GetTaskList(),
			HasRetryPolicy:           decoded.GetHasRetryPolicy(),
			InitialInterval:          common.SecondsToDuration(int64(decoded.GetRetryInitialIntervalSeconds())),
			BackoffCoefficient:       decoded.GetRetryBackoffCoefficient(),
			MaximumInterval:          common.SecondsToDuration(int64(decoded.GetRetryMaximumIntervalSeconds())),
			ExpirationTime:           time.Unix(0, decoded.GetRetryExpirationTimeNanos()),
			MaximumAttempts:          decoded.GetRetryMaximumAttempts(),
			NonRetriableErrors:       decoded.GetRetryNonRetryableErrors(),
			LastFailureReason:        decoded.GetRetryLastFailureReason(),
			LastWorkerIdentity:       decoded.GetRetryLastWorkerIdentity(),
			LastFailureDetails:       decoded.GetRetryLastFailureDetails(),
		}
		if decoded.StartedEvent != nil {
			info.StartedEvent = persistence.NewDataBlob(decoded.StartedEvent, common.EncodingType(decoded.GetStartedEventEncoding()))
		}
		ret[row.ScheduleID] = info
	}

	return ret, nil
}

func deleteActivityInfoMap(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
) error {

	if _, err := tx.DeleteFromActivityInfoMaps(ctx, &sqlplugin.ActivityInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	}); err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to delete activity info map. Error: %v", err),
		}
	}
	return nil
}

func updateTimerInfos(
	ctx context.Context,
	tx sqlplugin.Tx,
	timerInfos []*persistence.TimerInfo,
	deleteInfos []string,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) error {

	if len(timerInfos) > 0 {
		rows := make([]sqlplugin.TimerInfoMapsRow, len(timerInfos))
		for i, timerInfo := range timerInfos {
			blob, err := parser.TimerInfoToBlob(&sqlblobs.TimerInfo{
				Version:         &timerInfo.Version,
				StartedID:       &timerInfo.StartedID,
				ExpiryTimeNanos: common.Int64Ptr(timerInfo.ExpiryTime.UnixNano()),
				// TaskID is a misleading variable, it actually serves
				// the purpose of indicating whether a timer task is
				// generated for this timer info
				TaskID: &timerInfo.TaskStatus,
			})
			if err != nil {
				return err
			}
			rows[i] = sqlplugin.TimerInfoMapsRow{
				ShardID:      int64(shardID),
				DomainID:     domainID,
				WorkflowID:   workflowID,
				RunID:        runID,
				TimerID:      timerInfo.TimerID,
				Data:         blob.Data,
				DataEncoding: string(blob.Encoding),
			}
		}
		if _, err := tx.ReplaceIntoTimerInfoMaps(ctx, rows); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update timer info. Failed to execute update query. Error: %v", err),
			}
		}
	}

	for _, deleteInfo := range deleteInfos {
		if _, err := tx.DeleteFromTimerInfoMaps(ctx, &sqlplugin.TimerInfoMapsFilter{
			ShardID:    int64(shardID),
			DomainID:   domainID,
			WorkflowID: workflowID,
			RunID:      runID,
			TimerID:    &deleteInfo,
		}); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update timer info. Failed to execute delete query. Error: %v", err),
			}
		}
	}

	return nil
}

func getTimerInfoMap(
	ctx context.Context,
	db sqlplugin.DB,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) (map[string]*persistence.TimerInfo, error) {

	rows, err := db.SelectFromTimerInfoMaps(ctx, &sqlplugin.TimerInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to get timer info. Error: %v", err),
		}
	}
	ret := make(map[string]*persistence.TimerInfo)
	for _, row := range rows {
		info, err := parser.TimerInfoFromBlob(row.Data, row.DataEncoding)
		if err != nil {
			return nil, err
		}
		ret[row.TimerID] = &persistence.TimerInfo{
			TimerID:    row.TimerID,
			Version:    info.GetVersion(),
			StartedID:  info.GetStartedID(),
			ExpiryTime: time.Unix(0, info.GetExpiryTimeNanos()),
			// TaskID is a misleading variable, it actually serves
			// the purpose of indicating whether a timer task is
			// generated for this timer info
			TaskStatus: info.GetTaskID(),
		}
	}

	return ret, nil
}

func deleteTimerInfoMap(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
) error {

	if _, err := tx.DeleteFromTimerInfoMaps(ctx, &sqlplugin.TimerInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	}); err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to delete timer info map. Error: %v", err),
		}
	}
	return nil
}

func updateChildExecutionInfos(
	ctx context.Context,
	tx sqlplugin.Tx,
	childExecutionInfos []*persistence.InternalChildExecutionInfo,
	deleteInfos []int64,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) error {

	if len(childExecutionInfos) > 0 {
		rows := make([]sqlplugin.ChildExecutionInfoMapsRow, len(childExecutionInfos))
		for i, childExecutionInfo := range childExecutionInfos {
			initiateEvent, initiateEncoding := persistence.FromDataBlob(childExecutionInfo.InitiatedEvent)
			startEvent, startEncoding := persistence.FromDataBlob(childExecutionInfo.StartedEvent)

			info := &sqlblobs.ChildExecutionInfo{
				Version:                &childExecutionInfo.Version,
				InitiatedEventBatchID:  &childExecutionInfo.InitiatedEventBatchID,
				InitiatedEvent:         initiateEvent,
				InitiatedEventEncoding: &initiateEncoding,
				StartedEvent:           startEvent,
				StartedEventEncoding:   &startEncoding,
				StartedID:              &childExecutionInfo.StartedID,
				StartedWorkflowID:      &childExecutionInfo.StartedWorkflowID,
				StartedRunID:           sqlplugin.MustParseUUID(childExecutionInfo.StartedRunID),
				CreateRequestID:        &childExecutionInfo.CreateRequestID,
				DomainName:             &childExecutionInfo.DomainName,
				WorkflowTypeName:       &childExecutionInfo.WorkflowTypeName,
				ParentClosePolicy:      common.Int32Ptr(int32(childExecutionInfo.ParentClosePolicy)),
			}
			blob, err := parser.ChildExecutionInfoToBlob(info)
			if err != nil {
				return err
			}
			rows[i] = sqlplugin.ChildExecutionInfoMapsRow{
				ShardID:      int64(shardID),
				DomainID:     domainID,
				WorkflowID:   workflowID,
				RunID:        runID,
				InitiatedID:  childExecutionInfo.InitiatedID,
				Data:         blob.Data,
				DataEncoding: string(blob.Encoding),
			}
		}
		if _, err := tx.ReplaceIntoChildExecutionInfoMaps(ctx, rows); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update child execution info. Failed to execute update query. Error: %v", err),
			}
		}
	}

	for _, deleteInfo := range deleteInfos {
		if _, err := tx.DeleteFromChildExecutionInfoMaps(ctx, &sqlplugin.ChildExecutionInfoMapsFilter{
			ShardID:     int64(shardID),
			DomainID:    domainID,
			WorkflowID:  workflowID,
			RunID:       runID,
			InitiatedID: common.Int64Ptr(deleteInfo),
		}); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update child execution info. Failed to execute delete query. Error: %v", err),
			}
		}
	}

	return nil
}

func getChildExecutionInfoMap(
	ctx context.Context,
	db sqlplugin.DB,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) (map[int64]*persistence.InternalChildExecutionInfo, error) {

	rows, err := db.SelectFromChildExecutionInfoMaps(ctx, &sqlplugin.ChildExecutionInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to get timer info. Error: %v", err),
		}
	}

	ret := make(map[int64]*persistence.InternalChildExecutionInfo)
	for _, row := range rows {
		rowInfo, err := parser.ChildExecutionInfoFromBlob(row.Data, row.DataEncoding)
		if err != nil {
			return nil, err
		}
		info := &persistence.InternalChildExecutionInfo{
			InitiatedID:           row.InitiatedID,
			InitiatedEventBatchID: rowInfo.GetInitiatedEventBatchID(),
			Version:               rowInfo.GetVersion(),
			StartedID:             rowInfo.GetStartedID(),
			StartedWorkflowID:     rowInfo.GetStartedWorkflowID(),
			StartedRunID:          sqlplugin.UUID(rowInfo.GetStartedRunID()).String(),
			CreateRequestID:       rowInfo.GetCreateRequestID(),
			DomainName:            rowInfo.GetDomainName(),
			WorkflowTypeName:      rowInfo.GetWorkflowTypeName(),
			ParentClosePolicy:     types.ParentClosePolicy(rowInfo.GetParentClosePolicy()),
		}
		if rowInfo.InitiatedEvent != nil {
			info.InitiatedEvent = persistence.NewDataBlob(rowInfo.InitiatedEvent, common.EncodingType(rowInfo.GetInitiatedEventEncoding()))
		}
		if rowInfo.StartedEvent != nil {
			info.StartedEvent = persistence.NewDataBlob(rowInfo.StartedEvent, common.EncodingType(rowInfo.GetStartedEventEncoding()))
		}
		ret[row.InitiatedID] = info
	}

	return ret, nil
}

func deleteChildExecutionInfoMap(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
) error {

	if _, err := tx.DeleteFromChildExecutionInfoMaps(ctx, &sqlplugin.ChildExecutionInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	}); err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to delete timer info map. Error: %v", err),
		}
	}
	return nil
}

func updateRequestCancelInfos(
	ctx context.Context,
	tx sqlplugin.Tx,
	requestCancelInfos []*persistence.RequestCancelInfo,
	deleteInfos []int64,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) error {

	if len(requestCancelInfos) > 0 {
		rows := make([]sqlplugin.RequestCancelInfoMapsRow, len(requestCancelInfos))
		for i, requestCancelInfo := range requestCancelInfos {
			blob, err := parser.RequestCancelInfoToBlob(&sqlblobs.RequestCancelInfo{
				Version:               &requestCancelInfo.Version,
				InitiatedEventBatchID: &requestCancelInfo.InitiatedEventBatchID,
				CancelRequestID:       &requestCancelInfo.CancelRequestID,
			})
			if err != nil {
				return err
			}
			rows[i] = sqlplugin.RequestCancelInfoMapsRow{
				ShardID:      int64(shardID),
				DomainID:     domainID,
				WorkflowID:   workflowID,
				RunID:        runID,
				InitiatedID:  requestCancelInfo.InitiatedID,
				Data:         blob.Data,
				DataEncoding: string(blob.Encoding),
			}
		}

		if _, err := tx.ReplaceIntoRequestCancelInfoMaps(ctx, rows); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update request cancel info. Failed to execute update query. Error: %v", err),
			}
		}
	}

	for _, deleteInfo := range deleteInfos {
		if _, err := tx.DeleteFromRequestCancelInfoMaps(ctx, &sqlplugin.RequestCancelInfoMapsFilter{
			ShardID:     int64(shardID),
			DomainID:    domainID,
			WorkflowID:  workflowID,
			RunID:       runID,
			InitiatedID: common.Int64Ptr(deleteInfo),
		}); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update request cancel info. Failed to execute delete query. Error: %v", err),
			}
		}
	}

	return nil
}

func getRequestCancelInfoMap(
	ctx context.Context,
	db sqlplugin.DB,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) (map[int64]*persistence.RequestCancelInfo, error) {

	rows, err := db.SelectFromRequestCancelInfoMaps(ctx, &sqlplugin.RequestCancelInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to get request cancel info. Error: %v", err),
		}
	}

	ret := make(map[int64]*persistence.RequestCancelInfo)
	for _, row := range rows {
		rowInfo, err := parser.RequestCancelInfoFromBlob(row.Data, row.DataEncoding)
		if err != nil {
			return nil, err
		}
		ret[row.InitiatedID] = &persistence.RequestCancelInfo{
			Version:               rowInfo.GetVersion(),
			InitiatedID:           row.InitiatedID,
			InitiatedEventBatchID: rowInfo.GetInitiatedEventBatchID(),
			CancelRequestID:       rowInfo.GetCancelRequestID(),
		}
	}

	return ret, nil
}

func deleteRequestCancelInfoMap(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
) error {

	if _, err := tx.DeleteFromRequestCancelInfoMaps(ctx, &sqlplugin.RequestCancelInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	}); err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to delete request cancel info map. Error: %v", err),
		}
	}
	return nil
}

func updateSignalInfos(
	ctx context.Context,
	tx sqlplugin.Tx,
	signalInfos []*persistence.SignalInfo,
	deleteInfos []int64,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) error {

	if len(signalInfos) > 0 {
		rows := make([]sqlplugin.SignalInfoMapsRow, len(signalInfos))
		for i, signalInfo := range signalInfos {
			blob, err := parser.SignalInfoToBlob(&sqlblobs.SignalInfo{
				Version:               &signalInfo.Version,
				InitiatedEventBatchID: &signalInfo.InitiatedEventBatchID,
				RequestID:             &signalInfo.SignalRequestID,
				Name:                  &signalInfo.SignalName,
				Input:                 signalInfo.Input,
				Control:               signalInfo.Control,
			})
			if err != nil {
				return err
			}
			rows[i] = sqlplugin.SignalInfoMapsRow{
				ShardID:      int64(shardID),
				DomainID:     domainID,
				WorkflowID:   workflowID,
				RunID:        runID,
				InitiatedID:  signalInfo.InitiatedID,
				Data:         blob.Data,
				DataEncoding: string(blob.Encoding),
			}
		}

		if _, err := tx.ReplaceIntoSignalInfoMaps(ctx, rows); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update signal info. Failed to execute update query. Error: %v", err),
			}
		}
	}

	for _, deleteInfo := range deleteInfos {
		if _, err := tx.DeleteFromSignalInfoMaps(ctx, &sqlplugin.SignalInfoMapsFilter{
			ShardID:     int64(shardID),
			DomainID:    domainID,
			WorkflowID:  workflowID,
			RunID:       runID,
			InitiatedID: common.Int64Ptr(deleteInfo),
		}); err != nil {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to update signal info. Failed to execute delete query. Error: %v", err),
			}
		}
	}

	return nil
}

func getSignalInfoMap(
	ctx context.Context,
	db sqlplugin.DB,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
	parser serialization.Parser,
) (map[int64]*persistence.SignalInfo, error) {

	rows, err := db.SelectFromSignalInfoMaps(ctx, &sqlplugin.SignalInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to get signal info. Error: %v", err),
		}
	}

	ret := make(map[int64]*persistence.SignalInfo)
	for _, row := range rows {
		rowInfo, err := parser.SignalInfoFromBlob(row.Data, row.DataEncoding)
		if err != nil {
			return nil, err
		}
		ret[row.InitiatedID] = &persistence.SignalInfo{
			Version:               rowInfo.GetVersion(),
			InitiatedID:           row.InitiatedID,
			InitiatedEventBatchID: rowInfo.GetInitiatedEventBatchID(),
			SignalRequestID:       rowInfo.GetRequestID(),
			SignalName:            rowInfo.GetName(),
			Input:                 rowInfo.GetInput(),
			Control:               rowInfo.GetControl(),
		}
	}

	return ret, nil
}

func deleteSignalInfoMap(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID sqlplugin.UUID,
	workflowID string,
	runID sqlplugin.UUID,
) error {

	if _, err := tx.DeleteFromSignalInfoMaps(ctx, &sqlplugin.SignalInfoMapsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	}); err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to delete signal info map. Error: %v", err),
		}
	}
	return nil
}
