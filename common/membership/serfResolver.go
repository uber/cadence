package membership

import (
	"github.com/dgryski/go-farm"
	"github.com/hashicorp/serf/serf"
)

type serfResolver struct {
	service string
	serf    *serf.Serf
}

func newSerfResolver(service string, serf *serf.Serf) ServiceResolver {
	return &serfResolver{service, serf}
}

func (s *serfResolver) Lookup(key string) (*HostInfo, error) {
	members := make([]string, 0, 10)
	for _, member := range s.serf.Members() {
		if val, ok := member.Tags[RoleKey]; ok && val == s.service {
			members = append(members, member.Name)
		}
	}
	hash := int(farm.Fingerprint32([]byte(key)))
	idx := hash % len(members)
	hostInfo := NewHostInfo(members[idx], map[string]string{RoleKey: s.service})
	return hostInfo, nil
}

func (s *serfResolver) AddListener(name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *serfResolver) RemoveListener(name string) error {
	return nil
}
