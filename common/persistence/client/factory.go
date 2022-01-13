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

package client

import (
	"sync"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	es "github.com/uber/cadence/common/elasticsearch"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/elasticsearch"
	"github.com/uber/cadence/common/persistence/nosql"
	"github.com/uber/cadence/common/persistence/serialization"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/service"
)

type (
	// Factory defines the interface for any implementation that can vend
	// persistence layer objects backed by a datastore. The actual datastore
	// is implementation detail hidden behind this interface
	Factory interface {
		// Close the factory
		Close()
		// NewTaskManager returns a new task manager
		NewTaskManager() (p.TaskManager, error)
		// NewShardManager returns a new shard manager
		NewShardManager() (p.ShardManager, error)
		// NewHistoryManager returns a new history manager
		NewHistoryManager() (p.HistoryManager, error)
		// NewDomainManager returns a new metadata manager
		NewDomainManager() (p.DomainManager, error)
		// NewExecutionManager returns a new execution manager for a given shardID
		NewExecutionManager(shardID int) (p.ExecutionManager, error)
		// NewVisibilityManager returns a new visibility manager
		NewVisibilityManager(params *Params, serviceConfig *service.Config) (p.VisibilityManager, error)
		// NewDomainReplicationQueueManager returns a new queue for domain replication
		NewDomainReplicationQueueManager() (p.QueueManager, error)
		// NewConfigStoreManager returns a new config store manager
		NewConfigStoreManager() (p.ConfigStoreManager, error)
	}
	// DataStoreFactory is a low level interface to be implemented by a datastore
	// Examples of datastores are cassandra, mysql etc
	DataStoreFactory interface {
		// Close closes the factory
		Close()
		// NewTaskStore returns a new task store
		NewTaskStore() (p.TaskStore, error)
		// NewShardStore returns a new shard store
		NewShardStore() (p.ShardStore, error)
		// NewHistoryStore returns a new history store
		NewHistoryStore() (p.HistoryStore, error)
		// NewDomainStore returns a new metadata store
		NewDomainStore() (p.DomainStore, error)
		// NewExecutionStore returns an execution store for given shardID
		NewExecutionStore(shardID int) (p.ExecutionStore, error)
		// NewVisibilityStore returns a new visibility store,
		// TODO We temporarily using sortByCloseTime to determine whether or not ListClosedWorkflowExecutions should
		// be ordering by CloseTime. This will be removed when implementing https://github.com/uber/cadence/issues/3621
		NewVisibilityStore(sortByCloseTime bool) (p.VisibilityStore, error)
		NewQueue(queueType p.QueueType) (p.Queue, error)
		// NewConfigStore returns a new config store
		NewConfigStore() (p.ConfigStore, error)
	}

	// Datastore represents a datastore
	Datastore struct {
		factory   DataStoreFactory
		ratelimit quotas.Limiter
	}
	factoryImpl struct {
		sync.RWMutex
		config        *config.Persistence
		metricsClient metrics.Client
		logger        log.Logger
		datastores    map[storeType]Datastore
		clusterName   string
	}

	storeType int
)

const (
	storeTypeHistory storeType = iota + 1
	storeTypeTask
	storeTypeShard
	storeTypeMetadata
	storeTypeExecution
	storeTypeVisibility
	storeTypeQueue
	storeTypeConfigStore
)

var storeTypes = []storeType{
	storeTypeHistory,
	storeTypeTask,
	storeTypeShard,
	storeTypeMetadata,
	storeTypeExecution,
	storeTypeVisibility,
	storeTypeQueue,
	storeTypeConfigStore,
}

// NewFactory returns an implementation of factory that vends persistence objects based on
// specified configuration. This factory takes as input a config.Persistence object
// which specifies the datastore to be used for a given type of object. This config
// also contains config for individual datastores themselves.
//
// The objects returned by this factory enforce ratelimit and maxconns according to
// given configuration. In addition, all objects will emit metrics automatically
func NewFactory(
	cfg *config.Persistence,
	persistenceMaxQPS dynamicconfig.IntPropertyFn,
	clusterName string,
	metricsClient metrics.Client,
	logger log.Logger,
) Factory {
	factory := &factoryImpl{
		config:        cfg,
		metricsClient: metricsClient,
		logger:        logger,
		clusterName:   clusterName,
	}
	limiters := buildRatelimiters(cfg, persistenceMaxQPS)
	factory.init(clusterName, limiters)
	return factory
}

