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
	"errors"
	"net"
	"strconv"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/service"
)

type (
	HostAddressMapper interface {
		GetGRPCAddress(service, hostAddress string) (string, error)
	}

	GRPCPorts map[string]int
)

func NewGRPCPorts(c *config.Config) GRPCPorts {
	grpcPorts := map[string]int{}
	for name, config := range c.Services {
		grpcPorts[service.FullName(name)] = config.RPC.GRPCPort
	}
	return grpcPorts
}

func (p GRPCPorts) GetGRPCAddress(service, hostAddress string) (string, error) {
	port, ok := p[service]
	if !ok {
		return hostAddress, errors.New("unknown service: " + service)
	}
	if port == 0 {
		return hostAddress, errors.New("GRPC port not configured for service: " + service)
	}

	// Drop port if provided
	if newHostAddress, _, err := net.SplitHostPort(hostAddress); err == nil {
		hostAddress = newHostAddress
	}

	return net.JoinHostPort(hostAddress, strconv.Itoa(port)), nil
}
