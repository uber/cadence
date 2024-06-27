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

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/quotas"
)

type (
	shadowedLimiter struct {
		primary quotas.Limiter
		shadow  quotas.Limiter
	}
	shadowedReservation struct {
		primary clock.Reservation
		shadow  clock.Reservation
	}
)

var _ quotas.Limiter = shadowedLimiter{}
var _ clock.Reservation = shadowedReservation{}

// NewShadowedLimiter mirrors all quotas.Limiter to its two wrapped limiters,
// but only returns results from / waits as long as the "primary" one (first arg).
//
// This is intended for when you want to use one limit but also run a second
// limit at the same time for comparison (to decide before switching) or
// to warm data (to reduce spikes when switching).
func NewShadowedLimiter(primary, secondary quotas.Limiter) quotas.Limiter {
	return shadowedLimiter{
		primary: primary,
		shadow:  secondary,
	}
}

func (s shadowedLimiter) Allow() bool {
	_ = s.shadow.Allow()
	return s.primary.Allow()
}

func (s shadowedLimiter) Wait(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		_ = s.shadow.Wait(ctx) // not waited for because it should end ~immediately on its own
	}()

	return s.primary.Wait(ctx)
}

func (s shadowedLimiter) Reserve() clock.Reservation {
	return shadowedReservation{
		primary: s.primary.Reserve(),
		shadow:  s.shadow.Reserve(),
	}
}

func (s shadowedReservation) Allow() bool {
	_ = s.shadow.Allow()
	return s.primary.Allow()
}

func (s shadowedReservation) Used(wasUsed bool) {
	s.primary.Used(wasUsed)
	s.shadow.Used(wasUsed)
}
