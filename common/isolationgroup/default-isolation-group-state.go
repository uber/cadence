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
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/uber/cadence/common/persistence"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/dynamicconfig/configstore"
	csc "github.com/uber/cadence/common/dynamicconfig/configstore/config"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
)

// defaultConfig values for the partitioning library for segmenting portions of workflows into isolation-groups - a resiliency
// concept meant to help move workflows around and away from failure zones.
type defaultConfig struct {
	// IsolationGroupEnabled is a domain-based configuration value for whether this feature is enabled at all
	IsolationGroupEnabled dynamicconfig.BoolPropertyFnWithDomainFilter
	// AllIsolationGroups is a static list of all the possible isolation group names
	AllIsolationGroups []string
}

type defaultIsolationGroupStateHandler struct {
	status                          int32
	done                            chan struct{}
	log                             log.Logger
	domainCache                     cache.DomainCache
	globalIsolationGroupDrains      dynamicconfig.Client
	globalIsolationGroupDrainClient dynamicconfig.Client
	config                          defaultConfig
	subscriptionMu                  sync.Mutex
	valuesMu                        sync.RWMutex
	lastSeen                        *isolationGroups
	updateCB                        func()
	// subscriptions is a map of domains->subscription-keys-> subscription channels
	// for notifying when there's a state change
	subscriptions map[string]map[string]chan<- ChangeEvent
}

const (
	DefaultIsolationGroupConfigStoreManagerGlobalMapping dynamicconfig.MapKey = iota
)

// NewDefaultIsolationGroupStateWatcher is the default constructor
func NewDefaultIsolationGroupStateWatcher(
	logger log.Logger,
	dc *dynamicconfig.Collection,
	persistenceCfg *config.Persistence,
	domainCache cache.DomainCache,
) (State, error) {
	stopChan := make(chan struct{})
	cscConfig := &csc.ClientConfig{
		PollInterval:        dc.GetDurationProperty(dynamicconfig.IsolationGroupStateRefreshInterval)(),
		UpdateRetryAttempts: dc.GetIntProperty(dynamicconfig.IsolationGroupStateUpdateRetryAttempts)(),
		FetchTimeout:        dc.GetDurationProperty(dynamicconfig.IsolationGroupStateFetchTimeout)(),
		UpdateTimeout:       dc.GetDurationProperty(dynamicconfig.IsolationGroupStateUpdateTimeout)(),
	}
	cfgStoreClient, err := configstore.NewConfigStoreClient(cscConfig, persistenceCfg, logger, stopChan, persistence.GlobalIsolationGroupConfig)
	if err != nil {
		return nil, fmt.Errorf("failure during setup for the IsolationGroupStateWatcher: %w", err)
	}
	return NewDefaultIsolationGroupStateWatcherWithConfigStoreClient(logger, dc, domainCache, cfgStoreClient, stopChan)
}

// NewDefaultIsolationGroupStateWatcherWithConfigStoreClient Is a constructor which allows passing in the dynamic config client
func NewDefaultIsolationGroupStateWatcherWithConfigStoreClient(
	logger log.Logger,
	dc *dynamicconfig.Collection,
	domainCache cache.DomainCache,
	cfgStoreClient dynamicconfig.Client,
	stopChan chan struct{},
) (State, error) {

	allIGs := dc.GetListProperty(dynamicconfig.AllIsolationGroups)()
	allIsolationGroups, err := mapAllIsolationGroupsResponse(allIGs)
	if err != nil {
		return nil, fmt.Errorf("could not get all isolation groups fron dynamic config: %w", err)
	}

	config := defaultConfig{
		IsolationGroupEnabled: dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableTasklistIsolation),
		AllIsolationGroups:    allIsolationGroups,
	}

	return &defaultIsolationGroupStateHandler{
		done:                       stopChan,
		domainCache:                domainCache,
		globalIsolationGroupDrains: cfgStoreClient,
		status:                     common.DaemonStatusInitialized,
		log:                        logger,
		config:                     config,
		subscriptionMu:             sync.Mutex{},
		subscriptions:              make(map[string]map[string]chan<- ChangeEvent),
	}, nil
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
	mappedInput, err := mapUpdateGlobalIsolationGroupsRequest(in.IsolationGroups)
	if err != nil {
		return err
	}
	return z.globalIsolationGroupDrains.UpdateValue(
		DefaultIsolationGroupConfigStoreManagerGlobalMapping,
		mappedInput,
	)
}

