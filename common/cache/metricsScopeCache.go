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
	"bytes"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common/metrics"
)

const flushBufferedMetricsScopeDuration = 10 * time.Second

type (
	metricsScopeMap map[string]metrics.Scope

	buffer struct {
		sync.RWMutex
		bufferMap metricsScopeMap
	}

	domainMetricsScopeCache struct {
		buffer   *buffer
		cache atomic.Value
		closeCh chan int
	}
)

// NewDomainMetricsScopeCache constructs a new domainMetricsScopeCache
func NewDomainMetricsScopeCache(flushDuration ...time.Duration) DomainMetricsScopeCache {

	mc := new(domainMetricsScopeCache)
	mc.buffer = &buffer{
		bufferMap: make(metricsScopeMap),
	}
	mc.closeCh = make(chan int)
	mc.cache.Store(make(metricsScopeMap))

	var refreshDuration time.Duration
	if len(flushDuration) == 0 {
		refreshDuration = flushBufferedMetricsScopeDuration
	} else {
		refreshDuration = flushDuration[0]
	}

	go mc.flushBufferedMetricsScope(refreshDuration)
	return mc
}

func (c *domainMetricsScopeCache) flushBufferedMetricsScope(flushDuration time.Duration) {
	for {
		select {
		case <-time.After(flushDuration):
			c.buffer.Lock()
			if len(c.buffer.bufferMap) > 0 {
				scopeMap := make(metricsScopeMap)

				data := c.cache.Load().(metricsScopeMap)
				// Copy everything over after atomic load
				for key, val := range data {
					scopeMap[key] = val
				}

				// Copy from buffered array
				for key, val := range c.buffer.bufferMap {
					scopeMap[key] = val
				}

				c.cache.Store(scopeMap)
				c.buffer.bufferMap = make(metricsScopeMap)
			}
			c.buffer.Unlock()

		case <- c.closeCh:
			break
		}
	}
}

// Get retrieves scope for domainID and scopeIdx
func (c *domainMetricsScopeCache) Get(domainID string, scopeIdx int) (metrics.Scope, bool) {
	key := joinStrings(domainID, "_", strconv.Itoa(scopeIdx))

	data := c.cache.Load().(metricsScopeMap)

	if data == nil {
		return nil, false
	}

	metricsScope, ok := data[key]

	return metricsScope, ok
}

// Put puts map of domainID and scopeIdx to metricsScope
func (c *domainMetricsScopeCache) Put(domainID string, scopeIdx int, scope metrics.Scope) {
	key := joinStrings(domainID, "_", strconv.Itoa(scopeIdx))

	c.buffer.Lock()
	defer c.buffer.Unlock()

	c.buffer.bufferMap[key] = scope
}

func (c *domainMetricsScopeCache) Stop() {
	c.closeCh <- 0
}

func joinStrings(str ...string) string {
	var buffer bytes.Buffer
	for _, s := range str {
		buffer.WriteString(s)
	}
	return buffer.String()
}