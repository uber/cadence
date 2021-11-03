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
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/yarpc/transport/tchannel"

	"github.com/uber/ringpop-go"
	"github.com/uber/ringpop-go/discovery"
	"github.com/uber/ringpop-go/discovery/jsonfile"
	"github.com/uber/ringpop-go/discovery/statichosts"
	"github.com/uber/ringpop-go/swim"
	tcg "github.com/uber/tchannel-go"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/service"
)

const (
	// BootstrapModeNone represents a bootstrap mode set to nothing or invalid
	BootstrapModeNone BootstrapMode = iota
	// BootstrapModeFile represents a file-based bootstrap mode
	BootstrapModeFile
	// BootstrapModeHosts represents a list of hosts passed in the configuration
	BootstrapModeHosts
	// BootstrapModeCustom represents a custom bootstrap mode
	BootstrapModeCustom
	// BootstrapModeDNS represents a list of hosts passed in the configuration
	// to be resolved, and the resulting addresses are used for bootstrap
	BootstrapModeDNS
	// BootstrapModeDNSSRV represents a list of DNS hosts passed in the configuration to resolve host and ports, resulting in a list of addresses
	BootstrapModeDNSSRV
)

const (
	defaultMaxJoinDuration = 10 * time.Second
)

// RingpopFactory implements the RingpopFactory interface
type RingpopFactory struct {
	config      *Ringpop
	channel     tchannel.Channel
	serviceName string
	logger      log.Logger

	sync.Mutex
	ringPop           *membership.RingPop
	membershipMonitor membership.Monitor
}

// NewFactory builds a ringpop factory conforming
// to the underlying configuration
func (rpConfig *Ringpop) NewFactory(
	channel tchannel.Channel,
	serviceName string,
	logger log.Logger,
) (*RingpopFactory, error) {

	return newRingpopFactory(rpConfig, channel, serviceName, logger)
}

func (rpConfig *Ringpop) validate() error {
	if len(rpConfig.Name) == 0 {
		return fmt.Errorf("ringpop config missing `name` param")
	}
	return validateBootstrapMode(rpConfig)
}

// UnmarshalYAML is called by the yaml package to convert
// the config YAML into a BootstrapMode.
func (m *BootstrapMode) UnmarshalYAML(
	unmarshal func(interface{}) error,
) error {

	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	var err error
	*m, err = parseBootstrapMode(s)
	return err
}

// parseBootstrapMode reads a string value and returns a bootstrap mode.
func parseBootstrapMode(
	mode string,
) (BootstrapMode, error) {

	switch strings.ToLower(mode) {
	case "hosts":
		return BootstrapModeHosts, nil
	case "file":
		return BootstrapModeFile, nil
	case "custom":
		return BootstrapModeCustom, nil
	case "dns":
		return BootstrapModeDNS, nil
	case "dns-srv":
		return BootstrapModeDNSSRV, nil
	}
	return BootstrapModeNone, errors.New("invalid or no ringpop bootstrap mode")
}

func validateBootstrapMode(
	rpConfig *Ringpop,
) error {

	switch rpConfig.BootstrapMode {
	case BootstrapModeFile:
		if len(rpConfig.BootstrapFile) == 0 {
			return fmt.Errorf("ringpop config missing bootstrap file param")
		}
	case BootstrapModeHosts, BootstrapModeDNS, BootstrapModeDNSSRV:
		if len(rpConfig.BootstrapHosts) == 0 {
			return fmt.Errorf("ringpop config missing boostrap hosts param")
		}
	case BootstrapModeCustom:
		if rpConfig.DiscoveryProvider == nil {
			return fmt.Errorf("ringpop bootstrapMode is set to custom but discoveryProvider is nil")
		}
	default:
		return fmt.Errorf("ringpop config with unknown boostrap mode")
	}
	return nil
}

func newRingpopFactory(
	rpConfig *Ringpop,
	channel tchannel.Channel,
	serviceName string,
	logger log.Logger,
) (*RingpopFactory, error) {

	if err := rpConfig.validate(); err != nil {
		return nil, err
	}
	if rpConfig.MaxJoinDuration == 0 {
		rpConfig.MaxJoinDuration = defaultMaxJoinDuration
	}
	return &RingpopFactory{
		config:      rpConfig,
		channel:     channel,
		serviceName: serviceName,
		logger:      logger,
	}, nil
}

// GetMembershipMonitor return a membership monitor
func (factory *RingpopFactory) GetMembershipMonitor() (membership.Monitor, error) {
	factory.Lock()
	defer factory.Unlock()

	return factory.getMembership()
}

func (factory *RingpopFactory) getMembership() (membership.Monitor, error) {
	if factory.membershipMonitor != nil {
		return factory.membershipMonitor, nil
	}

	membershipMonitor, err := factory.createMembership()
	if err != nil {
		return nil, err
	}
	factory.membershipMonitor = membershipMonitor
	return membershipMonitor, nil
}

func (factory *RingpopFactory) createMembership() (membership.Monitor, error) {
	// use actual listen port (in case service is bound to :0 or 0.0.0.0:0)
	rp, err := factory.getRingpop()
	if err != nil {
		return nil, fmt.Errorf("ringpop creation failed: %v", err)
	}

	membershipMonitor := membership.NewRingpopMonitor(factory.serviceName, service.List, rp, factory.logger)
	return membershipMonitor, nil
}

