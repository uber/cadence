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
	"time"

	"github.com/uber-go/tally"
)

type metricsScope struct {
	scope tally.Scope
	// cached scope with tags = { domain:all, tasklist:all }.
	// this is non-nil when isTaskListTagged is true for this scope.
	// when emitting metrics with tasklist tag, it is often necessary
	// to also visualize the same metrics across all domains / all tasklists
	// To facilitate that, this scope will dual emit metrics automatically when
	// with special tags = { domain: all, tasklist: all} when taskListTagged is true
	taskListAllScope tally.Scope
	defs             map[int]metricDefinition
	isDomainTagged   bool
	isTaskListTagged bool
}

func newMetricsScope(scope tally.Scope, defs map[int]metricDefinition, isDomain bool, isTaskList bool) Scope {
	mScope := &metricsScope{
		scope:            scope,
		defs:             defs,
		isDomainTagged:   isDomain,
		isTaskListTagged: isTaskList,
	}
	if mScope.isDomainTagged && mScope.isTaskListTagged {
		// pre-allocate { domain:all, tasklist:all } scope to save cost on each call
		mScope.taskListAllScope = scope.Tagged(map[string]string{domain: domainAllValue, taskList: taskListAllValue})
	}
	return mScope
}

// NoopScope returns a noop scope of metrics
func NoopScope(serviceIdx ServiceIdx) Scope {
	return &metricsScope{
		scope: tally.NoopScope,
		defs:  getMetricDefs(serviceIdx),
	}
}

func (m *metricsScope) IncCounter(id int) {
	m.AddCounter(id, 1)
}

func (m *metricsScope) AddCounter(id int, delta int64) {
	name := string(m.defs[id].metricName)
	m.scope.Counter(name).Inc(delta)
	if m.isTaskListTagged {
		m.taskListAllScope.Counter(name).Inc(delta)
	}
}

func (m *metricsScope) UpdateGauge(id int, value float64) {
	name := string(m.defs[id].metricName)
	m.scope.Gauge(name).Update(value)
}

func (m *metricsScope) StartTimer(id int) Stopwatch {
	name := string(m.defs[id].metricName)
	timer := m.scope.Timer(name)
	switch {
	case m.isTaskListTagged:
		return NewStopwatch(timer, m.taskListAllScope.Timer(name))
	case m.isDomainTagged:
		timerAll := m.scope.Tagged(map[string]string{domain: domainAllValue}).Timer(name)
		return NewStopwatch(timer, timerAll)
	default:
		return NewStopwatch(timer)
	}
}

func (m *metricsScope) RecordTimer(id int, d time.Duration) {
	name := string(m.defs[id].metricName)
	m.scope.Timer(name).Record(d)
	switch {
	case m.isTaskListTagged:
		m.taskListAllScope.Timer(name).Record(d)
	case m.isDomainTagged:
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
	domainTagged := false
	taskListTagged := false
	tagMap := make(map[string]string, len(tags))
	for _, tag := range tags {
		switch {
		case isDomainTagged(tag):
			domainTagged = true
		case isTaskListTagged(tag):
			taskListTagged = true
		}
		tagMap[tag.Key()] = tag.Value()
	}
	return newMetricsScope(m.scope.Tagged(tagMap), m.defs, domainTagged, taskListTagged)
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

func isTaskListTagged(tag Tag) bool {
	return tag.Key() == taskList && tag.Value() != taskListAllValue
}