// NewTaskManager returns a new task manager
func (f *factoryImpl) NewTaskManager() (p.TaskManager, error) {
	ds := f.datastores[storeTypeTask]
	store, err := ds.factory.NewTaskStore()
	if err != nil {
		return nil, err
	}
	result := p.NewTaskManager(store)
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = p.NewTaskPersistenceErrorInjectionClient(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = p.NewTaskPersistenceRateLimitedClient(result, ds.ratelimit, f.logger)
	}
	if f.metricsClient != nil {
		result = p.NewTaskPersistenceMetricsClient(result, f.metricsClient, f.logger, f.config)
	}
	return result, nil
}

// NewShardManager returns a new shard manager
func (f *factoryImpl) NewShardManager() (p.ShardManager, error) {
	ds := f.datastores[storeTypeShard]
	store, err := ds.factory.NewShardStore()
	if err != nil {
		return nil, err
	}
	result := p.NewShardManager(store)
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = p.NewShardPersistenceErrorInjectionClient(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = p.NewShardPersistenceRateLimitedClient(result, ds.ratelimit, f.logger)
	}
	if f.metricsClient != nil {
		result = p.NewShardPersistenceMetricsClient(result, f.metricsClient, f.logger, f.config)
	}
	return result, nil
}

// NewHistoryManager returns a new history manager
func (f *factoryImpl) NewHistoryManager() (p.HistoryManager, error) {
	ds := f.datastores[storeTypeHistory]
	store, err := ds.factory.NewHistoryStore()
	if err != nil {
		return nil, err
	}
	result := p.NewHistoryV2ManagerImpl(store, f.logger, f.config.TransactionSizeLimit)
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = p.NewHistoryPersistenceErrorInjectionClient(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = p.NewHistoryPersistenceRateLimitedClient(result, ds.ratelimit, f.logger)
	}
	if f.metricsClient != nil {
		result = p.NewHistoryPersistenceMetricsClient(result, f.metricsClient, f.logger, f.config)
	}
	return result, nil
}

// NewDomainManager returns a new metadata manager
func (f *factoryImpl) NewDomainManager() (p.DomainManager, error) {
	var err error
	var store p.DomainStore
	ds := f.datastores[storeTypeMetadata]
	store, err = ds.factory.NewDomainStore()
	if err != nil {
		return nil, err
	}
	result := p.NewDomainManagerImpl(store, f.logger)
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = p.NewDomainPersistenceErrorInjectionClient(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = p.NewDomainPersistenceRateLimitedClient(result, ds.ratelimit, f.logger)
	}
	if f.metricsClient != nil {
		result = p.NewDomainPersistenceMetricsClient(result, f.metricsClient, f.logger, f.config)
	}
	return result, nil
}

// NewExecutionManager returns a new execution manager for a given shardID
func (f *factoryImpl) NewExecutionManager(shardID int) (p.ExecutionManager, error) {
	ds := f.datastores[storeTypeExecution]
	store, err := ds.factory.NewExecutionStore(shardID)
	if err != nil {
		return nil, err
	}
	result := p.NewExecutionManagerImpl(store, f.logger)
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = p.NewWorkflowExecutionPersistenceErrorInjectionClient(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = p.NewWorkflowExecutionPersistenceRateLimitedClient(result, ds.ratelimit, f.logger)
	}
	if f.metricsClient != nil {
		result = p.NewWorkflowExecutionPersistenceMetricsClient(result, f.metricsClient, f.logger, f.config)
	}
	return result, nil
}

// NewVisibilityManager returns a new visibility manager
func (f *factoryImpl) NewVisibilityManager(
	params *Params,
	resourceConfig *service.Config,
) (p.VisibilityManager, error) {
	if resourceConfig.EnableReadVisibilityFromES == nil && resourceConfig.AdvancedVisibilityWritingMode == nil {
		// No need to create visibility manager as no read/write needed
		return nil, nil
	}
	var visibilityFromDB, visibilityFromES p.VisibilityManager
	var err error
	if params.PersistenceConfig.VisibilityStore != "" {
		visibilityFromDB, err = f.newDBVisibilityManager(resourceConfig)
		if err != nil {
			return nil, err
		}
	}
	if params.PersistenceConfig.AdvancedVisibilityStore != "" {
		visibilityIndexName := params.ESConfig.Indices[common.VisibilityAppName]
		visibilityProducer, err := params.MessagingClient.NewProducer(common.VisibilityAppName)
		if err != nil {
			f.logger.Fatal("Creating visibility producer failed", tag.Error(err))
		}
		visibilityFromES = newESVisibilityManager(
			visibilityIndexName, params.ESClient, resourceConfig, visibilityProducer, params.MetricsClient, f.logger,
		)
	}
	return p.NewVisibilityDualManager(
		visibilityFromDB,
		visibilityFromES,
		resourceConfig.EnableReadVisibilityFromES,
		resourceConfig.AdvancedVisibilityWritingMode,
		f.logger,
	), nil
}

