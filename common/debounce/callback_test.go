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

package debounce

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/uber/cadence/common/clock"
)

const (
	testDebounceInterval = time.Second
	// this one is required since we're often testing for absence of extra calls, so we have to wait
	testSleepAmount = time.Second / 2
)

type callbackTestData struct {
	mockedTimeSource  clock.MockedTimeSource
	debouncedCallback *DebouncedCallback
	calls             atomic.Int32
}

func waitCondition(fn func() bool, duration time.Duration) bool {
	started := time.Now()

	for time.Since(started) < duration {
		if fn() {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return false
}

func newCallbackTestData(t *testing.T) *callbackTestData {
	var td callbackTestData

	t.Cleanup(
		func() {
			td.debouncedCallback.Stop()
			goleak.VerifyNone(t)
		},
	)

	td.mockedTimeSource = clock.NewMockedTimeSourceAt(time.Now())
	callback := func() {
		td.calls.Add(1)
	}

	td.debouncedCallback = NewDebouncedCallback(td.mockedTimeSource, testDebounceInterval, callback)
	td.debouncedCallback.Start()
	td.mockedTimeSource.BlockUntil(1) // we should wait until ticker is created

	return &td
}

func TestDebouncedCallbackWorks(t *testing.T) {
	td := newCallbackTestData(t)

	td.debouncedCallback.Handler()
	require.True(
		t,
		waitCondition(func() bool { return td.calls.Load() > 0 }, testTimeout),
		"first callback is expected to be issued immediately after handler",
	)
	assert.Equal(t, 1, int(td.calls.Load()), "should be just once call since handler() called once")

	// issue more calls to handler(); they all should be postponed to testDebounceInterval
	for i := 0; i < 10; i++ {
		td.debouncedCallback.Handler()
	}

	td.mockedTimeSource.Advance(testDebounceInterval)
	time.Sleep(testSleepAmount)
	assert.Equal(t, 2, int(td.calls.Load()))

	// now call handler again, but advance time only by little - no callbacks are expected
	for i := 0; i < 10; i++ {
		td.debouncedCallback.Handler()
	}

	td.mockedTimeSource.Advance(testDebounceInterval / 2)
	time.Sleep(testSleepAmount)
	assert.Equal(t, 2, int(td.calls.Load()), "should not have new callbacks")
}

func TestDebouncedCallbackDoesntCallHandlerIfThereWereNoUpdates(t *testing.T) {
	td := newCallbackTestData(t)

	td.mockedTimeSource.Advance(2 * testDebounceInterval)
	time.Sleep(testSleepAmount)
	assert.Equal(t, 0, int(td.calls.Load()))
}

func TestDebouncedCallbackDoubleStopIsOK(t *testing.T) {
	td := newCallbackTestData(t)

	td.debouncedCallback.Stop()
	assert.NotPanics(t, func() { td.debouncedCallback.Stop() }, "double stop should be OK")
}
