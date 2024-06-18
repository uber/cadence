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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/testdata"
)

func TestThriftInAny(t *testing.T) {
	// pushing thrift or proto data through this type is very much expected,
	// so this is both evidence that it's possible and an example for how to do so.
	//
	// note that there is an equivalent test in mapper/proto/shared_test.go.
	// they are structurally the same, but encoding/decoding details vary a bit.

	// --- create the original data, a thrift object, and encode it by hand
	orig := &shared.WorkflowExecution{
		WorkflowId: common.StringPtr(testdata.WorkflowID),
		RunId:      common.StringPtr(testdata.RunID),
	}
	internalBytes, err := EncodeToBytes(orig)
	require.NoError(t, err)

	// --- put that data into the custom Any type
	// thrift has no registry like proto has, so there is no canonical name.
	// no problem, just generate one.
	var typeName = reflect.TypeOf(orig).PkgPath() + reflect.TypeOf(orig).Name()
	anyVal := &types.Any{
		ValueType: typeName,
		Value:     internalBytes, // store thrift bytes in the Any
	}

	// --- convert the whole container to thrift (mimics making a call via yarpc)
	thriftAny := FromAny(anyVal)                  // we map to the rpc type
	networkBytes, err := EncodeToBytes(thriftAny) // yarpc does this
	require.NoError(t, err)
	// ^ this is what's sent over the network.

	// as a side note:
	// the final data is not double-encoded, so this "encode -> wrap -> encode" process is reasonably efficient.
	//
	// Thrift and Proto can efficiently move around binary blobs like this, as it's essentially just a memcpy between
	// the input and the output, and there's no `\0` to `\\0` escaping or base64 encoding or whatever needed.
	//
	// no behavior depends on this, it's just presented here as evidence that this Any-wrapper does not meaningfully
	// change any RPC design concerns: anything you would do with normal RPC can be done through an Any if you need
	// loose typing, the change-stability / performance / etc is entirely unaffected.
	//
	// compare via a sliding window to find the place it overlaps, to prove that this is true:
	found := false
	for i := 0; i <= len(networkBytes)-len(internalBytes); i++ {
		if reflect.DeepEqual(internalBytes, networkBytes[i:i+len(internalBytes)]) {
			found = true
			t.Logf("Found matching bytes at index %v", i) // currently at index 14
		}
	}
	// *should* be true for efficiency's sake, but is not truly necessary for correct behavior
	assert.Truef(t, found, "did not find internal bytes within network bytes, might be paying double-encoding costs:\n\tinternal: %v\n\tnetwork:  %v", internalBytes, networkBytes)

	// --- the network pushes the data to a new location ---

	// --- on the receiving side, we map to internal types like normal
	var outAny shared.Any
	err = DecodeStructFromBytes(networkBytes, &outAny) // yarpc does this
	require.NoError(t, err)
	outAnyVal := ToAny(&outAny) // we map to internal types

	// --- and finally decode the any-typed data by hand
	require.Equal(t, typeName, outAnyVal.ValueType, "type name through RPC should match the original type name")
	var outOrig shared.WorkflowExecution                   // selected based on the ValueType contents
	err = DecodeStructFromBytes(outAnyVal.Value, &outOrig) // do the actual custom decoding
	require.NoError(t, err)
	assert.NotEmpty(t, outOrig, "sanity check, decoded value should not be empty")
	assert.Equal(t, orig, &outOrig, "final round-tripped Any-contained data should be identical to the original object")
}
