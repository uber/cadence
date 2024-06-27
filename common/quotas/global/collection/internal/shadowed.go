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
	ShadowedLimiter struct {
		primary quotas.Limiter
		shadow  quotas.Limiter
	}
	shadowedReservation struct {
		primary clock.Reservation
		shadow  clock.Reservation
	}
)

var _ quotas.Limiter = ShadowedLimiter{}
var _ clock.Reservation = shadowedReservation{}

func NewShadowedLimiter(primary, secondary quotas.Limiter) quotas.Limiter {
	return ShadowedLimiter{
		primary: primary,
		shadow:  secondary,
	}
}

func (s ShadowedLimiter) Allow() bool {
	_ = s.shadow.Allow()
	return s.primary.Allow()
}

func (s ShadowedLimiter) Wait(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		_ = s.shadow.Wait(ctx)
	}()

	return s.primary.Wait(ctx)
}

func (s ShadowedLimiter) Reserve() clock.Reservation {
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
