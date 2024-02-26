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

package algorithm

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
)

// just simplifies newForTest usage as most tests only care about rate
func defaultConfig(rate time.Duration) configSnapshot {
	return configSnapshot{
		// now:    , ignored

		// intentionally avoiding 0.5 because it cannot tell if the code uses
		// `weight` or `1-weight`, which is usually relevant.
		//
		// 0.1 is relatively human-math-friendly for a single step,
		// but is otherwise arbitrary.
		weight:    0.1,
		rate:      rate,
		maxMissed: 2.0,
		gcAfter:   10,
	}
}

func newForTest(t require.TestingT, snap configSnapshot) (*impl, clock.MockedTimeSource) {
	cfg := Config{
		NewDataWeight: func(opts ...dynamicconfig.FilterOption) float64 {
			return snap.weight
		},
		UpdateRate: func(opts ...dynamicconfig.FilterOption) time.Duration {
			return snap.rate
		},
		MaxMissedUpdates: func(opts ...dynamicconfig.FilterOption) float64 {
			return snap.maxMissed
		},
		GcAfter: func(opts ...dynamicconfig.FilterOption) int {
			return snap.gcAfter
		},
	}
	agg, err := New(cfg)
	require.NoError(t, err)

	underlying := agg.(*impl)
	tick := clock.NewMockedTimeSource()
	underlying.clock = tick

	// adjust time to get rid of sub-second output, it's just harder to read.
	// doesn't matter if this goes forward or backward.
	tick.Advance(tick.Now().Sub(tick.Now().Round(time.Second)))

	return underlying, tick
}

func TestMissedUpdateHandling(t *testing.T) {
	agg, tick := newForTest(t, defaultConfig(time.Second))

	h1, h2 := Identity("host 1"), Identity("host 2")
	key := Limit("key")
	agg.Update(h1, map[Limit]Requests{key: {1, 1}}, time.Second)
	agg.Update(h2, map[Limit]Requests{key: {1, 1}}, time.Second)
	weights, used := agg.HostWeights(h1, []Limit{key})
	assert.Len(t, weights, 1)                // 1 key
	assert.Equal(t, used[key], PerSecond(2)) // only 2 accepted

	// advance 1 second, should not change anything
	tick.Advance(time.Second)
	weights, used = agg.HostWeights(h1, []Limit{key})
	assert.Len(t, weights, 1)                // 1 key
	assert.Equal(t, used[key], PerSecond(2)) // only 2 accepted

	// advance to 2 seconds, this is still within "2 missed updates" so no changes
	tick.Advance(time.Second)
	weights, used = agg.HostWeights(h1, []Limit{key})
	assert.Len(t, weights, 1)                // 1 key
	assert.Equal(t, used[key], PerSecond(2)) // only 2 accepted

	// advance to 3 seconds, just crossing the threshold.
	// should retroactively count missed updates, immediately freeing up RPS
	// as if it was tracking 0s all along, because we're assuming it has been inactive for some reason.
	tick.Advance(time.Second)
	weights, used = agg.HostWeights(h1, []Limit{key})
	assert.Len(t, weights, 1)                            // still tracking the key / not GC'd
	assert.InDelta(t, float64(used[key]), 1.458, 0.0001) // 3 simulated zero-updates: 2 -> 1.8 -> 1.62 -> 1.458
}

