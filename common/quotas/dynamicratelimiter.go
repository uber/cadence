package quotas

import (
	"context"

	"golang.org/x/time/rate"
)

// DynamicRateLimiter implements a dynamic config wrapper around the rate limiter
type DynamicRateLimiter struct {
	rps RPSFunc
	rl  *RateLimiter
}

// DynamicRateLimiterFactory creates a factory function for creating DynamicRateLimiters
func DynamicRateLimiterFactory(rps RPSKeyFunc) func(string) Limiter {
	return func(key string) Limiter {
		return NewDynamicRateLimiter(func() float64 { return rps(key) })
	}
}

// NewDynamicRateLimiter returns a rate limiter which handles dynamic config
func NewDynamicRateLimiter(rps RPSFunc) *DynamicRateLimiter {
	initialRps := rps()
	rl := NewRateLimiter(&initialRps, _defaultRPSTTL, _burstSize)
	return &DynamicRateLimiter{rps, rl}
}

// Allow immediately returns with true or false indicating if a rate limit
// token is available or not
func (d *DynamicRateLimiter) Allow() bool {
	rps := d.rps()
	d.rl.UpdateMaxDispatch(&rps)
	return d.rl.Allow()
}

// Wait waits up till deadline for a rate limit token
func (d *DynamicRateLimiter) Wait(ctx context.Context) error {
	rps := d.rps()
	d.rl.UpdateMaxDispatch(&rps)
	return d.rl.Wait(ctx)
}

// Reserve reserves a rate limit token
func (d *DynamicRateLimiter) Reserve() *rate.Reservation {
	rps := d.rps()
	d.rl.UpdateMaxDispatch(&rps)
	return d.rl.Reserve()
}