// NewESVisibilityManager create a visibility manager for ElasticSearch
// In history, it only needs kafka producer for writing data;
// In frontend, it only needs ES client and related config for reading data
func newESVisibilityManager(
	indexName string,
	esClient es.GenericClient,
	visibilityConfig *service.Config,
	producer messaging.Producer,
	metricsClient metrics.Client,
	log log.Logger,
) p.VisibilityManager {

	visibilityFromESStore := elasticsearch.NewElasticSearchVisibilityStore(esClient, indexName, producer, visibilityConfig, log)
	visibilityFromES := p.NewVisibilityManagerImpl(visibilityFromESStore, log)

	// wrap with rate limiter
	if visibilityConfig.PersistenceMaxQPS != nil && visibilityConfig.PersistenceMaxQPS() != 0 {
		esRateLimiter := quotas.NewDynamicRateLimiter(
			func() float64 {
				return float64(visibilityConfig.PersistenceMaxQPS())
			},
		)
		visibilityFromES = p.NewVisibilityPersistenceRateLimitedClient(visibilityFromES, esRateLimiter, log)
	}
	if metricsClient != nil {
		// wrap with metrics
		visibilityFromES = elasticsearch.NewVisibilityMetricsClient(visibilityFromES, metricsClient, log)
	}

	return visibilityFromES
}

func (f *factoryImpl) newDBVisibilityManager(
	visibilityConfig *service.Config,
) (p.VisibilityManager, error) {
	enableReadFromClosedExecutionV2 := false
	if visibilityConfig.EnableReadDBVisibilityFromClosedExecutionV2 != nil {
		enableReadFromClosedExecutionV2 = visibilityConfig.EnableReadDBVisibilityFromClosedExecutionV2()
	} else {
		f.logger.Warn("missing visibility and EnableReadFromClosedExecutionV2 config", tag.Value(visibilityConfig))
	}

	ds := f.datastores[storeTypeVisibility]
	store, err := ds.factory.NewVisibilityStore(enableReadFromClosedExecutionV2)
	if err != nil {
		return nil, err
	}
	result := p.NewVisibilityManagerImpl(store, f.logger)
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = p.NewVisibilityPersistenceErrorInjectionClient(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = p.NewVisibilityPersistenceRateLimitedClient(result, ds.ratelimit, f.logger)
	}
	if visibilityConfig.EnableDBVisibilitySampling != nil && visibilityConfig.EnableDBVisibilitySampling() {
		result = p.NewVisibilitySamplingClient(result, &p.SamplingConfig{
			VisibilityClosedMaxQPS: visibilityConfig.WriteDBVisibilityClosedMaxQPS,
			VisibilityListMaxQPS:   visibilityConfig.DBVisibilityListMaxQPS,
			VisibilityOpenMaxQPS:   visibilityConfig.WriteDBVisibilityOpenMaxQPS,
		}, f.metricsClient, f.logger)
	}
	if f.metricsClient != nil {
		result = p.NewVisibilityPersistenceMetricsClient(result, f.metricsClient, f.logger, f.config)
	}

	return result, nil
}

func (f *factoryImpl) NewDomainReplicationQueueManager() (p.QueueManager, error) {
	ds := f.datastores[storeTypeQueue]
	store, err := ds.factory.NewQueue(p.DomainReplicationQueueType)
	if err != nil {
		return nil, err
	}
	result := p.NewQueueManager(store)
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = p.NewQueuePersistenceErrorInjectionClient(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = p.NewQueuePersistenceRateLimitedClient(result, ds.ratelimit, f.logger)
	}
	if f.metricsClient != nil {
		result = p.NewQueuePersistenceMetricsClient(result, f.metricsClient, f.logger, f.config)
	}

	return result, nil
}

