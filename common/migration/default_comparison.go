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

package migration

import (
	"reflect"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
)

var _ ComparisonFn[struct{}] = defaultComparisonFn[struct{}]

// Compare is a helper for this implementation specifically which, if implemented
// allows an input struct to describe what is different with the variable that its
// being compared with
type Compare[T any] interface {
	With(Compare[T]) string
}

// DefaultComparisonFn is a simple, suitable-for-low-throughput comparison function which
// uses reflection to determine if the results match. Probably not good for high RPS endpoints
// as it'll cause a significant amount of log noise. This doesn't attempt to print diffs
func defaultComparisonFn[T comparable](
	log log.Logger,
	scope metrics.Scope,
	callsite string,
	activeRes T,
	activeErr error,
	backgroundRes T,
	backgroundErr error) (isEqual bool) {

	log = log.WithTags(tag.Callsite(callsite))

	if activeErr != nil && backgroundErr != nil {

		errsAreEqual := activeErr == backgroundErr
		if errsAreEqual {
			scope.IncCounter(metrics.ShadowerComparisonMatchCount)
			return true
		}
		log.Warn("mismatched result with both flows returning an error")
		scope.IncCounter(metrics.ShadowerComparisonDifferenceCount)
		return false
	}

	if backgroundErr != nil && activeErr == nil {
		log.Warn("mismatched result with background flow returning an error, whereas the active flow succeeded",
			tag.Error(backgroundErr))
		scope.IncCounter(metrics.ShadowerComparisonDifferenceCount)
		return false
	}

	if activeErr != nil && backgroundErr == nil {
		log.Warn("mismatched result with active flow returning an error, whereas the background flow succeeded",
			tag.Error(activeErr))
		scope.IncCounter(metrics.ShadowerComparisonDifferenceCount)
		return false
	}

	if reflect.DeepEqual(activeRes, backgroundRes) {
		scope.IncCounter(metrics.ShadowerComparisonMatchCount)
		return true
	}

	diff := getDiff(activeRes, backgroundRes)

	log.Warn("mismatched results between active and background", tag.Dynamic("comparison-diff", diff))
	scope.IncCounter(metrics.ShadowerComparisonDifferenceCount)
	return false
}

func getDiff[T any](active T, background T) string {

	a1, ok := any(active).(Compare[T])
	if !ok {
		return "couldn't compare active"
	}

	b1, ok := any(background).(Compare[T])
	if !ok {
		return "couldn't compare background"
	}
	return a1.With(b1)
}
