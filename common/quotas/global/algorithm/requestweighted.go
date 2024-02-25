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
Package algorithm contains a running-weighted-average calculator for ratelimits,
and some associated types to prevent accidental confusion between the various
floats/ints/etc involved.

This package is intentionally unaware of RPC or any desired RPS limits, it just
tracks the observed rates of requests and returns proportions, so other code can
determine RPS values.

This is both to simplify testing and to keep its internals as fast as possible, as
it is relatively CPU heavy and needs to be quick to prevent a snowballing backlog
of concurrent updates and requests for aggregated data.

# How this fits in the global ratelimiting system

Going back to the [github.com/uber/cadence/common/quotas/global] package diagram:
each aggregating host will have one or more instances of this weight-calculator,
and it will receive a shard worth of request data.

Though this package has no direct connection to RPC structures, the request data is
expected to closely match the argument to RequestWeighted.Update, and the response
data like the return value of RequestWeighted.HostWeights.

This makes aggregating hosts intentionally very simple outside of this package:
they essentially just forward the request and response, multiplying by dynamicconfig
intended-ratelimit values (which occurs outside the mutex, to minimize contention).

# Expected use

Limiting hosts collect metrics and submit it to an aggregating host via
RequestWeighted.Update, through [some kind of RPC setup].

Once updated, the aggregating host can get the RequestWeighted.HostWeights for
that host's ratelimits (== the updated limits), multiply those 0..1 weights by
the dynamicconfig configured RPS, and return them to the limiting host that
triggered the request.

If there are unused RPS remaining, the aggregating host *may* increase the RPS
it returns by [some amount], to allow the limiting host to pre-emptively allow
more requests than is "fair" before the next update.  Even if it does not, an
increase in attempted usage will increase that limiter's weight on the next
update cycle, so this is mostly intended as a tool for reducing incorrectly-rejected
requests when a ratelimit's usage is well below its allowed limit.

# Dealing with expired data

As user calls change, or the aggregating-host ring changes, Limit keys may become
effectively unused in an instance.  During normal use, any accessed Limit will
clean up "expired" data if it is found, and there is essentially no ongoing
"upkeep" cost for an un-accessed Limit (aside from memory).

Hopefully this will be sufficient to keep memory use and performance reasonable.

If it is not, a trivial goroutine to periodically call RequestWeighted.GC will
clear *all* old data.  Every minute or so should be more than sufficient.
This method returns some simple metrics about how much data exists / was removed,
so it can be reported to help us estimate how necessary it is in practice.

# Dealing with excessive contention

In large clusters, there will be likely be many update-requests, many Limit keys,
and more data to process to return responses (more hosts per limit).  At some point
this could become a bottleneck, preventing timely updates.

On even our largest internal clusters, I do not believe we will run that risk.
At least, not with the planned frontend-only limits.  `pprof` should be able
to easily validate this and let us better estimate headroom for adding more in
the future.

If too much contention does occur, there are 3 relatively simple mitigations:
 1. turn it completely off, go back to host-local limits
 2. use N instances and shard keys across them
 3. slow down the update frequency

1 is pretty obvious and has clear behavior, though it means degrading behavior
for our imbalanced-load clusters.

2 is a relatively simple mitigation: create N instances instead of 1, and shard
the Limit keys to each instance.  Since this is mostly CPU bound and each one
would be fully independent, making GOMAXPROCS instances is an obvious first choice,
and this does not need to be dynamically reconfigured at runtime so there should
be no need to build a "smart" re-balancing / re-sharding system of any kind.

Otherwise, 3 (changing the limiter-update-rate to be slower) is essentially the only
other short-term and dynamically-apply-able option.  This will impact how quickly the
algorithm converges, so you may also want to adjust the new-data weight to be higher,
though "how much" depends on what kind of behavior you want.

