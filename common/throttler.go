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

package common

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	// Throttler is a generic component that can be used to implement rate limiting.
	// Before issuing a request the throttler should be consulted to determine
	// whether that request should be allowed to go through, or be rejected.
	// The throttler implementation determines whether to accept or reject a request
	// based on the (weighted) RPS rate and throttling level.
	Throttler interface {
		Start()
		Stop()
		TryStartRequest(weight int32) bool
		ThrottlingLevel() uint
		IncreaseThrottlingLevel(level uint)
		DecreaseThrottlingLevel(level uint)
	}

	timeSlotInfo struct {
		rateLimiter  TokenBucket
		rpsLimit     int
		requestCount int32
		startTime    time.Time
	}

	throttlerImpl struct {
		timeSlots                         []*timeSlotInfo
		samplingInterval                  time.Duration
		decreasePercentPerThrottlingLevel int
		numberOfSamples                   int
		shutdownWG                        sync.WaitGroup
		shutdownCh                        chan struct{}

		sync.RWMutex
		throttlingLevel     uint
		activeTimeSlotIndex int
	}
)

// New creates a new Throttler
func New(samplingInterval time.Duration, numberOfSamples int, decreasePercentPerThrottlingLevel int) Throttler {
	return &throttlerImpl{
		timeSlots:                         make([]*timeSlotInfo, 0, numberOfSamples),
		numberOfSamples:                   numberOfSamples,
		samplingInterval:                  samplingInterval,
		decreasePercentPerThrottlingLevel: decreasePercentPerThrottlingLevel,
	}
}

func (t *throttlerImpl) Start() {
	t.timeSlots = append(t.timeSlots, &timeSlotInfo{
		requestCount: 0,
		startTime:    time.Now(),
	})
	t.shutdownWG.Add(1)
	go t.samplerPump()
}

func (t *throttlerImpl) Stop() {
	close(t.shutdownCh)
	t.shutdownWG.Wait()
}

func (t *throttlerImpl) TryStartRequest(weight int32) bool {
	t.RLock()
	slot := t.timeSlots[t.activeTimeSlotIndex]
	t.RUnlock()

	atomic.AddInt32(&slot.requestCount, weight)
	if slot.rateLimiter != nil {
		// check if we exceeded the allowed rate
		if ok, _ := slot.rateLimiter.TryConsume(int(weight)); !ok {
			atomic.AddInt32(&slot.requestCount, -weight)
			return false
		}
	}
	return true
}

func (t *throttlerImpl) ThrottlingLevel() uint {
	t.RLock()
	defer t.RUnlock()
	return t.throttlingLevel
}

func (t *throttlerImpl) IncreaseThrottlingLevel(level uint) {
	t.Lock()
	defer t.Unlock()
	if t.throttlingLevel >= level {
		return
	}

	if len(t.timeSlots) < t.numberOfSamples {
		// don't have enough data to capture the rate.
		return
	}

	slot := t.timeSlots[t.activeTimeSlotIndex]
	if slot.rateLimiter != nil && slot.rpsLimit == 0 {
		// Already not allowing any requests, skip.
		return
	}

	for t.throttlingLevel < level {
		if t.throttlingLevel == 0 {
			// moving from not throttling to throttling
			// capture the rate
			slot.rpsLimit = t.computeRPS()
		} else {
			slot.rpsLimit = (t.decreasePercentPerThrottlingLevel * slot.rpsLimit) / 100
			// check doesn't drop below minimum?
		}
		t.throttlingLevel++
	}

	slot.rateLimiter = NewTokenBucket(slot.rpsLimit, NewRealTimeSource())
}

func (t *throttlerImpl) DecreaseThrottlingLevel(level uint) {
	t.Lock()
	defer t.Unlock()
	if t.throttlingLevel <= level {
		return
	}
	slot := t.timeSlots[t.activeTimeSlotIndex]
	if level == 0 {
		slot.rateLimiter = nil
		t.throttlingLevel = 0
	} else {
		for t.throttlingLevel > level {
			slot.rpsLimit = (100 * slot.rpsLimit) / t.decreasePercentPerThrottlingLevel
			t.throttlingLevel--
		}
		slot.rateLimiter = NewTokenBucket(slot.rpsLimit, NewRealTimeSource())
	}
}

func (t *throttlerImpl) samplerPump() {
	defer t.shutdownWG.Done()
	sampleTimer := time.NewTimer(t.samplingInterval)

samplerPumpLoop:
	for {
		select {
		case <-t.shutdownCh:
			break samplerPumpLoop
		case <-sampleTimer.C:
			{
				t.Lock()
				slot := *t.timeSlots[t.activeTimeSlotIndex]
				slot.requestCount = 0
				slot.startTime = time.Now()
				t.addNewTimeSlot(&slot)
				t.Unlock()
				sampleTimer = time.NewTimer(t.samplingInterval)
			}
		}
	}
}

func (t *throttlerImpl) addNewTimeSlot(slot *timeSlotInfo) {
	if len(t.timeSlots) < t.numberOfSamples {
		t.timeSlots = append(t.timeSlots, slot)
		t.activeTimeSlotIndex = len(t.timeSlots) - 1
	} else {
		nextActiveSlotIndex := (t.activeTimeSlotIndex + 1) % t.numberOfSamples
		t.timeSlots[nextActiveSlotIndex] = slot
		t.activeTimeSlotIndex = nextActiveSlotIndex
	}
}

func (t *throttlerImpl) computeRPS() int {
	totalRPS := 0
	slotCount := 0
	for _, slot := range t.timeSlots {
		elapsedSec := time.Now().Sub(slot.startTime).Seconds()
		if elapsedSec > 0 {
			totalRPS += int(float64(slot.requestCount) / elapsedSec)
			slotCount++
		}
	}
	return totalRPS / slotCount
}
