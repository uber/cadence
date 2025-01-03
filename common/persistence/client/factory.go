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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination factory_mock.go

package client

import (
	"sync"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/config"
	es "github.com/uber/cadence/common/elasticsearch"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/elasticsearch"
	"github.com/uber/cadence/common/persistence/nosql"
	pinotVisibility "github.com/uber/cadence/common/persistence/pinot"
	"github.com/uber/cadence/common/persistence/serialization"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/persistence/wrappers/errorinjectors"
	"github.com/uber/cadence/common/persistence/wrappers/metered"
	"github.com/uber/cadence/common/persistence/wrappers/ratelimited"
	"github.com/uber/cadence/common/persistence/wrappers/sampled"
	pnt "github.com/uber/cadence/common/pinot"
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
		dc            *p.DynamicConfiguration
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
	persistenceMaxQPS quotas.RPSFunc,
	clusterName string,
	metricsClient metrics.Client,
	logger log.Logger,
	dc *p.DynamicConfiguration,
) Factory {
	factory := &factoryImpl{
		config:        cfg,
		metricsClient: metricsClient,
		logger:        logger,
		clusterName:   clusterName,
		dc:            dc,
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
		result = errorinjectors.NewTaskManager(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = ratelimited.NewTaskManager(result, ds.ratelimit)
	}
	if f.metricsClient != nil {
		result = metered.NewTaskManager(result, f.metricsClient, f.logger, f.config)
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
		result = errorinjectors.NewShardManager(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = ratelimited.NewShardManager(result, ds.ratelimit)
	}
	if f.metricsClient != nil {
		result = metered.NewShardManager(result, f.metricsClient, f.logger, f.config)
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
	result := p.NewHistoryV2ManagerImpl(store, f.logger, p.NewPayloadSerializer(), codec.NewThriftRWEncoder(), f.config.TransactionSizeLimit)
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = errorinjectors.NewHistoryManager(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = ratelimited.NewHistoryManager(result, ds.ratelimit)
	}
	if f.metricsClient != nil {
		result = metered.NewHistoryManager(result, f.metricsClient, f.logger, f.config)
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
	result := p.NewDomainManagerImpl(store, f.logger, p.NewPayloadSerializer())
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = errorinjectors.NewDomainManager(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = ratelimited.NewDomainManager(result, ds.ratelimit)
	}
	if f.metricsClient != nil {
		result = metered.NewDomainManager(result, f.metricsClient, f.logger, f.config)
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
	result := p.NewExecutionManagerImpl(store, f.logger, p.NewPayloadSerializer())
	if errorRate := f.config.ErrorInjectionRate(); errorRate != 0 {
		result = errorinjectors.NewExecutionManager(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = ratelimited.NewExecutionManager(result, ds.ratelimit)
	}
	if f.metricsClient != nil {
		result = metered.NewExecutionManager(result, f.metricsClient, f.logger, f.config, f.dc.PersistenceSampleLoggingRate, f.dc.EnableShardIDMetrics)
	}
	return result, nil
}

// NewVisibilityManager returns a new visibility manager
func (f *factoryImpl) NewVisibilityManager(
	params *Params,
	resourceConfig *service.Config,
) (p.VisibilityManager, error) {
	if resourceConfig.EnableReadVisibilityFromES == nil && resourceConfig.EnableReadVisibilityFromPinot == nil && resourceConfig.AdvancedVisibilityWritingMode == nil {
		// No need to create visibility manager as no read/write needed
		return nil, nil
	}
	var visibilityFromDB, visibilityFromES, visibilityFromPinot, visibilityFromOS p.VisibilityManager
	var err error
	if params.PersistenceConfig.VisibilityStore != "" {
		visibilityFromDB, err = f.newDBVisibilityManager(resourceConfig)
		if err != nil {
			return nil, err
		}
	}

	switch params.PersistenceConfig.AdvancedVisibilityStore {
	case common.PinotVisibilityStoreName:
		visibilityFromPinot, err = setupPinotVisibilityManager(params, resourceConfig, f.logger)
		if err != nil {
			f.logger.Fatal("Creating Pinot advanced visibility manager failed", tag.Error(err))
		}

		if params.PinotConfig.Migration.Enabled {
			visibilityFromES, err = setupESVisibilityManager(params, resourceConfig, f.logger)
			if err != nil {
				f.logger.Fatal("Creating ES advanced visibility manager failed", tag.Error(err))
			}

			return p.NewVisibilityTripleManager(
				visibilityFromDB,
				visibilityFromPinot,
				visibilityFromES,
				resourceConfig.EnableReadVisibilityFromPinot,
				resourceConfig.EnableReadVisibilityFromES,
				resourceConfig.AdvancedVisibilityMigrationWritingMode,
				resourceConfig.EnableLogCustomerQueryParameter,
				resourceConfig.EnableVisibilityDoubleRead,
				f.logger,
			), nil
		}

		return p.NewVisibilityDualManager(
			visibilityFromDB,
			visibilityFromPinot,
			resourceConfig.EnableReadVisibilityFromPinot,
			resourceConfig.AdvancedVisibilityWritingMode,
			f.logger,
		), nil
	case common.OSVisibilityStoreName:
		visibilityFromOS, err = setupOSVisibilityManager(params, resourceConfig, f.logger)
		if err != nil {
			f.logger.Fatal("Creating OS advanced visibility manager failed", tag.Error(err))
		}
		if params.OSConfig.Migration.Enabled {
			// this should be always true when using os-visibility
			visibilityFromES, err = setupESVisibilityManager(params, resourceConfig, f.logger)
			if err != nil {
				f.logger.Fatal("Creating ES advanced visibility manager failed", tag.Error(err))
			}
			return p.NewVisibilityTripleManager(
				visibilityFromDB,
				visibilityFromOS,
				visibilityFromES,
				resourceConfig.EnableReadVisibilityFromES, // Didn't add new config for EnableReadVisibilityFromOS since we will use es-visibility and version: "os2" when migration is done
				resourceConfig.EnableReadVisibilityFromOS, // this controls read from source(ES), will be the primary read source
				resourceConfig.AdvancedVisibilityMigrationWritingMode,
				resourceConfig.EnableLogCustomerQueryParameter,
				resourceConfig.EnableVisibilityDoubleRead,
				f.logger,
			), nil
		}
		return p.NewVisibilityDualManager(
			visibilityFromDB,
			visibilityFromOS,
			resourceConfig.EnableReadVisibilityFromOS, //Didn't add new config for EnableReadVisibilityFromOS since we will use es-visibility and version: "os2" when migration is done
			resourceConfig.AdvancedVisibilityWritingMode,
			f.logger,
		), nil
	case common.ESVisibilityStoreName:
		visibilityFromES, err = setupESVisibilityManager(params, resourceConfig, f.logger)
		if err != nil {
			f.logger.Fatal("Creating advanced visibility manager failed", tag.Error(err))
		}
		return p.NewVisibilityDualManager(
			visibilityFromDB,
			visibilityFromES,
			resourceConfig.EnableReadVisibilityFromES,
			resourceConfig.AdvancedVisibilityWritingMode,
			f.logger,
		), nil
	default:
		return p.NewVisibilityDualManager(
			visibilityFromDB,
			visibilityFromES,
			resourceConfig.EnableReadVisibilityFromES,
			resourceConfig.AdvancedVisibilityWritingMode,
			f.logger,
		), nil
	}
}

// NewESVisibilityManager create a visibility manager for ElasticSearch
// In history, it only needs kafka producer for writing data;
// In frontend, it only needs ES client and related config for reading data
func newPinotVisibilityManager(
	pinotClient pnt.GenericClient,
	visibilityConfig *service.Config,
	producer messaging.Producer,
	metricsClient metrics.Client,
	log log.Logger,
) p.VisibilityManager {
	visibilityFromPinotStore := pinotVisibility.NewPinotVisibilityStore(pinotClient, visibilityConfig, producer, log)
	visibilityFromPinot := p.NewVisibilityManagerImpl(visibilityFromPinotStore, log)

	// wrap with rate limiter
	if visibilityConfig.PersistenceMaxQPS != nil && visibilityConfig.PersistenceMaxQPS() != 0 {
		pinotRateLimiter := quotas.NewDynamicRateLimiter(visibilityConfig.PersistenceMaxQPS.AsFloat64())
		visibilityFromPinot = ratelimited.NewVisibilityManager(visibilityFromPinot, pinotRateLimiter)
	}

	if metricsClient != nil {
		// wrap with metrics
		visibilityFromPinot = pinotVisibility.NewPinotVisibilityMetricsClient(visibilityFromPinot, metricsClient, log)
	}

	return visibilityFromPinot
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
		esRateLimiter := quotas.NewDynamicRateLimiter(visibilityConfig.PersistenceMaxQPS.AsFloat64())
		visibilityFromES = ratelimited.NewVisibilityManager(visibilityFromES, esRateLimiter)
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
		result = errorinjectors.NewVisibilityManager(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = ratelimited.NewVisibilityManager(result, ds.ratelimit)
	}
	if visibilityConfig.EnableDBVisibilitySampling != nil && visibilityConfig.EnableDBVisibilitySampling() {
		result = sampled.NewVisibilityManager(result, sampled.Params{
			Config: &sampled.Config{
				VisibilityClosedMaxQPS: visibilityConfig.WriteDBVisibilityClosedMaxQPS,
				VisibilityListMaxQPS:   visibilityConfig.DBVisibilityListMaxQPS,
				VisibilityOpenMaxQPS:   visibilityConfig.WriteDBVisibilityOpenMaxQPS,
			},
			MetricClient:           f.metricsClient,
			Logger:                 f.logger,
			TimeSource:             clock.NewRealTimeSource(),
			RateLimiterFactoryFunc: sampled.NewDomainToBucketMap,
		})
	}
	if f.metricsClient != nil {
		result = metered.NewVisibilityManager(result, f.metricsClient, f.logger, f.config)
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
		result = errorinjectors.NewQueueManager(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = ratelimited.NewQueueManager(result, ds.ratelimit)
	}
	if f.metricsClient != nil {
		result = metered.NewQueueManager(result, f.metricsClient, f.logger, f.config)
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
		result = errorinjectors.NewConfigStoreManager(result, errorRate, f.logger)
	}
	if ds.ratelimit != nil {
		result = ratelimited.NewConfigStoreManager(result, ds.ratelimit)
	}
	if f.metricsClient != nil {
		result = metered.NewConfigStoreManager(result, f.metricsClient, f.logger, f.config)
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
		shardedNoSQLConfig := defaultCfg.NoSQL.ConvertToShardedNoSQLConfig()
		defaultDataStore.factory = nosql.NewFactory(*shardedNoSQLConfig, clusterName, f.logger, f.dc)
	case defaultCfg.ShardedNoSQL != nil:
		defaultDataStore.factory = nosql.NewFactory(*defaultCfg.ShardedNoSQL, clusterName, f.logger, f.dc)
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
			getSQLParser(f.logger, common.EncodingType(defaultCfg.SQL.EncodingType), decodingTypes...),
			f.dc)
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
		shardedNoSQLConfig := visibilityCfg.NoSQL.ConvertToShardedNoSQLConfig()
		visibilityDataStore.factory = nosql.NewFactory(*shardedNoSQLConfig, clusterName, f.logger, f.dc)
	case visibilityCfg.SQL != nil:
		var decodingTypes []common.EncodingType
		for _, dt := range visibilityCfg.SQL.DecodingTypes {
			decodingTypes = append(decodingTypes, common.EncodingType(dt))
		}
		visibilityDataStore.factory = sql.NewFactory(
			*visibilityCfg.SQL,
			clusterName,
			f.logger,
			getSQLParser(f.logger, common.EncodingType(visibilityCfg.SQL.EncodingType), decodingTypes...),
			f.dc)
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

func buildRatelimiters(cfg *config.Persistence, maxQPS quotas.RPSFunc) map[string]quotas.Limiter {
	result := make(map[string]quotas.Limiter, len(cfg.DataStores))
	for dsName := range cfg.DataStores {
		if maxQPS != nil && maxQPS() > 0 {
			result[dsName] = quotas.NewDynamicRateLimiter(maxQPS)
		}
	}
	return result
}

func setupPinotVisibilityManager(params *Params, resourceConfig *service.Config, logger log.Logger) (p.VisibilityManager, error) {
	visibilityProducer, err := params.MessagingClient.NewProducer(common.PinotVisibilityAppName)
	if err != nil {
		return nil, err
	}
	return newPinotVisibilityManager(params.PinotClient, resourceConfig, visibilityProducer, params.MetricsClient, logger), nil
}

func setupESVisibilityManager(params *Params, resourceConfig *service.Config, logger log.Logger) (p.VisibilityManager, error) {
	visibilityIndexName := params.ESConfig.Indices[common.VisibilityAppName]
	visibilityProducer, err := params.MessagingClient.NewProducer(common.VisibilityAppName)
	if err != nil {
		return nil, err
	}
	return newESVisibilityManager(visibilityIndexName, params.ESClient, resourceConfig, visibilityProducer, params.MetricsClient, logger), nil
}

func setupOSVisibilityManager(params *Params, resourceConfig *service.Config, logger log.Logger) (p.VisibilityManager, error) {
	visibilityIndexName := params.OSConfig.Indices[common.VisibilityAppName]
	visibilityProducer, err := params.MessagingClient.NewProducer(common.VisibilityAppName)
	if err != nil {
		return nil, err
	}
	return newESVisibilityManager(visibilityIndexName, params.OSClient, resourceConfig, visibilityProducer, params.MetricsClient, logger), nil
}
