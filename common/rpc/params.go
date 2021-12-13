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
	"crypto/tls"
	"fmt"
	"net"
	"strconv"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/service"

	"go.uber.org/yarpc"
)

// Params allows to configure rpc.Factory
type Params struct {
	ServiceName       string
	TChannelAddress   string
	GRPCAddress       string
	GRPCMaxMsgSize    int
	HostAddressMapper HostAddressMapper

	InboundTLS  *tls.Config
	OutboundTLS map[string]*tls.Config

	InboundMiddleware  yarpc.InboundMiddleware
	OutboundMiddleware yarpc.OutboundMiddleware

	OutboundsBuilder OutboundsBuilder
}

// NewParams creates parameters for rpc.Factory from the given config
func NewParams(serviceName string, config *config.Config, dc *dynamicconfig.Collection) (Params, error) {
	serviceConfig, err := config.GetServiceConfig(serviceName)
	if err != nil {
		return Params{}, err
	}

	listenIP, err := getListenIP(serviceConfig.RPC)
	if err != nil {
		return Params{}, fmt.Errorf("get listen IP: %v", err)
	}

	inboundTLS, err := serviceConfig.RPC.TLS.ToTLSConfig()
	if err != nil {
		return Params{}, fmt.Errorf("inbound TLS config: %v", err)
	}
	outboundTLS := map[string]*tls.Config{}
	for _, outboundServiceName := range service.List {
		outboundServiceConfig, err := config.GetServiceConfig(outboundServiceName)
		if err != nil {
			continue
		}
		outboundTLS[outboundServiceName], err = outboundServiceConfig.RPC.TLS.ToTLSConfig()
		if err != nil {
			return Params{}, fmt.Errorf("outbound %s TLS config: %v", outboundServiceName, err)
		}
	}

	enableGRPCOutbound := dc.GetBoolProperty(dynamicconfig.EnableGRPCOutbound, true)()

	publicClientOutbound, err := newPublicClientOutbound(config)
	if err != nil {
		return Params{}, fmt.Errorf("public client outbound: %v", err)
	}

	return Params{
		ServiceName:       serviceName,
		TChannelAddress:   net.JoinHostPort(listenIP.String(), strconv.Itoa(serviceConfig.RPC.Port)),
		GRPCAddress:       net.JoinHostPort(listenIP.String(), strconv.Itoa(serviceConfig.RPC.GRPCPort)),
		GRPCMaxMsgSize:    serviceConfig.RPC.GRPCMaxMsgSize,
		HostAddressMapper: NewGRPCPorts(config),
		OutboundsBuilder: CombineOutbounds(
			NewDirectOutbound(service.History, enableGRPCOutbound, outboundTLS[service.History]),
			NewDirectOutbound(service.Matching, enableGRPCOutbound, outboundTLS[service.Matching]),
			publicClientOutbound,
		),
		InboundTLS:  inboundTLS,
		OutboundTLS: outboundTLS,
		InboundMiddleware: yarpc.InboundMiddleware{
			Unary: &InboundMetricsMiddleware{},
		},
		OutboundMiddleware: yarpc.OutboundMiddleware{
			Unary: &HeaderForwardingMiddleware{},
		},
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
		if ip != nil && ip.To16() != nil {
			return ip.To16(), nil
		}
		return nil, fmt.Errorf("unable to parse bindOnIP value or it is not an IPv4 or IPv6 address: %s", config.BindOnIP)
	}
	return ListenIP()
}
