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

package host

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/uber-go/tally"

	adminClient "github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/filestore"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/elasticsearch"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/messaging/kafka"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql"
	"github.com/uber/cadence/common/persistence/persistence-tests/testcluster"

	// the import is a test dependency
	_ "github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql/public"
	persistencetests "github.com/uber/cadence/common/persistence/persistence-tests"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin/mysql"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin/postgres"
)

type (
	// TestCluster is a base struct for integration tests
	TestCluster struct {
		testBase     persistencetests.TestBase
		archiverBase *ArchiverBase
		host         Cadence
	}

	// ArchiverBase is a base struct for archiver provider being used in integration tests
	ArchiverBase struct {
		metadata                 archiver.ArchivalMetadata
		provider                 provider.ArchiverProvider
		historyStoreDirectory    string
		visibilityStoreDirectory string
		historyURI               string
		visibilityURI            string
	}

	// TestClusterConfig are config for a test cluster
	TestClusterConfig struct {
		FrontendAddress       string
		EnableArchival        bool
		IsPrimaryCluster      bool
		ClusterNo             int
		ClusterGroupMetadata  config.ClusterGroupMetadata
		MessagingClientConfig *MessagingClientConfig
		Persistence           persistencetests.TestBaseOptions
		HistoryConfig         *HistoryConfig
		ESConfig              *config.ElasticSearchConfig
		WorkerConfig          *WorkerConfig
		MockAdminClient       map[string]adminClient.Client
	}

	// MessagingClientConfig is the config for messaging config
	MessagingClientConfig struct {
		UseMock     bool
		KafkaConfig *config.KafkaConfig
	}

	// WorkerConfig is the config for enabling/disabling cadence worker
	WorkerConfig struct {
		EnableArchiver   bool
		EnableIndexer    bool
		EnableReplicator bool
	}
)

const (
	defaultTestValueOfESIndexMaxResultWindow = 5
	defaultTestPersistenceTimeout            = 5 * time.Second
)

// NewCluster creates and sets up the test cluster
func NewCluster(options *TestClusterConfig, logger log.Logger, params persistencetests.TestBaseParams) (*TestCluster, error) {
	testBase := persistencetests.NewTestBaseFromParams(params)
	testBase.Setup()
	setupShards(testBase, options.HistoryConfig.NumHistoryShards, logger)
	archiverBase := newArchiverBase(options.EnableArchival, logger)
	messagingClient := getMessagingClient(options.MessagingClientConfig, logger)
	var esClient elasticsearch.GenericClient
	if options.WorkerConfig.EnableIndexer {
		var err error
		esClient, err = elasticsearch.NewGenericClient(options.ESConfig, logger)
		if err != nil {
			return nil, err
		}
	}

	pConfig := testBase.Config()
	pConfig.NumHistoryShards = options.HistoryConfig.NumHistoryShards
	scope := tally.NewTestScope("integration-test", nil)
	metricsClient := metrics.NewClient(scope, metrics.ServiceIdx(0))
	domainReplicationQueue := domain.NewReplicationQueue(
		testBase.DomainReplicationQueueMgr,
		options.ClusterGroupMetadata.CurrentClusterName,
		metricsClient,
		logger,
	)
	aConfig := noopAuthorizationConfig()
	cadenceParams := &CadenceParams{
		ClusterMetadata:               params.ClusterMetadata,
		PersistenceConfig:             pConfig,
		MessagingClient:               messagingClient,
		DomainManager:                 testBase.DomainManager,
		HistoryV2Mgr:                  testBase.HistoryV2Mgr,
		ExecutionMgrFactory:           testBase.ExecutionMgrFactory,
		DomainReplicationQueue:        domainReplicationQueue,
		Logger:                        logger,
		ClusterNo:                     options.ClusterNo,
		ESConfig:                      options.ESConfig,
		ESClient:                      esClient,
		ArchiverMetadata:              archiverBase.metadata,
		ArchiverProvider:              archiverBase.provider,
		HistoryConfig:                 options.HistoryConfig,
		WorkerConfig:                  options.WorkerConfig,
		MockAdminClient:               options.MockAdminClient,
		DomainReplicationTaskExecutor: domain.NewReplicationTaskExecutor(testBase.DomainManager, clock.NewRealTimeSource(), logger),
		AuthorizationConfig:           aConfig,
	}
	cluster := NewCadence(cadenceParams)
	if err := cluster.Start(); err != nil {
		return nil, err
	}

	return &TestCluster{testBase: testBase, archiverBase: archiverBase, host: cluster}, nil
}

func noopAuthorizationConfig() config.Authorization {
	return config.Authorization{
		OAuthAuthorizer: config.OAuthAuthorizer{
			Enable: false,
		},
		NoopAuthorizer: config.NoopAuthorizer{
			Enable: true,
		},
	}
}

// NewClusterMetadata returns cluster metdata from config
func NewClusterMetadata(options *TestClusterConfig, logger log.Logger) cluster.Metadata {
	clusterMetadata := cluster.GetTestClusterMetadata(
		options.ClusterGroupMetadata.EnableGlobalDomain,
		options.IsPrimaryCluster,
	)
	if !options.IsPrimaryCluster && options.ClusterGroupMetadata.PrimaryClusterName != "" { // xdc cluster metadata setup
		clusterMetadata = cluster.NewMetadata(
			logger,
			dynamicconfig.GetBoolPropertyFn(options.ClusterGroupMetadata.EnableGlobalDomain),
			options.ClusterGroupMetadata.FailoverVersionIncrement,
			options.ClusterGroupMetadata.PrimaryClusterName,
			options.ClusterGroupMetadata.CurrentClusterName,
			options.ClusterGroupMetadata.ClusterGroup,
		)
	}
	return clusterMetadata
}

