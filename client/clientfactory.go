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

package client

import (
	"fmt"
	"time"

	"go.uber.org/yarpc/api/transport"

	"github.com/uber/cadence/.gen/go/admin/adminserviceclient"
	"github.com/uber/cadence/.gen/go/cadence/workflowserviceclient"
	"github.com/uber/cadence/.gen/go/history/historyserviceclient"
	"github.com/uber/cadence/.gen/go/matching/matchingserviceclient"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	adminv1 "github.com/uber/cadence/.gen/proto/admin/v1"
	historyv1 "github.com/uber/cadence/.gen/proto/history/v1"
	matchingv1 "github.com/uber/cadence/.gen/proto/matching/v1"

	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/rpc"
	"github.com/uber/cadence/common/service"
)

type (
	// Factory can be used to create RPC clients for cadence services
	Factory interface {
		NewHistoryClient() (history.Client, error)
		NewMatchingClient(domainIDToName DomainIDToNameFunc) (matching.Client, error)

		NewHistoryClientWithTimeout(timeout time.Duration) (history.Client, error)
		NewMatchingClientWithTimeout(domainIDToName DomainIDToNameFunc, timeout time.Duration, longPollTimeout time.Duration) (matching.Client, error)

		NewAdminClientWithTimeoutAndConfig(config transport.ClientConfig, timeout time.Duration, largeTimeout time.Duration) (admin.Client, error)
		NewFrontendClientWithTimeoutAndConfig(config transport.ClientConfig, timeout time.Duration, longPollTimeout time.Duration) (frontend.Client, error)
	}

	// DomainIDToNameFunc maps a domainID to domain name. Returns error when mapping is not possible.
	DomainIDToNameFunc func(string) (string, error)

	rpcClientFactory struct {
		rpcFactory            common.RPCFactory
		resolver              membership.Resolver
		metricsClient         metrics.Client
		dynConfig             *dynamicconfig.Collection
		numberOfHistoryShards int
		logger                log.Logger
	}
)

// NewRPCClientFactory creates an instance of client factory that knows how to dispatch RPC calls.
func NewRPCClientFactory(
	rpcFactory common.RPCFactory,
	resolver membership.Resolver,
	metricsClient metrics.Client,
	dc *dynamicconfig.Collection,
	numberOfHistoryShards int,
	logger log.Logger,
) Factory {
	return &rpcClientFactory{
		rpcFactory:            rpcFactory,
		resolver:              resolver,
		metricsClient:         metricsClient,
		dynConfig:             dc,
		numberOfHistoryShards: numberOfHistoryShards,
		logger:                logger,
	}
}

func (cf *rpcClientFactory) NewHistoryClient() (history.Client, error) {
	return cf.NewHistoryClientWithTimeout(history.DefaultTimeout)
}

func (cf *rpcClientFactory) NewMatchingClient(domainIDToName DomainIDToNameFunc) (matching.Client, error) {
	return cf.NewMatchingClientWithTimeout(domainIDToName, matching.DefaultTimeout, matching.DefaultLongPollTimeout)
}

func (cf *rpcClientFactory) NewHistoryClientWithTimeout(timeout time.Duration) (history.Client, error) {
	var rawClient history.Client
	var addressMapper history.AddressMapperFn
	outboundConfig := cf.rpcFactory.GetDispatcher().ClientConfig(service.History)
	if rpc.IsGRPCOutbound(outboundConfig) {
		rawClient = history.NewGRPCClient(historyv1.NewHistoryAPIYARPCClient(outboundConfig))
		addressMapper = func(address string) (string, error) {
			return cf.rpcFactory.ReplaceGRPCPort(service.History, address)
		}
	} else {
		rawClient = history.NewThriftClient(historyserviceclient.New(outboundConfig))
	}

	peerResolver := history.NewPeerResolver(cf.numberOfHistoryShards, cf.resolver, addressMapper)

	supportedMessageSize := cf.rpcFactory.GetMaxMessageSize()
	maxSizeConfig := cf.dynConfig.GetIntProperty(dynamicconfig.GRPCMaxSizeInByte, supportedMessageSize)
	if maxSizeConfig() > supportedMessageSize {
		return nil, fmt.Errorf(
			"GRPCMaxSizeInByte dynamic config value %v is larger than supported value %v",
			maxSizeConfig(),
			supportedMessageSize,
		)
	}
	client := history.NewClient(
		cf.numberOfHistoryShards,
		maxSizeConfig,
		timeout,
		rawClient,
		peerResolver,
		cf.logger,
	)
	if errorRate := cf.dynConfig.GetFloat64Property(dynamicconfig.HistoryErrorInjectionRate, 0)(); errorRate != 0 {
		client = history.NewErrorInjectionClient(client, errorRate, cf.logger)
	}
	if cf.metricsClient != nil {
		client = history.NewMetricClient(client, cf.metricsClient)
	}
	return client, nil
}

