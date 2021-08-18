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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterMetadataDefaults(t *testing.T) {
	config := ClusterMetadata{
		MasterClusterName: "active",
		ClusterInformation: map[string]ClusterInformation{
			"active": {},
		},
	}

	config.fillDefaults()

	assert.Equal(t, "active", config.PrimaryClusterName)
	assert.Equal(t, "cadence-frontend", config.ClusterInformation["active"].RPCName)
}

func TestClusterMetadataValidate(t *testing.T) {
	tests := []struct {
		msg    string
		config *ClusterMetadata
		err    string
	}{
		{
			msg:    "valid",
			config: validClusterMetadata(),
		},
		{
			msg:    "empty",
			config: nil,
			err:    "ClusterMetadata cannot be empty",
		},
		{
			msg: "primary cluster name is empty",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				m.PrimaryClusterName = ""
			}),
			err: "primary cluster name is empty",
		},
		{
			msg: "current cluster name is empty",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				m.CurrentClusterName = ""
			}),
			err: "current cluster name is empty",
		},
		{
			msg: "primary cluster is not specified in cluster info",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				m.PrimaryClusterName = "non-existing"
			}),
			err: "primary cluster is not specified in cluster info",
		},
		{
			msg: "current cluster is not specified in cluster info",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				m.CurrentClusterName = "non-existing"
			}),
			err: "current cluster is not specified in cluster info",
		},
		{
			msg: "version increment is 0",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				m.FailoverVersionIncrement = 0
			}),
			err: "version increment is 0",
		},
		{
			msg: "empty cluster information",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				m.ClusterInformation = nil
			}),
			err: "empty cluster information",
		},
		{
			msg: "cluster with empty name defined",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				m.ClusterInformation[""] = ClusterInformation{}
			}),
			err: "cluster with empty name defined",
		},
		{
			msg: "increment smaller than initial version",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				m.FailoverVersionIncrement = 1
			}),
			err: "cluster standby: version increment 1 is smaller than initial version: 2",
		},
		{
			msg: "empty rpc name",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				active := m.ClusterInformation["active"]
				active.RPCName = ""
				m.ClusterInformation["active"] = active
			}),
			err: "cluster active: rpc name / address is empty",
		},
		{
			msg: "empty rpc address",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				active := m.ClusterInformation["active"]
				active.RPCAddress = ""
				m.ClusterInformation["active"] = active
			}),
			err: "cluster active: rpc name / address is empty",
		},
		{
			msg: "initial version duplicated",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				standby := m.ClusterInformation["standby"]
				standby.InitialFailoverVersion = 0
				m.ClusterInformation["standby"] = standby
			}),
			err: "cluster info initial versions have duplicates",
		},
		{
			msg: "multiple errors",
			config: modifyClusterMetadata(validClusterMetadata(), func(m *ClusterMetadata) {
				*m = ClusterMetadata{}
			}),
			err: "primary cluster name is empty; current cluster name is empty; version increment is 0; empty cluster information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			err := tt.config.validate()
			if tt.err == "" {
				assert.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
		})
	}
}

func validClusterMetadata() *ClusterMetadata {
	return &ClusterMetadata{
		EnableGlobalDomain:       true,
		FailoverVersionIncrement: 10,
		PrimaryClusterName:       "active",
		CurrentClusterName:       "standby",
		ClusterInformation: map[string]ClusterInformation{
			"active": {
				Enabled:                true,
				InitialFailoverVersion: 0,
				RPCName:                "cadence-frontend",
				RPCAddress:             "localhost:7933",
			},
			"standby": {
				Enabled:                true,
				InitialFailoverVersion: 2,
				RPCName:                "cadence-frontend",
				RPCAddress:             "localhost:7933",
			},
		},
	}
}

func modifyClusterMetadata(initial *ClusterMetadata, modify func(metadata *ClusterMetadata)) *ClusterMetadata {
	modify(initial)
	return initial
}
