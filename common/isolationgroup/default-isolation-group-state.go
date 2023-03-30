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
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common/config"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type defaultIsolationGroupStateHandler struct {
	status                     int32
	done                       chan struct{}
	log                        log.Logger
	domainCache                cache.DomainCache
	globalIsolationGroupDrains persistence.ConfigStoreManager
	config                     config.IsolationGroups
	subscriptionMu             sync.Mutex
	valuesMu                   sync.RWMutex
	lastSeen                   *isolationGroups
	updateCB                   func()
	// subscriptions is a map of domains->subscription-keys-> subscription channels
	// for notifying when there's a state change
	subscriptions map[string]map[string]chan<- ChangeEvent
}

func NewDefaultIsolationGroupStateWatcher(
	logger log.Logger,
	config config.IsolationGroups,
	done chan struct{},
) State {
	return &defaultIsolationGroupStateHandler{
		done:           done,
		status:         common.DaemonStatusInitialized,
		log:            logger,
		config:         config,
		subscriptionMu: sync.Mutex{},
		subscriptions:  make(map[string]map[string]chan<- ChangeEvent),
	}
}

func (z *defaultIsolationGroupStateHandler) AvailableIsolationGroupsByDomainID(ctx context.Context, domainID string) (types.IsolationGroupConfiguration, error) {
	state, err := z.getByDomainID(ctx, domainID)
	if err != nil {
		return nil, fmt.Errorf("unable to get isolation group state: %w", err)
	}
	return availableIG(z.config.AllIsolationGroups, state.Global, state.Domain), nil
}

func (z *defaultIsolationGroupStateHandler) IsDrained(ctx context.Context, domain string, isolationGroup string) (bool, error) {
	state, err := z.get(ctx, domain)
	if err != nil {
		return false, fmt.Errorf("could not determine if drained: %w", err)
	}
	return isDrained(isolationGroup, state.Global, state.Domain), nil
}

func (z *defaultIsolationGroupStateHandler) IsDrainedByDomainID(ctx context.Context, domainID string, isolationGroup string) (bool, error) {
	domain, err := z.domainCache.GetDomainByID(domainID)
	if err != nil {
		return false, fmt.Errorf("could not determine if drained: %w", err)
	}
	return z.IsDrained(ctx, domain.GetInfo().Name, isolationGroup)
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

func (z *defaultIsolationGroupStateHandler) UpdateGlobalState(ctx context.Context, in types.UpdateGlobalIsolationGroupsRequest) error {
	return z.globalIsolationGroupDrains.UpdateDynamicConfig(ctx, updateGlobalIGRequest(in), persistence.GlobalIsolationGroupConfig)
}

func (z *defaultIsolationGroupStateHandler) GetGlobalState(ctx context.Context) (*types.GetGlobalIsolationGroupsResponse, error) {
	res, err := z.globalIsolationGroupDrains.FetchDynamicConfig(ctx, persistence.GlobalIsolationGroupConfig)
	if err != nil {
		return nil, nil
	}
	return getGlobalIGResponse(*res), nil
}

func (z *defaultIsolationGroupStateHandler) getByDomainID(ctx context.Context, domainID string) (*isolationGroups, error) {
	domain, err := z.domainCache.GetDomainByID(domainID)
	if err != nil {
		return nil, fmt.Errorf("could not resolve domain in isolationGroup handler: %w", err)
	}
	return z.get(ctx, domain.GetInfo().Name)
}

// Get the statue of a isolationGroup, with respect to both domain and global drains. Domain-specific drains override global config
// will return nil, nil when it is not enabled
func (z *defaultIsolationGroupStateHandler) get(ctx context.Context, domain string) (*isolationGroups, error) {
	if !z.config.IsolationGroupEnabled(domain) {
		return nil, nil
	}

	domainData, err := z.domainCache.GetDomain(domain)
	if err != nil {
		return nil, fmt.Errorf("could not resolve domain in isolationGroup handler: %w", err)
	}
	domainState := domainData.GetInfo().IsolationGroupConfig
	globalCfg, err := z.globalIsolationGroupDrains.FetchDynamicConfig(ctx, persistence.GlobalIsolationGroupConfig)

	globalState, err := fromCfgStore(globalCfg)
	if err != nil {
		return nil, fmt.Errorf("could not resolve global drains in isolationGroup handler: %w", err)
	}

	ig := &isolationGroups{
		Global: globalState,
		Domain: domainState,
	}

	return ig, nil
}

func (z *defaultIsolationGroupStateHandler) ProvideDomainCache(domainCache cache.DomainCache) State {
	z.domainCache = domainCache
	return z
}

func (z *defaultIsolationGroupStateHandler) ProvideConfigStoreManager(cfgMgr persistence.ConfigStoreManager) State {
	z.globalIsolationGroupDrains = cfgMgr
	return z
}

func (z *defaultIsolationGroupStateHandler) checkIfChanged() {
	// todo (david.porter)
	// check new values against existing cached ones
	// get the difference
	// if any difference, notify subscribers for whom the change is applicable
	// ie, global changes for all, domain changes for the domain-listeners
	panic("not implemented")
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

func isDrained(isolationGroup string, global types.IsolationGroupConfiguration, domain types.IsolationGroupConfiguration) bool {
	globalCfg, hasGlobalConfig := global[string(isolationGroup)]
	domainCfg, hasDomainConfig := domain[string(isolationGroup)]
	if hasGlobalConfig {
		if globalCfg.State == types.IsolationGroupStateDrained {
			return true
		}
	}
	if hasDomainConfig {
		if domainCfg.State == types.IsolationGroupStateDrained {
			return true
		}
	}
	return false
}

// A simple explicit deny-based isolation group implementation
func availableIG(allIsolationGroups []string, global types.IsolationGroupConfiguration, domain types.IsolationGroupConfiguration) types.IsolationGroupConfiguration {
	out := types.IsolationGroupConfiguration{}
	for _, isolationGroup := range allIsolationGroups {
		globalCfg, hasGlobalConfig := global[string(isolationGroup)]
		domainCfg, hasDomainConfig := domain[string(isolationGroup)]
		if hasGlobalConfig {
			if globalCfg.State == types.IsolationGroupStateDrained {
				continue
			}
		}
		if hasDomainConfig {
			if domainCfg.State == types.IsolationGroupStateDrained {
				continue
			}
		}
		out[isolationGroup] = types.IsolationGroupPartition{
			Name:  isolationGroup,
			State: types.IsolationGroupStateHealthy,
		}
	}
	return out
}

func fromCfgStore(in *persistence.FetchDynamicConfigResponse) (types.IsolationGroupConfiguration, error) {
	panic("not implemented")
}

func updateGlobalIGRequest(in types.UpdateGlobalIsolationGroupsRequest) *persistence.UpdateDynamicConfigRequest {

	types.DynamicConfigBlob{
		SchemaVersion: 0,
		Entries:       nil,
	}

	return &persistence.UpdateDynamicConfigRequest{
		Snapshot: &persistence.DynamicConfigSnapshot{
			Version: 0,
			Values:  nil,
		},
	}
}

func getGlobalIGResponse(in persistence.FetchDynamicConfigResponse) *types.GetGlobalIsolationGroupsResponse {
	panic("not implement")
	return nil
}
