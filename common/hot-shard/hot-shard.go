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

package hotshard

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/dgryski/go-farm"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
)

// Options for Hotshard rate-limiter
type Options struct {
	// Main configuration options
	Limit int // how many events (at the sampled rate) before considering a value to be a hot shard (Default: 10)

	// Additional knobs
	TotalHashMapSizeLimit int           // how many values to track in its map of hashed values (Default: 1000)
	WindowSize            time.Duration // The value to sample over, higher values risk greater false-positives if the sample rate is also not commensurately large. (Default:1m)
	SampleRate            float64       // how frequently to sample input check events as a percentage expressed as a value between 0-1 (Default: 0.001)
}

// Detector is a fixed size cache which samples inputs to see if they're being
// seen frequently, and, if they have been, to flag them.
type Detector interface {
	// Check is a syncronous method which samples the workflows it is given
	// and checks (at the sample rate) if the workflows are a hot-shard over the window period.
	// Returns true when the hot-shard is detected.
	Check(now time.Time, workflowID string, additionalDimensions ...string) bool
}

const (
	MetricHotShardDetected   = "hot_shard_detected"
	_defaultLimit            = 10
	_defaultSampleRate       = 0.01
	_defaultWindow           = time.Minute
	_defaultHashMapSizeLimit = 1000
)

type detector struct {
	mu                    sync.RWMutex
	sb                    strings.Builder
	windowSize            time.Duration
	windowCount           map[uint64]int
	windowStartTime       time.Time
	limit                 int
	log                   log.Logger
	metrics               metrics.Client
	sampleRate            float64
	totalHashmapSizeLimit int
}

func (d *detector) Check(now time.Time, workflowID string, dim ...string) bool {
	// first, hash the inputs, and see
	// then if we've seen this value before
	d.mu.RLock()
	hashedIdentifier := d.hashInputs(workflowID, dim)
	existingCount, ok := d.windowCount[hashedIdentifier]
	d.mu.RUnlock()

	// if the value has been seen before, and we know it's over limit don't sample this, reply immediately that
	// we've seen this value before
	if ok && existingCount > d.limit {
		d.log.Warn("hot-shard detected", tag.WorkflowID(workflowID), tag.Dynamic("entries", dim))
		d.metrics.IncCounter(metrics.HotShardRateLimitTriggered, 1)
		go d.sampleAndCheck(now, hashedIdentifier, workflowID, dim)
		return true
	}
	return d.sampleAndCheck(now, hashedIdentifier, workflowID, dim)
}

func (d *detector) sampleAndCheck(now time.Time, hashedIdentifier uint64, workflowID string, dim []string) bool {
	defer func() {
		if r := recover(); r != nil {
			d.log.Error("Unexpected panic in hot-shard-detector", tag.Dynamic("panic", r))
		}
	}()

	if !d.shouldSample() {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	if now.After(d.windowStartTime.Add(d.windowSize)) {
		// we're rolling over the monitoring window, start afresh
		d.createNewWindow(now, d.windowStartTime)
	}

	existingCount, ok := d.windowCount[hashedIdentifier]
	if !ok {
		if len(d.windowCount) > d.totalHashmapSizeLimit {
			// don't keep adding endlessly to the hash map as, for very high
			// cardinality samples this will just become too large and add no additional
			// value, but *do* keep adding to existing values > 1 and evict
			// a not-interesting value
			for key := range d.windowCount {
				if d.windowCount[key] == 1 {
					delete(d.windowCount, key)
					break
				}
			}

			return false
		}
		d.windowCount[hashedIdentifier] = 1
		return false
	}

	d.windowCount[hashedIdentifier] = existingCount + 1
	if d.windowCount[hashedIdentifier] > d.limit {
		return true
	}
	return false
}

// NewDetector creates a hotshard detector with reasonable defaults. The default values should catch
// (approximately) hot-shard values at a rate greater than 1000 per minute
func NewDetector(log log.Logger, metricsClient metrics.Client, options Options) Detector {
	var sampleRate float64
	if options.SampleRate != 0.0 {
		sampleRate = options.SampleRate
	} else {
		sampleRate = _defaultSampleRate
	}

	var limit int
	if options.Limit != 0 {
		limit = options.Limit
	} else {
		limit = _defaultLimit
	}

	var window time.Duration
	if options.WindowSize != 0 {
		window = options.WindowSize
	} else {
		window = _defaultWindow
	}

	var totalSizeLimit int
	if options.TotalHashMapSizeLimit != 0 {
		totalSizeLimit = options.TotalHashMapSizeLimit
	} else {
		totalSizeLimit = _defaultHashMapSizeLimit
	}

	return &detector{
		mu:                    sync.RWMutex{},
		limit:                 limit,
		windowSize:            window,
		windowCount:           make(map[uint64]int),
		log:                   log,
		metrics:               metricsClient,
		sampleRate:            sampleRate,
		sb:                    strings.Builder{},
		totalHashmapSizeLimit: totalSizeLimit,
	}
}

func (d *detector) createNewWindow(now time.Time, lastUpdateTime time.Time) {
	// if we wanted to add any kind of gradual backoff for value greater than the window time
	// we could add it here
	d.windowStartTime = now
	d.windowCount = make(map[uint64]int)
}

func (d *detector) shouldSample() bool {
	if rand.Float64() < d.sampleRate {
		return true
	}
	return false
}

func (d *detector) hashInputs(workflowID string, dim []string) uint64 {
	d.sb.Reset()
	d.sb.WriteString(workflowID)
	for i := range dim {
		d.sb.WriteString(":")
		d.sb.WriteString(dim[i])
	}
	res := farm.Fingerprint64([]byte(d.sb.String()))
	return res
}
