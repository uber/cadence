package migration

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"sync"
	"testing"
	"time"
)

func TestControllerImpl_ReadAndReturnActive_BothNewAndOrdFlowsWellBehaved_ReturningDataFromOldFlow(t *testing.T) {

	type testStruct struct {
		A string
		B *int
		c bool
	}

	dataReturnedByNew := &testStruct{
		A: "A",
		B: nil,
		c: false,
	}

	dataReturnedByOld := &testStruct{
		A: "B",
		B: nil,
		c: false,
	}

	newOp := func(ctx context.Context) (*testStruct, error) {
		time.Sleep(time.Millisecond * 10)
		return dataReturnedByNew, nil
	}

	oldOp := func(ctx context.Context) (*testStruct, error) {
		time.Sleep(time.Millisecond * 10)
		return dataReturnedByOld, nil
	}

	constraints := Constraints{Domain: "testdomain"}
	parentCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	scope := metrics.NewNoopMetricsClient().Scope(123)

	comparator := func(log log.Logger, scope metrics.Scope, active *testStruct, activeErr error, background *testStruct, backgroundErr error) bool {
		assert.Equal(t, dataReturnedByOld, active, "The data given to the comparison function didn't match that of the old flow as expected")
		assert.Equal(t, dataReturnedByNew, background, "The data given to the comparison function didn't match that of the new flow as expected")

		assert.NoError(t, activeErr, "no error was expected on the active flow")
		assert.NoError(t, backgroundErr, background, "no error was expected on the background flow")

		comarisonRes := defaultComparisonFn[*testStruct](log, scope, active, activeErr, background, backgroundErr)

		assert.False(t, comarisonRes, "the default camparitor didn't return the expected result. In this test scenario the results are expected to not match")
		return comarisonRes
	}

	reader := NewDualReaderWithCustomComparisonFn[testStruct](
		func(_ ...dynamicconfig.FilterOption) string { return string(ReaderRolloutCallBothAndReturnOld) },
		func(_ ...dynamicconfig.FilterOption) bool { return true },
		loggerimpl.NewNopLogger(),
		scope,
		time.Millisecond*200,
		comparator)

	res, err := reader.ReadAndReturnActive(parentCtx, constraints, oldOp, newOp)

	assert.NoError(t, err)
	assert.Equal(t, dataReturnedByOld, res)
}

func TestControllerImpl_ReadAndReturnActive_BothNewAndOrdFlowsWellBehaved_ReturningDataFromNewFlow(t *testing.T) {

	type testStruct struct {
		A string
		B *int
		c bool
	}

	dataReturnedByNew := &testStruct{
		A: "A",
		B: nil,
		c: false,
	}

	dataReturnedByOld := &testStruct{
		A: "B",
		B: nil,
		c: false,
	}

	newOp := func(ctx context.Context) (*testStruct, error) {
		time.Sleep(time.Millisecond * 10)
		return dataReturnedByNew, nil
	}

	oldOp := func(ctx context.Context) (*testStruct, error) {
		time.Sleep(time.Millisecond * 10)
		return dataReturnedByOld, nil
	}

	constraints := Constraints{Domain: "testdomain"}
	parentCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	scope := metrics.NewNoopMetricsClient().Scope(123)

	comparator := func(log log.Logger, scope metrics.Scope, active *testStruct, activeErr error, background *testStruct, backgroundErr error) bool {
		assert.Equal(t, dataReturnedByOld, active, "The data given to the comparison function didn't match that of the old flow as expected")
		assert.Equal(t, dataReturnedByNew, background, "The data given to the comparison function didn't match that of the new flow as expected")

		assert.NoError(t, activeErr, "no error was expected on the active flow")
		assert.NoError(t, backgroundErr, background, "no error was expected on the background flow")

		comparisonRes := defaultComparisonFn[*testStruct](log, scope, active, activeErr, background, backgroundErr)

		assert.False(t, comparisonRes, "the default camparator didn't return the expected result. In this test scenario the results are expected to not match")
		return comparisonRes
	}

	reader := NewDualReaderWithCustomComparisonFn[testStruct](
		func(_ ...dynamicconfig.FilterOption) string { return string(ReaderRolloutCallBothAndReturnNew) },
		func(_ ...dynamicconfig.FilterOption) bool { return true },
		loggerimpl.NewNopLogger(),
		scope,
		time.Millisecond*200,
		comparator)

	res, err := reader.ReadAndReturnActive(parentCtx, constraints, oldOp, newOp)

	assert.NoError(t, err)
	assert.Equal(t, dataReturnedByNew, res)
}

