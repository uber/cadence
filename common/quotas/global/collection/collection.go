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

/*
Package collection contains the limiting-host ratelimit usage tracking and enforcing logic,
which acts as a quotas.Collection.

At a very high level, this wraps a [quotas.Limiter] to do a few additional things
in the context of the [github.com/uber/cadence/common/quotas/global] ratelimiter system:
  - keep track of usage per key (quotas.Limiter does not support this natively, nor should it)
  - periodically report usage to each key's "aggregator" host (batched and fanned out in parallel)
  - apply the aggregator's returned per-key RPS limits to future requests
  - fall back to the wrapped limiter in case of failures (handled internally in internal.FallbackLimiter)
*/
package collection

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/quotas/global/collection/internal"
)

type (
	Collection struct {
		updateInterval dynamicconfig.DurationPropertyFn

		logger log.Logger
		scope  metrics.Scope

		limits *internal.AtomicMap[string, *internal.FallbackLimiter]

		ctx       context.Context // context used for background operations, canceled when stopping
		ctxCancel func()
		stopped   chan struct{} // closed when stopping is complete

		keyModes dynamicconfig.StringPropertyWithRatelimitKeyFilter

		// now exists largely for tests, elsewhere it is always time.Now
		timesource clock.TimeSource
	}

	// Limiter is a simplified quotas.Limiter API, covering the only global-ratelimiter-used API: Allow.
	//
	// This is largely because Reserve() and Wait() are difficult to monitor for usage-tracking purposes.
	// If converted to the wrapper, Reserve can probably be added pretty easily.
	Limiter interface {
		Allow() bool
	}

	// calls is a small temporary pair[T] container
	calls struct {
		allowed, rejected int
	}
)

func New(
	fallback quotas.LimiterFactory,
	updateInterval dynamicconfig.DurationPropertyFn,
	keyModes dynamicconfig.StringPropertyWithRatelimitKeyFilter,
	logger log.Logger,
	met metrics.Client,
) *Collection {
	contents := internal.NewAtomicMap(func(key string) *internal.FallbackLimiter {
		return internal.NewFallbackLimiter(fallback.GetLimiter(key))
	})
	ctx, cancel := context.WithCancel(context.Background())
	c := &Collection{
		limits:         contents,
		updateInterval: updateInterval,

		logger:   logger.WithTags(tag.ComponentGlobalRatelimiter),
		scope:    met.Scope(metrics.FrontendGlobalRatelimiter),
		keyModes: keyModes,

		ctx:       ctx,
		ctxCancel: cancel,
		stopped:   make(chan struct{}),

		// override externally in tests
		timesource: clock.NewRealTimeSource(),
	}
	return c
}

func (c *Collection) TestOverrides(t *testing.T, timesource clock.TimeSource) {
	t.Helper()
	c.timesource = timesource
}

// OnStart follows fx's OnStart hook semantics.
func (c *Collection) OnStart(ctx context.Context) error {
	go func() {
		defer func() {
			// bad but not worth crashing the process
			log.CapturePanic(recover(), c.logger, nil)
		}()
		defer func() {
			close(c.stopped)
		}()
		c.backgroundUpdateLoop()
	}()

	return nil
}

// OnStop follows fx's OnStop hook semantics.
func (c *Collection) OnStop(ctx context.Context) error {
	c.ctxCancel()
	// context timeout is for everything to shut down, but this one should be fast.
	// don't wait very long.
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		return fmt.Errorf("ollection failed to stop, context canceled: %w", ctx.Err())
	case <-c.stopped:
		return nil
	}
}

func (c *Collection) For(key string) Limiter {
	TODO: need the key mapper to continue.  blaaah.  so much stuff to do at the same time.
	right here I should be leaning on the fallback collection (not limiter) if it is not shadowing.
	that way fallback truly is "fallback" and does not even trigger dual-collection.
		so like 5 modes?
		- disable (use fallback collection, ignore everything)
		- fallback (use this code, but do not submit anything)
		- fallback-shadow (use this code, submit everything, but rely on fallback for behavior)
		- global, global-shadow
	return c.limits.Load(key)
}

