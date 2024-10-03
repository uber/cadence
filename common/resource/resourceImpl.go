// Copyright (c) 2019 Uber Technologies, Inc.
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

package resource

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/client/wrappers/retryable"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/asyncworkflow/queue"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/dynamicconfig/configstore"
	csc "github.com/uber/cadence/common/dynamicconfig/configstore/config"
	"github.com/uber/cadence/common/isolationgroup"
	"github.com/uber/cadence/common/isolationgroup/defaultisolationgroupstate"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/partition"
	"github.com/uber/cadence/common/persistence"
	persistenceClient "github.com/uber/cadence/common/persistence/client"
	"github.com/uber/cadence/common/quotas/global/rpc"
	"github.com/uber/cadence/common/quotas/permember"
	"github.com/uber/cadence/common/service"
)

type (

	// VisibilityManagerInitializer is the function each service should implement
	// for visibility manager initialization
	VisibilityManagerInitializer func(
		persistenceBean persistenceClient.Bean,
		logger log.Logger,
	) (persistence.VisibilityManager, error)

	// Impl contains all common resources shared across frontend / matching / history / worker
	Impl struct {
		status int32

		// static infos
		numShards       int
		serviceName     string
		hostInfo        membership.HostInfo
		metricsScope    tally.Scope
		clusterMetadata cluster.Metadata

		// other common resources

		domainCache             cache.DomainCache
		domainMetricsScopeCache cache.DomainMetricsScopeCache
		timeSource              clock.TimeSource
		payloadSerializer       persistence.PayloadSerializer
		metricsClient           metrics.Client
		messagingClient         messaging.Client
		blobstoreClient         blobstore.Client
		archivalMetadata        archiver.ArchivalMetadata
		archiverProvider        provider.ArchiverProvider
		domainReplicationQueue  domain.ReplicationQueue

		// membership infos

		membershipResolver membership.Resolver

		// internal services clients

		sdkClient         workflowserviceclient.Interface
		frontendRawClient frontend.Client
		frontendClient    frontend.Client
		matchingRawClient matching.Client
		matchingClient    matching.Client
		historyRawClient  history.Client
		historyClient     history.Client
		clientBean        client.Bean

		// persistence clients
		persistenceBean persistenceClient.Bean

		// hostName
		hostName string

		// loggers
		logger          log.Logger
		throttledLogger log.Logger

		// for registering handlers
		dispatcher *yarpc.Dispatcher

		// internal vars

		pprofInitializer       common.PProfInitializer
		runtimeMetricsReporter *metrics.RuntimeMetricsReporter
		rpcFactory             common.RPCFactory

		isolationGroups           isolationgroup.State
		isolationGroupConfigStore configstore.Client
		partitioner               partition.Partitioner

		asyncWorkflowQueueProvider queue.Provider

		ratelimiterAggregatorClient rpc.Client
	}
)

var _ Resource = (*Impl)(nil)