func (f *factoryImpl) NewConfigStoreManager() (p.ConfigStoreManager, error) {
	ds := f.datastores[storeTypeConfigStore]
	store, err := ds.factory.NewConfigStore()
	if err != nil {
		return nil, err
	}
	result := p.NewConfigStoreManagerImpl(store, f.logger)
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = p.NewConfigStoreErrorInjectionPersistenceClient(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = p.NewConfigStorePersistenceRateLimitedClient(result, ds.ratelimit, f.logger)
	}
	if f.metricsClient != nil {
		result = p.NewConfigStorePersistenceMetricsClient(result, f.metricsClient, f.logger, f.config)
	}

	return result, nil
}

// Close closes this factory
func (f *factoryImpl) Close() {
	ds := f.datastores[storeTypeExecution]
	ds.factory.Close()
}

func (f *factoryImpl) init(clusterName string, limiters map[string]quotas.Limiter) {
	f.datastores = make(map[storeType]Datastore, len(storeTypes))
	defaultCfg := f.config.DataStores[f.config.DefaultStore]
	if defaultCfg.Cassandra != nil {
		f.logger.Warn("Cassandra config is deprecated, please use NoSQL with pluginName of cassandra.")
	}
	defaultDataStore := Datastore{ratelimit: limiters[f.config.DefaultStore]}
	switch {
	case defaultCfg.NoSQL != nil:
		defaultDataStore.factory = nosql.NewFactory(*defaultCfg.NoSQL, clusterName, f.logger)
	case defaultCfg.SQL != nil:
		if defaultCfg.SQL.EncodingType == "" {
			defaultCfg.SQL.EncodingType = string(common.EncodingTypeThriftRW)
		}
		if len(defaultCfg.SQL.DecodingTypes) == 0 {
			defaultCfg.SQL.DecodingTypes = []string{
				string(common.EncodingTypeThriftRW),
			}
		}
		var decodingTypes []common.EncodingType
		for _, dt := range defaultCfg.SQL.DecodingTypes {
			decodingTypes = append(decodingTypes, common.EncodingType(dt))
		}
		defaultDataStore.factory = sql.NewFactory(
			*defaultCfg.SQL,
			clusterName,
			f.logger,
			getSQLParser(f.logger, common.EncodingType(defaultCfg.SQL.EncodingType), decodingTypes...))
	default:
		f.logger.Fatal("invalid config: one of nosql or sql params must be specified for defaultDataStore")
	}

	for _, st := range storeTypes {
		if st != storeTypeVisibility {
			f.datastores[st] = defaultDataStore
		}
	}

	visibilityCfg, ok := f.config.DataStores[f.config.VisibilityStore]
	if !ok {
		f.logger.Info("no visibilityStore is configured, will use advancedVisibilityStore")
		// NOTE: f.datastores[storeTypeVisibility] will be nil
		return
	}

	if visibilityCfg.Cassandra != nil {
		f.logger.Warn("Cassandra config is deprecated, please use NoSQL with pluginName of cassandra.")
	}
	visibilityDataStore := Datastore{ratelimit: limiters[f.config.VisibilityStore]}
	switch {
	case visibilityCfg.NoSQL != nil:
		visibilityDataStore.factory = nosql.NewFactory(*visibilityCfg.NoSQL, clusterName, f.logger)
	case visibilityCfg.SQL != nil:
		var decodingTypes []common.EncodingType
		for _, dt := range visibilityCfg.SQL.DecodingTypes {
			decodingTypes = append(decodingTypes, common.EncodingType(dt))
		}
		visibilityDataStore.factory = sql.NewFactory(
			*visibilityCfg.SQL,
			clusterName,
			f.logger,
			getSQLParser(f.logger, common.EncodingType(visibilityCfg.SQL.EncodingType), decodingTypes...))
	default:
		f.logger.Fatal("invalid config: one of nosql or sql params must be specified for visibilityStore")
	}

	f.datastores[storeTypeVisibility] = visibilityDataStore
}

func getSQLParser(logger log.Logger, encodingType common.EncodingType, decodingTypes ...common.EncodingType) serialization.Parser {
	parser, err := serialization.NewParser(encodingType, decodingTypes...)
	if err != nil {
		logger.Fatal("failed to construct sql parser", tag.Error(err))
	}
	return parser
}

func buildRatelimiters(cfg *config.Persistence, maxQPS dynamicconfig.IntPropertyFn) map[string]quotas.Limiter {
	result := make(map[string]quotas.Limiter, len(cfg.DataStores))
	for dsName := range cfg.DataStores {
		if maxQPS != nil && maxQPS() > 0 {
			result[dsName] = quotas.NewDynamicRateLimiter(func() float64 { return float64(maxQPS()) })
		}
	}
	return result
}