func (c *Collection) backgroundUpdateLoop() {
	tickInterval := c.updateInterval()

	c.logger.Debug("update loop starting")

	ticker := c.timesource.NewTicker(tickInterval)
	defer ticker.Stop()

	lastGatherTime := c.timesource.Now()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.Chan():
			now := c.timesource.Now()
			c.logger.Debug("update tick")
			// update  if it changed
			newTickInterval := c.updateInterval()
			if tickInterval != newTickInterval {
				tickInterval = newTickInterval
				ticker.Reset(newTickInterval)
			}

			capacity := c.limits.Len()
			capacity += capacity / 10 // size may grow while we range, try to avoid reallocating in that case
			usage := make(map[string]calls, capacity)
			fallbacks, globals := 0, 0
			c.limits.Range(func(k string, v *internal.FallbackLimiter) bool {
				limit, fallback, actual, usingFallback := v.Collect()
				if usingFallback {
					fallbacks++
				} else {
					globals++
				}
				usage[k] = calls{
					allowed:  actual.Allowed,
					rejected: actual.Rejected,
				}
				c.sendMetrics(k, "global", limit)      // for shadow verification
				c.sendMetrics(k, "fallback", fallback) // for shadow verification
				c.sendMetrics(k, "actual", actual)     // for user-side behavior monitoring

				return true
			})
			// track how often we're using fallbacks vs non-fallbacks.
			// this also tells us about how many keys are used per host.
			// timer-sums unfortunately must be divided by the update frequency to get correct totals.
			c.scope.RecordTimer(metrics.GlobalRatelimiterFallbackUsageTimer, time.Duration(fallbacks))
			c.scope.RecordTimer(metrics.GlobalRatelimiterGlobalUsageTimer, time.Duration(globals))

			c.doUpdate(now.Sub(lastGatherTime), usage)

			lastGatherTime = now
		}
	}
}

func (c *Collection) sendMetrics(globalKey string, limitType string, usage internal.UsageMetrics) {
	scope := c.scope.Tagged(
		metrics.GlobalRatelimitKeyTag(globalKey),
		metrics.GlobalRatelimitTypeTag(limitType),
	)
	scope.AddCounter(metrics.GlobalRatelimiterAllowedRequestsCount, int64(usage.Allowed))
	scope.AddCounter(metrics.GlobalRatelimiterRejectedRequestsCount, int64(usage.Rejected))
}

// doUpdate pushes usage data to aggregators
func (c *Collection) doUpdate(since time.Duration, usage map[string]calls) {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second) // TODO: configurable
	defer cancel()
	_ = ctx

	// TODO: use the RPC code to shard requests to all aggregators, and apply the updated RPS values.
	perAggregator := func(ctx context.Context, since time.Duration, requestedBatch map[string]calls) error {
		// do the call, get the result
		resultBatch, err := make(map[string]float64, len(requestedBatch)), error(nil)

		if err == nil {
			for k, v := range resultBatch {
				delete(requestedBatch, k) // clean up the list so we know what was missed

				if v < 0 {
					// negative values cannot be valid, so they're a failure.
					//
					// this is largely for future-proofing and to cover all possibilities,
					// so unrecognized values lead to fallback behavior because they cannot be understood.
					c.limits.Load(k).FailedUpdate()
				} else {
					c.limits.Load(k).Update(v)
				}
			}
		}

		// mark all non-returned limits as failures.
		// this also handles request errors: all requested values are failed
		// because none of them have been deleted above.
		//
		// aside from request errors this should be extremely rare and might
		// represent a race in the aggregator, but the semantics for handling it
		// are pretty clear either way.
		for k := range requestedBatch {
			// requested but not returned, bump the fallback fuse
			c.limits.Load(k).FailedUpdate()
		}

		if err != nil {
			c.logger.Error("aggregator batch failed to update",
				tag.Error(err),
				tag.Dynamic("agg_host", "hostname?"))
		}
		return err
	}
	_ = perAggregator // make the compiler happy about this pseudocode

}
