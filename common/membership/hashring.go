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
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgryski/go-farm"
	"github.com/uber/ringpop-go/hashring"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
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

	GetMembers(service string) ([]string, error)
	WhoAmI() (string, error)
	SelfEvict() error
	Subscribe(name string, notifyChannel chan<- *ChangedEvent) error
}

// HostInfo is a type that contains the info about a cadence host
type HostInfo struct {
	addr string // ip:port
}

type ring struct {
	status       int32
	service      string
	peerProvider PeerProvider
	refreshChan  chan *ChangedEvent
	shutdownCh   chan struct{}
	shutdownWG   sync.WaitGroup
	logger       log.Logger

	value atomic.Value // this stores the current hashring

	members struct {
		refreshed time.Time

		sync.Mutex
		keys map[string]struct{} // for de-duping change notifications

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
) *ring {
	hashring := &ring{
		status:       common.DaemonStatusInitialized,
		service:      service,
		peerProvider: provider,
		shutdownCh:   make(chan struct{}),
		logger:       logger.WithTags(tag.ComponentServiceResolver),
		refreshChan:  make(chan *ChangedEvent),
	}

	hashring.members.keys = make(map[string]struct{})
	hashring.subscribers.keys = make(map[string]chan<- *ChangedEvent)

	hashring.value.Store(emptyHashring())
	return hashring
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
	r.peerProvider.Subscribe(r.service, r.refreshChan)

	if err := r.refreshLocked(); err != nil {
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
) (*HostInfo, error) {
	addr, found := r.ring().Lookup(key)
	if !found {
		select {
		case r.refreshChan <- &ChangedEvent{}:
		default:
		}
		return nil, ErrInsufficientHosts
	}
	return NewHostInfo(addr), nil
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

func (r *ring) Members() []*HostInfo {
	var servers []*HostInfo
	for _, s := range r.ring().Servers() {
		servers = append(servers, NewHostInfo(s))
	}

	return servers
}

func (r *ring) refreshLocked() error {
	r.members.Lock()
	defer r.members.Unlock()
	addrs, err := r.peerProvider.GetMembers(r.service)

	if err != nil {
		return err
	}

	newMembersMap, changed := r.compareMembers(addrs)
	if !changed {
		return nil
	}

	ring := emptyHashring()
	for _, addr := range addrs {
		host := NewHostInfo(addr)
		ring.AddMembers(host)
	}

	r.members.keys = newMembersMap
	r.members.refreshed = time.Now()
	r.value.Store(ring)
	r.logger.Info(
		fmt.Sprintf("refreshed ring members for %s", r.service),
		tag.Addresses(addrs),
	)
	return nil
}

func (r *ring) refreshWithBackoff() error {
	if r.members.refreshed.After(time.Now().Add(-minRefreshInternal)) {
		// refreshLocked too frequently
		return nil
	}
	return r.refreshLocked()
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
			if err := r.refreshWithBackoff(); err != nil {
				r.logger.Error("refreshing ring", tag.Error(err))
			}
		case <-refreshTicker.C: // periodically refreshLocked membership
			if err := r.refreshWithBackoff(); err != nil {
				r.logger.Error("periodically refreshing ring", tag.Error(err))
			}
		}
	}
}

func (r *ring) ring() *hashring.HashRing {
	return r.value.Load().(*hashring.HashRing)
}

func (r *ring) compareMembers(addrs []string) (map[string]struct{}, bool) {
	changed := false
	newMembersMap := make(map[string]struct{}, len(addrs))
	for _, addr := range addrs {
		newMembersMap[addr] = struct{}{}
		if _, ok := r.members.keys[addr]; !ok {
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

// NewHostInfo creates a new HostInfo instance
func NewHostInfo(addr string) *HostInfo {
	return &HostInfo{
		addr: addr,
	}
}

// GetAddress returns the ip:port address
func (hi *HostInfo) GetAddress() string {
	return hi.addr
}

// Identity implements ringpop's Membership interface
func (hi *HostInfo) Identity() string {
	// for now we just use the address as the identity
	return hi.addr
}

// Label implements ringpop's Membership interface
func (hi *HostInfo) Label(key string) (value string, has bool) {
	return "", false
}

// SetLabel sets the label.
func (hi *HostInfo) SetLabel(key string, value string) {
}
