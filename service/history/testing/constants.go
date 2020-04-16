// Copyright (c) 2020 Uber Technologies, Inc.
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

package testing

import (
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/persistence"
)

var (
	Version          = int64(1234)
	DomainID         = "deadbeef-0123-4567-890a-bcdef0123456"
	DomainName       = "some random domain name"
	ParentDomainID   = "deadbeef-0123-4567-890a-bcdef0123457"
	ParentDomainName = "some random parent domain name"
	TargetDomainID   = "deadbeef-0123-4567-890a-bcdef0123458"
	TargetDomainName = "some random target domain name"
	ChildDomainID    = "deadbeef-0123-4567-890a-bcdef0123459"
	ChildDomainName  = "some random child domain name"
	WorkflowID       = "random-workflow-id"
	RunID            = "0d00698f-08e1-4d36-a3e2-3bf109f5d2d6"

	LocalDomainEntry = cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: DomainID, Name: DomainName},
		&persistence.DomainConfig{Retention: 1},
		cluster.TestCurrentClusterName,
		nil,
	)

	GlobalDomainEntry = cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: DomainID, Name: DomainName},
		&persistence.DomainConfig{
			Retention:                1,
			VisibilityArchivalStatus: workflow.ArchivalStatusEnabled,
			VisibilityArchivalURI:    "test:///visibility/archival",
		},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		Version,
		nil,
	)

	GlobalParentDomainEntry = cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: ParentDomainID, Name: ParentDomainName},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		Version,
		nil,
	)

	GlobalTargetDomainEntry = cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: TargetDomainID, Name: TargetDomainName},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		Version,
		nil,
	)

	GlobalChildDomainEntry = cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: ChildDomainID, Name: ChildDomainName},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		Version,
		nil,
	)
)
