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

	"golang.org/x/time/rate"

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
func NewShadowedLimiter(primary, shadow quotas.Limiter) quotas.Limiter {
	return shadowedLimiter{
		primary: primary,
		shadow:  shadow,
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

func (s shadowedLimiter) Limit() rate.Limit {
	return s.primary.Limit()
}

func (s shadowedReservation) Allow() bool {
	_ = s.shadow.Allow()
	return s.primary.Allow()
}

func (s shadowedReservation) Used(wasUsed bool) {
	s.primary.Used(wasUsed)

	// Shadow reservations are more complicated than they seem.
	//
	// Mirroring the *exact* call here will not ever let the shadow allow more RPS
	// than the primary, because the primary will reject and not-use requests exceeding its limit.
	//
	// This means you'll see fewer rejects (because shadow-cancels are not rejects), but:
	// - You can't assume each reject would have been an allow because it returned
	//   the shadow-token, allowing ~infinite not-counted-reject-RPS on shadows.
	// - You can't tell what the shadow's peak RPS actually is, because it will
	//   always be limited to match the primary (or lower).
	//
	// This doesn't really tell you anything useful about the shadow's behavior.
	// To improve this, we need to try to differentiate between "was this
	// rejected directly by the primary" vs "was this rejected by something else,
	// and that's transitively canceling this reservation".
	//
	// That's pretty easy: we can tell by the primary's Allow state.

	if s.primary.Allow() {
		// If wasUsed: this request was allowed, so the shadow should also try to consume any token
		// it received (pass `true`).
		//
		// If !wasUsed: this request was rejected by something else, e.g. a second-stage limiter.
		// this second-stage limiter would also have rejected the request if shadow was primary,
		// so try to return the shadow's token (pass `false`).
		// This will cap the shadow's allow-RPS to match the later-stage limiter, but that would
		// still be true if the shadow limiter became the primary, so it's accurate behavior.
		//
		// Tn other words: just forward `wasUsed` to shadow.
		s.shadow.Used(wasUsed)
	} else {
		// No transitive rejection is possible (we do not currently try to get two limiters separately before checking,
		// so this must be due to the primary rejecting the request), so we must assume that any token available would
		// have been used.
		//
		// This will be misleading about the behavior of later-stage limiters (shadow may allow more RPS than if it was
		// primary because a later-stage limiter could lead to it canceling reservations, effectively capping its RPS
		// below its configured RPS), but there isn't much we can do to know about that here.
		// Metrics on that second-stage limiter should show why a multi-stage operation is being rejected, and we could
		// start reporting "allowed but canceled transitively" metrics to show when these transitive limits are affecting
		// a limiter, and that's about as good as we can get.
		//
		// But regardless, for this shadowed reservation: always try to consume the token.
		// - This allows shadow-RPS to go above primary-RPS, which is not possible if `false` was forwarded here.
		// - Shadow-RPS can still go below primary-RPS because it can reject additional requests.
		//   it'll essentially just be "primary-rejects + difference-in-allows" which isn't all that useful, but it's relatively accurate.
		s.shadow.Used(true)
	}
}
