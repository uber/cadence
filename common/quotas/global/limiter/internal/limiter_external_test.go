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
package internal_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/quotas/global/limiter/internal"
)

func TestLimiter(t *testing.T) {
	t.Run("uses fallback initially", func(t *testing.T) {
		m := quotas.NewMockLimiter(gomock.NewController(t))
		m.EXPECT().Allow().Times(1).Return(true)
		m.EXPECT().Allow().Times(2).Return(false)
		lim := internal.NewFallbackLimiter(m)

		assert.True(t, lim.Allow(), "should return fallback's first response")
		assert.False(t, lim.Allow(), "should return fallback's second response")
		assert.False(t, lim.Allow(), "should return fallback's third response")

		accept, reject, fallback := lim.Collect()
		assert.Equal(t, 1, accept, "should have counted one accepted request")
		assert.Equal(t, 2, reject, "should have counted two rejected requests")
		assert.True(t, fallback, "should be using fallback")
	})
	t.Run("uses real after update", func(t *testing.T) {
		m := quotas.NewMockLimiter(gomock.NewController(t)) // no allowances, must not be used
		lim := internal.NewFallbackLimiter(m)
		lim.Update(1_000_000) // large enough to allow millisecond sleeps to refill

		time.Sleep(time.Millisecond)
		assert.True(t, lim.Allow(), "limiter allows after enough time has passed")

		allowed := lim.Allow() // could be either due to timing
		accept, reject, fallback := lim.Collect()
		if allowed {
			assert.Equal(t, 2, accept, "should have counted two accepted request")
			assert.Equal(t, 0, reject, "should have counted no rejected requests")
		} else {
			assert.Equal(t, 1, accept, "should have counted one accepted request")
			assert.Equal(t, 1, reject, "should have counted one rejected request")
		}
		assert.False(t, fallback, "should not be using fallback")
	})

	t.Run("collecting usage data resets counts", func(t *testing.T) {
		lim := internal.NewFallbackLimiter(nil)
		lim.Update(1) // so there's no need to mock
		lim.Allow()
		accepted, rejected, _ := lim.Collect()
		assert.Equal(t, 1, accepted+rejected, "should count one request")
		accepted, rejected, _ = lim.Collect()
		assert.Zero(t, accepted+rejected, "collect should have cleared the counts")
	})

	t.Run("use-fallback fuse", func(t *testing.T) {
		// duplicate to allow this test to be external, keep in sync by hand
		const maxFailedUpdates = 9
		t.Cleanup(func() {
			if t.Failed() { // notices sub-test failures
				t.Logf("maxFailedUpdates may be out of sync (%v), check hardcoded values", maxFailedUpdates)
			}
		})

		t.Run("falls back after too many failures", func(t *testing.T) {
			m := quotas.NewMockLimiter(gomock.NewController(t)) // no allowances
			lim := internal.NewFallbackLimiter(m)
			lim.Update(1)
			_, _, fallback := lim.Collect()
			require.False(t, fallback, "should not be using fallback")
			// bucket starts out empty / with whatever contents it had before (zero).
			// this is possibly surprising, so it's asserted.
			require.False(t, lim.Allow(), "rate.Limiter should reject requests until filled")

			// fail enough times to trigger a fallback
			for i := 0; i < maxFailedUpdates; i++ {
				// build up to the edge...
				lim.FailedUpdate()
				_, _, fallback := lim.Collect()
				require.False(t, fallback, "should not be using fallback after %n failed updates", i+1)
			}
			lim.FailedUpdate() // ... and push it over
			_, _, fallback = lim.Collect()
			require.True(t, fallback, "%vth update should switch to fallback", maxFailedUpdates+1)

			m.EXPECT().Allow().Times(1).Return(true)
			assert.True(t, lim.Allow(), "should return fallback's allowed request")
		})
		t.Run("failing many times does not accidentally switch away from fallback", func(t *testing.T) {
			lim := internal.NewFallbackLimiter(nil)
			for i := 0; i < maxFailedUpdates*10; i++ {
				lim.FailedUpdate()
				_, _, fallback := lim.Collect()
				require.True(t, fallback, "should use fallback after %v failed updates", i+1)
			}
		})
	})
}

func TestLimiterNotRacy(t *testing.T) {
	lim := internal.NewFallbackLimiter(allowlimiter{})
	var g errgroup.Group
	const loops = 1000
	for i := 0; i < loops; i++ {
		// clear ~10% of the time
		if rand.Intn(10) == 0 {
			g.Go(func() error {
				lim.Clear()
				return nil
			})
		}
		// update ~10% of the time, fail the rest.
		// this should randomly clear occasionally via failures.
		if rand.Intn(10) == 0 {
			g.Go(func() error {
				lim.Update(rand.Float64()) // essentially never exercises "skip no update" logic
				return nil
			})
		} else {
			g.Go(func() error {
				lim.FailedUpdate()
				return nil
			})
		}
		// collect occasionally
		if rand.Intn(10) == 0 {
			g.Go(func() error {
				lim.Collect()
				return nil
			})
		}
		g.Go(func() error {
			lim.Allow()
			return nil
		})
	}
}

var _ internal.Limiter = allowlimiter{}

type allowlimiter struct{}

func (allowlimiter) Allow() bool { return true }
