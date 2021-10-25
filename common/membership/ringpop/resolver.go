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
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgryski/go-farm"
	"github.com/uber/ringpop-go/events"
	"github.com/uber/ringpop-go/hashring"
	"github.com/uber/ringpop-go/swim"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
)

const (
	minRefreshInternal     = time.Second * 4
	defaultRefreshInterval = time.Second * 10
	replicaPoints          = 100
)

type ringpopServiceResolver struct {
	status      int32
	service     string
	rp          *RingPop
	refreshChan chan struct{}
	shutdownCh  chan struct{}
	shutdownWG  sync.WaitGroup
	logger      log.Logger

	ringValue atomic.Value // this stores the current hashring

	refreshLock     sync.Mutex
	lastRefreshTime time.Time
	membersMap      map[string]struct{} // for de-duping change notifications

	listenerLock sync.RWMutex
	listeners    map[string]chan<- *membership.ChangedEvent
}

var _ membership.ServiceResolver = (*ringpopServiceResolver)(nil)

func newRingpopServiceResolver(
	service string,
	rp *RingPop,
	logger log.Logger,
) *ringpopServiceResolver {

	resolver := &ringpopServiceResolver{
		status:      common.DaemonStatusInitialized,
		service:     service,
		rp:          rp,
		refreshChan: make(chan struct{}),
		shutdownCh:  make(chan struct{}),
		logger:      logger.WithTags(tag.ComponentServiceResolver),
		membersMap:  make(map[string]struct{}),
		listeners:   make(map[string]chan<- *membership.ChangedEvent),
	}
	resolver.ringValue.Store(newHashRing())
	return resolver
}

func newHashRing() *hashring.HashRing {
	return hashring.New(farm.Fingerprint32, replicaPoints)
}

// Start starts the oracle
func (r *ringpopServiceResolver) Start() {
	if !atomic.CompareAndSwapInt32(
		&r.status,
		common.DaemonStatusInitialized,
		common.DaemonStatusStarted,
	) {
		return
	}

	r.rp.ringpop.AddListener(r)
	if err := r.refresh(); err != nil {
		r.logger.Fatal("unable to start ringpop service resolver", tag.Error(err))
	}

	r.shutdownWG.Add(1)
	go r.refreshRingWorker()
}

// Stop stops the resolver
func (r *ringpopServiceResolver) Stop() {
	if !atomic.CompareAndSwapInt32(
		&r.status,
		common.DaemonStatusStarted,
		common.DaemonStatusStopped,
	) {
		return
	}

	r.listenerLock.Lock()
	defer r.listenerLock.Unlock()
	r.rp.ringpop.RemoveListener(r)
	r.ringValue.Store(newHashRing())
	r.listeners = make(map[string]chan<- *membership.ChangedEvent)
	close(r.shutdownCh)

	if success := common.AwaitWaitGroup(&r.shutdownWG, time.Minute); !success {
		r.logger.Warn("service resolver timed out on shutdown.")
	}
}

// Lookup finds the host in the ring responsible for serving the given key
func (r *ringpopServiceResolver) Lookup(
	key string,
) (*membership.HostInfo, error) {

	addr, found := r.ring().Lookup(key)
	if !found {
		select {
		case r.refreshChan <- struct{}{}:
		default:
		}
		return nil, membership.ErrInsufficientHosts
	}
	return membership.NewHostInfo(addr, r.getLabelsMap()), nil
}

func (r *ringpopServiceResolver) AddListener(
	name string,
	notifyChannel chan<- *membership.ChangedEvent,
) error {
	r.listenerLock.Lock()
	defer r.listenerLock.Unlock()
	if _, ok := r.listeners[name]; ok {
		return fmt.Errorf("listener for service %q already exist", name)
	}
	r.listeners[name] = notifyChannel
	return nil
}

func (r *ringpopServiceResolver) RemoveListener(
	name string,
) error {

	r.listenerLock.Lock()
	defer r.listenerLock.Unlock()
	_, ok := r.listeners[name]
	if !ok {
		return nil
	}
	delete(r.listeners, name)
	return nil
}

func (r *ringpopServiceResolver) MemberCount() int {
	return r.ring().ServerCount()
}

func (r *ringpopServiceResolver) Members() []*membership.HostInfo {
	var servers []*membership.HostInfo
	for _, s := range r.ring().Servers() {
		servers = append(servers, membership.NewHostInfo(s, r.getLabelsMap()))
	}

	return servers
}

