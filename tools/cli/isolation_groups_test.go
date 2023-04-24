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

package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/types"
)

func TestValidateIsolationGroupArgs(t *testing.T) {

	tests := map[string]struct {
		domainArgs          string
		globalDomainArg     string
		setDrainsArgs       []string
		jsonConfigArgs      string
		removeAllDrainsArgs bool

		requiresDomain bool
		expectedErr    error
	}{
		"valid inputs for doing a drain": {
			domainArgs:      "some-domain",
			globalDomainArg: "",
			setDrainsArgs:   []string{"zone-1", "zone-2"},
			jsonConfigArgs:  "",

			expectedErr: nil,
		},
		"valid json input": {
			domainArgs:      "some-domain",
			globalDomainArg: "",
			setDrainsArgs:   nil,
			jsonConfigArgs:  "{}",

			expectedErr: nil,
		},
		"invalid - no domain": {
			domainArgs:      "",
			globalDomainArg: "",
			setDrainsArgs:   nil,
			jsonConfigArgs:  "{}",
			requiresDomain:  true,

			expectedErr: errors.New("the --domain flag is required"),
		},
		"invalid - global domain": {
			domainArgs:      "",
			globalDomainArg: "second domain",
			setDrainsArgs:   nil,
			jsonConfigArgs:  "{}",
			requiresDomain:  true,

			expectedErr: errors.New("the flag '--domain' has to go at the end"),
		},
		"invalid - no config domain": {
			domainArgs:      "domain",
			globalDomainArg: "",
			setDrainsArgs:   nil,
			jsonConfigArgs:  "",
			requiresDomain:  true,

			expectedErr: errors.New("need to specify either \"set-drains\", \"json\" or \"remove-all-drains\" flags"),
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expectedErr, validateIsolationGroupUpdateArgs(
				td.domainArgs,
				td.globalDomainArg,
				td.setDrainsArgs,
				td.jsonConfigArgs,
				td.removeAllDrainsArgs,
				td.requiresDomain))
		})
	}
}

func TestParseCliInput(t *testing.T) {

	tests := map[string]struct {
		setDrainsArgs  []string
		jsonConfigArgs string

		expected    *types.IsolationGroupConfiguration
		expectedErr error
	}{
		"valid inputs for doing a drain": {
			setDrainsArgs:  []string{"zone-1", "zone-2"},
			jsonConfigArgs: "",

			expected: &types.IsolationGroupConfiguration{
				"zone-1": {Name: "zone-1", State: types.IsolationGroupStateDrained},
				"zone-2": {Name: "zone-2", State: types.IsolationGroupStateDrained},
			},
		},
		"valid json input": {
			setDrainsArgs:  nil,
			jsonConfigArgs: "[{\"Name\": \"zone-1\", \"State\": 2}, {\"Name\": \"zone-2\", \"State\": 1}]",
			expected: &types.IsolationGroupConfiguration{
				"zone-1": {Name: "zone-1", State: types.IsolationGroupStateDrained},
				"zone-2": {Name: "zone-2", State: types.IsolationGroupStateHealthy},
			},

			expectedErr: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := parseIsolationGroupCliInputCfg(
				td.setDrainsArgs,
				td.jsonConfigArgs,
			)
			assert.Equal(t, td.expected, res)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}
