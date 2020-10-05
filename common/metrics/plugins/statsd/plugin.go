// Copyright (c) 2020 Uber Technologies, Inc.
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

package statsd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/uber-go/tally"
	tallystatsdreporter "github.com/uber-go/tally/statsd"

	"github.com/uber/cadence/common/metrics/plugins"
	statsdreporter "github.com/uber/cadence/common/metrics/tally/statsd"
	"github.com/uber/cadence/common/service/config"
)

const (
	// PluginName is the name of the plugin
	PluginName = "statsd"
)

type plugin struct{}

var _ plugins.MetricReporterPlugin = (*plugin)(nil)

func init() {
	plugins.RegisterPlugin(PluginName, &plugin{})
}

func (p *plugin) NewTallyScope(cfg config.Metrics) (tally.Scope, error) {
	configKVs := cfg.ThirdParty.ConfigKVs
	hostPort := configKVs["hostPort"]
	if hostPort == "" {
		return nil, fmt.Errorf("hostPort is missing")
	}
	prefix := configKVs["prefix"]
	flushInterval, err := time.ParseDuration(configKVs["flushInterval"])
	if err != nil {
		return nil, err
	}
	flushBytes, err := strconv.Atoi(configKVs["flushBytes"])
	if err != nil {
		return nil, err
	}
	statter, err := statsd.NewBufferedClient(hostPort, prefix, flushInterval, flushBytes)
	if err != nil {
		return tally.NoopScope, fmt.Errorf("error creating statsd client %v", err)
	}

	//NOTE: according to ( https://github.com/uber-go/tally )Tally's statsd implementation doesn't support tagging.
	// Therefore, we implement Tally interface to have a statsd reporter that can support tagging
	reporter := statsdreporter.NewReporter(statter, tallystatsdreporter.Options{})
	scopeOpts := tally.ScopeOptions{
		Tags:     cfg.Tags,
		Reporter: reporter,
		Prefix:   cfg.Prefix,
	}
	scope, _ := tally.NewRootScope(scopeOpts, time.Second)
	return scope, nil
}
