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
