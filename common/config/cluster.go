// Copyright (c) 2021 Uber Technologies, Inc.
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

package config

import (
	"errors"
	"fmt"
	"log"

	"go.uber.org/multierr"
)

type (
	// ClusterGroupMetadata contains all the clusters participating in a replication group(aka XDC/GlobalDomain)
	ClusterGroupMetadata struct {
		EnableGlobalDomain bool `yaml:"enableGlobalDomain"`
		// FailoverVersionIncrement is the increment of each cluster version when failover happens
		// It decides the maximum number clusters in this replication groups
		FailoverVersionIncrement int64 `yaml:"failoverVersionIncrement"`
		// PrimaryClusterName is the primary cluster name, only the primary cluster can register / update domain
		// all clusters can do domain failover
		PrimaryClusterName string `yaml:"primaryClusterName"`
		// MasterClusterName is deprecated. Please use PrimaryClusterName.
		MasterClusterName string `yaml:"masterClusterName"`
		// CurrentClusterName is the name of the cluster of current deployment
		CurrentClusterName string `yaml:"currentClusterName"`
		// ClusterGroup contains information for each cluster within the replication group
		// Key is the clusterName
		ClusterGroup map[string]ClusterInformation `yaml:"clusterGroup"`
		// ClusterInformation is deprecated. Please use ClusterGroup.
		ClusterInformation map[string]ClusterInformation `yaml:"clusterInformation"`
	}

	// ClusterInformation contains the information about each cluster participating in cross DC
	ClusterInformation struct {
		Enabled bool `yaml:"enabled"`
		// InitialFailoverVersion is the identifier of each cluster. 0 <= the value < failoverVersionIncrement
		InitialFailoverVersion int64 `yaml:"initialFailoverVersion"`
		// RPCName indicate the remote service name
		RPCName string `yaml:"rpcName"`
		// Address indicate the remote service address(Host:Port). Host can be DNS name.
		// For currentCluster, it's usually the same as publicClient.hostPort
		RPCAddress string `yaml:"rpcAddress" validate:"nonzero"`
	}
)

func (m *ClusterGroupMetadata) validate() error {
	if m == nil {
		return errors.New("ClusterGroupMetadata cannot be empty")
	}

	if !m.EnableGlobalDomain {
		log.Println("[WARN] Local domain is now deprecated. Please update config to enable global domain(ClusterGroupMetadata->EnableGlobalDomain)." +
			"Global domain of single cluster has zero overhead, but only advantages for future migration and fail over. Please check Cadence documentation for more details.")
	}

	var errs error

	if len(m.PrimaryClusterName) == 0 {
		errs = multierr.Append(errs, errors.New("primary cluster name is empty"))
	}
	if len(m.CurrentClusterName) == 0 {
		errs = multierr.Append(errs, errors.New("current cluster name is empty"))
	}
	if m.FailoverVersionIncrement == 0 {
		errs = multierr.Append(errs, errors.New("version increment is 0"))
	}
	if len(m.ClusterGroup) == 0 {
		errs = multierr.Append(errs, errors.New("empty cluster group"))
	}
	if _, ok := m.ClusterGroup[m.PrimaryClusterName]; len(m.PrimaryClusterName) > 0 && !ok {
		errs = multierr.Append(errs, errors.New("primary cluster is not specified in the cluster group"))
	}
	if _, ok := m.ClusterGroup[m.CurrentClusterName]; len(m.CurrentClusterName) > 0 && !ok {
		errs = multierr.Append(errs, errors.New("current cluster is not specified in the cluster group"))
	}

	versionToClusterName := make(map[int64]string)
	for clusterName, info := range m.ClusterGroup {
		if len(clusterName) == 0 {
			errs = multierr.Append(errs, errors.New("cluster with empty name defined"))
		}
		versionToClusterName[info.InitialFailoverVersion] = clusterName

		if m.FailoverVersionIncrement <= info.InitialFailoverVersion || info.InitialFailoverVersion < 0 {
			errs = multierr.Append(errs, fmt.Errorf(
				"cluster %s: version increment %v is smaller than initial version: %v",
				clusterName,
				m.FailoverVersionIncrement,
				info.InitialFailoverVersion,
			))
		}

		if info.Enabled && (len(info.RPCName) == 0 || len(info.RPCAddress) == 0) {
			errs = multierr.Append(errs, fmt.Errorf("cluster %v: rpc name / address is empty", clusterName))
		}
	}
	if len(versionToClusterName) != len(m.ClusterGroup) {
		errs = multierr.Append(errs, errors.New("initial versions of the cluster group have duplicates"))
	}

	return errs
}

func (m *ClusterGroupMetadata) fillDefaults() {
	if m == nil {
		return
	}

	// TODO: remove this after 0.23 and mention a breaking change in config.
	if len(m.PrimaryClusterName) == 0 && len(m.MasterClusterName) != 0 {
		m.PrimaryClusterName = m.MasterClusterName
		log.Println("[WARN] masterClusterName config is deprecated. Please replace it with primaryClusterName.")
	}

	// TODO: remove this after 0.23 and mention a breaking change in config.
	if len(m.ClusterGroup) == 0 && len(m.ClusterInformation) != 0 {
		m.ClusterGroup = m.ClusterInformation
		log.Println("[WARN] clusterInformation config is deprecated. Please replace it with clusterGroup.")
	}

	for name, cluster := range m.ClusterGroup {
		if cluster.RPCName == "" {
			// filling RPCName with a default value if empty
			cluster.RPCName = "cadence-frontend"
			m.ClusterGroup[name] = cluster
		}
	}
}
