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

	"github.com/uber/cadence/common/authorization"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/service"

	"go.uber.org/multierr"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/middleware"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/peer/direct"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/yarpc/transport/tchannel"
)

const (
	// OutboundPublicClient is the name of configured public client outbound
	OutboundPublicClient = "public-client"

	crossDCCaller = "cadence-xdc-client"
)

// OutboundsBuilder allows defining outbounds for the dispatcher
type OutboundsBuilder interface {
	Build(*grpc.Transport, *tchannel.Transport) (yarpc.Outbounds, error)
}

type multiOutbounds struct {
	builders []OutboundsBuilder
}

// CombineOutbounds takes multiple outbound builders and combines them
func CombineOutbounds(builders ...OutboundsBuilder) OutboundsBuilder {
	return multiOutbounds{builders}
}

func (b multiOutbounds) Build(grpc *grpc.Transport, tchannel *tchannel.Transport) (yarpc.Outbounds, error) {
	outbounds := yarpc.Outbounds{}
	var errs error
	for _, builder := range b.builders {
		builderOutbounds, err := builder.Build(grpc, tchannel)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		for name, outbound := range builderOutbounds {
			if _, exists := outbounds[name]; exists {
				errs = multierr.Append(errs, fmt.Errorf("outbound %q already configured", name))
				break
			}
			outbounds[name] = outbound
		}
	}
	return outbounds, errs
}

type publicClientOutbound struct {
	address        string
	isGRPC         bool
	authMiddleware middleware.UnaryOutbound
}

func newPublicClientOutbound(config *config.Config) (publicClientOutbound, error) {
	if len(config.PublicClient.HostPort) == 0 {
		return publicClientOutbound{}, fmt.Errorf("need to provide an endpoint config for PublicClient")
	}

	var authMiddleware middleware.UnaryOutbound
	if config.Authorization.OAuthAuthorizer.Enable {
		clusterName := config.ClusterGroupMetadata.CurrentClusterName
		clusterInfo := config.ClusterGroupMetadata.ClusterGroup[clusterName]
		authProvider, err := authorization.GetAuthProviderClient(clusterInfo.AuthorizationProvider.PrivateKey)
		if err != nil {
			return publicClientOutbound{}, fmt.Errorf("create AuthProvider: %v", err)
		}
		authMiddleware = &authOutboundMiddleware{authProvider}
	}

	isGrpc := config.PublicClient.Transport == grpc.TransportName

	return publicClientOutbound{config.PublicClient.HostPort, isGrpc, authMiddleware}, nil
}

func (b publicClientOutbound) Build(grpc *grpc.Transport, tchannel *tchannel.Transport) (yarpc.Outbounds, error) {
	var outbound transport.UnaryOutbound
	if b.isGRPC {
		outbound = grpc.NewSingleOutbound(b.address)
	} else {
		outbound = tchannel.NewSingleOutbound(b.address)
	}
	return yarpc.Outbounds{
		OutboundPublicClient: {
			ServiceName: service.Frontend,
			Unary:       middleware.ApplyUnaryOutbound(outbound, b.authMiddleware),
		},
	}, nil
}

type crossDCOutbounds struct {
	clusterGroup map[string]config.ClusterInformation
	pcf          PeerChooserFactory
}

func NewCrossDCOutbounds(clusterGroup map[string]config.ClusterInformation, pcf PeerChooserFactory) OutboundsBuilder {
	return crossDCOutbounds{clusterGroup, pcf}
}

func (b crossDCOutbounds) Build(grpcTransport *grpc.Transport, tchannelTransport *tchannel.Transport) (yarpc.Outbounds, error) {
	outbounds := yarpc.Outbounds{}
	for clusterName, clusterInfo := range b.clusterGroup {
		if !clusterInfo.Enabled {
			continue
		}

		var outbound transport.UnaryOutbound
		switch clusterInfo.RPCTransport {
		case tchannel.TransportName:
			peerChooser, err := b.pcf.CreatePeerChooser(tchannelTransport, clusterInfo.RPCAddress)
			if err != nil {
				return nil, err
			}
			outbound = tchannelTransport.NewOutbound(peerChooser)
		case grpc.TransportName:
			tlsConfig, err := clusterInfo.TLS.ToTLSConfig()
			if err != nil {
				return nil, err
			}
			peerChooser, err := b.pcf.CreatePeerChooser(createDialer(grpcTransport, tlsConfig), clusterInfo.RPCAddress)
			if err != nil {
				return nil, err
			}
			outbound = grpcTransport.NewOutbound(peerChooser)
		default:
			return nil, fmt.Errorf("unknown cross DC transport type: %s", clusterInfo.RPCTransport)
		}

		var authMiddleware middleware.UnaryOutbound
		if clusterInfo.AuthorizationProvider.Enable {
			authProvider, err := authorization.GetAuthProviderClient(clusterInfo.AuthorizationProvider.PrivateKey)
			if err != nil {
				return nil, fmt.Errorf("create AuthProvider: %v", err)
			}
			authMiddleware = &authOutboundMiddleware{authProvider}
		}

		outbounds[clusterName] = transport.Outbounds{
			ServiceName: clusterInfo.RPCName,
			Unary: middleware.ApplyUnaryOutbound(outbound, yarpc.UnaryOutboundMiddleware(
				authMiddleware,
				&overrideCallerMiddleware{crossDCCaller},
			)),
		}
	}
	return outbounds, nil
}

type directOutbound struct {
	serviceName string
	grpcEnabled bool
	tlsConfig   *tls.Config
}

func NewDirectOutbound(serviceName string, grpcEnabled bool, tlsConfig *tls.Config) OutboundsBuilder {
	return directOutbound{serviceName, grpcEnabled, tlsConfig}
}

func (o directOutbound) Build(grpc *grpc.Transport, tchannel *tchannel.Transport) (yarpc.Outbounds, error) {
	var outbound transport.UnaryOutbound
	if o.grpcEnabled {
		directChooser, err := direct.New(direct.Configuration{}, createDialer(grpc, o.tlsConfig))
		if err != nil {
			return nil, err
		}
		outbound = grpc.NewOutbound(directChooser)
	} else {
		directChooser, err := direct.New(direct.Configuration{}, tchannel)
		if err != nil {
			return nil, err
		}
		outbound = tchannel.NewOutbound(directChooser)
	}

	return yarpc.Outbounds{
		o.serviceName: {
			ServiceName: o.serviceName,
			Unary:       middleware.ApplyUnaryOutbound(outbound, &ResponseInfoMiddleware{}),
		},
	}, nil
}

func IsGRPCOutbound(config transport.ClientConfig) bool {
	namer, ok := config.GetUnaryOutbound().(transport.Namer)
	if !ok {
		// This should not happen, unless yarpc older than v1.43.0 is used
		panic("Outbound does not implement transport.Namer")
	}
	return namer.TransportName() == grpc.TransportName
}
