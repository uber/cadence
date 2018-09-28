// Copyright (c) 2017 Uber Technologies, Inc.
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

package persistence

import (
	"encoding/json"
	"fmt"

	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/logging"
)

type (

	// executionManagerImpl implements ExecutionManager based on PersistenceExecutionManager, statsComputer and HistorySerializer
	executionManagerImpl struct {
		serializer    HistorySerializer
		persistence   PersistenceExecutionManager
		statsComputer statsComputer
	}

	// historyManagerImpl implements HistoryManager based on PersistenceHistoryManager and HistorySerializer
	historyManagerImpl struct {
		serializer  HistorySerializer
		persistence PersistenceHistoryManager
		logger      bark.Logger
	}

	historyToken struct {
		LastEventBatchVersion int64
		LastEventID           int64
		Data                  []byte
	}
)

var _ ExecutionManager = (*executionManagerImpl)(nil)
var _ HistoryManager = (*historyManagerImpl)(nil)

// NewExecutionManagerImpl returns new ExecutionManager
func NewExecutionManagerImpl(persistence PersistenceExecutionManager) ExecutionManager {
	return &executionManagerImpl{
		serializer:    NewHistorySerializer(),
		persistence:   persistence,
		statsComputer: statsComputer{},
	}
}

//The below three APIs are related to serialization/deserialization
func (m *executionManagerImpl) GetWorkflowExecution(request *GetWorkflowExecutionRequest) (*GetWorkflowExecutionResponse, error) {
	response, err := m.persistence.GetWorkflowExecution(request)
	if err != nil {
		return nil, err
	}
	newResponse := &GetWorkflowExecutionResponse{
		State: &WorkflowMutableState{
			TimerInfos:         response.State.TimerInfos,
			RequestCancelInfos: response.State.RequestCancelInfos,
			SignalInfos:        response.State.SignalInfos,
			SignalRequestedIDs: response.State.SignalRequestedIDs,
			ReplicationState:   response.State.ReplicationState,
		},
	}

	newResponse.State.ActivitInfos, err = m.DeserializeActivityInfos(response.State.ActivitInfos)
	if err != nil {
		return nil, err
	}
	newResponse.State.ChildExecutionInfos, err = m.DeserializeChildExecutionInfos(response.State.ChildExecutionInfos)
	if err != nil {
		return nil, err
	}
	newResponse.State.BufferedEvents, err = m.DeserializeBufferedEvents(response.State.BufferedEvents)
	if err != nil {
		return nil, err
	}
	newResponse.State.BufferedReplicationTasks, err = m.DeserializeBufferedReplicationTasks(response.State.BufferedReplicationTasks)
	if err != nil {
		return nil, err
	}
	newResponse.State.ExecutionInfo, err = m.DeserializeExecutionInfo(response.State.ExecutionInfo)
	if err != nil {
		return nil, err
	}

	return newResponse, nil
}

