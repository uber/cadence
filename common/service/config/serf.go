package config

import (
	"fmt"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/membership"
	"go.uber.org/yarpc"
)

// SerfFactory implements the SerfFactory interface
type SerfFactory struct {
	config      *Serf
	logger      log.Logger
	serviceName string
}

// NewFactory builds a ringpop factory conforming
// to the underlying configuration
func (serfConfig *Serf) NewFactory(logger log.Logger, serviceName string) (*SerfFactory, error) {
	return &SerfFactory{serfConfig, logger, serviceName}, nil
}

// Create - creates a membership monitor backed by serf
func (s *SerfFactory) Create(dispatcher *yarpc.Dispatcher) (membership.Monitor, error) {
	config := serf.DefaultConfig()

	// set the hostname as identifier for serf
	config.NodeName = s.config.Name
	config.Tags = map[string]string{membership.RoleKey: s.serviceName}
	config.MemberlistConfig = memberlist.DefaultLANConfig()
	config.MemberlistConfig.BindAddr = "127.0.0.1"
	config.MemberlistConfig.BindPort = int(s.config.Port)
	config.MemberlistConfig.AdvertiseAddr = "127.0.0.1"
	config.MemberlistConfig.AdvertisePort = int(s.config.Port)
	s.logger.Info(fmt.Sprintf("creating serf monitor on port %v for service %v", s.config.Port, s.serviceName))
	return membership.NewSerfMonitor(CadenceServices, s.config.BootstrapHosts, config, s.logger), nil
}
