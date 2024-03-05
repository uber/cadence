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
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgryski/go-farm"
	"github.com/uber/ringpop-go/hashring"
	"github.com/uber/ringpop-go/membership"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination peerprovider_mock.go github.com/uber/cadence/common/membership PeerProvider

// ErrInsufficientHosts is thrown when there are not enough hosts to serve the request
var ErrInsufficientHosts = &types.InternalServiceError{Message: "Not enough hosts to serve the request"}

const (
	minRefreshInternal     = time.Second * 4
	defaultRefreshInterval = time.Second * 10
	replicaPoints          = 100
)

// PeerProvider is used to retrieve membership information from provider
type PeerProvider interface {
	common.Daemon

	GetMembers(service string) ([]HostInfo, error)
	WhoAmI() (HostInfo, error)
	SelfEvict() error
	Subscribe(name string, notifyChannel chan<- *ChangedEvent) error
}

type ring struct {
	status       int32
	service      string
	peerProvider PeerProvider
	refreshChan  chan *ChangedEvent
	shutdownCh   chan struct{}
	shutdownWG   sync.WaitGroup
	scope        metrics.Scope
	logger       log.Logger

	value atomic.Value // this stores the current hashring

	members struct {
		sync.RWMutex
		refreshed time.Time
		keys      map[string]HostInfo // for mapping ip:port to HostInfo
	}

	subscribers struct {
		sync.RWMutex
		keys map[string]chan<- *ChangedEvent
	}
}

func newHashring(
	service string,
	provider PeerProvider,
	logger log.Logger,
	scope metrics.Scope,
) *ring {
	ring := &ring{
		status:       common.DaemonStatusInitialized,
		service:      service,
		peerProvider: provider,
		shutdownCh:   make(chan struct{}),
		logger:       logger,
		refreshChan:  make(chan *ChangedEvent),
		scope:        scope,
	}

	ring.members.keys = make(map[string]HostInfo)
	ring.subscribers.keys = make(map[string]chan<- *ChangedEvent)

	ring.value.Store(emptyHashring())
	return ring
}

func emptyHashring() *hashring.HashRing {
	return hashring.New(farm.Fingerprint32, replicaPoints)
}

// Start starts the hashring
func (r *ring) Start() {
	if !atomic.CompareAndSwapInt32(
		&r.status,
		common.DaemonStatusInitialized,
		common.DaemonStatusStarted,
	) {
		return
	}
	if err := r.peerProvider.Subscribe(r.service, r.refreshChan); err != nil {
		r.logger.Fatal("subscribing to peer provider", tag.Error(err))
	}

	if err := r.refresh(); err != nil {
		r.logger.Fatal("failed to start service resolver", tag.Error(err))
	}

	r.shutdownWG.Add(1)
	go r.refreshRingWorker()
}

// Stop stops the resolver
func (r *ring) Stop() {
	if !atomic.CompareAndSwapInt32(
		&r.status,
		common.DaemonStatusStarted,
		common.DaemonStatusStopped,
	) {
		return
	}

	r.peerProvider.Stop()
	r.value.Store(emptyHashring())

	r.subscribers.Lock()
	defer r.subscribers.Unlock()
	r.subscribers.keys = make(map[string]chan<- *ChangedEvent)
	close(r.shutdownCh)

	if success := common.AwaitWaitGroup(&r.shutdownWG, time.Minute); !success {
		r.logger.Warn("service resolver timed out on shutdown.")
	}
}

// Lookup finds the host in the ring responsible for serving the given key
func (r *ring) Lookup(
	key string,
) (HostInfo, error) {
	addr, found := r.ring().Lookup(key)
	if !found {
		select {
		case r.refreshChan <- &ChangedEvent{}:
		default:
		}
		return HostInfo{}, ErrInsufficientHosts
	}
	r.members.RLock()
	defer r.members.RUnlock()
	host, ok := r.members.keys[addr]
	if !ok {
		return HostInfo{}, fmt.Errorf("host not found in member keys, host: %q", addr)
	}
	return host, nil
}

// Subscribe registers callback watcher.
// All subscribers are notified about ring change.
func (r *ring) Subscribe(
	service string,
	notifyChannel chan<- *ChangedEvent,
) error {
	r.subscribers.Lock()
	defer r.subscribers.Unlock()

	_, ok := r.subscribers.keys[service]
	if ok {
		return fmt.Errorf("service %q already subscribed", service)
	}

	r.subscribers.keys[service] = notifyChannel
	return nil
}

