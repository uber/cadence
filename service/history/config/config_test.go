// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/types"
)

type configTestCase struct {
	key   dynamicconfig.Key
	value interface{}
}

func TestNewConfig(t *testing.T) {
	hostname := "hostname"
	numberOfShards := 8192
	maxMessageSize := 1024
	isAdvancedVisConfigExist := true
	fields := map[string]configTestCase{
		"NumberOfShards":                                       {nil, numberOfShards},
		"IsAdvancedVisConfigExist":                             {nil, isAdvancedVisConfigExist},
		"RPS":                                                  {dynamicconfig.HistoryRPS, 100},
		"MaxIDLengthWarnLimit":                                 {dynamicconfig.MaxIDLengthWarnLimit, 100},
		"DomainNameMaxLength":                                  {dynamicconfig.DomainNameMaxLength, 100},
		"IdentityMaxLength":                                    {dynamicconfig.IdentityMaxLength, 100},
		"WorkflowIDMaxLength":                                  {dynamicconfig.WorkflowIDMaxLength, 100},
		"SignalNameMaxLength":                                  {dynamicconfig.SignalNameMaxLength, 100},
		"WorkflowTypeMaxLength":                                {dynamicconfig.WorkflowTypeMaxLength, 100},
		"RequestIDMaxLength":                                   {dynamicconfig.RequestIDMaxLength, 100},
		"TaskListNameMaxLength":                                {dynamicconfig.TaskListNameMaxLength, 100},
		"ActivityIDMaxLength":                                  {dynamicconfig.ActivityIDMaxLength, 100},
		"ActivityTypeMaxLength":                                {dynamicconfig.ActivityTypeMaxLength, 100},
		"MarkerNameMaxLength":                                  {dynamicconfig.MarkerNameMaxLength, 100},
		"TimerIDMaxLength":                                     {dynamicconfig.TimerIDMaxLength, 100},
		"PersistenceMaxQPS":                                    {dynamicconfig.HistoryPersistenceMaxQPS, 100},
		"PersistenceGlobalMaxQPS":                              {dynamicconfig.HistoryPersistenceGlobalMaxQPS, 100},
		"EnableVisibilitySampling":                             {dynamicconfig.EnableVisibilitySampling, true},
		"EnableReadFromClosedExecutionV2":                      {dynamicconfig.EnableReadFromClosedExecutionV2, true},
		"VisibilityOpenMaxQPS":                                 {dynamicconfig.HistoryVisibilityOpenMaxQPS, 100},
		"VisibilityClosedMaxQPS":                               {dynamicconfig.HistoryVisibilityClosedMaxQPS, 100},
		"AdvancedVisibilityWritingMode":                        {dynamicconfig.AdvancedVisibilityWritingMode, "dual"},
		"AdvancedVisibilityMigrationWritingMode":               {dynamicconfig.AdvancedVisibilityMigrationWritingMode, "dual"},
		"EmitShardDiffLog":                                     {dynamicconfig.EmitShardDiffLog, true},
		"MaxAutoResetPoints":                                   {dynamicconfig.HistoryMaxAutoResetPoints, 100},
		"ThrottledLogRPS":                                      {dynamicconfig.HistoryThrottledLogRPS, 100},
		"EnableStickyQuery":                                    {dynamicconfig.EnableStickyQuery, true},
		"ShutdownDrainDuration":                                {dynamicconfig.HistoryShutdownDrainDuration, time.Second},
		"WorkflowDeletionJitterRange":                          {dynamicconfig.WorkflowDeletionJitterRange, 1},
		"DeleteHistoryEventContextTimeout":                     {dynamicconfig.DeleteHistoryEventContextTimeout, 1},
		"MaxResponseSize":                                      {nil, maxMessageSize},
		"HistoryCacheInitialSize":                              {dynamicconfig.HistoryCacheInitialSize, 100},
		"HistoryCacheMaxSize":                                  {dynamicconfig.HistoryCacheMaxSize, 100},
		"HistoryCacheTTL":                                      {dynamicconfig.HistoryCacheTTL, time.Second},
		"EventsCacheInitialCount":                              {dynamicconfig.EventsCacheInitialCount, 100},
		"EventsCacheMaxCount":                                  {dynamicconfig.EventsCacheMaxCount, 100},
		"EventsCacheMaxSize":                                   {dynamicconfig.EventsCacheMaxSize, 100},
		"EventsCacheTTL":                                       {dynamicconfig.EventsCacheTTL, time.Second},
		"EventsCacheGlobalEnable":                              {dynamicconfig.EventsCacheGlobalEnable, true},
		"EventsCacheGlobalInitialCount":                        {dynamicconfig.EventsCacheGlobalInitialCount, 100},
		"EventsCacheGlobalMaxCount":                            {dynamicconfig.EventsCacheGlobalMaxCount, 100},
		"RangeSizeBits":                                        {nil, uint(20)},
		"AcquireShardInterval":                                 {dynamicconfig.AcquireShardInterval, time.Second},
		"AcquireShardConcurrency":                              {dynamicconfig.AcquireShardConcurrency, 100},
		"StandbyClusterDelay":                                  {dynamicconfig.StandbyClusterDelay, time.Second},
		"StandbyTaskMissingEventsResendDelay":                  {dynamicconfig.StandbyTaskMissingEventsResendDelay, time.Second},
		"StandbyTaskMissingEventsDiscardDelay":                 {dynamicconfig.StandbyTaskMissingEventsDiscardDelay, time.Second},
		"TaskProcessRPS":                                       {dynamicconfig.TaskProcessRPS, 100},
		"TaskSchedulerType":                                    {dynamicconfig.TaskSchedulerType, 1},
		"TaskSchedulerWorkerCount":                             {dynamicconfig.TaskSchedulerWorkerCount, 100},
		"TaskSchedulerShardWorkerCount":                        {dynamicconfig.TaskSchedulerShardWorkerCount, 100},
		"TaskSchedulerQueueSize":                               {dynamicconfig.TaskSchedulerQueueSize, 100},
		"TaskSchedulerDispatcherCount":                         {dynamicconfig.TaskSchedulerDispatcherCount, 100},
		"TaskSchedulerRoundRobinWeights":                       {dynamicconfig.TaskSchedulerRoundRobinWeights, map[string]interface{}{"key": 1}},
		"TaskSchedulerShardQueueSize":                          {dynamicconfig.TaskSchedulerShardQueueSize, 100},
		"TaskCriticalRetryCount":                               {dynamicconfig.TaskCriticalRetryCount, 100},
		"ActiveTaskRedispatchInterval":                         {dynamicconfig.ActiveTaskRedispatchInterval, time.Second},
		"StandbyTaskRedispatchInterval":                        {dynamicconfig.StandbyTaskRedispatchInterval, time.Second},
		"TaskRedispatchIntervalJitterCoefficient":              {dynamicconfig.TaskRedispatchIntervalJitterCoefficient, 1.0},
		"StandbyTaskReReplicationContextTimeout":               {dynamicconfig.StandbyTaskReReplicationContextTimeout, time.Second},
		"EnableDropStuckTaskByDomainID":                        {dynamicconfig.EnableDropStuckTaskByDomainID, true},
		"ResurrectionCheckMinDelay":                            {dynamicconfig.ResurrectionCheckMinDelay, time.Second},
		"QueueProcessorEnableSplit":                            {dynamicconfig.QueueProcessorEnableSplit, true},
		"QueueProcessorSplitMaxLevel":                          {dynamicconfig.QueueProcessorSplitMaxLevel, 100},
		"QueueProcessorEnableRandomSplitByDomainID":            {dynamicconfig.QueueProcessorEnableRandomSplitByDomainID, true},
		"QueueProcessorRandomSplitProbability":                 {dynamicconfig.QueueProcessorRandomSplitProbability, 1.0},
		"QueueProcessorEnablePendingTaskSplitByDomainID":       {dynamicconfig.QueueProcessorEnablePendingTaskSplitByDomainID, true},
		"QueueProcessorPendingTaskSplitThreshold":              {dynamicconfig.QueueProcessorPendingTaskSplitThreshold, map[string]interface{}{"a": 100}},
		"QueueProcessorEnableStuckTaskSplitByDomainID":         {dynamicconfig.QueueProcessorEnableStuckTaskSplitByDomainID, true},
		"QueueProcessorStuckTaskSplitThreshold":                {dynamicconfig.QueueProcessorStuckTaskSplitThreshold, map[string]interface{}{"b": 1}},
		"QueueProcessorSplitLookAheadDurationByDomainID":       {dynamicconfig.QueueProcessorSplitLookAheadDurationByDomainID, time.Second},
		"QueueProcessorPollBackoffInterval":                    {dynamicconfig.QueueProcessorPollBackoffInterval, time.Second},
		"QueueProcessorPollBackoffIntervalJitterCoefficient":   {dynamicconfig.QueueProcessorPollBackoffIntervalJitterCoefficient, 1.0},
		"QueueProcessorEnableLoadQueueStates":                  {dynamicconfig.QueueProcessorEnableLoadQueueStates, true},
		"QueueProcessorEnableGracefulSyncShutdown":             {dynamicconfig.QueueProcessorEnableGracefulSyncShutdown, true},
		"QueueProcessorEnablePersistQueueStates":               {dynamicconfig.QueueProcessorEnablePersistQueueStates, true},
		"TimerTaskBatchSize":                                   {dynamicconfig.TimerTaskBatchSize, 100},
		"TimerTaskDeleteBatchSize":                             {dynamicconfig.TimerTaskDeleteBatchSize, 100},
		"TimerProcessorGetFailureRetryCount":                   {dynamicconfig.TimerProcessorGetFailureRetryCount, 100},
		"TimerProcessorCompleteTimerFailureRetryCount":         {dynamicconfig.TimerProcessorCompleteTimerFailureRetryCount, 100},
		"TimerProcessorUpdateAckInterval":                      {dynamicconfig.TimerProcessorUpdateAckInterval, time.Second},
		"TimerProcessorUpdateAckIntervalJitterCoefficient":     {dynamicconfig.TimerProcessorUpdateAckIntervalJitterCoefficient, 1.0},
		"TimerProcessorCompleteTimerInterval":                  {dynamicconfig.TimerProcessorCompleteTimerInterval, time.Second},
		"TimerProcessorFailoverMaxStartJitterInterval":         {dynamicconfig.TimerProcessorFailoverMaxStartJitterInterval, time.Second},
		"TimerProcessorFailoverMaxPollRPS":                     {dynamicconfig.TimerProcessorFailoverMaxPollRPS, 100},
		"TimerProcessorMaxPollRPS":                             {dynamicconfig.TimerProcessorMaxPollRPS, 100},
		"TimerProcessorMaxPollInterval":                        {dynamicconfig.TimerProcessorMaxPollInterval, time.Second},
		"TimerProcessorMaxPollIntervalJitterCoefficient":       {dynamicconfig.TimerProcessorMaxPollIntervalJitterCoefficient, 1.0},
		"TimerProcessorSplitQueueInterval":                     {dynamicconfig.TimerProcessorSplitQueueInterval, time.Second},
		"TimerProcessorSplitQueueIntervalJitterCoefficient":    {dynamicconfig.TimerProcessorSplitQueueIntervalJitterCoefficient, 1.0},
		"TimerProcessorMaxRedispatchQueueSize":                 {dynamicconfig.TimerProcessorMaxRedispatchQueueSize, 100},
		"TimerProcessorMaxTimeShift":                           {dynamicconfig.TimerProcessorMaxTimeShift, time.Second},
		"TimerProcessorHistoryArchivalSizeLimit":               {dynamicconfig.TimerProcessorHistoryArchivalSizeLimit, 100},
		"TimerProcessorArchivalTimeLimit":                      {dynamicconfig.TimerProcessorArchivalTimeLimit, time.Second},
		"TransferTaskBatchSize":                                {dynamicconfig.TransferTaskBatchSize, 100},
		"TransferTaskDeleteBatchSize":                          {dynamicconfig.TransferTaskDeleteBatchSize, 100},
		"TransferProcessorCompleteTransferFailureRetryCount":   {dynamicconfig.TransferProcessorCompleteTransferFailureRetryCount, 100},
		"TransferProcessorFailoverMaxStartJitterInterval":      {dynamicconfig.TransferProcessorFailoverMaxStartJitterInterval, time.Second},
		"TransferProcessorFailoverMaxPollRPS":                  {dynamicconfig.TransferProcessorFailoverMaxPollRPS, 100},
		"TransferProcessorMaxPollRPS":                          {dynamicconfig.TransferProcessorMaxPollRPS, 100},
		"TransferProcessorMaxPollInterval":                     {dynamicconfig.TransferProcessorMaxPollInterval, time.Second},
		"TransferProcessorMaxPollIntervalJitterCoefficient":    {dynamicconfig.TransferProcessorMaxPollIntervalJitterCoefficient, 1.0},
		"TransferProcessorSplitQueueInterval":                  {dynamicconfig.TransferProcessorSplitQueueInterval, time.Second},
		"TransferProcessorSplitQueueIntervalJitterCoefficient": {dynamicconfig.TransferProcessorSplitQueueIntervalJitterCoefficient, 1.0},
		"TransferProcessorUpdateAckInterval":                   {dynamicconfig.TransferProcessorUpdateAckInterval, time.Second},
		"TransferProcessorUpdateAckIntervalJitterCoefficient":  {dynamicconfig.TransferProcessorUpdateAckIntervalJitterCoefficient, 1.0},
		"TransferProcessorCompleteTransferInterval":            {dynamicconfig.TransferProcessorCompleteTransferInterval, time.Second},
		"TransferProcessorMaxRedispatchQueueSize":              {dynamicconfig.TransferProcessorMaxRedispatchQueueSize, 100},
		"TransferProcessorEnableValidator":                     {dynamicconfig.TransferProcessorEnableValidator, true},
		"TransferProcessorValidationInterval":                  {dynamicconfig.TransferProcessorValidationInterval, time.Second},
		"TransferProcessorVisibilityArchivalTimeLimit":         {dynamicconfig.TransferProcessorVisibilityArchivalTimeLimit, time.Second},
		"ReplicatorTaskDeleteBatchSize":                        {dynamicconfig.ReplicatorTaskDeleteBatchSize, 100},
		"ReplicatorReadTaskMaxRetryCount":                      {dynamicconfig.ReplicatorReadTaskMaxRetryCount, 100},
		"ReplicatorProcessorFetchTasksBatchSize":               {dynamicconfig.ReplicatorTaskBatchSize, 100},
		"ReplicatorUpperLatency":                               {dynamicconfig.ReplicatorUpperLatency, time.Second},
		"ReplicatorCacheCapacity":                              {dynamicconfig.ReplicatorCacheCapacity, 100},
		"ExecutionMgrNumConns":                                 {dynamicconfig.ExecutionMgrNumConns, 100},
		"HistoryMgrNumConns":                                   {dynamicconfig.HistoryMgrNumConns, 100},
		"MaximumBufferedEventsBatch":                           {dynamicconfig.MaximumBufferedEventsBatch, 100},
		"MaximumSignalsPerExecution":                           {dynamicconfig.MaximumSignalsPerExecution, 100},
		"ShardUpdateMinInterval":                               {dynamicconfig.ShardUpdateMinInterval, time.Second},
		"ShardSyncMinInterval":                                 {dynamicconfig.ShardSyncMinInterval, time.Second},
		"ShardSyncTimerJitterCoefficient":                      {dynamicconfig.TransferProcessorMaxPollIntervalJitterCoefficient, 1.0},
		"LongPollExpirationInterval":                           {dynamicconfig.HistoryLongPollExpirationInterval, time.Second},
		"EventEncodingType":                                    {dynamicconfig.DefaultEventEncoding, "eventEncodingType"},
		"EnableParentClosePolicy":                              {dynamicconfig.EnableParentClosePolicy, true},
		"EnableParentClosePolicyWorker":                        {dynamicconfig.EnableParentClosePolicyWorker, true},
		"ParentClosePolicyThreshold":                           {dynamicconfig.ParentClosePolicyThreshold, 100},
		"ParentClosePolicyBatchSize":                           {dynamicconfig.ParentClosePolicyBatchSize, 100},
		"NumParentClosePolicySystemWorkflows":                  {dynamicconfig.NumParentClosePolicySystemWorkflows, 100},
		"NumArchiveSystemWorkflows":                            {dynamicconfig.NumArchiveSystemWorkflows, 100},
		"ArchiveRequestRPS":                                    {dynamicconfig.ArchiveRequestRPS, 100},
		"ArchiveInlineHistoryRPS":                              {dynamicconfig.ArchiveInlineHistoryRPS, 100},
		"ArchiveInlineHistoryGlobalRPS":                        {dynamicconfig.ArchiveInlineHistoryGlobalRPS, 100},
		"ArchiveInlineVisibilityRPS":                           {dynamicconfig.ArchiveInlineVisibilityRPS, 100},
		"ArchiveInlineVisibilityGlobalRPS":                     {dynamicconfig.ArchiveInlineVisibilityGlobalRPS, 100},
		"AllowArchivingIncompleteHistory":                      {dynamicconfig.AllowArchivingIncompleteHistory, true},
		"BlobSizeLimitError":                                   {dynamicconfig.BlobSizeLimitError, 100},
		"BlobSizeLimitWarn":                                    {dynamicconfig.BlobSizeLimitWarn, 100},
		"HistorySizeLimitError":                                {dynamicconfig.HistorySizeLimitError, 100},
		"HistorySizeLimitWarn":                                 {dynamicconfig.HistorySizeLimitWarn, 100},
		"HistoryCountLimitError":                               {dynamicconfig.HistoryCountLimitError, 100},
		"HistoryCountLimitWarn":                                {dynamicconfig.HistoryCountLimitWarn, 100},
		"PendingActivitiesCountLimitError":                     {dynamicconfig.PendingActivitiesCountLimitError, 100},
		"PendingActivitiesCountLimitWarn":                      {dynamicconfig.PendingActivitiesCountLimitWarn, 100},
		"PendingActivityValidationEnabled":                     {dynamicconfig.EnablePendingActivityValidation, true},
		"EnableQueryAttributeValidation":                       {dynamicconfig.EnableQueryAttributeValidation, true},
		"ValidSearchAttributes":                                {dynamicconfig.ValidSearchAttributes, map[string]interface{}{"key": 1}},
		"SearchAttributesNumberOfKeysLimit":                    {dynamicconfig.SearchAttributesNumberOfKeysLimit, 100},
		"SearchAttributesSizeOfValueLimit":                     {dynamicconfig.SearchAttributesSizeOfValueLimit, 100},
		"SearchAttributesTotalSizeLimit":                       {dynamicconfig.SearchAttributesTotalSizeLimit, 100},
		"StickyTTL":                                            {dynamicconfig.StickyTTL, time.Second},
		"DecisionHeartbeatTimeout":                             {dynamicconfig.DecisionHeartbeatTimeout, time.Second},
		"MaxDecisionStartToCloseSeconds":                       {dynamicconfig.MaxDecisionStartToCloseSeconds, 100},
		"DecisionRetryCriticalAttempts":                        {dynamicconfig.DecisionRetryCriticalAttempts, 100},
		"DecisionRetryMaxAttempts":                             {dynamicconfig.DecisionRetryMaxAttempts, 100},
		"NormalDecisionScheduleToStartMaxAttempts":             {dynamicconfig.NormalDecisionScheduleToStartMaxAttempts, 100},
		"NormalDecisionScheduleToStartTimeout":                 {dynamicconfig.NormalDecisionScheduleToStartTimeout, time.Second},
		"ReplicationTaskFetcherParallelism":                    {dynamicconfig.ReplicationTaskFetcherParallelism, 100},
		"ReplicationTaskFetcherAggregationInterval":            {dynamicconfig.ReplicationTaskFetcherAggregationInterval, time.Second},
		"ReplicationTaskFetcherTimerJitterCoefficient":         {dynamicconfig.ReplicationTaskFetcherTimerJitterCoefficient, 1.0},
		"ReplicationTaskFetcherErrorRetryWait":                 {dynamicconfig.ReplicationTaskFetcherErrorRetryWait, time.Second},
		"ReplicationTaskFetcherServiceBusyWait":                {dynamicconfig.ReplicationTaskFetcherServiceBusyWait, time.Second},
		"ReplicationTaskFetcherEnableGracefulSyncShutdown":     {dynamicconfig.ReplicationTaskFetcherEnableGracefulSyncShutdown, true},
		"ReplicationTaskProcessorErrorRetryWait":               {dynamicconfig.ReplicationTaskProcessorErrorRetryWait, time.Second},
		"ReplicationTaskProcessorErrorRetryMaxAttempts":        {dynamicconfig.ReplicationTaskProcessorErrorRetryMaxAttempts, 100},
		"ReplicationTaskProcessorErrorSecondRetryWait":         {dynamicconfig.ReplicationTaskProcessorErrorSecondRetryWait, time.Second},
		"ReplicationTaskProcessorErrorSecondRetryExpiration":   {dynamicconfig.ReplicationTaskProcessorErrorSecondRetryExpiration, time.Second},
		"ReplicationTaskProcessorErrorSecondRetryMaxWait":      {dynamicconfig.ReplicationTaskProcessorErrorSecondRetryMaxWait, time.Second},
		"ReplicationTaskProcessorNoTaskRetryWait":              {dynamicconfig.ReplicationTaskProcessorNoTaskInitialWait, time.Second},
		"ReplicationTaskProcessorCleanupInterval":              {dynamicconfig.ReplicationTaskProcessorCleanupInterval, time.Second},
		"ReplicationTaskProcessorCleanupJitterCoefficient":     {dynamicconfig.ReplicationTaskProcessorCleanupJitterCoefficient, 1.0},
		"ReplicationTaskProcessorStartWait":                    {dynamicconfig.ReplicationTaskProcessorStartWait, time.Second},
		"ReplicationTaskProcessorStartWaitJitterCoefficient":   {dynamicconfig.ReplicationTaskProcessorStartWaitJitterCoefficient, 1.0},
		"ReplicationTaskProcessorHostQPS":                      {dynamicconfig.ReplicationTaskProcessorHostQPS, 100.0},
		"ReplicationTaskProcessorShardQPS":                     {dynamicconfig.ReplicationTaskProcessorShardQPS, 100.0},
		"ReplicationTaskGenerationQPS":                         {dynamicconfig.ReplicationTaskGenerationQPS, 100.0},
		"EnableReplicationTaskGeneration":                      {dynamicconfig.EnableReplicationTaskGeneration, true},
		"EnableRecordWorkflowExecutionUninitialized":           {dynamicconfig.EnableRecordWorkflowExecutionUninitialized, true},
		"WorkflowIDCacheExternalEnabled":                       {dynamicconfig.WorkflowIDCacheExternalEnabled, true},
		"WorkflowIDCacheInternalEnabled":                       {dynamicconfig.WorkflowIDCacheInternalEnabled, true},
		"WorkflowIDExternalRateLimitEnabled":                   {dynamicconfig.WorkflowIDExternalRateLimitEnabled, true},
		"WorkflowIDInternalRateLimitEnabled":                   {dynamicconfig.WorkflowIDInternalRateLimitEnabled, true},
		"WorkflowIDExternalRPS":                                {dynamicconfig.WorkflowIDExternalRPS, 100},
		"WorkflowIDInternalRPS":                                {dynamicconfig.WorkflowIDInternalRPS, 100},
		"EnableConsistentQuery":                                {dynamicconfig.EnableConsistentQuery, true},
		"EnableConsistentQueryByDomain":                        {dynamicconfig.EnableConsistentQueryByDomain, true},
		"MaxBufferedQueryCount":                                {dynamicconfig.MaxBufferedQueryCount, 100},
		"EnableContextHeaderInVisibility":                      {dynamicconfig.EnableContextHeaderInVisibility, true},
		"EnableCrossClusterOperationsForDomain":                {dynamicconfig.EnableCrossClusterOperationsForDomain, true},
		"MutableStateChecksumGenProbability":                   {dynamicconfig.MutableStateChecksumGenProbability, 10},
		"MutableStateChecksumVerifyProbability":                {dynamicconfig.MutableStateChecksumVerifyProbability, 10},
		"MutableStateChecksumInvalidateBefore":                 {dynamicconfig.MutableStateChecksumInvalidateBefore, 10.0},
		"EnableRetryForChecksumFailure":                        {dynamicconfig.EnableRetryForChecksumFailure, true},
		"EnableHistoryCorruptionCheck":                         {dynamicconfig.EnableHistoryCorruptionCheck, true},
		"NotifyFailoverMarkerInterval":                         {dynamicconfig.NotifyFailoverMarkerInterval, time.Second},
		"NotifyFailoverMarkerTimerJitterCoefficient":           {dynamicconfig.NotifyFailoverMarkerTimerJitterCoefficient, 1.0},
		"EnableGracefulFailover":                               {dynamicconfig.EnableGracefulFailover, true},
		"EnableActivityLocalDispatchByDomain":                  {dynamicconfig.EnableActivityLocalDispatchByDomain, true},
		"MaxActivityCountDispatchByDomain":                     {dynamicconfig.MaxActivityCountDispatchByDomain, 100},
		"ActivityMaxScheduleToStartTimeoutForRetry":            {dynamicconfig.ActivityMaxScheduleToStartTimeoutForRetry, time.Second},
		"EnableDebugMode":                                      {dynamicconfig.EnableDebugMode, true},
		"EnableTaskInfoLogByDomainID":                          {dynamicconfig.HistoryEnableTaskInfoLogByDomainID, true},
		"EnableTimerDebugLogByDomainID":                        {dynamicconfig.EnableTimerDebugLogByDomainID, true},
		"SampleLoggingRate":                                    {dynamicconfig.SampleLoggingRate, 100},
		"EnableShardIDMetrics":                                 {dynamicconfig.EnableShardIDMetrics, true},
		"LargeShardHistorySizeMetricThreshold":                 {dynamicconfig.LargeShardHistorySizeMetricThreshold, 100},
		"LargeShardHistoryEventMetricThreshold":                {dynamicconfig.LargeShardHistoryEventMetricThreshold, 100},
		"LargeShardHistoryBlobMetricThreshold":                 {dynamicconfig.LargeShardHistoryBlobMetricThreshold, 100},
		"EnableStrongIdempotency":                              {dynamicconfig.EnableStrongIdempotency, true},
		"EnableStrongIdempotencySanityCheck":                   {dynamicconfig.EnableStrongIdempotencySanityCheck, true},
		"GlobalRatelimiterNewDataWeight":                       {dynamicconfig.HistoryGlobalRatelimiterNewDataWeight, 1.0},
		"GlobalRatelimiterUpdateInterval":                      {dynamicconfig.GlobalRatelimiterUpdateInterval, time.Second},
		"GlobalRatelimiterDecayAfter":                          {dynamicconfig.HistoryGlobalRatelimiterDecayAfter, time.Second},
		"GlobalRatelimiterGCAfter":                             {dynamicconfig.HistoryGlobalRatelimiterGCAfter, time.Second},
		"HostName":                                             {nil, hostname},
	}
	client := dynamicconfig.NewInMemoryClient()
	for fieldName, expected := range fields {
		if expected.key != nil {
			err := client.UpdateValue(expected.key, expected.value)
			if err != nil {
				t.Errorf("Failed to update config for %s: %s", fieldName, err)
			}
		}
	}
	dc := dynamicconfig.NewCollection(client, testlogger.New(t))

	config := New(dc, numberOfShards, maxMessageSize, isAdvancedVisConfigExist, hostname)

	assertFieldsMatch(t, *config, fields)
}