func (cf *rpcClientFactory) NewMatchingClientWithTimeout(
	domainIDToName DomainIDToNameFunc,
	timeout time.Duration,
	longPollTimeout time.Duration,
) (matching.Client, error) {
	var rawClient matching.Client
	var addressMapper matching.AddressMapperFn
	outboundConfig := cf.rpcFactory.GetDispatcher().ClientConfig(service.Matching)
	if rpc.IsGRPCOutbound(outboundConfig) {
		rawClient = matching.NewGRPCClient(matchingv1.NewMatchingAPIYARPCClient(outboundConfig))
		addressMapper = func(address string) (string, error) {
			return cf.rpcFactory.ReplaceGRPCPort(service.Matching, address)
		}
	} else {
		rawClient = matching.NewThriftClient(matchingserviceclient.New(outboundConfig))
	}

	peerResolver := matching.NewPeerResolver(cf.resolver, addressMapper)

	client := matching.NewClient(
		timeout,
		longPollTimeout,
		rawClient,
		peerResolver,
		matching.NewLoadBalancer(domainIDToName, cf.dynConfig),
	)
	if errorRate := cf.dynConfig.GetFloat64Property(dynamicconfig.MatchingErrorInjectionRate, 0)(); errorRate != 0 {
		client = matching.NewErrorInjectionClient(client, errorRate, cf.logger)
	}
	if cf.metricsClient != nil {
		client = matching.NewMetricClient(client, cf.metricsClient)
	}
	return client, nil

}

func (cf *rpcClientFactory) NewAdminClientWithTimeoutAndConfig(
	config transport.ClientConfig,
	timeout time.Duration,
	largeTimeout time.Duration,
) (admin.Client, error) {
	var client admin.Client
	if rpc.IsGRPCOutbound(config) {
		client = admin.NewGRPCClient(adminv1.NewAdminAPIYARPCClient(config))
	} else {
		client = admin.NewThriftClient(adminserviceclient.New(config))
	}

	client = admin.NewClient(timeout, largeTimeout, client)
	if errorRate := cf.dynConfig.GetFloat64Property(dynamicconfig.AdminErrorInjectionRate, 0)(); errorRate != 0 {
		client = admin.NewErrorInjectionClient(client, errorRate, cf.logger)
	}
	if cf.metricsClient != nil {
		client = admin.NewMetricClient(client, cf.metricsClient)
	}
	return client, nil
}

func (cf *rpcClientFactory) NewFrontendClientWithTimeoutAndConfig(
	config transport.ClientConfig,
	timeout time.Duration,
	longPollTimeout time.Duration,
) (frontend.Client, error) {
	var client frontend.Client
	if rpc.IsGRPCOutbound(config) {
		client = frontend.NewGRPCClient(
			apiv1.NewDomainAPIYARPCClient(config),
			apiv1.NewWorkflowAPIYARPCClient(config),
			apiv1.NewWorkerAPIYARPCClient(config),
			apiv1.NewVisibilityAPIYARPCClient(config),
		)
	} else {
		client = frontend.NewThriftClient(workflowserviceclient.New(config))
	}

	client = frontend.NewClient(timeout, longPollTimeout, client)
	if errorRate := cf.dynConfig.GetFloat64Property(dynamicconfig.FrontendErrorInjectionRate, 0)(); errorRate != 0 {
		client = frontend.NewErrorInjectionClient(client, errorRate, cf.logger)
	}
	if cf.metricsClient != nil {
		client = frontend.NewMetricClient(client, cf.metricsClient)
	}
	return client, nil
}
