// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
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

package serialization

import (
	"bytes"
	"fmt"

	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"

	"github.com/uber/cadence/common"
	p "github.com/uber/cadence/common/persistence"
)

type thriftRWType interface {
	ToWire() (wire.Value, error)
	FromWire(w wire.Value) error
}

func validateEncoding(encoding string) error {
	switch common.EncodingType(encoding) {
	case common.EncodingTypeProto, common.EncodingTypeThriftRW:
		return nil
	default:
		return fmt.Errorf("invalid encoding type: %v", encoding)
	}
}

func encodeErr(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("error serializing struct to blob: %v", err)
}

func decodeErr(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("error deserializing blob to struct: %v", err)
}

func thriftRWEncode(t thriftRWType) (p.DataBlob, error) {
	blob := p.DataBlob{Encoding: common.EncodingTypeThriftRW}
	value, err := t.ToWire()
	if err != nil {
		return blob, encodeErr(err)
	}
	var b bytes.Buffer
	if err := protocol.Binary.Encode(value, &b); err != nil {
		return blob, encodeErr(err)
	}
	blob.Data = b.Bytes()
	return blob, nil
}

func thriftRWDecode(data []byte, result thriftRWType) error {
	value, err := protocol.Binary.Decode(bytes.NewReader(data), wire.TStruct)
	if err != nil {
		return decodeErr(err)
	}
	return decodeErr(result.FromWire(value))
}