func (m *executionManagerImpl) DeserializeExecutionInfo(info *PersistenceWorkflowExecutionInfo) (*WorkflowExecutionInfo, error) {
	completionEvent, err := m.serializer.DeserializeEvent(info.CompletionEvent)
	if err != nil {
		return nil, err
	}
	newInfo := &WorkflowExecutionInfo{
		CompletionEvent: completionEvent,

		DomainID:                     info.DomainID,
		WorkflowID:                   info.WorkflowID,
		RunID:                        info.RunID,
		ParentDomainID:               info.ParentDomainID,
		ParentWorkflowID:             info.ParentWorkflowID,
		ParentRunID:                  info.ParentRunID,
		InitiatedID:                  info.InitiatedID,
		TaskList:                     info.TaskList,
		WorkflowTypeName:             info.WorkflowTypeName,
		WorkflowTimeout:              info.WorkflowTimeout,
		DecisionTimeoutValue:         info.DecisionTimeoutValue,
		ExecutionContext:             info.ExecutionContext,
		State:                        info.State,
		CloseStatus:                  info.CloseStatus,
		LastFirstEventID:             info.LastFirstEventID,
		NextEventID:                  info.NextEventID,
		LastProcessedEvent:           info.LastProcessedEvent,
		StartTimestamp:               info.StartTimestamp,
		LastUpdatedTimestamp:         info.LastUpdatedTimestamp,
		CreateRequestID:              info.CreateRequestID,
		HistorySize:                  info.HistorySize,
		DecisionVersion:              info.DecisionVersion,
		DecisionScheduleID:           info.DecisionScheduleID,
		DecisionStartedID:            info.DecisionStartedID,
		DecisionRequestID:            info.DecisionRequestID,
		DecisionTimeout:              info.DecisionTimeout,
		DecisionAttempt:              info.DecisionAttempt,
		DecisionTimestamp:            info.DecisionTimestamp,
		CancelRequested:              info.CancelRequested,
		CancelRequestID:              info.CancelRequestID,
		StickyTaskList:               info.StickyTaskList,
		StickyScheduleToStartTimeout: info.StickyScheduleToStartTimeout,
		ClientLibraryVersion:         info.ClientLibraryVersion,
		ClientFeatureVersion:         info.ClientFeatureVersion,
		ClientImpl:                   info.ClientImpl,
		Attempt:                      info.Attempt,
		HasRetryPolicy:               info.HasRetryPolicy,
		InitialInterval:              info.InitialInterval,
		BackoffCoefficient:           info.BackoffCoefficient,
		MaximumInterval:              info.MaximumInterval,
		ExpirationTime:               info.ExpirationTime,
		MaximumAttempts:              info.MaximumAttempts,
		NonRetriableErrors:           info.NonRetriableErrors,
	}
	return newInfo, nil
}

func (m *executionManagerImpl) DeserializeBufferedReplicationTasks(tasks map[int64]*PersistenceBufferedReplicationTask) (map[int64]*BufferedReplicationTask, error) {
	newBRTs := make(map[int64]*BufferedReplicationTask, 0)
	for k, v := range tasks {
		history, err := m.serializer.DeserializeBatchEvents(v.History)
		if err != nil {
			return nil, err
		}
		newHistory, err := m.serializer.DeserializeBatchEvents(v.NewRunHistory)
		if err != nil {
			return nil, err
		}
		b := &BufferedReplicationTask{
			FirstEventID: v.FirstEventID,
			NextEventID:  v.NextEventID,
			Version:      v.Version,

			History:       history.Events,
			NewRunHistory: newHistory.Events,
		}
		newBRTs[k] = b
	}
	return newBRTs, nil
}

func (m *executionManagerImpl) DeserializeBufferedEvents(blobs []*DataBlob) ([]*workflow.HistoryEvent, error) {
	events := make([]*workflow.HistoryEvent, 0)
	for _, b := range blobs {
		history, err := m.serializer.DeserializeBatchEvents(b)
		if err != nil {
			return nil, err
		}
		events = append(events, history.Events...)
	}
	return events, nil
}

func (m *executionManagerImpl) DeserializeChildExecutionInfos(infos map[int64]*PersistenceChildExecutionInfo) (map[int64]*ChildExecutionInfo, error) {
	newInfos := make(map[int64]*ChildExecutionInfo, 0)
	for k, v := range infos {
		initiatedEvent, err := m.serializer.DeserializeEvent(v.InitiatedEvent)
		if err != nil {
			return nil, err
		}
		startedEvent, err := m.serializer.DeserializeEvent(v.StartedEvent)
		if err != nil {
			return nil, err
		}
		c := &ChildExecutionInfo{
			InitiatedEvent: initiatedEvent,
			StartedEvent:   startedEvent,

			Version:         v.Version,
			InitiatedID:     v.InitiatedID,
			StartedID:       v.StartedID,
			CreateRequestID: v.CreateRequestID,
		}
		newInfos[k] = c
	}
	return newInfos, nil
}

