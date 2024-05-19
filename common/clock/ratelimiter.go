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

package clock

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

type (
	// Ratelimiter is a test-friendly version of [golang.org/x/time/rate.Limiter],
	// which can be backed by any TimeSource.  The changes summarize as:
	//   - Wait has been removed, as it is difficult to precisely mimic, though an
	//     approximation is likely good enough. However, our Wait-using ratelimiters
	//     *only* use Wait, and they can have a drastically simplified API and
	//     implementation if they are handled separately.
	//   - All MethodAt APIs have been excluded as we do not use them, and they would
	//     have to bypass the contained TimeSource, which seems error-prone.
	//   - All `MethodN` APIs have been excluded because we do not use them, but they
	//     seem fairly easy to add if needed later.
	Ratelimiter interface {
		// Allow returns true if an operation can be performed now, i.e. there are burst-able tokens available to use.
		//
		// To perform operations at a steady rate, use Wait instead.
		Allow() bool
		// Burst returns the maximum burst size allowed by this Ratelimiter.
		// A Burst of zero returns false from Allow, unless its Limit is infinity.
		//
		// To see the number of burstable tokens available now, use Tokens.
		Burst() int
		// Limit returns the maximum overall rate of events that are allowed.
		Limit() rate.Limit
		// Reserve claims a single Allow() call that can be canceled if not used.
		Reserve() Reservation
		// SetBurst sets the Burst value
		SetBurst(newBurst int)
		// SetLimit sets the Limit value
		SetLimit(newLimit rate.Limit)
		// Tokens returns the number of immediately-allowable operations when called.
		// Values >= 1 will lead to Allow returning true.
		//
		// These values are NOT guaranteed, as other calls to Allow or Reserve may
		// occur before you are able to act on the returned value.
		Tokens() float64
		// Wait blocks until one token is available (and consumes it), or:
		//   - context is canceled
		//   - a token will never become available (burst == 0)
		//   - a token is not expected to become available in time (short deadline + many pending reservations)
		//
		// If an error is returned, regardless of the reason, you may NOT proceed with the
		// limited action you were waiting for.
		Wait(ctx context.Context) error
	}
	// Reservation is a simplified and test-friendly version of [golang.org/x/time/rate.Reservation].
	//
	// The semantics are not quite the same, as [golang.org/x/time/rate.Reservation] does not match our usage very well.
	Reservation interface {
		// Used marks this Reservation as either used or not-used, and it must be called
		// once on every Reservation:
		//  - If true, the operation is assumed to be allowed, and the Ratelimiter token
		//    will remain consumed.
		//  - If false, the Reservation will be rolled back, restoring a Ratelimiter token.
		//    This is equivalent to calling Cancel() on a golang.org/x/time/rate.Reservation
		Used(wasUsed bool)
		// Allow returns true if a request should be allowed.
		//
		// This is equivalent to `OK() == true && Delay() == 0` with [golang.org/x/time/rate.Reservation].
		Allow() bool
	}

	ratelimiter struct {
		timesource TimeSource
		limiter    *rate.Limiter
	}

	reservation struct {
		timesource TimeSource
		res        *rate.Reservation
	}
)

var (
	_ Ratelimiter = (*ratelimiter)(nil)
	_ Reservation = (*reservation)(nil)
)

// err constants to keep allocations low
var (
	// ErrCannotWait is used as the common base for two internal reasons why Ratelimiter.Wait call cannot
	// succeed, without directly exposing the reason except as an explanatory string.
	//
	// Callers should only ever care about ErrCannotWait to be resistant to racing behavior.
	// Or perhaps ideally: not care about the reason at all beyond "not a ctx.Err()".
	ErrCannotWait             = fmt.Errorf("ratelimiter.Wait cannot be satisfied")
	errWaitLongerThanDeadline = fmt.Errorf("would wait longer than ctx deadline: %w", ErrCannotWait)
	errWaitZeroBurst          = fmt.Errorf("zero burst will never allow: %w", ErrCannotWait)
)

