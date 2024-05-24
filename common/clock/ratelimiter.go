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
	//   - Time is never allowed to rewind, and does not advance when canceling,
	//     fixing core issues with [golang.org/x/time/rate.Limiter] and greatly
	//     improving the chances of successfully canceling a reserved token in our
	//     usage (from "almost literally never" to "likely 99%+ in our usage").
	//   - Reservations are simplified and do not allow waiting, but are MUCH more
	//     likely to return tokens when canceled.
	//   - All MethodAt APIs have been excluded as we do not use them, and they would
	//     have to bypass the contained TimeSource, which seems error-prone.
	//   - All `MethodN` APIs (specifically the ">1 event" parts) have been excluded
	//     because we do not use them, but they seem fairly easy to add if needed later.
	Ratelimiter interface {
		// Allow returns true if an event can be performed now, i.e. there
		// are burst-able tokens available to use.
		//
		// To perform events at a steady rate, use Wait instead.
		Allow() bool
		// Burst returns the maximum burst size allowed by this Ratelimiter.
		// A Burst of zero returns false from Allow, unless its Limit is infinity.
		//
		// To see the number of burstable tokens available now, use Tokens.
		Burst() int
		// Limit returns the maximum overall rate of events that are allowed.
		Limit() rate.Limit
		// Reserve claims a single Allow() call that can be canceled if not used.
		// Check Reservation.Allow() to see if the event is allowed.
		//
		// This Reservation does not support waiting, and needs to have
		// Reservation.Used(true/false) called as soon as possible to have the
		// best chance of returning its unused token.
		Reserve() Reservation
		// SetBurst sets the Burst value
		SetBurst(newBurst int)
		// SetLimit sets the Limit value
		SetLimit(newLimit rate.Limit)
		// Tokens returns the number of immediately-allowable events when called.
		// Values >= 1 will lead to Allow returning true.
		//
		// These values are NOT able to guarantee that a later call will succeed
		// or fail, as other calls to Allow or Reserve may occur before you are
		// able to act on the returned value.
		// It is essentially only suitable for monitoring or test use.
		Tokens() float64
		// Wait blocks until one token is available (and consumes it), or it returns
		// an error and does not consume any tokens if:
		//   - context is canceled
		//   - a token will never become available (burst == 0)
		//   - a token is not expected to become available in time
		//     (short deadline and/or many pending reservations or waiters)
		//
		// If an error is returned, regardless of the reason, you may NOT proceed
		// with the limited event you were waiting for.
		Wait(ctx context.Context) error
	}

	// Reservation is a simplified and test-friendly version of [golang.org/x/time/rate.Reservation]
	// that only allows checking if it is successful or not, and (possibly) returning unused tokens.
	//
	// This Reservation MUST have Allow() checked and marked as Used(true/false)
	// as soon as possible, or an unused token is unlikely to be returned.
	//
	// Recommended use is either by using defer:
	//
	// 	func allow() (allowed bool) {
	//     	r := ratelimiter.Reserve()
	// 		// defer in a closure to read the most-recent value, not the value when deferred
	// 		defer func() { r.Used(allowed) }()
	// 		allowed = r.Allow()
	// 		if !allowed {
	//			return false // also marks the reservation as not-used
	// 		}
	// 		if somethingElseAllowsIt() {
	// 			// allow the event, and mark the reservation as used so the token is consumed
	//			return true
	// 		}
	// 		// cancel the reservation and possibly return the token
	// 		return false
	// 	}
	//
	// Or directly calling Used when appropriate:
	//
	// 	func allow() bool {
	//     	r := ratelimiter.Reserve()
	// 		allowed = r.Allow()
	// 		if !allowed {
	// 			r.Used(false) // mark the reservation as not-used
	//			return false
	// 		}
	// 		if somethingElseAllowsIt() {
	// 			// allow the event, and mark the reservation as used
	// 			r.Used(true)
	//			return true
	// 		}
	// 		r.Used(false) // mark the reservation as not-used, and restore the unused token
	// 		return false
	// 	}
	Reservation interface {
		// Allow returns true if a request should be allowed.
		//
		// This may be called at any time and it will return the same value each time,
		// but it becomes less correct as time passes, so you are expected to call it ~immediately.
		//
		// As soon as possible, also call Used(true/false) to consume / return the reserved token.
		//
		// This is equivalent to `OK() == true && DelayFrom(reservationTime) == 0`
		// with [rate.Reservation].
		Allow() bool

		// Used marks this Reservation as either used or not-used, and it must be called
		// once on every Reservation:
		//  - If true, the event is assumed to be allowed, and the Ratelimiter token
		//    will remain consumed.
		//  - If false, the Reservation will be rolled back, possibly restoring a Ratelimiter token.
		//    This is equivalent to calling Cancel() on a [rate.Reservation], but it is more likely
		//    to return the token.
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

	// a reservation that allows an immediate call, and can be canceled
	allowedReservation struct {
		res *rate.Reservation
		// reservedAt is used to cancel at the reservation time, restoring tokens
		// if (and only if) no other time-advancing calls have interleaved
		reservedAt time.Time
		// needs access to the parent-wrapped-ratelimiter to ensure time is not rewound
		limiter *ratelimiter
	}
	// a reservation that does NOT allow an immediate call.
	// the *rate.Reservation has already been canceled, so this type is trivial.
	deniedReservation struct{}
)

