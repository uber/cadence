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
	"fmt"
	"strings"

	"github.com/hashicorp/serf/serf"
	"github.com/uber/cadence/common/log"
)

type serfMonitor struct {
	service   string
	logger    log.Logger
	serf      *serf.Serf
	hosts     []string
	resolvers map[string]ServiceResolver
}

// NewSerfMonitor returns a new serf based monitor
func NewSerfMonitor(services []string, hosts []string, config *serf.Config, logger log.Logger) Monitor {
	serf, err := serf.Create(config)
	if err != nil {
		logger.Fatal(fmt.Sprintf("unable to create serf %v", config.Tags[RoleKey]))
	}
	resolvers := make(map[string]ServiceResolver, len(services))
	for _, service := range services {
		resolvers[service] = newSerfResolver(service, serf)
	}
	return &serfMonitor{service: config.Tags[RoleKey], logger: logger, hosts: hosts, serf: serf, resolvers: resolvers}
}

func (s *serfMonitor) Start() error {
	if strings.Contains(s.service, "history") {
		return nil
	}
	if _, err := s.serf.Join(s.hosts, false); err != nil {
		return err
	}
	return nil
}

func (s *serfMonitor) Stop() {
	s.serf.Leave()
	s.serf.Shutdown()
}

func (s *serfMonitor) WhoAmI() (*HostInfo, error) {
	member := s.serf.LocalMember()
	return NewHostInfo(member.Name, member.Tags), nil
}

func (s *serfMonitor) GetResolver(service string) (ServiceResolver, error) {
	return s.resolvers[service], nil
}

func (s *serfMonitor) Lookup(service string, key string) (*HostInfo, error) {
	return s.resolvers[service].Lookup(key)
}

func (s *serfMonitor) AddListener(service string, name string, notifyChannel chan<- *ChangedEvent) error {
	return nil
}

func (s *serfMonitor) RemoveListener(service string, name string) error {
	return nil
}
