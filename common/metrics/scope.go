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

package metrics

import (
	"strings"
	"time"

	"github.com/uber-go/tally"
)

type metricsScope struct {
	scope          tally.Scope
	defs           map[int]metricDefinition
	isDomainTagged bool
	cachedScopes   map[string]Scope
}

func newMetricsScope(scope tally.Scope, defs map[int]metricDefinition, isDomain bool) Scope {
	return &metricsScope{
		scope:          scope,
		defs:           defs,
		isDomainTagged: isDomain,
		cachedScopes:   make(map[string]Scope),
	}
}

// NoopScope returns a noop scope of metrics
func NoopScope(serviceIdx ServiceIdx) Scope {
	return &metricsScope{
		scope:          tally.NoopScope,
		defs:           getMetricDefs(serviceIdx),
		isDomainTagged: false,
		cachedScopes:   make(map[string]Scope),
	}
}

func (m *metricsScope) IncCounter(id int) {
	name := string(m.defs[id].metricName)
	m.scope.Counter(name).Inc(1)
}

func (m *metricsScope) AddCounter(id int, delta int64) {
	name := string(m.defs[id].metricName)
	m.scope.Counter(name).Inc(delta)
}

func (m *metricsScope) UpdateGauge(id int, value float64) {
	name := string(m.defs[id].metricName)
	m.scope.Gauge(name).Update(value)
}

func (m *metricsScope) StartTimer(id int) Stopwatch {
	name := string(m.defs[id].metricName)
	timer := m.scope.Timer(name)
	if m.isDomainTagged {
		timerAll := m.scope.Tagged(map[string]string{domain: domainAllValue}).Timer(name)
		return NewStopwatch(timer, timerAll)
	}
	return NewStopwatch(timer)
}

func (m *metricsScope) RecordTimer(id int, d time.Duration) {
	name := string(m.defs[id].metricName)
	m.scope.Timer(name).Record(d)
	if m.isDomainTagged {
		m.scope.Tagged(map[string]string{domain: domainAllValue}).Timer(name).Record(d)
	}
}

func (m *metricsScope) RecordHistogramDuration(id int, value time.Duration) {
	name := string(m.defs[id].metricName)
	m.scope.Histogram(name, m.getBuckets(id)).RecordDuration(value)
}

func (m *metricsScope) RecordHistogramValue(id int, value float64) {
	name := string(m.defs[id].metricName)
	m.scope.Histogram(name, m.getBuckets(id)).RecordValue(value)
}

func (m *metricsScope) Tagged(tags ...Tag) Scope {
	strTags := stringifyTags(tags...)
	scope, ok := m.cachedScopes[strTags]
	if ok {
		return scope
	}

	domainTagged := false
	tagMap := make(map[string]string, len(tags))
	for _, tag := range tags {
		if isDomainTagged(tag) {
			domainTagged = true
		}
		tagMap[tag.Key()] = tag.Value()
	}
	scope = newMetricsScope(m.scope.Tagged(tagMap), m.defs, domainTagged)
	m.cachedScopes[strTags] = scope
	return scope
}

func (m *metricsScope) getBuckets(id int) tally.Buckets {
	if m.defs[id].buckets != nil {
		return m.defs[id].buckets
	}
	return tally.DefaultBuckets
}

func isDomainTagged(tag Tag) bool {
	return tag.Key() == domain && tag.Value() != domainAllValue
}

func stringifyTags(tags ...Tag) string {
	var b strings.Builder
	for _, tag := range tags {
		b.WriteString(tag.Key())
		b.WriteString(tag.Value())
	}
	return b.String()
}
