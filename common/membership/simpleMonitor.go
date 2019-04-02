package membership

import "fmt"

type simpleMonitor struct {
	hostInfo  *HostInfo
	resolvers map[string]ServiceResolver
}

// NewSimpleMonitor returns a simple monitor interface
func NewSimpleMonitor(serviceName string, hosts map[string][]string) Monitor {
	resolvers := make(map[string]ServiceResolver, len(hosts))
	for service, hostList := range hosts {
		resolvers[service] = newSimpleResolver(service, hostList)
	}
	hostInfo := NewHostInfo(hosts[serviceName][0], map[string]string{RoleKey: serviceName})
	return &simpleMonitor{hostInfo, resolvers}
}

func (s *simpleMonitor) Start() error {
	return nil
}

func (s *simpleMonitor) Stop() {
}

func (s *simpleMonitor) WhoAmI() (*HostInfo, error) {
	return s.hostInfo, nil
}

func (s *simpleMonitor) GetResolver(service string) (ServiceResolver, error) {
	return s.resolvers[service], nil
}

func (s *simpleMonitor) Lookup(service string, key string) (*HostInfo, error) {
	resolver, ok := s.resolvers[service]
	if !ok {
		return nil, fmt.Errorf("cannot lookup host for service %v", service)
	}
	return resolver.Lookup(key)
}

func (s *simpleMonitor) AddListener(service string, name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *simpleMonitor) RemoveListener(service string, name string) error {
	return nil
}
