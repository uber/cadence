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

package cache

import (
	"os"
	"sync"
	"testing"

	"github.com/pborman/uuid"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
)

type (
	domainCacheSuite struct {
		suite.Suite

		logger          bark.Logger
		clusterMetadata *mocks.ClusterMetadata
		metadataMgr     *mocks.MetadataManager
		domainCache     *domainCache
	}
)

func TestDomainCacheSuite(t *testing.T) {
	s := new(domainCacheSuite)
	suite.Run(t, s)
}

func (s *domainCacheSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}

}

func (s *domainCacheSuite) TearDownSuite() {

}

func (s *domainCacheSuite) SetupTest() {
	log2 := log.New()
	log2.Level = log.DebugLevel
	s.logger = bark.NewLoggerFromLogrus(log2)
	s.clusterMetadata = &mocks.ClusterMetadata{}
	s.metadataMgr = &mocks.MetadataManager{}
	s.domainCache = NewDomainCache(s.metadataMgr, s.clusterMetadata, s.logger).(*domainCache)
}

func (s *domainCacheSuite) TearDownTest() {
	s.domainCache.Stop()
	s.clusterMetadata.AssertExpectations(s.T())
	s.metadataMgr.AssertExpectations(s.T())
}

func (s *domainCacheSuite) TestListDomain() {
	domainRecord1 := &persistence.GetDomainResponse{
		Info:   &persistence.DomainInfo{ID: uuid.New(), Name: "some random domain name"},
		Config: &persistence.DomainConfig{Retention: 1},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestCurrentClusterName},
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
	}
	entry1 := s.buildEntryFromRecord(domainRecord1)

	domainRecord2 := &persistence.GetDomainResponse{
		Info:   &persistence.DomainInfo{ID: uuid.New(), Name: "another random domain name"},
		Config: &persistence.DomainConfig{Retention: 2},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestAlternativeClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestCurrentClusterName},
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
	}
	entry2 := s.buildEntryFromRecord(domainRecord2)

	pageToken := []byte("some random page token")

	s.metadataMgr.On("GetMetadata").Return(int64(123), nil)
	s.clusterMetadata.On("IsGlobalDomainEnabled").Return(true)
	s.metadataMgr.On("ListDomain", &persistence.ListDomainRequest{
		PageSize:      domainCacheRefreshPageSize,
		NextPageToken: nil,
	}).Return(&persistence.ListDomainResponse{
		Domains:       []*persistence.GetDomainResponse{domainRecord1},
		NextPageToken: pageToken,
	}, nil).Once()

	s.metadataMgr.On("ListDomain", &persistence.ListDomainRequest{
		PageSize:      domainCacheRefreshPageSize,
		NextPageToken: pageToken,
	}).Return(&persistence.ListDomainResponse{
		Domains:       []*persistence.GetDomainResponse{domainRecord2},
		NextPageToken: nil,
	}, nil).Once()

	// load domains
	s.domainCache.Start()

	entryByName1, err := s.domainCache.GetDomain(domainRecord1.Info.Name)
	s.Nil(err)
	s.Equal(entry1, entryByName1)
	entryByID1, err := s.domainCache.GetDomainByID(domainRecord1.Info.ID)
	s.Nil(err)
	s.Equal(entry1, entryByID1)

	entryByName2, err := s.domainCache.GetDomain(domainRecord2.Info.Name)
	s.Nil(err)
	s.Equal(entry2, entryByName2)
	entryByID2, err := s.domainCache.GetDomainByID(domainRecord2.Info.ID)
	s.Nil(err)
	s.Equal(entry2, entryByID2)
}

func (s *domainCacheSuite) TestGetDomain_NonLoaded_GetByName() {
	s.clusterMetadata.On("IsGlobalDomainEnabled").Return(true)
	domainRecord := &persistence.GetDomainResponse{
		Info:   &persistence.DomainInfo{ID: uuid.New(), Name: "some random domain name"},
		Config: &persistence.DomainConfig{Retention: 1},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestCurrentClusterName},
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
	}
	entry := s.buildEntryFromRecord(domainRecord)

	s.metadataMgr.On("GetDomain", &persistence.GetDomainRequest{Name: entry.info.Name}).Return(domainRecord, nil).Once()

	entryByName, err := s.domainCache.GetDomain(domainRecord.Info.Name)
	s.Nil(err)
	s.Equal(entry, entryByName)
	entryByName, err = s.domainCache.GetDomain(domainRecord.Info.Name)
	s.Nil(err)
	s.Equal(entry, entryByName)
}

