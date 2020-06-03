// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
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

package cache

import (
	"strconv"
	"sync"

	"github.com/uber/cadence/common/metrics"
)

type metricsScopeCache struct {
	sync.RWMutex
	scopeMap map[string]metrics.Scope
}

// NewMetricsScopeCache constructs a new metricsScopeCache
func NewMetricsScopeCache() MetricsScopeCache {
	return &metricsScopeCache{
		scopeMap: make(map[string]metrics.Scope),
	}
}

// Get retrieves scope using domainID and scopeIdx from scopeMap
func (mc *metricsScopeCache) Get(domainID string, scopeIdx int) (metrics.Scope, bool) {
	mc.RLock()
	defer mc.RUnlock()

	key := domainID + "_" + strconv.Itoa(scopeIdx)

	metricsScope, found := mc.scopeMap[key]
	return metricsScope, found
}

// Put puts map of domainID and scopeIdx in the metricsCache scopeMap
func (mc *metricsScopeCache) Put(domainID string, scopeIdx int, scope metrics.Scope) {
	mc.Lock()
	defer mc.Unlock()

	key := domainID + "_" + strconv.Itoa(scopeIdx)
	mc.scopeMap[key] = scope
}