func TestGC(t *testing.T) {
	h1, h2 := Identity("host 1"), Identity("host 2")
	key := Limit("key")

	// creates an aggregator and advances time 9 seconds, ensuring that data still exists.
	// advance 1 more second to trigger garbage collection.
	setup := func(t *testing.T) (*impl, clock.MockedTimeSource) {
		agg, tick := newForTest(t, configSnapshot{
			rate:    time.Second,
			gcAfter: 10,

			// irrelevant for these tests but must be non-zero:
			weight:    0.1,
			maxMissed: 2,
		})

		agg.Update(h1, map[Limit]Requests{key: {1, 1}}, time.Second)
		agg.Update(h2, map[Limit]Requests{key: {1, 1}}, time.Second)
		weights, used := agg.HostWeights(h1, []Limit{key})
		// sanity check that we have data
		require.Len(t, weights, 1, "sanity check: should have inserted limit's data")
		require.Equal(t, used[key], PerSecond(2), "sanity check: should have inserted usage data")

		// partially advance, sanity check
		tick.Advance(9 * time.Second)
		weights, used = agg.HostWeights(h1, []Limit{key})
		require.Len(t, weights, 1, "sanity check: should have inserted limit's data after 9s")
		require.Equal(t, used[key], PerSecond(2*math.Pow(0.9, 9)), "sanity check: should have inserted usage data, reduced after 9s")

		return agg, tick
	}

	t.Run("no cleanup before expiration", func(t *testing.T) {
		agg, _ := setup(t)
		met := agg.GC()
		require.Equal(t, Metrics{
			HostLimits:        2,
			Limits:            1,
			RemovedHostLimits: 0,
			RemovedLimits:     0,
		}, met)
	})

	t.Run("cleans up during read", func(t *testing.T) {
		agg, tick := setup(t)

		tick.Advance(time.Second) // advance to 10th second
		// read it out, should detect out-of-date data and clean it up
		weights, used := agg.HostWeights(h1, []Limit{key})
		require.Len(t, weights, 0, "should be no weights for h1")
		require.Len(t, used, 0, "should be no usage data at all")

		// internals should also be empty
		require.Len(t, agg.usage, 0) // also frees memory
	})
	t.Run("retains recent data while cleaning", func(t *testing.T) {
		agg, tick := setup(t)

		// refresh data for one host
		agg.Update(h1, map[Limit]Requests{key: {1, 1}}, time.Second)
		tick.Advance(time.Second) // advance to 10th second

		// read both hosts.  h1 should exist, h2 should not
		weights, used := agg.HostWeights(h1, []Limit{key})
		require.Len(t, weights, 1, "h1 was refreshed and should remain")
		require.Len(t, used, 1, "h1 was refreshed and usage data should remain")
		weights, used = agg.HostWeights(h2, []Limit{key})
		require.Len(t, weights, 0, "h2 should have no weight at all")
		require.Len(t, used, 1, "h1 was refreshed and usage data should remain")

		// internals should also be partly emptied
		require.Len(t, agg.usage, 1, "limit should remain")
		require.Len(t, agg.usage[key], 1, "only one host's data should be in limit")
	})
	t.Run("cleans up by explicit gc", func(t *testing.T) {
		agg, tick := setup(t)
		tick.Advance(time.Second)
		met := agg.GC()
		require.Equal(t, Metrics{
			HostLimits: 0, // none remain
			Limits:     0, // none remain

			RemovedHostLimits: 2, // h1 and h2 for the single key, both removed
			RemovedLimits:     1, // single key removed
		}, met)

		// internals should also be empty
		require.Len(t, agg.usage, 0)
	})
}

func TestRapidlyCoalesces(t *testing.T) {
	// This test ensures that, regardless of starting weights, the algorithm
	// "rapidly" achieves near-actual weight distribution after a small number of rounds.
	//
	// Otherwise, the exact numbers here don't really matter, it's just handy to show the
	// behavior in semi-extreme scenarios.  Logs show a quick adjustment which is what we want.
	// If you're making changes, check with like 10k rounds to make sure it's stable.
	//
	// Time is also not advanced because it doesn't actually need to be advanced.
	// An update is an update, and the caller's elapsed time is assumed to be correct.
	agg, _ := newForTest(t, configSnapshot{
		// Using 0.5 weight because that's what we expect to use IRL, and this test is
		// ensuring that weight is good enough for the behavior we want.
		// Weight-math-correctness is ensured by other tests.
		weight: 0.5,

		// irrelevant / ignored
		rate:      time.Second,
		maxMissed: 2.0,
		gcAfter:   10,
	})

	key := Limit("start workflow")
	h1, h2, h3 := Identity("one"), Identity("two"), Identity("three")

	weights, used := agg.getWeightsLocked(key, agg.snapshot())
	assert.Zero(t, weights, "should have no weights")
	assert.Zero(t, used, "should have no used RPS")

	push := func(host Identity, accept, reject int) {
		agg.Update(host, map[Limit]Requests{
			key: {
				Accepted: accept,
				Rejected: reject,
			},
		}, time.Second) // 1s just to make rps in == rps out
	}

	// init with anything <~1000, too large and even a small fraction of the original value can be too big.
	push(h1, rand.Intn(1000), rand.Intn(1000))
	push(h2, rand.Intn(1000), rand.Intn(1000))
	push(h3, rand.Intn(1000), rand.Intn(1000))

	// now update multiple times and make sure it gets to 90% within 4 steps == 12s (normally).
	//
	// 4 steps with 0.5 weight should mean only 0.5^4 => 6.25% of the original influence remains,
	// which feels pretty reasonable: after ~10 seconds (3s updates), the oldest data only has ~10% weight.
	const target = 10 + 200 + 999
	for i := 0; i < 4; i++ {
		weights, used = agg.getWeightsLocked(key, agg.snapshot())
		t.Log("used:", used, "of actual:", target)
		t.Log("weights so far:", weights)
		push(h1, 10, 10)
		push(h2, 200, 200)
		push(h3, 999, 999)
	}
	weights, used = agg.getWeightsLocked(key, agg.snapshot())
	t.Log("used:", used, "of actual:", target)
	t.Log("weights so far:", weights)

	// aggregated allowed-request values should be less than 10% off
	assert.InDeltaf(t, target, float64(used), target*0.1, "should have allowed >90%% of target rps by the 5th round") // actually ~94%
	// also check weights, they should be within 10%
	assert.InDeltaMapValues(t, map[Identity]float64{
		h1: 10 / 1209.0,  // 0.07698229407
		h2: 200 / 1209.0, // 0.1539645881
		h3: 999 / 1209.0, // 0.7690531178
	}, floaty(weights), 0.1, "should be close to true load balance")
}