func assertFieldsMatch(t *testing.T, config interface{}, fields map[string]configTestCase) {
	configType := reflect.ValueOf(config)

	for i := 0; i < configType.NumField(); i++ {
		f := configType.Field(i)
		fieldName := configType.Type().Field(i).Name

		if expected, ok := fields[fieldName]; ok {
			actual := getValue(&f)
			if f.Kind() == reflect.Slice {
				assert.ElementsMatch(t, expected.value, actual, "Incorrect value for field: %s", fieldName)
			} else {
				assert.Equal(t, expected.value, actual, "Incorrect value for field: %s", fieldName)
			}

		} else {
			t.Errorf("Unknown property on Config: %s", fieldName)
		}
	}
}

func getValue(f *reflect.Value) interface{} {
	switch f.Kind() {
	case reflect.Func:
		switch fn := f.Interface().(type) {
		case dynamicconfig.IntPropertyFn:
			return fn()
		case dynamicconfig.IntPropertyFnWithDomainFilter:
			return fn("domain")
		case dynamicconfig.IntPropertyFnWithTaskListInfoFilters:
			return fn("domain", "tasklist", int(types.TaskListTypeDecision))
		case dynamicconfig.BoolPropertyFn:
			return fn()
		case dynamicconfig.BoolPropertyFnWithDomainFilter:
			return fn("domain")
		case dynamicconfig.BoolPropertyFnWithDomainIDFilter:
			return fn("domain")
		case dynamicconfig.BoolPropertyFnWithTaskListInfoFilters:
			return fn("domain", "tasklist", int(types.TaskListTypeDecision))
		case dynamicconfig.DurationPropertyFn:
			return fn()
		case dynamicconfig.DurationPropertyFnWithDomainFilter:
			return fn("domain")
		case dynamicconfig.DurationPropertyFnWithTaskListInfoFilters:
			return fn("domain", "tasklist", int(types.TaskListTypeDecision))
		case dynamicconfig.FloatPropertyFn:
			return fn()
		case dynamicconfig.MapPropertyFn:
			return fn()
		case dynamicconfig.StringPropertyFn:
			return fn()
		case dynamicconfig.DurationPropertyFnWithDomainIDFilter:
			return fn("domain")
		case dynamicconfig.IntPropertyFnWithShardIDFilter:
			return fn(0)
		case dynamicconfig.StringPropertyFnWithDomainFilter:
			return fn("domain")
		case dynamicconfig.DurationPropertyFnWithShardIDFilter:
			return fn(0)
		case dynamicconfig.FloatPropertyFnWithShardIDFilter:
			return fn(0)
		case dynamicconfig.BoolPropertyFnWithDomainIDAndWorkflowIDFilter:
			return fn("domain", "workflowID")
		case func() []string:
			return fn()
		default:
			panic("Unable to handle type: " + f.Type().Name())
		}
	default:
		return f.Interface()
	}
}

func isolationGroupsHelper() []string {
	return []string{"zone-1", "zone-2"}
}
