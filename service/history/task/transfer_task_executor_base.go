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

package task

import (
	"context"
	"time"

	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/workflowcache"
	"github.com/uber/cadence/service/worker/archiver"
)

const (
	taskDefaultTimeout             = 3 * time.Second
	taskGetExecutionContextTimeout = 500 * time.Millisecond
	taskRPCCallTimeout             = 2 * time.Second

	secondsInDay      = int32(24 * time.Hour / time.Second)
	defaultDomainName = "defaultDomainName"
)

type (
	transferTaskExecutorBase struct {
		shard          shard.Context
		archiverClient archiver.Client
		executionCache *execution.Cache
		logger         log.Logger
		metricsClient  metrics.Client
		matchingClient matching.Client
		visibilityMgr  persistence.VisibilityManager
		config         *config.Config
		throttleRetry  *backoff.ThrottleRetry
		wfIDCache      workflowcache.WFCache
	}
)

func newTransferTaskExecutorBase(
	shard shard.Context,
	archiverClient archiver.Client,
	executionCache *execution.Cache,
	logger log.Logger,
	config *config.Config,
	wfIDCache workflowcache.WFCache,
) *transferTaskExecutorBase {
	return &transferTaskExecutorBase{
		shard:          shard,
		archiverClient: archiverClient,
		executionCache: executionCache,
		logger:         logger,
		metricsClient:  shard.GetMetricsClient(),
		matchingClient: shard.GetService().GetMatchingClient(),
		visibilityMgr:  shard.GetService().GetVisibilityManager(),
		config:         config,
		throttleRetry: backoff.NewThrottleRetry(
			backoff.WithRetryPolicy(taskRetryPolicy),
			backoff.WithRetryableError(common.IsServiceTransientError),
		),
		wfIDCache: wfIDCache,
	}
}

func (t *transferTaskExecutorBase) pushActivity(
	ctx context.Context,
	task *persistence.TransferTaskInfo,
	activityScheduleToStartTimeout int32,
	partitionConfig map[string]string,
) error {

	ctx, cancel := context.WithTimeout(ctx, taskRPCCallTimeout)
	defer cancel()

	if task.TaskType != persistence.TransferTaskTypeActivityTask {
		t.logger.Fatal("Cannot process non activity task", tag.TaskType(task.GetTaskType()))
	}

	// Ratelimiting is not done. This is only to count the number of requests via metrics
	t.wfIDCache.AllowInternal(task.DomainID, task.WorkflowID)

	return t.matchingClient.AddActivityTask(ctx, &types.AddActivityTaskRequest{
		DomainUUID:       task.TargetDomainID,
		SourceDomainUUID: task.DomainID,
		Execution: &types.WorkflowExecution{
			WorkflowID: task.WorkflowID,
			RunID:      task.RunID,
		},
		TaskList:                      &types.TaskList{Name: task.TaskList},
		ScheduleID:                    task.ScheduleID,
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(activityScheduleToStartTimeout),
		PartitionConfig:               partitionConfig,
	})
}

func (t *transferTaskExecutorBase) pushDecision(
	ctx context.Context,
	task *persistence.TransferTaskInfo,
	tasklist *types.TaskList,
	decisionScheduleToStartTimeout int32,
	partitionConfig map[string]string,
) error {

	ctx, cancel := context.WithTimeout(ctx, taskRPCCallTimeout)
	defer cancel()

	if task.TaskType != persistence.TransferTaskTypeDecisionTask {
		t.logger.Fatal("Cannot process non decision task", tag.TaskType(task.GetTaskType()))
	}

	return t.matchingClient.AddDecisionTask(ctx, &types.AddDecisionTaskRequest{
		DomainUUID: task.DomainID,
		Execution: &types.WorkflowExecution{
			WorkflowID: task.WorkflowID,
			RunID:      task.RunID,
		},
		TaskList:                      tasklist,
		ScheduleID:                    task.ScheduleID,
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(decisionScheduleToStartTimeout),
		PartitionConfig:               partitionConfig,
	})
}