// converts for testify/assert as it requires same types, not just same underlying type
func floaty[K comparable, V numeric](m map[K]V) map[K]float64 {
	out := make(map[K]float64, len(m))
	for k, v := range m {
		out[k] = float64(v)
	}
	return out
}

func TestConcurrent(t *testing.T) {
	// essentially a fuzz-test for race purposes.
	// values aren't checked, but it shouldn't panic / shouldn't race / etc.
	// low timeout + real time clock to also have wall-clock changes race with logic.
	//
	// this test frequently reaches 100% coverage all on its own (currently),
	// though it's not guaranteed and that is not the intent.
	// other tests should cover sufficiently even if this test is skipped:
	// t.Skip("skipped to check coverage")

	// should be "much" longer than expected update rate
	const (
		updateRate      = time.Millisecond // fairly arbitrary, max gap between updates
		targetDuration  = 100 * updateRate // also minimum number of updates
		numHosts        = 10
		numUpdaters     = 10           // fairly arbitrary
		updatesPerBatch = numHosts / 3 // intentionally below len(hosts) to allow some to gc normally
	)

	agg, _ := newForTest(t, configSnapshot{
		rate:      updateRate,
		maxMissed: 2,
		gcAfter:   3, // relatively low to trigger implicit gc, check coverage if changing the values

		weight: 0.1, // irrelevant but must be non-zero
	})
	agg.clock = clock.NewRealTimeSource()

	var hosts []Identity
	for i := 0; i < numHosts; i++ {
		hosts = append(hosts, Identity(fmt.Sprintf("host %d", i)))
	}
	var keys []Limit
	for i := 'a'; i <= 'z'; i++ {
		keys = append(keys, Limit(i))
	}

	do := make(chan struct{}, numUpdaters)
	var wg sync.WaitGroup

	// run some goroutines to update/read in batches
	for i := 0; i < numUpdaters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				_, ok := <-do
				if !ok {
					// chan's closed, stop
					return
				}
				num := rand.Intn(len(keys))
				updates := make(map[Limit]Requests, num)
				for i := 0; i < num; i++ {
					updates[keys[rand.Intn(len(keys))]] = Requests{
						Accepted: rand.Intn(100),
						Rejected: rand.Intn(100),
					}
				}
				host := hosts[rand.Intn(len(hosts))]

				// randomly update or read
				if rand.Intn(2) == 0 {
					agg.Update(host, updates, updateRate)
				} else {
					agg.HostWeights(host, maps.Keys(updates))
				}
			}
		}()
	}

	// run a "trigger some work occasionally" goroutine
	wg.Add(1)
	start := time.Now()
	go func() {
		defer wg.Done()

		for {
			if time.Since(start) > targetDuration {
				close(do)
				return
			}
			// sleep a random portion of the update rate
			time.Sleep(time.Duration(rand.Intn(int(updateRate))))
			// allow a random number of updates.
			// should be non-blocking to further encourage racing, when possible
			for i := rand.Intn(updatesPerBatch); i > 0; i-- {
				do <- struct{}{}
			}
		}
	}()

	assert.True(t,
		common.AwaitWaitGroup(&wg, 5*targetDuration),
		"blocked test? still waiting after %v", 5*targetDuration)
	// non-racy even if waiting failed
	t.Logf("%#v", agg.GC()) // should (usually) not be "full" + should (usually) remove some data
}

