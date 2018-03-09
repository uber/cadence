// Copyright (c) 2017 Uber Technologies, Inc.
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

package history

import (
	"sync"
	"time"
)

type (
	// Timer interface
	Timer interface {
		// GetTimerChan return the channel which will be fired when time is up
		GetTimerChan() <-chan bool
		// WillFireAfter check will the timer get fired after a certain time
		WillFireAfter(now time.Time) bool
		// UpdateTimer update the timer, return true if update is a success
		// success means timer is idle or timer is set with a sooner time to fire
		UpdateTimer(nextTime time.Time) bool
		// Close shutdown the timer
		Close()
	}

	// TimerImpl is an timer implementation,
	// which basically is an wrrapper of golang's timer and
	// additional feature
	TimerImpl struct {
		// the channel which will be used to proxy the fired timer
		timerChan chan bool
		closeChan chan bool

		// lock for timer and next wake up time
		sync.Mutex
		// the actual timer which will fire
		timer *time.Timer
		// variable indicating when the above timer will fire
		nextWakeupTime time.Time
	}
)

// NewTimer create a new timer instance
func NewTimer() Timer {
	timer := &TimerImpl{
		timer:          time.NewTimer(0),
		nextWakeupTime: time.Time{},
		timerChan:      make(chan bool),
		closeChan:      make(chan bool),
	}

	go func() {
		defer close(timer.timerChan)
		defer timer.timer.Stop()
	loop:
		for {
			select {
			case <-timer.timer.C:
				// re-transmit on gateC
				timer.timerChan <- true

			case <-timer.closeChan:
				// closed; cleanup and quit
				break loop
			}
		}
	}()

	return timer
}

// GetTimerChan return the channel which will be fired when time is up
func (timer *TimerImpl) GetTimerChan() <-chan bool {
	return timer.timerChan
}

// WillFireAfter check will the timer get fired after a certain time
func (timer *TimerImpl) WillFireAfter(now time.Time) bool {
	return timer.nextWakeupTime.After(now)
}

// UpdateTimer update the timer, return true if update is a success
// success means timer is idle or timer is set with a sooner time to fire
func (timer *TimerImpl) UpdateTimer(nextTime time.Time) bool {
	now := time.Now()

	timer.Lock()
	defer timer.Unlock()
	if !timer.WillFireAfter(now) || timer.nextWakeupTime.After(nextTime) {
		// if timer will not fire or next wake time is after the "next"
		// then we need to update the timer to fire
		timer.nextWakeupTime = nextTime
		// reset timer to fire when the next message should be made 'visible'
		timer.timer.Reset(nextTime.Sub(now))

		// Notifies caller that next notification is reset to fire at passed in 'next' visibility time
		return true
	}

	return false
}

// Close shutdown the timer
func (timer *TimerImpl) Close() {
	close(timer.closeChan)
}
