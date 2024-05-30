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
	sync.Mutex

	startingSecond time.Time
	count          int
}

func (w *workflowIDCountMetric) reset(now time.Time) {
	w.startingSecond = now
	w.count = 0
}

func (c *wfCache) updatePerDomainMaxWFRequestCount(domainName string, value *cacheValue) {
	value.countMetric.Lock()
	defer value.countMetric.Unlock()

	if c.timeSource.Since(value.countMetric.startingSecond) > time.Second {
		value.countMetric.reset(c.timeSource.Now().UTC())
	}
	value.countMetric.count++

	// We can just use the upper of the metric, so it is not an issue to emit all the counts
	c.metricsClient.Scope(metrics.HistoryClientWfIDCacheScope, metrics.DomainTag(domainName)).
		RecordTimer(metrics.WorkflowIDCacheRequestsExternalMaxRequestsPerSecondsTimer, time.Duration(value.countMetric.count))
}
