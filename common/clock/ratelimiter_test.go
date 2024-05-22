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
	"strings"
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
    the added latency is ~50%, and even extremes are less than double.
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

Essentially, when one call sets the time to "now", and an out-of-sync call sets it to "now-1ms" (time is acquired outside
the limiter's lock, and they may make progress out of order), a token may be restored when a time-consistent limiter would
not do so.

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
almost exactly 1µs as they should.

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
						// anything not extremely higher or lower is probably fine.
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
						// any value should work, but smaller than normalLimit revealed
						// some sub-optimal behavior in the past, where time-thrashing
						// would lead to extremely low allowed calls.
						//
						// this is now resolved with synchronous cancels when reservations
						// are not synchronously allowed.
						ts.Advance(normalLimit / 10)
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
								// advance enough to restore a token so it does not block forever.
								// "mock-wait times out when it should" is asserted in the tests.
								ts.Advance(dur)
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
		maxBurst = 5 // half of events, improves odds of Wait waiting

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

		// debug-level output is very noisy on successful runs, so hide it by default.
		// using a custom seed will also set this to true.
		logDebug   = false
		customSeed = 0 // override with non-zero seed to run just one test with that seed
	)

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
			t.Logf("please replay with the failing seed (set `customSeed`) and check detailed output to see what the behavior should be.")
			t.Logf("try setting `logDebug` to false to true to show Tokens() progression, this can help explain many issues.")
			t.Logf("")
			t.Logf("if you are intentionally making changes, be sure to run a few hundred rounds to make sure your changes are stable,")
			t.Logf("and in particular make sure an 'Advancing mock time ...' event occurs and times match, as that is a bit rare")
			t.Logf("---- CAUTION ----")
		}
	})

	debug := func(t *testing.T, format string, args ...interface{}) {
		// do nothing
	}
	if logDebug || customSeed != 0 {
		debug = func(t *testing.T, format string, args ...interface{}) {
			t.Logf(format, args...)
		}
	}
	_ = debug

	const (
		components  = "%-11s %-14s %-15s %-14s" // "{what:} {actual} {wrapped} {mocked}" with padding
		normalRound = "round[%v][%v] " + components
	)
	assertLimitersAgree := func(t *testing.T, what string, round, event int, actual, wrapped, mocked bool, compressed *bool) {
		t.Helper()
		sWhat := what + ":"
		sActual := fmt.Sprintf("actual: %v,", actual)
		sWrapped := fmt.Sprintf("wrapped: %v,", wrapped)
		sMocked := fmt.Sprintf("mocked: %v,", mocked)
		// format strings should share padding so output aligns and it's easy to read.
		// unfortunately only strings do this nicely.s
		if compressed != nil {
			sCompressed := "compressed: " + fmt.Sprint(*compressed)
			t.Logf(normalRound+" %s", round, event, sWhat, sActual, sWrapped, sMocked, sCompressed)
			assert.True(t, actual == wrapped && wrapped == mocked && mocked == *compressed, "ratelimiters disagree")
		} else {
			t.Logf(normalRound, round, event, sWhat, sActual, sWrapped, sMocked)
			assert.True(t, actual == wrapped && wrapped == mocked, "ratelimiters disagree")
		}
	}
	logLatency := func(t *testing.T, round, event int, mustBeLessThan, actual, wrapped, mocked time.Duration, compressed *time.Duration) {
		// none are currently expected to wait, but if they do, they must not wait the full timeout.
		// this also helps handle small cpu stutter, as some may wait 1ms or so.
		maxLatency := maxDur(actual, wrapped, mocked)
		minLatency := minDur(actual, wrapped, mocked)
		compressedString := ""
		if compressed != nil {
			maxLatency = maxDur(maxLatency, *compressed)
			minLatency = minDur(minLatency, *compressed)
			compressedString = fmt.Sprintf("compressed: %v", *compressed)
		}
		// this is unfortunately not asserted.
		//
		// local testing was quite flaky, with even non-blocking "actual" calls (many avilable tokens)
		// taking >10ms with some regularity, and mocks/wrappers randomly doing that as well.
		//
		// checking each's tokens and behavior before/after, they look entirely like CPU noise
		// and they do not usually break other tests, so asserting is just skipped for now.
		//
		// logging the time information can still help show why other failures occurred though, e.g. due to
		// a very large wait causing real-time-backed limiters to restore a token when the mock does not.
		//
		// log format tries to align with assertLimitersAgree
		indent := strings.Repeat(" ", len("round[1][1]"))
		sActual := fmt.Sprintf("actual: %v,", actual)
		sWrapped := fmt.Sprintf("wrapped: %v,", wrapped)
		sMocked := fmt.Sprintf("mocked: %v,", mocked)
		t.Logf("%s "+components+" %s",
			indent, "Wait time:", sActual, sWrapped, sMocked, compressedString)
	}

	deadline := time.Now().Add(testDuration)
	for testnum := 0; !t.Failed() && time.Now().Before(deadline); testnum++ {
		if testnum > 0 && customSeed != 0 {
			// already ran the seeded test, stop early.
			break
		}
		t.Run(fmt.Sprintf("attempt %v", testnum), func(t *testing.T) {
			// parallel saves a fair bit of time but introduces a lot of CPU contention
			// and that leads to a moderate amount of flakiness.
			//
			// unfortunately not recommended here.
			// t.Parallel()

			seed := time.Now().UnixNano()
			if customSeed != 0 {
				seed = customSeed // override seed to test a specific scenario
			}
			rng := rand.New(rand.NewSource(seed))
			t.Logf("rand seed: %v", seed)

			burst := rng.Intn(maxBurst)               // zero is uninteresting but allowed
			limit := rate.Every(granularity * events) // refresh after each full round
			t.Logf("limit: %v, burst: %v", granularity, burst)

			initialNow := time.Now().Round(time.Millisecond) // easier to read when logging/debugging when rounded
			ts := NewMockedTimeSourceAt(initialNow)
			compressedTS := NewMockedTimeSourceAt(initialNow)

			actual := rate.NewLimiter(limit, burst)
			wrapped := NewRatelimiter(limit, burst)
			mocked := NewMockRatelimiter(ts, limit, burst)
			compressed := NewMockRatelimiter(compressedTS, limit, burst)

			// generate some non-colliding "A == perform an Allow call" rounds.
			calls := [rounds][events]string{} // using string so it prints nicely
			for i := range calls {
				for j := range calls[i] {
					calls[i][j] = "_" // "do nothing" marker
				}
			}
			skipWait := false
			for i := 0; i < len(calls[0]); i++ {
				if skipWait {
					skipWait = false
					continue
				}
				// pick one to set to true, or none (1 chance for none)
				set := rng.Intn(len(calls) + 1)
				if set < len(calls) {
					// 1/3rd of them are waits
					switch rng.Intn(4) {
					case 0:
						calls[set][i] = "a" // allow
					case 1:
						calls[set][i] = "r" // reserve
					case 2:
						calls[set][i] = "c" // reserve but cancel
					case 3:
						calls[set][i] = "w" // wait
						// waits can pause for an event and a half, so skip the next event to avoid overlapping
						skipWait = true
					default:
						panic("missing a case")
					}
				}
			}
			t.Log("round setup:")
			for _, round := range calls {
				t.Logf("\t%v", round)
			}
			/*
				rounds look like:
					[r _ a w _ _ _ _ _ w]
					[_ c _ _ _ _ w _ _ _]

				which means that the first round will:
					- reserve()
					- do nothing
					- allow()
					- wait() (may extend to next event, so the next one is always "_")
					- ...

				and by the end of that first array it'll reach the end of
				the `rate.Every` time, and may recover tokens during the next
				array's runtime.  so the second round will:
					- do nothing, just refresh .1 while waiting ==> will have 1.0 tokens regenerated,
					  as it was reserved earlier (if non-zero burst).
					  (this could be consumed by last round's final wait)
					- reserve but cancel, restoring the token (if any), and maybe refresh .1 token
					- do nothing, maybe refresh .1 more token...
					- ... and try to wait on the 7th event.

				non-blank events do not overlap to avoid triggering a race between the
				old consumed tokens being refreshed and a new call consuming them, as these
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
					ts.Advance(granularity) // may also be advanced inside wait
					debug(t, "Tokens before round, real: %0.2f, wrapped: %0.2f, mocked: %0.2f", actual.Tokens(), wrapped.Tokens(), mocked.Tokens())

					switch calls[round][event] {
					case "a":
						// call Allow on everything
						a, w, m := actual.Allow(), wrapped.Allow(), mocked.Allow()
						assertLimitersAgree(t, "Allow", round, event, a, w, m, nil)
						compressedReplay[round][event] = func(t *testing.T) {
							c := compressed.Allow()
							assertLimitersAgree(t, "Allow", round, event, a, w, m, &c)
						}
					case "r":
						// call Reserve on everything, don't cancel any we acquired successfully
						now := time.Now()
						_a, _w, _m := actual.ReserveN(now, 1), wrapped.Reserve(), mocked.Reserve()
						a, w, m := _a.OK() && _a.DelayFrom(now) == 0, _w.Allow(), _m.Allow()

						// un-reserve any that were not successful, we don't leave them dangling.
						if !a {
							_a.CancelAt(now)
						}
						_w.Used(w)
						_m.Used(m)
						assertLimitersAgree(t, "Reserve", round, event, a, w, m, nil)
						compressedReplay[round][event] = func(t *testing.T) {
							_c := compressed.Reserve()
							c := _c.Allow()
							_c.Used(c)
							assertLimitersAgree(t, "Reserve", round, event, a, w, m, &c)
						}
					case "c":
						// call Reserve on everything, and cancel them all
						now := time.Now() // needed for the real ratelimiter to cancel successfully like the wrapper does
						_a, _w, _m := actual.ReserveN(now, 1), wrapped.Reserve(), mocked.Reserve()
						_a.CancelAt(now)
						_w.Used(false)
						_m.Used(false)
						t.Logf("round[%v][%v] Cancel:     canceled", round, event)
						compressedReplay[round][event] = func(t *testing.T) {
							compressed.Reserve().Used(false)
							t.Logf("round[%v][%v] Cancel:     canceled", round, event)
						}
					case "w":
						/*
							Try a brief Wait on everything.

							ctx should expire in the middle of the *next* event, so all three possibilities are exercised:
							  - sometimes unblocks immediately because a token is available
							  - sometimes waits because a token will become available soon
							  - sometimes unblocks immediately because a token will not become available in time

							you can see waits logged like this, they're somewhat rare though.  otherwise wait times are ~0s:
								regular:
									ratelimiter_test.go:402: Wait elapsed: 20ms, wrapped err: <nil>
									ratelimiter_test.go:639: Advancing mock time by 20ms
									ratelimiter_test.go:402: Wait elapsed: 20ms, mocked err: <nil>
									ratelimiter_test.go:402: Wait elapsed: 20ms, actual err: <nil>
									ratelimiter_test.go:655: round[0][9] Wait:       actual: true,  wrapped: true,  mocked: true,
									ratelimiter_test.go:455:             Wait time:  actual: 20ms,  wrapped: 20ms,  mocked: 20ms,
								compressed:
									ratelimiter_test.go:672: Advancing mock time by 20ms
									ratelimiter_test.go:402: Wait elapsed: 20ms, compressed err: <nil>
									ratelimiter_test.go:684: round[0][9] Wait:       actual: true,  wrapped: true,  mocked: true,  compressed: true
									ratelimiter_test.go:455:             Wait time:  actual: 20ms,  wrapped: 20ms,  mocked: 20ms,  compressed: 20ms
						*/
						timeout := granularity + (granularity / 2)
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
							// mocked time needs something to advance it, or it'll block until timeout...
							// but don't advance time if it returned immediately, the others won't wait either
							done := make(chan struct{})
							go func() {
								select {
								case <-time.After(granularity):
									t.Logf("Advancing mock time by %v", granularity)
									ts.Advance(granularity)
								case <-done:
									// do nothing, it is not waiting
								}
							}()

							_m := mocked.Wait(ctx)
							close(done)
							mLatency = time.Since(started).Round(time.Millisecond)
							debug(t, "Wait elapsed: %v, mocked err: %v", mLatency, _m)
							m = _m == nil
							return nil
						})
						_ = g.Wait()

						assertLimitersAgree(t, "Wait", round, event, a, w, m, nil)
						logLatency(t, round, event, timeout, aLatency, wLatency, mLatency, nil)
						compressedReplay[round][event] = func(t *testing.T) {
							// need a mocked-time context, or the real deadline will not match the mocked deadline
							ctx := &timesourceContext{
								ts:       compressedTS,
								deadline: compressedTS.Now().Add(granularity + (granularity / 2)),
							}

							done := make(chan struct{})
							go func() {
								compressedTS.BlockUntil(1) // need Wait to block
								select {
								case <-done:
									// do nothing, wait did not block.
									// this goroutine may leak if no blockers occur. it's "fine".
								default:
									t.Logf("Advancing mock time by %v", granularity)
									compressedTS.Advance(granularity)
								}
							}()

							compressedStarted := compressedTS.Now()
							_c := compressed.Wait(ctx)
							close(done)
							cLatency := time.Since(started).Round(time.Millisecond)
							cLatency = compressedTS.Since(compressedStarted)
							debug(t, "Wait elapsed: %v, compressed err: %v", cLatency, _c)
							c := _c == nil
							assertLimitersAgree(t, "Wait", round, event, a, w, m, &c)
							logLatency(t, round, event, timeout, aLatency, wLatency, mLatency, &cLatency)
						}
						cancel()
					case "_":
						// do nothing
					default:
						panic("unhandled case: " + calls[round][event])
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
	return t.deadline, !t.deadline.IsZero()
}

func (t *timesourceContext) Done() <-chan struct{} {
	// not currently expected to be used, but it would look like this:
	t.mut.Lock()
	defer t.mut.Unlock()
	c := make(chan struct{})
	delay := t.deadline.Sub(t.ts.Now())
	t.ts.AfterFunc(delay, func() {
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
