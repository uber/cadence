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

package rpc

import (
	"testing"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/service"

	"github.com/stretchr/testify/assert"
)

func TestGRPCPorts(t *testing.T) {
	config := config.Config{
		Services: map[string]config.Service{
			"frontend": {RPC: config.RPC{GRPCPort: 9999}},
			"history":  {RPC: config.RPC{}},
		},
	}
	grpcPorts := NewGRPCPorts(&config)

	_, err := grpcPorts.GetGRPCAddress("some-service", "1.2.3.4")
	assert.EqualError(t, err, "unknown service: some-service")

	_, err = grpcPorts.GetGRPCAddress(service.History, "1.2.3.4")
	assert.EqualError(t, err, "GRPC port not configured for service: cadence-history")

	grpcAddress, err := grpcPorts.GetGRPCAddress(service.Frontend, "1.2.3.4")
	assert.Nil(t, err)
	assert.Equal(t, grpcAddress, "1.2.3.4:9999")

	grpcAddress, err = grpcPorts.GetGRPCAddress(service.Frontend, "1.2.3.4:8888")
	assert.Nil(t, err)
	assert.Equal(t, grpcAddress, "1.2.3.4:9999")
}
