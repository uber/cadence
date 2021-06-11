// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package future

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
)

type (
	// Future represents a value that will be available in the future
	Future interface {
		Get(ctx context.Context, valuePtr interface{}) error
		IsReady() bool
	}

	// Settable is for setting value and error for the corresponding Future
	// The Set method can only be called once. Later calls will result in a
	// panic since the value and error has already been set.
	Settable interface {
		Set(interface{}, error)
	}

	futureImpl struct {
		value   interface{}
		err     error
		readyCh chan struct{}
		status  int32
	}
)

const (
	valueNotSet = iota
	valueSet
)

// NewFuture creates a new Future and the Settable for setting its value
func NewFuture() (Future, Settable) {
	future := &futureImpl{
		readyCh: make(chan struct{}),
		status:  valueNotSet,
	}
	return future, future
}

func (f *futureImpl) Get(
	ctx context.Context,
	valuePtr interface{},
) error {
	if err := ctx.Err(); err != nil {
		// if the given context is invalid,
		// guarantee to return an error
		return err
	}

	select {
	case <-f.readyCh:
		return f.populateValue(valuePtr)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (f *futureImpl) IsReady() bool {
	select {
	case <-f.readyCh:
		return true
	default:
		return false
	}
}

func (f *futureImpl) Set(
	value interface{},
	err error,
) {
	if !atomic.CompareAndSwapInt32(&f.status, valueNotSet, valueSet) {
		panic("future has already been set")
	}

	f.value = value
	f.err = err
	close(f.readyCh)
}

func (f *futureImpl) populateValue(
	valuePtr interface{},
) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("failed to populate valuePtr: %v", p)
		}
	}()

	if f.err != nil || f.value == nil || valuePtr == nil {
		return f.err
	}

	rf := reflect.ValueOf(valuePtr)
	if rf.Type().Kind() != reflect.Ptr {
		return errors.New("valuePtr parameter is not a pointer")
	}

	fv := reflect.ValueOf(f.value)
	if fv.IsValid() {
		rf.Elem().Set(fv)
	}
	return nil
}