// New create a new resource containing common dependencies
func New(
	params *Params,
	serviceName string,
	serviceConfig *service.Config,
) (impl *Impl, retError error) {

	hostname := params.HostName

	logger := params.Logger
	throttledLogger := loggerimpl.NewThrottledLogger(logger, serviceConfig.ThrottledLoggerMaxRPS)

	numShards := params.PersistenceConfig.NumHistoryShards
	dispatcher := params.RPCFactory.GetDispatcher()
	membershipResolver := params.MembershipResolver

	ensureGetAllIsolationGroupsFnIsSet(params)

	dynamicCollection := dynamicconfig.NewCollection(
		params.DynamicConfig,
		logger,
		dynamicconfig.ClusterNameFilter(params.ClusterMetadata.GetCurrentClusterName()),
	)
	clientBean, err := client.NewClientBean(
		client.NewRPCClientFactory(
			params.RPCFactory,
			membershipResolver,
			params.MetricsClient,
			dynamicCollection,
			numShards,
			logger,
		),
		params.RPCFactory.GetDispatcher(),
		params.ClusterMetadata,
	)
	if err != nil {
		return nil, err
	}

	persistenceBean, err := persistenceClient.NewBeanFromFactory(persistenceClient.NewFactory(
		&params.PersistenceConfig,
		func() float64 {
			return permember.PerMember(
				serviceName,
				float64(serviceConfig.PersistenceGlobalMaxQPS()),
				float64(serviceConfig.PersistenceMaxQPS()),
				membershipResolver,
			)
		},
		params.ClusterMetadata.GetCurrentClusterName(),
		params.MetricsClient,
		logger,
		persistence.NewDynamicConfiguration(dynamicCollection),
	), &persistenceClient.Params{
		PersistenceConfig: params.PersistenceConfig,
		MetricsClient:     params.MetricsClient,
		MessagingClient:   params.MessagingClient,
		ESClient:          params.ESClient,
		ESConfig:          params.ESConfig,
		PinotConfig:       params.PinotConfig,
		PinotClient:       params.PinotClient,
		OSClient:          params.OSClient,
		OSConfig:          params.OSConfig,
	}, serviceConfig)
	if err != nil {
		return nil, err
	}

	domainCache := cache.NewDomainCache(
		persistenceBean.GetDomainManager(),
		params.ClusterMetadata,
		params.MetricsClient,
		logger,
		cache.WithTimeSource(params.TimeSource),
	)

	domainMetricsScopeCache := cache.NewDomainMetricsScopeCache()
	domainReplicationQueue := domain.NewReplicationQueue(
		persistenceBean.GetDomainReplicationQueueManager(),
		params.ClusterMetadata.GetCurrentClusterName(),
		params.MetricsClient,
		logger,
	)

	frontendRawClient := clientBean.GetFrontendClient()
	frontendClient := retryable.NewFrontendClient(
		frontendRawClient,
		common.CreateFrontendServiceRetryPolicy(),
		serviceConfig.IsErrorRetryableFunction,
	)

	matchingRawClient, err := clientBean.GetMatchingClient(domainCache.GetDomainName)
	if err != nil {
		return nil, err
	}
	matchingClient := retryable.NewMatchingClient(
		matchingRawClient,
		common.CreateMatchingServiceRetryPolicy(),
		serviceConfig.IsErrorRetryableFunction,
	)

	var historyRawClient history.Client
	if params.HistoryClientFn != nil {
		logger.Debug("Using history client from HistoryClientFn")
		historyRawClient = params.HistoryClientFn()
	} else {
		logger.Debug("Using history client from bean")
		historyRawClient = clientBean.GetHistoryClient()
	}
	historyClient := retryable.NewHistoryClient(
		historyRawClient,
		common.CreateHistoryServiceRetryPolicy(),
		serviceConfig.IsErrorRetryableFunction,
	)

	historyArchiverBootstrapContainer := &archiver.HistoryBootstrapContainer{
		HistoryV2Manager: persistenceBean.GetHistoryManager(),
		Logger:           logger,
		MetricsClient:    params.MetricsClient,
		ClusterMetadata:  params.ClusterMetadata,
		DomainCache:      domainCache,
	}
	visibilityArchiverBootstrapContainer := &archiver.VisibilityBootstrapContainer{
		Logger:          logger,
		MetricsClient:   params.MetricsClient,
		ClusterMetadata: params.ClusterMetadata,
		DomainCache:     domainCache,
	}
	if err := params.ArchiverProvider.RegisterBootstrapContainer(
		serviceName,
		historyArchiverBootstrapContainer,
		visibilityArchiverBootstrapContainer,
	); err != nil {
		return nil, err
	}

	isolationGroupStore := createConfigStoreOrDefault(params, dynamicCollection)

	isolationGroupState, err := ensureIsolationGroupStateHandlerOrDefault(
		params,
		dynamicCollection,
		domainCache,
		isolationGroupStore,
	)
	if err != nil {
		return nil, err
	}
	partitioner := ensurePartitionerOrDefault(params, isolationGroupState)

	ratelimiterAggs := rpc.New(
		historyRawClient, // no retries, will retry internally if needed
		clientBean.GetHistoryPeers(),
		logger,
		params.MetricsClient,
	)

	impl = &Impl{
		status: common.DaemonStatusInitialized,

		// static infos

		numShards:       numShards,
		serviceName:     params.Name,
		metricsScope:    params.MetricScope,
		clusterMetadata: params.ClusterMetadata,

		// other common resources

		domainCache:             domainCache,
		domainMetricsScopeCache: domainMetricsScopeCache,
		timeSource:              clock.NewRealTimeSource(),
		payloadSerializer:       persistence.NewPayloadSerializer(),
		metricsClient:           params.MetricsClient,
		messagingClient:         params.MessagingClient,
		blobstoreClient:         params.BlobstoreClient,
		archivalMetadata:        params.ArchivalMetadata,
		archiverProvider:        params.ArchiverProvider,
		domainReplicationQueue:  domainReplicationQueue,

		// membership infos
		membershipResolver: membershipResolver,

		// internal services clients

		sdkClient:         params.PublicClient,
		frontendRawClient: frontendRawClient,
		frontendClient:    frontendClient,
		matchingRawClient: matchingRawClient,
		matchingClient:    matchingClient,
		historyRawClient:  historyRawClient,
		historyClient:     historyClient,
		clientBean:        clientBean,

		// persistence clients
		persistenceBean: persistenceBean,

		// hostname
		hostName: hostname,

		// loggers

		logger:          logger,
		throttledLogger: throttledLogger,

		// for registering handlers
		dispatcher: dispatcher,

		// internal vars
		pprofInitializer: params.PProfInitializer,
		runtimeMetricsReporter: metrics.NewRuntimeMetricsReporter(
			params.MetricScope,
			time.Minute,
			logger,
			params.InstanceID,
		),
		rpcFactory:                params.RPCFactory,
		isolationGroups:           isolationGroupState,
		isolationGroupConfigStore: isolationGroupStore, // can be nil where persistence is not available
		partitioner:               partitioner,

		asyncWorkflowQueueProvider: params.AsyncWorkflowQueueProvider,

		ratelimiterAggregatorClient: ratelimiterAggs,
	}
	return impl, nil
}

