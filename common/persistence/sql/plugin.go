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

package sql

import (
	"fmt"
	"sort"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
)

var supportedPlugins = map[string]sqlplugin.Plugin{}

// RegisterPlugin will register a SQL plugin
func RegisterPlugin(pluginName string, plugin sqlplugin.Plugin) {
	if _, ok := supportedPlugins[pluginName]; ok {
		panic("plugin " + pluginName + " already registered")
	}
	supportedPlugins[pluginName] = plugin
}

// RegisterPluginIfNotExists will register a SQL plugin only if a plugin with same name has not already been registered
func RegisterPluginIfNotExists(pluginName string, plugin sqlplugin.Plugin) {
	if _, ok := supportedPlugins[pluginName]; !ok {
		supportedPlugins[pluginName] = plugin
	}
}

// PluginRegistered returns true if plugin with given name has been registered, false otherwise
func PluginRegistered(pluginName string) bool {
	_, ok := supportedPlugins[pluginName]
	return ok
}

// GetRegisteredPluginNames returns the list of registered plugin names
func GetRegisteredPluginNames() []string {
	var plugins []string
	for k := range supportedPlugins {
		plugins = append(plugins, k)
	}
	sort.Strings(plugins)
	return plugins
}

// NewSQLDB creates a returns a reference to a logical connection to the
// underlying SQL database. The returned object is to tied to a single
// SQL database and the object can be used to perform CRUD operations on
// the tables in the database
func NewSQLDB(cfg *config.SQL) (sqlplugin.DB, error) {
	plugin, ok := supportedPlugins[cfg.PluginName]

	if !ok {
		return nil, fmt.Errorf("not supported plugin %v, only supported: %v", cfg.PluginName, supportedPlugins)
	}

	return plugin.CreateDB(cfg)
}

// NewSQLAdminDB returns a AdminDB
func NewSQLAdminDB(cfg *config.SQL) (sqlplugin.AdminDB, error) {
	plugin, ok := supportedPlugins[cfg.PluginName]

	if !ok {
		return nil, fmt.Errorf("not supported plugin %v, only supported: %v", cfg.PluginName, supportedPlugins)
	}

	return plugin.CreateAdminDB(cfg)
}
