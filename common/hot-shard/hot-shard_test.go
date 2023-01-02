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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dgryski/go-farm"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
)

func TestWindowRollover(t *testing.T) {

}

func TestE2ESampleExampleWithSmallSet(t *testing.T) {
	t1 := time.Now()

	//detector := NewDetector(100, time.Second, 100, loggerimpl.NewNopLogger(), metrics.NewNoopMetricsClient())
	detector := NewDetector(loggerimpl.NewNopLogger(), metrics.NewNoopMetricsClient(), Options{})
	for i := 0; i < 100; i++ {
		assert.False(t, detector.Check(t1, "wf-1", "entry 2"))
	}
}

func TestE2ESampleExampleWithModerateSet(t *testing.T) {
	t1 := time.Now()

	detector := NewDetector(loggerimpl.NewNopLogger(), metrics.NewNoopMetricsClient(), Options{})

	notAHotShard := "wf-1"
	hotShard := "hs1"

	// everything should be not a hit on the detector initially
	for i := 0; i < 100; i++ {
		assert.False(t, detector.Check(t1, notAHotShard))
		assert.False(t, detector.Check(t1, hotShard))
		assert.False(t, detector.Check(t1, uuid.New().String()))
	}

	// some warm data
	for i := 0; i < 100; i++ {
		detector.Check(t1, notAHotShard)
	}

	// add some random guff
	for i := 0; i < 10000; i++ {
		detector.Check(t1, uuid.New().String())
	}

	// trigger the hot shard
	for i := 0; i < 4000; i++ {
		detector.Check(t1, hotShard)
	}

	// by now, the hashtable should contain enough samples of the hot-shard to identify
	// the problematic shard (wf-2) but not give any false positives
	for i := 0; i < 10; i++ {
		assert.False(t, detector.Check(t1, uuid.New().String()))
		assert.False(t, detector.Check(t1, uuid.New().String()))
		assert.True(t, detector.Check(t1, hotShard))
	}
}

func TestWindowMoving(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now().Add(time.Minute + time.Second)

	d := detector{
		mu:                    sync.RWMutex{},
		sb:                    strings.Builder{},
		windowSize:            _defaultWindow,
		windowCount:           make(map[uint64]int),
		windowStartTime:       time.Time{},
		limit:                 1000,
		log:                   loggerimpl.NewNopLogger(),
		metrics:               metrics.NewNoopMetricsClient(),
		sampleRate:            1,
		totalHashmapSizeLimit: _defaultHashMapSizeLimit,
	}

	for i := 0; i < 100; i++ {
		d.Check(t1, "123")
	}

	hash := farm.Fingerprint64([]byte("123"))
	assert.Equal(t, 100, d.windowCount[hash])

	d.Check(t2, "123")
	assert.Equal(t, 1, d.windowCount[hash])
}
