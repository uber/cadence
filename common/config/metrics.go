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

package config

import (
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	prom "github.com/m3db/prometheus_client_golang/prometheus"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	tallystatsdreporter "github.com/uber-go/tally/statsd"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	mprom "github.com/uber/cadence/common/metrics/tally/prometheus"
	statsdreporter "github.com/uber/cadence/common/metrics/tally/statsd"
)

const (
	_defaultReportingInterval = time.Second
)

// tally sanitizer options that satisfy both Prometheus and M3 restrictions.
// This will rename metrics at the tally emission level, so metrics name we
// use maybe different from what gets emitted. In the current implementation
// it will replace - and . with _
// We should still ensure that the base metrics are prometheus compatible,
// but this is necessary as the same prom client initialization is used by
// our system workflows.
var (
	safeCharacters = []rune{'_'}

	sanitizeOptions = tally.SanitizeOptions{
		NameCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		KeyCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ValueCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ReplacementCharacter: tally.DefaultReplacementCharacter,
	}
)

// NewScope builds a new tally scope for this metrics configuration
// Only one reporter type is allowed
func (c *Metrics) NewScope(logger log.Logger, service string) tally.Scope {
	if c.ReportingInterval <= 0 {
		c.ReportingInterval = _defaultReportingInterval
	}
	rootScope := tally.NoopScope
	if c.M3 != nil {
		rootScope = c.newM3Scope(logger)
	}
	if c.Statsd != nil {
		if rootScope != tally.NoopScope {
			logger.Fatal("error creating metric reporter: cannot have more than one types of metric configuration")
		}
		rootScope = c.newStatsdScope(logger)
	}
	if c.Prometheus != nil {
		if rootScope != tally.NoopScope {
			logger.Fatal("error creating metric reporter: cannot have more than one types of metric configuration")
		}
		rootScope = c.newPrometheusScope(logger)
	}
	rootScope = rootScope.Tagged(map[string]string{metrics.CadenceServiceTagName: service})
	return rootScope
}

// newM3Scope returns a new m3 scope with
// a default reporting interval of a second
func (c *Metrics) newM3Scope(logger log.Logger) tally.Scope {
	reporter, err := c.M3.NewReporter()
	if err != nil {
		logger.Fatal("error creating m3 reporter", tag.Error(err))
	}
	scopeOpts := tally.ScopeOptions{
		Tags:           c.Tags,
		CachedReporter: reporter,
		Prefix:         c.Prefix,
	}
	scope, _ := tally.NewRootScope(scopeOpts, c.ReportingInterval)
	return scope
}

// newM3Scope returns a new statsd scope with
// a default reporting interval of a second
func (c *Metrics) newStatsdScope(logger log.Logger) tally.Scope {
	config := c.Statsd
	if len(config.HostPort) == 0 {
		return tally.NoopScope
	}
	statter, err := statsd.NewClientWithConfig(&statsd.ClientConfig{
		Address:       config.HostPort,
		Prefix:        config.Prefix,
		UseBuffered:   true,
		FlushInterval: config.FlushInterval,
		FlushBytes:    config.FlushBytes,
	})
	if err != nil {
		logger.Fatal("error creating statsd client", tag.Error(err))
	}
	// NOTE: according to ( https://github.com/uber-go/tally )Tally's statsd implementation doesn't support tagging.
	// Therefore, we implement Tally interface to have a statsd reporter that can support tagging
	reporter := statsdreporter.NewReporter(statter, tallystatsdreporter.Options{})
	scopeOpts := tally.ScopeOptions{
		Tags:     c.Tags,
		Reporter: reporter,
		Prefix:   c.Prefix,
	}
	scope, _ := tally.NewRootScope(scopeOpts, c.ReportingInterval)
	return scope
}

// newPrometheusScope returns a new prometheus scope with
// a default reporting interval of a second
func (c *Metrics) newPrometheusScope(logger log.Logger) tally.Scope {
	if len(c.Prometheus.DefaultHistogramBuckets) == 0 {
		c.Prometheus.DefaultHistogramBuckets = mprom.DefaultHistogramBuckets()
	}
	reporter, err := c.Prometheus.NewReporter(
		prometheus.ConfigurationOptions{
			Registry: prom.NewRegistry(),
			OnError: func(err error) {
				logger.Warn("error in prometheus reporter", tag.Error(err))
			},
		},
	)
	if err != nil {
		logger.Fatal("error creating prometheus reporter", tag.Error(err))
	}
	scopeOpts := tally.ScopeOptions{
		Tags:            c.Tags,
		CachedReporter:  reporter,
		Separator:       prometheus.DefaultSeparator,
		SanitizeOptions: &sanitizeOptions,
		Prefix:          c.Prefix,
	}
	scope, _ := tally.NewRootScope(scopeOpts, c.ReportingInterval)
	return scope
}