func (m *executionManagerImpl) DeserializeActivityInfos(infos map[int64]*PersistenceActivityInfo) (map[int64]*ActivityInfo, error) {
	newInfos := make(map[int64]*ActivityInfo, 0)
	for k, v := range infos {
		scheduledEvent, err := m.serializer.DeserializeEvent(v.ScheduledEvent)
		if err != nil {
			return nil, err
		}
		startedEvent, err := m.serializer.DeserializeEvent(v.StartedEvent)
		if err != nil {
			return nil, err
		}
		a := &ActivityInfo{
			ScheduledEvent: scheduledEvent,
			StartedEvent:   startedEvent,

			Version:                  v.Version,
			ScheduleID:               v.ScheduleID,
			ScheduledTime:            v.ScheduledTime,
			StartedID:                v.StartedID,
			StartedTime:              v.StartedTime,
			ActivityID:               v.ActivityID,
			RequestID:                v.RequestID,
			Details:                  v.Details,
			ScheduleToStartTimeout:   v.ScheduleToStartTimeout,
			ScheduleToCloseTimeout:   v.ScheduleToCloseTimeout,
			StartToCloseTimeout:      v.StartToCloseTimeout,
			HeartbeatTimeout:         v.HeartbeatTimeout,
			CancelRequested:          v.CancelRequested,
			CancelRequestID:          v.CancelRequestID,
			LastHeartBeatUpdatedTime: v.LastHeartBeatUpdatedTime,
			TimerTaskStatus:          v.TimerTaskStatus,
			Attempt:                  v.Attempt,
			DomainID:                 v.DomainID,
			StartedIdentity:          v.StartedIdentity,
			TaskList:                 v.TaskList,
			HasRetryPolicy:           v.HasRetryPolicy,
			InitialInterval:          v.InitialInterval,
			BackoffCoefficient:       v.BackoffCoefficient,
			MaximumInterval:          v.MaximumInterval,
			ExpirationTime:           v.ExpirationTime,
			MaximumAttempts:          v.MaximumAttempts,
			NonRetriableErrors:       v.NonRetriableErrors,
			LastTimeoutVisibility:    v.LastTimeoutVisibility,
		}
		newInfos[k] = a
	}
	return newInfos, nil
}

func (m *executionManagerImpl) UpdateWorkflowExecution(request *UpdateWorkflowExecutionRequest) (*UpdateWorkflowExecutionResponse, error) {
	executionInfo, err := m.SerializeExecutionInfo(request.ExecutionInfo, request.Encoding)
	if err != nil {
		return nil, err
	}
	upsertActivityInfos, err := m.SerializeUpsertActivityInfos(request.UpsertActivityInfos, request.Encoding)
	if err != nil {
		return nil, err
	}
	upsertChildExecutionInfos, err := m.SerializeUpsertChildExecutionInfos(request.UpsertChildExecutionInfos, request.Encoding)
	if err != nil {
		return nil, err
	}
	newBufferedEvents, err := m.serializer.SerializeBatchEvents(&workflow.History{Events: request.NewBufferedEvents}, request.Encoding)
	if err != nil {
		return nil, err
	}
	newBufferedReplicationTask, err := m.SerializeNewBufferedReplicationTask(request.NewBufferedReplicationTask, request.Encoding)
	if err != nil {
		return nil, err
	}

	newRequest := &PersistenceUpdateWorkflowExecutionRequest{
		ExecutionInfo:              executionInfo,
		UpsertActivityInfos:        upsertActivityInfos,
		UpsertChildExecutionInfos:  upsertChildExecutionInfos,
		NewBufferedEvents:          newBufferedEvents,
		NewBufferedReplicationTask: newBufferedReplicationTask,

		ReplicationState:              request.ReplicationState,
		TransferTasks:                 request.TransferTasks,
		TimerTasks:                    request.TimerTasks,
		ReplicationTasks:              request.ReplicationTasks,
		DeleteTimerTask:               request.DeleteTimerTask,
		Condition:                     request.Condition,
		RangeID:                       request.RangeID,
		ContinueAsNew:                 request.ContinueAsNew,
		FinishExecution:               request.FinishExecution,
		FinishedExecutionTTL:          request.FinishedExecutionTTL,
		DeleteActivityInfos:           request.DeleteActivityInfos,
		UpserTimerInfos:               request.UpserTimerInfos,
		DeleteTimerInfos:              request.DeleteTimerInfos,
		DeleteChildExecutionInfo:      request.DeleteChildExecutionInfo,
		UpsertRequestCancelInfos:      request.UpsertRequestCancelInfos,
		DeleteRequestCancelInfo:       request.DeleteRequestCancelInfo,
		UpsertSignalInfos:             request.UpsertSignalInfos,
		DeleteSignalInfo:              request.DeleteSignalInfo,
		UpsertSignalRequestedIDs:      request.UpsertSignalRequestedIDs,
		DeleteSignalRequestedID:       request.DeleteSignalRequestedID,
		ClearBufferedEvents:           request.ClearBufferedEvents,
		DeleteBufferedReplicationTask: request.DeleteBufferedReplicationTask,
	}
	mss := m.statsComputer.computeMutableStateStats(newRequest)
	msuss := m.statsComputer.computeMutableStateUpdateStats(newRequest)
	return &UpdateWorkflowExecutionResponse{MutableStateStats: mss, MutableStateUpdateSessionStats: msuss}, m.persistence.UpdateWorkflowExecution(newRequest)
}