var (
	_ Ratelimiter = (*ratelimiter)(nil)
	_ Reservation = (*allowedReservation)(nil)
	_ Reservation = deniedReservation{}
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
// This should be equivalent to setting `r.timesource.Now()`, outside of tests
// that abuse mocked time and Advance with negative values.
//
// The lock MUST be held until all time-accepting methods are called on the
// underlying rate.Limiter, as otherwise the internal "now" value may rewind
// and cause undefined behavior.
func (r *ratelimiter) lockNow() (now time.Time, unlock func()) {
	r.mut.Lock()
	// important design note: `r.timesource.Now()` MUST be gathered while the lock is held,
	// or the times we compare are not guaranteed to be monotonic due to waiting on mutexes.
	newNow := r.timesource.Now()
	if newNow.After(r.latestNow) {
		// this should always be true because `time.Now()` is monotonic, and it is
		// always acquired inside the lock (so values are strictly ordered and
		// not arbitrarily delayed).
		//
		// that means e.g. lockReservedNow should not ever receive a value newer than
		// the Now we just acquired, so r.latestNow should never be newer either, so
		// this branch is always selected (unless time is equal).
		//
		// that said: it still shouldn't be allowed to rewind, so it's checked.
		r.latestNow = newNow
	}
	return r.latestNow, r.mut.Unlock
}

// lockReservedNow gets the correct "now" time to use, ensuring it never rolls back.
// This is intended to be used by reservations, to allow them to return claimed tokens
// after arbitrary waits, as long as no other calls have interleaved.
//
// The returned value MUST be used instead of the passed time: it may be the same or newer.
// The passed value may rewind the rate.Limiter's internal time record, and cause undefined behavior.
//
// The lock MUST be held until all time-accepting methods are called on the
// underlying rate.Limiter, as otherwise the internal "now" value may rewind
// and cause undefined behavior.
func (r *ratelimiter) lockReservedNow(tryNow time.Time) (now time.Time, unlock func()) {
	r.mut.Lock()
	if tryNow.After(r.latestNow) {
		// this should be true if a cancellation arrives before any other call
		// has advanced time (e.g. other Allow() calls).
		//
		// "old" cancels are expected to skip this path, and return the r.latestNow
		// without updating it.
		r.latestNow = tryNow
	}
	return r.latestNow, r.mut.Unlock
}

func (r *ratelimiter) Allow() bool {
	now, unlock := r.lockNow()
	defer unlock()
	return r.limiter.AllowN(now, 1)
}

func (r *ratelimiter) Reserve() Reservation {
	now, unlock := r.lockNow()
	defer unlock()
	res := r.limiter.ReserveN(now, 1)
	if res.OK() && res.DelayFrom(now) == 0 {
		// usable token, return a cancel-able handler in case it's not used
		return &allowedReservation{
			res:        res,
			reservedAt: now,
			limiter:    r,
		}
	}
	// unusable, immediately cancel it to get rid of the future-reservation
	// since we don't allow waiting for it.
	//
	// this restores MANY more tokens when heavily contended, as time skews
	// less before canceling, so it gets MUCH closer to the intended limit.
	res.CancelAt(now)
	return deniedReservation{}
}

func (r *ratelimiter) Burst() int {
	// rate.Limiter.Burst does not currently advance time,
	// so this does not need to hold the ratelimiter lock.
	// if this changes, use lockNow.
	return r.limiter.Burst()
}

func (r *ratelimiter) Limit() rate.Limit {
	// rate.Limiter.Limit does not currently advance time,
	// so this does not need to hold the ratelimiter lock.
	// if this changes, use lockNow.
	return r.limiter.Limit()
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
	defer once.Do(unlock) // unlock if panic or returned early with no err

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
			// that the token will be restored, but if there was any waiting
			// it's pretty likely that other calls occurred while unlocked and
			// the most-recently-observed time must be maintained.
			now, unlock := r.lockReservedNow(now)
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
		// burst alone is sufficient to tell these apart.
		//
		// if we begin allowing >1 reservations, this may have to change, so
		// these error values are NOT exposed to prevent building behavior
		// that depends on this detail.
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

	// wait for cancellation or the waiter's turn.
	// re-acquiring Now() here is valid and may shorten the delay, but may add
	// more branches and isn't strictly necessary as sleeps aren't precise anyway.
	timer := r.timesource.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.Chan():
		return nil
	}
}

func (r *allowedReservation) Used(used bool) {
	if !used {
		now, unlock := r.limiter.lockReservedNow(r.reservedAt)
		// lock must be held while canceling, because it can affect the limiter's time
		defer unlock()

		// if no calls interleaved, this will be the same as the reservation time,
		// and the cancellation will be able to roll back an immediately-available call.
		//
		// otherwise, the ratelimiter's time has been advanced, and this will fail to
		// restore any tokens.
		r.res.CancelAt(now)
	}
	// if used, do nothing.
	//
	// this method largely exists so it can be mocked and asserted,
	// or customized by an implementation that tracks allowed/rejected metrics.
}

func (r *allowedReservation) Allow() bool {
	// this is equivalent to `r.res.OK() && r.res.DelayFrom(r.reservedAt) == 0`,
	// but this is known to be `true` when it is first reserved and `r.reservedAt`
	// does not ever change, so it can be skipped here.
	return true
}

func (r deniedReservation) Used(used bool) {
	// reservation already canceled, nothing to roll back
}

func (r deniedReservation) Allow() bool {
	return false
}