func (z *defaultIsolationGroupStateHandler) GetGlobalState(ctx context.Context) (*types.GetGlobalIsolationGroupsResponse, error) {
	res, err := z.globalIsolationGroupDrains.ListValue(DefaultIsolationGroupConfigStoreManagerGlobalMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to get global isolation groups from datastore: %w", err)
	}
	resp, err := mapDynamicConfigResponse(res)
	if err != nil {
		return nil, fmt.Errorf("failed to get global isolation groups from datastore: %w", err)
	}
	return &types.GetGlobalIsolationGroupsResponse{IsolationGroups: resp}, nil
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
	globalCfg, err := z.globalIsolationGroupDrains.ListValue(DefaultIsolationGroupConfigStoreManagerGlobalMapping)
	if err != nil {
		return nil, fmt.Errorf("could not resolve global drains in %w", err)
	}

	globalState, err := mapDynamicConfigResponse(globalCfg)
	if err != nil {
		return nil, fmt.Errorf("could not resolve global drains in isolationGroup handler: %w", err)
	}

	ig := &isolationGroups{
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

func isDrained(isolationGroup string, global types.IsolationGroupConfiguration, domain types.IsolationGroupConfiguration) bool {
	globalCfg, hasGlobalConfig := global[isolationGroup]
	domainCfg, hasDomainConfig := domain[isolationGroup]
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

// ----- Mappers -----
// dynamic config library does much of the decoding already, so this is asymmetric
// to the serialization step. Incoming value should be just
// map[string]interface{}
func mapDynamicConfigResponse(in []*types.DynamicConfigEntry) (out types.IsolationGroupConfiguration, err error) {
	if in == nil {
		return nil, nil
	}
	out = make(types.IsolationGroupConfiguration)

	for _, entry := range in {
		for _, v := range entry.Values {
			if v.Value == nil {
				continue
			}
			if v.Value.GetEncodingType() != types.EncodingTypeJSON {
				return nil, fmt.Errorf("failed to decode values: %v, (%T)", v.Value, v.Value.EncodingType)
			}
			var partition types.IsolationGroupPartition
			err := json.Unmarshal(v.Value.GetData(), &partition)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal entry: %v, got %w", string(v.Value.GetData()), err)
			}
			out[partition.Name] = partition
		}
	}
	return out, nil
}

func mapAllIsolationGroupsResponse(in []interface{}) ([]string, error) {
	var allIsolationGroups []string
	for k := range in {
		v, ok := in[k].(string)
		if !ok {
			return nil, fmt.Errorf("failed to get all-isolation-groups resonse from dynamic config: got %v (%T)", in[k], in[k])
		}
		allIsolationGroups = append(allIsolationGroups, v)
	}
	return allIsolationGroups, nil
}

func mapUpdateGlobalIsolationGroupsRequest(in types.IsolationGroupConfiguration) ([]*types.DynamicConfigValue, error) {
	var out []*types.DynamicConfigValue
	for _, v := range in {
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input for dynamic config: %w", err)
		}
		out = append(out, &types.DynamicConfigValue{
			Value: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonData,
			},
			Filters: nil,
		})
	}
	// ensure sort-order
	sort.Slice(out, func(i, j int) bool {
		return string(out[i].Value.GetData()) < string(out[j].Value.GetData())
	})
	return out, nil
}
