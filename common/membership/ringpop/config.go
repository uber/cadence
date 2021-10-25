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

package ringpop

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/uber/ringpop-go/discovery"
	"github.com/uber/ringpop-go/discovery/jsonfile"
	"github.com/uber/ringpop-go/discovery/statichosts"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

// Config contains the ringpop config items
type Config struct {
	// Name to be used in ringpop advertisement
	Name string `yaml:"name" validate:"nonzero"`
	// BootstrapMode is a enum that defines the ringpop bootstrap method
	BootstrapMode BootstrapMode `yaml:"bootstrapMode"`
	// BootstrapHosts is a list of seed hosts to be used for ringpop bootstrap
	BootstrapHosts []string `yaml:"bootstrapHosts"`
	// BootstrapFile is the file path to be used for ringpop bootstrap
	BootstrapFile string `yaml:"bootstrapFile"`
	// MaxJoinDuration is the max wait time to join the ring
	MaxJoinDuration time.Duration `yaml:"maxJoinDuration"`
	// Custom discovery provider, cannot be specified through yaml
	DiscoveryProvider discovery.DiscoverProvider `yaml:"-"`
}

// BootstrapMode is an enum type for ringpop bootstrap mode
type BootstrapMode int

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
)

const (
	defaultMaxJoinDuration = 10 * time.Second
)

func (rpConfig *Config) validate() error {
	if len(rpConfig.Name) == 0 {
		return fmt.Errorf("ringpop config missing `name` param")
	}
	return validateBootstrapMode(rpConfig)
}

func validateBootstrapMode(
	rpConfig *Config,
) error {
	switch rpConfig.BootstrapMode {
	case BootstrapModeFile:
		if len(rpConfig.BootstrapFile) == 0 {
			return fmt.Errorf("ringpop config missing bootstrap file param")
		}
	case BootstrapModeHosts, BootstrapModeDNS:
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
	}
	return BootstrapModeNone, errors.New("invalid or no ringpop bootstrap mode")
}

type dnsHostResolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
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

func newDiscoveryProvider(
	cfg Config,
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
	}
	return nil, fmt.Errorf("unknown bootstrap mode")
}
