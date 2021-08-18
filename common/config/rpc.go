// Copyright (c) 2017 Uber Technologies, Inc.
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
	"net"
	"strings"
	"sync"

	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/yarpc/transport/tchannel"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

// RPCFactory is an implementation of service.RPCFactory interface
type RPCFactory struct {
	config      *RPC
	serviceName string
	ch          *tchannel.ChannelTransport
	grpc        *grpc.Transport
	logger      log.Logger
	grpcPorts   GRPCPorts

	sync.Mutex
	dispatcher *yarpc.Dispatcher
}

// NewFactory builds a new RPCFactory
// conforming to the underlying configuration
func (cfg *RPC) NewFactory(sName string, logger log.Logger, grpcPorts GRPCPorts) *RPCFactory {
	return newRPCFactory(cfg, sName, logger, grpcPorts)
}

func newRPCFactory(cfg *RPC, sName string, logger log.Logger, grpcPorts GRPCPorts) *RPCFactory {
	factory := &RPCFactory{config: cfg, serviceName: sName, logger: logger, grpcPorts: grpcPorts}
	return factory
}

// GetDispatcher return a cached dispatcher
func (d *RPCFactory) GetDispatcher() *yarpc.Dispatcher {
	d.Lock()
	defer d.Unlock()

	if d.dispatcher != nil {
		return d.dispatcher
	}

	d.dispatcher = d.createInboundDispatcher()
	return d.dispatcher
}

// createInboundDispatcher creates a dispatcher for inbound
func (d *RPCFactory) createInboundDispatcher() *yarpc.Dispatcher {
	// Setup dispatcher for onebox
	var err error
	inbounds := yarpc.Inbounds{}

	hostAddress := fmt.Sprintf("%v:%v", d.getListenIP(), d.config.Port)
	d.ch, err = tchannel.NewChannelTransport(
		tchannel.ServiceName(d.serviceName),
		tchannel.ListenAddr(hostAddress))
	if err != nil {
		d.logger.Fatal("Failed to create transport channel", tag.Error(err))
	}
	inbounds = append(inbounds, d.ch.NewInbound())
	d.logger.Info("Listening for TChannel requests", tag.Address(hostAddress))

	var options []grpc.TransportOption
	if d.config.GRPCMaxMsgSize > 0 {
		options = append(options, grpc.ServerMaxRecvMsgSize(d.config.GRPCMaxMsgSize))
		options = append(options, grpc.ClientMaxRecvMsgSize(d.config.GRPCMaxMsgSize))
	}
	d.grpc = grpc.NewTransport(options...)
	if d.config.GRPCPort > 0 {
		grpcAddress := fmt.Sprintf("%v:%v", d.getListenIP(), d.config.GRPCPort)
		listener, err := net.Listen("tcp", grpcAddress)
		if err != nil {
			d.logger.Fatal("Failed to listen on GRPC port", tag.Error(err))
		}

		inbounds = append(inbounds, d.grpc.NewInbound(listener))
		d.logger.Info("Listening for GRPC requests", tag.Address(grpcAddress))
	}

	return yarpc.NewDispatcher(yarpc.Config{
		Name:     d.serviceName,
		Inbounds: inbounds,
	})
}

// CreateDispatcherForOutbound creates a dispatcher for outbound connection
func (d *RPCFactory) CreateDispatcherForOutbound(
	callerName string,
	serviceName string,
	hostName string,
) (*yarpc.Dispatcher, error) {
	return d.createOutboundDispatcher(callerName, serviceName, hostName, d.ch.NewSingleOutbound(hostName))
}

// CreateGRPCDispatcherForOutbound creates a dispatcher for GRPC outbound connection
func (d *RPCFactory) CreateGRPCDispatcherForOutbound(
	callerName string,
	serviceName string,
	hostName string,
) (*yarpc.Dispatcher, error) {
	return d.createOutboundDispatcher(callerName, serviceName, hostName, d.grpc.NewSingleOutbound(hostName))
}

// ReplaceGRPCPort replaces port in the address to grpc for a given service
func (d *RPCFactory) ReplaceGRPCPort(serviceName, hostAddress string) (string, error) {
	return d.grpcPorts.GetGRPCAddress(serviceName, hostAddress)
}

func (d *RPCFactory) createOutboundDispatcher(
	callerName string,
	serviceName string,
	hostName string,
	outbound transport.UnaryOutbound,
) (*yarpc.Dispatcher, error) {

	// Setup dispatcher(outbound) for onebox
	d.logger.Info("Created RPC dispatcher outbound", tag.Address(hostName))
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: callerName,
		Outbounds: yarpc.Outbounds{
			serviceName: {Unary: outbound},
		},
	})
	if err := dispatcher.Start(); err != nil {
		d.logger.Error("Failed to create outbound transport channel", tag.Error(err))
		return nil, err
	}
	return dispatcher, nil
}

func (d *RPCFactory) getListenIP() net.IP {
	if d.config.BindOnLocalHost && len(d.config.BindOnIP) > 0 {
		d.logger.Fatal("ListenIP failed, bindOnLocalHost and bindOnIP are mutually exclusive")
	}

	if d.config.BindOnLocalHost {
		return net.IPv4(127, 0, 0, 1)
	}

	if len(d.config.BindOnIP) > 0 {
		ip := net.ParseIP(d.config.BindOnIP)
		if ip != nil && ip.To4() != nil {
			return ip.To4()
		}
		d.logger.Fatal("ListenIP failed, unable to parse bindOnIP value or it is not an IPv4 address", tag.Address(d.config.BindOnIP))
	}
	ip, err := ListenIP()
	if err != nil {
		d.logger.Fatal("ListenIP failed", tag.Error(err))
	}
	return ip
}

type GRPCPorts map[string]int

func (c *Config) NewGRPCPorts() GRPCPorts {
	grpcPorts := map[string]int{}
	for service, config := range c.Services {
		grpcPorts["cadence-"+service] = config.RPC.GRPCPort
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
	if index := strings.Index(hostAddress, ":"); index > 0 {
		hostAddress = hostAddress[:index]
	}

	return fmt.Sprintf("%s:%d", hostAddress, port), nil
}
