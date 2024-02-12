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

package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/types"
)

func TestDecode(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name           string
		blob           *types.DataBlob
		want           *testStruct
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "valid JSON encoding",
			blob: &types.DataBlob{
				Data:         []byte(`{"name":"test"}`),
				EncodingType: types.EncodingTypeJSON.Ptr(),
			},
			want:    &testStruct{Name: "test"},
			wantErr: false,
		},
		{
			name: "unsupported encoding type",
			blob: &types.DataBlob{
				Data:         []byte("aa"),
				EncodingType: types.EncodingTypeThriftRW.Ptr(),
			},
			want:           nil,
			wantErr:        true,
			expectedErrMsg: "unsupported encoding type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := newDecoder(tt.blob)
			var got testStruct
			err := decoder.Decode(&got)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, &got)
			}
		})
	}
}
