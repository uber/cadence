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

package testdatagen

import (
	fuzz "github.com/google/gofuzz"

	"github.com/uber/cadence/common/types"
)

// GenHistoryEvent is a function to use with gofuzz which
// skips the majority of difficult to generate values
// for the sake of simplicity in testing. Use it with the fuzz.Funcs(...) generation function
func GenHistoryEvent(o *types.HistoryEvent, c fuzz.Continue) {
	t1 := int64(1704127933)
	e1 := types.EventType(1)
	*o = types.HistoryEvent{
		ID:        1,
		Timestamp: &t1,
		EventType: &e1,
		Version:   2,
		TaskID:    3,
		// todo (david.porter) Find a better way to handle substructs
	}
	return
}
