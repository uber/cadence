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

package gocql

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/config"
)

func Test_GetRegisteredClient(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	GetRegisteredClient()
}

func Test_GetRegisteredClientNotNil(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	registered = NewMockClient(mockCtrl)
	assert.NotNil(t, GetRegisteredClient())
}

func Test_RegisterClient(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	RegisterClient(nil)
}

func Test_RegisterClientNotNil(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	newClient := NewMockClient(mockCtrl)
	registered = nil
	RegisterClient(newClient)
}

func Test_newCassandraCluster(t *testing.T) {
	testFullConfig := ClusterConfig{
		Hosts:      "testHost",
		Port:       123,
		User:       "testUser",
		Password:   "testPassword",
		Keyspace:   "testKeyspace",
		Datacenter: "testDatacenter",
		Region:     "testRegion",
		TLS: &config.TLS{
			Enabled: true,
		},
		MaxConns: 10,
	}
	clusterConfig := newCassandraCluster(testFullConfig)
	assert.NotNil(t, clusterConfig)
}

func Test_regionHostFilter(t *testing.T) {
	region := "testRegion"
	cluster := regionHostFilter(region)
	assert.NotNil(t, cluster)
}
