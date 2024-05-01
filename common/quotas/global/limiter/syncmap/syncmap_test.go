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

package syncmap

import (
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
)

func TestBasics(t *testing.T) {

}

func TestNotRacy(t *testing.T) {
	count := atomic.NewInt64(0)
	// using a string pointer just to make things a bit riskier / more sensitive to races since mutation is possible.
	// no mutation currently occurs, but it seems slightly safer to leave it here for future changes.
	m := New(func(key string) *string {
		s := key
		s += "-"
		s += strconv.Itoa(int(count.Inc())) // just to be recognizable
		return &s
	})

	// call ALL the methods concurrently
	var g errgroup.Group
	const loops = 1000 // 100 has had some coverage flapping, raise further if needed
	for i := 0; i < loops; i++ {
		key := strconv.Itoa(i)
		g.Go(func() error {
			v := m.Load(key)
			assert.NotEmpty(t, *v) // "never nil" also asserted by crashing
			return nil
		})
		// try to load the same key multiple times
		g.Go(func() error {
			v := m.Load(key)
			assert.NotEmpty(t, *v)
			return nil
		})
		// range over it while reading/writing
		g.Go(func() error {
			m.Range(func(k string, v *string) bool {
				assert.NotEmpty(t, k)
				assert.NotEmpty(t, *v)
				return true
			})
			return nil
		})
		g.Go(func() error {
			_ = m.Len() // value does not matter / hard to check usefully
			return nil
		})
		g.Go(func() error {
			_, _ = m.Try(key) // value does not matter / hard to check usefully.
			// could count true vs false to ensure both orders are hit...
			return nil
		})
		// delete ~10% of keys to exercise that logic, and mostly ensure coverage
		if rand.Intn(10) == 0 {
			g.Go(func() error {
				m.Delete(key)
				return nil
			})
		}
	}
	require.NoError(t, g.Wait())

	// sanity-check to show decent concurrency:
	// - out-of-order inits (values can be both higher and lower than the key)
	// - duplicate inits (values higher than 100)
	same, higher, lower, upper := 0, 0, 0, int64(0)
	m.Range(func(k string, v *string) bool {
		parts := strings.SplitN(*v, "-", 2)

		// sanity check that keys and values stay associated
		assert.Equal(t, k, parts[0], "key %q and first part of value must match: %q", k, *v)

		if parts[0] == parts[1] {
			same++
		} else if parts[0] < parts[1] {
			higher++
		} else {
			lower++
		}

		vint, err := strconv.ParseInt(parts[1], 10, 64)
		assert.NoError(t, err, "count-%v should be parse-able as an int", parts[1])
		if vint > upper {
			upper = vint
		}
		return true
	})

	assert.LessOrEqual(t,
		int64(loops), upper,
		// regrettably not guaranteed due to deletions, but I have yet to see it.
		// if this becomes an issue, probably just delete it.
		"did not observe a value at least as high as the number of loops.  "+
			"not technically impossible, just very unlikely",
	)

	t.Logf(
		"Metrics for cpu %v:\n"+
			"\tKey == value  (1=>1-1):      %v\n"+
			"\tValue higher  (5=>5-100):    %v\n"+
			"\tValue lower   (100=>100-5):  %v\n"+
			"\tNumber of iterations:        %v\n"+
			"\tHighest saved create:        %v\n"+ // same or higher than iterations
			"\tTotal num of creates:        %v", // same or higher than saved
		runtime.GOMAXPROCS(0), same, higher, lower, loops, upper, count.Load(),
	)
}
