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

// Package internal protects these types' concurrency primitives and other internals from accidents.
// this has been rolled back in favor of simplicity and avoiding premature optimization before measuring,
// but the internal-separation is still broadly a good thing.  these are not dumb objects.
package internal

import (
	"math"
	"sync/atomic"

	"golang.org/x/time/rate"

	"github.com/uber/cadence/common/quotas"
)

type (
	// FallbackLimiter wraps a [rate.Limiter] with a fallback [quotas.Limiter] to use
	// after a configurable number of failed updates.
	//
	// Intended use is:
	//   - collect allowed vs rejected metrics
	//   - periodically, the limiting host gathers all FallbackLimiter metrics and zeros them
	//   - this info is submitted to aggregating hosts, who compute new target RPS values
	//   - these new updates adjust this ratelimiter
	//
	// During this sequence, an individual limit may not be returned for two major reasons:
	//   - this limit is legitimately unused and no data exists for it
	//   - ring re-sharding led to losing track of the limit, and it is now owned by a host without sufficient data
	//
	// To mitigate the impact of the second case, "insufficient data" responses from aggregating hosts
	// are temporarily ignored, and the previously-configured update is used.  This gives the aggregating host time
	// to fill in its data, and then the next cycle should use "real" values.
	//
	// If no data has been returned for a sufficiently long time, the "smart" ratelimit will be dropped, and
	// the fallback limit will be used exclusively.  This is intended as a safety fallback, e.g. during initial
	// rollout and outage scenarios, normal use is not expected to rely on it.
	//
	// ---
	//
	// Note that this object has no locks, despite being "mutated".
	// Both [quotas.Limiter] and [rate.Limiter] have internal locks and the instances
	// are never changed, and the same is true in a sense for the atomic counters.
	//
	// If this struct changes, top-level locks may be necessary.
	// `go vet -copylocks` should detect this, but be careful regardless.
	FallbackLimiter struct {
		// usage data cannot be gathered from rate.Limiter, sadly.
		// so we need to gather it separately. or maybe find a fork.
		//
		// note that these atomics are values, not pointers.
		// this requires that FallbackLimiter is used as a pointer, not a value, and is never copied.
		// if this is not done, `go vet` will produce an error:
		//     ❯ go vet .
		//     # github.com/uber/cadence/common/quotas/global/limiter
		//     ./limiter.go:30:27: func passes lock by value: github.com/uber/cadence/common/quotas/global/limiter/internal.FallbackLimiter contains sync/atomic.Int64 contains sync/atomic.noCopy
		// which is checked by `make lint` during CI.

		accepted      atomic.Int64 // accepted-request usage data, value-typed because the parent is never copied
		rejected      atomic.Int64 // rejected-request usage data, value-typed because the parent is never copied
		failedUpdates atomic.Int64 // number of failed updates, value-typed because the parent is never copied

		// ratelimiters in use

		fallback quotas.Limiter // fallback when limit is nil, accepted until updated
		limit    *rate.Limiter  // local-only limiter based on remote data.
	}
)

const (
	// when failed updates exceeds this value, use the fallback
	maxFailedUpdates = 9

	// at startup / new limits in use, use the fallback logic, because that's
	// expected as we have no data yet.
	//
	// this risks being interpreted as a "failure" though, so start deeply
	// negative so it can be identified as "still starting up".
	// min-int64 has more than enough room to "count" failed updates for eons
	// without becoming positive, so it should not risk being misinterpreted.
	initialFailedUpdates = math.MinInt64
)

func New(fallback quotas.Limiter) *FallbackLimiter {
	l := &FallbackLimiter{
		fallback: fallback,
		limit:    rate.NewLimiter(rate.Limit(1), 0), // 0 allows no requests, will be unused until we receive an update
	}
	l.failedUpdates.Store(initialFailedUpdates)
	return l
}

// Collect returns the current accepted/rejected values, and resets them to zero.
// Small bits of imprecise counting / time-bucketing due to concurrent limiting is expected and allowed,
// as it should be more than small enough in practice to not matter.
func (b *FallbackLimiter) Collect() (accepted int, rejected int, usingFallback bool) {
	accepted = int(b.accepted.Swap(0))
	rejected = int(b.rejected.Swap(0))
	return accepted, rejected, b.useFallback()
}

func (b *FallbackLimiter) useFallback() bool {
	failed := b.failedUpdates.Load()
	return failed < 0 /* not yet set */ || failed > maxFailedUpdates /* too many failures */
}

// Update adjusts the underlying "main" ratelimit, and resets the fallback fuse.
func (b *FallbackLimiter) Update(rps float64) {
	// caution: order here matters, to prevent potentially-old limiter values from being used
	// before they are updated.
	//
	// this is probably not going to be noticeable, but some users are sensitive to ANY
	// requests being ratelimited.  updating the fallback fuse last should be more reliable
	// in preventing that from happening when they would not otherwise be limited, e.g. with
	// the initial value of 0 burst, or if their rps was very low long ago and is now high.
	defer func() {
		// reset the use-fallback fuse, which may also (re)enable using the main limiter
		b.failedUpdates.Store(0)
	}()

	if b.limit.Limit() == rate.Limit(rps) {
		return
	}

	b.limit.SetLimit(rate.Limit(rps))
	b.limit.SetBurst(max(1, int(rps))) // 0 burst disallows all requests, so allow at least 1 and rely on rps to fill sanely
}

// FailedUpdate should be called when a key fails to update from an aggregator,
// possibly implying some kind of problem, possibly with this key.
//
// After crossing a threshold of failures (currently 10), the fallback will be switched to.
func (b *FallbackLimiter) FailedUpdate() (failures int) {
	failures = int(b.failedUpdates.Add(1)) // always increment the count for monitoring purposes
	return failures
}

// Clear erases the internal ratelimit, and defers to the fallback until an update is received.
// This is intended to be used when the current limit is no longer trustworthy for some reason,
// determined externally rather than from too many FailedUpdate calls.
func (b *FallbackLimiter) Clear() {
	b.limit = nil
}

// Allow returns true if a request is allowed right now.
func (b *FallbackLimiter) Allow() bool {
	var allowed bool
	if b.useFallback() {
		allowed = b.fallback.Allow()
	} else {
		allowed = b.limit.Allow()
	}

	if allowed {
		b.accepted.Add(1)
	} else {
		b.rejected.Add(1)
	}
	return allowed
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
