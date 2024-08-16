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

package quotas

import (
	"context"

	"golang.org/x/time/rate"

	"github.com/uber/cadence/common/clock"
)

// DynamicRateLimiter implements a dynamic config wrapper around the rate limiter,
// checks for updates to the dynamic config and updates the rate limiter accordingly
type DynamicRateLimiter struct {
	rps RPSFunc
	rl  *RateLimiter
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
func (d *DynamicRateLimiter) Reserve() clock.Reservation {
	rps := d.rps()
	d.rl.UpdateMaxDispatch(&rps)
	return d.rl.Reserve()
}

func (d *DynamicRateLimiter) Limit() rate.Limit {
	return rate.Limit(d.rps())
}
