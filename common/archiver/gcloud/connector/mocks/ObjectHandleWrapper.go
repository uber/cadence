// Copyright (c) 2020 Uber Technologies, Inc.
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

// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	storage "cloud.google.com/go/storage"
	mock "github.com/stretchr/testify/mock"

	connector "github.com/uber/cadence/common/archiver/gcloud/connector"
)

// ObjectHandleWrapper is an autogenerated mock type for the ObjectHandleWrapper type
type ObjectHandleWrapper struct {
	mock.Mock
}

// Attrs provides a mock function with given fields: ctx
func (_m *ObjectHandleWrapper) Attrs(ctx context.Context) (*storage.ObjectAttrs, error) {
	ret := _m.Called(ctx)

	var r0 *storage.ObjectAttrs
	if rf, ok := ret.Get(0).(func(context.Context) *storage.ObjectAttrs); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.ObjectAttrs)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewReader provides a mock function with given fields: ctx
func (_m *ObjectHandleWrapper) NewReader(ctx context.Context) (connector.ReaderWrapper, error) {
	ret := _m.Called(ctx)

	var r0 connector.ReaderWrapper
	if rf, ok := ret.Get(0).(func(context.Context) connector.ReaderWrapper); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(connector.ReaderWrapper)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewWriter provides a mock function with given fields: ctx
func (_m *ObjectHandleWrapper) NewWriter(ctx context.Context) connector.WriterWrapper {
	ret := _m.Called(ctx)

	var r0 connector.WriterWrapper
	if rf, ok := ret.Get(0).(func(context.Context) connector.WriterWrapper); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(connector.WriterWrapper)
		}
	}

	return r0
}
