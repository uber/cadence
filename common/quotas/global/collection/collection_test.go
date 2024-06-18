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
	"go.uber.org/zap/zapcore"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/quotas/global/collection/internal"
	"github.com/uber/cadence/common/quotas/global/rpc"
)

func TestLifecycleBasics(t *testing.T) {
	defer goleak.VerifyNone(t) // must shut down cleanly

	ctrl := gomock.NewController(t)
	logger, logs := testlogger.NewObserved(t)
	c, err := New(
		"test",
		quotas.NewCollection(quotas.NewMockLimiterFactory(ctrl)),
		quotas.NewCollection(quotas.NewMockLimiterFactory(ctrl)),
		func(opts ...dynamicconfig.FilterOption) time.Duration { return time.Second },
		nil, // not used
		func(globalRatelimitKey string) string { return string(modeGlobal) },
		nil, // no rpc expected as there are no metrics to submit
		logger,
		metrics.NewNoopMetricsClient(),
	)
	require.NoError(t, err)
	mts := clock.NewMockedTimeSource()
	ts := mts.(clock.TimeSource)
	c.TestOverrides(t, &ts, nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, c.OnStart(ctx))

	mts.BlockUntil(1)        // created timer in background updater
	mts.Advance(time.Second) // trigger the update
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
				"test",
				quotas.NewCollection(localLimiters),
				quotas.NewCollection(globalLimiters),
				func(opts ...dynamicconfig.FilterOption) time.Duration { return time.Second },
				func(domain string) int { return 5 },
				func(globalRatelimitKey string) string { return string(test.mode) },
				nil,
				testlogger.New(t),
				metrics.NewNoopMetricsClient(),
			)
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

func TestCollectionSubmitsDataAndUpdates(t *testing.T) {
	defer goleak.VerifyNone(t) // must shut down cleanly

	ctrl := gomock.NewController(t)
	// anything non-zero
	limiters := quotas.NewSimpleDynamicRateLimiterFactory(func(domain string) int {
		return 1
	})
	logger, observed := testlogger.NewObserved(t)
	aggs := rpc.NewMockClient(ctrl)
	c, err := New(
		"test",
		quotas.NewCollection(limiters),
		quotas.NewCollection(limiters),
		func(opts ...dynamicconfig.FilterOption) time.Duration { return time.Second },
		func(domain string) int { return 10 }, // target 10 rps / one per 100ms
		func(globalRatelimitKey string) string { return string(modeGlobal) },
		aggs,
		logger,
		metrics.NewNoopMetricsClient(),
	)
	require.NoError(t, err)
	mts := clock.NewMockedTimeSource()
	ts := mts.(clock.TimeSource)
	c.TestOverrides(t, &ts, nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, c.OnStart(ctx))

	// generate some data
	someLimiter := c.For("something")
	res := someLimiter.Reserve()
	assert.True(t, res.Allow(), "first request should have been allowed")
	res.Used(true)
	assert.False(t, someLimiter.Allow(), "second request on the same domain should have been rejected")
	assert.NoError(t, c.For("other").Wait(ctx), "request on a different domain should be allowed")

	// prep for the calls
	called := make(chan struct{}, 1)
	aggs.EXPECT().Update(gomock.Any(), gomock.Any(), map[string]rpc.Calls{
		"test:other": {
			Allowed:  1,
			Rejected: 0,
		},
		"test:something": {
			Allowed:  1,
			Rejected: 1,
		},
	}).DoAndReturn(func(ctx context.Context, period time.Duration, load map[string]rpc.Calls) rpc.UpdateResult {
		called <- struct{}{}
		return rpc.UpdateResult{
			Weights: map[string]float64{
				"test:something": 1, // should recover a token in 100ms
				// "test:other":   // not returned, should not change weight
			},
			Err: nil,
		}
	})

	mts.BlockUntil(1)        // need to have created timer in background updater
	mts.Advance(time.Second) // trigger the update

	// wait until the calls occur
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("did not make an rpc call after 1s")
	}
	// panic if more calls occur
	close(called)

	// wait for the updates to be sent to the ratelimiters, and for "something"'s 100ms token to recover
	time.Sleep(150 * time.Millisecond)

	// and make sure updates occurred
	assert.False(t, c.For("other").Allow(), "should be no recovered tokens yet on the slow limit")
	assert.True(t, c.For("something").Allow(), "should have allowed one request on the fast limit")            // should use weight, not target rps
	assert.False(t, c.For("something").Allow(), "should not have allowed as second request on the fast limit") // should use weight, not target rps

	assert.NoError(t, c.OnStop(ctx))

	// and make sure no non-fatal errors were logged
	for _, log := range observed.All() {
		t.Logf("observed log: %v", log.Message)
		assert.Lessf(t, log.Level, zapcore.ErrorLevel, "should not be any error logs, got: %v", log.Message)
	}
}
