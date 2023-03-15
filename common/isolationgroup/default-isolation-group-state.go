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

package isolationgroup

import (
	"context"
	"fmt"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"sync"
	"sync/atomic"
	"time"
)

type defaultIsolationGroupStateHandler struct {
	status                     int32
	done                       chan bool
	log                        log.Logger
	domainCache                cache.DomainCache
	globalIsolationGroupDrains persistence.GlobalIsolationGroupDrains
	config                     Config
	subscriptionMu             sync.Mutex
	valuesMu                   sync.RWMutex
	lastSeen                   *IsolationGroups
	updateCB                   func()
	// subscriptions is a map of domains->subscription-keys-> subscription channels
	// for notifying when there's a state change
	subscriptions map[string]map[string]chan<- ChangeEvent
}

func NewDefaultIsolationGroupStateWatcher(
	logger log.Logger,
	domainCache cache.DomainCache,
	config Config,
	globalIsolationGroupDrains persistence.GlobalIsolationGroupDrains,
	done chan bool,
) State {
	return &defaultIsolationGroupStateHandler{
		status:                     common.DaemonStatusInitialized,
		log:                        logger,
		config:                     config,
		domainCache:                domainCache,
		globalIsolationGroupDrains: globalIsolationGroupDrains,
		subscriptionMu:             sync.Mutex{},
		subscriptions:              make(map[string]map[string]chan<- ChangeEvent),
	}
}

func (z *defaultIsolationGroupStateHandler) GetByDomainID(ctx context.Context, domainID string) (*IsolationGroups, error) {
	domain, err := z.domainCache.GetDomainByID(domainID)
	if err != nil {
		return nil, fmt.Errorf("could not resolve domain in isolationGroup handler: %w", err)
	}
	return z.Get(ctx, domain.GetInfo().Name)
}

// Get the statue of a isolationGroup, with respect to both domain and global drains. Domain-specific drains override global config
// will return nil, nil when it is not enabled
func (z *defaultIsolationGroupStateHandler) Get(ctx context.Context, domain string) (*IsolationGroups, error) {
	if !z.config.zonalPartitioningEnabledForDomain(domain) {
		return nil, nil
	}

	domainData, err := z.domainCache.GetDomain(domain)
	if err != nil {
		return nil, fmt.Errorf("could not resolve domain in isolationGroup handler: %w", err)
	}
	domainState := domainData.GetInfo().IsolationGroupConfig
	globalState, err := z.globalIsolationGroupDrains.GetClusterDrains(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not resolve global drains in isolationGroup handler: %w", err)
	}

	ig := &IsolationGroups{
		Global: globalState,
		Domain: domainState,
	}

	return ig, nil
}

func (z *defaultIsolationGroupStateHandler) checkIfChanged() {
	// todo (david.porter)
	// check new values against existing cached ones
	// get the difference
	// if any difference, notify subscribers for whom the change is applicable
	// ie, global changes for all, domain changes for the domain-listeners
	panic("not implemented")
}

func (z *defaultIsolationGroupStateHandler) Subscribe(domainID, key string, notifyChannel chan<- ChangeEvent) error {
	z.subscriptionMu.Lock()
	defer z.subscriptionMu.Unlock()

	panic("not implemented")
	return nil
}

func (z *defaultIsolationGroupStateHandler) Unsubscribe(domainID, key string) error {
	z.subscriptionMu.Lock()
	defer z.subscriptionMu.Unlock()
	panic("not implemented")
	return nil
}

func (z *defaultIsolationGroupStateHandler) pollForChanges() {
	ticker := time.NewTicker(time.Duration(z.config.UpdateFrequency()) * time.Second)
	for {
		select {
		case <-z.done:
			return
		case <-ticker.C:
			z.checkIfChanged()
		}
	}
}

// Start the state handler
func (z *defaultIsolationGroupStateHandler) Start() {
	if !atomic.CompareAndSwapInt32(&z.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}
	go z.updateCB()
}

func (z *defaultIsolationGroupStateHandler) Stop() {
	if !atomic.CompareAndSwapInt32(&z.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}
	close(z.done)
}
