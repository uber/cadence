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

package rpc

import (
	"net"

	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/yarpc/transport/tchannel"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

// Factory is an implementation of common.RPCFactory interface
type Factory struct {
	logger            log.Logger
	hostAddressMapper HostAddressMapper
	tchannel          *tchannel.Transport
	grpc              *grpc.Transport
	dispatcher        *yarpc.Dispatcher
}

// NewFactory builds a new rpc.Factory
func NewFactory(logger log.Logger, p Params) *Factory {
	inbounds := yarpc.Inbounds{}

	// Create TChannel transport
	// This is here only because ringpop extracts inbound from the dispatcher and expects tchannel.ChannelTransport,
	// everywhere else we use regular tchannel.Transport.
	ch, err := tchannel.NewChannelTransport(
		tchannel.ServiceName(p.ServiceName),
		tchannel.ListenAddr(p.TChannelAddress))
	if err != nil {
		logger.Fatal("Failed to create transport channel", tag.Error(err))
	}
	tchannel, err := tchannel.NewTransport(tchannel.ServiceName(p.ServiceName))
	if err != nil {
		logger.Fatal("Failed to create tchannel transport", tag.Error(err))
	}

	inbounds = append(inbounds, ch.NewInbound())
	logger.Info("Listening for TChannel requests", tag.Address(p.TChannelAddress))

	// Create gRPC transport
	var options []grpc.TransportOption
	if p.GRPCMaxMsgSize > 0 {
		options = append(options, grpc.ServerMaxRecvMsgSize(p.GRPCMaxMsgSize))
		options = append(options, grpc.ClientMaxRecvMsgSize(p.GRPCMaxMsgSize))
	}
	grpc := grpc.NewTransport(options...)
	if len(p.GRPCAddress) > 0 {
		listener, err := net.Listen("tcp", p.GRPCAddress)
		if err != nil {
			logger.Fatal("Failed to listen on GRPC port", tag.Error(err))
		}

		inbounds = append(inbounds, grpc.NewInbound(listener))
		logger.Info("Listening for GRPC requests", tag.Address(p.GRPCAddress))
	}

	// Create outbounds
	outbounds := yarpc.Outbounds{}
	if p.OutboundsBuilder != nil {
		outbounds, err = p.OutboundsBuilder.Build(grpc, tchannel)
		if err != nil {
			logger.Fatal("Failed to create outbounds", tag.Error(err))
		}
	}

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name:               p.ServiceName,
		Inbounds:           inbounds,
		Outbounds:          outbounds,
		InboundMiddleware:  p.InboundMiddleware,
		OutboundMiddleware: p.OutboundMiddleware,
	})

	return &Factory{
		logger:            logger,
		hostAddressMapper: p.HostAddressMapper,
		tchannel:          tchannel,
		grpc:              grpc,
		dispatcher:        dispatcher,
	}
}

// GetDispatcher return a cached dispatcher
func (d *Factory) GetDispatcher() *yarpc.Dispatcher {
	return d.dispatcher
}

// CreateDispatcherForOutbound creates a dispatcher for outbound connection
func (d *Factory) CreateDispatcherForOutbound(
	callerName string,
	serviceName string,
	hostName string,
) (*yarpc.Dispatcher, error) {
	return d.createOutboundDispatcher(callerName, serviceName, hostName, d.tchannel.NewSingleOutbound(hostName))
}

// CreateGRPCDispatcherForOutbound creates a dispatcher for GRPC outbound connection
func (d *Factory) CreateGRPCDispatcherForOutbound(
	callerName string,
	serviceName string,
	hostName string,
) (*yarpc.Dispatcher, error) {
	return d.createOutboundDispatcher(callerName, serviceName, hostName, d.grpc.NewSingleOutbound(hostName))
}

// ReplaceGRPCPort replaces port in the address to grpc for a given service
func (d *Factory) ReplaceGRPCPort(serviceName, hostAddress string) (string, error) {
	return d.hostAddressMapper.GetGRPCAddress(serviceName, hostAddress)
}

func (d *Factory) createOutboundDispatcher(
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