func (m *executionManagerImpl) SerializeNewBufferedReplicationTask(task *BufferedReplicationTask, encoding common.EncodingType) (*PersistenceBufferedReplicationTask, error) {
	if task == nil {
		return &PersistenceBufferedReplicationTask{
			History:       &DataBlob{},
			NewRunHistory: &DataBlob{},
		}, nil
	}
	var history, newHistory *DataBlob
	var err error
	if task.History != nil {
		history, err = m.serializer.SerializeBatchEvents(&workflow.History{Events: task.History}, encoding)
		if err != nil {
			return nil, err
		}
	}

	if task.NewRunHistory != nil {
		newHistory, err = m.serializer.SerializeBatchEvents(&workflow.History{Events: task.NewRunHistory}, encoding)
		if err != nil {
			return nil, err
		}
	}

	return &PersistenceBufferedReplicationTask{
		FirstEventID: task.FirstEventID,
		NextEventID:  task.NextEventID,
		Version:      task.Version,

		History:       history,
		NewRunHistory: newHistory,
	}, nil
}

func (m *executionManagerImpl) SerializeUpsertChildExecutionInfos(infos []*ChildExecutionInfo, encoding common.EncodingType) ([]*PersistenceChildExecutionInfo, error) {
	newInfos := make([]*PersistenceChildExecutionInfo, 0)
	for _, v := range infos {
		initiatedEvent, err := m.serializer.SerializeEvent(v.InitiatedEvent, encoding)
		if err != nil {
			return nil, err
		}
		startedEvent, err := m.serializer.SerializeEvent(v.StartedEvent, encoding)
		if err != nil {
			return nil, err
		}
		i := &PersistenceChildExecutionInfo{
			InitiatedEvent: initiatedEvent,
			StartedEvent:   startedEvent,

			Version:         v.Version,
			InitiatedID:     v.InitiatedID,
			CreateRequestID: v.CreateRequestID,
			StartedID:       v.StartedID,
		}
		newInfos = append(newInfos, i)
	}
	return newInfos, nil
}

func (m *executionManagerImpl) SerializeUpsertActivityInfos(infos []*ActivityInfo, encoding common.EncodingType) ([]*PersistenceActivityInfo, error) {
	newInfos := make([]*PersistenceActivityInfo, 0)
	for _, v := range infos {
		fmt.Printf("longer: %+v\n", v)
		scheduledEvent, err := m.serializer.SerializeEvent(v.ScheduledEvent, encoding)
		if err != nil {
			return nil, err
		}
		startedEvent, err := m.serializer.SerializeEvent(v.StartedEvent, encoding)
		if err != nil {
			return nil, err
		}
		i := &PersistenceActivityInfo{
			Version:                  v.Version,
			ScheduleID:               v.ScheduleID,
			ScheduledEvent:           scheduledEvent,
			ScheduledTime:            v.ScheduledTime,
			StartedID:                v.StartedID,
			StartedEvent:             startedEvent,
			StartedTime:              v.StartedTime,
			ActivityID:               v.ActivityID,
			RequestID:                v.RequestID,
			Details:                  v.Details,
			ScheduleToStartTimeout:   v.ScheduleToStartTimeout,
			ScheduleToCloseTimeout:   v.ScheduleToCloseTimeout,
			StartToCloseTimeout:      v.StartToCloseTimeout,
			HeartbeatTimeout:         v.HeartbeatTimeout,
			CancelRequested:          v.CancelRequested,
			CancelRequestID:          v.CancelRequestID,
			LastHeartBeatUpdatedTime: v.LastHeartBeatUpdatedTime,
			TimerTaskStatus:          v.TimerTaskStatus,
			Attempt:                  v.Attempt,
			DomainID:                 v.DomainID,
			StartedIdentity:          v.StartedIdentity,
			TaskList:                 v.TaskList,
			HasRetryPolicy:           v.HasRetryPolicy,
			InitialInterval:          v.InitialInterval,
			BackoffCoefficient:       v.BackoffCoefficient,
			MaximumInterval:          v.MaximumInterval,
			ExpirationTime:           v.ExpirationTime,
			MaximumAttempts:          v.MaximumAttempts,
			NonRetriableErrors:       v.NonRetriableErrors,
			LastTimeoutVisibility:    v.LastTimeoutVisibility,
		}
		fmt.Printf("longer: %+v\n", *i)
		newInfos = append(newInfos, i)
	}
	return newInfos, nil
}

