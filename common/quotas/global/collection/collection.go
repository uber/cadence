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
	"errors"
	"fmt"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/quotas/global/collection/internal"
)

type (
	// Collection wraps three kinds of ratelimiter collections:
	//   1. a "global" collection, which tracks usage, and is weighted to match request patterns.
	//   2. a "local" collection, which tracks usage, and can be optionally used (and/or shadowed) instead of "global"
	//   3. a "disabled" collection, which does NOT track usage, and is used to bypass as much of this Collection as possible
	//
	// 1 is the reason this Collection exists - limiter-usage has to be tracked and submitted to aggregating hosts to
	// drive the whole "global load-balanced ratelimiter" system.  internally, this will fall back to a local collection
	// if there is insufficient data or too many errors.
	//
	// 2 is the previous non-global behavior, where ratelimits are determined locally, more simply, and load information is not shared.
	// this is a lower-cost and MUCH less complex system, and it SHOULD be used if your Cadence cluster receives requests
	// in a roughly random way (e.g. any client-side request goes to a roughly-fair roughly-random Frontend host).
	//
	// 3 disables as much of 1 and 2 as possible, and is intended to be temporary.
	// it is essentially a maximum-safety fallback during initial rollout, and should be removed once 2 is demonstrated
	// to be safe enough to use in all cases.
	//
	// And last but not least:
	// 1's local-fallback and 2 MUST NOT share ratelimiter instances, or the local instances will be double-counted when shadowing.
	// they should likely be configured to behave identically, but they need to be separate instances.
	Collection struct {
		updateInterval dynamicconfig.DurationPropertyFn

		logger log.Logger
		scope  metrics.Scope

		global   *internal.AtomicMap[string, *internal.FallbackLimiter]
		local    *internal.AtomicMap[string, internal.CountedLimiter]
		disabled *quotas.Collection

		ctx       context.Context // context used for background operations, canceled when stopping
		ctxCancel func()
		stopped   chan struct{} // closed when stopping is complete

		keyModes dynamicconfig.StringPropertyWithRatelimitKeyFilter

		// now exists largely for tests, elsewhere it is always time.Now
		timesource clock.TimeSource
	}

	// calls is a small temporary pair[T] container
	calls struct {
		allowed, rejected int
	}
)

// basically an enum for key values.
// when plug-in behavior is allowed, this will eventually be from a parsed string,
// and values (aside from disabled) are not known at compile time.
type keyMode string

var (
	modeDisabled          keyMode = "modeDisabled"
	modeLocal             keyMode = "local"
	modeGlobal            keyMode = "global"
	modeLocalShadowGlobal keyMode = "local-shadow-global"
	modeGlobalShadowLocal keyMode = "global-shadow-global"
)

