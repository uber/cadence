package membership

import (
	"strings"

	"github.com/hashicorp/serf/serf"
	"github.com/uber-common/bark"
)

type serfMonitor struct {
	service   string
	logger    bark.Logger
	serf      *serf.Serf
	hosts     []string
	resolvers map[string]ServiceResolver
}

// NewSerfMonitor returns a new serf based monitor
func NewSerfMonitor(services []string, hosts []string, config *serf.Config, logger bark.Logger) Monitor {
	serf, err := serf.Create(config)
	if err != nil {
		logger.Fatalf("unable to create serf %v", config.Tags[RoleKey])
	}
	resolvers := make(map[string]ServiceResolver, len(services))
	for _, service := range services {
		resolvers[service] = newSerfResolver(service, serf)
	}
	return &serfMonitor{service: config.Tags[RoleKey], logger: logger, hosts: hosts, serf: serf, resolvers: resolvers}
}

func (s *serfMonitor) Start() error {
	if strings.Contains(s.service, "history") {
		return nil
	}
	if _, err := s.serf.Join(s.hosts, false); err != nil {
		return err
	}
	return nil
}

func (s *serfMonitor) Stop() {
	s.serf.Leave()
}

func (s *serfMonitor) WhoAmI() (*HostInfo, error) {
	member := s.serf.LocalMember()
	return NewHostInfo(member.Name, member.Tags), nil
}

func (s *serfMonitor) GetResolver(service string) (ServiceResolver, error) {
	return s.resolvers[service], nil
}

func (s *serfMonitor) Lookup(service string, key string) (*HostInfo, error) {
	return s.resolvers[service].Lookup(key)
}

func (s *serfMonitor) AddListener(service string, name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *serfMonitor) RemoveListener(service string, name string) error {
	return nil
}