func (m *executionManagerImpl) SerializeExecutionInfo(info *WorkflowExecutionInfo, encoding common.EncodingType) (*PersistenceWorkflowExecutionInfo, error) {
	if info == nil {
		return &PersistenceWorkflowExecutionInfo{
			CompletionEvent: &DataBlob{},
		}, nil
	}
	var completionEvent *DataBlob
	var err error
	completionEvent, err = m.serializer.SerializeEvent(info.CompletionEvent, encoding)
	if err != nil {
		return nil, err
	}

	return &PersistenceWorkflowExecutionInfo{
		DomainID:                     info.DomainID,
		WorkflowID:                   info.WorkflowID,
		RunID:                        info.RunID,
		ParentDomainID:               info.ParentDomainID,
		ParentWorkflowID:             info.ParentWorkflowID,
		ParentRunID:                  info.ParentRunID,
		InitiatedID:                  info.InitiatedID,
		CompletionEvent:              completionEvent,
		TaskList:                     info.TaskList,
		WorkflowTypeName:             info.WorkflowTypeName,
		WorkflowTimeout:              info.WorkflowTimeout,
		DecisionTimeoutValue:         info.DecisionTimeoutValue,
		ExecutionContext:             info.ExecutionContext,
		State:                        info.State,
		CloseStatus:                  info.CloseStatus,
		LastFirstEventID:             info.LastFirstEventID,
		NextEventID:                  info.NextEventID,
		LastProcessedEvent:           info.LastProcessedEvent,
		StartTimestamp:               info.StartTimestamp,
		LastUpdatedTimestamp:         info.LastUpdatedTimestamp,
		CreateRequestID:              info.CreateRequestID,
		HistorySize:                  info.HistorySize,
		DecisionVersion:              info.DecisionVersion,
		DecisionScheduleID:           info.DecisionScheduleID,
		DecisionStartedID:            info.DecisionStartedID,
		DecisionRequestID:            info.DecisionRequestID,
		DecisionTimeout:              info.DecisionTimeout,
		DecisionAttempt:              info.DecisionAttempt,
		DecisionTimestamp:            info.DecisionTimestamp,
		CancelRequested:              info.CancelRequested,
		CancelRequestID:              info.CancelRequestID,
		StickyTaskList:               info.StickyTaskList,
		StickyScheduleToStartTimeout: info.StickyScheduleToStartTimeout,
		ClientLibraryVersion:         info.ClientLibraryVersion,
		ClientFeatureVersion:         info.ClientFeatureVersion,
		ClientImpl:                   info.ClientImpl,
		Attempt:                      info.Attempt,
		HasRetryPolicy:               info.HasRetryPolicy,
		InitialInterval:              info.InitialInterval,
		BackoffCoefficient:           info.BackoffCoefficient,
		MaximumInterval:              info.MaximumInterval,
		ExpirationTime:               info.ExpirationTime,
		MaximumAttempts:              info.MaximumAttempts,
		NonRetriableErrors:           info.NonRetriableErrors,
	}, nil
}

