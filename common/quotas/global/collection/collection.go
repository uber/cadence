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
  - periodically report usage to an "aggregator" host
  - apply the aggregator's returned per-key RPS limits to future requests
  - fall back to the wrapped limiter in case of failures
*/
package collection

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"

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
		updateRate dynamicconfig.DurationPropertyFn

		logger log.Logger
		scope  metrics.Scope

		limits *internal.AtomicMap[string, *internal.FallbackLimiter]

		lifecycle *internal.OneShot

		// now exists largely for tests, elsewhere it is always time.Now
		timesource clock.TimeSource
	}

	// Limiter is a simplified quotas.Limiter API, covering the only global-ratelimiter-used API: Allow.
	//
	// This is largely because Reserve() is essentially impossible to monitor for usage-tracking purposes.
	// Once that's wrapped with a replacement, we can likely support the full API here.
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
	updateRate dynamicconfig.DurationPropertyFn,
	logger log.Logger,
	scope metrics.Scope,
) *Collection {
	contents := internal.NewAtomicMap(func(key string) *internal.FallbackLimiter {
		return internal.NewFallbackLimiter(fallback.GetLimiter(key))
	})
	return &Collection{
		limits:     contents,
		updateRate: updateRate,

		logger: logger, // TODO: tag it
		scope:  scope,  // TODO: tag it

		lifecycle: internal.NewOneShot(),

		// override externally in tests
		timesource: clock.NewRealTimeSource(),
	}
}

func (c *Collection) TestOverrides(t *testing.T, timesource clock.TimeSource) {
	t.Helper()
	c.timesource = timesource
}

func (c *Collection) Start(ctx context.Context) error {
	err := c.lifecycle.Start(func() error {
		go func() {
			defer func() {
				// bad but not worth crashing the process
				log.CapturePanic(recover(), c.logger, nil)
			}()
			defer func() {
				c.lifecycle.Stopped()
			}()
			c.backgroundUpdateLoop()
		}()
		return nil
	})
	if err != nil {
		return fmt.Errorf("*Collection failed to start: %w", err)
	}
	return nil
}

func (c *Collection) Stop(ctx context.Context) error {
	err := c.lifecycle.StopAndWait(ctx)
	if err != nil {
		return fmt.Errorf("*Collection failed to stop: %w", err)
	}
	return nil
}

func (c *Collection) For(key string) Limiter {

}

func (c *Collection) backgroundUpdateLoop() {
	tickRate := c.updateRate()

	logger := c.logger.WithTags(tag.Dynamic("limiter", "update"))
	logger.Debug("update loop starting")

	ticker := c.timesource.NewTicker(tickRate)
	lastGatherTime := c.timesource.Now()
	for {
		select {
		case <-c.lifecycle.Context().Done():
			return
		case <-ticker.Chan():
			now := c.timesource.Now()
			logger.Debug("update tick")
			// update tick-rate if it changed
			newTickRate := c.updateRate()
			if tickRate != newTickRate {
				tickRate = newTickRate
				ticker.Reset(newTickRate)
			}

			capacity := c.limits.Len()
			capacity += capacity / 10 // size may grow while we range, try to avoid that realloc too
			usage := make(map[string]calls, capacity)
			c.limits.Range(func(k string, v *internal.FallbackLimiter) bool {
				allowed, rejected, _ := v.Collect()
				usage[k] = calls{allowed: allowed, rejected: rejected}
				return true
			})

			c.doUpdate(now.Sub(lastGatherTime), usage)

			lastGatherTime = now
		}
	}
}

// doUpdate pushes usage data to aggregators
func (c *Collection) doUpdate(since time.Duration, usage map[string]calls) {
	ctx, cancel := context.WithTimeout(c.lifecycle.Context(), 10*time.Second) // TODO: configurable
	defer cancel()
	_ = ctx

	// TODO: use the RPC code to shard requests to all aggregators, and apply the updated RPS values.
	perAggregator := func(since time.Duration, requestedBatch map[string]calls) error {
		// do the call
		resultBatch, err := make(map[string]float64, len(requestedBatch)), error(nil)

		if err == nil {
			for k, v := range resultBatch {
				delete(requestedBatch, k) // clean up the list so we know what was missed

				if v < 0 {
					// error loading this key, do something
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
				tag.Dynamic("host", "hostname?"))
		}
		return err
	}
	_ = perAggregator
	}
}
