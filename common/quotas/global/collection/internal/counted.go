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

package internal

import (
	"context"

	"go.uber.org/atomic"
	"golang.org/x/time/rate"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/quotas"
)

type (
	CountedLimiter struct {
		wrapped quotas.Limiter
		usage   *AtomicUsage
	}

	countedReservation struct {
		wrapped clock.Reservation
		usage   *AtomicUsage // reference to the reservation's limiter's usage
	}

	AtomicUsage struct {
		allowed, rejected, idle atomic.Int64
	}

	UsageMetrics struct {
		Allowed, Rejected, Idle int
	}
)

var _ quotas.Limiter = CountedLimiter{}
var _ clock.Reservation = countedReservation{}

func NewCountedLimiter(limiter quotas.Limiter) CountedLimiter {
	return CountedLimiter{
		wrapped: limiter,
		usage:   &AtomicUsage{},
	}
}

func (c CountedLimiter) Allow() bool {
	allowed := c.wrapped.Allow()
	c.usage.Count(allowed)
	return allowed
}

func (c CountedLimiter) Wait(ctx context.Context) error {
	err := c.wrapped.Wait(ctx)
	c.usage.Count(err == nil)
	return err
}

func (c CountedLimiter) Reserve() clock.Reservation {
	return countedReservation{
		wrapped: c.wrapped.Reserve(),
		usage:   c.usage,
	}
}

func (c CountedLimiter) Limit() rate.Limit {
	return c.wrapped.Limit()
}

func (c CountedLimiter) Collect() UsageMetrics {
	return c.usage.Collect()
}

func (c countedReservation) Allow() bool {
	return c.wrapped.Allow()
}

func (c countedReservation) Used(wasUsed bool) {
	c.wrapped.Used(wasUsed)
	c.usage.idle.Store(0)
	if c.Allow() {
		if wasUsed {
			// only counts as allowed if used, else it is hopefully rolled back.
			// this may or may not restore the token, but it does imply "this limiter did not limit the event".
			c.usage.Count(true)
		}

		// else it was canceled, and not "used".
		//
		// currently these are not tracked because some other rejection will occur
		// and be emitted in all our current uses, but with bad enough luck or
		// latency before canceling it could lead to misleading metrics.
	} else {
		// these reservations cannot be waited on so they cannot become allowed,
		// and they cannot be returned, so they are always rejected.
		//
		// specifically: it is likely that `wasUsed == Allow()`, so false cannot be
		// trusted to mean "will not use for some other reason", and the underlying
		// rate.Limiter did not change state anyway because it returned the
		// pending-token before becoming a clock.Reservation.
		c.usage.Count(false)
	}
}

func (a *AtomicUsage) Count(allowed bool) {
	if allowed {
		a.allowed.Add(1)
	} else {
		a.rejected.Add(1)
	}
	a.idle.Store(0)
}

func (a *AtomicUsage) Collect() UsageMetrics {
	m := UsageMetrics{
		Allowed:  int(a.allowed.Swap(0)),
		Rejected: int(a.rejected.Swap(0)),
	}
	if m.Allowed+m.Rejected == 0 {
		m.Idle = int(a.idle.Add(1))
	} // else 0
	return m
}