Tests contain a single-threaded benchmark and laptop-local results for estimating
contention, and for judging if changes will help or hinder, but real-world use and
different CPUs will of course be somewhat different.
Personally I expect "few mutexes, GOMAXPROCS instances" is roughly ideal for CPU
throughput with the current setup, but an entirely different internal structure
might exceed it.
*/
package algorithm

import (
	"math"
	"sync"
	"time"

	"golang.org/x/exp/constraints"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
)

type (
	// history holds the running per-second running average for a single key from a single host, and the last time it was updated.
	history struct {
		lastUpdate         time.Time // only so we know if elapsed times are exceeded, not used to compute per-second rates
		accepted, rejected PerSecond // requests received, per second (conceptually divided by update rate)
	}

	// Identity is an arbitrary (stable) identifier for a "limiting" host.
	Identity string
	// Limit is the key being used to Limit requests.
	Limit string
	// PerSecond represents already-scaled-to-per-second RPS values
	PerSecond float64
	// HostWeight is a float between 0 and 1, showing how much of a ratelimit should be allocated to a host.
	HostWeight float64

	impl struct {
		// intentionally value-typed so caller cannot mutate the fields.
		// manually copy data if this changes.
		cfg Config

		// mut protects usage, as it is the only mutable data
		mut sync.Mutex
		// usage data for ratelimits, per host.
		//
		// data is first keyed on the limit and then on the host, as it's assumed that there will be
		// many times more limit keys than limiting hosts, and this will reduce cardinality more rapidly.
		usage map[Limit]map[Identity]history

		clock clock.TimeSource
	}

	// Metrics reports overall counts discovered as part of garbage collection.
	//
	// This is not necessarily worth maintaining verbatim, but while it's easy to
	// collect it might give us some insight into overall behavior.
	Metrics struct {
		// HostLimits is the number of per-host limits that remain after cleanup
		HostLimits int
		// Limits is the number of cross-host limits that remain after cleanup
		Limits int

		// RemovedHostLimits is the number of per-host limits that were removed as part of this GC pass
		RemovedHostLimits int
		// RemovedLimits is the number of cross-host limits that were removed as part of this GC pass
		RemovedLimits int
	}

	// Requests contains accepted/rejected request counts for a ratelimit, as part of an update operation.
	// Only intended to be used in [RequestWeighted.Update].
	Requests struct {
		Accepted, Rejected int
	}

	// RequestWeighted returns aggregated RPS and distribution data for its Limit keys,
	// based on how many requests each Identity has received.
	RequestWeighted interface {
		// Update load-data for this host's requests, given a known elapsed time spent accumulating this load info.
		//
		// Elapsed time must be non-zero, but is not otherwise constrained.
		Update(host Identity, load map[Limit]Requests, elapsed time.Duration)

		// HostWeights returns the per-[Limit] weights for all requested + known keys for this Identity,
		// as well as the Limit's overall used RPS (to decide RPS to allow for new hosts).
		HostWeights(host Identity, limits []Limit) (weights map[Limit]HostWeight, usedRPS map[Limit]PerSecond)

		// GC can be called periodically to pre-emptively prune old ratelimits.
		//
		// Limit keys that are accessed normally will automatically garbage-collect themselves and old host data,
		// and an unused limit only costs memory, so this may prove to be unnecessary.
		GC() Metrics
	}

	Config struct {
		// How much each update should be weighted vs prior data.
		// Must be between 0 and 1, recommend starting with 0.5 (4 updates until data has <10% influence)
		NewDataWeight dynamicconfig.FloatPropertyFn

		// Expected rate of updates.  Should match the cluster's config for how often limiters check in.
		UpdateRate dynamicconfig.DurationPropertyFn

		// How many UpdateRate periods can pass without new information before considering a host-limit's
		// RPS "probably inactive", rather than simply delayed.
		//
		// Should always be 1 or larger.  Unsure about a good default (try 2+?), but larger numbers mean smoother
		// behavior but longer delays on adjusting to hosts that have disappeared or stopped receiving requests.
		MaxMissedUpdates dynamicconfig.FloatPropertyFn

		// How many UpdateRate periods can pass without receiving an update before completely deleting data.
		//
		// Due to ever-reducing weight after MaxMissedUpdates, this is intended to reduce calculation costs,
		// not influence behavior / weights returned.  Even extremely-low weighted hosts will still be retained
		// as long as they keep reporting "in use".
		//
		// E.g. with 0.5 NewDataWeight and 2.0 MaxMissedUpdates, after 10 missed updates the most-recent data
		// will be reduced to ~0.1% and may not be worth keeping.
		// "Good" values depend on a lot of details, but >=10 seems reasonably safe.
		GcAfter dynamicconfig.IntPropertyFn
	}
	configSnapshot struct {
		now       time.Time
		weight    float64
		rate      time.Duration
		maxMissed float64
		gcAfter   int
	}
)

