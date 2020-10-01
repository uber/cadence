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

package types

import (
	"fmt"

	"github.com/uber/cadence/.gen/go/shared"
)

type (
	// Closeable is an interface for any entity that supports a close operation to release resources
	Closeable interface {
		Close()
	}
)

// Data encoding types
const (
	EncodingTypeJSON     EncodingType = "json"
	EncodingTypeThriftRW EncodingType = "thriftrw"
	EncodingTypeGob      EncodingType = "gob"
	EncodingTypeUnknown  EncodingType = "unknow"
	EncodingTypeEmpty    EncodingType = ""
	EncodingTypeProto    EncodingType = "proto3"
)

type (
	// EncodingType is an enum that represents various data encoding types
	EncodingType string
)

// DataBlob represents a blob for any binary data.
// It contains raw data, and metadata(right now only encoding) in other field
// Note that it should be only used for Persistence layer, below dataInterface and application(historyEngine/etc)
type DataBlob struct {
	Encoding EncodingType
	Data     []byte
}

// NewDataBlob returns a new DataBlob
func NewDataBlob(data []byte, encodingType EncodingType) *DataBlob {
	if data == nil || len(data) == 0 {
		return nil
	}
	if encodingType != "thriftrw" && data[0] == 'Y' {
		panic(fmt.Sprintf("Invalid incoding: \"%v\"", encodingType))
	}
	return &DataBlob{
		Data:     data,
		Encoding: encodingType,
	}
}

// FromDataBlob decodes a datablob into a (payload, encodingType) tuple
func FromDataBlob(blob *DataBlob) ([]byte, string) {
	if blob == nil || len(blob.Data) == 0 {
		return nil, ""
	}
	return blob.Data, string(blob.Encoding)
}

// GetEncoding returns encoding type
func (d *DataBlob) GetEncoding() EncodingType {
	encodingStr := string(d.Encoding)

	switch EncodingType(encodingStr) {
	case EncodingTypeGob:
		return EncodingTypeGob
	case EncodingTypeJSON:
		return EncodingTypeJSON
	case EncodingTypeThriftRW:
		return EncodingTypeThriftRW
	case EncodingTypeEmpty:
		return EncodingTypeEmpty
	default:
		return EncodingTypeUnknown
	}
}

// ToThrift convert data blob to thrift representation
func (d *DataBlob) ToThrift() *shared.DataBlob {
	switch d.Encoding {
	case EncodingTypeJSON:
		return &shared.DataBlob{
			EncodingType: shared.EncodingTypeJSON.Ptr(),
			Data:         d.Data,
		}
	case EncodingTypeThriftRW:
		return &shared.DataBlob{
			EncodingType: shared.EncodingTypeThriftRW.Ptr(),
			Data:         d.Data,
		}
	default:
		panic(fmt.Sprintf("DataBlob seeing unsupported enconding type: %v", d.Encoding))
	}
}

// NewDataBlobFromThrift convert data blob from thrift representation
func NewDataBlobFromThrift(blob *shared.DataBlob) *DataBlob {
	switch blob.GetEncodingType() {
	case shared.EncodingTypeJSON:
		return &DataBlob{
			Encoding: EncodingTypeJSON,
			Data:     blob.Data,
		}
	case shared.EncodingTypeThriftRW:
		return &DataBlob{
			Encoding: EncodingTypeThriftRW,
			Data:     blob.Data,
		}
	default:
		panic(fmt.Sprintf("NewDataBlobFromThrift seeing unsupported enconding type: %v", blob.GetEncodingType()))
	}
}