func TestSimulate(t *testing.T) {
	// Semi-fuzzy simulated sequence with fully computed values, exercising most behaviors.
	//
	// Everything about this test is sensitive to changes in behavior,
	// so if that occurs just update the values after ensuring they're reasonable.
	// Exact matches after changes are not at all important.

	updateRate := 3 * time.Second // both expected and duration fed to update
	agg, tick := newForTest(t, configSnapshot{
		// now:    , ignored
		weight:    0.75, // fairly fast adjustment, and semi-human-friendly math
		rate:      updateRate,
		maxMissed: 2.0,
		gcAfter:   10,
	})

	// keeping var == string simplifies copy/paste
	start, query := Limit("start"), Limit("query")
	all := []Limit{start, query}
	h1, h2, h3 := Identity("one"), Identity("two"), Identity("three")

	weights, used := agg.getWeightsLocked(start, agg.snapshot())
	assert.Zero(t, weights, "should have no weights")
	assert.Zero(t, used, "should have no used RPS")

	// just simplifies arg-construction
	push := func(host Identity, key Limit, accept, reject int) {
		agg.Update(host, map[Limit]Requests{
			key: {
				Accepted: accept,
				Rejected: reject,
			},
		}, time.Second)
	}

	// these tests intentionally share data and run sequentially,
	// the sub-testing is mostly to help group semantics.

	// init with Some Numbers.  updates to different keys with the same timestamp
	// can be grouped or separate, it doesn't matter - only same-key changes behavior.
	push(h1, query, 5, 5)
	push(h1, start, 5, 5)
	tick.Advance(time.Second)
	push(h2, query, 1, 1)
	push(h2, start, 1, 1)
	tick.Advance(time.Second)
	push(h3, query, 0, 0)
	push(h3, start, 1, 1)
	tick.Advance(time.Second)

	// changing new-data-weight does not affect this test because
	// zero -> nonzero keeps 100% to jump-start the initial state,
	// rather than gradually growing from zero (which would be biased
	// towards lower values during movement - not bad, but not intended).
	t.Run("initial weights at 3s", func(t *testing.T) {
		// 3s elapsed, all fresh.  h1 is "old" but within max-missed-updates so it's assumed valid still
		myweights, rps := agg.HostWeights(h1, all)

		// h1 has most of the weight.
		expectSimilar(t,
			query, 0.83,
			start, 0.71, // h3 uses more start than query, so h1 has less weight
			myweights)
		expectSimilar(t,
			query, 6.0,
			start, 7.0,
			rps)

		myweights, rps = agg.HostWeights(h2, all)
		expectSimilar(t,
			query, 0.17, // only h1 and h2 called this, so (0.83 + 0.17)==1
			start, 0.14, // h3 also had a small use, so less than query
			myweights)

		myweights, rps = agg.HostWeights(h3, all)
		expectSimilar(t,
			query, 0, // no query at all
			start, 0.14, // same num of starts as h2
			myweights)
	})
	if t.Failed() {
		return // later tests likely invalid, verify each step before moving to the next
	}

	// advance to second h2 update, skip the others.
	// no data has expired yet, but h1 is now at 5s old
	tick.Advance(2 * time.Second)
	push(h2, query, 1, 2)
	push(h2, start, 3, 3)

	// 4.5s
	t.Run("increased h2 weight at 5s", func(t *testing.T) {
		myweights, rps := agg.HostWeights(h1, all)
		// h1's weight reduces due to increased h2 usage
		expectSimilar(t,
			query, 0.78,
			start, 0.59,
			myweights)
		// overall usage increased
		expectSimilar(t,
			query, 6.00, // accepted query requests did not change (1 before, 1 after)
			start, 8.5, // but start calls from h2 increased by 2 -> 0.75 weight -> +1.5 total
			rps)

		myweights, rps = agg.HostWeights(h2, all)
		//
		expectSimilar(t,
			query, 0.22, // h3 has no query, so h1+h2=1.0 but the balance has shifted a bit towards h2
			start, 0.29, // increased over last round due to more calls
			myweights)
		// sanity check: same rps as h1 saw
		expectSimilar(t,
			query, 6.00,
			start, 8.5,
			rps)

		myweights, rps = agg.HostWeights(h3, all)
		// h3 is almost idle but 0.1 weight changes slowly
		expectSimilar(t,
			query, 0, // never sent any query requests
			start, 0.12, // decreased since last round, as relative usage is lower
			myweights)
		// sanity check: same rps as h1 saw
		expectSimilar(t,
			query, 6.00,
			start, 8.5,
			rps)
	})
	if t.Failed() {
		return // later tests likely invalid, verify each step before moving to the next
	}

	// advance to 10s.
	// this puts h1's original data beyond max-updates, so it will get degraded
	tick.Advance(5 * time.Second)
	push(h2, query, 5, 5)
	push(h2, start, 5, 5)
	push(h3, query, 5, 5)
	push(h3, start, 5, 5)

	t.Run("h1 degraded at 10s", func(t *testing.T) {
		myweights, rps := agg.HostWeights(h1, all)
		// h1's weight reduces a ton due to a long period of no updates -> retroactively treat
		// as if we got zeros, plus increased use in others
		expectSimilar(t,
			query, 0.01,
			start, 0.01,
			myweights)
		expectSimilar(t,
			query, 7.83, // 10 accepted in last round, but only 0.75 weight, 6.0 -> 7.8 due to older data
			start, 8.22, // similar
			rps)
		myweights, _ = agg.HostWeights(h2, all)
		// h2 has slightly over half weight due to greater historical use
		expectSimilar(t,
			query, 0.52,
			start, 0.53, // still more starts than queries
			myweights)
		myweights, _ = agg.HostWeights(h3, all)
		// h2 slightly less, due to low historical + some used by h1
		expectSimilar(t,
			// this looks odd, but it's higher due to lower historical *total* queries,
			// leading to a somewhat counter-intuitive:
			// - smaller numerator (lower calls by this host)
			// - smaller denominator (lower total calls)
			// - higher final value (smaller denominator has greater influence)
			query, 0.47,
			start, 0.46,
			myweights)
	})
	if t.Failed() {
		return // later tests likely invalid, verify each step before moving to the next
	}

	// update everything, should flatten compared to previous round
	tick.Advance(1 * time.Second)
	push(h1, query, 5, 5)
	push(h1, start, 5, 5)
	push(h2, query, 5, 5)
	push(h2, start, 5, 5)
	push(h3, query, 5, 5)
	push(h3, start, 5, 5)

	t.Run("all equal at 11s is relatively flatter than before", func(t *testing.T) {
		myweights, rps := agg.HostWeights(h1, all)
		// historically lower weight = current lower weight, but a big jump from 0.01 before.
		//
		// this is likely a faster shift than we want in practice, as it'll make allowed-request
		// behavior quite jumpy, which is why the initial weight is likely to be around 0.5.
		expectSimilar(t,
			query, 0.29,
			start, 0.28,
			myweights)
		// used is adjusting towards 15
		expectSimilar(t,
			query, 13.21,
			start, 13.31,
			rps)
		// else flattening towards 0.33
		myweights, _ = agg.HostWeights(h2, all)
		expectSimilar(t,
			query, 0.36,
			start, 0.36,
			myweights)
		myweights, _ = agg.HostWeights(h3, all)
		expectSimilar(t,
			query, 0.35,
			start, 0.35,
			myweights)
	})
	if t.Failed() {
		return // later tests likely invalid, verify each step before moving to the next
	}

	// do that again, should flatten further
	tick.Advance(1 * time.Second)
	push(h1, query, 5, 5)
	push(h1, start, 5, 5)
	push(h2, query, 5, 5)
	push(h2, start, 5, 5)
	push(h3, query, 5, 5)
	push(h3, start, 5, 5)

	t.Run("all equal at 12s is even flatter", func(t *testing.T) {
		myweights, rps := agg.HostWeights(h1, all)
		// still slightly below the others but it hardly matters now
		expectSimilar(t,
			query, 0.32,
			start, 0.32,
			myweights)
		// quite close to 15
		expectSimilar(t,
			query, 14.55,
			start, 14.58,
			rps)
		// else flattening towards 0.33
		myweights, _ = agg.HostWeights(h2, all)
		expectSimilar(t,
			query, 0.34,
			start, 0.34,
			myweights)
		myweights, _ = agg.HostWeights(h3, all)
		expectSimilar(t,
			query, 0.34,
			start, 0.34,
			myweights)
	})
	if t.Failed() {
		return // later tests likely invalid, verify each step before moving to the next
	}

	// make everything fully expired so the "from expired data" branch is exercised too
	tick.Advance(time.Hour)
	// push anything (need all keys to get same HostWeight keys for `expectSimilar`)
	push(h1, query, 5, 5)
	push(h1, start, 5, 5)
	push(h2, query, 5, 5)
	push(h2, start, 5, 5)
	push(h3, query, 5, 5)
	push(h3, start, 5, 5)

	t.Run("updating expired data acts like deleted data", func(t *testing.T) {
		// should leap to exact values, not weighted-from-zero.
		myweights, rps := agg.HostWeights(h1, all)
		// still slightly below the others but it hardly matters now
		expectSimilar(t,
			query, 0.333,
			start, 0.333,
			myweights)
		// note: exactly 15. if weighting from or very near zero, would be: (0*0.25 + 15*0.75) == 11.25
		expectSimilar(t,
			query, 15,
			start, 15,
			rps)
		// else flattening towards 0.33
		myweights, _ = agg.HostWeights(h2, all)
		expectSimilar(t,
			query, 0.333,
			start, 0.333,
			myweights)
		myweights, _ = agg.HostWeights(h3, all)
		expectSimilar(t,
			query, 0.333,
			start, 0.333,
			myweights)
	})
}

