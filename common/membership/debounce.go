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

package membership

import (
	"context"
	"sync"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
)

type DebouncedCallback struct {
	status               common.DaemonStatusManager
	lastHandlerCallMutex sync.Mutex
	lastHandlerCall      time.Time
	callback             func()
	interval             time.Duration
	timeSource           clock.TimeSource
	updateCh             chan struct{}

	loopContext context.Context
	cancelLoop  context.CancelFunc
	waitGroup   sync.WaitGroup
}

// NewDebouncedCallback creates a new DebouncedCallback which will
// accumulate calls to .Handler function and will only call `callback` on given interval if there were any calls to
// .Handler in between. It is almost like rate limiter, but it never ignores calls to handler - only postpone calling the callback once.
func NewDebouncedCallback(timeSource clock.TimeSource, interval time.Duration, callback func()) *DebouncedCallback {
	ctx, cancelFn := context.WithCancel(context.Background())

	return &DebouncedCallback{
		timeSource:  timeSource,
		interval:    interval,
		callback:    callback,
		updateCh:    make(chan struct{}, 1),
		loopContext: ctx,
		cancelLoop:  cancelFn,
	}
}

func (d *DebouncedCallback) Start() {
	if !d.status.TransitionToStart() {
		return
	}

	d.waitGroup.Add(1)
	go func() {
		defer d.waitGroup.Done()
		d.backgroundLoop()
	}()
}

func (d *DebouncedCallback) Stop() {
	if !d.status.TransitionToStop() {
		return
	}

	d.cancelLoop()
	d.waitGroup.Wait()
}

func (d *DebouncedCallback) Handler() {
	select {
	case d.updateCh <- struct{}{}:
	default:
	}

	// special case for the handler call which wasn't issued for some time
	// in this case we call the callback immediately
	d.lastHandlerCallMutex.Lock()
	defer d.lastHandlerCallMutex.Unlock()
	if d.timeSource.Since(d.lastHandlerCall) > 2*d.interval {
		d.callbackIfRequired()
	}
	d.lastHandlerCall = d.timeSource.Now()
}

func (d *DebouncedCallback) backgroundLoop() {
	ticker := d.timeSource.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-d.loopContext.Done():
			return
		case <-ticker.Chan():
			d.callbackIfRequired()
		}
	}
}

func (d *DebouncedCallback) callbackIfRequired() {
	select {
	case <-d.updateCh:
		d.callback()
	default:
	}
}

// DebouncedChannel is a wrapper around DebouncedCallback
// providing channel instead of callback
type DebouncedChannel struct {
	*DebouncedCallback
	signalCh chan struct{}
}

// NewDebouncedChannel creates a new NewDebouncedChannel which will
// accumulate calls to .Handler function and will write data to Chan() on given interval if there were any calls to
// .Handler in between. It is almost like rate limiter, but it never ignores calls to handler - only postpone.
func NewDebouncedChannel(timeSource clock.TimeSource, interval time.Duration) *DebouncedChannel {
	res := &DebouncedChannel{}

	res.signalCh = make(chan struct{}, 1)
	callback := func() {
		select {
		case res.signalCh <- struct{}{}:
		default:
		}
	}
	res.DebouncedCallback = NewDebouncedCallback(timeSource, interval, callback)
	return res
}

func (d *DebouncedChannel) Chan() <-chan struct{} { return d.signalCh }
