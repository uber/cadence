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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination resolver_mock.go -self_package github.com/uber/cadence/common/membership

package membership

import (
	"fmt"
	"sync/atomic"

	"github.com/uber/cadence/common/service"
	"github.com/uber/ringpop-go"
	"github.com/uber/ringpop-go/swim"
	tcg "github.com/uber/tchannel-go"
	"go.uber.org/yarpc/transport/tchannel"

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

	// Resolver provides membership information for all cadence services.
	Resolver interface {
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
		Subscribe(service, name string, notifyChannel chan<- *ChangedEvent) error
		// RemoveListener removes a listener for this service.
		Unsubscribe(service, name string) error
		// MemberCount returns host count in a hashring
		MemberCount(service string) (int, error)
		// Members returns all host addresses in a hashring
		Members(service string) ([]*HostInfo, error)
	}
)

type RingpopResolver struct {
	status int32

	serviceName    string
	ringpopWrapper *RingpopWrapper
	rings          map[string]*ringpopHashring
	logger         log.Logger
}

var _ Resolver = (*RingpopResolver)(nil)

// NewResolver builds a ringpop monitor conforming
// to the underlying configuration
func NewResolver(
	config *RingpopConfig,
	channel tchannel.Channel,
	serviceName string,
	logger log.Logger,
) (*RingpopResolver, error) {

	if err := config.validate(); err != nil {
		return nil, err
	}

	rp, err := ringpop.New(config.Name, ringpop.Channel(channel.(*tcg.Channel)))
	if err != nil {
		return nil, fmt.Errorf("ringpop creation failed: %v", err)
	}

	discoveryProvider, err := newDiscoveryProvider(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get discovery provider %v", err)
	}

	bootstrapOpts := &swim.BootstrapOptions{
		MaxJoinDuration:  config.MaxJoinDuration,
		DiscoverProvider: discoveryProvider,
	}
	rpw := NewRingpopWraper(rp, bootstrapOpts, logger)

	return NewRingpopResolver(serviceName, service.List, rpw, logger), nil

}

// NewRingpopResolver returns a ringpop-based membership monitor
func NewRingpopResolver(
	serviceName string,
	services []string,
	rp *RingpopWrapper,
	logger log.Logger,
) *RingpopResolver {

	rpo := &RingpopResolver{
		status:         common.DaemonStatusInitialized,
		serviceName:    serviceName,
		ringpopWrapper: rp,
		logger:         logger,
		rings:          make(map[string]*ringpopHashring),
	}
	for _, s := range services {
		rpo.rings[s] = newRingpopHashring(s, rp, logger)
	}
	return rpo
}

func (rpo *RingpopResolver) Start() {
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

func (rpo *RingpopResolver) Stop() {
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

func (rpo *RingpopResolver) WhoAmI() (*HostInfo, error) {
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

func (rpo *RingpopResolver) EvictSelf() error {
	return rpo.ringpopWrapper.SelfEvict()
}

func (rpo *RingpopResolver) getRing(service string) (*ringpopHashring, error) {
	ring, found := rpo.rings[service]
	if !found {
		return nil, fmt.Errorf("service %q is not tracked by Resolver", service)
	}
	return ring, nil
}

func (rpo *RingpopResolver) Lookup(service string, key string) (*HostInfo, error) {
	ring, err := rpo.getRing(service)
	if err != nil {
		return nil, err
	}
	return ring.Lookup(key)
}

func (rpo *RingpopResolver) Subscribe(service string, name string, notifyChannel chan<- *ChangedEvent) error {
	ring, err := rpo.getRing(service)
	if err != nil {
		return err
	}
	return ring.AddListener(name, notifyChannel)
}

func (rpo *RingpopResolver) Unsubscribe(service string, name string) error {
	ring, err := rpo.getRing(service)
	if err != nil {
		return err
	}
	return ring.RemoveListener(name)
}

func (rpo *RingpopResolver) Members(service string) ([]*HostInfo, error) {
	ring, err := rpo.getRing(service)
	if err != nil {
		return nil, err
	}
	return ring.Members(), nil
}

func (rpo *RingpopResolver) MemberCount(service string) (int, error) {
	ring, err := rpo.getRing(service)
	if err != nil {
		return 0, err
	}
	return ring.MemberCount(), nil
}
