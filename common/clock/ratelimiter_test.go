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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

/*
Results on my machines:

	M1 mac:
		Allow:
			Sequential goes from 75ns to 100ns
			Parallel goes from 200ns to 250ns
		Reserve and 1/3rd cancel:
			Sequential goes from 175ns to 220ns
			Parallel goes from 300ns to 550ns (but real limiters have badly flawed behavior)
		Wait:
	Linux cloud machine:
		Allow:
		Reserve:
		Wait:
*/
func BenchmarkLimiter(b *testing.B) {
	type runType func(b *testing.B, each func(int) bool)

	// runs a callback in a sequential benchmark, and tracks allowed/denied metrics
	var runSerial runType = func(b *testing.B, each func(int) bool) {
		allowed, denied := 0, 0
		for i := 0; i < b.N; i++ {
			if each(i) {
				allowed++
			} else {
				denied++
			}
		}
		b.Logf("allowed: %v, denied: %v", allowed, denied)
	}
	// runs a callback in a parallel benchmark, and tracks allowed/denied metrics
	var runParallel runType = func(b *testing.B, each func(int) bool) {
		var allowed, denied atomic.Int64
		b.RunParallel(func(pb *testing.PB) {
			n := 0
			for pb.Next() {
				if each(n) {
					allowed.Inc()
				} else {
					denied.Inc()
				}
				n++
			}
		})
		b.Logf("allowed: %v, denied: %v", allowed.Load(), denied.Load())
	}

	both := map[string]runType{
		"serial":   runSerial,
		"parallel": runParallel,
	}

	// Allow is a significant amount of our ratelimiter usage,
	// so this should probably take top priority.
	b.Run("allow", func(b *testing.B) {
		for name, runner := range both {
			b.Run(name, func(b *testing.B) {
				b.Run("real", func(b *testing.B) {
					// very fast to jump back and forth, rather than slamming into "deny"
					rl := rate.NewLimiter(rate.Every(time.Microsecond), 1000)
					runner(b, func(i int) bool {
						return rl.Allow()
					})
				})
				b.Run("wrapped", func(b *testing.B) {
					rl := NewRatelimiter(rate.Every(time.Microsecond), 1000)
					runner(b, func(i int) bool {
						return rl.Allow()
					})
				})
				b.Run("mocked timesource", func(b *testing.B) {
					ts := NewMockedTimeSource()
					rl := NewMockRatelimiter(ts, rate.Every(time.Microsecond), 1000)
					runner(b, func(i int) bool {
						// adjusted by eye, to try to very roughly match the above values for the final runs.
						// probably needs to be tweaked per machine.
						ts.Advance(time.Microsecond / 5)
						return rl.Allow()
					})
				})
			})
		}
	})

	// Reserve is used when tiering ratelimiters, which is only done in a few
	// primarily-user-facing limits that aren't super high perf need...
	// ... BUT Reserve is used in the wrapper's Wait, so this serves to separate
	// out its cost from the Wait benchmarks.
	cancelNth := 3
	b.Run(fmt.Sprintf("reserve-canceling-every-%v", cancelNth), func(b *testing.B) {
		for name, runner := range both {
			b.Run(name, func(b *testing.B) {
				// CAUTION: in parallel, these real ratelimiters run quickly, but note how many
				// requests are allowed.  FAR more than should be allowed, by a few
				// orders of magnitude.
				//
				// example:
				//   BenchmarkLimiter/reserve-canceling-every-3/serial/real-8         6572617	     174.3 ns/op
				//     ratelimiter_test.go:62: allowed: 1196, denied: 6571421
				//   BenchmarkLimiter/reserve-canceling-every-3/parallel/real-8       4066992	     296.7 ns/op
				//     ratelimiter_test.go:77: allowed: 3973266, denied: 93726
				//                                      ^ ~4,000x too many!
				//
				// this occurs because time is not monotonic between goroutines, so
				// time is being moved forward and backward repeatedly, and this incorrectly
				// returns tokens.  you can see this if you rewind/advance time by hand when
				// calling AllowN(at, 1) and watch what it does to `*rate.Limiter.Tokens()`.
				//
				// this is one of the reasons this wrapper was built.
				// I'm honestly surprised that *rate.Limiter handles non-monotonic time like this.
				b.Run("real", func(b *testing.B) {
					rl := rate.NewLimiter(rate.Every(time.Microsecond), 1000)
					runner(b, func(i int) (allowed bool) {
						r := rl.Reserve()
						allowed = r.OK() && r.Delay() == 0
						if i%cancelNth == 0 {
							r.Cancel()
							return
						}
						return
					})
				})
				b.Run("real pinned time", func(b *testing.B) {
					rl := rate.NewLimiter(rate.Every(time.Microsecond), 1000)
					runner(b, func(i int) (allowed bool) {
						// expected to be faster as the limiter's time does not
						// advance as often, and "now" is only gathered once.
						//
						// but this is not possible to do correctly with concurrent use,
						// so it's purely synthetic and serves as a lower bound only.
						now := time.Now()
						r := rl.ReserveN(now, 1)
						allowed = r.OK() && r.DelayFrom(now) == 0
						if i%cancelNth == 0 {
							r.CancelAt(now)
							return
						}
						return
					})
				})

				// calls Reserve on the wrapped limiter and sometimes cancels, sequentially or in parallel
				runWrapped := func(b *testing.B, rl Ratelimiter, runner runType, advance func()) {
					runner(b, func(i int) (allowed bool) {
						if advance != nil {
							advance() // for mock time
						}
						r := rl.Reserve()
						allowed = r.Allow()
						canceled := i%cancelNth == 0
						r.Used(!canceled)
						return
					})
				}

				// notice that the allowed/denied metrics are roughly the same whether parallel or not.
				// this is how it should be, as the benchmark run times are similar.
				b.Run("wrapped", func(b *testing.B) {
					rl := NewRatelimiter(rate.Every(time.Microsecond), 1000)
					runWrapped(b, rl, runner, nil)
				})
				b.Run("mocked timesource", func(b *testing.B) {
					ts := NewMockedTimeSource()
					rl := NewMockRatelimiter(ts, rate.Every(time.Microsecond), 1000)
					runWrapped(b, rl, runner, func() {
						ts.Advance(time.Microsecond / 5)
					})
				})
			})
		}
	})

	// Wait is implemented quite differently between real and wrapped.
	// For the most part our use is in batches, where queueing is expected.
	// This makes benchmarks kinda synthetic.  Queueing won't run any faster than the intended rate.
	//
	// Still, the cost to allow a request should be similar.
	//
	// Measured benchmark time is checked, and if (non-mocked) they run too quickly
	// or too slowly it'll fail the benchmark.
	b.Run("wait", func(b *testing.B) {
		for name, runner := range both {
			b.Run(name, func(b *testing.B) {
				durations := []time.Duration{
					// test with an insanely fast refresh, to see maximum throughput with "zero" waiting.
					// intentionally avoiding 1 nanosecond in case it gets special-cased, like 0 and math.MaxInt64.
					2 * time.Nanosecond,
					// test with a rate that should contend and wait heavily.
					// this SHOULD be how long the per-iteration time is (when not mocked)... but read comments below.
					time.Microsecond,
				}
				for _, dur := range durations {
					b.Run(fmt.Sprintf("%v rate", dur), func(b *testing.B) {
						limit := rate.Every(dur)
						// set a wait timeout long enough to allow ~all attempts.
						// benchmark seems to target below 1s, above that should be fine and tests should not take this long.
						timeout := 5 * time.Second

						b.Run("real", func(b *testing.B) {
							// CAUTION: in parallel, these real ratelimiters allow too many requests through.
							//
							// when run sequentially with 1µs, this behaves the way you would expect: it takes ~1µs per loop.
							// when run in parallel, this takes <500ms per loop!
							//
							// this is likely for the same reason as Reserve's extremely incorrect behavior:
							// time between goroutines is not monotonic when viewed globally, so it's jumping
							// back and forth and allowing something near 2x more than it should.
							rl := rate.NewLimiter(limit, 1000)
							runner(b, func(i int) bool {
								ctx, cancel := context.WithTimeout(context.Background(), timeout)
								defer cancel()
								return rl.Wait(ctx) == nil
							})
						})
						b.Run("wrapped", func(b *testing.B) {
							rl := NewRatelimiter(limit, 1000)
							runner(b, func(i int) bool {
								ctx, cancel := context.WithTimeout(context.Background(), timeout)
								defer cancel()
								return rl.Wait(ctx) == nil
							})
						})
						b.Run("mocked timesource", func(b *testing.B) {
							ts := NewMockedTimeSource()
							rl := NewMockRatelimiter(ts, limit, 1000)
							runner(b, func(i int) bool {
								ts.Advance(dur) // not entirely sure what a reasonable value is, but this seems... fine?
								ctx := &timesourceContext{
									ts:       ts,
									deadline: time.Now().Add(timeout),
								}
								return rl.Wait(ctx) == nil
							})
						})
					})
				}
			})
		}
	})
}

