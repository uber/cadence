package algorithm

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeightAlgorithms(t *testing.T) {
	/*
		No algorithm can perfectly predict the future, and some issues are discovered only after trying them out on widely-varying production usage.

		So here are some samples of things we've thought of, how they behave, and maybe some comments on how we feel about them.
		They are re-implemented here in a simplified form, because it's much easier and clearer (and more long-term implement-able) to do so.

		All are bound to a single key across 3 hosts with a simple 15rps target (5 per host when balanced) (3 hosts chosen because there are many easy mistakes with 2).
	*/
	const rps = 15
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

	roundSchedule := func(low, high int) [][][3]float64 {
		rounds := [][3]float64{
			{float64(high), float64(high), float64(high)},
			{float64(low), float64(low), float64(high)},
			{float64(low), float64(high), float64(high)},
			{float64(low), float64(low), float64(high)}, // TODO: test all this with much longer 1-sequences.  likely deserves a DSL.
		}
		// and repeat that cycle like 16x to let it stabilize
		all := [][][3]float64{rounds} // 1
		all = append(all, rounds)     // 2
		all = append(all, all...)     // 4
		all = append(all, all...)     // 8
		all = append(all, all...)     // 16
		return all
	}

	// helper to track accept and reject rates.
	// requests must be an int-float64 (ensured above, by the round schedule)
	track := func(accepts, rejects *float64, requests, rps float64) {
		accepted := min(requests, math.Floor(rps))
		*accepts += accepted
		*rejects += requests - accepted
	}

	t.Run("v1 - simple running average", func(t *testing.T) {
		// original algorithm, unfortunately has rather bad behavior with intermittent / low-volume usage.

		update := func(prev float64, curr float64) float64 {
			if prev == 0 {
				return curr
			}
			return (prev / 2) + (curr / 2)
		}
		run := func(t *testing.T, low, high int) (h1Allowed, h2Allowed, h3Allowed float64, accepts, rejects [3]float64) {
			var rps1, rps2, rps3 float64
			for _, round := range roundSchedule(low, high) {
				// clear each round so only the last round remains.
				// these are nonsense on the first round, when rps start at zero (unrealistic)
				accepts = [3]float64{0, 0, 0}
				rejects = [3]float64{0, 0, 0}
				for ri, r := range round {
					h1, h2, h3 := r[0], r[1], r[2]
					// find out how many would've been accepted and track it (this is what happens in the limiters, as they do this based on previous-round data)
					track(&accepts[0], &rejects[0], h1, h1Allowed)
					track(&accepts[1], &rejects[1], h2, h2Allowed)
					track(&accepts[2], &rejects[2], h3, h3Allowed)
					// update rps
					rps1 = update(rps1, h1)
					rps2 = update(rps2, h2)
					rps3 = update(rps3, h3)
					// figure out what to return
					h1Allowed = rps * (rps1 / (rps1 + rps2 + rps3))
					//                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^---- host's weight in cluster
					h2Allowed = rps * (rps2 / (rps1 + rps2 + rps3))
					h3Allowed = rps * (rps3 / (rps1 + rps2 + rps3))
					t.Logf("round %d added [%.0f, %.0f, %.0f], "+
						"got: [%0.3f, %0.3f, %0.3f], "+
						"rps: [%0.3f, %0.3f, %0.3f]",
						ri, h1, h2, h3,
						rps1, rps2, rps3,
						h1Allowed, h2Allowed, h3Allowed)
				}
			}
			return
		}
		t.Run("zeros", func(t *testing.T) {
			// zeros are the worst-case scenario here, greatly favoring the non-zero values.
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 0, 5)
			assert.InDeltaf(t, 0.625, h1Allowed, 0.1, "intermittent host doesn't even allow one per second, well below actual usage")
			assert.InDeltaf(t, 3.57, h2Allowed, 0.1, "alternating host allows better than average usage")
			assert.InDeltaf(t, 10.7, h3Allowed, 0.1, "constant host allows all plus lots of room for more")
			assert.Equal(t, [3]float64{0, 6, 20}, accepts, "bad accept outside constant")                           // these are the main problem
			assert.Equal(t, [3]float64{5, 4, 0}, rejects, "high reject outside constant")                           // with this algorithm
			assert.InDeltaf(t, rps, h1Allowed+h2Allowed+h3Allowed, 0.1, "total allowed rps does not exceed target") // a useful quality of this approach, it stays at the limit in all scenarios
		})
		t.Run("low non-zeros", func(t *testing.T) {
			// even 1 instead of 0 does quite a lot better, though still not great
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 1, 5)
			assert.InDeltaf(t, 2.21, h1Allowed, 0.1, "intermittent host allows well below ideal")
			assert.InDeltaf(t, 4.07, h2Allowed, 0.1, "alternating host allows most but still below ideal")
			assert.InDeltaf(t, 8.72, h3Allowed, 0.1, "constant host allows all plus room for more")
			assert.Equal(t, [3]float64{5, 9, 20}, accepts, "most accepted, all in constant")
			assert.Equal(t, [3]float64{3, 3, 0}, rejects, "some rejects")
			assert.InDeltaf(t, rps, h1Allowed+h2Allowed+h3Allowed, 0.1, "total allowed rps does not exceed target")
		})
		t.Run("balanced below", func(t *testing.T) {
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 3, 3)
			assert.InDeltaf(t, 5, h1Allowed, 0.1, "allows exactly a third when load is evenly balanced")
			assert.InDeltaf(t, 5, h2Allowed, 0.1, "allows exactly a third when load is evenly balanced")
			assert.InDeltaf(t, 5, h3Allowed, 0.1, "allows exactly a third when load is evenly balanced")
			assert.Equal(t, [3]float64{12, 12, 12}, accepts, "all accepted")
			assert.Equal(t, [3]float64{0, 0, 0}, rejects, "no rejects")
			assert.InDeltaf(t, rps, h1Allowed+h2Allowed+h3Allowed, 0.1, "total allowed rps does not exceed target")
		})
		t.Run("balanced above", func(t *testing.T) {
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 10, 10)
			assert.InDeltaf(t, 5, h1Allowed, 0.1, "allows exactly a third when load is evenly balanced")
			assert.InDeltaf(t, 5, h2Allowed, 0.1, "allows exactly a third when load is evenly balanced")
			assert.InDeltaf(t, 5, h3Allowed, 0.1, "allows exactly a third when load is evenly balanced")
			assert.Equal(t, [3]float64{20, 20, 20}, accepts, "balanced == accept at 5rps each")
			assert.Equal(t, [3]float64{20, 20, 20}, rejects, "balanced == reject at 5rps each")
			assert.InDeltaf(t, rps, h1Allowed+h2Allowed+h3Allowed, 0.1, "total allowed rps does not exceed target")
		})
		t.Run("over and under budget", func(t *testing.T) {
			// goes between over and under budget, alternating rounds:
			//   added [8, 8, 8], got: [4.733, 5.667, 8.000], rps: [3.859, 4.620, 6.522]
			//   added [1, 1, 8], got: [2.867, 3.333, 8.000], rps: [3.028, 3.521, 8.451]
			//   added [1, 8, 8], got: [1.933, 5.667, 8.000], rps: [1.859, 5.449, 7.692]
			//   added [1, 1, 8], got: [1.467, 3.333, 8.000], rps: [1.719, 3.906, 9.375]
			// low-usage keeps degrading despite there being more room for it to run than it does
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 1, 8)
			assert.InDeltaf(t, 1.7, h1Allowed, 0.1, "allows a small number when under-budget")
			assert.InDeltaf(t, 3.9, h2Allowed, 0.1, "moderate boost above ")
			assert.InDeltaf(t, 9.38, h3Allowed, 0.1, "would allow next request completely")
			assert.Equal(t, [3]float64{4, 8, 29}, accepts, "mediocre accept outside constant host")
			assert.Equal(t, [3]float64{7, 10, 3}, rejects, "high reject rates outside constant")
			assert.InDeltaf(t, 15, h1Allowed+h2Allowed+h3Allowed, 0.1, "should be precisely at limit")
		})
	})
	t.Run("v1 nonzero tweak - running average ignoring zeros", func(t *testing.T) {
		// an idea to improve on the original: hosts never see their own degraded values, but others do.
		// this allows others to act as if the host is idle (when it has been recently), but allows a periodic spike to go through.
		updater := func() func(prev float64, curr float64) (actual, observed float64) {
			var lastNonZero float64
			return func(prev float64, curr float64) (actual, observed float64) {
				defer func() {
					if curr != 0 {
						lastNonZero = curr
					}
				}()
				if prev == 0 {
					return curr, curr
				}
				actual = (prev / 2) + (curr / 2)
				observed = (float64(lastNonZero) / 2) + (curr / 2)
				// largely a hack so an abnormally low lastNonZero doesn't badly skew a weight.
				// I don't really like this whole setup because of this.
				if actual > observed {
					observed = actual
				}
				return
			}
		}
		run := func(t *testing.T, low, high int) (h1Allowed, h2Allowed, h3Allowed float64, accepts, rejects [3]float64) {
			var rps1, rps2, rps3, obs1, obs2, obs3 float64
			u1, u2, u3 := updater(), updater(), updater()
			for _, round := range roundSchedule(low, high) {
				// clear each round so only the last round remains.
				// these are nonsense the first round, when rps are not known
				accepts = [3]float64{0, 0, 0}
				rejects = [3]float64{0, 0, 0}
				for ri, r := range round {
					h1, h2, h3 := r[0], r[1], r[2]
					// find out how many would've been accepted and track it
					track(&accepts[0], &rejects[0], h1, h1Allowed)
					track(&accepts[1], &rejects[1], h2, h2Allowed)
					track(&accepts[2], &rejects[2], h3, h3Allowed)
					// update rps
					rps1, obs1 = u1(rps1, h1)
					rps2, obs2 = u2(rps2, h2)
					rps3, obs3 = u3(rps3, h3)
					h1Allowed = rps * (obs1 / (obs1 + rps2 + rps3))
					//                 ^^^^----^^^^----------- each host uses their "imagined" value, not their real one.
					h2Allowed = rps * (obs2 / (rps1 + obs2 + rps3))
					h3Allowed = rps * (obs3 / (rps1 + rps2 + obs3))
					t.Logf("round %d added [%.0f, %.0f, %.0f], got "+
						"real: [%0.3f, %0.3f, %0.3f], "+
						"observed: [%0.3f, %0.3f, %0.3f], "+
						"rps: [%0.3f, %0.3f, %0.3f]",
						ri, h1, h2, h3,
						rps1, rps2, rps3,
						obs1, obs2, obs3,
						h1Allowed, h2Allowed, h3Allowed)
				}
			}
			return
		}
		t.Run("zeros", func(t *testing.T) {
			// handles zeros pretty well, the intermittent host allows most of the requests
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 0, 5)
			assert.InDeltaf(t, 4.09, h1Allowed, 0.1, "intermittent host allows most of the requests")
			assert.InDeltaf(t, 4.79, h2Allowed, 0.1, "alternating host allows rps near its average")
			assert.InDeltaf(t, 10.71, h3Allowed, 0.1, "constant host allows rps well above what it receives")
			assert.Equal(t, [3]float64{4, 8, 20}, accepts, "nearly ideal accept")
			assert.Equal(t, [3]float64{1, 2, 0}, rejects, "low reject")
			assert.InDeltaf(t, 19.6, h1Allowed+h2Allowed+h3Allowed, 0.1, "may allow a fair bit more requests than target RPS")
		})
		t.Run("low nonzeros", func(t *testing.T) {
			// nonzeros are *nearly* the same as the original.  not awful but not good.
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 1, 5)
			assert.InDeltaf(t, 2.21, h1Allowed, 0.1, "same as original")
			assert.InDeltaf(t, 4.87, h2Allowed, 0.1, "slightly above original, due to last-non-zero being higher than average more often")
			assert.InDeltaf(t, 8.72, h3Allowed, 0.1, "same as original")
			assert.Equal(t, [3]float64{5, 10, 20}, accepts, "one better than original")
			assert.Equal(t, [3]float64{3, 2, 0}, rejects, "one better than original")
			assert.InDeltaf(t, 15.79, h1Allowed+h2Allowed+h3Allowed, 0.1, "slightly above normal due to last-non-zero hack")
		})
		// with constant load it's always identical to the original.
		t.Run("balanced below", func(t *testing.T) {
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 3, 3)
			assert.InDeltaf(t, 5, h1Allowed, 0.1, "same as original")
			assert.InDeltaf(t, 5, h2Allowed, 0.1, "same as original")
			assert.InDeltaf(t, 5, h3Allowed, 0.1, "same as original")
			assert.Equal(t, [3]float64{12, 12, 12}, accepts, "same as original")
			assert.Equal(t, [3]float64{0, 0, 0}, rejects, "same as original")
			assert.InDeltaf(t, rps, h1Allowed+h2Allowed+h3Allowed, 0.1, "same as original")
		})
		t.Run("balanced above", func(t *testing.T) {
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 10, 10)
			assert.InDeltaf(t, 5, h1Allowed, 0.1, "same as original")
			assert.InDeltaf(t, 5, h2Allowed, 0.1, "same as original")
			assert.InDeltaf(t, 5, h3Allowed, 0.1, "same as original")
			assert.Equal(t, [3]float64{20, 20, 20}, accepts, "same as original")
			assert.Equal(t, [3]float64{20, 20, 20}, rejects, "same as original")
			assert.InDeltaf(t, rps, h1Allowed+h2Allowed+h3Allowed, 0.1, "same as original")
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
		update := func(prev float64, curr float64) float64 {
			if prev == 0 {
				return curr
			}
			return (prev / 2) + (curr / 2)
		}
		run := func(t *testing.T, low, high int) (h1Allowed, h2Allowed, h3Allowed float64, accepts, rejects [3]float64) {
			var rps1, rps2, rps3 float64
			for _, round := range roundSchedule(low, high) {
				// clear each round so only the last round remains.
				// these are nonsense the first round, when rps are not known
				accepts = [3]float64{0, 0, 0}
				rejects = [3]float64{0, 0, 0}
				for ri, r := range round {
					h1, h2, h3 := r[0], r[1], r[2]
					// find out how many would've been accepted and track it
					track(&accepts[0], &rejects[0], h1, h1Allowed)
					track(&accepts[1], &rejects[1], h2, h2Allowed)
					track(&accepts[2], &rejects[2], h3, h3Allowed)
					// update the rps
					rps1 = update(rps1, h1)
					rps2 = update(rps2, h2)
					rps3 = update(rps3, h3)
					allRps := rps1 + rps2 + rps3
					// calculate fair weighted rps (same as original)
					h1Fair := rps * (rps1 / allRps)
					h2Fair := rps * (rps2 / allRps)
					h3Fair := rps * (rps3 / allRps)

					// and now: allow each host to use part of the unused rps (each host decides blindly because they're distributed)
					//
					// this would be done on the limiter side, comparing either:
					// - (targetRPS - usedRPS) / numHostsInRing    (does not require change in exchanged data currently, stays near limit)
					// - (targetRPS - usedRPS) / numHostsInvolved  (requires additional data to be returned, can approach 2x limit)
					//
					// it could also be done on agg side, but the limiter already needs to know rps and this does not *quite* need it in the agg.
					// I'm not sure that moving this into the agg helps anything, and it'd need more state tracking to hold "real" and "allowed"
					// rps values separately, so this will just stay in the limiter as part of a collection update cycle.
					splitUnused := max(0, (rps-allRps)/3)
					local := rps / 3.0
					boostUnused := func(fair float64, used float64) float64 {
						if fair < local {
							// if allowed less than fallback local value, add unused up to local.
							// this helps low-weight hosts recover when usage is low.
							return min(local, fair+splitUnused)
						} else {
							// weighted higher than the average, trust the rps we get
							return fair
						}
					}

					// the only change on top of the original
					h1Allowed = boostUnused(h1Fair, h1)
					h2Allowed = boostUnused(h2Fair, h2)
					h3Allowed = boostUnused(h3Fair, h3)
					t.Logf("round %d added [%.0f, %.0f, %.0f], "+
						"got usage: [%0.3f, %0.3f, %0.3f], "+
						"unused: %0.3f (split %0.3f), "+
						"fair rps: [%0.3f, %0.3f, %0.3f], "+
						"adjusted rps: [%0.3f, %0.3f, %0.3f] ",
						ri, h1, h2, h3,
						rps1, rps2, rps3,
						rps-allRps, splitUnused,
						h1Fair, h2Fair, h3Fair,
						h1Allowed, h2Allowed, h3Allowed)
				}
			}
			return
		}
		t.Run("zeros", func(t *testing.T) {
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 0, 5)
			assert.InDeltaf(t, 3.38, h1Allowed, 0.1, "intermittent host allows most of the requests")
			assert.InDeltaf(t, 5.0, h2Allowed, 0.1, "alternating host allows all while usage is low")
			assert.InDeltaf(t, 10.7, h3Allowed, 0.1, "constant host can grow if needed")
			assert.Equal(t, [3]float64{3, 10, 20}, accepts, "very good accept") // this is the motivating reason for this algorithm
			assert.Equal(t, [3]float64{2, 0, 0}, rejects, "just 2 rejects")     // along with this
			assert.InDeltaf(t, 19.1, h1Allowed+h2Allowed+h3Allowed, 0.1, "may burst above limits until rebalanced")
		})
		t.Run("low nonzeros", func(t *testing.T) {
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 1, 5)
			assert.InDeltaf(t, 4.34, h1Allowed, 0.1, "allows most")
			assert.InDeltaf(t, 5, h2Allowed, 0.1, "allows all")
			assert.InDeltaf(t, 8.72, h3Allowed, 0.1, "constant host allows rps well above what it receives")
			assert.Equal(t, [3]float64{7, 12, 20}, accepts, "almost perfect accept") // this is the motivating reason for this algorithm
			assert.Equal(t, [3]float64{1, 0, 0}, rejects, "just one reject")         // along with this
			assert.InDeltaf(t, 18.1, h1Allowed+h2Allowed+h3Allowed, 0.1, "may burst above limits until rebalanced")
		})
		t.Run("balanced below", func(t *testing.T) {
			// low balanced usage is fine, just splits evenly.
			// specifically: fair == actual+boost, which feels nicely balanced.
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 3, 3)
			assert.InDeltaf(t, 5, h1Allowed, 0.1, "exactly a third when balanced")
			assert.InDeltaf(t, 5, h2Allowed, 0.1, "exactly a third when balanced")
			assert.InDeltaf(t, 5, h3Allowed, 0.1, "exactly a third when balanced")
			assert.Equal(t, [3]float64{12, 12, 12}, accepts, "same as original")
			assert.Equal(t, [3]float64{0, 0, 0}, rejects, "same as original")
			assert.InDeltaf(t, rps, h1Allowed+h2Allowed+h3Allowed, 0.1, "should be precisely at limit")
		})
		t.Run("balanced above", func(t *testing.T) {
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 10, 10)
			assert.InDeltaf(t, 5, h1Allowed, 0.1, "exactly a third when balanced + over limit")
			assert.InDeltaf(t, 5, h2Allowed, 0.1, "exactly a third when balanced + over limit")
			assert.InDeltaf(t, 5, h3Allowed, 0.1, "exactly a third when balanced + over limit")
			assert.Equal(t, [3]float64{20, 20, 20}, accepts, "same as original")
			assert.Equal(t, [3]float64{20, 20, 20}, rejects, "same as original")
			assert.InDeltaf(t, rps, h1Allowed+h2Allowed+h3Allowed, 0.1, "should be precisely at limit")
		})
		t.Run("over and under budget", func(t *testing.T) {
			// goes between over and under budget each round:
			//   added [8, 8, 8], usage: [4.733, 5.667, 8.000], unused: -3.400 (split 0.000), fair rps: [3.859, 4.620, 6.522], adjusted rps: [3.859, 4.620, 6.522]
			//   added [1, 1, 8], usage: [2.867, 3.333, 8.000], unused:  0.800 (split 0.267), fair rps: [3.028, 3.521, 8.451], adjusted rps: [3.295, 3.788, 8.451]
			//   added [1, 8, 8], usage: [1.933, 5.667, 8.000], unused: -0.600 (split 0.000), fair rps: [1.859, 5.449, 7.692], adjusted rps: [1.859, 5.449, 7.692]
			//   added [1, 1, 8], usage: [1.467, 3.333, 8.000], unused:  2.200 (split 0.733), fair rps: [1.719, 3.906, 9.375], adjusted rps: [2.452, 4.640, 9.375]
			// the important part of this one is that the positive-unused values go to the lowest weights,
			// on the assumption that the higher-weighted ones do not need it (their overage is already accounted for by their weight, note ^ 9.3 is above 8)
			h1Allowed, h2Allowed, h3Allowed, accepts, rejects := run(t, 1, 8)
			assert.InDeltaf(t, 2.45, h1Allowed, 0.1, "moderately better than original")
			assert.InDeltaf(t, 4.64, h2Allowed, 0.1, "moderately better than original")
			assert.InDeltaf(t, 9.38, h3Allowed, 0.1, "same as original")
			assert.Equal(t, [3]float64{5, 9, 29}, accepts, "slightly better than original")
			assert.Equal(t, [3]float64{6, 9, 3}, rejects, "slightly better than original")
			assert.InDeltaf(t, 16.4, h1Allowed+h2Allowed+h3Allowed, 0.1, "may burst above limits until rebalanced")
		})
	})
}
