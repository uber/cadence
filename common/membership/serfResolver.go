package membership

import (
	"fmt"

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
	//	fmt.Printf("looking up members for service %v\n", s.service)
	members := make([]string, 0, 10)
	for _, member := range s.serf.Members() {
		if val, ok := member.Tags[RoleKey]; ok && val == s.service {
			members = append(members, member.Name)
		}
	}
	//	fmt.Printf("list of members are %v\n", members)
	hash := int(farm.Fingerprint32([]byte(key)))
	idx := hash % len(members)
	fmt.Printf("returning host %v\n", members[idx])
	return NewHostInfo(members[idx], map[string]string{RoleKey: s.service}), nil
}

func (s *serfResolver) AddListener(name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *serfResolver) RemoveListener(name string) error {
	return nil
}
