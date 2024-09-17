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

package stats

import (
	"testing"
	"time"

	"github.com/uber/cadence/common/clock"
)

func TestEmaFixedWindowQPSTracker(t *testing.T) {
	timeSource := clock.NewMockedTimeSourceAt(time.Now())
	exp := 0.4
	bucketInterval := time.Second

	r := NewEmaFixedWindowQPSTracker(timeSource, exp, bucketInterval)
	r.Start()

	// Test ReportCounter
	r.ReportCounter(10)
	r.ReportCounter(20)

	qps := r.QPS()
	if qps != 0 {
		t.Errorf("QPS mismatch, expected: 0, got: %f", qps)
	}

	timeSource.BlockUntil(1)
	timeSource.Advance(bucketInterval)
	time.Sleep(10 * time.Millisecond)
	// Test QPS
	qps = r.QPS()
	expectedQPS := float64(30) / (float64(bucketInterval) / float64(time.Second))
	if qps != expectedQPS {
		t.Errorf("QPS mismatch, expected: %f, got: %f", expectedQPS, qps)
	}

	r.ReportCounter(10)
	timeSource.BlockUntil(1)
	timeSource.Advance(bucketInterval)
	time.Sleep(10 * time.Millisecond)
	// Test QPS
	qps = r.QPS()
	expectedQPS = float64(22) / (float64(bucketInterval) / float64(time.Second))
	if qps != expectedQPS {
		t.Errorf("QPS mismatch, expected: %f, got: %f", expectedQPS, qps)
	}

	r.Stop()
}
