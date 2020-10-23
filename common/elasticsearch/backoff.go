// Copyright (c) 2020 Uber Technologies, Inc.
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

package elasticsearch

import (
	"math"
	"math/rand"
	"time"
)

// exponentialBackoff implements the simple exponential backoff described by 
// Douglas Thain at http://dthain.blogspot.de/2009/02/exponential-backoff-in-distributed.html.
// TODO https://github.com/uber/cadence/issues/3676
type exponentialBackoff struct {
	initialTimeout float64 // initial timeout (in msec)
	factor         float64 // exponential factor (e.g. 2)
	maximumTimeout float64 // maximum timeout (in msec)
}

var _ GenericBackoff = (*exponentialBackoff)(nil)

// NewExponentialBackoff returns a exponentialBackoff backoff policy.
// Use initialTimeout to set the first/minimal interval
// and maxTimeout to set the maximum wait interval.
func NewExponentialBackoff(initialTimeout, maxTimeout time.Duration) GenericBackoff {
	return &exponentialBackoff{
		initialTimeout: float64(int64(initialTimeout / time.Millisecond)),
		factor:         2.0,
		maximumTimeout: float64(int64(maxTimeout / time.Millisecond)),
	}
}

// Next implements BackoffFunc for exponentialBackoff.
func (b *exponentialBackoff) Next(retry int) (time.Duration, bool) {
	r := 1.0 + rand.Float64() // random number in [1..2]
	m := math.Min(r*b.initialTimeout*math.Pow(b.factor, float64(retry)), b.maximumTimeout)
	if m >= b.maximumTimeout {
		return 0, false
	}
	d := time.Duration(int64(m)) * time.Millisecond
	return d, true
}
