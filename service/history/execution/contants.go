package execution

import (
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/persistence"
)

var testVersion = int64(1234)
var testDomainID = "deadbeef-0123-4567-890a-bcdef0123456"
var testDomainName = "some random domain name"
var testRunID = "0d00698f-08e1-4d36-a3e2-3bf109f5d2d6"
var testLocalDomainEntry = cache.NewLocalDomainCacheEntryForTest(
	&persistence.DomainInfo{ID: testDomainID, Name: testDomainName},
	&persistence.DomainConfig{Retention: 1},
	cluster.TestCurrentClusterName,
	nil,
)
var testGlobalDomainEntry = cache.NewGlobalDomainCacheEntryForTest(
	&persistence.DomainInfo{ID: testDomainID, Name: testDomainName},
	&persistence.DomainConfig{
		Retention:                1,
		VisibilityArchivalStatus: shared.ArchivalStatusEnabled,
		VisibilityArchivalURI:    "test:///visibility/archival",
	},
	&persistence.DomainReplicationConfig{
		ActiveClusterName: cluster.TestCurrentClusterName,
		Clusters: []*persistence.ClusterReplicationConfig{
			{ClusterName: cluster.TestCurrentClusterName},
			{ClusterName: cluster.TestAlternativeClusterName},
		},
	},
	testVersion,
	nil,
)