// Start all resources
func (h *Impl) Start() {

	if !atomic.CompareAndSwapInt32(
		&h.status,
		common.DaemonStatusInitialized,
		common.DaemonStatusStarted,
	) {
		return
	}

	h.metricsScope.Counter(metrics.RestartCount).Inc(1)
	h.runtimeMetricsReporter.Start()

	if err := h.pprofInitializer.Start(); err != nil {
		h.logger.WithTags(tag.Error(err)).Fatal("fail to start PProf")
	}
	if err := h.dispatcher.Start(); err != nil {
		h.logger.WithTags(tag.Error(err)).Fatal("fail to start dispatcher")
	}
	h.membershipResolver.Start()
	h.domainCache.Start()
	h.domainMetricsScopeCache.Start()

	hostInfo, err := h.membershipResolver.WhoAmI()
	if err != nil {
		h.logger.WithTags(tag.Error(err)).Fatal("fail to get host info from membership monitor")
	}
	h.hostInfo = hostInfo

	if h.isolationGroupConfigStore != nil {
		h.isolationGroupConfigStore.Start()
	}
	// The service is now started up
	h.logger.Info("service started")
	// seed the random generator once for this service
	rand.Seed(time.Now().UTC().UnixNano())
}

// Stop stops all resources
func (h *Impl) Stop() {

	if !atomic.CompareAndSwapInt32(
		&h.status,
		common.DaemonStatusStarted,
		common.DaemonStatusStopped,
	) {
		return
	}

	h.domainCache.Stop()
	h.domainMetricsScopeCache.Stop()
	h.membershipResolver.Stop()
	if err := h.dispatcher.Stop(); err != nil {
		h.logger.WithTags(tag.Error(err)).Error("failed to stop dispatcher")
	}
	h.runtimeMetricsReporter.Stop()
	h.persistenceBean.Close()
	if h.isolationGroupConfigStore != nil {
		h.isolationGroupConfigStore.Stop()
	}
	h.isolationGroups.Stop()
}

// GetServiceName return service name
func (h *Impl) GetServiceName() string {
	return h.serviceName
}

// GetHostInfo return host info
func (h *Impl) GetHostInfo() membership.HostInfo {
	return h.hostInfo
}

// GetClusterMetadata return cluster metadata
func (h *Impl) GetClusterMetadata() cluster.Metadata {
	return h.clusterMetadata
}

// other common resources

