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

package testutils

import (
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
)

func EnsureFuzzCoverage(t *testing.T, expected []string, cb func(t *testing.T, f *fuzz.Fuzzer) string) {
	t.Helper()

	var details []string
	results := make(map[string]bool, len(expected))
	for _, e := range expected {
		results[e] = false
	}
	seed := time.Now().UnixNano()
	f := fuzz.NewWithSeed(seed) // helps with troubleshooting

	defer func() {
		if t.Failed() { // else a bit noisy
			t.Logf("expected to see:  %#v", expected)
			t.Logf("observed results: %#v", results)
			t.Logf("detailed results: %#v", details)
			t.Logf("fuzz seed: %v", seed)
		}
	}()

	for tries := 0; tries < 100; tries++ { // retry a few times if needed
		for i := 0; i < 100; i++ { // always fuzz a moderate amount, don't stop immediately
			res := cb(t, f)
			details = append(details, res)
			if res == "" {
				t.Errorf("invalid empty response from fuzzing callback on iteration %v", (tries*100)+i)
			}
			if _, ok := results[res]; !ok {
				t.Errorf("unrecognized response from fuzzing callback on iteration %v: %v", (tries*100)+i, res)
			}
			if t.Failed() {
				return // already failed either internally or in the callback, either way stop trying
			}
			results[res] = true
		}
		stop := true
		for _, v := range results {
			stop = stop && v
		}
		if stop {
			return // covered all expected values, stop retrying
		}
	}
	missing := make([]string, 0, len(results))
	for _, v := range expected {
		if !results[v] {
			missing = append(missing, v)
		}
	}
	t.Errorf("fuzzy coverage func did not check enough cases after 10k attempts, missing: %#v", missing)
}
