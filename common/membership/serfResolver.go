package membership

import "github.com/hashicorp/serf/serf"

type serfResolver struct {
	service string
	serf    *serf.Serf
}

func newSerfResolver(service string, serf *serf.Serf) ServiceResolver {
	return &serfResolver{service, serf}
}

func (s *serfResolver) Lookup(key string) (*HostInfo, error) {
	return nil, nil
}

func (s *serfResolver) AddListener(name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *serfResolver) RemoveListener(name string) error {
	return nil
}
