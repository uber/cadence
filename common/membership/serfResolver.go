// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