// these could be configurable, but it's not expected to be a noticeable performance concern.
const (
	guessNumKeys   = 1024 // guesstimate at num of ratelimit keys in a cluster
	guessHostCount = 32   // guesstimate at num of frontend hosts in a cluster that receive traffic for each key
)

// New returns a concurrency-safe host-weight aggregator.
//
// This instance is effectively single-threaded, but a small sharding wrapper should allow better concurrent
// throughput if needed (bound by CPU cores, as it's moderately CPU-costly).
func New(cfg Config) (RequestWeighted, error) {
	return &impl{
		cfg:   cfg,
		usage: make(map[Limit]map[Identity]history, guessNumKeys), // start out relatively large

		clock: clock.NewRealTimeSource(),
	}, nil
}

// Update performs a weighted update to the running RPS for this host's per-key request data
func (a *impl) Update(id Identity, load map[Limit]Requests, elapsed time.Duration) {
	a.mut.Lock()
	defer a.mut.Unlock()

	snap := a.snapshot()
	for key, req := range load {
		ih := a.usage[key]
		if ih == nil {
			ih = make(map[Identity]history, guessHostCount)
		}

		var next history
		prev := ih[id]
		aps := PerSecond(float64(req.Accepted) / float64(elapsed/time.Second))
		rps := PerSecond(float64(req.Rejected) / float64(elapsed/time.Second))
		if prev.lastUpdate.IsZero() {
			next = history{
				lastUpdate: snap.now,
				accepted:   aps, // no history == 100% weight
				rejected:   rps, // no history == 100% weight
			}
		} else {
			// account for missed updates.
			// ignoring gc as we are updating, and a very-low reduce just means similar results to zero data.
			reduce, _ := missedUpdateScalar(snap.now.Sub(prev.lastUpdate), snap)

			next = history{
				lastUpdate: snap.now,
				accepted:   weighted(aps, prev.accepted*reduce, snap.weight),
				rejected:   weighted(rps, prev.rejected*reduce, snap.weight),
			}
		}

		ih[id] = next
		a.usage[key] = ih
	}
}

// getWeightsLocked returns the weights of observed hosts (based on ALL requests), and the total number of requests accepted per second.
func (a *impl) getWeightsLocked(key Limit, snap configSnapshot) (weights map[Identity]HostWeight, usedRPS PerSecond) {
	ih := a.usage[key]
	if len(ih) == 0 {
		return
	}

	weights = make(map[Identity]HostWeight, len(ih))
	total := HostWeight(0.0)
	for id, history := range ih {
		// account for missed updates
		reduce, gc := missedUpdateScalar(snap.now.Sub(history.lastUpdate), snap)
		if gc {
			// old, clean up
			delete(ih, id)
			continue
		}

		actual := HostWeight((history.accepted + history.rejected) * reduce)

		weights[id] = actual // populate with the reduced values so it doesn't have to be calculated again
		total += actual      // keep a running total to scale all values when done
		usedRPS += history.accepted * reduce
	}

	if len(ih) == 0 {
		// completely empty Limit, gc it as well
		delete(a.usage, key)
		return nil, 0
	}

	for id := range ih {
		// scale by the total.
		// this also ensures all values are between 0 and 1 (inclusive)
		weights[id] = weights[id] / total
	}
	return weights, usedRPS
}

