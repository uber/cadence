package migration

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"reflect"
)

var _ ComparisonFn[struct{}] = defaultComparisonFn[struct{}]

// DefaultComparisonFn is a simple, suitable-for-low-throughput comparison function which
// uses reflection to determine if the results match. Probably not good for high RPS endpoints
// as it'll cause a significant amount of log noise. Can cause panics due to the use of the
// cmp library (which panics for a number of conditions) so *not* for use without a recover.
//
// This function only compares structs/pointers to structs and errors. It will not attempt to compare
// strings or primitive types and will just state them as not matching.
func defaultComparisonFn[T any](
	log log.Logger,
	scope metrics.Scope,
	activeRes T,
	activeErr error,
	backgroundRes T,
	backgroundErr error) (isEqual bool) {

	if activeErr != nil && backgroundErr != nil {
		errDiff := cmp.Diff(activeErr, backgroundErr, cmpopts.EquateErrors())
		if errDiff != "" {
			log.Warn("mismatched result with both flows returning an error", tag.Dynamic("comparison-result-errdiff", errDiff))
			scope.IncCounter(metrics.ShadowerComparisonDifferenceCount)
			return false
		} else {
			scope.IncCounter(metrics.ShadowerComparisonMatchCount)
			return true
		}
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

	// so, the problem here is that cmp.Diff doesn't love to receive pointers, and tends to panic
	// for a fairly wide range of inputs when it hits a panic. This is not a test system, but rather
	// an online flow, and so the use of a generic reflection tool is both slow and not particularly suited
	// to this usecase, and quite dangerous insofar as that it panics a lot.
	//
	// That said, the ability to spot mapping differences, serialization differences and similar through its deep
	// reflection is invaluable, so while this is a dangerous tradeoff, it's chosen here carefully, with the knowledge
	// that this comparison is to be used in a context where panics are recovered and precautions such as the below are
	// taken to try and catch the egregious cases which will cause panics.

	// Firstly, use a bit of reflection to work out of the given types are pointers
	// and try and derefrence them. this is not super robust, fairly obviously pointers to pointers
	// will panic
	backgroundVal, d1, err := followPtrsAndReturnValue(backgroundRes)
	if err != nil {
		return false // fail to compare due to being unable to deref safely
	}

	activeVal, d2, err := followPtrsAndReturnValue(activeRes)
	if err != nil {
		return false // fail to compare due to being unable to deref safely
	}

	if d1 != d2 {
		log.Warn("shadow flow got mismatched results, the results have a different number of pointers. Probably one of the values is a pointer and the other a value")
		scope.IncCounter(metrics.ShadowerComparisonDifferenceCount)
		return false
	}

	if backgroundVal == nil && activeVal == nil {
		scope.IncCounter(metrics.ShadowerComparisonMatchCount)
		return true
	}

	// and then go through and short-circuit if anything doesn't appear to be a struct
	backgroundType := reflect.TypeOf(backgroundVal)
	activeType := reflect.TypeOf(activeVal)

	if backgroundType == nil || backgroundType.Kind() != reflect.Struct ||
		activeType == nil || activeType.Kind() != reflect.Struct {

		log.Warn("shadow flow got mismatched results, the types to compare are not struct or pointer to structs")
		scope.IncCounter(metrics.ShadowerComparisonDifferenceCount)
		return false
	}

	diff := cmp.Diff(
		backgroundVal,
		activeVal,
		cmpopts.IgnoreUnexported(backgroundVal),
		cmpopts.IgnoreUnexported(activeVal))

	if diff != "" {
		log.Warn("mismatched results between active and background",
			tag.Dynamic("comparison-result-res-diff", diff))
		scope.IncCounter(metrics.ShadowerComparisonDifferenceCount)
		return false
	} else {
		scope.IncCounter(metrics.ShadowerComparisonMatchCount)
		return true
	}
}

func followPtrsAndReturnValue(in interface{}) (out interface{}, derefCount int, err error) {
	derefCount = 0

	out = in
	for {
		rv := reflect.ValueOf(out)
		if rv.Kind() != reflect.Ptr {
			break
		}
		derefCount++ // prevent infinite loops following pointers
		if derefCount > 10 {
			return nil, derefCount, fmt.Errorf("unable to find value, too many pointers")
		}
		out = reflect.Indirect(rv).Interface()
	}
	return out, derefCount, nil
}