func expectSimilar[V ~float64](
	t *testing.T,
	k1 Limit, v1 float64, k2 Limit, v2 float64,
	actual map[Limit]V) {
	t.Helper() // report caller's line as the logger, not this expect-er

	if !assert.InDeltaMapValues(t, map[Limit]float64{
		k1: v1,
		k2: v2,
	}, floaty(actual), 0.01) {
		var zeroV V // for %T convenience
		t.Logf("verify the %T values, and if it's correct just update args to:\n"+
			"\t%v, %0.2f,\n"+
			"\t%v, %0.2f,", zeroV, k1, actual[k1], k2, actual[k2])
	}
}

// fairly fuzzy but somewhat representative of expected use:
// benchmark "update a bunch of keys and get my load" requests, and accumulate data across all iterations.
// this is roughly what a server update operation will look like.
// there are intentionally overlaps in both hosts and keys to get multiple hosts per key and updates to existing data.
//
// while this obviously has major caveats and fake costs, the benchmark regularly runs several thousand iterations
// before it settles, filling a moderate amount of data along the way (e.g. consistently all 10k keys, and >200k records in memory).
//
// general results imply:
//   - host-weight reading takes a few times longer than updating, but on my laptop the
//     bench takes about 3-4ms per loop for the current 100/1000/10000 values, without mutexes.
//   - this is relatively CPU costly, but even if our largest cluster checks in every 3s to a *single*
//     host, this should work out with some room to spare... and it will be far more spread
//     out in practice due to the number of history hosts involved.
//   - an attempt to cache weight calculations per key until invalidated saved about 1/2 the cpu
//     with a perfect cache.
//   - seems not worth the complexity, but it'd be fairly easy to add if we change our minds
//
// since there is only a coarse mutex in the implementation itself, the whole aggregator forces itself
// to be processed serially.  if this turns out to be too costly, the easy fix is probably
// to shard the keys to e.g. 8 separate aggregators that can be processed concurrently.
//
// mutexes could be added to each Limit, but that could end up costing more in memory delays
// than the fine-grained locking gains us in concurrency flexibility.  benchmark first!
func BenchmarkNormalUse(b *testing.B) {
	updateRate := 3 * time.Second
	agg, _ := newForTest(b, defaultConfig(updateRate))
	// intentionally using real clock source, time-gathering cost is relevant to benchmark
	agg.clock = clock.NewRealTimeSource()

	// aiming for vaguely realistic-ish values
	const (
		hosts       = 100   // num of hosts in pool, if rounds > hosts there will be duplicates (intentional)
		keysPerHost = 1000  // more domains than hosts, multiple limits per domain.
		globalKeys  = 10000 // each host gets ~10% of requests (not consistent across calls tho)
		requests    = 1000  // accept/reject counts
	)

	// rand and string-formatting is a non-trivial amount of load,
	// so this is prepared up-front so it can be excluded.
	type round struct {
		host    Identity
		load    map[Limit]Requests
		keys    []Limit
		elapsed time.Duration
	}

	rounds := make([]round, 0, b.N)
	for i := 0; i < b.N; i++ {
		keys := rand.Intn(keysPerHost)
		reqs := make(map[Limit]Requests, keys)
		for len(reqs) < keys {
			key := Limit(fmt.Sprintf("key %d", rand.Intn(globalKeys)))
			if _, ok := reqs[key]; ok {
				// duplicate, just try again.
				// dups wouldn't really invalidate the benchmark,
				// but it's easy to avoid and should make more stable results.
				continue
			}

			reqs[key] = Requests{
				// values don't really matter but there's no need for them all to use the same value
				Accepted: rand.Intn(requests),
				Rejected: rand.Intn(requests),
			}
		}
		rounds = append(rounds, round{
			host:    Identity(fmt.Sprintf("host %d", rand.Intn(hosts))),
			load:    reqs,
			keys:    maps.Keys(reqs),
			elapsed: time.Duration(rand.Int63n(updateRate.Nanoseconds())),
		})
	}

	b.ResetTimer()
	b.ReportAllocs()

	sawnonzero := 0
	for _, r := range rounds { // == b.N times
		agg.Update(r.host, r.load, r.elapsed)
		var unused map[Limit]PerSecond // ensure non-error second return for test safety
		weights, unused := agg.HostWeights(r.host, r.keys)
		_ = unused // ignore unused rps
		if len(weights) > 0 {
			// 	wrote data and later read it out, benchmark is likely functional
			sawnonzero++
		}
	}
	b.StopTimer() // not benchmarking gc, just using it to show sanity-check metrics

	b.Log("N was:", b.N)
	b.Log("gc metrics:", agg.GC()) // shows how many keys / hosts / etc we stored for manual validation

	// sanity check, as "all zero" should mean something like "always looking at nonexistent keys" / bad benchmark code.
	b.Log("nonzero results:", sawnonzero)
	assert.True(b, b.N < 500 || sawnonzero > 0, "no non-zero result found on a large enough benchmark, likely not benchmarking anything useful")
}

