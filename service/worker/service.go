// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
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

package worker

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/archiver"
	"github.com/uber/cadence/service/worker/batcher"
	"github.com/uber/cadence/service/worker/esanalyzer"
	"github.com/uber/cadence/service/worker/failovermanager"
	"github.com/uber/cadence/service/worker/indexer"
	"github.com/uber/cadence/service/worker/parentclosepolicy"
	"github.com/uber/cadence/service/worker/replicator"
	"github.com/uber/cadence/service/worker/scanner"
	"github.com/uber/cadence/service/worker/scanner/executions"
	"github.com/uber/cadence/service/worker/scanner/shardscanner"
	"github.com/uber/cadence/service/worker/scanner/tasklist"
	"github.com/uber/cadence/service/worker/scanner/timers"
	"github.com/uber/cadence/service/worker/shadower"
	"github.com/uber/cadence/service/worker/watchdog"
)

type (
	// Service represents the cadence-worker service. This service hosts all background processing needed for cadence cluster:
	// 1. Replicator: Handles applying replication tasks generated by remote clusters.
	// 2. Indexer: Handles uploading of visibility records to elastic search.
	// 3. Archiver: Handles archival of workflow histories.
	Service struct {
		resource.Resource

		status int32
		stopC  chan struct{}
		params *resource.Params
		config *Config
	}

	// Config contains all the service config for worker
	Config struct {
		ArchiverConfig                      *archiver.Config
		IndexerCfg                          *indexer.Config
		ScannerCfg                          *scanner.Config
		BatcherCfg                          *batcher.Config
		ESAnalyzerCfg                       *esanalyzer.Config
		WatchdogConfig                      *watchdog.Config
		failoverManagerCfg                  *failovermanager.Config
		ThrottledLogRPS                     dynamicconfig.IntPropertyFn
		PersistenceGlobalMaxQPS             dynamicconfig.IntPropertyFn
		PersistenceMaxQPS                   dynamicconfig.IntPropertyFn
		EnableBatcher                       dynamicconfig.BoolPropertyFn
		EnableParentClosePolicyWorker       dynamicconfig.BoolPropertyFn
		NumParentClosePolicySystemWorkflows dynamicconfig.IntPropertyFn
		EnableFailoverManager               dynamicconfig.BoolPropertyFn
		EnableWorkflowShadower              dynamicconfig.BoolPropertyFn
		DomainReplicationMaxRetryDuration   dynamicconfig.DurationPropertyFn
		EnableESAnalyzer                    dynamicconfig.BoolPropertyFn
		EnableWatchDog                      dynamicconfig.BoolPropertyFn
	}
)

// NewService builds a new cadence-worker service
func NewService(
	params *resource.Params,
) (resource.Resource, error) {

	serviceConfig := NewConfig(params)

	serviceResource, err := resource.New(
		params,
		service.Worker,
		&service.Config{
			PersistenceMaxQPS:       serviceConfig.PersistenceMaxQPS,
			PersistenceGlobalMaxQPS: serviceConfig.PersistenceGlobalMaxQPS,
			ThrottledLoggerMaxRPS:   serviceConfig.ThrottledLogRPS,
			// worker service doesn't need visibility config as it never call visibilityManager API
		},
	)
	if err != nil {
		return nil, err
	}

	return &Service{
		Resource: serviceResource,
		status:   common.DaemonStatusInitialized,
		config:   serviceConfig,
		params:   params,
		stopC:    make(chan struct{}),
	}, nil
}

