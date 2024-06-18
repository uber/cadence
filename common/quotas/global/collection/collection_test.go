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

package collection

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/quotas/global/collection/internal"
)

func TestLifecycleBasics(t *testing.T) {
	defer goleak.VerifyNone(t) // must shut down cleanly

	ctrl := gomock.NewController(t)
	logger, logs := testlogger.NewObserved(t)
	c, err := New(
		quotas.NewCollection(quotas.NewMockLimiterFactory(ctrl)),
		quotas.NewCollection(quotas.NewMockLimiterFactory(ctrl)),
		func(opts ...dynamicconfig.FilterOption) time.Duration {
			return time.Second
		},
		func(globalRatelimitKey string) string {
			return string(modeGlobal)
		},
		logger, metrics.NewNoopMetricsClient())
	require.NoError(t, err)
	ts := clock.NewMockedTimeSource()
	c.TestOverrides(t, ts)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, c.OnStart(ctx))

	ts.BlockUntil(1)        // created timer in background updater
	ts.Advance(time.Second) // trigger the update
	// clockwork does not provide a way to wait for "a ticker or timer event was consumed",
	// so just sleep for a bit to allow that branch of the select statement to be followed.
	//
	// it seems *possible* for clockwork to provide this, maybe we should build it?
	time.Sleep(time.Millisecond)

	require.NoError(t, c.OnStop(ctx))

	// will fail if the tick did not begin processing, e.g. no tick or it raced with shutdown.
	// without a brief sleep after advancing time, this extremely rarely fails (try 10k tests to see)
	assert.Len(t, logs.FilterMessage("update tick").All(), 1, "no update logs, did it process the tick? all: %#v", logs.All())
}

func TestCollectionLimitersCollectMetrics(t *testing.T) {
	tests := map[string]struct {
		mode   keyMode
		local  internal.UsageMetrics
		global internal.UsageMetrics
	}{
		"disabled": {
			mode: modeDisabled,
		},
		"local": {
			mode:  modeLocal,
			local: internal.UsageMetrics{Allowed: 1},
		},
		"global": {
			mode:   modeGlobal,
			global: internal.UsageMetrics{Allowed: 1},
		},
		"local-shadow": {
			mode:   modeLocalShadowGlobal,
			local:  internal.UsageMetrics{Allowed: 1},
			global: internal.UsageMetrics{Allowed: 1},
		},
		"global-shadow": {
			mode:   modeGlobalShadowLocal,
			local:  internal.UsageMetrics{Allowed: 1},
			global: internal.UsageMetrics{Allowed: 1},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// anything non-zero
			localLimiters := quotas.NewSimpleDynamicRateLimiterFactory(func(domain string) int {
				return 1
			})
			globalLimiters := quotas.NewSimpleDynamicRateLimiterFactory(func(domain string) int {
				return 10
			})

			c, err := New(
				quotas.NewCollection(localLimiters),
				quotas.NewCollection(globalLimiters),
				func(opts ...dynamicconfig.FilterOption) time.Duration {
					return time.Second
				},
				func(globalRatelimitKey string) string {
					return string(test.mode)
				},
				testlogger.New(t), metrics.NewNoopMetricsClient())
			require.NoError(t, err)
			// not starting, in principle it could Collect() in the background and break this test.

			// perform one call
			_ = c.For("test").Allow()

			// check counts on limits
			globalEvents := make(map[string]internal.UsageMetrics)
			c.global.Range(func(k string, v *internal.FallbackLimiter) bool {
				usage, _, _ := v.Collect()
				_, ok := globalEvents[k]
				require.False(t, ok, "data key should not already exist: %v", k)
				globalEvents[k] = internal.UsageMetrics{
					Allowed:  usage.Allowed,
					Rejected: usage.Rejected,
				}
				return true
			})
			localEvents := make(map[string]internal.UsageMetrics)
			c.local.Range(func(k string, v internal.CountedLimiter) bool {
				_, ok := localEvents[k]
				require.False(t, ok, "data key should not already exist: %v", k)
				localEvents[k] = v.Collect()
				return true
			})

			if reflect.ValueOf(test.global).IsZero() {
				assert.Empty(t, globalEvents, "unexpected global events, should be none")
			} else {
				assert.Equal(t, map[string]internal.UsageMetrics{"test": test.global}, globalEvents, "incorrect global metrics")
			}
			if reflect.ValueOf(test.local).IsZero() {
				assert.Empty(t, localEvents, "unexpected local events, should be none")
			} else {
				assert.Equal(t, map[string]internal.UsageMetrics{"test": test.local}, localEvents, "incorrect local metrics")
			}
		})
	}
}