func New(
	// quotas for "local only" behavior.
	// used for both "local" and "modeDisabled" behavior, though "local" wraps the values.
	local *quotas.Collection,
	// quotas for the global limiter's internal fallback.
	//
	// this should be configured the same as the local collection, but it
	// MUST NOT actually be the same collection, or shadowing will double-count
	// events on the fallback.
	global *quotas.Collection,
	updateInterval dynamicconfig.DurationPropertyFn,
	keyModes dynamicconfig.StringPropertyWithRatelimitKeyFilter,
	logger log.Logger,
	met metrics.Client,
) (*Collection, error) {
	if local == global {
		return nil, errors.New("local and global collections must be different")
	}

	globalCollection := internal.NewAtomicMap(func(key string) *internal.FallbackLimiter {
		return internal.NewFallbackLimiter(global.For(key))
	})
	localCollection := internal.NewAtomicMap(func(key string) internal.CountedLimiter {
		return internal.NewCountedLimiter(local.For(key))
	})
	ctx, cancel := context.WithCancel(context.Background())
	c := &Collection{
		global:         globalCollection,
		local:          localCollection,
		disabled:       local,
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
	return c, nil
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

func (c *Collection) keyMode(key string) keyMode {
	rawMode := keyMode(c.keyModes(key))
	switch rawMode {
	case modeLocal, modeGlobal, modeLocalShadowGlobal, modeGlobalShadowLocal, modeDisabled:
		return rawMode
	default:
		return modeDisabled
	}
}

func (c *Collection) For(key string) quotas.Limiter {
	switch c.keyMode(key) {
	case modeLocal:
		return c.local.Load(key)
	case modeGlobal:
		return c.global.Load(key)
	case modeLocalShadowGlobal:
		return internal.NewShadowedLimiter(c.local.Load(key), c.global.Load(key))
	case modeGlobalShadowLocal:
		return internal.NewShadowedLimiter(c.global.Load(key), c.local.Load(key))

	default:
		// pass through to the fallback, as if this collection did not exist.
		// this means usage cannot be collected or shadowed, so changing to or
		// from "disable" may allow a burst of requests beyond intended limits.
		//
		// this is largely intended for safety during initial rollouts, not
		// normal use - normally "fallback" or "global" should be used.
		// "fallback" SHOULD behave the same, but with added monitoring, and
		// the ability to warm caches in either direction before switching.
		return c.disabled.For(key)
	}
}

func (c *Collection) shouldDeleteKey(mode keyMode) bool {
	switch mode {
	case modeLocal, modeGlobal, modeLocalShadowGlobal, modeGlobalShadowLocal:
		return false

	default:
		return true
	}
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
			// update interval if it changed
			newTickInterval := c.updateInterval()
			if tickInterval != newTickInterval {
				tickInterval = newTickInterval
				ticker.Reset(newTickInterval)
			}

			// submit local metrics asynchronously, because there's no need to do it synchronously
			localMetricsDone := make(chan struct{})
			go func() {
				defer close(localMetricsDone)
				c.local.Range(func(k string, v internal.CountedLimiter) bool {
					if c.shouldDeleteKey(c.keyMode(k)) {
						c.local.Delete(k) // clean up so maintenance cost is near zero
						return true       // continue iterating, delete others too
					}

					c.sendMetrics(k, "local", v.Collect())
					return true
				})
			}()

			capacity := c.global.Len()
			capacity += capacity / 10 // size may grow while we range, try to avoid reallocating in that case
			usage := make(map[string]calls, capacity)
			startups, failings, globals := 0, 0, 0
			c.global.Range(func(k string, v *internal.FallbackLimiter) bool {
				if c.shouldDeleteKey(c.keyMode(k)) {
					c.global.Delete(k) // clean up so maintenance cost is near zero
					return true        // continue iterating, delete others too
				}

				counts, startup, failing := v.Collect()
				if startup {
					startups++
				}
				if failing {
					failings++
				}
				if !(startup || failing) {
					globals++
				}
				usage[k] = calls{
					allowed:  counts.Allowed,
					rejected: counts.Rejected,
				}
				c.sendMetrics(k, "global", counts)

				return true
			})

			// track how often we're using fallbacks vs non-fallbacks.
			// this also tells us about how many keys are used per host.
			// timer-sums unfortunately must be divided by the update frequency to get correct totals.
			c.scope.RecordTimer(metrics.GlobalRatelimiterStartupUsageTimer, time.Duration(startups))
			c.scope.RecordTimer(metrics.GlobalRatelimiterFallbackUsageTimer, time.Duration(failings))
			c.scope.RecordTimer(metrics.GlobalRatelimiterGlobalUsageTimer, time.Duration(globals))

			c.doUpdate(now.Sub(lastGatherTime), usage)

			<-localMetricsDone // should be much faster than doUpdate, unless it's no-opped

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
	if len(usage) == 0 {
		return // modeDisabled or just no activity on this collection of limits, nothing to do
	}

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
					c.global.Load(k).FailedUpdate()
				} else {
					// TODO: needs to be scaled by the configured ratelimit
					c.global.Load(k).Update(rate.Limit(v))
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
			c.global.Load(k).FailedUpdate()
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
