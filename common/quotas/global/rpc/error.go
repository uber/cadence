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

package rpc

// RPCError wraps errors to mark them as coming from RPC, i.e. basically expected.
type RPCError struct {
	cause error
}

// SerializationError wraps errors to mark them as coming from (de)serialization.
//
// In practice this should not ever occur, as it should imply one of:
// - incorrect Any Value/ValueType pairing, which should not pass tests
// - non-backwards-compatible type changes deployed in an unsafe way
// - incompatible Thrift-binary changes
type SerializationError struct {
	cause error
}
type errwrapper interface {
	error
	Unwrap() error
}

var _ errwrapper = (*RPCError)(nil)
var _ errwrapper = (*SerializationError)(nil)

func (e *RPCError) Error() string           { return e.cause.Error() }
func (e *RPCError) Unwrap() error           { return e.cause }
func (e *SerializationError) Error() string { return e.cause.Error() }
func (e *SerializationError) Unwrap() error { return e.cause }