func (m *executionManagerImpl) ResetMutableState(request *ResetMutableStateRequest) error {
	executionInfo, err := m.SerializeExecutionInfo(request.ExecutionInfo, request.Encoding)
	if err != nil {
		return err
	}
	insertActivityInfos, err := m.SerializeUpsertActivityInfos(request.InsertActivityInfos, request.Encoding)
	if err != nil {
		return err
	}
	insertChildExecutionInfos, err := m.SerializeUpsertChildExecutionInfos(request.InsertChildExecutionInfos, request.Encoding)
	if err != nil {
		return err
	}

	newRequest := &PersistenceResetMutableStateRequest{
		PrevRunID:                 request.PrevRunID,
		ExecutionInfo:             executionInfo,
		ReplicationState:          request.ReplicationState,
		Condition:                 request.Condition,
		RangeID:                   request.RangeID,
		InsertActivityInfos:       insertActivityInfos,
		InsertTimerInfos:          request.InsertTimerInfos,
		InsertChildExecutionInfos: insertChildExecutionInfos,
		InsertRequestCancelInfos:  request.InsertRequestCancelInfos,
		InsertSignalInfos:         request.InsertSignalInfos,
		InsertSignalRequestedIDs:  request.InsertSignalRequestedIDs,
	}
	return m.persistence.ResetMutableState(newRequest)
}

func (m *executionManagerImpl) CreateWorkflowExecution(request *CreateWorkflowExecutionRequest) (*CreateWorkflowExecutionResponse, error) {
	return m.persistence.CreateWorkflowExecution(request)
}
func (m *executionManagerImpl) DeleteWorkflowExecution(request *DeleteWorkflowExecutionRequest) error {
	return m.persistence.DeleteWorkflowExecution(request)
}

func (m *executionManagerImpl) GetCurrentExecution(request *GetCurrentExecutionRequest) (*GetCurrentExecutionResponse, error) {
	return m.persistence.GetCurrentExecution(request)
}

// Transfer task related methods
func (m *executionManagerImpl) GetTransferTasks(request *GetTransferTasksRequest) (*GetTransferTasksResponse, error) {
	return m.persistence.GetTransferTasks(request)
}
func (m *executionManagerImpl) CompleteTransferTask(request *CompleteTransferTaskRequest) error {
	return m.persistence.CompleteTransferTask(request)
}
func (m *executionManagerImpl) RangeCompleteTransferTask(request *RangeCompleteTransferTaskRequest) error {
	return m.persistence.RangeCompleteTransferTask(request)
}

// Replication task related methods
func (m *executionManagerImpl) GetReplicationTasks(request *GetReplicationTasksRequest) (*GetReplicationTasksResponse, error) {
	return m.persistence.GetReplicationTasks(request)
}
func (m *executionManagerImpl) CompleteReplicationTask(request *CompleteReplicationTaskRequest) error {
	return m.persistence.CompleteReplicationTask(request)
}

// Timer related methods.
func (m *executionManagerImpl) GetTimerIndexTasks(request *GetTimerIndexTasksRequest) (*GetTimerIndexTasksResponse, error) {
	return m.persistence.GetTimerIndexTasks(request)
}
func (m *executionManagerImpl) CompleteTimerTask(request *CompleteTimerTaskRequest) error {
	return m.persistence.CompleteTimerTask(request)
}
func (m *executionManagerImpl) RangeCompleteTimerTask(request *RangeCompleteTimerTaskRequest) error {
	return m.persistence.RangeCompleteTimerTask(request)
}

//NewHistoryManagerImpl returns new HistoryManager
func NewHistoryManagerImpl(persistence PersistenceHistoryManager, logger bark.Logger) HistoryManager {
	return &historyManagerImpl{
		serializer:  NewHistorySerializer(),
		persistence: persistence,
		logger:      logger,
	}
}

func (m *historyManagerImpl) AppendHistoryEvents(request *AppendHistoryEventsRequest) (*AppendHistoryEventsResponse, error) {
	eventsData, err := m.serializer.SerializeBatchEvents(&workflow.History{Events: request.Events}, request.Encoding)
	if err != nil {
		return nil, err
	}

	resp := &AppendHistoryEventsResponse{Size: len(eventsData.Data)}
	return resp, m.persistence.AppendHistoryEvents(
		&PersistenceAppendHistoryEventsRequest{
			DomainID:          request.DomainID,
			Execution:         request.Execution,
			FirstEventID:      request.FirstEventID,
			EventBatchVersion: request.EventBatchVersion,
			RangeID:           request.RangeID,
			TransactionID:     request.TransactionID,
			Events:            eventsData,
			Overwrite:         request.Overwrite,
		})
}

