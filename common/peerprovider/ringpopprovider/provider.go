// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ringpopprovider

import (
	"fmt"
	"sync"
	"sync/atomic"

	"go.uber.org/yarpc/transport/tchannel"

	"github.com/uber/ringpop-go"
	"github.com/uber/ringpop-go/events"
	"github.com/uber/ringpop-go/swim"
	tcg "github.com/uber/tchannel-go"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
)

type (
	// Provider use ringpop to announce membership changes
	Provider struct {
		status     int32
		service    string
		ringpop    *ringpop.Ringpop
		bootParams *swim.BootstrapOptions
		logger     log.Logger

		smu         sync.RWMutex
		subscribers map[string]chan<- *membership.ChangedEvent
	}
)

const (
	// RoleKey label is set by every single service as soon as it bootstraps its
	// ringpop instance. The data for this key is the service name
	RoleKey = "serviceName"
)

var _ membership.PeerProvider = (*Provider)(nil)

func New(
	service string,
	config *Config,
	channel tchannel.Channel,
	logger log.Logger,
) (*Provider, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	rp, err := ringpop.New(config.Name, ringpop.Channel(channel.(*tcg.Channel)))
	if err != nil {
		return nil, fmt.Errorf("ringpop creation failed: %v", err)
	}

	discoveryProvider, err := newDiscoveryProvider(config, logger)
	if err != nil {
		return nil, fmt.Errorf("ringpop discovery provider: %v", err)
	}

	bootstrapOpts := &swim.BootstrapOptions{
		MaxJoinDuration:  config.MaxJoinDuration,
		DiscoverProvider: discoveryProvider,
	}

	return NewRingpopProvider(service, rp, bootstrapOpts, logger), nil
}

// NewRingpopProvider sets up ringpop based peer provider
func NewRingpopProvider(service string,
	rp *ringpop.Ringpop,
	bootstrapOpts *swim.BootstrapOptions,
	logger log.Logger,
) *Provider {
	return &Provider{
		service:     service,
		status:      common.DaemonStatusInitialized,
		ringpop:     rp,
		bootParams:  bootstrapOpts,
		logger:      logger,
		subscribers: map[string]chan<- *membership.ChangedEvent{},
	}
}

// Start starts ringpop
func (r *Provider) Start() {
	if !atomic.CompareAndSwapInt32(
		&r.status,
		common.DaemonStatusInitialized,
		common.DaemonStatusStarted,
	) {
		return
	}

	_, err := r.ringpop.Bootstrap(r.bootParams)
	if err != nil {
		r.logger.Fatal("unable to bootstrap ringpop", tag.Error(err))
	}

	// Get updates from ringpop ring
	r.ringpop.AddListener(r)

	labels, err := r.ringpop.Labels()
	if err != nil {
		r.logger.Fatal("unable to get ring pop labels", tag.Error(err))
	}

	if err = labels.Set(RoleKey, r.service); err != nil {
		r.logger.Fatal("unable to set ring pop labels", tag.Error(err))
	}
}

// HandleEvent handles updates from ringpop
func (r *Provider) HandleEvent(
	event events.Event,
) {
	// We only care about RingChangedEvent
	e, ok := event.(events.RingChangedEvent)
	if !ok {
		return
	}

	r.logger.Info("Received a ringpop ring changed event")
	// Marshall the event object into the required type
	change := &membership.ChangedEvent{
		HostsAdded:   e.ServersAdded,
		HostsUpdated: e.ServersUpdated,
		HostsRemoved: e.ServersRemoved,
	}

	// Notify subscribers
	r.smu.RLock()
	defer r.smu.RUnlock()

	for name, ch := range r.subscribers {
		select {
		case ch <- change:
		default:
			r.logger.Error("Failed to send listener notification, channel full", tag.Subscriber(name))
		}
	}

}

func (r *Provider) SelfEvict() error {
	return r.ringpop.SelfEvict()
}

// GetMembers returns all hosts with a specified role value
func (r *Provider) GetMembers(service string) ([]string, error) {
	return r.ringpop.GetReachableMembers(swim.MemberWithLabelAndValue(RoleKey, service))
}

// WhoAmI returns address of this instance
func (r *Provider) WhoAmI() (string, error) {
	return r.ringpop.WhoAmI()
}

// Stop stops ringpop
func (r *Provider) Stop() {
	if !atomic.CompareAndSwapInt32(
		&r.status,
		common.DaemonStatusStarted,
		common.DaemonStatusStopped,
	) {
		return
	}

	r.ringpop.Destroy()
}

// Subscribe allows to be subscribed for ring changes
func (r *Provider) Subscribe(name string, notifyChannel chan<- *membership.ChangedEvent) error {
	r.smu.RLock()
	defer r.smu.RUnlock()

	r.subscribers[name] = notifyChannel
	return nil
}