func TestHashmapIterationIsRandom(t *testing.T) {
	// is partial hashmap iteration reliably random each time so we can statistically ensure coverage,
	// or is it relatively fixed per map?
	// language spec is somewhat vague on the details here, so let's check.
	//
	// verdict: yes!  looks random each time.
	// so we can abuse it for an amortized / interruptible cleanup tool.

	const keys = 100
	m := map[int]struct{}{}
	for i := 0; i < keys; i++ {
		m[i] = struct{}{}
	}

	allObserved := make(map[int]struct{}, len(m))
	var orderObserved [][]int
	singlePass := func() {
		for i := 0; i < keys; i++ {
			observed := make([]int, 0, keys/10)
			for k := range m {
				allObserved[k] = struct{}{}
				observed = append(observed, k)
				if len(observed) == cap(observed) {
					break // interrupt part way through
				}
			}
			orderObserved = append(orderObserved, observed)
		}
	}

	// keep trying up to 10x to make it sufficiently-unlikely that randomness will fail.
	// with only a single round it fails like 5% of the time, which is reasonable behavior
	// but too much noise to allow to fail the test suite.
	for i := 0; i < 10 && len(allObserved) < keys; i++ {
		if i > 0 {
			t.Logf("insufficient keys observed (%d), trying again in round %d", len(allObserved), i+1)
		}
		singlePass()
	}

	// complain if it still hasn't observed all keys
	if !assert.Len(t, allObserved, keys) {
		// super noisy when successful, so only log when failing
		for idx, pass := range orderObserved {
			t.Log("Pass", idx)
			for _, keys := range pass {
				t.Log("\t", keys)
			}
		}
	}
}