// NewConfig builds the new Config for cadence-worker service
func NewConfig(params *resource.Params) *Config {
	dc := dynamicconfig.NewCollection(
		params.DynamicConfig,
		params.Logger,
		dynamicconfig.ClusterNameFilter(params.ClusterMetadata.GetCurrentClusterName()),
	)
	config := &Config{
		ArchiverConfig: &archiver.Config{
			ArchiverConcurrency:             dc.GetIntProperty(dynamicconfig.WorkerArchiverConcurrency),
			ArchivalsPerIteration:           dc.GetIntProperty(dynamicconfig.WorkerArchivalsPerIteration),
			TimeLimitPerArchivalIteration:   dc.GetDurationProperty(dynamicconfig.WorkerTimeLimitPerArchivalIteration),
			AllowArchivingIncompleteHistory: dc.GetBoolProperty(dynamicconfig.AllowArchivingIncompleteHistory),
		},
		ScannerCfg: &scanner.Config{
			ScannerPersistenceMaxQPS: dc.GetIntProperty(dynamicconfig.ScannerPersistenceMaxQPS),
			TaskListScannerOptions: tasklist.Options{
				GetOrphanTasksPageSizeFn: dc.GetIntProperty(dynamicconfig.ScannerGetOrphanTasksPageSize),
				TaskBatchSizeFn:          dc.GetIntProperty(dynamicconfig.ScannerBatchSizeForTasklistHandler),
				EnableCleaning:           dc.GetBoolProperty(dynamicconfig.EnableCleaningOrphanTaskInTasklistScavenger),
				MaxTasksPerJobFn:         dc.GetIntProperty(dynamicconfig.ScannerMaxTasksProcessedPerTasklistJob),
			},
			Persistence:            &params.PersistenceConfig,
			ClusterMetadata:        params.ClusterMetadata,
			TaskListScannerEnabled: dc.GetBoolProperty(dynamicconfig.TaskListScannerEnabled),
			HistoryScannerEnabled:  dc.GetBoolProperty(dynamicconfig.HistoryScannerEnabled),
			ShardScanners: []*shardscanner.ScannerConfig{
				executions.ConcreteExecutionScannerConfig(dc),
				executions.CurrentExecutionScannerConfig(dc),
				timers.ScannerConfig(dc),
			},
			MaxWorkflowRetentionInDays: dc.GetIntProperty(dynamicconfig.MaxRetentionDays),
		},
		BatcherCfg: &batcher.Config{
			AdminOperationToken: dc.GetStringProperty(dynamicconfig.AdminOperationToken),
			ClusterMetadata:     params.ClusterMetadata,
		},
		failoverManagerCfg: &failovermanager.Config{
			AdminOperationToken: dc.GetStringProperty(dynamicconfig.AdminOperationToken),
			ClusterMetadata:     params.ClusterMetadata,
		},
		ESAnalyzerCfg: &esanalyzer.Config{
			ESAnalyzerPause:                          dc.GetBoolProperty(dynamicconfig.ESAnalyzerPause),
			ESAnalyzerTimeWindow:                     dc.GetDurationProperty(dynamicconfig.ESAnalyzerTimeWindow),
			ESAnalyzerMaxNumDomains:                  dc.GetIntProperty(dynamicconfig.ESAnalyzerMaxNumDomains),
			ESAnalyzerMaxNumWorkflowTypes:            dc.GetIntProperty(dynamicconfig.ESAnalyzerMaxNumWorkflowTypes),
			ESAnalyzerLimitToTypes:                   dc.GetStringProperty(dynamicconfig.ESAnalyzerLimitToTypes),
			ESAnalyzerEnableAvgDurationBasedChecks:   dc.GetBoolProperty(dynamicconfig.ESAnalyzerEnableAvgDurationBasedChecks),
			ESAnalyzerLimitToDomains:                 dc.GetStringProperty(dynamicconfig.ESAnalyzerLimitToDomains),
			ESAnalyzerNumWorkflowsToRefresh:          dc.GetIntPropertyFilteredByWorkflowType(dynamicconfig.ESAnalyzerNumWorkflowsToRefresh),
			ESAnalyzerBufferWaitTime:                 dc.GetDurationPropertyFilteredByWorkflowType(dynamicconfig.ESAnalyzerBufferWaitTime),
			ESAnalyzerMinNumWorkflowsForAvg:          dc.GetIntPropertyFilteredByWorkflowType(dynamicconfig.ESAnalyzerMinNumWorkflowsForAvg),
			ESAnalyzerWorkflowDurationWarnThresholds: dc.GetStringProperty(dynamicconfig.ESAnalyzerWorkflowDurationWarnThresholds),
			ESAnalyzerWorkflowVersionDomains:         dc.GetStringProperty(dynamicconfig.ESAnalyzerWorkflowVersionMetricDomains),
		},
		WatchdogConfig: &watchdog.Config{
			CorruptWorkflowWatchdogPause: dc.GetBoolProperty(dynamicconfig.CorruptWorkflowWatchdogPause),
		},
		EnableBatcher:                       dc.GetBoolProperty(dynamicconfig.EnableBatcher),
		EnableParentClosePolicyWorker:       dc.GetBoolProperty(dynamicconfig.EnableParentClosePolicyWorker),
		NumParentClosePolicySystemWorkflows: dc.GetIntProperty(dynamicconfig.NumParentClosePolicySystemWorkflows),
		EnableESAnalyzer:                    dc.GetBoolProperty(dynamicconfig.EnableESAnalyzer),
		EnableWatchDog:                      dc.GetBoolProperty(dynamicconfig.EnableWatchDog),
		EnableFailoverManager:               dc.GetBoolProperty(dynamicconfig.EnableFailoverManager),
		EnableWorkflowShadower:              dc.GetBoolProperty(dynamicconfig.EnableWorkflowShadower),
		ThrottledLogRPS:                     dc.GetIntProperty(dynamicconfig.WorkerThrottledLogRPS),
		PersistenceGlobalMaxQPS:             dc.GetIntProperty(dynamicconfig.WorkerPersistenceGlobalMaxQPS),
		PersistenceMaxQPS:                   dc.GetIntProperty(dynamicconfig.WorkerPersistenceMaxQPS),
		DomainReplicationMaxRetryDuration:   dc.GetDurationProperty(dynamicconfig.WorkerReplicationTaskMaxRetryDuration),
	}
	advancedVisWritingMode := dc.GetStringProperty(
		dynamicconfig.AdvancedVisibilityWritingMode,
	)
	if common.IsAdvancedVisibilityWritingEnabled(advancedVisWritingMode(), params.PersistenceConfig.IsAdvancedVisibilityConfigExist()) {
		config.IndexerCfg = &indexer.Config{
			IndexerConcurrency:       dc.GetIntProperty(dynamicconfig.WorkerIndexerConcurrency),
			ESProcessorNumOfWorkers:  dc.GetIntProperty(dynamicconfig.WorkerESProcessorNumOfWorkers),
			ESProcessorBulkActions:   dc.GetIntProperty(dynamicconfig.WorkerESProcessorBulkActions),
			ESProcessorBulkSize:      dc.GetIntProperty(dynamicconfig.WorkerESProcessorBulkSize),
			ESProcessorFlushInterval: dc.GetDurationProperty(dynamicconfig.WorkerESProcessorFlushInterval),
			ValidSearchAttributes:    dc.GetMapProperty(dynamicconfig.ValidSearchAttributes),
		}
	}
	return config
}