// Unsubscribe removes subscriber
func (r *ring) Unsubscribe(
	name string,
) error {
	r.subscribers.Lock()
	defer r.subscribers.Unlock()
	delete(r.subscribers.keys, name)
	return nil
}

// MemberCount returns number of hosts in a ring
func (r *ring) MemberCount() int {
	return r.ring().ServerCount()
}

func (r *ring) Members() []HostInfo {
	servers := r.ring().Servers()

	var hosts = make([]HostInfo, 0, len(servers))
	r.members.RLock()
	defer r.members.RUnlock()
	for _, s := range servers {
		host, ok := r.members.keys[s]
		if !ok {
			r.logger.Warn("host not found in hashring keys", tag.Address(s))
			continue
		}
		hosts = append(hosts, host)
	}

	return hosts
}

func (r *ring) refresh() error {
	if r.members.refreshed.After(time.Now().Add(-minRefreshInternal)) {
		// refreshed too frequently
		return nil
	}

	members, err := r.peerProvider.GetMembers(r.service)
	if err != nil {
		return fmt.Errorf("getting members from peer provider: %w", err)
	}

	r.members.Lock()
	defer r.members.Unlock()
	newMembersMap, changed := r.compareMembers(members)
	if !changed {
		return nil
	}

	ring := emptyHashring()
	ring.AddMembers(castToMembers(members)...)

	r.members.keys = newMembersMap
	r.members.refreshed = time.Now()
	r.value.Store(ring)
	r.logger.Info("refreshed ring members", tag.Value(members))

	return nil
}

func (r *ring) refreshRingWorker() {
	defer r.shutdownWG.Done()

	refreshTicker := time.NewTicker(defaultRefreshInterval)
	defer refreshTicker.Stop()
	for {
		select {
		case <-r.shutdownCh:
			return
		case <-r.refreshChan: // local signal or signal from provider
			if err := r.refresh(); err != nil {
				r.logger.Error("refreshing ring", tag.Error(err))
			}
		case <-refreshTicker.C: // periodically refresh membership
			r.emitHashIdentifier()
			if err := r.refresh(); err != nil {
				r.logger.Error("periodically refreshing ring", tag.Error(err))
			}
		}
	}
}

func (r *ring) ring() *hashring.HashRing {
	return r.value.Load().(*hashring.HashRing)
}

func (r *ring) emitHashIdentifier() float64 {
	members, err := r.peerProvider.GetMembers(r.service)
	if err != nil {
		r.logger.Error("Observed a problem getting peer members while emitting hash identifier metrics", tag.Error(err))
		return -1
	}
	self, err := r.peerProvider.WhoAmI()
	if err != nil {
		r.logger.Error("Observed a problem looking up self from the membership provider while emitting hash identifier metrics", tag.Error(err))
		self = HostInfo{
			identity: "unknown",
		}
	}

	sort.Slice(members, func(i int, j int) bool {
		return members[i].addr > members[j].addr
	})
	var sb strings.Builder
	for i := range members {
		sb.WriteString(members[i].addr)
		sb.WriteString("\n")
	}
	hashedView := farm.Hash32([]byte(sb.String()))
	// Trimming the metric because collisions are unlikely and I didn't want to use the full Float64
	// in-case it overflowed something. The number itself is meaningless, so additional precision
	// doesn't really give any advantage, besides reducing the risk of collision
	trimmedForMetric := float64(hashedView % 1000)
	r.logger.Debug("Hashring view", tag.Dynamic("hashring-view", sb.String()), tag.Dynamic("trimmed-hash-id", trimmedForMetric), tag.Service(r.service))
	r.scope.Tagged(
		metrics.ServiceTag(r.service),
		metrics.HostTag(self.identity),
	).UpdateGauge(metrics.HashringViewIdentifier, trimmedForMetric)
	return trimmedForMetric
}

func (r *ring) compareMembers(members []HostInfo) (map[string]HostInfo, bool) {
	changed := false
	newMembersMap := make(map[string]HostInfo, len(members))
	for _, member := range members {
		newMembersMap[member.GetAddress()] = member
		if _, ok := r.members.keys[member.GetAddress()]; !ok {
			changed = true
		}
	}
	for addr := range r.members.keys {
		if _, ok := newMembersMap[addr]; !ok {
			changed = true
			break
		}
	}
	return newMembersMap, changed
}

func castToMembers[T membership.Member](members []T) []membership.Member {
	result := make([]membership.Member, 0, len(members))
	for _, h := range members {
		result = append(result, h)
	}
	return result
}
