package membership

import (
	"github.com/hashicorp/serf/serf"
	"github.com/uber-common/bark"
)

type serfMonitor struct {
	logger bark.Logger
	serf   *serf.Serf
}

// NewSerfMonitor returns a new serf based monitor
func NewSerfMonitor(logger bark.Logger) Monitor {
	serf, err := serf.Create(nil)
	if err != nil {
		logger.Fatal("unable to create serf")
	}
	return &serfMonitor{logger: logger, serf: serf}
}

func (s *serfMonitor) Start() error {
	return nil
}

func (s *serfMonitor) Stop() {
}

func (s *serfMonitor) WhoAmI() (*HostInfo, error) {
	return NewHostInfo("", nil), nil
}

func (s *serfMonitor) GetResolver(service string) (ServiceResolver, error) {
	return nil, nil
}

func (s *serfMonitor) Lookup(service string, key string) (*HostInfo, error) {
	return nil, nil
}

func (s *serfMonitor) AddListener(service string, name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *serfMonitor) RemoveListener(service string, name string) error {
	return nil
}
