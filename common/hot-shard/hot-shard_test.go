package hot_shard

import (
	"github.com/dgryski/go-farm"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"strings"
	"sync"
	"testing"
	"time"
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
		scope:                 metrics.NewNoopMetricsClient(),
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
