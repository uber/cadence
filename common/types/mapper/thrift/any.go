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

package thrift

import (
	"bytes"
	"fmt"

	"go.uber.org/thriftrw/protocol/binary"
	"go.uber.org/thriftrw/wire"
)

// any.go contains Any-helpers for encoding/decoding thrift data

// ThriftObject is satisfied by any thrift type, as they all have To/FromWire methods.
// go.uber.org/cadence has exactly this in an internal API, it's just copied here for simplicity.
type ThriftObject interface {
	FromWire(w wire.Value) error
	ToWire() (wire.Value, error)
}

func EncodeToBytes(obj ThriftObject) ([]byte, error) {
	// get the intermediate format, ready for byte-encoding
	wireValue, err := obj.ToWire()
	if err != nil {
		return nil, fmt.Errorf("failed to produce intermediate wire.Value from type %T", obj)
	}
	// write to binary
	var writer bytes.Buffer
	err = binary.Default.Encode(wireValue, &writer)
	if err != nil {
		return nil, fmt.Errorf("failed to encode wire.Value from type %T", obj)
	}
	// and return the bytes
	return writer.Bytes(), nil
}

func DecodeStructFromBytes(data []byte, target ThriftObject) error {
	// decode to the intermediate format, like encoding.
	// all top-most types likely *should* be a wire.TStruct, but if Any is at some point used
	// to contain a map/int/etc, just make sure to use wire.TMap / wire.TI64 / etc as is appropriate.
	intermediate, err := binary.Default.Decode(bytes.NewReader(data), wire.TStruct)
	if err != nil {
		return fmt.Errorf("could not decode thrift bytes to intermediate wire.Value for type %T", target)
	}
	// and populate the target with the wire.Value contents
	err = target.FromWire(intermediate)
	if err != nil {
		return fmt.Errorf("could not FromWire to the target type %T", target)
	}
	return nil
}
