package config

import (
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

	// set the hostname as identifier for serf
	config.NodeName = s.config.Name
	config.Tags = map[string]string{membership.RoleKey: s.serviceName}
	return membership.NewSerfMonitor(CadenceServices, s.config.BootstrapHosts, config, s.logger), nil
}
