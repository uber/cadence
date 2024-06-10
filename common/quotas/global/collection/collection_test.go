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
	c := New(quotas.NewMockLimiterFactory(ctrl), func(opts ...dynamicconfig.FilterOption) time.Duration {
		return time.Second
	}, logger, metrics.NewNoopMetricsClient())
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
	fallbackLimiters := quotas.NewSimpleDynamicRateLimiterFactory(func(domain string) int {
		return 1
	})
	c := New(fallbackLimiters, func(opts ...dynamicconfig.FilterOption) time.Duration {
		return time.Second
	}, testlogger.New(t), metrics.NewNoopMetricsClient())
	// not starting as it's not currently necessary.  it should not affect this test.

	_ = c.For("test").Allow()
	// make sure we have some keys and a call was counted.
	type data struct {
		allowed, rejected int
		fallback          bool
	}
	events := make(map[string]data)
	c.limits.Range(func(k string, v *internal.FallbackLimiter) bool {
		_, _, actual, usingFallback := v.Collect()
		events[k] = data{
			allowed:  actual.Allowed,
			rejected: actual.Rejected,
			fallback: usingFallback,
		}
		return true
	})

	assert.Equalf(t, map[string]data{
		"test": data{
			allowed:  1,
			rejected: 0,
			fallback: true, // fallback is always used because there have been no updates
		},
	}, events, "should have observed one allowed request on the fallback")
}
