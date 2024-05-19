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
		detailed = false
	)

	debug := func(t *testing.T, format string, args ...interface{}) {
		// do not log normally
	}
	if detailed {
		debug = func(t *testing.T, format string, args ...interface{}) {
			t.Logf(format, args...)
		}
	}
	_ = debug

	for p := 0; p < numTests; p++ {
		t.Run(fmt.Sprintf("goroutine %v", p), func(t *testing.T) {
			// parallel saves a fair bit of time but introduces a lot of CPU contention
			// and that leads to a moderate amount of flakiness.
			//
			// unfortunately not recommended here.
			// t.Parallel()

			seed := time.Now().UnixNano()
			// seed = 1716086576261550000 // override seed to test a failing scenario
			rng := rand.New(rand.NewSource(seed))
			t.Logf("rand seed: %v", seed)

			burst := rng.Intn(maxBurst)               // zero is uninteresting but allowed
			limit := rate.Every(granularity * events) // refresh after each full round
			t.Logf("limit: %v, burst: %v", granularity, burst)
			now := time.Now()
			ts := NewMockedTimeSourceAt(now)

			actual := rate.NewLimiter(limit, burst)
			wrapped := NewRatelimiter(limit, burst)
			mocked := NewMockRatelimiter(ts, limit, burst)

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

			ticker := time.NewTicker(granularity)
			for round := range calls {
				for event := range calls[round] {
					<-ticker.C
					ts.Advance(granularity)
					debug(t, "Tokens before round, real: %0.2f, wrapped: %0.2f, mocked: %0.2f", actual.Tokens(), wrapped.Tokens(), mocked.Tokens())

					if calls[round][event] == 1 {
						const options = 4
						switch rng.Intn(options) % options {
						case 0:
							// call Allow on everything
							a, w, m := actual.Allow(), wrapped.Allow(), mocked.Allow()
							t.Logf("round[%v][%v] Allow, real limiter: %v, wrapped real: %v, mocked: %v", round, event, a, w, m)
							assert.True(t, a == w && w == m, "ratelimiters disagree on round[%v][%v] Allow, real limiter: %v, wrapped real: %v, mocked: %v", round, event, a, w, m)
						case 1:
							// call Reserve on everything
							_a, _w, _m := actual.Reserve(), wrapped.Reserve(), mocked.Reserve()
							a, w, m := _a.OK() && _a.Delay() == 0, _w.Allow(), _m.Allow()
							t.Logf("round[%v][%v] Reserve, real limiter: %v (ok: %v, delay: %v), wrapped real: %v, mocked: %v", round, event, a, _a.OK(), _a.Delay().Round(time.Millisecond), w, m)
							assert.True(t, a == w && w == m, "ratelimiters disagree on round[%v][%v] Reserve, real limiter: %v, wrapped real: %v, mocked: %v", round, event, a, w, m)
						case 2:
							// Try a brief Wait on everything.
							//
							// ctx must expire before the next event or things can collide, but otherwise the timeout should not matter.
							ctx, cancel := context.WithTimeout(context.Background(), granularity/10)

							a, w, m := false, false, false
							var g errgroup.Group
							g.Go(func() error {
								a = actual.Wait(ctx) == nil
								return nil
							})
							g.Go(func() error {
								w = wrapped.Wait(ctx) == nil
								return nil
							})
							g.Go(func() error {
								m = mocked.Wait(ctx) == nil
								return nil
							})
							_ = g.Wait()

							t.Logf("round[%v][%v] Wait, real limiter: %v, wrapped real: %v, mocked: %v", round, event, a, w, m)
							assert.True(t, a == w && w == m, "ratelimiters disagree on round[%v][%v] Wait, real limiter: %v, wrapped real: %v, mocked: %v", round, event, a, w, m)

							cancel()
						case 3:
							// call Reserve on everything, and cancel half of them
							_a, _w, _m := actual.Reserve(), wrapped.Reserve(), mocked.Reserve()
							if rng.Intn(2)%2 == 0 {
								_a.Cancel()
								_w.Used(false)
								_m.Used(false)
								t.Logf("round[%v][%v] Reserve, canceled", round, event)
							} else {
								a, w, m := _a.OK() && _a.Delay() == 0, _w.Allow(), _m.Allow()
								t.Logf("round[%v][%v] Reserve, real limiter: %v (ok: %v, delay: %v), wrapped real: %v, mocked: %v", round, event, a, _a.OK(), _a.Delay().Round(time.Microsecond), w, m)
								assert.True(t, a == w && w == m, "ratelimiters disagree on round[%v][%v] Reserve, real limiter: %v, wrapped real: %v, mocked: %v", round, event, a, w, m)
								_w.Used(true)
								_m.Used(true)
							}
						}
					}
					debug(t, "Tokens after round, real: %0.2f, wrapped: %0.2f, mocked: %0.2f", actual.Tokens(), wrapped.Tokens(), mocked.Tokens())
				}
			}
		})
	}
}
