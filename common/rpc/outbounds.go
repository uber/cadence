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

	"github.com/uber/cadence/common/authorization"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/service"

	"go.uber.org/multierr"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/middleware"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/yarpc/transport/tchannel"
)

const (
	// OutboundPublicClient is the name of configured public client outbound
	OutboundPublicClient = "public-client"
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
				errs = multierr.Append(errs, fmt.Errorf("outbound %s already configured", name))
				break
			}
			outbounds[name] = outbound
		}
	}
	return outbounds, errs
}

type publicClientOutbound struct {
	address        string
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

	return publicClientOutbound{config.PublicClient.HostPort, authMiddleware}, nil
}

func (b publicClientOutbound) Build(_ *grpc.Transport, tchannel *tchannel.Transport) (yarpc.Outbounds, error) {
	return yarpc.Outbounds{
		OutboundPublicClient: {
			ServiceName: service.Frontend,
			Unary:       middleware.ApplyUnaryOutbound(tchannel.NewSingleOutbound(b.address), b.authMiddleware),
		},
	}, nil
}