/*
Results on my machine: quite small impact when contended.

	goos: darwin
	goarch: arm64
	pkg: github.com/uber/cadence/common/clock
	BenchmarkLimiterParallel/real-8         			5924394	       205.6 ns/op
		ratelimiter_test.go:130: allowed: 5158558, denied: 765836
	BenchmarkLimiterParallel/mocked-8       			4757856	       264.6 ns/op
		ratelimiter_test.go:145: allowed: 1260049, denied: 3497807
	BenchmarkLimiterParallel/mocked_timesource-8        6227589	       192.4 ns/op
		ratelimiter_test.go:164: allowed: 1246517, denied: 4981072
*/
func BenchmarkLimiterParallel(b *testing.B) {
	run := func(b *testing.B, each func() bool) {
		var allowed, denied atomic.Int64
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if each() {
					allowed.Inc()
				} else {
					denied.Inc()
				}
			}
		})
		b.Logf("allowed: %v, denied: %v", allowed.Load(), denied.Load())
	}
	b.Run("real", func(b *testing.B) {
		rl := rate.NewLimiter(rate.Every(time.Microsecond), 1000)
		run(b, rl.Allow)
	})
	b.Run("mocked", func(b *testing.B) {
		rl := NewRatelimiter(rate.Every(time.Microsecond), 1000)
		run(b, rl.Allow)
	})
	b.Run("mocked timesource", func(b *testing.B) {
		ts := NewMockedTimeSource()
		rl := NewMockRatelimiter(ts, rate.Every(time.Microsecond), 1000)
		run(b, func() bool {
			ts.Advance(time.Microsecond / 2)
			return rl.Allow()
		})
	})
}

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
		// sadly even 10ms has some excessive latency a few % of the time, particularly during the Wait check (10ms just to start a goroutine!).
		granularity = 20 * time.Millisecond
		// number of [granularity] events to trigger per round,
		// and also the number of [granularity] periods before a token is added.
		//
		// each test takes at least granularity*events*rounds, so keep it kinda small.
		events = 10
		// number of rounds of "N events" to try
		rounds = 2

		// keep running tests until this amount of time has passed, or a failure has occurred.
		// this could be a static count of tests to run, but a target duration
		// is a bit less fiddly.  "many attempts" is the goal, not any specific number.
		//
		// alternatively, tests have a deadline, this could run until near that deadline.
		// but right now that's not fine-tuned per package, so it's quite long.
		testDuration = 4 * time.Second // mostly stays under 5s, feels reasonable
	)

	// normally, this produces a ton of text that's just noise during successful runs.
	// so for bulk verbose modes, like `make test` does, skip the debug logs unless force-enabled.
	//
	// regrettably this makes failures less useful when verbose, but the
	// failure logs do at least inform about this behavior.
	lowVerbosity := testing.Verbose()

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
			if lowVerbosity {
				t.Logf("try setting `lowVerbosity` to false to verbosely log these tests at all times.")
			}
			t.Logf("")
			t.Logf("if you are intentionally making changes, be sure to run a few hundred rounds to make sure your changes are stable")
			t.Logf("---- CAUTION ----")
		}
	})

	debug := func(t *testing.T, format string, args ...interface{}) {
		t.Logf(format, args...)
	}
	if lowVerbosity {
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
	checkLatency := func(t *testing.T, round, event int, mustBeLessThan, actual, wrapped, mocked time.Duration, compressed *time.Duration) {
		// none are currently expected to wait, but if they do, they must not wait the full timeout.
		// this also helps handle small cpu stutter, as some may wait 1ms or so.
		maxLatency := maxDur(actual, wrapped, mocked)
		minLatency := minDur(actual, wrapped, mocked)
		compressedString := "<n/a>"
		if compressed != nil {
			maxLatency = maxDur(maxLatency, *compressed)
			minLatency = minDur(minLatency, *compressed)
			compressedString = fmt.Sprintf("%v", *compressed)
		}
		// this is not actually asserted.
		// local testing was quite flaky, with even non-blocking "actual" calls (many avilable tokens)
		// taking >10ms with some regularity, and mocks/wrappers randomly doing that as well.
		//
		// checking each's tokens and behavior before/after look entirely like CPU noise,
		// and they do not usually break other tests, so asserting is just skipped for now.
		//
		// logging the time information can still help show why other failures occurred though, e.g. due to
		// a very large wait causing real-time-backed limiters to restore a token when the mock does not.
		t.Logf("all limiters should wait a similar amount of time, got actual: %v, wrapped: %v, mocked: %v, compressed: %v",
			actual, wrapped, mocked, compressedString)
	}

	// aim for the desired duration, and try to avoid timing out from a CLI-enforced deadline too.
	deadline := time.Now().Add(testDuration)
	if testDeadline, ok := t.Deadline(); ok {
		// a test takes like half a second each, leave some buffer so we don't time out
		buffer := 5 * time.Second
		testDeadline = testDeadline.Add(-buffer)
		if testDeadline.Before(time.Now()) {
			t.Fatalf("not enough time to run tests, need at least %v", buffer)
		}
		// tests want to end before the hardcoded deadline, just run shorter.
		if testDeadline.Before(deadline) {
			deadline = testDeadline
		}
	}
	for testnum := 0; !t.Failed() && time.Now().Before(deadline); testnum++ {
		t.Run(fmt.Sprintf("attempt %v", testnum), func(t *testing.T) {
			// parallel saves a fair bit of time but introduces a lot of CPU contention
			// and that leads to a moderate amount of flakiness.
			//
			// unfortunately not recommended here.
			// t.Parallel()

			seed := time.Now().UnixNano()
			seed = 1716164357569243000 // override seed to test a failing scenario
			rng := rand.New(rand.NewSource(seed))
			t.Logf("rand seed: %v", seed)

			burst := rng.Intn(maxBurst)               // zero is uninteresting but allowed
			limit := rate.Every(granularity * events) // refresh after each full round
			t.Logf("limit: %v, burst: %v", granularity, burst)
			initialNow := time.Now()
			ts := NewMockedTimeSourceAt(initialNow)
			compressedTS := NewMockedTimeSourceAt(initialNow)

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
							timeout := granularity / 2
							started := time.Now() // intentionally gathered outside the goroutines, to reveal goroutine-starting lag
							ctx, cancel := context.WithTimeout(context.Background(), timeout)

							a, w, m := false, false, false
							var aLatency, wLatency, mLatency time.Duration
							var g errgroup.Group
							g.Go(func() error {
								_a := actual.Wait(ctx)
								aLatency = time.Since(started).Round(time.Millisecond)
								debug(t, "Wait elapsed: %v, actual err: %v", aLatency, _a)
								a = _a == nil
								return nil
							})
							g.Go(func() error {
								_w := wrapped.Wait(ctx)
								wLatency = time.Since(started).Round(time.Millisecond)
								debug(t, "Wait elapsed: %v, wrapped err: %v", wLatency, _w)
								w = _w == nil
								return nil
							})
							g.Go(func() error {
								_m := mocked.Wait(ctx)
								mLatency = time.Since(started).Round(time.Millisecond)
								debug(t, "Wait elapsed: %v, mocked err: %v", mLatency, _m)
								m = _m == nil
								return nil
							})
							_ = g.Wait()

							check(t, "Wait", round, event, a, w, m, nil)
							checkLatency(t, round, event, timeout, aLatency, wLatency, mLatency, nil)
							compressedReplay[round][event] = func(t *testing.T) {
								// need a mocked-time context, or the real deadline will not match the mocked deadline
								ctx := &timesourceContext{
									ts:       compressedTS,
									deadline: compressedTS.Now().Add(granularity / 2),
								}

								started := time.Now()
								// as we have no mock deadline ctx.Done() chan, this is expected to take some time.
								// rather than running instantly - the mock-timer will not fire.
								_c := compressed.Wait(ctx)
								cLatency := time.Since(started).Round(time.Millisecond)
								debug(t, "Wait elapsed: %v, compressed err: %v", cLatency, _c)
								c := _c == nil
								check(t, "Wait (Compressed)", round, event, a, w, m, &c)
								checkLatency(t, round, event, timeout, aLatency, wLatency, mLatency, &cLatency)
							}
							cancel()
						case 3:
							// call Reserve on everything, and cancel half of them
							now := time.Now() // needed for the real ratelimiter to cancel successfully like the wrapper does
							_a, _w, _m := actual.ReserveN(now, 1), wrapped.Reserve(), mocked.Reserve()
							if rng.Intn(2)%2 == 0 {
								_a.CancelAt(now)
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

func maxDur(d time.Duration, ds ...time.Duration) time.Duration {
	for _, tmp := range ds {
		if tmp > d {
			d = tmp
		}
	}
	return d
}
func minDur(d time.Duration, ds ...time.Duration) time.Duration {
	for _, tmp := range ds {
		if tmp < d {
			d = tmp
		}
	}
	return d
}

type timesourceContext struct {
	// does not contain a parent context as we currently have no need,
	// but a "real" one would for forwarding Value and deadline lookups.
	ts       TimeSource
	deadline time.Time
	mut      sync.Mutex
}

func (t *timesourceContext) Deadline() (deadline time.Time, ok bool) {
	t.mut.Lock()
	defer t.mut.Unlock()
	return t.deadline, t.deadline != time.Time{}
}

func (t *timesourceContext) Done() <-chan struct{} {
	// not currently expected to be used, but it would look like this:
	t.mut.Lock()
	defer t.mut.Unlock()
	c := make(chan struct{})
	t.ts.AfterFunc(t.ts.Now().Sub(t.deadline), func() {
		// this stack may leak if time is not advanced past it in tests.
		close(c)
	})
	return c
}

func (t *timesourceContext) Err() error {
	t.mut.Lock()
	defer t.mut.Unlock()
	if t.ts.Now().After(t.deadline) {
		return context.DeadlineExceeded
	}
	return nil
}

func (t *timesourceContext) Value(key any) any {
	panic("unimplemented")
}

var _ context.Context = (*timesourceContext)(nil)

/*
Benchmark results, removing some noisy logs with `grep ^Benchmark -B1 | grep -v -e '--'`

On an M1 Pro Mac:
❯ go test -bench . -test.run xxx -cpu 1,2,4,8,32 .
goos: darwin
goarch: arm64
pkg: github.com/uber/cadence/common/clock
BenchmarkLimiter/allow/serial/real       	13618585	        79.40 ns/op
    ratelimiter_test.go:68: allowed: 1082245, denied: 12536340
BenchmarkLimiter/allow/serial/real-2     	15085507	        79.27 ns/op
    ratelimiter_test.go:68: allowed: 1196772, denied: 13888735
BenchmarkLimiter/allow/serial/real-4     	15130706	        79.26 ns/op
    ratelimiter_test.go:68: allowed: 1200201, denied: 13930505
BenchmarkLimiter/allow/serial/real-8     	15166467	        79.97 ns/op
    ratelimiter_test.go:68: allowed: 1213767, denied: 13952700
BenchmarkLimiter/allow/serial/real-32    	15186853	        80.26 ns/op
    ratelimiter_test.go:68: allowed: 1219812, denied: 13967041
BenchmarkLimiter/allow/serial/wrapped    	10713946	       118.2 ns/op
    ratelimiter_test.go:68: allowed: 1267080, denied: 9446866
BenchmarkLimiter/allow/serial/wrapped-2  	10821831	       110.3 ns/op
    ratelimiter_test.go:68: allowed: 1194989, denied: 9626842
BenchmarkLimiter/allow/serial/wrapped-4  	10790121	       110.9 ns/op
    ratelimiter_test.go:68: allowed: 1197546, denied: 9592575
BenchmarkLimiter/allow/serial/wrapped-8  	10698758	       111.1 ns/op
    ratelimiter_test.go:68: allowed: 1189432, denied: 9509326
BenchmarkLimiter/allow/serial/wrapped-32 	10760622	       110.6 ns/op
    ratelimiter_test.go:68: allowed: 1190917, denied: 9569705
BenchmarkLimiter/allow/serial/mocked_timesource            	11918107	       106.8 ns/op
    ratelimiter_test.go:68: allowed: 2384621, denied: 9533486
BenchmarkLimiter/allow/serial/mocked_timesource-2          	11764609	       100.4 ns/op
    ratelimiter_test.go:68: allowed: 2353921, denied: 9410688
BenchmarkLimiter/allow/serial/mocked_timesource-4          	11542063	        99.76 ns/op
    ratelimiter_test.go:68: allowed: 2309412, denied: 9232651
BenchmarkLimiter/allow/serial/mocked_timesource-8          	12157638	        99.47 ns/op
    ratelimiter_test.go:68: allowed: 2432527, denied: 9725111
BenchmarkLimiter/allow/serial/mocked_timesource-32         	12005817	        99.16 ns/op
    ratelimiter_test.go:68: allowed: 2402163, denied: 9603654
BenchmarkLimiter/allow/parallel/real                       	14720223	        81.69 ns/op
    ratelimiter_test.go:84: allowed: 1203508, denied: 13516715
BenchmarkLimiter/allow/parallel/real-2                     	12110864	        97.91 ns/op
    ratelimiter_test.go:84: allowed: 1187031, denied: 10923833
BenchmarkLimiter/allow/parallel/real-4                     	 8842538	       123.1 ns/op
    ratelimiter_test.go:84: allowed: 1093257, denied: 7749281
BenchmarkLimiter/allow/parallel/real-8                     	 6893467	       178.3 ns/op
    ratelimiter_test.go:84: allowed: 4181778, denied: 2711689
BenchmarkLimiter/allow/parallel/real-32                    	 5012532	       246.8 ns/op
    ratelimiter_test.go:84: allowed: 4719665, denied: 292867
BenchmarkLimiter/allow/parallel/wrapped                    	 9481642	       115.2 ns/op
    ratelimiter_test.go:84: allowed: 1093546, denied: 8388096
BenchmarkLimiter/allow/parallel/wrapped-2                  	 8678460	       138.4 ns/op
    ratelimiter_test.go:84: allowed: 1201988, denied: 7476472
BenchmarkLimiter/allow/parallel/wrapped-4                  	 6188439	       193.0 ns/op
    ratelimiter_test.go:84: allowed: 1195103, denied: 4993336
BenchmarkLimiter/allow/parallel/wrapped-8                  	 4316008	       264.9 ns/op
    ratelimiter_test.go:84: allowed: 1144162, denied: 3171846
BenchmarkLimiter/allow/parallel/wrapped-32                 	 3078057	       366.7 ns/op
    ratelimiter_test.go:84: allowed: 1129621, denied: 1948436
BenchmarkLimiter/allow/parallel/mocked_timesource          	11716285	       110.3 ns/op
    ratelimiter_test.go:84: allowed: 2344256, denied: 9372029
BenchmarkLimiter/allow/parallel/mocked_timesource-2        	 8348055	       142.8 ns/op
    ratelimiter_test.go:84: allowed: 1670610, denied: 6677445
BenchmarkLimiter/allow/parallel/mocked_timesource-4        	 6955520	       172.9 ns/op
    ratelimiter_test.go:84: allowed: 1392103, denied: 5563417
BenchmarkLimiter/allow/parallel/mocked_timesource-8        	 6043203	       195.8 ns/op
    ratelimiter_test.go:84: allowed: 1209640, denied: 4833563
BenchmarkLimiter/allow/parallel/mocked_timesource-32       	 6134875	       216.2 ns/op
    ratelimiter_test.go:84: allowed: 1227974, denied: 4906901
BenchmarkLimiter/reserve-canceling-every-3/serial/real     	 6522390	       186.0 ns/op
    ratelimiter_test.go:68: allowed: 1178, denied: 6521212
BenchmarkLimiter/reserve-canceling-every-3/serial/real-2   	 6890377	       173.8 ns/op
    ratelimiter_test.go:68: allowed: 1177, denied: 6889200
BenchmarkLimiter/reserve-canceling-every-3/serial/real-4   	 6609655	       180.0 ns/op
    ratelimiter_test.go:68: allowed: 1183, denied: 6608472
BenchmarkLimiter/reserve-canceling-every-3/serial/real-8   	 6943302	       174.1 ns/op
    ratelimiter_test.go:68: allowed: 1191, denied: 6942111
BenchmarkLimiter/reserve-canceling-every-3/serial/real-32  	 6864550	       175.8 ns/op
    ratelimiter_test.go:68: allowed: 1183, denied: 6863367
BenchmarkLimiter/reserve-canceling-every-3/serial/real_pinned_time            	11740456	       102.6 ns/op
    ratelimiter_test.go:68: allowed: 1724, denied: 11738732
BenchmarkLimiter/reserve-canceling-every-3/serial/real_pinned_time-2          	11738256	       102.6 ns/op
    ratelimiter_test.go:68: allowed: 1730, denied: 11736526
BenchmarkLimiter/reserve-canceling-every-3/serial/real_pinned_time-4          	11573120	       102.8 ns/op
    ratelimiter_test.go:68: allowed: 1727, denied: 11571393
BenchmarkLimiter/reserve-canceling-every-3/serial/real_pinned_time-8          	11677448	       102.8 ns/op
    ratelimiter_test.go:68: allowed: 1734, denied: 11675714
BenchmarkLimiter/reserve-canceling-every-3/serial/real_pinned_time-32         	11776212	       103.1 ns/op
    ratelimiter_test.go:68: allowed: 1739, denied: 11774473
BenchmarkLimiter/reserve-canceling-every-3/serial/wrapped                     	 4711150	       247.7 ns/op
    ratelimiter_test.go:68: allowed: 2333, denied: 4708817
BenchmarkLimiter/reserve-canceling-every-3/serial/wrapped-2                   	 5511319	       208.9 ns/op
    ratelimiter_test.go:68: allowed: 2151, denied: 5509168
BenchmarkLimiter/reserve-canceling-every-3/serial/wrapped-4                   	 5655414	       211.9 ns/op
    ratelimiter_test.go:68: allowed: 2102, denied: 5653312
BenchmarkLimiter/reserve-canceling-every-3/serial/wrapped-8                   	 5629956	       212.7 ns/op
    ratelimiter_test.go:68: allowed: 2111, denied: 5627845
BenchmarkLimiter/reserve-canceling-every-3/serial/wrapped-32                  	 5341764	       214.0 ns/op
    ratelimiter_test.go:68: allowed: 2189, denied: 5339575
BenchmarkLimiter/reserve-canceling-every-3/serial/mocked_timesource           	 5575029	       218.6 ns/op
    ratelimiter_test.go:68: allowed: 2142, denied: 5572887
BenchmarkLimiter/reserve-canceling-every-3/serial/mocked_timesource-2         	 6456499	       187.8 ns/op
    ratelimiter_test.go:68: allowed: 2142, denied: 6454357
BenchmarkLimiter/reserve-canceling-every-3/serial/mocked_timesource-4         	 6344293	       188.5 ns/op
    ratelimiter_test.go:68: allowed: 2142, denied: 6342151
BenchmarkLimiter/reserve-canceling-every-3/serial/mocked_timesource-8         	 6388226	       193.3 ns/op
    ratelimiter_test.go:68: allowed: 2142, denied: 6386084
BenchmarkLimiter/reserve-canceling-every-3/serial/mocked_timesource-32        	 5958344	       192.0 ns/op
    ratelimiter_test.go:68: allowed: 2142, denied: 5956202
BenchmarkLimiter/reserve-canceling-every-3/parallel/real                      	 6390843	       189.0 ns/op
    ratelimiter_test.go:84: allowed: 1216, denied: 6389627
BenchmarkLimiter/reserve-canceling-every-3/parallel/real-2                    	 8914696	       131.2 ns/op
    ratelimiter_test.go:84: allowed: 1135, denied: 8913561
BenchmarkLimiter/reserve-canceling-every-3/parallel/real-4                    	 5277980	       217.4 ns/op
    ratelimiter_test.go:84: allowed: 1673, denied: 5276307
BenchmarkLimiter/reserve-canceling-every-3/parallel/real-8                    	 4012483	       294.0 ns/op
    ratelimiter_test.go:84: allowed: 3918350, denied: 94133
BenchmarkLimiter/reserve-canceling-every-3/parallel/real-32                   	 3438741	       347.8 ns/op
    ratelimiter_test.go:84: allowed: 3404112, denied: 34629
BenchmarkLimiter/reserve-canceling-every-3/parallel/real_pinned_time          	10423504	       104.2 ns/op
    ratelimiter_test.go:84: allowed: 1733, denied: 10421771
BenchmarkLimiter/reserve-canceling-every-3/parallel/real_pinned_time-2        	 9515350	       123.6 ns/op
    ratelimiter_test.go:84: allowed: 1775, denied: 9513575
BenchmarkLimiter/reserve-canceling-every-3/parallel/real_pinned_time-4        	 6880100	       174.7 ns/op
    ratelimiter_test.go:84: allowed: 3497, denied: 6876603
BenchmarkLimiter/reserve-canceling-every-3/parallel/real_pinned_time-8        	 5263464	       213.0 ns/op
    ratelimiter_test.go:84: allowed: 4781739, denied: 481725
BenchmarkLimiter/reserve-canceling-every-3/parallel/real_pinned_time-32       	 3952268	       324.6 ns/op
    ratelimiter_test.go:84: allowed: 3951409, denied: 859
BenchmarkLimiter/reserve-canceling-every-3/parallel/wrapped                   	 5057402	       240.3 ns/op
    ratelimiter_test.go:84: allowed: 2063, denied: 5055339
BenchmarkLimiter/reserve-canceling-every-3/parallel/wrapped-2                 	 3224310	       359.6 ns/op
    ratelimiter_test.go:84: allowed: 2666, denied: 3221644
BenchmarkLimiter/reserve-canceling-every-3/parallel/wrapped-4                 	 2820698	       422.3 ns/op
    ratelimiter_test.go:84: allowed: 2612, denied: 2818086
BenchmarkLimiter/reserve-canceling-every-3/parallel/wrapped-8                 	 2131276	       538.6 ns/op
    ratelimiter_test.go:84: allowed: 6946, denied: 2124330
BenchmarkLimiter/reserve-canceling-every-3/parallel/wrapped-32                	 1543068	       784.4 ns/op
    ratelimiter_test.go:84: allowed: 1457368, denied: 85700
BenchmarkLimiter/reserve-canceling-every-3/parallel/mocked_timesource         	 5244424	       221.1 ns/op
    ratelimiter_test.go:84: allowed: 2142, denied: 5242282
BenchmarkLimiter/reserve-canceling-every-3/parallel/mocked_timesource-2       	 3810586	       314.6 ns/op
    ratelimiter_test.go:84: allowed: 2078, denied: 3808508
BenchmarkLimiter/reserve-canceling-every-3/parallel/mocked_timesource-4       	 2904224	       385.6 ns/op
    ratelimiter_test.go:84: allowed: 1905, denied: 2902319
BenchmarkLimiter/reserve-canceling-every-3/parallel/mocked_timesource-8       	 2420433	       499.5 ns/op
    ratelimiter_test.go:84: allowed: 1774, denied: 2418659
BenchmarkLimiter/reserve-canceling-every-3/parallel/mocked_timesource-32      	 1649202	       740.3 ns/op
    ratelimiter_test.go:84: allowed: 1559, denied: 1647643
BenchmarkLimiter/wait/serial/2ns_rate/real                                    	 2402028	       496.4 ns/op
    ratelimiter_test.go:68: allowed: 2402028, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/real-2                                  	 2979196	       431.4 ns/op
    ratelimiter_test.go:68: allowed: 2979196, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/real-4                                  	 3018830	       400.7 ns/op
    ratelimiter_test.go:68: allowed: 3018830, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/real-8                                  	 2954853	       416.0 ns/op
    ratelimiter_test.go:68: allowed: 2954853, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/real-32                                 	 2862715	       408.1 ns/op
    ratelimiter_test.go:68: allowed: 2862715, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/wrapped                                 	 2292128	       536.9 ns/op
    ratelimiter_test.go:68: allowed: 2292128, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/wrapped-2                               	 2834373	       437.3 ns/op
    ratelimiter_test.go:68: allowed: 2834373, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/wrapped-4                               	 2824138	       422.2 ns/op
    ratelimiter_test.go:68: allowed: 2824138, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/wrapped-8                               	 2879480	       428.0 ns/op
    ratelimiter_test.go:68: allowed: 2879480, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/wrapped-32                              	 2539308	       438.4 ns/op
    ratelimiter_test.go:68: allowed: 2539308, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/mocked_timesource                       	 5113411	       235.9 ns/op
    ratelimiter_test.go:68: allowed: 5113411, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/mocked_timesource-2                     	 5947539	       199.2 ns/op
    ratelimiter_test.go:68: allowed: 5947539, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/mocked_timesource-4                     	 6080888	       196.1 ns/op
    ratelimiter_test.go:68: allowed: 6080888, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/mocked_timesource-8                     	 6101578	       195.8 ns/op
    ratelimiter_test.go:68: allowed: 6101578, denied: 0
BenchmarkLimiter/wait/serial/2ns_rate/mocked_timesource-32                    	 6003756	       205.0 ns/op
    ratelimiter_test.go:68: allowed: 6003756, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/real                                    	 1201179	       999.2 ns/op
    ratelimiter_test.go:68: allowed: 1201179, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/real-2                                  	 1201159	       999.2 ns/op
    ratelimiter_test.go:68: allowed: 1201159, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/real-4                                  	 1201179	       999.2 ns/op
    ratelimiter_test.go:68: allowed: 1201179, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/real-8                                  	 1201182	       999.2 ns/op
    ratelimiter_test.go:68: allowed: 1201182, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/real-32                                 	 1201178	       999.2 ns/op
    ratelimiter_test.go:68: allowed: 1201178, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/wrapped                                 	 1200878	       999.2 ns/op
    ratelimiter_test.go:68: allowed: 1200878, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/wrapped-2                               	 1201171	      1000 ns/op
    ratelimiter_test.go:68: allowed: 1201171, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/wrapped-4                               	 1201180	       999.2 ns/op
    ratelimiter_test.go:68: allowed: 1201180, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/wrapped-8                               	 1200398	       999.2 ns/op
    ratelimiter_test.go:68: allowed: 1200398, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/wrapped-32                              	 1201142	       999.2 ns/op
    ratelimiter_test.go:68: allowed: 1201142, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/mocked_timesource                       	 5126984	       238.0 ns/op
    ratelimiter_test.go:68: allowed: 5126984, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/mocked_timesource-2                     	 6061310	       199.3 ns/op
    ratelimiter_test.go:68: allowed: 6061310, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/mocked_timesource-4                     	 6107908	       200.6 ns/op
    ratelimiter_test.go:68: allowed: 6107908, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/mocked_timesource-8                     	 6011832	       200.5 ns/op
    ratelimiter_test.go:68: allowed: 6011832, denied: 0
BenchmarkLimiter/wait/serial/1µs_rate/mocked_timesource-32                    	 5872417	       202.2 ns/op
    ratelimiter_test.go:68: allowed: 5872417, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/real                                  	 2416035	       497.3 ns/op
    ratelimiter_test.go:84: allowed: 2416035, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/real-2                                	 3957284	       352.1 ns/op
    ratelimiter_test.go:84: allowed: 3957284, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/real-4                                	 3406399	       348.8 ns/op
    ratelimiter_test.go:84: allowed: 3406399, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/real-8                                	 2549136	       469.1 ns/op
    ratelimiter_test.go:84: allowed: 2549136, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/real-32                               	 2143489	       541.1 ns/op
    ratelimiter_test.go:84: allowed: 2143489, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/wrapped                               	 2177626	       544.1 ns/op
    ratelimiter_test.go:84: allowed: 2177626, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/wrapped-2                             	 3984897	       294.6 ns/op
    ratelimiter_test.go:84: allowed: 3984897, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/wrapped-4                             	 3445614	       366.8 ns/op
    ratelimiter_test.go:84: allowed: 3445614, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/wrapped-8                             	 1841377	       642.8 ns/op
    ratelimiter_test.go:84: allowed: 1841377, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/wrapped-32                            	 1589787	       748.8 ns/op
    ratelimiter_test.go:84: allowed: 1589787, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/mocked_timesource                     	 5076840	       234.8 ns/op
    ratelimiter_test.go:84: allowed: 5076840, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/mocked_timesource-2                   	 5381004	       218.5 ns/op
    ratelimiter_test.go:84: allowed: 5381004, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/mocked_timesource-4                   	 4833823	       245.8 ns/op
    ratelimiter_test.go:84: allowed: 4833823, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/mocked_timesource-8                   	 4418911	       279.7 ns/op
    ratelimiter_test.go:84: allowed: 4418911, denied: 0
BenchmarkLimiter/wait/parallel/2ns_rate/mocked_timesource-32                  	 4012911	       301.5 ns/op
    ratelimiter_test.go:84: allowed: 4012911, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/real                                  	 1201164	       999.2 ns/op
    ratelimiter_test.go:84: allowed: 1201164, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/real-2                                	 1219983	       983.1 ns/op
    ratelimiter_test.go:84: allowed: 1219983, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/real-4                                	 1297210	       925.0 ns/op
    ratelimiter_test.go:84: allowed: 1297210, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/real-8                                	 2537467	       470.7 ns/op
    ratelimiter_test.go:84: allowed: 2537467, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/real-32                               	 2110921	       539.8 ns/op
    ratelimiter_test.go:84: allowed: 2110921, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/wrapped                               	 1201170	       999.2 ns/op
    ratelimiter_test.go:84: allowed: 1201170, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/wrapped-2                             	 1201172	       999.2 ns/op
    ratelimiter_test.go:84: allowed: 1201172, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/wrapped-4                             	 1201144	       999.2 ns/op
    ratelimiter_test.go:84: allowed: 1201144, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/wrapped-8                             	 1201162	       999.2 ns/op
    ratelimiter_test.go:84: allowed: 1201162, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/wrapped-32                            	 1201062	       999.3 ns/op
    ratelimiter_test.go:84: allowed: 1201062, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/mocked_timesource                     	 5097742	       244.0 ns/op
    ratelimiter_test.go:84: allowed: 5097742, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/mocked_timesource-2                   	 5727841	       212.0 ns/op
    ratelimiter_test.go:84: allowed: 5727841, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/mocked_timesource-4                   	 4811091	       247.9 ns/op
    ratelimiter_test.go:84: allowed: 4811091, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/mocked_timesource-8                   	 4357129	       269.6 ns/op
    ratelimiter_test.go:84: allowed: 4357129, denied: 0
BenchmarkLimiter/wait/parallel/1µs_rate/mocked_timesource-32                  	 4148046	       296.5 ns/op
    ratelimiter_test.go:84: allowed: 4148046, denied: 0
BenchmarkLimiterParallel/real                                                 	14492636	        83.45 ns/op
    ratelimiter_test.go:306: allowed: 1210390, denied: 13282246
BenchmarkLimiterParallel/real-2                                               	11602707	       100.8 ns/op
    ratelimiter_test.go:306: allowed: 1171781, denied: 10430926
BenchmarkLimiterParallel/real-4                                               	 9222276	       129.0 ns/op
    ratelimiter_test.go:306: allowed: 1196377, denied: 8025899
BenchmarkLimiterParallel/real-8                                               	 7036296	       170.0 ns/op
    ratelimiter_test.go:306: allowed: 3823811, denied: 3212485
BenchmarkLimiterParallel/real-32                                              	 4980267	       243.1 ns/op
    ratelimiter_test.go:306: allowed: 4963185, denied: 17082
BenchmarkLimiterParallel/mocked                                               	10211008	       116.7 ns/op
    ratelimiter_test.go:306: allowed: 1192235, denied: 9018773
BenchmarkLimiterParallel/mocked-2                                             	 8599737	       138.1 ns/op
    ratelimiter_test.go:306: allowed: 1188970, denied: 7410767
BenchmarkLimiterParallel/mocked-4                                             	 5911910	       204.9 ns/op
    ratelimiter_test.go:306: allowed: 1212087, denied: 4699823
BenchmarkLimiterParallel/mocked-8                                             	 4406569	       249.9 ns/op
    ratelimiter_test.go:306: allowed: 1101959, denied: 3304610
BenchmarkLimiterParallel/mocked-32                                            	 3157717	       374.9 ns/op
    ratelimiter_test.go:306: allowed: 1184783, denied: 1972934
BenchmarkLimiterParallel/mocked_timesource                                    	11131578	       112.3 ns/op
    ratelimiter_test.go:306: allowed: 5566788, denied: 5564790
BenchmarkLimiterParallel/mocked_timesource-2                                  	 8479593	       140.7 ns/op
    ratelimiter_test.go:306: allowed: 4240796, denied: 4238797
BenchmarkLimiterParallel/mocked_timesource-4                                  	 6908698	       172.1 ns/op
    ratelimiter_test.go:306: allowed: 3455348, denied: 3453350
BenchmarkLimiterParallel/mocked_timesource-8                                  	 6280804	       191.8 ns/op
    ratelimiter_test.go:306: allowed: 3141401, denied: 3139403
BenchmarkLimiterParallel/mocked_timesource-32                                 	 5643787	       246.0 ns/op
    ratelimiter_test.go:306: allowed: 2822893, denied: 2820894
PASS
ok  	github.com/uber/cadence/common/clock	225.366s

*/