func NewPersistenceTestCluster(clusterConfig *TestClusterConfig) testcluster.PersistenceTestCluster {
	// NOTE: Override here to keep consistent. clusterConfig will be used in the test for some purposes.
	clusterConfig.Persistence.StoreType = TestFlags.PersistenceType
	clusterConfig.Persistence.DBPluginName = TestFlags.SQLPluginName

	var testCluster testcluster.PersistenceTestCluster
	if TestFlags.PersistenceType == config.StoreTypeCassandra {
		// TODO refactor to support other NoSQL
		ops := clusterConfig.Persistence
		ops.DBPluginName = "cassandra"
		testCluster = nosql.NewTestCluster(ops.DBPluginName, ops.DBName, ops.DBUsername, ops.DBPassword, ops.DBHost, ops.DBPort, ops.ProtoVersion, "")
	} else if TestFlags.PersistenceType == config.StoreTypeSQL {
		var ops *persistencetests.TestBaseOptions
		if TestFlags.SQLPluginName == mysql.PluginName {
			ops = mysql.GetTestClusterOption()
		} else if TestFlags.SQLPluginName == postgres.PluginName {
			ops = postgres.GetTestClusterOption()
		} else {
			panic("not supported plugin " + TestFlags.SQLPluginName)
		}
		testCluster = sql.NewTestCluster(TestFlags.SQLPluginName, clusterConfig.Persistence.DBName, ops.DBUsername, ops.DBPassword, ops.DBHost, ops.DBPort, ops.SchemaDir)
	} else {
		panic("not supported storage type" + TestFlags.PersistenceType)
	}
	return testCluster
}

func setupShards(testBase persistencetests.TestBase, numHistoryShards int, logger log.Logger) {
	// shard 0 is always created, we create additional shards if needed
	for shardID := 1; shardID < numHistoryShards; shardID++ {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTestPersistenceTimeout)
		err := testBase.CreateShard(ctx, shardID, "", 0)
		if err != nil {
			cancel()
			logger.Fatal("Failed to create shard", tag.Error(err))
		}
		cancel()
	}
}

func newArchiverBase(enabled bool, logger log.Logger) *ArchiverBase {
	dcCollection := dynamicconfig.NewNopCollection()
	if !enabled {
		return &ArchiverBase{
			metadata: archiver.NewArchivalMetadata(dcCollection, "", false, "", false, &config.ArchivalDomainDefaults{}),
			provider: provider.NewArchiverProvider(nil, nil),
		}
	}

	historyStoreDirectory, err := ioutil.TempDir("", "test-history-archival")
	if err != nil {
		logger.Fatal("Failed to create temp dir for history archival", tag.Error(err))
	}
	visibilityStoreDirectory, err := ioutil.TempDir("", "test-visibility-archival")
	if err != nil {
		logger.Fatal("Failed to create temp dir for visibility archival", tag.Error(err))
	}
	cfg := &config.FilestoreArchiver{
		FileMode: "0666",
		DirMode:  "0766",
	}
	provider := provider.NewArchiverProvider(
		&config.HistoryArchiverProvider{
			Filestore: cfg,
		},
		&config.VisibilityArchiverProvider{
			Filestore: cfg,
		},
	)
	return &ArchiverBase{
		metadata: archiver.NewArchivalMetadata(dcCollection, "enabled", true, "enabled", true, &config.ArchivalDomainDefaults{
			History: config.HistoryArchivalDomainDefaults{
				Status: "enabled",
				URI:    "testScheme://test/history/archive/path",
			},
			Visibility: config.VisibilityArchivalDomainDefaults{
				Status: "enabled",
				URI:    "testScheme://test/visibility/archive/path",
			},
		}),
		provider:                 provider,
		historyStoreDirectory:    historyStoreDirectory,
		visibilityStoreDirectory: visibilityStoreDirectory,
		historyURI:               filestore.URIScheme + "://" + historyStoreDirectory,
		visibilityURI:            filestore.URIScheme + "://" + visibilityStoreDirectory,
	}
}

func getMessagingClient(config *MessagingClientConfig, logger log.Logger) messaging.Client {
	if config == nil || config.UseMock {
		return mocks.NewMockMessagingClient(&mocks.KafkaProducer{}, nil)
	}
	checkApp := len(config.KafkaConfig.Applications) != 0
	return kafka.NewKafkaClient(config.KafkaConfig, metrics.NewNoopMetricsClient(), logger, tally.NoopScope, checkApp)
}

// TearDownCluster tears down the test cluster
func (tc *TestCluster) TearDownCluster() {
	tc.host.Stop()
	tc.host = nil
	tc.testBase.TearDownWorkflowStore()
	os.RemoveAll(tc.archiverBase.historyStoreDirectory)
	os.RemoveAll(tc.archiverBase.visibilityStoreDirectory)
}

// GetFrontendClient returns a frontend client from the test cluster
func (tc *TestCluster) GetFrontendClient() FrontendClient {
	return tc.host.GetFrontendClient()
}

// GetAdminClient returns an admin client from the test cluster
func (tc *TestCluster) GetAdminClient() AdminClient {
	return tc.host.GetAdminClient()
}

// GetHistoryClient returns a history client from the test cluster
func (tc *TestCluster) GetHistoryClient() HistoryClient {
	return tc.host.GetHistoryClient()
}

// GetExecutionManagerFactory returns an execution manager factory from the test cluster
func (tc *TestCluster) GetExecutionManagerFactory() persistence.ExecutionManagerFactory {
	return tc.host.GetExecutionManagerFactory()
}
