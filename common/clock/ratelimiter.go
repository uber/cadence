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
	"sync"
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
		//
		// This Reservation MUST have Allow() checked and marked as Used(true/false)
		// as soon as possible, or behavior may be confusing and it may be unable to
		// restore the claimed ratelimiter token if not used.
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
	// Reservation is a simplified and test-friendly version of [golang.org/x/time/rate.Reservation]
	// that only allows checking if it is successful or not.
	//
	// This Reservation MUST have Allow() checked and marked as Used(true/false)
	// as soon as possible, or behavior may be confusing and it may be unable to
	// restore the claimed ratelimiter token if not used.
	Reservation interface {
		// Allow returns true if a request should be allowed.
		//
		// This is expected to be called ~immediately, and only once.  Doing otherwise may behave
		// irrationally due to its internally-pinned time.
		// In particular, do not "Reserve -> time.Sleep -> Allow" as it may or may not account for
		// any of the time spent sleeping.
		//
		// As soon as possible, also call Used(true/false) to consume / return the reserved token.
		//
		// This is equivalent to `OK() == true && Delay() == 0` with [golang.org/x/time/rate.Reservation].
		Allow() bool

		// Used marks this Reservation as either used or not-used, and it must be called
		// once on every Reservation:
		//  - If true, the operation is assumed to be allowed, and the Ratelimiter token
		//    will remain consumed.
		//  - If false, the Reservation will be rolled back, restoring a Ratelimiter token.
		//    This is equivalent to calling Cancel() on a golang.org/x/time/rate.Reservation
		Used(wasUsed bool)
	}

	ratelimiter struct {
		// important concurrency note: the lock MUST be held while acquiring AND handling Now(),
		// to ensure that times are handled monotonically.
		// otherwise, even though `time.Now()` is monotonic, they may make progress
		// past the mutex in a different order and no longer be handled in a monotonic way.
		timesource TimeSource
		latestNow  time.Time // updated on each call, never allowed to rewind
		limiter    *rate.Limiter
		mut        sync.Mutex
	}

	reservation struct {
		res *rate.Reservation

		// reservedAt is used to un-reserve at the reservation time, restoring tokens if no calls have interleaved
		reservedAt time.Time
		// need access to the parent-wrapped-ratelimiter to ensure time is not rewound
		limiter *ratelimiter
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
	ErrCannotWait             = fmt.Errorf("ratelimiter.Wait() cannot be satisfied")
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

// lockNow updates the current "now" time, and ensures it never rolls back.
// this should be equivalent to `r.timesource.Now()` outside of tests that abuse mocked time.
//
// The lock MUST be held until all time-accepting methods are called on the
// underlying rate.Limiter, as otherwise the internal "now" value may rewind
// and cause undefined behavior.
//
// Important design note: `r.timesource.Now()` must be gathered while the lock is held,
// or the times we compare are not guaranteed to be monotonic due to waiting on mutexes.
// The `.Before` comparison should ensure *better* behavior in spite of this, but it's
// much more correct if it's acquired within the lock.
func (r *ratelimiter) lockNow() (now time.Time, unlock func()) {
	r.mut.Lock()
	newNow := r.timesource.Now()
	if r.latestNow.Before(newNow) {
		// this should always be true because `time.Now()` is monotonic, and it is
		// always acquired inside the lock (so values are strictly ordered and
		// not arbitrarily delayed).
		//
		// that means e.g. lockCorrectNow should not ever receive a value newer than
		// the Now we just acquired, so r.latestNow should never be newer either.
		//
		// that said: it still shouldn't be allowed to rewind.
		r.latestNow = newNow
	}
	return r.latestNow, r.mut.Unlock
}

// lockCorrectNow gets the correct "now" time to use, and ensures it never rolls back.
// this is intended to be used by reservations, to allow them to return claimed tokens
// after arbitrary waits, as long as no other calls have interleaved.
//
// the returned value MUST be used instead of the passed time: it may be the same or newer.
// the passed value may rewind the rate.Limiter's internal time record, and cause undefined behavior.
//
// The lock MUST be held until all time-accepting methods are called on the
// underlying rate.Limiter, as otherwise the internal "now" value may rewind
// and cause undefined behavior.
func (r *ratelimiter) lockCorrectNow(tryNow time.Time) (now time.Time, unlock func()) {
	r.mut.Lock()
	if r.latestNow.Before(tryNow) {
		// this should be true if a cancellation arrives before any other call
		// has advanced time (e.g. other Allow() calls).
		//
		// "old" cancels are expected to skip this path, and return the r.latestNow
		// without updating it.  they may or may not release their token in this case.
		r.latestNow = tryNow
	}
	return r.latestNow, r.mut.Unlock
}

func (r *ratelimiter) Allow() bool {
	now, unlock := r.lockNow()
	defer unlock()
	return r.limiter.AllowN(now, 1)
}

func (r *ratelimiter) Burst() int {
	return r.limiter.Burst()
}

func (r *ratelimiter) Limit() rate.Limit {
	return r.limiter.Limit()
}

func (r *ratelimiter) Reserve() Reservation {
	now, unlock := r.lockNow()
	defer unlock()
	return &reservation{
		res:        r.limiter.ReserveN(now, 1),
		reservedAt: now,
		limiter:    r,
	}
}

func (r *ratelimiter) SetBurst(newBurst int) {
	now, unlock := r.lockNow()
	defer unlock()
	r.limiter.SetBurstAt(now, newBurst)
}

func (r *ratelimiter) SetLimit(newLimit rate.Limit) {
	now, unlock := r.lockNow()
	defer unlock()
	r.limiter.SetLimitAt(now, newLimit)
}

func (r *ratelimiter) Tokens() float64 {
	now, unlock := r.lockNow()
	defer unlock()
	return r.limiter.TokensAt(now)
}

func (r *ratelimiter) Wait(ctx context.Context) (err error) {
	now, unlock := r.lockNow()
	var once sync.Once
	defer once.Do(unlock) // unlock if panic or returned early with no err, noop if it waited

	res := r.limiter.ReserveN(now, 1)
	defer func() {
		if err != nil {
			// err return means "not allowed", so cancel the reservation.
			//
			// note that this makes a separate call to get the current time:
			// this is intentional, as time may have passed while waiting.
			//
			// if the cancellation happened before the time-to-act (the delay),
			// the reservation will still be successfully rolled back.

			// ensure we are unlocked, which will not have happened if an err
			// was returned immediately.
			once.Do(unlock)

			// (re)-acquire the latest now value.
			//
			// it should not be advanced to the "real" now, to improve the chance
			// that the token will be restored, but it's pretty likely that other
			// calls occurred while unlocked and the most-recently-observed time
			// must be maintained.
			now, unlock := r.lockCorrectNow(now)
			defer unlock()
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
	if deadline, ok := ctx.Deadline(); ok && now.Add(delay).After(deadline) {
		return errWaitLongerThanDeadline // don't wait for a known failure
	}

	once.Do(unlock) // unlock before waiting

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
		now, unlock := r.limiter.lockCorrectNow(r.reservedAt)
		// lock must be held while canceling, because it updates the limiter's time
		defer unlock()

		// if no calls interleaved, this will be the same as the reservation time,
		// and the cancellation will be able to roll back an immediately-available call.
		//
		// otherwise, the ratelimiter's time has been advanced, and this may fail to
		// restore any tokens.
		r.res.CancelAt(now)
	}
	// if used, do nothing.
	//
	// this method largely exists so it can be mocked and asserted,
	// or customized by an implementation that tracks allowed/rejected metrics.
}

func (r *reservation) Allow() bool {
	// because this uses a snapshotted DelayFrom time rather than whatever the
	// "current" time might be, this result will not change as time passes.
	//
	// that's intentional because it's more stable... and a bit surprisingly
	// it's also important, as interleaving calls may advance the limiter's
	// "now" time, and turn this "would not allow" into a "would allow".
	//
	// in high-CPU benchmarks, if you use the limiter's "now", you can sometimes
	// see this allowing many times more requests than it should, due to the
	// interleaving time advancement.
	return r.res.OK() && r.res.DelayFrom(r.reservedAt) == 0
}
