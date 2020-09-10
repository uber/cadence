// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
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

package common

import (
	"github.com/uber/cadence/common/pagination"
)

type (
	persistenceIterator struct {
		itr pagination.Iterator
	}
)

// NewPersistenceIterator returns a new paginated iterator over persistence
func NewPersistenceIterator(
	pf PersistenceFetcher,
) ExecutionIterator {
	pi := &persistenceIterator{}
	pi.itr = pagination.NewIterator(nil, pf.Fetch)
	return pi
}

// Next returns the next execution
func (pi *persistenceIterator) Next() (interface{}, error) {
	exec, err := pi.itr.Next()
	// TODO consider to remove the ExecutionIterator
	if exec != nil {
		return exec, nil
	}
	return nil, err
}

// HasNext returns true if there is another execution, false otherwise.
func (pi *persistenceIterator) HasNext() bool {
	return pi.itr.HasNext()
}
