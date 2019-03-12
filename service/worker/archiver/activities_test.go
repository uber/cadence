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

package archiver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-common/bark"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/metrics"
	mmocks "github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
)

const (
	testArchivalBucket     = "test-archival-bucket"
	testCurrentClusterName = "test-current-cluster-name"
)

var (
	persistenceNonRetryableErr = errors.New("persistence non-retryable error")
	persistenceRetryableErr    = &shared.InternalServiceError{}
	blobstoreNonRetryableErr   = &shared.BadRequestError{}
	blobstoreRetryableErr      = errors.New("blobstore retryable error")
)

type activitiesSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	logger        bark.Logger
	metricsClient *mmocks.Client
}

func TestActivitiesSuite(t *testing.T) {
	suite.Run(t, new(activitiesSuite))
}

func (s *activitiesSuite) SetupTest() {
	s.logger = bark.NewNopLogger()
	s.metricsClient = &mmocks.Client{}
	s.metricsClient.On("StartTimer", metrics.ArchiverUploadHistoryActivityScope, metrics.CadenceLatency).Return(tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})).Once()
}

func (s *activitiesSuite) TearDownTest() {
	s.metricsClient.AssertExpectations(s.T())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_DomainCacheNonRetryableError() {
	domainCache := &cache.DomainCacheMock{}
	domainCache.On("GetDomainByID", mock.Anything).Return(nil, persistenceNonRetryableErr).Once()
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.ArchiverNonRetryableErrorCount).Once()
	container := &BootstrapContainer{
		Logger:        s.logger,
		MetricsClient: s.metricsClient,
		DomainCache:   domainCache,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(errGetDomainByID, err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_TimeoutGettingDomainCacheEntry() {
	domainCache := &cache.DomainCacheMock{}
	domainCache.On("GetDomainByID", mock.Anything).Return(nil, persistenceRetryableErr).Once()
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.CadenceErrContextTimeoutCounter).Once()
	container := &BootstrapContainer{
		Logger:        s.logger,
		MetricsClient: s.metricsClient,
		DomainCache:   domainCache,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(getCanceledContext(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(contextTimeoutErr.Error(), err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Skip_ClusterArchivalNotEnabled() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.ArchiverSkipUploadCount).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, false)
	container := &BootstrapContainer{
		Logger:          s.logger,
		MetricsClient:   s.metricsClient,
		DomainCache:     domainCache,
		ClusterMetadata: mockClusterMetadata,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.NoError(err)
}

func (s *activitiesSuite) TestUploadHistoryActivity_Skip_DomainArchivalNotEnabled() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.ArchiverSkipUploadCount).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(false, "", true)
	container := &BootstrapContainer{
		Logger:          s.logger,
		MetricsClient:   s.metricsClient,
		DomainCache:     domainCache,
		ClusterMetadata: mockClusterMetadata,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.NoError(err)
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_DomainConfigMissingBucket() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.ArchiverNonRetryableErrorCount).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, "", true)
	container := &BootstrapContainer{
		Logger:          s.logger,
		MetricsClient:   s.metricsClient,
		DomainCache:     domainCache,
		ClusterMetadata: mockClusterMetadata,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(errEmptyBucket, err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_NextBlobNonRetryableError() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.ArchiverNonRetryableErrorCount).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, true)
	mockHistoryBlobIterator := &HistoryBlobIteratorMock{}
	mockHistoryBlobIterator.On("HasNext").Return(true)
	mockHistoryBlobIterator.On("Next").Return(nil, persistenceNonRetryableErr)
	container := &BootstrapContainer{
		Logger:              s.logger,
		MetricsClient:       s.metricsClient,
		DomainCache:         domainCache,
		ClusterMetadata:     mockClusterMetadata,
		HistoryBlobIterator: mockHistoryBlobIterator,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(errNextBlob, err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_TimeoutGettingNextBlob() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.CadenceErrContextTimeoutCounter).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, true)
	mockHistoryBlobIterator := &HistoryBlobIteratorMock{}
	mockHistoryBlobIterator.On("HasNext").Return(true)
	mockHistoryBlobIterator.On("Next").Return(nil, persistenceRetryableErr)
	container := &BootstrapContainer{
		Logger:              s.logger,
		MetricsClient:       s.metricsClient,
		DomainCache:         domainCache,
		ClusterMetadata:     mockClusterMetadata,
		HistoryBlobIterator: mockHistoryBlobIterator,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(getCanceledContext(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(contextTimeoutErr.Error(), err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_ConstructBlobKeyError() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.ArchiverNonRetryableErrorCount).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, true)
	mockHistoryBlobIterator := &HistoryBlobIteratorMock{}
	mockHistoryBlobIterator.On("HasNext").Return(true)
	historyBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(1),
		},
	}
	mockHistoryBlobIterator.On("Next").Return(historyBlob, nil)
	container := &BootstrapContainer{
		Logger:              s.logger,
		MetricsClient:       s.metricsClient,
		DomainCache:         domainCache,
		ClusterMetadata:     mockClusterMetadata,
		HistoryBlobIterator: mockHistoryBlobIterator,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           "", // this causes an error when creating the blob key
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(errConstructBlob, err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_BlobExistsNonRetryableError() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.ArchiverNonRetryableErrorCount).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, true)
	mockHistoryBlobIterator := &HistoryBlobIteratorMock{}
	mockHistoryBlobIterator.On("HasNext").Return(true)
	historyBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(1),
		},
	}
	mockHistoryBlobIterator.On("Next").Return(historyBlob, nil)
	mockBlobstore := &mocks.BlobstoreClient{}
	mockBlobstore.On("Exists", mock.Anything, mock.Anything, mock.Anything).Return(false, blobstoreNonRetryableErr)
	container := &BootstrapContainer{
		Logger:              s.logger,
		MetricsClient:       s.metricsClient,
		DomainCache:         domainCache,
		ClusterMetadata:     mockClusterMetadata,
		HistoryBlobIterator: mockHistoryBlobIterator,
		Blobstore:           mockBlobstore,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(errBlobExists, err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_TimeoutOnBlobExists() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.CadenceErrContextTimeoutCounter).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, true)
	mockHistoryBlobIterator := &HistoryBlobIteratorMock{}
	mockHistoryBlobIterator.On("HasNext").Return(true)
	historyBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(1),
		},
	}
	mockHistoryBlobIterator.On("Next").Return(historyBlob, nil)
	mockBlobstore := &mocks.BlobstoreClient{}
	mockBlobstore.On("Exists", mock.Anything, mock.Anything, mock.Anything).Return(false, blobstoreRetryableErr)
	container := &BootstrapContainer{
		Logger:              s.logger,
		MetricsClient:       s.metricsClient,
		DomainCache:         domainCache,
		ClusterMetadata:     mockClusterMetadata,
		HistoryBlobIterator: mockHistoryBlobIterator,
		Blobstore:           mockBlobstore,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(getCanceledContext(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(contextTimeoutErr.Error(), err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Nop_BlobAlreadyExists() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.ArchiverBlobAlreadyExistsCount).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, true)
	mockHistoryBlobIterator := &HistoryBlobIteratorMock{}
	mockHistoryBlobIterator.On("HasNext").Return(true).Once()
	mockHistoryBlobIterator.On("HasNext").Return(false).Once()
	historyBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(1),
		},
	}
	mockHistoryBlobIterator.On("Next").Return(historyBlob, nil)
	mockBlobstore := &mocks.BlobstoreClient{}
	mockBlobstore.On("Exists", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	container := &BootstrapContainer{
		Logger:              s.logger,
		MetricsClient:       s.metricsClient,
		DomainCache:         domainCache,
		ClusterMetadata:     mockClusterMetadata,
		HistoryBlobIterator: mockHistoryBlobIterator,
		Blobstore:           mockBlobstore,
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.NoError(err)
	mockBlobstore.AssertNotCalled(s.T(), "Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_UploadBlobNonRetryableError() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.ArchiverNonRetryableErrorCount).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, true)
	mockHistoryBlobIterator := &HistoryBlobIteratorMock{}
	mockHistoryBlobIterator.On("HasNext").Return(true).Once()
	mockHistoryBlobIterator.On("HasNext").Return(false).Once()
	historyBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(1),
		},
	}
	mockHistoryBlobIterator.On("Next").Return(historyBlob, nil)
	mockBlobstore := &mocks.BlobstoreClient{}
	mockBlobstore.On("Exists", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	mockBlobstore.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(blobstoreNonRetryableErr)
	container := &BootstrapContainer{
		Logger:              s.logger,
		MetricsClient:       s.metricsClient,
		DomainCache:         domainCache,
		ClusterMetadata:     mockClusterMetadata,
		HistoryBlobIterator: mockHistoryBlobIterator,
		Blobstore:           mockBlobstore,
		Config:              constructConfig(testDefaultPersistencePageSize, testDefaultTargetArchivalBlobSize),
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(errUploadBlob, err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Fail_TimeoutOnUploadBlob() {
	s.metricsClient.On("IncCounter", metrics.ArchiverUploadHistoryActivityScope, metrics.CadenceErrContextTimeoutCounter).Once()
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, true)
	mockHistoryBlobIterator := &HistoryBlobIteratorMock{}
	mockHistoryBlobIterator.On("HasNext").Return(true).Once()
	mockHistoryBlobIterator.On("HasNext").Return(false).Once()
	historyBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(1),
		},
	}
	mockHistoryBlobIterator.On("Next").Return(historyBlob, nil)
	mockBlobstore := &mocks.BlobstoreClient{}
	mockBlobstore.On("Exists", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	mockBlobstore.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(blobstoreRetryableErr)
	container := &BootstrapContainer{
		Logger:              s.logger,
		MetricsClient:       s.metricsClient,
		DomainCache:         domainCache,
		ClusterMetadata:     mockClusterMetadata,
		HistoryBlobIterator: mockHistoryBlobIterator,
		Blobstore:           mockBlobstore,
		Config:              constructConfig(testDefaultPersistencePageSize, testDefaultTargetArchivalBlobSize),
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(getCanceledContext(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.Equal(contextTimeoutErr.Error(), err.Error())
}

func (s *activitiesSuite) TestUploadHistoryActivity_Success() {
	domainCache, mockClusterMetadata := s.archivalConfig(true, testArchivalBucket, true)
	mockHistoryBlobIterator := &HistoryBlobIteratorMock{}
	mockHistoryBlobIterator.On("HasNext").Return(true).Once()
	mockHistoryBlobIterator.On("HasNext").Return(false).Once()
	historyBlob := &HistoryBlob{
		Header: &HistoryBlobHeader{
			CurrentPageToken: common.IntPtr(1),
		},
	}
	mockHistoryBlobIterator.On("Next").Return(historyBlob, nil)
	mockBlobstore := &mocks.BlobstoreClient{}
	mockBlobstore.On("Exists", mock.Anything, mock.Anything, mock.Anything).Return(false, nil).Once()
	mockBlobstore.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	container := &BootstrapContainer{
		Logger:              s.logger,
		MetricsClient:       s.metricsClient,
		DomainCache:         domainCache,
		ClusterMetadata:     mockClusterMetadata,
		HistoryBlobIterator: mockHistoryBlobIterator,
		Blobstore:           mockBlobstore,
		Config:              constructConfig(testDefaultPersistencePageSize, testDefaultTargetArchivalBlobSize),
	}
	env := s.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), bootstrapContainerKey, container),
	})
	request := ArchiveRequest{
		DomainID:             testDomainID,
		WorkflowID:           testWorkflowID,
		RunID:                testRunID,
		BranchToken:          testBranchToken,
		NextEventID:          testNextEventID,
		CloseFailoverVersion: testCloseFailoverVersion,
	}
	_, err := env.ExecuteActivity(uploadHistoryActivity, request)
	s.NoError(err)
}

func (s *activitiesSuite) archivalConfig(
	domainEnablesArchival bool,
	domainArchivalBucket string,
	clusterEnablesArchival bool,
) (cache.DomainCache, cluster.Metadata) {
	domainArchivalStatus := shared.ArchivalStatusDisabled
	if domainEnablesArchival {
		domainArchivalStatus = shared.ArchivalStatusEnabled
	}
	clusterArchivalStatus := cluster.ArchivalDisabled
	clusterDefaultBucket := ""
	if clusterEnablesArchival {
		clusterDefaultBucket = "default-bucket"
		clusterArchivalStatus = cluster.ArchivalEnabled
	}
	mockMetadataMgr := &mocks.MetadataManager{}
	mockClusterMetadata := &mocks.ClusterMetadata{}
	mockClusterMetadata.On("ArchivalConfig").Return(cluster.NewArchivalConfig(clusterArchivalStatus, clusterDefaultBucket))
	mockClusterMetadata.On("IsGlobalDomainEnabled").Return(false)
	mockClusterMetadata.On("GetCurrentClusterName").Return(testCurrentClusterName)
	mockMetadataMgr.On("GetDomain", mock.Anything).Return(
		&persistence.GetDomainResponse{
			Info: &persistence.DomainInfo{ID: testDomainID, Name: testDomain},
			Config: &persistence.DomainConfig{
				Retention:      1,
				ArchivalBucket: domainArchivalBucket,
				ArchivalStatus: domainArchivalStatus,
			},
			ReplicationConfig: &persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
				},
			},
			TableVersion: persistence.DomainTableVersionV1,
		},
		nil,
	)
	return cache.NewDomainCache(mockMetadataMgr, mockClusterMetadata, s.metricsClient, s.logger), mockClusterMetadata
}

func getCanceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