func (a *impl) HostWeights(host Identity, limits []Limit) (weights map[Limit]HostWeight, usedRPS map[Limit]PerSecond) {
	a.mut.Lock()
	defer a.mut.Unlock()

	weights = make(map[Limit]HostWeight, len(limits))
	usedRPS = make(map[Limit]PerSecond, len(limits))
	snap := a.snapshot()
	for _, lim := range limits {
		hosts, used := a.getWeightsLocked(lim, snap)
		if len(hosts) > 0 {
			usedRPS[lim] = used // limit is known, has some usage
			if weight, ok := hosts[host]; ok {
				weights[lim] = weight // host has a known weight
			}
		}
	}
	return weights, usedRPS
}

func (a *impl) GC() Metrics {
	a.mut.Lock()
	defer a.mut.Unlock()

	m := Metrics{}
	snap := a.snapshot()
	// TODO: too costly? can check the first-N% each time and it'll eventually visit all keys, demonstrated in tests.
	for lim, dat := range a.usage {
		for host, hist := range dat {
			if _, gc := missedUpdateScalar(snap.now.Sub(hist.lastUpdate), snap); gc {
				// clean up stale host data within limits
				delete(dat, host)
				m.RemovedHostLimits++
			} else {
				m.HostLimits++
			}
		}

		// clean up stale limits
		if len(dat) == 0 {
			delete(a.usage, lim)
			m.RemovedLimits++
		} else {
			m.Limits++
		}
	}

	return m
}

// returns a snapshot of config and "now" for easier chaining through calls, and reducing calls to
// non-trivial field types like dynamic config.
func (a *impl) snapshot() configSnapshot {
	return configSnapshot{
		now:       a.clock.Now(),
		weight:    a.cfg.NewDataWeight(),
		rate:      a.cfg.UpdateRate(),
		maxMissed: a.cfg.MaxMissedUpdates(),
		gcAfter:   a.cfg.GcAfter(),
	}
}

// missedUpdateScalar returns an amount to multiply old RPS data by, to account for missed updates outside SLA.
func missedUpdateScalar(elapsed time.Duration, cfg configSnapshot) (scalar PerSecond, gc bool) {
	reduce := PerSecond(1.0)

	// fast path: check the bounds for "new enough to ignore"
	if elapsed <= time.Duration(float64(cfg.rate)*cfg.maxMissed) {
		return reduce, false // within SLA, no changes
	}
	// fast path: check the bounds for "old enough to prune"
	if elapsed >= cfg.rate*time.Duration(cfg.gcAfter) {
		return reduce, true
	}

	// slow path: account for missed updates by simulating 0-value updates.

	// calculate missed updates, and compute an exponential decay as if we had kept track along the way
	missed := float64(elapsed) / float64(cfg.rate) // how many missed updates (fractional, 1 == one full duration passed, 2.5 == 2.5 durations passed, etc)
	if missed >= cfg.maxMissed {
		// missed at least one update period beyond decay-after, calculate an exponential decay to the old values.
		// floor to an int when doing so because:
		// - precision isn't important
		// - tests are a bit easier (more stable / less crazy-looking values)
		// - as a bonus freebie: integer exponents are typically faster to compute
		reduce = PerSecond(math.Pow(cfg.weight, math.Floor(missed)))
	}

	return reduce, false
}

func weighted[T numeric](newer, older T, weight float64) T {
	return T((float64(newer) * weight) +
		(float64(older) * (1 - weight)))
}

type numeric interface {
	constraints.Integer | constraints.Float
}
