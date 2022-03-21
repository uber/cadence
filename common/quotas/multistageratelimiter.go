// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package quotas

import "golang.org/x/time/rate"

// MultiStageRateLimiter indicates a domain specific rate limit policy
type MultiStageRateLimiter struct {
	// "user"-request limits, e.g. for start, signal, query.
	domainUserLimiters *Collection
	// "worker"-request limits, currently just polling.  task-responses are not limited.
	domainWorkerLimiters *Collection

	globalUserLimiter   Limiter
	globalWorkerLimiter Limiter
}

// NewMultiStageRateLimiter returns a new domain quota rate limiter. This is about
// an order of magnitude slower than
func NewMultiStageRateLimiter(globalUser, globalWorker Limiter, domainUser, domainWorker *Collection) *MultiStageRateLimiter {
	return &MultiStageRateLimiter{
		domainUserLimiters:   domainUser,
		domainWorkerLimiters: domainWorker,
		globalUserLimiter:    globalUser,
		globalWorkerLimiter:  globalWorker,
	}
}

// Allow attempts to allow a request to go through. The method returns
// immediately with a true or false indicating if the request can make
// progress
func (d *MultiStageRateLimiter) Allow(info Info) bool {
	var global Limiter
	if info.IsUser {
		global = d.globalUserLimiter
	} else {
		global = d.globalWorkerLimiter
	}

	domain := info.Domain
	if len(domain) == 0 {
		return global.Allow()
	}

	// take a reservation with the domain limiter first
	var rsv *rate.Reservation
	if info.IsUser {
		rsv = d.domainUserLimiters.For(domain).Reserve()
	} else {
		rsv = d.domainWorkerLimiters.For(domain).Reserve()
	}
	if !rsv.OK() {
		return false
	}

	// check whether the reservation is valid now, otherwise
	// cancel and return right away so we can drop the request
	if rsv.Delay() != 0 {
		rsv.Cancel()
		return false
	}

	// ensure that the reservation does not break the global rate limit, if it
	// does, cancel the reservation and do not allow to proceed.
	if !global.Allow() {
		rsv.Cancel()
		return false
	}
	return true
}