// HandleEvent handles updates from ringpop
func (r *ringpopServiceResolver) HandleEvent(
	event events.Event,
) {

	// We only care about RingChangedEvent
	e, ok := event.(events.RingChangedEvent)
	if ok {
		r.logger.Info("Received a ring changed event")
		// Note that we receive events asynchronously, possibly out of order.
		// We cannot rely on the content of the event, rather we load everything
		// from ringpop when we get a notification that something changed.
		if err := r.refresh(); err != nil {
			r.logger.Error("error refreshing ring when receiving a ring changed event", tag.Error(err))
		}
		r.emitEvent(e)
	}
}

func (r *ringpopServiceResolver) refresh() error {
	r.refreshLock.Lock()
	defer r.refreshLock.Unlock()
	return r.refreshNoLock()
}

func (r *ringpopServiceResolver) refreshWithBackoff() error {
	r.refreshLock.Lock()
	defer r.refreshLock.Unlock()
	if r.lastRefreshTime.After(time.Now().Add(-minRefreshInternal)) {
		// refresh too frequently
		return nil
	}
	return r.refreshNoLock()
}

func (r *ringpopServiceResolver) refreshNoLock() error {
	addrs, err := r.rp.ringpop.GetReachableMembers(swim.MemberWithLabelAndValue(membership.RoleKey, r.service))
	if err != nil {
		return err
	}

	newMembersMap, changed := r.compareMembers(addrs)
	if !changed {
		return nil
	}

	ring := newHashRing()
	for _, addr := range addrs {
		host := membership.NewHostInfo(addr, r.getLabelsMap())
		ring.AddMembers(host)
	}

	r.membersMap = newMembersMap
	r.lastRefreshTime = time.Now()
	r.ringValue.Store(ring)
	r.logger.Info("Current reachable members", tag.Addresses(addrs))
	return nil
}

func (r *ringpopServiceResolver) emitEvent(
	rpEvent events.RingChangedEvent,
) {

	// Marshall the event object into the required type
	event := &membership.ChangedEvent{}
	for _, addr := range rpEvent.ServersAdded {
		event.HostsAdded = append(event.HostsAdded, membership.NewHostInfo(addr, r.getLabelsMap()))
	}
	for _, addr := range rpEvent.ServersRemoved {
		event.HostsRemoved = append(event.HostsRemoved, membership.NewHostInfo(addr, r.getLabelsMap()))
	}
	for _, addr := range rpEvent.ServersUpdated {
		event.HostsUpdated = append(event.HostsUpdated, membership.NewHostInfo(addr, r.getLabelsMap()))
	}

	// Notify listeners
	r.listenerLock.RLock()
	defer r.listenerLock.RUnlock()

	for name, ch := range r.listeners {
		select {
		case ch <- event:
		default:
			r.logger.Error("Failed to send listener notification, channel full", tag.ListenerName(name))
		}
	}
}

func (r *ringpopServiceResolver) refreshRingWorker() {
	defer r.shutdownWG.Done()

	refreshTicker := time.NewTicker(defaultRefreshInterval)
	defer refreshTicker.Stop()

	for {
		select {
		case <-r.shutdownCh:
			return
		case <-r.refreshChan:
			if err := r.refreshWithBackoff(); err != nil {
				r.logger.Error("refreshing ring", tag.Error(err))
			}
		case <-refreshTicker.C:
			if err := r.refreshWithBackoff(); err != nil {
				r.logger.Error("error periodically refreshing ring", tag.Error(err))
			}
		}
	}
}

func (r *ringpopServiceResolver) ring() *hashring.HashRing {
	return r.ringValue.Load().(*hashring.HashRing)
}

func (r *ringpopServiceResolver) getLabelsMap() map[string]string {
	labels := make(map[string]string)
	labels[membership.RoleKey] = r.service
	return labels
}

func (r *ringpopServiceResolver) compareMembers(addrs []string) (map[string]struct{}, bool) {
	changed := false
	newMembersMap := make(map[string]struct{}, len(addrs))
	for _, addr := range addrs {
		newMembersMap[addr] = struct{}{}
		if _, ok := r.membersMap[addr]; !ok {
			changed = true
		}
	}
	for addr := range r.membersMap {
		if _, ok := newMembersMap[addr]; !ok {
			changed = true
			break
		}
	}
	return newMembersMap, changed
}
