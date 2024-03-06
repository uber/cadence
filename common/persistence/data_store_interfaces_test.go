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

package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

func TestDataBlob(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		assert.Nil(t, NewDataBlob(nil, "anything"), "nil data should become nil blob")
		data, encoding := []byte("something"), common.EncodingTypeJSON
		assert.EqualValues(t, &DataBlob{
			Data:     data,
			Encoding: encoding,
		}, NewDataBlob(data, encoding))
	})
	t.Run("Y-prefixed data may panic", func(t *testing.T) {
		allEncodings := []common.EncodingType{
			common.EncodingTypeJSON,
			common.EncodingTypeThriftRW,
			common.EncodingTypeGob,
			common.EncodingTypeUnknown,
			common.EncodingTypeEmpty,
			common.EncodingTypeProto,
		}
		const problematic = "Y..."

		// I'm not entirely sure why this behavior exists, but to nail it down:
		assert.NotPanics(t, func() {
			NewDataBlob([]byte(problematic), common.EncodingTypeThriftRW)
		}, "only thriftrw data can start with Y without panicking")
		// all others panic
		for _, encoding := range allEncodings {
			if encoding == common.EncodingTypeThriftRW {
				continue // handled above
			}
			assert.Panicsf(t, func() {
				NewDataBlob([]byte(problematic), encoding)
			}, "non-thriftrw %v should panic if first byte is Y", encoding)
		}

		// make sure other data/encoding combinations do not panic for any encoding
		for _, dat := range []string{"Z...", "other"} {
			for _, encoding := range allEncodings {
				assert.NotPanicsf(t, func() {
					NewDataBlob([]byte(dat), encoding)
				}, "should not panic with %q on encoding %q", dat, encoding)
			}
		}
	})
	t.Run("from", func(t *testing.T) {
		assertEmpty := func(b []byte, s string) {
			assert.Nil(t, b, "should have nil bytes")
			assert.Empty(t, s, "should have empty string")
		}
		assertEmpty(FromDataBlob(nil))
		assertEmpty(FromDataBlob(&DataBlob{}))
		assertEmpty(FromDataBlob(&DataBlob{Encoding: common.EncodingTypeThriftRW}))

		dat, str := FromDataBlob(&DataBlob{
			Data:     []byte("data"),
			Encoding: common.EncodingTypeJSON,
		})
		assert.Equal(t, []byte("data"), dat, "data should be returned")
		assert.Equal(t, string(common.EncodingTypeJSON), str, "encoding should be returned as a string")
	})
	t.Run("ToNilSafeDataBlob", func(t *testing.T) {
		orig := &DataBlob{Encoding: common.EncodingTypeGob} // anything not empty
		assert.Equal(t, orig, orig.ToNilSafeDataBlob())
		assert.NotNil(t, (*DataBlob)(nil).ToNilSafeDataBlob(), "typed nils should convert to non-nils")
	})
	t.Run("GetEncodingString", func(t *testing.T) {
		assert.Equal(t, "", (*DataBlob)(nil).GetEncodingString())
		assert.Equal(t, "test", (&DataBlob{Encoding: "test"}).GetEncodingString())
	})
	t.Run("GetData", func(t *testing.T) {
		// this method returns empty slices, not nils.
		// I'm not sure if this needs to be maintained, but it must be checked before changing.
		assert.Equal(t, []byte{}, (*DataBlob)(nil).GetData())
		assert.Equal(t, []byte{}, (&DataBlob{}).GetData())
		assert.Equal(t, []byte{}, (&DataBlob{Data: []byte{}}).GetData())
		assert.Equal(t, []byte("test"), (&DataBlob{Data: []byte("test")}).GetData())
	})
	t.Run("GetEncoding", func(t *testing.T) {
		same := func(encoding common.EncodingType) {
			assert.Equal(t, encoding, (&DataBlob{Encoding: encoding}).GetEncoding())
		}
		unknown := func(encoding common.EncodingType) {
			assert.Equal(t, common.EncodingTypeUnknown, (&DataBlob{Encoding: encoding}).GetEncoding())
		}
		// obvious
		same(common.EncodingTypeGob)
		same(common.EncodingTypeJSON)
		same(common.EncodingTypeThriftRW)
		same(common.EncodingTypeEmpty)

		// highly suspicious
		unknown(common.EncodingTypeProto)

		// should be unknown
		unknown(common.EncodingTypeUnknown)
		unknown("any other value")
	})
	t.Run("To and FromInternal", func(t *testing.T) {
		data := []byte("some data")
		// safe
		internal := &types.DataBlob{
			EncodingType: types.EncodingTypeJSON.Ptr(),
			Data:         data,
		}
		blob := &DataBlob{
			Encoding: common.EncodingTypeJSON,
			Data:     data,
		}
		assert.EqualValues(t, internal, blob.ToInternal(), "json should encode to internal type")
		assert.Equal(t, blob, NewDataBlobFromInternal(internal), "json should decode from internal type")
		// safe
		internal.EncodingType = types.EncodingTypeThriftRW.Ptr()
		blob.Encoding = common.EncodingTypeThriftRW
		assert.EqualValues(t, internal, blob.ToInternal(), "thriftrw should encode to internal type")
		assert.Equal(t, blob, NewDataBlobFromInternal(internal), "thriftrw should decode from internal type")

		// all these panic
		for _, encoding := range []common.EncodingType{
			common.EncodingTypeUnknown,
			common.EncodingTypeProto,
			common.EncodingTypeGob,
			common.EncodingTypeEmpty,
			"any other value",
		} {
			assert.Panicsf(t, func() {
				(&DataBlob{
					Encoding: encoding,
					Data:     data,
				}).ToInternal()
			}, "should panic when encoding to unhandled encoding %q", encoding)
		}

		unknownType := types.EncodingType(1234)
		// sanity check that the value actually is unknown
		require.NotEqual(t, types.EncodingTypeJSON, unknownType)
		require.NotEqual(t, types.EncodingTypeThriftRW, unknownType)
		assert.Panicsf(t, func() {
			NewDataBlobFromInternal(&types.DataBlob{
				EncodingType: &unknownType,
				Data:         data,
			}).ToInternal()
		}, "should panic when decoding from unhandled encoding %q", unknownType)
	})
}
