package config

import (
	"strconv"
	"strings"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/membership"
	"go.uber.org/yarpc"
)

// SerfFactory implements the SerfFactory interface
type SerfFactory struct {
	config      *Serf
	logger      bark.Logger
	serviceName string
}

// NewFactory builds a ringpop factory conforming
// to the underlying configuration
func (serfConfig *Serf) NewFactory(logger bark.Logger, serviceName string) (*SerfFactory, error) {
	return &SerfFactory{serfConfig, logger, serviceName}, nil
}

// Create - creates a membership monitor backed by serf
func (s *SerfFactory) Create(dispatcher *yarpc.Dispatcher) (membership.Monitor, error) {
	config := serf.DefaultConfig()

	port, _ := strconv.ParseInt(strings.Split(s.config.Name, ":")[1], 0, 0)
	port = port + 800

	// set the hostname as identifier for serf
	config.NodeName = s.config.Name
	config.Tags = map[string]string{membership.RoleKey: s.serviceName}
	config.MemberlistConfig = memberlist.DefaultLANConfig()
	config.MemberlistConfig.BindAddr = "127.0.0.1"
	config.MemberlistConfig.BindPort = int(port)
	config.MemberlistConfig.AdvertiseAddr = "127.0.0.1"
	config.MemberlistConfig.AdvertisePort = int(port)
	s.logger.Infof("creating serf monitor on port %v for service %v", port, s.serviceName)
	return membership.NewSerfMonitor(CadenceServices, s.config.BootstrapHosts, config, s.logger), nil
}