func NewRatelimiter(lim rate.Limit, burst int) Ratelimiter {
	return &ratelimiter{
		timesource: NewRealTimeSource(),
		limiter:    rate.NewLimiter(lim, burst),
	}
}
func NewMockRatelimiter(ts TimeSource, lim rate.Limit, burst int) Ratelimiter {
	return &ratelimiter{
		timesource: ts,
		limiter:    rate.NewLimiter(lim, burst),
	}
}

func (r *ratelimiter) Allow() bool {
	return r.limiter.AllowN(r.timesource.Now(), 1)
}

func (r *ratelimiter) Burst() int {
	return r.limiter.Burst()
}

func (r *ratelimiter) Limit() rate.Limit {
	return r.limiter.Limit()
}

func (r *ratelimiter) Reserve() Reservation {
	return &reservation{
		timesource: r.timesource,
		res:        r.limiter.ReserveN(r.timesource.Now(), 1),
	}
}

func (r *ratelimiter) SetBurst(newBurst int) {
	r.limiter.SetBurstAt(r.timesource.Now(), newBurst)
}

func (r *ratelimiter) SetLimit(newLimit rate.Limit) {
	r.limiter.SetLimitAt(r.timesource.Now(), newLimit)
}

func (r *ratelimiter) Tokens() float64 {
	return r.limiter.TokensAt(r.timesource.Now())
}

func (r *ratelimiter) Wait(ctx context.Context) (err error) {
	now := r.timesource.Now()
	res := r.limiter.ReserveN(now, 1)
	defer func() {
		if err != nil {
			// err return means "not allowed", so cancel them
			res.CancelAt(now)
		}
	}()

	if !res.OK() {
		// impossible to wait for some reason, figure out why
		if err := ctx.Err(); err != nil {
			// context already errored or raced, just return it
			return err
		}
		// !OK and not already canceled means either:
		//   1: insufficient burst, can never satisfy
		//   2: sufficient burst, but longer wait than deadline allows
		//
		// 1 must imply 0 burst and finite limit, as we only reserve 1.
		// 2 implies too many reservations for a finite wait.
		//
		// burst alone is sufficient to tell these apart, and a race that changes
		// burst during this call will at worst lead to a minor miscategorization.
		burst := r.limiter.Burst()
		if burst > 0 {
			return errWaitLongerThanDeadline
		}
		return errWaitZeroBurst
	}
	delay := res.DelayFrom(now)
	if delay == 0 {
		return nil // available token, allow it
	}

	// wait for cancellation or the waiter's turn
	timer := r.timesource.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.Chan():
		return nil
	}
}

func (r *reservation) Used(used bool) {
	if !used {
		/*
			SURPRISING DETAIL:
			canceling an immediately-allowed reservation does nothing.

			the consumed tokens will not be returned because:
			- their "time to act" is the .Reserve() call time
			- the real Reservation does nothing on Cancel() if `r.timeToAct.Before(t)`

			this means our current tiered ratelimiter allowance logic does not work
			the way it is clearly intended, where successful limits are "held" until
			their need is verified, and rolled back if not needed.

			unfortunately, until we re-implement the ratelimiter, we cannot address this.
			tokens cannot be "inserted" into a ratelimiter, only passing time restores them.

			----

			unfortunately, mocked time sources are unlikely to advance time between
			acquiring a reservation and canceling it, so they DO restore the token:
			"now" is not before "now" (because it's equal).

			to work around this, 1 nanosecond is added to all calls here.
			this makes mocked time behave like real time (no longer equal to reservation time),
			and has essentially no effect on real time sources.
		*/
		r.res.CancelAt(r.timesource.Now().Add(time.Nanosecond))
	}
	// if used, do nothing.
	//
	// this method largely exists so it can be mocked and asserted,
	// or customized by an implementation that tracks allowed/rejected metrics.
}

func (r *reservation) Allow() bool {
	return r.res.OK() && r.res.DelayFrom(r.timesource.Now()) == 0
}
