package membership

type simpleResolver struct {
	hosts map[string][]string
}

// NewSimpleResolver returns a service resolver that maintains static mapping
// between services and host info
func NewSimpleResolver() ServiceResolver {
	return &simpleResolver{}
}

func (s *simpleResolver) Lookup(key string) (*HostInfo, error) {
	return nil, nil
}

func (s *simpleResolver) AddListener(name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *simpleResolver) RemoveListener(name string) error {
	return nil
}
