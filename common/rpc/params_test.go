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
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/service"
)

func TestNewParams(t *testing.T) {
	serviceName := service.Frontend
	dc := dynamicconfig.NewNopCollection()
	makeConfig := func(svc config.Service) *config.Config {
		return &config.Config{
			PublicClient: config.PublicClient{HostPort: "localhost:9999"},
			Services:     map[string]config.Service{"frontend": svc}}
	}

	_, err := NewParams(serviceName, &config.Config{}, dc)
	assert.EqualError(t, err, "no config section for service: frontend")

	_, err = NewParams(serviceName, makeConfig(config.Service{RPC: config.RPC{BindOnLocalHost: true, BindOnIP: "1.2.3.4"}}), dc)
	assert.EqualError(t, err, "get listen IP: bindOnLocalHost and bindOnIP are mutually exclusive")

	_, err = NewParams(serviceName, makeConfig(config.Service{RPC: config.RPC{BindOnIP: "invalidIP"}}), dc)
	assert.EqualError(t, err, "get listen IP: unable to parse bindOnIP value or it is not an IPv4 or IPv6 address: invalidIP")

	_, err = NewParams(serviceName, &config.Config{Services: map[string]config.Service{"frontend": {}}}, dc)
	assert.EqualError(t, err, "public client outbound: need to provide an endpoint config for PublicClient")

	_, err = NewParams(serviceName, makeConfig(config.Service{RPC: config.RPC{BindOnLocalHost: true, TLS: config.TLS{Enabled: true, CertFile: "invalid", KeyFile: "invalid"}}}), dc)
	assert.EqualError(t, err, "inbound TLS config: open invalid: no such file or directory")

	_, err = NewParams(serviceName, &config.Config{Services: map[string]config.Service{
		"frontend": {RPC: config.RPC{BindOnLocalHost: true}},
		"history":  {RPC: config.RPC{TLS: config.TLS{Enabled: true, CaFile: "invalid"}}},
	}}, dc)
	assert.EqualError(t, err, "outbound cadence-history TLS config: open invalid: no such file or directory")

	params, err := NewParams(serviceName, makeConfig(config.Service{RPC: config.RPC{BindOnLocalHost: true, Port: 1111, GRPCPort: 2222, GRPCMaxMsgSize: 3333}}), dc)
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:1111", params.TChannelAddress)
	assert.Equal(t, "127.0.0.1:2222", params.GRPCAddress)
	assert.Equal(t, 3333, params.GRPCMaxMsgSize)
	assert.Nil(t, params.InboundTLS)
	assert.IsType(t, GRPCPorts{}, params.HostAddressMapper)

	params, err = NewParams(serviceName, makeConfig(config.Service{RPC: config.RPC{BindOnIP: "1.2.3.4", GRPCPort: 2222}}), dc)
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3.4:2222", params.GRPCAddress)

	params, err = NewParams(serviceName, makeConfig(config.Service{RPC: config.RPC{GRPCPort: 2222, TLS: config.TLS{Enabled: true}}}), dc)
	assert.NoError(t, err)
	ip, port, err := net.SplitHostPort(params.GRPCAddress)
	assert.NoError(t, err)
	assert.Equal(t, "2222", port)
	assert.NotNil(t, net.ParseIP(ip))
	assert.NotNil(t, params.InboundTLS)
}
