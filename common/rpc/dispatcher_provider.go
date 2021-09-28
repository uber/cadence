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
	"context"

	clientworker "go.uber.org/cadence/worker"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"

	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/yarpc/transport/tchannel"
)

const (
	crossDCCaller = "cadence-xdc-client"
)

type (
	DispatcherOptions struct {
		AuthProvider clientworker.AuthorizationProvider
	}

	// DispatcherProvider provides a dispatcher to a given address
	DispatcherProvider interface {
		GetTChannel(name string, address string, options *DispatcherOptions) (*yarpc.Dispatcher, error)
		GetGRPC(name string, address string, options *DispatcherOptions) (*yarpc.Dispatcher, error)
	}

	dispatcherProvider struct {
		pcf    PeerChooserFactory
		logger log.Logger
	}
)

// NewDispatcherProvider create a dispatcher provider which handles with IP address
func NewDispatcherProvider(logger log.Logger, pcf PeerChooserFactory) DispatcherProvider {
	return &dispatcherProvider{
		pcf:    pcf,
		logger: logger,
	}
}

func (p *dispatcherProvider) GetTChannel(serviceName string, address string, options *DispatcherOptions) (*yarpc.Dispatcher, error) {
	tchanTransport, err := tchannel.NewTransport(
		tchannel.ServiceName(serviceName),
		// this aim to get rid of the annoying popup about accepting incoming network connections
		tchannel.ListenAddr("127.0.0.1:0"),
	)
	if err != nil {
		return nil, err
	}

	peerChooser, err := p.pcf.CreatePeerChooser(tchanTransport, address)
	if err != nil {
		return nil, err
	}
	outbound := tchanTransport.NewOutbound(peerChooser)

	p.logger.Info("Creating TChannel dispatcher outbound", tag.Address(address))
	return p.createOutboundDispatcher(serviceName, outbound, options)
}

func (p *dispatcherProvider) GetGRPC(serviceName string, address string, options *DispatcherOptions) (*yarpc.Dispatcher, error) {
	grpcTransport := grpc.NewTransport()

	peerChooser, err := p.pcf.CreatePeerChooser(grpcTransport, address)
	if err != nil {
		return nil, err
	}
	outbound := grpcTransport.NewOutbound(peerChooser)

	p.logger.Info("Creating GRPC dispatcher outbound", tag.Address(address))
	return p.createOutboundDispatcher(serviceName, outbound, options)
}

func (p *dispatcherProvider) createOutboundDispatcher(serviceName string, outbound transport.UnaryOutbound, options *DispatcherOptions) (*yarpc.Dispatcher, error) {
	cfg := yarpc.Config{
		Name: crossDCCaller,
		Outbounds: yarpc.Outbounds{
			serviceName: transport.Outbounds{
				Unary:       outbound,
				ServiceName: serviceName,
			},
		},
	}
	if options != nil && options.AuthProvider != nil {
		cfg.OutboundMiddleware = yarpc.OutboundMiddleware{
			Unary: &outboundMiddleware{authProvider: options.AuthProvider},
		}
	}

	// Attach the outbound to the dispatcher (this will add middleware/logging/etc)
	dispatcher := yarpc.NewDispatcher(cfg)

	if err := dispatcher.Start(); err != nil {
		return nil, err
	}
	return dispatcher, nil
}

type outboundMiddleware struct {
	authProvider clientworker.AuthorizationProvider
}

func (om *outboundMiddleware) Call(ctx context.Context, request *transport.Request, out transport.UnaryOutbound) (*transport.Response, error) {
	if om.authProvider != nil {
		token, err := om.authProvider.GetAuthToken()
		if err != nil {
			return nil, err
		}
		request.Headers = request.Headers.
			With(common.AuthorizationTokenHeaderName, string(token))
	}
	return out.Call(ctx, request)
}