func (s *domainCacheSuite) TestGetDomain_NonLoaded_GetByID() {
	s.clusterMetadata.On("IsGlobalDomainEnabled").Return(true)
	domainRecord := &persistence.GetDomainResponse{
		Info:   &persistence.DomainInfo{ID: uuid.New(), Name: "some random domain name"},
		Config: &persistence.DomainConfig{Retention: 1},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestCurrentClusterName},
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
	}
	entry := s.buildEntryFromRecord(domainRecord)

	s.metadataMgr.On("GetDomain", &persistence.GetDomainRequest{ID: entry.info.ID}).Return(domainRecord, nil).Once()

	entryByID, err := s.domainCache.GetDomainByID(domainRecord.Info.ID)
	s.Nil(err)
	s.Equal(entry, entryByID)
	entryByID, err = s.domainCache.GetDomainByID(domainRecord.Info.ID)
	s.Nil(err)
	s.Equal(entry, entryByID)
}

func (s *domainCacheSuite) TestUpdateCache_Trigger() {
	s.clusterMetadata.On("IsGlobalDomainEnabled").Return(true)
	domainRecordOld := &persistence.GetDomainResponse{
		Info:   &persistence.DomainInfo{ID: uuid.New(), Name: "some random domain name"},
		Config: &persistence.DomainConfig{Retention: 1},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestCurrentClusterName},
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
	}
	entryOld := s.buildEntryFromRecord(domainRecordOld)

	domainRecordNew := &persistence.GetDomainResponse{
		Info:   entryOld.info,
		Config: &persistence.DomainConfig{Retention: 2},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestAlternativeClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestCurrentClusterName},
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		ConfigVersion: domainRecordOld.ConfigVersion + 1,
	}
	entryNew := s.buildEntryFromRecord(domainRecordNew)

	s.metadataMgr.On("GetDomain", &persistence.GetDomainRequest{ID: entryOld.info.ID}).Return(domainRecordOld, nil).Once()
	entry, err := s.domainCache.GetDomainByID(entryOld.info.ID)
	s.Nil(err)
	s.Equal(entryOld, entry)

	callbackInvoked := false
	s.domainCache.RegisterDomainChangeCallback(0, func(prevDomain *DomainCacheEntry, nextDomain *DomainCacheEntry) {
		s.Equal(entryOld, prevDomain)
		s.Equal(entryNew, nextDomain)
		callbackInvoked = true
	})

	entry, err = s.domainCache.updateIDToDomainCache(domainRecordNew.Info.ID, domainRecordNew)
	s.Nil(err)
	s.Equal(entryNew, entry)
	s.True(callbackInvoked)
}

func (s *domainCacheSuite) TestGetUpdateCache_ConcurrentAccess() {
	s.clusterMetadata.On("IsGlobalDomainEnabled").Return(true)
	id := uuid.New()
	domainRecordOld := &persistence.GetDomainResponse{
		Info:   &persistence.DomainInfo{ID: id, Name: "some random domain name"},
		Config: &persistence.DomainConfig{Retention: 1},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestCurrentClusterName},
				&persistence.ClusterReplicationConfig{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		ConfigVersion:   0,
		FailoverVersion: 0,
	}
	entryOld := s.buildEntryFromRecord(domainRecordOld)

	s.metadataMgr.On("GetDomain", &persistence.GetDomainRequest{ID: id}).Return(domainRecordOld, nil).Once()
	s.domainCache.GetDomainByID(id)

	coroutineCountGet := 100
	coroutineCountUpdate := 100
	waitGroup := &sync.WaitGroup{}
	stopChan := make(chan struct{})
	testGetFn := func() {
		<-stopChan
		entryNew, err := s.domainCache.GetDomainByID(id)
		s.Nil(err)
		// make the config version the same so we can easily compare those
		entryNew.configVersion = 0
		entryNew.failoverVersion = 0
		s.Equal(entryOld, entryNew)
		waitGroup.Done()
	}

	testUpdateFn := func() {
		<-stopChan
		entryNew, err := s.domainCache.GetDomainByID(id)
		s.Nil(err)
		s.domainCache.updateIDToDomainCache(id, &persistence.GetDomainResponse{
			Info:              entryNew.GetInfo(),
			Config:            entryNew.GetConfig(),
			ReplicationConfig: entryNew.GetReplicationConfig(),
			ConfigVersion:     entryNew.GetConfigVersion() + 1,
			FailoverVersion:   entryNew.GetFailoverVersion() + 1,
		})
		waitGroup.Done()
	}

	for i := 0; i < coroutineCountGet; i++ {
		waitGroup.Add(1)
		go testGetFn()
	}
	for i := 0; i < coroutineCountUpdate; i++ {
		waitGroup.Add(1)
		go testUpdateFn()
	}
	close(stopChan)
	waitGroup.Wait()
}

func (s *domainCacheSuite) buildEntryFromRecord(record *persistence.GetDomainResponse) *DomainCacheEntry {
	newEntry := newDomainCacheEntry(s.clusterMetadata)
	newEntry.info = record.Info
	newEntry.config = record.Config
	newEntry.replicationConfig = record.ReplicationConfig
	newEntry.configVersion = record.ConfigVersion
	newEntry.failoverVersion = record.FailoverVersion
	newEntry.isGlobalDomain = record.IsGlobalDomain
	return newEntry
}
