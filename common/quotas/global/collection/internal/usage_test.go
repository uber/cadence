package internal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	"github.com/uber/cadence/common/clock"
)

func TestUsage(t *testing.T) {
	t.Run("tracks allow", func(t *testing.T) {
		ts := clock.NewMockedTimeSource()
		counted := NewCountedLimiter(clock.NewMockRatelimiter(ts, 1, 1))

		assert.True(t, counted.Allow(), "should match wrapped limiter")
		assert.Equal(t, UsageMetrics{1, 0, 0}, counted.Collect())

		assert.False(t, counted.Allow(), "should match wrapped limiter")
		assert.Equal(t, UsageMetrics{0, 1, 0}, counted.Collect(), "previous collect should have reset counts, and should now have just a reject")
	})
	t.Run("tracks wait", func(t *testing.T) {
		ts := clock.NewMockedTimeSource()
		counted := NewCountedLimiter(clock.NewMockRatelimiter(ts, 1, 1))

		// consume the available token
		requireQuickly(t, 100*time.Millisecond, func() {
			assert.NoError(t, counted.Wait(context.Background()), "should match wrapped limiter")
			assert.Equal(t, UsageMetrics{1, 0, 0}, counted.Collect())
		})
		// give up before the next token arrives
		requireQuickly(t, 100*time.Millisecond, func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			assert.Error(t, counted.Wait(ctx), "should match wrapped limiter")
			assert.Equal(t, UsageMetrics{0, 1, 0}, counted.Collect(), "previous collect should have reset counts, and should now have just a reject")
		})
		// wait for the next token to arrive
		requireQuickly(t, 100*time.Millisecond, func() {
			var g errgroup.Group
			g.Go(func() error {
				// waits for token to arrive
				assert.NoError(t, counted.Wait(context.Background()), "should match wrapped limiter")
				assert.Equal(t, UsageMetrics{1, 0, 0}, counted.Collect())
				return nil
			})
			g.Go(func() error {
				time.Sleep(time.Millisecond)
				ts.Advance(time.Second) // recover one token
				return nil
			})
			assert.NoError(t, g.Wait())
		})
	})
	t.Run("tracks reserve", func(t *testing.T) {
		ts := clock.NewMockedTimeSource()
		lim := NewCountedLimiter(clock.NewMockRatelimiter(ts, 1, 1))

		r := lim.Reserve()
		assert.True(t, r.Allow(), "should have used the available burst")
		assert.Equal(t, UsageMetrics{0, 0, 1}, lim.Collect(), "allowed tokens should not be counted until they're used")

		r.Used(true)
		assert.Equal(t, UsageMetrics{1, 0, 0}, lim.Collect(), "using the token should reset idle and count allowed")

		r = lim.Reserve()
		assert.False(t, r.Allow(), "should not have a token available")
		assert.Equal(t, UsageMetrics{0, 1, 0}, lim.Collect(), "not-allowed reservations immediately count rejection")
		r.Used(false)
	})

}

// Wait-based tests can block forever if there's an issue, better to fail fast.
func requireQuickly(t *testing.T, timeout time.Duration, cb func()) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		cb()
	}()
	wait := time.NewTimer(timeout)
	defer wait.Stop()
	select {
	case <-done:
	case <-wait.C: // should be far faster
		t.Fatal("timed out waiting for callback to return")
	}
}
