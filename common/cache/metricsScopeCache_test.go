// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
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

package cache

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/bmizerany/assert"

	"github.com/uber/cadence/common/metrics/mocks"
)

func TestGetMetricsScope(t *testing.T) {
	metricsCache := NewMetricsCache()

	mockMetricsScope := &mocks.Scope{}

	metricsCache.Put("A", mockMetricsScope)
	metricsCache.Put("B", mockMetricsScope)
	metricsCache.Put("C", mockMetricsScope)

	assert.Equal(t, 3, len(metricsCache.scopeMap))

	assert.Equal(t, mockMetricsScope, metricsCache.Get("A"))
	assert.Equal(t, mockMetricsScope, metricsCache.Get("B"))
	assert.Equal(t, mockMetricsScope, metricsCache.Get("C"))
}

func TestConcurrentMetricsScopeAccess(t *testing.T) {
	metricsCache := NewMetricsCache()

	mockMetricsScope := &mocks.Scope{}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		// concurrent get and put
		go func() {
			defer wg.Done()

			metricsCache.Get(strconv.Itoa(i))
			metricsCache.Put(strconv.Itoa(i), mockMetricsScope)
		}()
	}

	wg.Wait()

	assert.Equal(t, 1000, len(metricsCache.scopeMap))

	for i := 0; i < 1000; i++ {
		if metricsCache.scopeMap[strconv.Itoa(i)] == nil {
			t.Error(fmt.Sprintf("Metrics scope not set for %d", i))
		}
	}
}