// GetDomainCache return domain cache
func (h *Impl) GetDomainCache() cache.DomainCache {
	return h.domainCache
}

// GetDomainMetricsScopeCache return domainMetricsScope cache
func (h *Impl) GetDomainMetricsScopeCache() cache.DomainMetricsScopeCache {
	return h.domainMetricsScopeCache
}

// GetTimeSource return time source
func (h *Impl) GetTimeSource() clock.TimeSource {
	return h.timeSource
}

// GetPayloadSerializer return binary payload serializer
func (h *Impl) GetPayloadSerializer() persistence.PayloadSerializer {
	return h.payloadSerializer
}

// GetMetricsClient return metrics client
func (h *Impl) GetMetricsClient() metrics.Client {
	return h.metricsClient
}

// GetMessagingClient return messaging client
func (h *Impl) GetMessagingClient() messaging.Client {
	return h.messagingClient
}

// GetBlobstoreClient returns blobstore client
func (h *Impl) GetBlobstoreClient() blobstore.Client {
	return h.blobstoreClient
}

// GetArchivalMetadata return archival metadata
func (h *Impl) GetArchivalMetadata() archiver.ArchivalMetadata {
	return h.archivalMetadata
}

// GetArchiverProvider return archival provider
func (h *Impl) GetArchiverProvider() provider.ArchiverProvider {
	return h.archiverProvider
}

// GetDomainReplicationQueue return domain replication queue
func (h *Impl) GetDomainReplicationQueue() domain.ReplicationQueue {
	return h.domainReplicationQueue
}

// GetMembershipResolver return the membership resolver
func (h *Impl) GetMembershipResolver() membership.Resolver {
	return h.membershipResolver
}

// internal services clients

// GetSDKClient return sdk client
func (h *Impl) GetSDKClient() workflowserviceclient.Interface {
	return h.sdkClient
}

// GetFrontendRawClient return frontend client without retry policy
func (h *Impl) GetFrontendRawClient() frontend.Client {
	return h.frontendRawClient
}

// GetFrontendClient return frontend client with retry policy
func (h *Impl) GetFrontendClient() frontend.Client {
	return h.frontendClient
}

// GetMatchingRawClient return matching client without retry policy
func (h *Impl) GetMatchingRawClient() matching.Client {
	return h.matchingRawClient
}

// GetMatchingClient return matching client with retry policy
func (h *Impl) GetMatchingClient() matching.Client {
	return h.matchingClient
}

// GetHistoryRawClient return history client without retry policy
func (h *Impl) GetHistoryRawClient() history.Client {
	return h.historyRawClient
}

// GetHistoryClient return history client with retry policy
func (h *Impl) GetHistoryClient() history.Client {
	return h.historyClient
}

func (h *Impl) GetRatelimiterAggregatorsClient() rpc.Client {
	return h.ratelimiterAggregatorClient
}

// GetRemoteAdminClient return remote admin client for given cluster name
func (h *Impl) GetRemoteAdminClient(
	cluster string,
) admin.Client {

	return h.clientBean.GetRemoteAdminClient(cluster)
}

// GetRemoteFrontendClient return remote frontend client for given cluster name
func (h *Impl) GetRemoteFrontendClient(
	cluster string,
) frontend.Client {

	return h.clientBean.GetRemoteFrontendClient(cluster)
}

// GetClientBean return RPC client bean
func (h *Impl) GetClientBean() client.Bean {
	return h.clientBean
}

// persistence clients

// GetMetadataManager return metadata manager
func (h *Impl) GetDomainManager() persistence.DomainManager {
	return h.persistenceBean.GetDomainManager()
}

// GetTaskManager return task manager
func (h *Impl) GetTaskManager() persistence.TaskManager {
	return h.persistenceBean.GetTaskManager()
}

// GetVisibilityManager return visibility manager
func (h *Impl) GetVisibilityManager() persistence.VisibilityManager {
	return h.persistenceBean.GetVisibilityManager()
}

// GetShardManager return shard manager
func (h *Impl) GetShardManager() persistence.ShardManager {
	return h.persistenceBean.GetShardManager()
}

// GetHistoryManager return history manager
func (h *Impl) GetHistoryManager() persistence.HistoryManager {
	return h.persistenceBean.GetHistoryManager()
}

