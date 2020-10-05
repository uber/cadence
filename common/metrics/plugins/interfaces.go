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

package plugins

import (
	"fmt"

	"github.com/uber-go/tally"

	"github.com/uber/cadence/common/service/config"
)

type (
	MetricReporterPlugin interface {
		NewTallyScope(cfg config.Metrics) (tally.Scope, error)
	}
)

var supportedPlugins = map[string]MetricReporterPlugin{}

// RegisterPlugin will register a metric plugin
func RegisterPlugin(pluginName string, plugin MetricReporterPlugin) {
	if _, ok := supportedPlugins[pluginName]; ok {
		panic("plugin " + pluginName + " already registered")
	}
	supportedPlugins[pluginName] = plugin
}

func NewThirdPartyReporter(cfg config.Metrics) (tally.Scope, error) {
	plugin, ok := supportedPlugins[cfg.ThirdParty.PluginName]

	if !ok {
		return nil, fmt.Errorf("not supported plugin %v, only supported: %v", cfg.ThirdParty.PluginName, supportedPlugins)
	}

	return plugin.NewTallyScope(cfg)
}
