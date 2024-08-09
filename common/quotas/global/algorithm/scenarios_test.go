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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeightAlgorithms(t *testing.T) {
	/*
		No algorithm can perfectly predict the future, and some issues are discovered only after trying them out on widely-varying production usage.

		So here are some samples of things we've thought of, how they behave, and maybe some comments on how we feel about them.
		They are re-implemented here in a simplified form, because it's much easier and clearer (and more long-term implement-able) to do so.

		All are bound to a single key across 3 hosts with a simple 15rps target (5 per host when balanced) (3 hosts chosen because there are many easy mistakes with 2).
	*/
	const targetRPS = 15
	max := func(a, b float64) float64 {
		if a > b {
			return a
		}
		return b
	}
	min := func(a, b float64) float64 {
		if a < b {
			return a
		}
		return b
	}
	_, _ = max, min

	roundSchedule := func(low, high int) [][][3]int {
		rounds := [][3]int{
			{high, high, high},
			{low, low, high},
			{low, high, high},
			{low, low, high},
		}
		// and repeat that cycle like 16x to let it stabilize
		all := [][][3]int{rounds} // 1
		all = append(all, rounds) // 2
		all = append(all, all...) // 4
		all = append(all, all...) // 8
		all = append(all, all...) // 16
		return all
	}

	// helper to track accept and reject rates.
	// requests must be an int-float64 (ensured above, by the round schedule)
	track := func(accepts, rejects *float64, requests int, rps float64) {
		accepted := min(float64(requests), math.Floor(rps))
		*accepts += accepted
		*rejects += float64(requests) - accepted
	}

	type data struct {
		// target for the immediate next period of time, as a sample of how target RPSes settle over time
		targetRPS [3]float64 // rps that should be allowed by the limiter hosts at the end of the last round

		// these are all "last-round aggregate" info, currently 4 call/update cycles.

		callRPS       [3]float64 // rps that was sent
		accepts       [3]float64 // actually-accepted rps per host
		rejects       [3]float64 // actually-rejected requests per host
		cumulativeRPS float64
	}
	check := func(t *testing.T, expected, actual data) {
		t.Helper()
		deltaslice := func(e, a [3]float64, msg string) {
			t.Helper()
			assert.InDeltaSlicef(t,
				[]float64{e[0], e[1], e[2]}, // InDeltaSlicef does not handle arrays, only slices
				[]float64{a[0], a[1], a[2]},
				0.1, msg+", wanted [%.2f, %.2f, %.2f] but got [%.2f, %.2f, %.2f]", // copy/paste-friendly format
				e[0], e[1], e[2],
				a[0], a[1], a[2],
			)
		}

		// sanity checks to catch missing fields
		require.NotEmptyf(t, expected.targetRPS, "expected targetRPS = [%.2f, %.2f, %.2f]", actual.targetRPS[0], actual.targetRPS[1], actual.targetRPS[2])
		require.NotEmptyf(t, expected.callRPS, "expected callRPS = [%.2f, %.2f, %.2f]", actual.callRPS[0], actual.callRPS[1], actual.callRPS[2])
		require.NotEmptyf(t, expected.accepts, "expected accepts = [%.2f, %.2f, %.2f]", actual.accepts[0], actual.accepts[1], actual.accepts[2])
		// rejects can be empty, that's not crazy
		require.NotEmptyf(t, expected.cumulativeRPS, "expected cumulativeRPS = %v", actual.cumulativeRPS)

		deltaslice(expected.targetRPS, actual.targetRPS, "target RPS does not match")
		deltaslice(expected.callRPS, actual.callRPS, "call RPS does not match")
		deltaslice(expected.accepts, actual.accepts, "wrong number of per-host accepted requests")
		deltaslice(expected.rejects, actual.rejects, "wrong number of per-host rejected requests")
		assert.InDeltaf(t, expected.cumulativeRPS, actual.cumulativeRPS, 0.1, "cumulative RPS does not match")
	}
	divide := func(ary [3]float64, by int) [3]float64 {
		return [3]float64{
			ary[0] / float64(by),
			ary[1] / float64(by),
			ary[2] / float64(by),
		}
	}

	t.Run("v1 - simple running average", func(t *testing.T) {
		// original algorithm, unfortunately has rather bad behavior with intermittent / low-volume usage.

		average := func(prev float64, curr int) float64 {
			if prev == 0 {
				// special case, also matches real implementation:
				// don't grow from zero, assume traffic is steady-ish and the first data is a good starting point.
				// otherwise a common "deploy -> data loss" scenario may mean rejecting a lot of valid traffic.
				return float64(curr)
			}
			return (prev / 2) + (float64(curr) / 2)
		}
		run := func(t *testing.T, schedule [][][3]int) (results data) {
			hostRPS := [3]float64{}
			// results.targetRPS = [3]float64{0, 0, 0} // else they're NaN
			for _, round := range schedule {
				// cleared each round so only the last round remains,
				// and divided by num-rounds at the end to get an RPS-like count.
				//
				// these are nonsense on the first round, when rps start at zero,
				// which unrealistic as data is only returned when non-zero data is submitted.
				results.accepts = [3]float64{0, 0, 0}
				results.rejects = [3]float64{0, 0, 0}
				results.callRPS = [3]float64{0, 0, 0}
				for ri, r := range round {
					// keep track of call RPS
					results.callRPS[0] += float64(r[0])
					results.callRPS[1] += float64(r[1])
					results.callRPS[2] += float64(r[2])
					// find out how many would've been accepted and track it (uses previous-round target-RPS because limiters always lag by an update cycle)
					track(&results.accepts[0], &results.rejects[0], r[0], results.targetRPS[0])
					track(&results.accepts[1], &results.rejects[1], r[1], results.targetRPS[1])
					track(&results.accepts[2], &results.rejects[2], r[2], results.targetRPS[2])
					// update per-host RPS (running average in aggregator)
					hostRPS[0] = average(hostRPS[0], r[0])
					hostRPS[1] = average(hostRPS[1], r[1])
					hostRPS[2] = average(hostRPS[2], r[2])

					// update RPS that limiters allow.
					// this is based on weighted data only, nothing directly from the current batch of data
					totalRPS := hostRPS[0] + hostRPS[1] + hostRPS[2]
					//                            vvvvvvvvvvvvvvvvvvvvv---- host's weight in cluster
					results.targetRPS[0] = targetRPS * (hostRPS[0] / totalRPS)
					results.targetRPS[1] = targetRPS * (hostRPS[1] / totalRPS)
					results.targetRPS[2] = targetRPS * (hostRPS[2] / totalRPS)

					t.Logf("round %d: "+
						"calls: [%d, %d, %d], "+
						"real: [%0.3f, %0.3f, %0.3f], "+
						"rps: [%0.3f, %0.3f, %0.3f]",
						ri,
						r[0], r[1], r[2],
						hostRPS[0], hostRPS[1], hostRPS[2],
						results.targetRPS[0], results.targetRPS[1], results.targetRPS[2])
				}

				// and divide to get accept/reject-RPS (much more scale-agnostic)
				results.accepts = divide(results.accepts, len(round))
				results.rejects = divide(results.rejects, len(round))
				results.callRPS = divide(results.callRPS, len(round))
				results.cumulativeRPS = results.targetRPS[0] + results.targetRPS[1] + results.targetRPS[2]
			}
			return
		}
		t.Run("zeros", func(t *testing.T) {
			// zeros are the worst-case scenario here, greatly favoring the non-zero values.
			check(t, data{
				targetRPS:     [3]float64{0.63, 3.57, 10.7}, // intermittent allows far too few, constant is allowed over 2x what it uses
				callRPS:       [3]float64{1.25, 2.50, 5.00}, // all at or below per-host limit at all times, not just in aggregate
				accepts:       [3]float64{0.00, 1.50, 5.00}, // this is the main problems with this algorithm: bad behavior when imbalanced but low usage.
				rejects:       [3]float64{1.25, 1.00, 0.00},
				cumulativeRPS: targetRPS, // but this is useful and simple: it always tries to enforce the "real" target, never over or under.
			}, run(t, roundSchedule(0, 5)))
		})
		t.Run("low non-zeros", func(t *testing.T) {
			// even 1 instead of 0 does quite a lot better, though still not great
			check(t, data{
				targetRPS:     [3]float64{2.21, 4.07, 8.72},
				callRPS:       [3]float64{2.00, 3.00, 5.00},
				accepts:       [3]float64{1.25, 2.25, 5.00}, // constant use behaves well, others are not great
				rejects:       [3]float64{0.75, 0.75, 0.00},
				cumulativeRPS: targetRPS,
			}, run(t, roundSchedule(1, 5)))
		})
		t.Run("balanced below", func(t *testing.T) {
			// ideal scenario for this algorithm: constant and balanced
			check(t, data{
				targetRPS:     [3]float64{5, 5, 5}, // even usage == even weights == even allowances
				callRPS:       [3]float64{3, 3, 3},
				accepts:       [3]float64{3, 3, 3}, // accepts all requests perfectly.
				rejects:       [3]float64{0, 0, 0},
				cumulativeRPS: targetRPS,
			}, run(t, roundSchedule(3, 3)))
		})
		t.Run("balanced above", func(t *testing.T) {
			check(t, data{
				targetRPS:     [3]float64{5, 5, 5},
				callRPS:       [3]float64{10, 10, 10},
				accepts:       [3]float64{5, 5, 5}, // accepts exactly their intended rps
				rejects:       [3]float64{5, 5, 5}, // rejects everything else
				cumulativeRPS: targetRPS,
			}, run(t, roundSchedule(10, 10)))
		})
		t.Run("over and under budget", func(t *testing.T) {
			// goes between over and under budget, alternating rounds:
			//   added [8, 8, 8], got: [4.733, 5.667, 8.000], rps: [3.859, 4.620, 6.522]
			//   added [1, 1, 8], got: [2.867, 3.333, 8.000], rps: [3.028, 3.521, 8.451]
			//   added [1, 8, 8], got: [1.933, 5.667, 8.000], rps: [1.859, 5.449, 7.692]
			//   added [1, 1, 8], got: [1.467, 3.333, 8.000], rps: [1.719, 3.906, 9.375]
			// low-usage keeps degrading despite there being more room for it to run than it does
			check(t, data{
				targetRPS:     [3]float64{1.72, 3.91, 9.38}, // doesn't seem too unfair at a glance
				callRPS:       [3]float64{2.75, 4.50, 8.00},
				accepts:       [3]float64{1.00, 2.00, 7.25}, // but high unfair-reject rate outside constant (less than half accepted)
				rejects:       [3]float64{1.75, 2.50, 0.75},
				cumulativeRPS: targetRPS,
			}, run(t, roundSchedule(1, 8)))
		})
	})
	t.Run("v1 nonzero tweak - running average ignoring zeros", func(t *testing.T) {
		// an idea to improve on the original: hosts never see their own degraded values, but others do.
		// this allows others to act as if the host is idle (when it has been recently), but allows a periodic spike to go through.
		//
		// this has not been deployed, it was just an idea that IMO did not pan out.
		updater := func() func(prev float64, curr int) (actual, observed float64) {
			var lastNonZero float64
			return func(prev float64, curr int) (actual, observed float64) {
				defer func() {
					if curr != 0 {
						lastNonZero = float64(curr)
					}
				}()
				if prev == 0 {
					return float64(curr), float64(curr)
				}
				actual = (prev / 2) + (float64(curr) / 2)
				observed = (lastNonZero / 2) + (float64(curr) / 2)
				// largely a hack so an abnormally low lastNonZero doesn't badly skew a weight.
				// I don't really like this whole setup because of this, it feels hard to predict.
				if actual > observed {
					observed = actual
				}
				return
			}
		}
		run := func(t *testing.T, schedule [][][3]int) (results data) {
			var rps1, rps2, rps3, obs1, obs2, obs3 float64
			u1, u2, u3 := updater(), updater(), updater()
			for _, round := range schedule {
				// clear each round so only the last round remains.
				// these are nonsense the first round, when rps are not known
				results.accepts = [3]float64{0, 0, 0}
				results.rejects = [3]float64{0, 0, 0}
				results.callRPS = [3]float64{0, 0, 0}
				for ri, r := range round {
					results.callRPS[0] += float64(r[0])
					results.callRPS[1] += float64(r[1])
					results.callRPS[2] += float64(r[2])
					// find out how many would've been accepted and track it
					track(&results.accepts[0], &results.rejects[0], r[0], results.targetRPS[0])
					track(&results.accepts[1], &results.rejects[1], r[1], results.targetRPS[1])
					track(&results.accepts[2], &results.rejects[2], r[2], results.targetRPS[2])
					// update rps and "ignore-zero" observed-rps
					rps1, obs1 = u1(rps1, r[0])
					rps2, obs2 = u2(rps2, r[1])
					rps3, obs3 = u3(rps3, r[2])
					//                                 vvvv----vvvv----------- each host uses their "imagined" value, not their real one.
					results.targetRPS[0] = targetRPS * (obs1 / (obs1 + rps2 + rps3))
					results.targetRPS[1] = targetRPS * (obs2 / (rps1 + obs2 + rps3))
					results.targetRPS[2] = targetRPS * (obs3 / (rps1 + rps2 + obs3))
					t.Logf("round %d: "+
						"calls: [%d, %d, %d], "+
						"real: [%0.3f, %0.3f, %0.3f], "+
						"observed: [%0.3f, %0.3f, %0.3f], "+
						"rps: [%0.3f, %0.3f, %0.3f]",
						ri, r[0], r[1], r[2],
						rps1, rps2, rps3,
						obs1, obs2, obs3,
						results.targetRPS[0], results.targetRPS[1], results.targetRPS[2])
				}
				results.accepts = divide(results.accepts, len(round))
				results.rejects = divide(results.rejects, len(round))
				results.callRPS = divide(results.callRPS, len(round))
				results.cumulativeRPS = results.targetRPS[0] + results.targetRPS[1] + results.targetRPS[2]
			}
			return
		}
		t.Run("zeros", func(t *testing.T) {
			// handles zeros pretty well, the intermittent host allows most of the requests
			check(t, data{
				targetRPS:     [3]float64{4.09, 4.79, 10.7},
				callRPS:       [3]float64{1.25, 2.50, 5.00},
				accepts:       [3]float64{1.00, 2.00, 5.00}, // quite good accept/reject rates despite zeros
				rejects:       [3]float64{0.25, 0.50, 0.00},
				cumulativeRPS: targetRPS + 4.6, // may allow moderate bit over desired RPS
			}, run(t, roundSchedule(0, 5)))
		})
		t.Run("low nonzeros", func(t *testing.T) {
			// nonzeros are *nearly* the same as the original.  not awful but not good.
			check(t, data{
				targetRPS:     [3]float64{2.21, 4.87, 8.72},
				callRPS:       [3]float64{2.00, 3.00, 5.00},
				accepts:       [3]float64{1.25, 2.50, 5.00}, // almost the same as original
				rejects:       [3]float64{0.75, 0.50, 0.00},
				cumulativeRPS: targetRPS + 0.8, // close but still a bit above
			}, run(t, roundSchedule(1, 5)))
		})
		// with constant load it's always identical to the original.
		t.Run("balanced below", func(t *testing.T) {
			// same behavior as original
			check(t, data{
				targetRPS:     [3]float64{5, 5, 5},
				callRPS:       [3]float64{3, 3, 3},
				accepts:       [3]float64{3, 3, 3},
				rejects:       [3]float64{0, 0, 0},
				cumulativeRPS: targetRPS,
			}, run(t, roundSchedule(3, 3)))
		})
		t.Run("balanced above", func(t *testing.T) {
			// same behavior as original
			check(t, data{
				targetRPS:     [3]float64{5, 5, 5},
				callRPS:       [3]float64{10, 10, 10},
				accepts:       [3]float64{5, 5, 5},
				rejects:       [3]float64{5, 5, 5},
				cumulativeRPS: targetRPS,
			}, run(t, roundSchedule(10, 10)))
		})
	})
	t.Run("v1 limiter tweak - running average allowing portion of unused rps", func(t *testing.T) {
		// same update math as the blind running average,
		// but unused RPS get allocated to things below their local limit.
		//
		// the main goal here is to prevent pathological behavior when usage is low, infrequent, and imbalanced,
		// as weight alone can badly degrade a lower-volume host until it rejects clearly-below-limit traffic.
		//
		// there are no doubt many other ways to address this, but this one is quite simple and does not change
		// any core logic, it just lets low-RPS limiters "cheat fairly".
		// hopefully this behaves more like what we want, but the extra complexity will need more monitoring
		// to understand it whenever it fails to do so.

		average := func(prev float64, curr int) float64 {
			// same as original
			if prev == 0 {
				return float64(curr)
			}
			return (prev / 2) + (float64(curr) / 2)
		}
		run := func(t *testing.T, schedule [][][3]int) (results data) {
			var rps1, rps2, rps3 float64
			for _, round := range schedule {
				// clear each round so only the last round remains.
				// these are nonsense the first round, when rps are not known
				results.accepts = [3]float64{0, 0, 0}
				results.rejects = [3]float64{0, 0, 0}
				results.callRPS = [3]float64{0, 0, 0}
				for ri, r := range round {
					// all the same as the original
					results.callRPS[0] += float64(r[0])
					results.callRPS[1] += float64(r[1])
					results.callRPS[2] += float64(r[2])
					track(&results.accepts[0], &results.rejects[0], r[0], results.targetRPS[0])
					track(&results.accepts[1], &results.rejects[1], r[1], results.targetRPS[1])
					track(&results.accepts[2], &results.rejects[2], r[2], results.targetRPS[2])
					rps1 = average(rps1, r[0])
					rps2 = average(rps2, r[1])
					rps3 = average(rps3, r[2])
					allRps := rps1 + rps2 + rps3
					fair0 := targetRPS * (rps1 / allRps)
					fair1 := targetRPS * (rps2 / allRps)
					fair2 := targetRPS * (rps3 / allRps)

					// and now: allow each host to use part of the unused rps (each host decides blindly because they're distributed)
					//
					// this can be done entirely on the limiter side, comparing either:
					// - (targetRPS - usedRPS) / numHostsInRing    (does not require change in exchanged data currently, stays near limit)
					// - (targetRPS - usedRPS) / numHostsInvolved  (requires additional data to be returned, can approach 2x limit)
					//
					// it could also be done on agg side, but the limiter already needs to know rps and this does not *quite* need it in the agg.
					splitUnused := max(0, (targetRPS-allRps)/3)
					local := targetRPS / 3.0
					boostUnused := func(fair float64, used float64) float64 {
						if fair < local {
							// if allowed less than fallback local value, add unused up to local fallback.
							// this helps low-weight hosts accept requests when there is unused RPS.
							return min(local, fair+splitUnused)
						}
						// weighted higher than the average, trust the rps we get
						return fair
					}

					// the only change on top of the original
					results.targetRPS[0] = boostUnused(fair0, float64(r[0]))
					results.targetRPS[1] = boostUnused(fair1, float64(r[1]))
					results.targetRPS[2] = boostUnused(fair2, float64(r[2]))
					t.Logf("round %d: "+
						"calls: [%d, %d, %d], "+
						"real: [%0.3f, %0.3f, %0.3f], "+
						"unused: %0.3f (split %0.3f), "+
						"original rps: [%0.3f, %0.3f, %0.3f], "+
						"adjusted rps: [%0.3f, %0.3f, %0.3f] ",
						ri,
						r[0], r[1], r[2],
						rps1, rps2, rps3,
						targetRPS-allRps, splitUnused,
						fair0, fair1, fair2, // same as original logic
						results.targetRPS[0], results.targetRPS[1], results.targetRPS[2])
				}
				results.accepts = divide(results.accepts, len(round))
				results.rejects = divide(results.rejects, len(round))
				results.callRPS = divide(results.callRPS, len(round))
				results.cumulativeRPS = results.targetRPS[0] + results.targetRPS[1] + results.targetRPS[2]
			}
			return
		}

		t.Run("zeros", func(t *testing.T) {
			check(t, data{
				targetRPS:     [3]float64{3.38, 5.00, 10.7},
				callRPS:       [3]float64{1.25, 2.50, 5.00},
				accepts:       [3]float64{0.75, 2.50, 5.00}, // very close, a main motivating reason for this approach
				rejects:       [3]float64{0.50, 0.00, 0.00},
				cumulativeRPS: targetRPS + 4.1, // when very low, allows moderate bit above
			}, run(t, roundSchedule(0, 5)))
		})
		t.Run("low nonzeros", func(t *testing.T) {

			check(t, data{
				targetRPS:     [3]float64{4.34, 5.00, 8.72},
				callRPS:       [3]float64{2.00, 3.00, 5.00},
				accepts:       [3]float64{1.75, 3.00, 5.00}, // almost perfect, a main motivating reason for this approach
				rejects:       [3]float64{0.25, 0.00, 0.00},
				cumulativeRPS: targetRPS + 3.1, // less overage but still moderate
			}, run(t, roundSchedule(1, 5)))
		})
		t.Run("balanced below", func(t *testing.T) {
			// low balanced usage is fine, just splits evenly.
			// specifically: fair == actual+boost, which feels nicely balanced.

			check(t, data{
				targetRPS:     [3]float64{5, 5, 5},
				callRPS:       [3]float64{3, 3, 3},
				accepts:       [3]float64{3, 3, 3},
				rejects:       [3]float64{0, 0, 0},
				cumulativeRPS: targetRPS, // balanced use == limits to target
			}, run(t, roundSchedule(3, 3)))
		})
		t.Run("balanced above", func(t *testing.T) {
			check(t, data{
				targetRPS:     [3]float64{5, 5, 5},
				callRPS:       [3]float64{10, 10, 10},
				accepts:       [3]float64{5, 5, 5}, // balanced + over = fair
				rejects:       [3]float64{5, 5, 5},
				cumulativeRPS: targetRPS, // over use == limits to target
			}, run(t, roundSchedule(10, 10)))
		})
		t.Run("over and under budget", func(t *testing.T) {
			// goes between over and under budget each round:
			//   added [8, 8, 8], usage: [4.733, 5.667, 8.000], unused: -3.400 (split 0.000), fair rps: [3.859, 4.620, 6.522], adjusted rps: [3.859, 4.620, 6.522]
			//   added [1, 1, 8], usage: [2.867, 3.333, 8.000], unused:  0.800 (split 0.267), fair rps: [3.028, 3.521, 8.451], adjusted rps: [3.295, 3.788, 8.451]
			//   added [1, 8, 8], usage: [1.933, 5.667, 8.000], unused: -0.600 (split 0.000), fair rps: [1.859, 5.449, 7.692], adjusted rps: [1.859, 5.449, 7.692]
			//   added [1, 1, 8], usage: [1.467, 3.333, 8.000], unused:  2.200 (split 0.733), fair rps: [1.719, 3.906, 9.375], adjusted rps: [2.452, 4.640, 9.375]
			// the important part of this one is that the positive-unused values go to the lowest weights,
			// on the assumption that the higher-weighted ones do not need it (their overage is already accounted for by their weight, note ^ 9.3 is above 8)
			check(t, data{
				targetRPS:     [3]float64{2.45, 4.65, 9.38},
				callRPS:       [3]float64{2.75, 4.50, 8.00},
				accepts:       [3]float64{1.25, 2.25, 7.25}, // slightly better than orig
				rejects:       [3]float64{1.50, 2.25, 0.75},
				cumulativeRPS: targetRPS + 1.4, // allows moderate bit above as it's low after the calls
			}, run(t, roundSchedule(1, 8)))
		})
	})
}
