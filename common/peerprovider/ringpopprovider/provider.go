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
	"net"
	"strconv"
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
		channel    tchannel.Channel

		mu          sync.RWMutex
		subscribers map[string]chan<- *membership.ChangedEvent
	}
)

const (
	// roleKey label is set by every single service as soon as it bootstraps its
	// ringpop instance. The data for this key is the service name
	roleKey      = "serviceName"
	portTchannel = "tchannel"
	portgRPC     = "grpc"
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

	discoveryProvider, err := newDiscoveryProvider(config, logger)
	if err != nil {
		return nil, fmt.Errorf("ringpop discovery provider: %w", err)
	}

	bootstrapOpts := &swim.BootstrapOptions{
		MaxJoinDuration:  config.MaxJoinDuration,
		DiscoverProvider: discoveryProvider,
	}

	rp, err := ringpop.New(config.Name, ringpop.Channel(channel.(*tcg.Channel)))
	if err != nil {
		return nil, fmt.Errorf("ringpop instance creation: %w", err)
	}

	return NewRingpopProvider(service, rp, bootstrapOpts, channel, logger), nil
}

// NewRingpopProvider sets up ringpop based peer provider
func NewRingpopProvider(
	service string,
	rp *ringpop.Ringpop,
	bootstrapOpts *swim.BootstrapOptions,
	channel tchannel.Channel,
	logger log.Logger,
) *Provider {
	return &Provider{
		service:     service,
		status:      common.DaemonStatusInitialized,
		bootParams:  bootstrapOpts,
		logger:      logger,
		channel:     channel,
		ringpop:     rp,
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

	// set tchannel port to labels
	_, port, err := net.SplitHostPort(r.channel.PeerInfo().HostPort)
	if err != nil {
		r.logger.Fatal("unable get tchannel port", tag.Error(err))
	}

	if err = labels.Set(portTchannel, port); err != nil {
		r.logger.Fatal("unable to set ringpop tchannel label", tag.Error(err))
	}

	if err = labels.Set(roleKey, r.service); err != nil {
		r.logger.Fatal("unable to set ringpop role label", tag.Error(err))
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
	r.mu.RLock()
	defer r.mu.RUnlock()

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
func (r *Provider) GetMembers(service string) ([]membership.HostInfo, error) {
	var res []membership.HostInfo

	// filter member by service name, add port info to Hostinfo if they are present
	memberData := func(member swim.Member) bool {
		portMap := make(membership.PortMap)
		if v, ok := member.Label(roleKey); !ok || v != service {
			return false
		}

		if v, ok := member.Label(portTchannel); ok {
			port, err := labelToPort(v)
			if err != nil {
				r.logger.Warn("tchannel port cannot be converted", tag.Error(err), tag.Value(v))
			} else {
				portMap[portTchannel] = port
			}
		}

		if v, ok := member.Label(portgRPC); ok {
			port, err := labelToPort(v)
			if err != nil {
				r.logger.Warn("grpc port cannot be converted", tag.Error(err), tag.Value(v))
			} else {
				portMap[portgRPC] = port
			}
		}

		res = append(res, membership.NewDetailedHostInfo(member.GetAddress(), member.Identity(), portMap))

		return true
	}
	_, err := r.ringpop.GetReachableMembers(memberData)
	if err != nil {
		return nil, fmt.Errorf("ringpop get members: %w", err)
	}

	return res, nil
}

// WhoAmI returns address of this instance
func (r *Provider) WhoAmI() (membership.HostInfo, error) {
	address, err := r.ringpop.WhoAmI()
	if err != nil {
		return membership.HostInfo{}, fmt.Errorf("ringpop doesn't know Who Am I: %w", err)
	}
	return membership.NewHostInfo(address), nil
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
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.subscribers[name]
	if ok {
		return fmt.Errorf("%q already subscribed to ringpop provider", name)
	}

	r.subscribers[name] = notifyChannel
	return nil
}

func labelToPort(label string) (uint16, error) {
	port, err := strconv.ParseInt(label, 0, 16)
	if err != nil {
		return 0, err
	}
	return uint16(port), nil
}