func TestControllerImpl_ReadAndReturnActive_TimeoutAndErrorHandling(t *testing.T) {

	type testStruct struct {
		A string
		B *int
		c bool
	}

	dataReturnedA := &testStruct{
		A: "A",
		B: nil,
		c: false,
	}

	dataReturnedB := &testStruct{
		A: "B",
		B: nil,
		c: false,
	}

	opFastA := func(ctx context.Context) (*testStruct, error) {
		time.Sleep(time.Millisecond * 10)
		return dataReturnedA, nil
	}

	opSlowA := func(ctx context.Context) (*testStruct, error) {
		time.Sleep(time.Second * 10)
		return dataReturnedA, nil
	}

	opFastB := func(ctx context.Context) (*testStruct, error) {
		time.Sleep(time.Millisecond * 10)
		return dataReturnedB, nil
	}

	opSlowB := func(ctx context.Context) (*testStruct, error) {
		time.Sleep(time.Second * 10)
		return dataReturnedB, nil
	}

	opSlowErr := func(ctx context.Context) (*testStruct, error) {
		time.Sleep(time.Second * 10)
		return nil, errors.New("A slow error")
	}

	constraints := Constraints{Domain: "testdomain"}

	parentCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	scope := metrics.NewNoopMetricsClient().Scope(123)

	comparatorAssertEqual := func(t *testing.T, wg *sync.WaitGroup) ComparisonFn[*testStruct] {
		return func(log log.Logger, scope metrics.Scope, active *testStruct, activeErr error, background *testStruct, backgroundErr error) bool {
			comparisonRes := defaultComparisonFn[*testStruct](log, scope, active, activeErr, background, backgroundErr)
			assert.Equal(t, active, background)
			assert.Equal(t, activeErr, backgroundErr)
			wg.Done()
			return comparisonRes
		}
	}

	comparatorAssertNotEqual := func(t *testing.T, wg *sync.WaitGroup) ComparisonFn[*testStruct] {
		return func(log log.Logger, scope metrics.Scope, active *testStruct, activeErr error, background *testStruct, backgroundErr error) bool {
			comparisonRes := defaultComparisonFn[*testStruct](log, scope, active, activeErr, background, backgroundErr)
			assert.NotEqual(t, active, background)
			wg.Done()
			return comparisonRes
		}
	}

	tests := map[string]struct {
		operationNew func(ctx context.Context) (*testStruct, error)
		operationOld func(ctx context.Context) (*testStruct, error)

		migrationStatus ReaderRolloutState
		comparator      func(t *testing.T, group *sync.WaitGroup) ComparisonFn[*testStruct]

		expectedResult *testStruct
		expectedErr    error
	}{
		"Both operations fast, no errors, happy path, matching responses": {
			operationNew:    opFastA,
			operationOld:    opFastA,
			migrationStatus: ReaderRolloutCallBothAndReturnOld,
			comparator:      comparatorAssertEqual,
			expectedResult:  dataReturnedA,
		},
		"slow operation in background, no errors, matching responses": {
			operationNew:    opSlowA,
			operationOld:    opFastA,
			migrationStatus: ReaderRolloutCallBothAndReturnOld,
			comparator:      comparatorAssertNotEqual, // we expect the slow one to be error'd
			expectedResult:  dataReturnedA,
		},
		"Both operations fast, no errors, happy path, differing responses.": {
			operationNew:    opFastB,
			operationOld:    opFastA,
			migrationStatus: ReaderRolloutCallBothAndReturnOld,
			comparator:      comparatorAssertNotEqual,
			expectedResult:  dataReturnedA,
		},
		"slow operation in background, no errors, differing responses": {
			operationNew:    opSlowB,
			operationOld:    opFastA,
			migrationStatus: ReaderRolloutCallBothAndReturnOld,
			comparator:      comparatorAssertNotEqual,
			expectedResult:  dataReturnedA,
		},
		"slow operation in background with errors, differing responses": {
			operationNew:    opSlowErr,
			operationOld:    opFastA,
			migrationStatus: ReaderRolloutCallBothAndReturnOld,
			comparator:      comparatorAssertNotEqual,
			expectedResult:  dataReturnedA,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {

			start := time.Now()
			wg := sync.WaitGroup{}

			wg.Add(1)

			reader := NewDualReaderWithCustomComparisonFn[testStruct](
				func(_ ...dynamicconfig.FilterOption) string { return string(td.migrationStatus) },
				func(_ ...dynamicconfig.FilterOption) bool { return true },
				loggerimpl.NewNopLogger(),
				scope,
				time.Millisecond*200,
				td.comparator(t, &wg))

			res, err := reader.ReadAndReturnActive(parentCtx, constraints, td.operationOld, td.operationNew)

			intialCallEnd := time.Now()

			// wait for the comparison
			wg.Wait()

			comparisonEnd := time.Now()

			fmt.Println(".>", comparisonEnd.Sub(start))

			// 200 ms timeout
			assert.True(t, intialCallEnd.Sub(start) < time.Millisecond*210)
			assert.True(t, comparisonEnd.Sub(start) < time.Millisecond*300) // the comparison with all it's reflection is quite slow

			assert.Equal(t, td.expectedResult, res)
			assert.Equal(t, td.expectedErr, err)
		})
	}

}
