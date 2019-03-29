package membership

type simpleMonitor struct {
}

// NewSimpleMonitor returns a simple monitor interface
func NewSimpleMonitor() Monitor {
	return &simpleMonitor{}
}

func (s *simpleMonitor) Start() error {
	return nil
}

func (s *simpleMonitor) Stop() {
}

func (s *simpleMonitor) WhoAmI() (*HostInfo, error) {
	return NewHostInfo("address", map[string]string{}), nil
}

func (s *simpleMonitor) GetResolver(service string) (ServiceResolver, error) {
	return nil, nil
}

func (s *simpleMonitor) Lookup(service string, key string) (*HostInfo, error) {
	return nil, nil
}

func (s *simpleMonitor) AddListener(service string, name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *simpleMonitor) RemoveListener(service string, name string) error {
	return nil
}
