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

package ratelimited

import (
	"context"

	"golang.org/x/time/rate"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/quotas"
)

type limiterAlwaysAllow struct{}

func (l limiterAlwaysAllow) Allow() bool {
	return true
}

func (l limiterAlwaysAllow) Wait(ctx context.Context) error {
	return nil
}

func (l limiterAlwaysAllow) Reserve() clock.Reservation {
	return &reservationAlwaysAllow{}
}

func (l limiterAlwaysAllow) Limit() rate.Limit {
	return rate.Inf
}

type limiterNeverAllow struct{}

func (l limiterNeverAllow) Allow() bool {
	return false
}

func (l limiterNeverAllow) Wait(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (l limiterNeverAllow) Reserve() clock.Reservation {
	return &reservationNeverAllow{}
}

func (l limiterNeverAllow) Limit() rate.Limit {
	return 0
}

type reservationAlwaysAllow struct{}
type reservationNeverAllow struct{}

func (r reservationAlwaysAllow) Allow() bool { return true }
func (r reservationAlwaysAllow) Used(bool)   {}
func (r reservationNeverAllow) Allow() bool  { return false }
func (r reservationNeverAllow) Used(bool)    {}

var _ quotas.Limiter = (*limiterAlwaysAllow)(nil)
var _ quotas.Limiter = (*limiterNeverAllow)(nil)
var _ clock.Reservation = (*reservationAlwaysAllow)(nil)
var _ clock.Reservation = (*reservationNeverAllow)(nil)