// Start is called to start the service
func (s *Service) Start() {
	if !atomic.CompareAndSwapInt32(&s.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}
	logger := s.GetLogger()
	logger.Info("worker starting", tag.ComponentWorker)

	s.Resource.Start()
	s.Resource.GetDomainReplicationQueue().Start()

	s.ensureDomainExists(common.SystemLocalDomainName)
	s.startScanner()
	s.startFixerWorkflowWorker()
	if s.config.IndexerCfg != nil {
		s.startIndexer()
	}

	s.startReplicator()

	if s.GetArchivalMetadata().GetHistoryConfig().ClusterConfiguredForArchival() {
		s.startArchiver()
	}
	if s.config.EnableBatcher() {
		s.ensureDomainExists(common.BatcherLocalDomainName)
		s.startBatcher()
	}
	if s.config.EnableParentClosePolicyWorker() {
		s.startParentClosePolicyProcessor()
	}
	if s.config.EnableESAnalyzer() {
		s.startESAnalyzer()
	}
	if s.config.EnableWatchDog() {
		s.startWatchDog()
	}
	if s.config.EnableFailoverManager() {
		s.startFailoverManager()
	}
	if s.config.EnableWorkflowShadower() {
		s.ensureDomainExists(common.ShadowerLocalDomainName)
		s.startWorkflowShadower()
	}

	logger.Info("worker started", tag.ComponentWorker)
	<-s.stopC
}

