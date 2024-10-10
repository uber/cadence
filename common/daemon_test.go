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

package common

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaemonStatusManagerWorks(t *testing.T) {
	var dm DaemonStatusManager

	require.Equal(t, DaemonStatusInitialized, dm.status.Load())

	assert.True(t, dm.TransitionToStart())
	assert.False(t, dm.TransitionToStart(), "already started")
	assert.True(t, dm.TransitionToStop())
	assert.False(t, dm.TransitionToStop(), "already stopped")
}

func TestDaemonStatusManagerRequiresStartBeforeStop(t *testing.T) {
	var dm DaemonStatusManager

	assert.False(t, dm.TransitionToStop(), "never been started")
}

func TestDaemonStatusRaceCondition(t *testing.T) {
	var dm DaemonStatusManager
	var successes atomic.Int32
	var wg sync.WaitGroup
	var started sync.WaitGroup

	started.Add(1) // make sure we wait until all goroutines are started

	// try to issue multiple Start-s at once
	for i := 0; i < 10; i++ {
		wg.Add(1)
		started.Add(1)
		go func() {
			defer wg.Done()
			started.Done() // mark as "waiting"
			started.Wait() // wait for other goroutines to catch up
			if dm.TransitionToStart() {
				successes.Add(1)
			}
		}()
	}

	started.Done() // allows resuming goroutines once they're all started
	wg.Wait()
	assert.Equal(t, 1, int(successes.Load()), "only one Start call should succeed")

	// now do the same for stops
	successes.Store(0)
	started.Add(1) // make sure we wait until all goroutines are started
	for i := 0; i < 10; i++ {
		wg.Add(1)
		started.Add(1)
		go func() {
			defer wg.Done()
			started.Done() // mark as "waiting"
			started.Wait() // wait for other goroutines to catch up
			if dm.TransitionToStop() {
				successes.Add(1)
			}
		}()
	}

	started.Done() // allows resuming goroutines once they're all started
	wg.Wait()
	assert.Equal(t, 1, int(successes.Load()), "only one Stop call should succeed")
}
