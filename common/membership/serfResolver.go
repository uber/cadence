package membership

type serfResolver struct {
}

func newSerfResolver() ServiceResolver {
	return &serfResolver{}
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
