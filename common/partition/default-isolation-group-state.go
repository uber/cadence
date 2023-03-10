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

package partition

import (
	"context"
	"fmt"

	"github.com/uber/cadence/common/persistence"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
)

type DefaultIsolationGroupStateHandler struct {
	log                        log.Logger
	domainCache                cache.DomainCache
	globalIsolationGroupDrains persistence.GlobalIsolationGroupDrains
	config                     Config
}

func NewDefaultIsolationGroupStateWatcher(
	logger log.Logger,
	domainCache cache.DomainCache,
	config Config,
	globalIsolationGroupDrains persistence.GlobalIsolationGroupDrains,
) IsolationGroupState {
	return &DefaultIsolationGroupStateHandler{
		log:                        logger,
		config:                     config,
		domainCache:                domainCache,
		globalIsolationGroupDrains: globalIsolationGroupDrains,
	}
}

func (z *DefaultIsolationGroupStateHandler) GetByDomainID(ctx context.Context, domainID string) (*State, error) {
	domain, err := z.domainCache.GetDomainByID(domainID)
	if err != nil {
		return nil, fmt.Errorf("could not resolve domain in isolationGroup handler: %w", err)
	}
	return z.Get(ctx, domain.GetInfo().Name)
}

// Get the statue of a isolationGroup, with respect to both domain and global drains. Domain-specific drains override global config
// will return nil, nil when it is not enabled
func (z *DefaultIsolationGroupStateHandler) Get(ctx context.Context, domain string) (*State, error) {
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
	return &State{
		Global: globalState,
		Domain: domainState,
	}, nil
}
