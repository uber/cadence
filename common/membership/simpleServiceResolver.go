package membership

import (
	"github.com/dgryski/go-farm"
)

type simpleResolver struct {
	hosts    []*HostInfo
	hashfunc func([]byte) uint32
}

// newSimpleResolver returns a service resolver that maintains static mapping
// between services and host info
func newSimpleResolver(hosts []string) ServiceResolver {
	hostInfos := make([]*HostInfo, 0, len(hosts))
	for _, host := range hosts {
		hostInfos = append(hostInfos, NewHostInfo(host, nil))
	}
	return &simpleResolver{hostInfos, farm.Fingerprint32}
}

func (s *simpleResolver) Lookup(key string) (*HostInfo, error) {
	hash := int(s.hashfunc([]byte(key)))
	return s.hosts[hash%len(s.hosts)], nil
}

func (s *simpleResolver) AddListener(name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *simpleResolver) RemoveListener(name string) error {
	return nil
}
