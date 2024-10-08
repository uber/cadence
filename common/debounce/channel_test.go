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
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
)

const testTimeout = 10 * time.Second

type channelTestData struct {
	mockedTimeSource clock.MockedTimeSource
	debouncedChannel *DebouncedChannel
}

func newChannelTestData(t *testing.T) *channelTestData {
	var td channelTestData

	td.mockedTimeSource = clock.NewMockedTimeSourceAt(time.Now())
	td.debouncedChannel = NewDebouncedChannel(td.mockedTimeSource, testDebounceInterval)

	t.Cleanup(
		func() {
			td.debouncedChannel.Stop()
			goleak.VerifyNone(t)
		},
	)

	td.debouncedChannel.Start()
	// we should wait until ticker is created
	td.mockedTimeSource.BlockUntil(1)
	return &td
}

func (td *channelTestData) countCalls(t *testing.T, calls *int) {
	// create a reader from channel which will save calls:
	//   this way it should be easier to write tests by just evaluating td.calls
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-td.debouncedChannel.Chan():
				*calls++
			case <-ctx.Done():
				return
			}
		}
	}()

	t.Cleanup(
		func() {
			cancel()
			assert.True(
				t,
				common.AwaitWaitGroup(&wg, testTimeout),
				"debouncedChannel channel reader must be stopped",
			)
		},
	)
}

func TestDebouncedSignalWorks(t *testing.T) {
	td := newChannelTestData(t)

	var calls int
	td.countCalls(t, &calls)

	td.debouncedChannel.Handler()
	require.True(
		t,
		waitCondition(func() bool { return calls > 0 }, testTimeout),
		"first callback is expected to be issued immediately after handler",
	)
	assert.Equal(t, 1, calls, 1)

	// we call handler multiple times. There should be just one message in channel
	for i := 0; i < 10; i++ {
		td.debouncedChannel.Handler()
	}

	td.mockedTimeSource.Advance(testDebounceInterval)
	time.Sleep(testSleepAmount)
	assert.Equal(t, 2, calls)

	// now call handler again, but advance time only by little - no messages in channel are expected
	for i := 0; i < 10; i++ {
		td.debouncedChannel.Handler()
	}

	td.mockedTimeSource.Advance(testDebounceInterval / 2)
	time.Sleep(testSleepAmount)
	assert.Equal(t, 2, calls, "should not have new messages in channel")
}

func TestDebouncedSignalDoesntDuplicateIfWeDontReadChannel(t *testing.T) {
	td := newChannelTestData(t)

	// this should lead message in channel
	for i := 0; i < 10; i++ {
		td.debouncedChannel.Handler()
	}
	td.mockedTimeSource.Advance(testDebounceInterval)
	time.Sleep(testSleepAmount)

	// same, but we don't expect new message since we haven't read one
	for i := 0; i < 10; i++ {
		td.debouncedChannel.Handler()
	}
	td.mockedTimeSource.Advance(testDebounceInterval)
	time.Sleep(testSleepAmount)

	// only now we start reading messages from channel
	var calls int
	td.countCalls(t, &calls)
	time.Sleep(testSleepAmount)

	assert.Equal(t, 1, calls, "Only a single message is expected")
}