// GetExecutionManager return execution manager for given shard ID
func (h *Impl) GetExecutionManager(
	shardID int,
) (persistence.ExecutionManager, error) {

	return h.persistenceBean.GetExecutionManager(shardID)
}

// GetPersistenceBean return persistence bean
func (h *Impl) GetPersistenceBean() persistenceClient.Bean {
	return h.persistenceBean
}

func (h *Impl) GetHostName() string {
	return h.hostName
}

// loggers

// GetLogger return logger
func (h *Impl) GetLogger() log.Logger {
	return h.logger
}

// GetThrottledLogger return throttled logger
func (h *Impl) GetThrottledLogger() log.Logger {
	return h.throttledLogger
}

// GetDispatcher return YARPC dispatcher, used for registering handlers
func (h *Impl) GetDispatcher() *yarpc.Dispatcher {
	return h.dispatcher
}

// GetIsolationGroupState returns the isolationGroupState
func (h *Impl) GetIsolationGroupState() isolationgroup.State {
	return h.isolationGroups
}

// GetPartitioner returns the partitioner
func (h *Impl) GetPartitioner() partition.Partitioner {
	return h.partitioner
}

// GetIsolationGroupStore returns the isolation group configuration store or nil
func (h *Impl) GetIsolationGroupStore() configstore.Client {
	return h.isolationGroupConfigStore
}

// GetAsyncWorkflowQueueProvider returns the async workflow queue provider
func (h *Impl) GetAsyncWorkflowQueueProvider() queue.Provider {
	return h.asyncWorkflowQueueProvider
}

// due to the config store being only available for some
// persistence layers, *both* the configStoreClient and IsolationGroupState
// will be optionally available
func createConfigStoreOrDefault(
	params *Params,
	dc *dynamicconfig.Collection,
) configstore.Client {

	if params.IsolationGroupStore != nil {
		return params.IsolationGroupStore
	}
	cscConfig := &csc.ClientConfig{
		PollInterval:        dc.GetDurationProperty(dynamicconfig.IsolationGroupStateRefreshInterval)(),
		UpdateRetryAttempts: dc.GetIntProperty(dynamicconfig.IsolationGroupStateUpdateRetryAttempts)(),
		FetchTimeout:        dc.GetDurationProperty(dynamicconfig.IsolationGroupStateFetchTimeout)(),
		UpdateTimeout:       dc.GetDurationProperty(dynamicconfig.IsolationGroupStateUpdateTimeout)(),
	}
	cfgStoreClient, err := configstore.NewConfigStoreClient(cscConfig, &params.PersistenceConfig, params.Logger, persistence.GlobalIsolationGroupConfig)
	if err != nil {
		// not possible to create the client under some persistence configurations, so this is expected
		params.Logger.Warn("not instantiating Isolation group config store, this feature will not be enabled", tag.Error(err))
		return nil
	}
	return cfgStoreClient
}

// Use the provided IsolationGroupStateHandler or the default one
// due to the config store being only available for some
// persistence layers, *both* the configStoreClient and IsolationGroupState
// will be optionally available
func ensureIsolationGroupStateHandlerOrDefault(
	params *Params,
	dc *dynamicconfig.Collection,
	domainCache cache.DomainCache,
	isolationGroupStore dynamicconfig.Client,
) (isolationgroup.State, error) {

	if params.IsolationGroupState != nil {
		return params.IsolationGroupState, nil
	}

	return defaultisolationgroupstate.NewDefaultIsolationGroupStateWatcherWithConfigStoreClient(
		params.Logger,
		dc,
		domainCache,
		isolationGroupStore,
		params.MetricsClient,
		params.GetIsolationGroups,
	)
}

// Use the provided partitioner or the default one
func ensurePartitionerOrDefault(params *Params, state isolationgroup.State) partition.Partitioner {
	if params.Partitioner != nil {
		return params.Partitioner
	}
	return partition.NewDefaultPartitioner(params.Logger, state)
}

func ensureGetAllIsolationGroupsFnIsSet(params *Params) {
	if params.GetIsolationGroups == nil {
		params.GetIsolationGroups = func() []string { return []string{} }
	}
}
