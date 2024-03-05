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

package idlfuzzedtestdata

import (
	"testing"

	fuzz "github.com/google/gofuzz"

	"github.com/uber/cadence/common/testing/testdatagen"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/testdata"
)

// NewFuzzerWithIDLTypes creates a new fuzzer, notes down the deterministic seed
// this particular invocation is preconfigured to be able to handle idl structs
// correctly without generating completely invalid data (which, while good to test for
// in the context of an application is too wide a search to be useful)
func NewFuzzerWithIDLTypes(t *testing.T) *fuzz.Fuzzer {
	return testdatagen.New(t,
		// USE THESE VERY SPARINGLY, ONLY WHEN YOU MUST!
		//
		// The goal of providing these generators for specific types should be
		// to use them as little as possible, as they are fixed test data
		// which will not evolve with the idl or functions, therefore
		// the main benefit of fuzzing - evolving tests to handle all new fields in place -
		// will be defeated.
		//
		// for example, for mappers, if you add a new field that needs to be
		// mapped from protobuf to a native-go type (from the types folder)
		// and the testdata is fixed here *and not updated*, then the issue
		// will not be caught by any roundtrip tests.
		GenHistoryEvent,
	)
}

// GenHistoryEvent is a function to use with gofuzz which
// skips the majority of difficult to generate values
// for the sake of simplicity in testing. Use it with the fuzz.Funcs(...) generation function
func GenHistoryEvent(o *types.HistoryEvent, c fuzz.Continue) {
	// todo (david.porter) setup an assertion to ensure this list is exhaustive
	i := c.Rand.Intn(len(testdata.HistoryEventArray) - 1)
	o = testdata.HistoryEventArray[i]
	return
}
