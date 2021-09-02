// Copyright (c) 2018 Uber Technologies, Inc.
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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination metadata_mock.go

package cluster

import (
	"fmt"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
)

type (
	// Metadata provides information about clusters
	Metadata interface {
		// IsGlobalDomainEnabled whether the global domain is enabled,
		// this attr should be discarded when cross DC is made public
		IsGlobalDomainEnabled() bool
		// IsPrimaryCluster whether current cluster is the primary cluster
		IsPrimaryCluster() bool
		// GetNextFailoverVersion return the next failover version for domain failover
		GetNextFailoverVersion(string, int64) int64
		// IsVersionFromSameCluster return true if 2 version are used for the same cluster
		IsVersionFromSameCluster(version1 int64, version2 int64) bool
		// GetPrimaryClusterName return the primary cluster name
		GetPrimaryClusterName() string
		// GetCurrentClusterName return the current cluster name
		GetCurrentClusterName() string
		// GetAllClusterInfo return the all cluster name -> corresponding info
		GetAllClusterInfo() map[string]config.ClusterInformation
		// ClusterNameForFailoverVersion return the corresponding cluster name for a given failover version
		ClusterNameForFailoverVersion(failoverVersion int64) string
	}

	metadataImpl struct {
		logger log.Logger
		// EnableGlobalDomain whether the global domain is enabled,
		// this attr should be discarded when cross DC is made public
		enableGlobalDomain dynamicconfig.BoolPropertyFn
		// failoverVersionIncrement is the increment of each cluster's version when failover happen
		failoverVersionIncrement int64
		// primaryClusterName is the name of the primary cluster, only the primary cluster can register / update domain
		// all clusters can do domain failover
		primaryClusterName string
		// currentClusterName is the name of the current cluster
		currentClusterName string
		// clusterGroup contains all cluster name -> corresponding information
		clusterGroup map[string]config.ClusterInformation
		// versionToClusterName contains all initial version -> corresponding cluster name
		versionToClusterName map[int64]string
	}
)

// NewMetadata create a new instance of Metadata
func NewMetadata(
	logger log.Logger,
	enableGlobalDomain dynamicconfig.BoolPropertyFn,
	failoverVersionIncrement int64,
	primaryClusterName string,
	currentClusterName string,
	clusterGroup map[string]config.ClusterInformation,
) Metadata {
	versionToClusterName := make(map[int64]string)
	for clusterName, info := range clusterGroup {
		versionToClusterName[info.InitialFailoverVersion] = clusterName
	}

	return &metadataImpl{
		logger:                   logger,
		enableGlobalDomain:       enableGlobalDomain,
		failoverVersionIncrement: failoverVersionIncrement,
		primaryClusterName:       primaryClusterName,
		currentClusterName:       currentClusterName,
		clusterGroup:             clusterGroup,
		versionToClusterName:     versionToClusterName,
	}
}

// IsGlobalDomainEnabled whether the global domain is enabled,
// this attr should be discarded when cross DC is made public
func (metadata *metadataImpl) IsGlobalDomainEnabled() bool {
	return metadata.enableGlobalDomain()
}

// GetNextFailoverVersion return the next failover version based on input
func (metadata *metadataImpl) GetNextFailoverVersion(cluster string, currentFailoverVersion int64) int64 {
	info, ok := metadata.clusterGroup[cluster]
	if !ok {
		panic(fmt.Sprintf(
			"Unknown cluster name: %v with given cluster initial failover version map: %v.",
			cluster,
			metadata.clusterGroup,
		))
	}
	failoverVersion := currentFailoverVersion/metadata.failoverVersionIncrement*metadata.failoverVersionIncrement + info.InitialFailoverVersion
	if failoverVersion < currentFailoverVersion {
		return failoverVersion + metadata.failoverVersionIncrement
	}
	return failoverVersion
}

// IsVersionFromSameCluster return true if 2 version are used for the same cluster
func (metadata *metadataImpl) IsVersionFromSameCluster(version1 int64, version2 int64) bool {
	return (version1-version2)%metadata.failoverVersionIncrement == 0
}

func (metadata *metadataImpl) IsPrimaryCluster() bool {
	return metadata.primaryClusterName == metadata.currentClusterName
}

// GetPrimaryClusterName return the primary cluster name
func (metadata *metadataImpl) GetPrimaryClusterName() string {
	return metadata.primaryClusterName
}

// GetCurrentClusterName return the current cluster name
func (metadata *metadataImpl) GetCurrentClusterName() string {
	return metadata.currentClusterName
}

// GetAllClusterInfo return the all cluster name -> corresponding information
func (metadata *metadataImpl) GetAllClusterInfo() map[string]config.ClusterInformation {
	return metadata.clusterGroup
}

// ClusterNameForFailoverVersion return the corresponding cluster name for a given failover version
func (metadata *metadataImpl) ClusterNameForFailoverVersion(failoverVersion int64) string {
	if failoverVersion == common.EmptyVersion {
		return metadata.currentClusterName
	}

	initialFailoverVersion := failoverVersion % metadata.failoverVersionIncrement
	clusterName, ok := metadata.versionToClusterName[initialFailoverVersion]
	if !ok {
		panic(fmt.Sprintf(
			"Unknown initial failover version %v with given cluster initial failover version map: %v and failover version increment %v.",
			initialFailoverVersion,
			metadata.clusterGroup,
			metadata.failoverVersionIncrement,
		))
	}
	return clusterName
}
