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

package clock

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

func TestAgainstRealRatelimit(t *testing.T) {
	t.Cleanup(func() {
		if t.Failed() {
			/*
				Is this test becoming too slow or too noisy?

				The easiest "likely to actually work" fix is probably going to
				require detecting excessive lag, and retrying instead of failing.

				That could miss some real flaws if they are racy by nature, but
				seems like it might be good enough.
			*/
			t.Logf("---- CAUTION ----")
			t.Logf("these tests are randomized by design, a random failure may be a real flaw!")
			t.Logf("please replay with the failing seed and check detailed output to see what the behavior should be.")
			t.Logf("try enabling the 'detailedWhenVerbose' var to increase logging verbosity.")
			t.Logf("")
			t.Logf("if you are intentionally making changes, be sure to run a few hundred rounds to make sure your changes are stable")
			t.Logf("---- CAUTION ----")
		}
	})
	// The mock ratelimiter should behave the same as a real ratelimiter in all cases.
	//
	// So fuzz test it: throw random calls at non-overlapping times that are a bit
	// spaced out (to prevent noise due to busy CPUs) and make sure all impls agree.
	//
	// If a test fails, please verify by hand with the failing seed.
	// Enabling debug logs can help show what's happening and what *should* be happening,
	// though there is a chance it's just due to CPU contention.
	const (
		maxBurst = 10

		// amount of time between each conceptual "tick" of the test's clocks, both real and fake.
		// 10ms seems fairly good when not parallel.
		granularity = 10 * time.Millisecond
		// number of [granularity] events to trigger per round,
		// and also the number of [granularity] periods before a token is added.
		//
		// each test takes at least granularity*events*rounds, so keep it kinda small.
		events = 10
		// number of rounds of "N events" to try
		rounds = 2

		// number of tests to generate.
		// running these in parallel works and saves time, but leads to output that's hard to read,
		// and too many can cause noise due to cpu contention.
		// so keep it smallish, and while debugging try lowering + sequential.
		numTests = 10

		// set to true to log in more detail, which is generally too noisy
		detailedWhenVerbose = false
	)

	debug := func(t *testing.T, format string, args ...interface{}) {
		t.Logf(format, args...)
	}
	if testing.Verbose() && !detailedWhenVerbose {
		// normally, this is a ton of text that's just noise.
		// so for bulk verbose modes, skip the debug logs unless force-enabled.
		//
		// regrettably this makes failures less useful when verbose, but the
		// failure logs do at least inform about this behavior.
		debug = func(t *testing.T, format string, args ...interface{}) {
			// do nothing
		}
	}
	_ = debug

	check := func(t *testing.T, what string, round, event int, actual, wrapped, mocked bool, compressed *bool) {
		t.Helper()
		if compressed != nil {
			t.Logf("round[%v][%v] %v, actual limiter: %v, wrapped: %v, mocked: %v, compressed: %v", round, event, what, actual, wrapped, mocked, *compressed)
			assert.True(t, actual == wrapped && wrapped == mocked && mocked == *compressed, "ratelimiters disagree")
		} else {
			t.Logf("round[%v][%v] %v, actual limiter: %v, wrapped: %v, mocked: %v", round, event, what, actual, wrapped, mocked)
			assert.True(t, actual == wrapped && wrapped == mocked, "ratelimiters disagree")
		}
	}

	for p := 0; p < numTests; p++ {
		t.Run(fmt.Sprintf("goroutine %v", p), func(t *testing.T) {
			// parallel saves a fair bit of time but introduces a lot of CPU contention
			// and that leads to a moderate amount of flakiness.
			//
			// unfortunately not recommended here.
			// t.Parallel()

			seed := time.Now().UnixNano()
			// seed = 1716151675980912000 // override seed to test a failing scenario
			rng := rand.New(rand.NewSource(seed))
			t.Logf("rand seed: %v", seed)

			burst := rng.Intn(maxBurst)               // zero is uninteresting but allowed
			limit := rate.Every(granularity * events) // refresh after each full round
			t.Logf("limit: %v, burst: %v", granularity, burst)
			now := time.Now()

			ts := NewMockedTimeSourceAt(now)
			compressedTS := NewMockedTimeSourceAt(now)

			actual := rate.NewLimiter(limit, burst)
			wrapped := NewRatelimiter(limit, burst)
			mocked := NewMockRatelimiter(ts, limit, burst)
			compressed := NewMockRatelimiter(compressedTS, limit, burst)

			// generate some non-colliding "1 == perform an Allow call" rounds.
			calls := [rounds][events]int{} // using int just to print better than bool
			for i := 0; i < len(calls[0]); i++ {
				// pick one to set to true, or none (1 chance for none)
				set := rng.Intn(len(calls) + 1)
				if set < len(calls) {
					calls[set][i] = 1
				}
			}
			t.Log("round setup:")
			for _, round := range calls {
				t.Logf("\t%v", round)
			}
			/*
				rounds look like:
					[1 0 0 1 0 ...]
					[0 1 1 0 0 ...]
				which means that the first round will:
					- call
					- do nothing
					- do nothing
					- call ...
				and by the end of that first array it'll reach the end of
				the rate.Every time, and will begin recovering tokens during
				the next array's runtime.
				so the second round will:
					- refresh 1 token while waiting
					- consume 1 call
					- consume 1 call
					- refresh 1
					- wait, no tokens refresh because none were consumed here last round

				the `1`s do not overlap to avoid triggering a race between the
				old call being refreshed and the new call consuming, as these
				make for very flaky tests unless time is mocked.
			*/

			// record "to be executed" closures on the time-compressed ratelimiter too,
			// so it can be checked at a sped up rate.
			compressedReplay := [rounds][events]func(t *testing.T){}

			ticker := time.NewTicker(granularity)
			for round := range calls {
				round := round // for closure
				for event := range calls[round] {
					event := event // for closure
					<-ticker.C
					ts.Advance(granularity)
					debug(t, "Tokens before round, real: %0.2f, wrapped: %0.2f, mocked: %0.2f", actual.Tokens(), wrapped.Tokens(), mocked.Tokens())

					if calls[round][event] == 1 {
						const options = 4
						switch rng.Intn(options) % options {
						case 0:
							// call Allow on everything
							a, w, m := actual.Allow(), wrapped.Allow(), mocked.Allow()
							check(t, "Allow", round, event, a, w, m, nil)
							compressedReplay[round][event] = func(t *testing.T) {
								c := compressed.Allow()
								check(t, "Allow (Compressed)", round, event, a, w, m, &c)
							}
						case 1:
							// call Reserve on everything
							_a, _w, _m := actual.Reserve(), wrapped.Reserve(), mocked.Reserve()
							a, w, m := _a.OK() && _a.Delay() == 0, _w.Allow(), _m.Allow()
							check(t, "Reserve", round, event, a, w, m, nil)
							compressedReplay[round][event] = func(t *testing.T) {
								c := compressed.Reserve().Allow()
								check(t, "Reserve (Compressed)", round, event, a, w, m, &c)
							}
						case 2:
							// Try a brief Wait on everything.
							//
							// ctx must expire:
							//   - after Wait performs its internal checks
							//   - before the next event would occur
							//
							// so:
							//   - don't make it *too* short or the deadline may be passed before Wait sleeps
							//     (a timeout of 1ms has done this a few % of the time, it can happen to you too!)
							//   - don't make it *too* long or it may conflict with later events
							ctx, cancel := context.WithTimeout(context.Background(), granularity/2)

							a, w, m := false, false, false
							var g errgroup.Group
							g.Go(func() error {
								started := time.Now()
								_a := actual.Wait(ctx)
								debug(t, "Wait elapsed: %v, actual err: %v", time.Since(started).Round(time.Millisecond), _a)
								a = _a == nil
								return nil
							})
							g.Go(func() error {
								started := time.Now()
								_w := wrapped.Wait(ctx)
								debug(t, "Wait elapsed: %v, wrapped err: %v", time.Since(started).Round(time.Millisecond), _w)
								w = _w == nil
								return nil
							})
							g.Go(func() error {
								started := time.Now()
								_m := mocked.Wait(ctx)
								debug(t, "Wait elapsed: %v, mocked err: %v", time.Since(started).Round(time.Millisecond), _m)
								m = _m == nil
								return nil
							})
							_ = g.Wait()

							check(t, "Wait", round, event, a, w, m, nil)
							compressedReplay[round][event] = func(t *testing.T) {
								// hmm.  maybe we need a time-mocked context too.
								ctx, cancel := context.WithTimeout(context.Background(), granularity/1000)
								started := time.Now()
								_c := compressed.Wait(ctx)
								debug(t, "Wait elapsed: %v, compressed err: %v", time.Since(started).Round(time.Millisecond), _c)
								c := _c == nil
								check(t, "Wait (Compressed)", round, event, a, w, m, &c)
								cancel()
							}
							cancel()
						case 3:
							// call Reserve on everything, and cancel half of them
							_a, _w, _m := actual.Reserve(), wrapped.Reserve(), mocked.Reserve()
							if rng.Intn(2)%2 == 0 {
								_a.Cancel()
								_w.Used(false)
								_m.Used(false)
								t.Logf("round[%v][%v] ReserveWithCancel, canceled", round, event)
								compressedReplay[round][event] = func(t *testing.T) {
									compressed.Reserve().Used(false)
									t.Logf("compressed round[%v][%v] ReserveWithCancel, canceled", round, event)
								}
							} else {
								a, w, m := _a.OK() && _a.Delay() == 0, _w.Allow(), _m.Allow()
								check(t, "ReserveWithCancel", round, event, a, w, m, nil)
								if !a {
									// must cancel, or the not-yet-available reservation affects future calls.
									// that's valid if you intend to wait for your reservation, but the wrapper
									// does not allow that.
									_a.Cancel()
								}
								_w.Used(w)
								_m.Used(m)
								compressedReplay[round][event] = func(t *testing.T) {
									_c := compressed.Reserve()
									c := _c.Allow()
									_c.Used(c)
									check(t, "ReserveWithCancel (Compressed)", round, event, a, w, m, &c)
								}
							}
						}
					}
					debug(t, "Tokens after round, real: %0.2f, wrapped: %0.2f, mocked: %0.2f", actual.Tokens(), wrapped.Tokens(), mocked.Tokens())
				}
			}
			t.Run("compressed time", func(t *testing.T) {
				// and now replay the compressed ratelimiter and make sure it matches too,
				// as time-compressed must behave the same as real-time.
				//
				// this is primarily intended to detect cases where real-time is accidentally used,
				// as ~zero time actually passes, which is quite different from the above tests.
				//
				// it's not perfect, but it does eventually notice such bugs, as you can see by
				// changing literally any timesource.Now() calls into time.Now() and running
				// tests a few times.
				// depending on the change it might not notice in most tests, but eventually a
				// problematic combination is triggered and can be replayed with the logged seed.
				for round := range calls {
					for event := range calls[round] {
						compressedTS.Advance(granularity)
						debug(t, "Tokens before compressed round: %0.2f", compressed.Tokens())
						replay := compressedReplay[round][event]
						if replay != nil {
							replay(t)
						}
						debug(t, "Tokens after compressed round: %0.2f", compressed.Tokens())
					}
				}
			})
		})
	}
}
