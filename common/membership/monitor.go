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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination monitor_mock.go -self_package github.com/uber/cadence/common/membership

package membership

import (
	"fmt"
	"sync/atomic"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

type (

	// ChangedEvent describes a change in membership
	ChangedEvent struct {
		HostsAdded   []*HostInfo
		HostsUpdated []*HostInfo
		HostsRemoved []*HostInfo
	}

	// Monitor provides membership information for all cadence services.
	Monitor interface {
		common.Daemon
		// WhoAmI returns self address
		WhoAmI() (*HostInfo, error)
		// EvictSelf evicts this member from the membership ring. After this method is
		// called, other members will discover that this node is no longer part of the
		// ring. This primitive is useful to carry out graceful host shutdown during deployments.
		EvictSelf() error

		Lookup(service, key string) (*HostInfo, error)
		// AddListener adds a listener which will get notified on the given
		// channel, whenever membership changes.
		AddListener(service, name string, notifyChannel chan<- *ChangedEvent) error
		// RemoveListener removes a listener for this service.
		RemoveListener(service, name string) error
		// MemberCount returns host count in a hashring
		MemberCount(service string) (int, error)
		// Members returns all host addresses in a hashring
		Members(service string) ([]*HostInfo, error)
	}
)

// NewRingpopMonitor returns a ringpop-based membership monitor
func NewRingpopMonitor(
	serviceName string,
	services []string,
	rp *RingpopWrapper,
	logger log.Logger,
) *RingpopMonitor {

	rpo := &RingpopMonitor{
		status:         common.DaemonStatusInitialized,
		serviceName:    serviceName,
		ringpopWrapper: rp,
		logger:         logger,
		rings:          make(map[string]*ringpopServiceResolver),
	}
	for _, s := range services {
		rpo.rings[s] = newRingpopServiceResolver(s, rp, logger)
	}
	return rpo
}

func (rpo *RingpopMonitor) Start() {
	if !atomic.CompareAndSwapInt32(
		&rpo.status,
		common.DaemonStatusInitialized,
		common.DaemonStatusStarted,
	) {
		return
	}

	rpo.ringpopWrapper.Start()

	labels, err := rpo.ringpopWrapper.Labels()
	if err != nil {
		rpo.logger.Fatal("unable to get ring pop labels", tag.Error(err))
	}

	if err = labels.Set(RoleKey, rpo.serviceName); err != nil {
		rpo.logger.Fatal("unable to set ring pop labels", tag.Error(err))
	}

	for _, ring := range rpo.rings {
		ring.Start()
	}
}

func (rpo *RingpopMonitor) Stop() {
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

	rpo.ringpopWrapper.Stop()
}

func (rpo *RingpopMonitor) WhoAmI() (*HostInfo, error) {
	address, err := rpo.ringpopWrapper.WhoAmI()
	if err != nil {
		return nil, err
	}
	labels, err := rpo.ringpopWrapper.Labels()
	if err != nil {
		return nil, err
	}
	return NewHostInfo(address, labels.AsMap()), nil
}

func (rpo *RingpopMonitor) EvictSelf() error {
	return rpo.ringpopWrapper.SelfEvict()
}

func (rpo *RingpopMonitor) getRing(service string) (*ringpopServiceResolver, error) {
	ring, found := rpo.rings[service]
	if !found {
		return nil, fmt.Errorf("service %q is not tracked by Monitor", service)
	}
	return ring, nil
}

func (rpo *RingpopMonitor) Lookup(service string, key string) (*HostInfo, error) {
	ring, err := rpo.getRing(service)
	if err != nil {
		return nil, err
	}
	return ring.Lookup(key)
}

func (rpo *RingpopMonitor) AddListener(service string, name string, notifyChannel chan<- *ChangedEvent) error {
	ring, err := rpo.getRing(service)
	if err != nil {
		return err
	}
	return ring.AddListener(name, notifyChannel)
}

func (rpo *RingpopMonitor) RemoveListener(service string, name string) error {
	ring, err := rpo.getRing(service)
	if err != nil {
		return err
	}
	return ring.RemoveListener(name)
}

func (rpo *RingpopMonitor) Members(service string) ([]*HostInfo, error) {
	ring, err := rpo.getRing(service)
	if err != nil {
		return nil, err
	}
	return ring.Members(), nil
}

func (rpo *RingpopMonitor) MemberCount(service string) (int, error) {
	ring, err := rpo.getRing(service)
	if err != nil {
		return 0, err
	}
	return ring.MemberCount(), nil
}
