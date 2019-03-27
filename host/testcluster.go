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
	"github.com/uber-common/bark"
	"github.com/uber/cadence/client"
	"github.com/uber/cadence/common/blobstore/filestore"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/persistence/persistence-tests"
	"io/ioutil"
	"os"
)

type (
	// TestCluster is a base struct for integration tests
	TestCluster struct {
		testBase  persistencetests.TestBase
		blobstore *BlobstoreBase
		host      Cadence
	}

	// TestClusterOptions are options for test cluster
	TestClusterOptions struct {
		PersistOptions   *persistencetests.TestBaseOptions
		EnableWorker     bool
		EnableEventsV2   bool
		EnableArchival   bool
		ClusterNo        int
		NumHistoryShards int
		MessagingClient  messaging.Client
	}
)

// NewCluster creates and sets up the test cluster
func NewCluster(options *TestClusterOptions, logger bark.Logger) (*TestCluster, error) {
	testBase := persistencetests.NewTestBaseWithCassandra(options.PersistOptions)
	testBase.Setup()
	setupShards(testBase, options.NumHistoryShards, logger)
	blobstore := setupBlobstore(logger)

	cadenceParams := &CadenceParams{
		ClusterMetadata:               testBase.ClusterMetadata,
		DispatcherProvider:            client.NewIPYarpcDispatcherProvider(),
		MessagingClient:               options.MessagingClient,
		MetadataMgr:                   testBase.MetadataProxy,
		MetadataMgrV2:                 testBase.MetadataManagerV2,
		ShardMgr:                      testBase.ShardMgr,
		HistoryMgr:                    testBase.HistoryMgr,
		HistoryV2Mgr:                  testBase.HistoryV2Mgr,
		ExecutionMgrFactory:           testBase.ExecutionMgrFactory,
		TaskMgr:                       testBase.TaskMgr,
		VisibilityMgr:                 testBase.VisibilityMgr,
		NumberOfHistoryShards:         options.NumHistoryShards,
		NumberOfHistoryHosts:          testNumberOfHistoryHosts,
		Logger:                        logger,
		ClusterNo:                     options.ClusterNo,
		EnableWorker:                  options.EnableWorker,
		EnableEventsV2:                options.EnableEventsV2,
		EnableVisibilityToKafka:       false,
		EnableReadHistoryFromArchival: true,
		Blobstore:                     blobstore.client,
	}
	cluster := NewCadence(cadenceParams)
	if err := cluster.Start(); err != nil {
		return nil, err
	}

	return &TestCluster{testBase: testBase, blobstore: blobstore, host: cluster}, nil
}

func setupShards(testBase persistencetests.TestBase, numHistoryShards int, logger bark.Logger) {
	// shard 0 is always created, we create additional shards if needed
	for shardID := 1; shardID < numHistoryShards; shardID++ {
		err := testBase.CreateShard(shardID, "", 0)
		if err != nil {
			logger.WithField("error", err).Fatal("Failed to create shard")
		}
	}
}

func setupBlobstore(logger bark.Logger) *BlobstoreBase {
	bucketName := "default-test-bucket"
	storeDirectory, err := ioutil.TempDir("", "test-blobstore")
	if err != nil {
		logger.WithField("error", err).Fatal("Failed to create temp dir for blobstore")
	}
	cfg := &filestore.Config{
		StoreDirectory: storeDirectory,
		DefaultBucket: filestore.BucketConfig{
			Name:          bucketName,
			Owner:         "test-owner",
			RetentionDays: 10,
		},
	}
	client, err := filestore.NewClient(cfg)
	if err != nil {
		logger.WithField("error", err).Fatal("Failed to construct blobstore client")
	}
	return &BlobstoreBase{
		client:         client,
		storeDirectory: storeDirectory,
		bucketName:     bucketName,
	}
}

// TearDownCluster tears down the test cluster
func (tc *TestCluster) TearDownCluster() {
	tc.host.Stop()
	tc.host = nil
	tc.testBase.TearDownWorkflowStore()
	os.RemoveAll(tc.blobstore.storeDirectory)
}

// GetFrontendClient returns a frontend client from the test cluster
func (tc *TestCluster) GetFrontendClient() FrontendClient {
	return tc.host.GetFrontendClient()
}

// GetAdminClient returns an admin client from the test cluster
func (tc *TestCluster) GetAdminClient() AdminClient {
	return tc.host.GetAdminClient()
}