// GetWorkflowExecutionHistory retrieves the paginated list of history events for given execution
func (m *historyManagerImpl) GetWorkflowExecutionHistory(request *GetWorkflowExecutionHistoryRequest) (*GetWorkflowExecutionHistoryResponse, error) {
	token, err := m.deserializeToken(request)
	if err != nil {
		return nil, err
	}

	// persistence API expects the actual cassandra paging token
	request.NextPageToken = token.Data
	newRequest := &PersistenceGetWorkflowExecutionHistoryRequest{
		LastEventBatchVersion: token.LastEventBatchVersion,
		NextPageToken:         token.Data,

		DomainID:     request.DomainID,
		Execution:    request.Execution,
		FirstEventID: request.FirstEventID,
		NextEventID:  request.NextEventID,
		PageSize:     request.PageSize,
	}
	response, err := m.persistence.GetWorkflowExecutionHistory(newRequest)
	if err != nil {
		return nil, err
	}
	// we store LastEventBatchVersion in the token. The reason we do it here is for historic reason.
	token.LastEventBatchVersion = response.LastEventBatchVersion

	newResponse := &GetWorkflowExecutionHistoryResponse{}

	history := &workflow.History{
		Events: make([]*workflow.HistoryEvent, 0),
	}

	// first_event_id of the last batch
	lastFirstEventID := common.EmptyEventID
	size := 0

	for _, b := range response.History {
		size += len(b.Data)
		historyBatch, err := m.serializer.DeserializeBatchEvents(b)
		if err != nil {
			return nil, err
		}

		if len(historyBatch.Events) == 0 || historyBatch.Events[0].GetEventId() > token.LastEventID+1 {
			logger := m.logger.WithFields(bark.Fields{
				logging.TagWorkflowExecutionID: request.Execution.GetWorkflowId(),
				logging.TagWorkflowRunID:       request.Execution.GetRunId(),
				logging.TagDomainID:            request.DomainID,
			})
			logger.Error("Unexpected event batch")
			return nil, fmt.Errorf("corrupted history event batch")
		}

		if historyBatch.Events[0].GetEventId() != token.LastEventID+1 {
			// staled event batch, skip it
			continue
		}

		lastFirstEventID = historyBatch.Events[0].GetEventId()
		history.Events = append(history.Events, historyBatch.Events...)
		token.LastEventID = historyBatch.Events[len(historyBatch.Events)-1].GetEventId()

		history.Events = append(history.Events, historyBatch.Events...)
	}

	newResponse.Size = size
	newResponse.LastFirstEventID = lastFirstEventID
	newResponse.History = history
	newResponse.NextPageToken, err = m.serializeToken(token)
	if err != nil {
		return nil, err
	}

	return newResponse, nil
}

func (m *historyManagerImpl) deserializeToken(request *GetWorkflowExecutionHistoryRequest) (*historyToken, error) {
	token := &historyToken{
		LastEventBatchVersion: common.EmptyVersion,
		LastEventID:           request.FirstEventID - 1,
	}

	if len(request.NextPageToken) == 0 {
		return token, nil
	}

	err := json.Unmarshal(request.NextPageToken, token)
	if err == nil {
		return token, nil
	}

	// for backward compatible reason, the input data can be raw Cassandra token
	token.Data = request.NextPageToken
	return token, nil
}

func (m *historyManagerImpl) serializeToken(token *historyToken) ([]byte, error) {
	if len(token.Data) == 0 {
		return nil, nil
	}

	data, err := json.Marshal(token)
	if err != nil {
		return nil, &workflow.InternalServiceError{Message: "Error generating history event token."}
	}
	return data, nil
}

func (m *historyManagerImpl) DeleteWorkflowExecutionHistory(request *DeleteWorkflowExecutionHistoryRequest) error {
	return m.persistence.DeleteWorkflowExecutionHistory(request)
}

func (m *executionManagerImpl) Close() {
	m.persistence.Close()
}

func (m *historyManagerImpl) Close() {
	m.persistence.Close()
}
