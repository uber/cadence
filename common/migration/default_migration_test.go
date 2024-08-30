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

package migration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/testing/testdatagen"
	"github.com/uber/cadence/common/types"
)

func TestDefaultComparisonFn(t *testing.T) {
	tests := map[string]struct {
		active        interface{}
		activeErr     error
		background    interface{}
		backgroundErr error
		shouldBeEqual bool
	}{
		"Simple equality": {
			active:        struct{ A string }{A: "test123"},
			background:    struct{ A string }{A: "test123"},
			shouldBeEqual: true,
		},
		"Simple equality 1": {
			active:        struct{ A string }{A: "test123"},
			background:    struct{ A string }{A: "test123"},
			shouldBeEqual: true,
		},
		"equality - same values, though one is a pointer": {
			active:        common.Ptr(struct{ A string }{A: "test123"}),
			background:    struct{ A string }{A: "test123"},
			shouldBeEqual: false, // this is arguable either way, but it's more surprising to have them reported as equal
		},
		"Simple difference": {
			active:        struct{ A string }{A: "test123"},
			background:    struct{ A string }{A: "difference"},
			shouldBeEqual: false,
		},
		"Simple difference 1": {
			active:        common.Ptr(struct{ A string }{A: "test123"}),
			background:    common.Ptr(struct{ A string }{A: "difference"}),
			shouldBeEqual: false,
		},
		"Simple error handling 1": {
			active:        struct{ A string }{A: "a value"},
			background:    nil,
			backgroundErr: assert.AnError,
			shouldBeEqual: false,
		},
		"Simple error handling 2": {
			activeErr:     assert.AnError,
			background:    struct{ A string }{A: "a value"},
			shouldBeEqual: false,
		},
		"Simple error handling 3": {
			activeErr:     assert.AnError,
			backgroundErr: assert.AnError,
			shouldBeEqual: true,
		},
		"errors differ": {
			activeErr:     assert.AnError,
			backgroundErr: errors.New("a different error"),
			shouldBeEqual: false,
		},
		"types differ 1": {
			active:        123,
			background:    struct{ A string }{A: "a value"},
			shouldBeEqual: false,
		},
		"unexpected types handling": {
			active:        struct{ a string }{a: "123"},
			background:    struct{ a string }{a: "different"},
			shouldBeEqual: false,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			scope := tally.NewTestScope("", nil)
			metricsClient := metrics.NewClient(scope, 0)
			res := defaultComparisonFn[interface{}](loggerimpl.NewNopLogger(), metricsClient.Scope(123), "test", &td.active, td.activeErr, &td.background, td.backgroundErr)
			assert.Equal(t, td.shouldBeEqual, res)
		})
	}
}

func TestFuzzTestComparisonForMatchingTypes(t *testing.T) {

	gen := testdatagen.New(t)
	e := types.HistoryEvent{}
	gen.Fuzz(&e)

	scope := tally.NewTestScope("", nil)
	metricsClient := metrics.NewClient(scope, 0)

	equal := defaultComparisonFn[*types.HistoryEvent](
		loggerimpl.NewNopLogger(),
		metricsClient.Scope(123),
		"test",
		&e,
		nil,
		&e,
		nil,
	)
	assert.True(t, equal)
}