func (factory *RingpopFactory) getRingpop() (*membership.RingPop, error) {
	if factory.ringPop != nil {
		return factory.ringPop, nil
	}

	ringPop, err := factory.createRingpop()
	if err != nil {
		return nil, err
	}
	factory.ringPop = ringPop
	return ringPop, nil
}

func (factory *RingpopFactory) createRingpop() (*membership.RingPop, error) {
	rp, err := ringpop.New(
		factory.config.Name,
		ringpop.Channel(factory.channel.(*tcg.Channel)),
	)
	if err != nil {
		return nil, err
	}

	discoveryProvider, err := newDiscoveryProvider(factory.config, factory.logger)
	if err != nil {
		return nil, err
	}
	bootstrapOpts := &swim.BootstrapOptions{
		MaxJoinDuration:  factory.config.MaxJoinDuration,
		DiscoverProvider: discoveryProvider,
	}
	return membership.NewRingPop(rp, bootstrapOpts, factory.logger), nil
}

type dnsHostResolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
	LookupSRV(ctx context.Context, service string, proto string, name string) (cname string, addrs []*net.SRV, err error)
}

type dnsProvider struct {
	UnresolvedHosts []string
	Resolver        dnsHostResolver
	Logger          log.Logger
}

func newDNSProvider(
	hosts []string,
	resolver dnsHostResolver,
	logger log.Logger,
) *dnsProvider {

	set := map[string]struct{}{}
	for _, hostport := range hosts {
		set[hostport] = struct{}{}
	}

	var keys []string
	for key := range set {
		keys = append(keys, key)
	}
	return &dnsProvider{
		UnresolvedHosts: keys,
		Resolver:        resolver,
		Logger:          logger,
	}
}

func (provider *dnsProvider) Hosts() ([]string, error) {
	var results []string
	resolvedHosts := map[string][]string{}
	for _, hostPort := range provider.UnresolvedHosts {
		host, port, err := net.SplitHostPort(hostPort)
		if err != nil {
			provider.Logger.Warn("could not split host and port", tag.Address(hostPort), tag.Error(err))
			continue
		}

		resolved, exists := resolvedHosts[host]
		if !exists {
			resolved, err = provider.Resolver.LookupHost(context.Background(), host)
			if err != nil {
				provider.Logger.Warn("could not resolve host", tag.Address(host), tag.Error(err))
				continue
			}
			resolvedHosts[host] = resolved
		}
		for _, r := range resolved {
			results = append(results, net.JoinHostPort(r, port))
		}
	}
	if len(results) == 0 {
		return nil, errors.New("no hosts found, and bootstrap requires at least one")
	}
	return results, nil
}

type dnsSRVProvider struct {
	UnresolvedHosts []string
	Resolver        dnsHostResolver
	Logger          log.Logger
}

func newDNSSRVProvider(
	hosts []string,
	resolver dnsHostResolver,
	logger log.Logger,
) *dnsSRVProvider {

	set := map[string]struct{}{}
	for _, hostport := range hosts {
		set[hostport] = struct{}{}
	}

	var keys []string
	for key := range set {
		keys = append(keys, key)
	}

	return &dnsSRVProvider{
		UnresolvedHosts: keys,
		Resolver:        resolver,
		Logger:          logger,
	}
}

func (provider *dnsSRVProvider) Hosts() ([]string, error) {
	var results []string
	resolvedHosts := map[string][]string{}

	for _, service := range provider.UnresolvedHosts {
		serviceParts := strings.Split(service, ".")
		if len(serviceParts) <= 2 {
			provider.Logger.Warn("could seperate service name from domain", tag.Address(service))
			continue
		}
		serviceName := serviceParts[0]
		domain := strings.Join(serviceParts[1:], ".")
		resolved, exists := resolvedHosts[serviceName]
		if !exists {
			_, addrs, err := provider.Resolver.LookupSRV(context.Background(), serviceName, "tcp", domain)

			if err != nil {
				provider.Logger.Warn("could not resolve host", tag.Address(serviceName), tag.Error(err))
				continue
			}

			var targets []string
			for _, record := range addrs {
			   target, err := provider.Resolver.LookupHost(context.Background(), record.Target)

			   if err != nil {
					provider.Logger.Warn("could not resolve srv dns host", tag.Address(record.Target), tag.Error(err))
					continue
			   }
			   for _, host := range target {
			   		targets = append(targets, net.JoinHostPort(host, fmt.Sprintf("%d", record.Port)))
			   }
			}
			resolvedHosts[serviceName] = targets
			resolved = targets
		}

		for _, r := range resolved {
			results = append(results, r)
		}

	}

	if len(results) == 0 {
		return nil, errors.New("no hosts found, and bootstrap requires at least one")
	}
	return results, nil
}

func newDiscoveryProvider(
	cfg *Ringpop,
	logger log.Logger,
) (discovery.DiscoverProvider, error) {

	if cfg.DiscoveryProvider != nil {
		// custom discovery provider takes first precedence
		return cfg.DiscoveryProvider, nil
	}

	switch cfg.BootstrapMode {
	case BootstrapModeHosts:
		return statichosts.New(cfg.BootstrapHosts...), nil
	case BootstrapModeFile:
		return jsonfile.New(cfg.BootstrapFile), nil
	case BootstrapModeDNS:
		return newDNSProvider(cfg.BootstrapHosts, net.DefaultResolver, logger), nil
	case BootstrapModeDNSSRV:
		return newDNSSRVProvider(cfg.BootstrapHosts, net.DefaultResolver, logger), nil
	}
	return nil, fmt.Errorf("unknown bootstrap mode")
}
