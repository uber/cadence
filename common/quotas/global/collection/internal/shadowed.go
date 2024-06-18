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
	"sync"

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

	var wg sync.WaitGroup
	wg.Add(2)
	// wait on both limiters, but give up as soon as the primary one has a result.
	// it's not possible to get precisely-correct counts, but this feels pretty close.
	//
	// alternatively, calling `shadow.Allow()` if the primary succeeds might also
	// be good enough, and it'd be a bit more efficient.  it will be unable to
	// "reserve" a partial token though, which might matter in some scenarios,
	// and it would also mean a different call in case that matters at some point
	// (e.g. if "which API was called" counting is added).
	go func() {
		defer wg.Done()
		defer cancel()
		err = s.primary.Wait(ctx)
	}()
	go func() {
		defer wg.Done()
		_ = s.shadow.Wait(ctx)
	}()
	wg.Wait()

	return err
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
