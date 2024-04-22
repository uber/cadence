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

package peerprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/membership"
)

func TestRegisterAllowsPluginOnlyOnce(t *testing.T) {
	assert.NoError(t, Register("testConfig", func(cfg *config.YamlNode, container Container) (membership.PeerProvider, error) { return nil, nil }))
	assert.Error(t, Register("testConfig",
		func(cfg *config.YamlNode, container Container) (membership.PeerProvider, error) { return nil, nil }),
		"plugin can be registered only once",
	)

}
func TestProviderRetrunsErrorWhenNoProviderRegistered(t *testing.T) {
	a := Provider{
		config:    nil,
		container: Container{},
	}
	p, err := a.Provider()
	assert.Nil(t, p)
	assert.EqualError(t, err, "no configured peer providers found")
}

func TestProviderRetrunsErrorWhenNoConfigFound(t *testing.T) {
	err := Register("providerName", func(cfg *config.YamlNode, container Container) (membership.PeerProvider, error) {
		return nil, nil
	})
	assert.NoError(t, err)
	ppConfig := config.PeerProvider{
		"configKey": &config.YamlNode{},
	}
	p, err := New(ppConfig, Container{}).Provider()
	assert.Nil(t, p)
	assert.Error(t, err)
}
