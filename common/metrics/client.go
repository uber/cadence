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

// ClientImpl is used for reporting metrics by various Cadence services
type ClientImpl struct {
	// parentReporter is the parent scope for the metrics
	parentScope tally.Scope
	childScopes map[int]tally.Scope
	metricDefs  map[int]metricDefinition
	serviceIdx  ServiceIdx
}

// NewClient creates and returns a new instance of
// Client implementation
// reporter holds the common tags for the service
// serviceIdx indicates the service type in (InputhostIndex, ... StorageIndex)
func NewClient(scope tally.Scope, serviceIdx ServiceIdx) Client {
	totalScopes := len(ScopeDefs[Common]) + len(ScopeDefs[serviceIdx])
	metricsClient := &ClientImpl{
		parentScope: scope,
		childScopes: make(map[int]tally.Scope, totalScopes),
		metricDefs:  getMetricDefs(serviceIdx),
		serviceIdx:  serviceIdx,
	}

	for idx, def := range ScopeDefs[Common] {
		scopeTags := map[string]string{
			OperationTagName: def.operation,
		}
		mergeMapToRight(def.tags, scopeTags)
		metricsClient.childScopes[idx] = scope.Tagged(scopeTags)
	}

	for idx, def := range ScopeDefs[serviceIdx] {
		scopeTags := map[string]string{
			OperationTagName: def.operation,
		}
		mergeMapToRight(def.tags, scopeTags)
		metricsClient.childScopes[idx] = scope.Tagged(scopeTags)
	}

	return metricsClient
}

// IncCounter increments one for a counter and emits
// to metrics backend
func (m *ClientImpl) IncCounter(scopeIdx int, counterIdx int) {
	name := string(m.metricDefs[counterIdx].metricName)
	m.childScopes[scopeIdx].Counter(name).Inc(1)
}

// AddCounter adds delta to the counter and
// emits to the metrics backend
func (m *ClientImpl) AddCounter(scopeIdx int, counterIdx int, delta int64) {
	name := string(m.metricDefs[counterIdx].metricName)
	m.childScopes[scopeIdx].Counter(name).Inc(delta)
}

// StartTimer starts a timer for the given
// metric name
func (m *ClientImpl) StartTimer(scopeIdx int, timerIdx int) tally.Stopwatch {
	name := string(m.metricDefs[timerIdx].metricName)
	return m.childScopes[scopeIdx].Timer(name).Start()
}

// RecordTimer record and emit a timer for the given
// metric name
func (m *ClientImpl) RecordTimer(scopeIdx int, timerIdx int, d time.Duration) {
	name := string(m.metricDefs[timerIdx].metricName)
	m.childScopes[scopeIdx].Timer(name).Record(d)
}

// RecordHistogramDuration record and emit a duration
func (m *ClientImpl) RecordHistogramDuration(scopeIdx int, timerIdx int, d time.Duration) {
	name := string(m.metricDefs[timerIdx].metricName)
	m.childScopes[scopeIdx].Histogram(name, m.getBuckets(timerIdx)).RecordDuration(d)
}

// UpdateGauge reports Gauge type metric
func (m *ClientImpl) UpdateGauge(scopeIdx int, gaugeIdx int, value float64) {
	name := string(m.metricDefs[gaugeIdx].metricName)
	m.childScopes[scopeIdx].Gauge(name).Update(value)
}

// Scope return a new internal metrics scope that can be used to add additional
// information to the metrics emitted
func (m *ClientImpl) Scope(scopeIdx int, tags ...Tag) Scope {
	scope := m.childScopes[scopeIdx]
	return newMetricsScope(scope, scope, m.metricDefs, false).Tagged(tags...)
}

func (m *ClientImpl) getBuckets(id int) tally.Buckets {
	if m.metricDefs[id].buckets != nil {
		return m.metricDefs[id].buckets
	}
	return tally.DefaultBuckets
}

func getMetricDefs(serviceIdx ServiceIdx) map[int]metricDefinition {
	defs := make(map[int]metricDefinition)
	for idx, def := range MetricDefs[Common] {
		defs[idx] = def
	}

	for idx, def := range MetricDefs[serviceIdx] {
		defs[idx] = def
	}

	return defs
}

func mergeMapToRight(src map[string]string, dest map[string]string) {
	for k, v := range src {
		dest[k] = v
	}
}
