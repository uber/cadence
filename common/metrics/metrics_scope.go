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

import "github.com/uber-go/tally"

const domainTag = "domain"

// Tag is an interface to define metrics tags
type Tag interface {
	Key() string
	Value() string
}

// Scope is an interface for metrics
type Scope interface {
	Counter(id int) tally.Counter
	Gauge(id int) tally.Gauge
	Timer(id int) tally.Timer
	Tagged(tags ...Tag) Scope
}

type metricsScope struct {
	scope tally.Scope
	defs  map[int]metricDefinition
}

func newMetricsScope(scope tally.Scope, defs map[int]metricDefinition) Scope {
	return &metricsScope{scope, defs}
}

func (m *metricsScope) Counter(id int) tally.Counter {
	name := string(m.defs[id].metricName)
	return m.scope.Counter(name)
}

func (m *metricsScope) Gauge(id int) tally.Gauge {
	name := string(m.defs[id].metricName)
	return m.scope.Gauge(name)
}

func (m *metricsScope) Timer(id int) tally.Timer {
	name := string(m.defs[id].metricName)
	return m.scope.Timer(name)
}

func (m *metricsScope) Tagged(tags ...Tag) Scope {
	tagMap := make(map[string]string, len(tags))
	for _, tag := range tags {
		tagMap[tag.Key()] = tag.Value()
	}
	return newMetricsScope(m.scope.Tagged(tagMap), m.defs)
}

// DomainTag wraps a domain tag
type DomainTag struct {
	value string
}

// NewDomainTag returns a new domain tag
func NewDomainTag(value string) DomainTag {
	return DomainTag{value}
}

// Key returns the key of the domain tag
func (d DomainTag) Key() string {
	return domainTag
}

// Value returns the value of a domain tag
func (d DomainTag) Value() string {
	return d.value
}
