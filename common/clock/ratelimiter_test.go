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
Benchmark data can be seen in the testdata folder, for relevant machines.

Interpretation:
  - for essentially all operations, mostly regardless of sequential / parallel / how parallel,
    the added latency is ~25% to ~40%, and even extremes are less than double.
    this seems entirely tolerable and safe to use.
  - mock-time ratelimiter is faster than real-time in essentially all cases, as you would hope.
    though not by much, unless waiting was part of the test.
  - real *rate.Limiter instances have some pathological behavior that... might disqualify them from safe use tbh.
    essentially this summarizes as "time.Now() is not globally monotonic", and it can lead to significant over and under limiting.

The real *rate.Limiter instances show these quirks during parallel tests, due to time rewinding and jumping forward repeatedly
from the *rate.Limiter's point of view (as it has an internal last-Now value protected by a mutex):
 1. `Allow()`-like calls (both literally `Allow()`, and `Reserve()` followed by sometimes `reservation.Cancel()`)
    can allow both significantly more or significantly fewer calls than they should.
 2. `Wait()` can allow calls through at a faster rate than it should.  At peak, a bit over 2x.
    (and probably slower, but sleeping is not guaranteed to wake up at a precise time so would still be correct behavior)

Essentially, when one call sets the time to "now", and an out-of-sync call sets it to "now-1ns" (as time between
goroutines is not consistent, this is expected), a token may be restored when a time-consistent limiter would not do so.

You can recreate this by hand by feeding a ratelimiter "now" and "now+1s" randomly, and watching what it
allows / what the value of Tokens() is as time passes.  Doing this *literally* by hand, e.g. pressing enter on a
command-line loop that does this, easily shows it - high frequency is not necessary, just illogical "time travel".

The first can be seen in m1_mac.txt by comparing these tests:
  - BenchmarkLimiter/allow/parallel/real-4 (note the near-1µs time per allow, which serial calls match)
  - BenchmarkLimiter/allow/parallel/real-8 (time-per-allow is now 231ns, over 4x more allowed than it should)
  - BenchmarkLimiter/reserve-canceling-every-3/parallel/real-2 (1.5ms per allow, around 1,000x fewer allowed than it should)
  - BenchmarkLimiter/reserve-canceling-every-3/parallel/real-8 (450ns per allow, around 2x more than it should)

"real with pinned time" behaves similarly, for similar reasons: canceling can time-travel.
pinned time does make the behavior when called on a single goroutine roughly perfect though.

And the second can be seen in m1_mac.txt by comparing these tests:
  - BenchmarkLimiter/wait/parallel/1µs_rate/real (almost exactly 1µs per iteration)
  - BenchmarkLimiter/wait/parallel/1µs_rate/real-8 (~500ns, twice as fast as it should allow)

I have not tried to verify Wait's cause by hand, but it certainly seems like time-thrashing explains it as well.
This is also supported by the wrapper's time-locking *completely* eliminating this flaw, as all iterations take
almost exactly 1µs (or longer) as they should.

---