func (t *transferTaskExecutorBase) recordWorkflowStarted(
	ctx context.Context,
	domainID string,
	workflowID string,
	runID string,
	workflowTypeName string,
	startTimeUnixNano int64,
	executionTimeUnixNano int64,
	workflowTimeout int32,
	taskID int64,
	taskList string,
	isCron bool,
	numClusters int16,
	visibilityMemo *types.Memo,
	updateTimeUnixNano int64,
	searchAttributes map[string][]byte,
) error {

	domain := defaultDomainName

	if domainEntry, err := t.shard.GetDomainCache().GetDomainByID(domainID); err != nil {
		if _, ok := err.(*types.EntityNotExistsError); !ok {
			return err
		}
	} else {
		domain = domainEntry.GetInfo().Name
		// if sampled for longer retention is enabled, only record those sampled events
		if domainEntry.IsSampledForLongerRetentionEnabled(workflowID) &&
			!domainEntry.IsSampledForLongerRetention(workflowID) {
			return nil
		}
	}

	request := &persistence.RecordWorkflowExecutionStartedRequest{
		DomainUUID: domainID,
		Domain:     domain,
		Execution: types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
		WorkflowTypeName:   workflowTypeName,
		StartTimestamp:     startTimeUnixNano,
		ExecutionTimestamp: executionTimeUnixNano,
		WorkflowTimeout:    int64(workflowTimeout),
		TaskID:             taskID,
		Memo:               visibilityMemo,
		TaskList:           taskList,
		IsCron:             isCron,
		NumClusters:        numClusters,
		UpdateTimestamp:    updateTimeUnixNano,
		SearchAttributes:   searchAttributes,
		ShardID:            int16(t.shard.GetShardID()),
	}

	if t.config.EnableRecordWorkflowExecutionUninitialized(domain) {
		uninitializedRequest := &persistence.RecordWorkflowExecutionUninitializedRequest{
			DomainUUID: domainID,
			Domain:     domain,
			Execution: types.WorkflowExecution{
				WorkflowID: workflowID,
				RunID:      runID,
			},
			WorkflowTypeName: workflowTypeName,
			UpdateTimestamp:  updateTimeUnixNano,
			ShardID:          int64(t.shard.GetShardID()),
		}
		if err := t.visibilityMgr.RecordWorkflowExecutionUninitialized(ctx, uninitializedRequest); err != nil {
			t.logger.Error("Failed to record uninitialized workflow execution", tag.Error(err))
		}
	}

	return t.visibilityMgr.RecordWorkflowExecutionStarted(ctx, request)
}

func (t *transferTaskExecutorBase) upsertWorkflowExecution(
	ctx context.Context,
	domainID string,
	workflowID string,
	runID string,
	workflowTypeName string,
	startTimeUnixNano int64,
	executionTimeUnixNano int64,
	workflowTimeout int32,
	taskID int64,
	taskList string,
	visibilityMemo *types.Memo,
	isCron bool,
	numClusters int16,
	updateTimeUnixNano int64,
	searchAttributes map[string][]byte,
) error {

	domain, err := t.shard.GetDomainCache().GetDomainName(domainID)
	if err != nil {
		if _, ok := err.(*types.EntityNotExistsError); !ok {
			return err
		}
		domain = defaultDomainName
	}

	request := &persistence.UpsertWorkflowExecutionRequest{
		DomainUUID: domainID,
		Domain:     domain,
		Execution: types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
		WorkflowTypeName:   workflowTypeName,
		StartTimestamp:     startTimeUnixNano,
		ExecutionTimestamp: executionTimeUnixNano,
		WorkflowTimeout:    int64(workflowTimeout),
		TaskID:             taskID,
		Memo:               visibilityMemo,
		TaskList:           taskList,
		IsCron:             isCron,
		NumClusters:        numClusters,
		SearchAttributes:   searchAttributes,
		UpdateTimestamp:    updateTimeUnixNano,
		ShardID:            int64(t.shard.GetShardID()),
	}

	return t.visibilityMgr.UpsertWorkflowExecution(ctx, request)
}