// Stop is called to stop the service
func (s *Service) Stop() {
	if !atomic.CompareAndSwapInt32(&s.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	close(s.stopC)

	s.Resource.Stop()
	s.Resource.GetDomainReplicationQueue().Stop()

	s.params.Logger.Info("worker stopped", tag.ComponentWorker)
}

func (s *Service) startParentClosePolicyProcessor() {
	params := &parentclosepolicy.BootstrapParams{
		ServiceClient: s.params.PublicClient,
		MetricsClient: s.GetMetricsClient(),
		Logger:        s.GetLogger(),
		TallyScope:    s.params.MetricScope,
		ClientBean:    s.GetClientBean(),
		DomainCache:   s.GetDomainCache(),
		NumWorkflows:  s.config.NumParentClosePolicySystemWorkflows(),
	}
	processor := parentclosepolicy.New(params)
	if err := processor.Start(); err != nil {
		s.GetLogger().Fatal("error starting parentclosepolicy processor", tag.Error(err))
	}
}

func (s *Service) startESAnalyzer() {
	analyzer := esanalyzer.New(
		s.params.PublicClient,
		s.GetFrontendClient(),
		s.GetClientBean(),
		s.params.ESClient,
		s.params.ESConfig,
		s.GetLogger(),
		s.params.MetricScope,
		s.Resource,
		s.GetDomainCache(),
		s.config.ESAnalyzerCfg,
	)

	if err := analyzer.Start(); err != nil {
		s.GetLogger().Fatal("error starting esanalyzer", tag.Error(err))
	}
}

func (s *Service) startWatchDog() {
	watchdog := watchdog.New(
		s.params.PublicClient,
		s.GetFrontendClient(),
		s.GetClientBean(),
		s.GetLogger(),
		s.GetMetricsClient(),
		s.params.MetricScope,
		s.Resource,
		s.GetDomainCache(),
		s.config.WatchdogConfig,
	)

	if err := watchdog.Start(); err != nil {
		s.GetLogger().Fatal("error starting watchdog", tag.Error(err))
	}
}

func (s *Service) startBatcher() {
	params := &batcher.BootstrapParams{
		Config:        *s.config.BatcherCfg,
		ServiceClient: s.params.PublicClient,
		MetricsClient: s.GetMetricsClient(),
		Logger:        s.GetLogger(),
		TallyScope:    s.params.MetricScope,
		ClientBean:    s.GetClientBean(),
	}
	if err := batcher.New(params).Start(); err != nil {
		s.GetLogger().Fatal("error starting batcher", tag.Error(err))
	}
}

func (s *Service) startScanner() {
	params := &scanner.BootstrapParams{
		Config:     *s.config.ScannerCfg,
		TallyScope: s.params.MetricScope,
	}
	if err := scanner.New(s.Resource, params).Start(); err != nil {
		s.GetLogger().Fatal("error starting scanner", tag.Error(err))
	}
}

func (s *Service) startFixerWorkflowWorker() {
	params := &scanner.BootstrapParams{
		Config:     *s.config.ScannerCfg,
		TallyScope: s.params.MetricScope,
	}
	if err := scanner.NewDataCorruptionWorkflowWorker(s.Resource, params).StartDataCorruptionWorkflowWorker(); err != nil {
		s.GetLogger().Fatal("error starting fixer workflow worker", tag.Error(err))
	}
}

func (s *Service) startReplicator() {
	domainReplicationTaskExecutor := domain.NewReplicationTaskExecutor(
		s.Resource.GetDomainManager(),
		s.Resource.GetTimeSource(),
		s.Resource.GetLogger(),
	)
	msgReplicator := replicator.NewReplicator(
		s.GetClusterMetadata(),
		s.GetClientBean(),
		s.GetLogger(),
		s.GetMetricsClient(),
		s.GetHostInfo(),
		s.GetMembershipResolver(),
		s.GetDomainReplicationQueue(),
		domainReplicationTaskExecutor,
		s.config.DomainReplicationMaxRetryDuration(),
	)
	if err := msgReplicator.Start(); err != nil {
		msgReplicator.Stop()
		s.GetLogger().Fatal("fail to start replicator", tag.Error(err))
	}
}

func (s *Service) startIndexer() {
	visibilityIndexer := indexer.NewIndexer(
		s.config.IndexerCfg,
		s.GetMessagingClient(),
		s.params.ESClient,
		s.params.ESConfig,
		s.GetLogger(),
		s.GetMetricsClient(),
	)
	if err := visibilityIndexer.Start(); err != nil {
		visibilityIndexer.Stop()
		s.GetLogger().Fatal("fail to start indexer", tag.Error(err))
	}
}

func (s *Service) startArchiver() {
	bc := &archiver.BootstrapContainer{
		PublicClient:     s.GetSDKClient(),
		MetricsClient:    s.GetMetricsClient(),
		Logger:           s.GetLogger(),
		HistoryV2Manager: s.GetHistoryManager(),
		DomainCache:      s.GetDomainCache(),
		Config:           s.config.ArchiverConfig,
		ArchiverProvider: s.GetArchiverProvider(),
	}
	clientWorker := archiver.NewClientWorker(bc)
	if err := clientWorker.Start(); err != nil {
		clientWorker.Stop()
		s.GetLogger().Fatal("failed to start archiver", tag.Error(err))
	}
}

func (s *Service) startFailoverManager() {
	params := &failovermanager.BootstrapParams{
		Config:        *s.config.failoverManagerCfg,
		ServiceClient: s.params.PublicClient,
		MetricsClient: s.GetMetricsClient(),
		Logger:        s.GetLogger(),
		TallyScope:    s.params.MetricScope,
		ClientBean:    s.GetClientBean(),
	}
	if err := failovermanager.New(params).Start(); err != nil {
		s.Stop()
		s.GetLogger().Fatal("error starting failoverManager", tag.Error(err))
	}
}

func (s *Service) startWorkflowShadower() {
	params := &shadower.BootstrapParams{
		ServiceClient: s.params.PublicClient,
		DomainCache:   s.GetDomainCache(),
		TallyScope:    s.params.MetricScope,
	}
	if err := shadower.New(params).Start(); err != nil {
		s.Stop()
		s.GetLogger().Fatal("error starting workflow shadower", tag.Error(err))
	}
}

func (s *Service) ensureDomainExists(domain string) {
	_, err := s.GetDomainManager().GetDomain(context.Background(), &persistence.GetDomainRequest{Name: domain})
	switch err.(type) {
	case nil:
		// noop
	case *types.EntityNotExistsError:
		s.GetLogger().Info(fmt.Sprintf("domain %s does not exist, attempting to register domain", domain))
		s.registerSystemDomain(domain)
	default:
		s.GetLogger().Fatal("failed to verify if system domain exists", tag.Error(err))
	}
}

func (s *Service) registerSystemDomain(domain string) {

	currentClusterName := s.GetClusterMetadata().GetCurrentClusterName()
	_, err := s.GetDomainManager().CreateDomain(context.Background(), &persistence.CreateDomainRequest{
		Info: &persistence.DomainInfo{
			ID:          getDomainID(domain),
			Name:        domain,
			Description: "Cadence internal system domain",
		},
		Config: &persistence.DomainConfig{
			Retention:  common.SystemDomainRetentionDays,
			EmitMetric: true,
		},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: currentClusterName,
			Clusters:          cluster.GetOrUseDefaultClusters(currentClusterName, nil),
		},
		IsGlobalDomain:  false,
		FailoverVersion: common.EmptyVersion,
	})
	if err != nil {
		if _, ok := err.(*types.DomainAlreadyExistsError); ok {
			return
		}
		s.GetLogger().Fatal("failed to register system domain", tag.Error(err))
	}
}

func getDomainID(domain string) string {
	var domainID string
	switch domain {
	case common.SystemLocalDomainName:
		domainID = common.SystemDomainID
	case common.BatcherLocalDomainName:
		domainID = common.BatcherDomainID
	case common.ShadowerLocalDomainName:
		domainID = common.ShadowerDomainID
	}
	return domainID
}
