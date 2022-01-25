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

func TestClusterGroupMetadataDefaults(t *testing.T) {
	config := ClusterGroupMetadata{
		MasterClusterName: "active",
		ClusterInformation: map[string]ClusterInformation{
			"active": {},
		},
	}

	config.FillDefaults()

	assert.Equal(t, "active", config.PrimaryClusterName)
	assert.Equal(t, "cadence-frontend", config.ClusterGroup["active"].RPCName)
	assert.Equal(t, "tchannel", config.ClusterGroup["active"].RPCTransport)
}

func TestClusterGroupMetadataValidate(t *testing.T) {
	modify := func(initial *ClusterGroupMetadata, modify func(metadata *ClusterGroupMetadata)) *ClusterGroupMetadata {
		modify(initial)
		return initial
	}

	tests := []struct {
		msg    string
		config *ClusterGroupMetadata
		err    string
	}{
		{
			msg:    "valid",
			config: validClusterGroupMetadata(),
		},
		{
			msg:    "empty",
			config: nil,
			err:    "ClusterGroupMetadata cannot be empty",
		},
		{
			msg: "primary cluster name is empty",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				m.PrimaryClusterName = ""
			}),
			err: "primary cluster name is empty",
		},
		{
			msg: "current cluster name is empty",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				m.CurrentClusterName = ""
			}),
			err: "current cluster name is empty",
		},
		{
			msg: "primary cluster is not specified in the cluster group",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				m.PrimaryClusterName = "non-existing"
			}),
			err: "primary cluster is not specified in the cluster group",
		},
		{
			msg: "current cluster is not specified in the cluster group",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				m.CurrentClusterName = "non-existing"
			}),
			err: "current cluster is not specified in the cluster group",
		},
		{
			msg: "version increment is 0",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				m.FailoverVersionIncrement = 0
			}),
			err: "version increment is 0",
		},
		{
			msg: "empty cluster group",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				m.ClusterGroup = nil
			}),
			err: "empty cluster group",
		},
		{
			msg: "cluster with empty name defined",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				m.ClusterGroup[""] = ClusterInformation{}
			}),
			err: "cluster with empty name defined",
		},
		{
			msg: "increment smaller than initial version",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				m.FailoverVersionIncrement = 1
			}),
			err: "cluster standby: version increment 1 is smaller than initial version: 2",
		},
		{
			msg: "empty rpc name",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				active := m.ClusterGroup["active"]
				active.RPCName = ""
				m.ClusterGroup["active"] = active
			}),
			err: "cluster active: rpc name / address is empty",
		},
		{
			msg: "empty rpc address",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				active := m.ClusterGroup["active"]
				active.RPCAddress = ""
				m.ClusterGroup["active"] = active
			}),
			err: "cluster active: rpc name / address is empty",
		},
		{
			msg: "invalid rpc transport",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				active := m.ClusterGroup["active"]
				active.RPCTransport = "invalid"
				m.ClusterGroup["active"] = active
			}),
			err: "cluster active: rpc transport must tchannel or grpc",
		},
		{
			msg: "initial version duplicated",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				standby := m.ClusterGroup["standby"]
				standby.InitialFailoverVersion = 0
				m.ClusterGroup["standby"] = standby
			}),
			err: "initial versions of the cluster group have duplicates",
		},
		{
			msg: "multiple errors",
			config: modify(validClusterGroupMetadata(), func(m *ClusterGroupMetadata) {
				*m = ClusterGroupMetadata{}
			}),
			err: "primary cluster name is empty; current cluster name is empty; version increment is 0; empty cluster group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.err == "" {
				assert.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
		})
	}
}

func validClusterGroupMetadata() *ClusterGroupMetadata {
	return &ClusterGroupMetadata{
		EnableGlobalDomain:       true,
		FailoverVersionIncrement: 10,
		PrimaryClusterName:       "active",
		CurrentClusterName:       "standby",
		ClusterGroup: map[string]ClusterInformation{
			"active": {
				Enabled:                true,
				InitialFailoverVersion: 0,
				RPCName:                "cadence-frontend",
				RPCAddress:             "localhost:7833",
				RPCTransport:           "grpc",
			},
			"standby": {
				Enabled:                true,
				InitialFailoverVersion: 2,
				RPCName:                "cadence-frontend",
				RPCAddress:             "localhost:7833",
				RPCTransport:           "grpc",
			},
		},
	}
}