All machines I've tried it on behave similarly, though whether one parallel test triggers misbehavior or not
is fairly random.  Higher parallelism triggers it more, and it seems *fairly* likely that at least one misbehaves
in a full `-cpu 1,2,4,8,...[>=cores]` suite, and sometimes many more.
*/
func BenchmarkLimiter(b *testing.B) {
	const (
		normalLimit      = time.Microsecond // pretty arbitrary but has a good blend of allow and deny
		burst            = 1000             // arbitrary, but more interesting than 1, and better matches how we use limiters (rate/s == burst)
		allowedPeriodFmt = "time per allow: %9s"
	)
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
		as := fmt.Sprintf("allowed: %v,", allowed)
		ds := fmt.Sprintf("denied: %v,", denied)
		allowedPeriod := fmt.Sprintf(allowedPeriodFmt, "n/a")
		if allowed > 0 {
			allowedPeriod = fmt.Sprintf(allowedPeriodFmt, fmt.Sprint(b.Elapsed()/time.Duration(allowed)))
		}
		b.Logf("%-20s %-20s %s", as, ds, allowedPeriod) // allows controlling whitespace better
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
		as := fmt.Sprintf("allowed: %v,", allowed.Load())
		ds := fmt.Sprintf("denied: %v,", denied.Load())
		allowedPeriod := fmt.Sprintf(allowedPeriodFmt, "n/a")
		if allowed.Load() > 0 {
			allowedPeriod = fmt.Sprintf(allowedPeriodFmt, fmt.Sprint(b.Elapsed()/time.Duration(allowed.Load())))
		}
		b.Logf("%-20s %-20s %s", as, ds, allowedPeriod) // allows controlling whitespace better
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
					rl := rate.NewLimiter(rate.Every(normalLimit), burst)
					runner(b, func(i int) bool {
						return rl.Allow()
					})
				})
				b.Run("wrapped", func(b *testing.B) {
					rl := NewRatelimiter(rate.Every(normalLimit), burst)
					runner(b, func(i int) bool {
						return rl.Allow()
					})
				})
				b.Run("mocked timesource", func(b *testing.B) {
					ts := NewMockedTimeSource()
					rl := NewMockRatelimiter(ts, rate.Every(normalLimit), burst)
					runner(b, func(i int) bool {
						// adjusted by eye, to try to very roughly match the above values for the final runs.
						// probably needs to be tweaked per machine.
						ts.Advance(normalLimit / 5)
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
					rl := rate.NewLimiter(rate.Every(normalLimit), burst)
					runner(b, func(i int) (allowed bool) {
						r := rl.Reserve()
						allowed = r.OK() && r.Delay() == 0
						canceled := i%cancelNth == 0
						allowed = allowed && !canceled
						if !allowed {
							r.Cancel() // all not-allowed calls must be canceled, as they will not be "used"
							return
						}
						return
					})
				})
				b.Run("real pinned time", func(b *testing.B) {
					rl := rate.NewLimiter(rate.Every(normalLimit), burst)
					runner(b, func(i int) (allowed bool) {
						// expected to be faster as the limiter's time does not
						// advance as often, and "now" is only gathered once.
						//
						// but this is not possible to do correctly with concurrent use,
						// so it's purely synthetic and serves as a lower bound only.
						now := time.Now()
						r := rl.ReserveN(now, 1)
						allowed = r.OK() && r.DelayFrom(now) == 0
						canceled := i%cancelNth == 0
						allowed = allowed && !canceled
						if !allowed {
							r.CancelAt(now) // all not-allowed calls must be canceled, as they will not be "used"
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
						allowed = allowed && !canceled
						r.Used(allowed)
						return
					})
				}

				// notice that the allowed/denied metrics are roughly the same whether parallel or not.
				// this is how it should be, as the benchmark run times are similar.
				b.Run("wrapped", func(b *testing.B) {
					rl := NewRatelimiter(rate.Every(normalLimit), burst)
					runWrapped(b, rl, runner, nil)
				})
				b.Run("mocked timesource", func(b *testing.B) {
					ts := NewMockedTimeSource()
					rl := NewMockRatelimiter(ts, rate.Every(normalLimit), burst)
					runWrapped(b, rl, runner, func() {
						ts.Advance(normalLimit / 5)
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
					// try an insanely fast refresh, to see maximum throughput with "zero" waiting.
					// intentionally avoiding 1 nanosecond in case it gets special-cased, like 0 and math.MaxInt64.
					2 * time.Nanosecond,
					// try a rate that should contend and wait heavily.
					// this SHOULD be how long the per-iteration time is (when not mocked)... but read comments below.
					normalLimit,
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
							rl := rate.NewLimiter(limit, burst)
							runner(b, func(i int) bool {
								ctx, cancel := context.WithTimeout(context.Background(), timeout)
								defer cancel()
								return rl.Wait(ctx) == nil
							})
						})
						b.Run("wrapped", func(b *testing.B) {
							rl := NewRatelimiter(limit, burst)
							runner(b, func(i int) bool {
								ctx, cancel := context.WithTimeout(context.Background(), timeout)
								defer cancel()
								return rl.Wait(ctx) == nil
							})
						})
						b.Run("mocked timesource", func(b *testing.B) {
							ts := NewMockedTimeSource()
							rl := NewMockRatelimiter(ts, limit, burst)
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
