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

package engineimpl

import (
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"testing"
)

//func TestDomainChangeCallback(t *testing.T) {
//
//	domainInfo := persistence.DomainInfo{
//		ID:         "83b48dab-68cb-4f73-8752-c75d9271977f",
//		Name:       "samples-domain",
//		OwnerEmail: "test4@test.com",
//		Data:       map[string]string{},
//	}
//	domainConfig := persistence.DomainConfig{
//		Retention:                3,
//		EmitMetric:               true,
//		HistoryArchivalURI:       "file:///tmp/cadence_archival/development",
//		VisibilityArchivalURI:    "file:///tmp/cadence_vis_archival/development",
//		BadBinaries:              types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
//		IsolationGroups:          types.IsolationGroupConfiguration{},
//	}
//	replicationConfig := persistence.DomainReplicationConfig{
//		ActiveClusterName: "cluster0",
//		Clusters: []*persistence.ClusterReplicationConfig{
//			{ClusterName: "cluster0"},
//			{ClusterName: "cluster1"},
//			{ClusterName: "cluster2"},
//		},
//	}
//
//	domain1 := cache.NewDomainCacheEntryForTest(&domainInfo, &domainConfig, true, &replicationConfig, 1, nil)
//
//	nonFailoverDomainUpdate := []*cache.DomainCacheEntry{
//		domain1,
//	}
//
//	tests := map[string]struct{
//		domainCacheEntry []*cache.DomainCacheEntry
//	}{
//		"non-failover domain update": {
//			nonFailoverDomainUpdate,
//		},
//	}
//
//	for name, td := range tests {
//		t.Run(name, func(t *testing.T){
//			t.Fail() // todo
//			he := historyEngineImpl{}
//			he.domainChangeCB(td.domainCacheEntry)
//		})
//	}
//}

//func TestHistoryEngineImpl_CheckingIfDomainsHaveFailedOver(t *testing.T) {
//
//	tests := map[string]struct{
//		domainCacheEntry []*cache.DomainCacheEntry
//	}{}
//
//	for name, td := range tests {
//		t.Run(name, func(t *testing.T){
//			t.Fail() // todo
//			he := historyEngineImpl{}
//			he.
//		})
//	}
//}

func TestGenerateFailoverTasksForDomainCallback(t *testing.T) {

	domainInfo := persistence.DomainInfo{
		ID:         "83b48dab-68cb-4f73-8752-c75d9271977f",
		Name:       "samples-domain",
		OwnerEmail: "test4@test.com",
		Data:       map[string]string{},
	}
	domainConfig := persistence.DomainConfig{
		Retention:             3,
		EmitMetric:            true,
		HistoryArchivalURI:    "file:///tmp/cadence_archival/development",
		VisibilityArchivalURI: "file:///tmp/cadence_vis_archival/development",
		BadBinaries:           types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
		IsolationGroups:       types.IsolationGroupConfiguration{},
	}
	replicationConfig := persistence.DomainReplicationConfig{
		ActiveClusterName: "cluster0",
		Clusters: []*persistence.ClusterReplicationConfig{
			{ClusterName: "cluster0"},
			{ClusterName: "cluster1"},
			{ClusterName: "cluster2"},
		},
	}

	domain1 := cache.NewDomainCacheEntryForTest(&domainInfo, &domainConfig, true, &replicationConfig, 1, nil)

	nonFailoverDomainUpdate := []*cache.DomainCacheEntry{
		domain1,
	}

	tests := map[string]struct {
		domainUpdates []*cache.DomainCacheEntry
		expectedRes   []*persistence.FailoverMarkerTask
		expectedErr   error
	}{
		"non-failover domain update. This should generate no tasks as there's no failover taking place": {
			domainUpdates: nonFailoverDomainUpdate,
			expectedRes:   []*persistence.FailoverMarkerTask{},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {

			cluster := cluster.NewMetadata(
				100,
				"cluster0",
				"cluster0",
				map[string]config.ClusterInformation{
					"cluster0": config.ClusterInformation{
						Enabled:                   true,
						InitialFailoverVersion:    1,
						NewInitialFailoverVersion: nil,
						RPCName:                   "",
						RPCAddress:                "",
						RPCTransport:              "",
						AuthorizationProvider:     config.AuthorizationProvider{},
						TLS:                       config.TLS{},
					},
				},
				func(string) bool { return false },
				metrics.NewNoopMetricsClient(),
				loggerimpl.NewNopLogger(),
			)

			he := historyEngineImpl{
				clusterMetadata: cluster,
			}
			res := he.generateFailoverTasksForDomainUpdateCallback(1, td.domainUpdates)
			assert.Equal(t, td.expectedRes, res)
		})
	}
}
