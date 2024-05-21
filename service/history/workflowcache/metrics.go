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

package workflowcache

import (
	"sync"
	"time"

	"github.com/uber/cadence/common/metrics"
)

type workflowIDCountMetric struct {
	secondStart time.Time
	count       int
}

func (w *workflowIDCountMetric) reset(now time.Time) {
	w.secondStart = now
	w.count = 0
}

type domainMaxCount struct {
	mu          sync.Mutex
	maxCount    int
	secondStart time.Time
}

func (d *domainMaxCount) reset(now time.Time) {
	d.maxCount = 0
	d.secondStart = now
}

func (c *wfCache) getOrCreateDomain(domainName string, now time.Time) *domainMaxCount {
	newDomainCount := &domainMaxCount{
		mu:          sync.Mutex{},
		maxCount:    0,
		secondStart: now,
	}

	domainCount, _ := c.domainMaxCount.LoadOrStore(domainName, newDomainCount)

	return domainCount.(*domainMaxCount)
}

func (c *wfCache) updatePerDomainMaxWFRequestCount(domainName string, value *cacheValue) {
	now := c.timeSource.Now()
	domain := c.getOrCreateDomain(domainName, now)

	// It's enough to lock the domain, as the workflowIDCountMetric belongs to the domain
	domain.mu.Lock()
	defer domain.mu.Unlock()

	if c.timeSource.Since(domain.secondStart) > time.Second {
		// Emit the max count for the previous second and reset the count
		c.metricsClient.Scope(metrics.HistoryClientWfIDCacheScope, metrics.DomainTag(domainName)).
			UpdateGauge(metrics.WorkflowIDCacheRequestsExternalMaxRequestsPerSecondsGauge, float64(domain.maxCount))
		domain.reset(now)
	}

	if c.timeSource.Since(value.countMetric.secondStart) > time.Second {
		value.countMetric.reset(now)
	}

	value.countMetric.count++

	if value.countMetric.count > domain.maxCount {
		domain.maxCount = value.countMetric.count
	}
}
