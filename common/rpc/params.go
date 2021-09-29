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
	"fmt"
	"net"

	"github.com/uber/cadence/common/config"

	"go.uber.org/yarpc"
)

// Params allows to configure rpc.Factory
type Params struct {
	ServiceName       string
	TChannelAddress   string
	GRPCAddress       string
	GRPCMaxMsgSize    int
	HostAddressMapper HostAddressMapper

	InboundMiddleware  yarpc.InboundMiddleware
	OutboundMiddleware yarpc.OutboundMiddleware

	OutboundsBuilder OutboundsBuilder
}

// NewParams creates parameters for rpc.Factory from the given config
func NewParams(serviceName string, config *config.Config) (Params, error) {
	serviceConfig, err := config.GetServiceConfig(serviceName)
	if err != nil {
		return Params{}, err
	}

	listenIP, err := getListenIP(serviceConfig.RPC)
	if err != nil {
		return Params{}, fmt.Errorf("failed to get listen IP: %v", err)
	}
	return Params{
		ServiceName:       serviceName,
		TChannelAddress:   fmt.Sprintf("%v:%v", listenIP, serviceConfig.RPC.Port),
		GRPCAddress:       fmt.Sprintf("%v:%v", listenIP, serviceConfig.RPC.GRPCPort),
		GRPCMaxMsgSize:    serviceConfig.RPC.GRPCMaxMsgSize,
		HostAddressMapper: NewGRPCPorts(config),
	}, nil
}

func getListenIP(config config.RPC) (net.IP, error) {
	if config.BindOnLocalHost && len(config.BindOnIP) > 0 {
		return nil, fmt.Errorf("bindOnLocalHost and bindOnIP are mutually exclusive")
	}

	if config.BindOnLocalHost {
		return net.IPv4(127, 0, 0, 1), nil
	}

	if len(config.BindOnIP) > 0 {
		ip := net.ParseIP(config.BindOnIP)
		if ip != nil && ip.To4() != nil {
			return ip.To4(), nil
		}
		return nil, fmt.Errorf("unable to parse bindOnIP value or it is not an IPv4 address: %s", config.BindOnIP)
	}
	return ListenIP()
}