func (t *transferTaskExecutorBase) recordWorkflowClosed(
	ctx context.Context,
	domainID string,
	workflowID string,
	runID string,
	workflowTypeName string,
	startTimeUnixNano int64,
	executionTimeUnixNano int64,
	endTimeUnixNano int64,
	closeStatus types.WorkflowExecutionCloseStatus,
	historyLength int64,
	taskID int64,
	visibilityMemo *types.Memo,
	taskList string,
	isCron bool,
	numClusters int16,
	updateTimeUnixNano int64,
	searchAttributes map[string][]byte,
) error {

	// Record closing in visibility store
	retentionSeconds := int64(0)
	domain := defaultDomainName
	recordWorkflowClose := true
	archiveVisibility := false

	domainEntry, err := t.shard.GetDomainCache().GetDomainByID(domainID)
	if err != nil && !isWorkflowNotExistError(err) {
		return err
	}

	if err == nil {
		// retention in domain config is in days, convert to seconds
		retentionSeconds = int64(domainEntry.GetRetentionDays(workflowID)) * int64(secondsInDay)
		domain = domainEntry.GetInfo().Name
		// if sampled for longer retention is enabled, only record those sampled events
		if domainEntry.IsSampledForLongerRetentionEnabled(workflowID) &&
			!domainEntry.IsSampledForLongerRetention(workflowID) {
			recordWorkflowClose = false
		}

		clusterConfiguredForVisibilityArchival := t.shard.GetService().GetArchivalMetadata().GetVisibilityConfig().ClusterConfiguredForArchival()
		domainConfiguredForVisibilityArchival := domainEntry.GetConfig().VisibilityArchivalStatus == types.ArchivalStatusEnabled
		archiveVisibility = clusterConfiguredForVisibilityArchival && domainConfiguredForVisibilityArchival
	}

	if recordWorkflowClose {
		if err := t.visibilityMgr.RecordWorkflowExecutionClosed(ctx, &persistence.RecordWorkflowExecutionClosedRequest{
			DomainUUID: domainID,
			Domain:     domain,
			Execution: types.WorkflowExecution{
				WorkflowID: workflowID,
				RunID:      runID,
			},
			WorkflowTypeName:   workflowTypeName,
			StartTimestamp:     startTimeUnixNano,
			ExecutionTimestamp: executionTimeUnixNano,
			CloseTimestamp:     endTimeUnixNano,
			Status:             closeStatus,
			HistoryLength:      historyLength,
			RetentionSeconds:   retentionSeconds,
			TaskID:             taskID,
			Memo:               visibilityMemo,
			TaskList:           taskList,
			SearchAttributes:   searchAttributes,
			IsCron:             isCron,
			UpdateTimestamp:    updateTimeUnixNano,
			NumClusters:        numClusters,
			ShardID:            int16(t.shard.GetShardID()),
		}); err != nil {
			return err
		}
	}

	if archiveVisibility {
		archiveCtx, cancel := context.WithTimeout(ctx, t.config.TransferProcessorVisibilityArchivalTimeLimit())
		defer cancel()
		_, err := t.archiverClient.Archive(archiveCtx, &archiver.ClientRequest{
			ArchiveRequest: &archiver.ArchiveRequest{
				DomainID:           domainID,
				DomainName:         domain,
				WorkflowID:         workflowID,
				RunID:              runID,
				WorkflowTypeName:   workflowTypeName,
				StartTimestamp:     startTimeUnixNano,
				ExecutionTimestamp: executionTimeUnixNano,
				CloseTimestamp:     endTimeUnixNano,
				CloseStatus:        closeStatus,
				HistoryLength:      historyLength,
				Memo:               visibilityMemo,
				SearchAttributes:   searchAttributes,
				VisibilityURI:      domainEntry.GetConfig().VisibilityArchivalURI,
				URI:                domainEntry.GetConfig().HistoryArchivalURI,
				Targets:            []archiver.ArchivalTarget{archiver.ArchiveTargetVisibility},
			},
			CallerService:        service.History,
			AttemptArchiveInline: true, // archive visibility inline by default
		})
		return err
	}
	return nil
}

// Argument startEvent is to save additional call of msBuilder.GetStartEvent
func getWorkflowExecutionTimestamp(
	msBuilder execution.MutableState,
	startEvent *types.HistoryEvent,
) time.Time {
	// Use value 0 to represent workflows that don't need backoff. Since ES doesn't support
	// comparison between two field, we need a value to differentiate them from cron workflows
	// or later runs of a workflow that needs retry.
	executionTimestamp := time.Unix(0, 0)
	if startEvent == nil {
		return executionTimestamp
	}

	if backoffSeconds := startEvent.WorkflowExecutionStartedEventAttributes.GetFirstDecisionTaskBackoffSeconds(); backoffSeconds != 0 {
		startTimestamp := time.Unix(0, startEvent.GetTimestamp())
		executionTimestamp = startTimestamp.Add(time.Duration(backoffSeconds) * time.Second)
	}
	return executionTimestamp
}

func getWorkflowMemo(
	memo map[string][]byte,
) *types.Memo {

	if memo == nil {
		return nil
	}
	return &types.Memo{Fields: memo}
}

func copySearchAttributes(
	input map[string][]byte,
) map[string][]byte {

	if input == nil {
		return nil
	}

	result := make(map[string][]byte)
	for k, v := range input {
		val := make([]byte, len(v))
		copy(val, v)
		result[k] = val
	}
	return result
}

func isWorkflowNotExistError(err error) bool {
	_, ok := err.(*types.EntityNotExistsError)
	return ok
}
