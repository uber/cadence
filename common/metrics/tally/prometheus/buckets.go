// Copyright (c) 2021 Uber Technologies, Inc.
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

package prometheus

import (
	"time"

	"github.com/uber-go/tally/prometheus"
)

// DefaultHistogramBuckets is the default histogram buckets used when
// creating a new Histogram in the prometheus registry.
// Cadence uses Histogram to report both latency and non-latency metrics(like history_count/history_size).
// The buckets need to set correct boundary for both
// See more in https://github.com/uber/cadence/issues/4006
func DefaultHistogramBuckets() []prometheus.HistogramObjective {

	// For latency metrics, this is a nanoSecond.
	// For non-latency metrics, this is ONE.
	// divided by Second because of https://github.com/uber-go/tally/blob/04828d51c63b1d09b46824d7c0d7904b5eb1b3b6/prometheus/reporter.go#L183
	oneOrNanoSecond := float64(time.Nanosecond) / float64(time.Second)

	// convenient units
	microSecOr1K := oneOrNanoSecond * 1000
	milliSecOr1M := oneOrNanoSecond * 1000000

	return []prometheus.HistogramObjective{
		{Upper: 5 * oneOrNanoSecond},
		{Upper: 10 * oneOrNanoSecond},
		{Upper: 50 * oneOrNanoSecond},
		{Upper: 100 * oneOrNanoSecond},

		{Upper: microSecOr1K},
		{Upper: 10 * microSecOr1K},
		{Upper: 50 * microSecOr1K},
		{Upper: 200 * microSecOr1K},

		// Below are exactly the same as https://github.com/uber-go/tally/blob/04828d51c63b1d09b46824d7c0d7904b5eb1b3b6/prometheus/reporter.go#L51
		// to ensure backward compatibility
		{Upper: milliSecOr1M},
		{Upper: 2 * milliSecOr1M},
		{Upper: 5 * milliSecOr1M},
		{Upper: 10 * milliSecOr1M},
		{Upper: 20 * milliSecOr1M},
		{Upper: 50 * milliSecOr1M},
		{Upper: 100 * milliSecOr1M},
		{Upper: 200 * milliSecOr1M},
		{Upper: 500 * milliSecOr1M},
		{Upper: 1000 * milliSecOr1M},
		{Upper: 2000 * milliSecOr1M},
		{Upper: 5000 * milliSecOr1M},
		{Upper: 10000 * milliSecOr1M},
	}
}
