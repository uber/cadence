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

package ringpop

import (
	"fmt"
	"sync/atomic"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
)

type monitor struct {
	status int32

	serviceName string
	services    []string
	rp          *RingPop
	rings       map[string]*ringpopServiceResolver
	logger      log.Logger
}

var _ membership.Monitor = (*monitor)(nil)

// NewMonitor returns a ringpop-based membership monitor
func NewMonitor(
	serviceName string,
	services []string,
	rp *RingPop,
	logger log.Logger,
) membership.Monitor {

	rpo := &monitor{
		status:      common.DaemonStatusInitialized,
		serviceName: serviceName,
		services:    services,
		rp:          rp,
		logger:      logger,
		rings:       make(map[string]*ringpopServiceResolver),
	}
	for _, service := range services {
		rpo.rings[service] = newRingpopServiceResolver(service, rp, logger)
	}
	return rpo
}

func (rpo *monitor) Start() {
	if !atomic.CompareAndSwapInt32(
		&rpo.status,
		common.DaemonStatusInitialized,
		common.DaemonStatusStarted,
	) {
		return
	}

	rpo.rp.Start()

	labels, err := rpo.rp.ringpop.Labels()
	if err != nil {
		rpo.logger.Fatal("unable to get ringpop labels", tag.Error(err))
	}

	if err := labels.Set(membership.RoleKey, rpo.serviceName); err != nil {
		rpo.logger.Fatal("unable to set ringpop labels", tag.Error(err))
	}

	for _, ring := range rpo.rings {
		ring.Start()
	}
}

func (rpo *monitor) Stop() {
	if !atomic.CompareAndSwapInt32(
		&rpo.status,
		common.DaemonStatusStarted,
		common.DaemonStatusStopped,
	) {
		return
	}

	for _, ring := range rpo.rings {
		ring.Stop()
	}

	rpo.rp.Stop()
}

func (rpo *monitor) WhoAmI() (*membership.HostInfo, error) {
	address, err := rpo.rp.ringpop.WhoAmI()
	if err != nil {
		return nil, err
	}
	labels, err := rpo.rp.ringpop.Labels()
	if err != nil {
		return nil, err
	}
	return membership.NewHostInfo(address, labels.AsMap()), nil
}

func (rpo *monitor) EvictSelf() error {
	return rpo.rp.ringpop.SelfEvict()
}

func (rpo *monitor) GetResolver(service string) (membership.ServiceResolver, error) {
	ring, found := rpo.rings[service]
	if !found {
		return nil, fmt.Errorf("service %q is not tracked by monitor", service)
	}
	return ring, nil
}

func (rpo *monitor) Lookup(service string, key string) (*membership.HostInfo, error) {
	ring, err := rpo.GetResolver(service)
	if err != nil {
		return nil, err
	}
	return ring.Lookup(key)
}

func (rpo *monitor) AddListener(service string, name string, notifyChannel chan<- *membership.ChangedEvent) error {
	ring, err := rpo.GetResolver(service)
	if err != nil {
		return err
	}
	return ring.AddListener(name, notifyChannel)
}

func (rpo *monitor) RemoveListener(service string, name string) error {
	ring, err := rpo.GetResolver(service)
	if err != nil {
		return err
	}
	return ring.RemoveListener(name)
}

func (rpo *monitor) GetReachableMembers() ([]string, error) {
	return rpo.rp.ringpop.GetReachableMembers()
}

func (rpo *monitor) GetMemberCount(service string) (int, error) {
	ring, err := rpo.GetResolver(service)
	if err != nil {
		return 0, err
	}
	return ring.MemberCount(), nil
}
