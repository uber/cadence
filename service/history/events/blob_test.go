// The MIT License (MIT)

// Copyright (c) 2022 Uber Technologies Inc.

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

package events

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/persistence"
)

func TestPersistedBlobs_Find(t *testing.T) {
	blob1 := persistence.DataBlob{Data: []byte{1, 2, 3}}
	blob2 := persistence.DataBlob{Data: []byte{4, 5, 6}}
	blob3 := persistence.DataBlob{Data: []byte{7, 8, 9}}
	branchA := []byte{11, 11, 11}
	branchB := []byte{22, 22, 22}
	persistedBlobs := PersistedBlobs{
		PersistedBlob{BranchToken: branchA, FirstEventID: 100, DataBlob: blob1},
		PersistedBlob{BranchToken: branchA, FirstEventID: 105, DataBlob: blob2},
		PersistedBlob{BranchToken: branchB, FirstEventID: 100, DataBlob: blob3},
	}
	assert.Equal(t, blob1, *persistedBlobs.Find(branchA, 100))
	assert.Equal(t, blob2, *persistedBlobs.Find(branchA, 105))
	assert.Equal(t, blob3, *persistedBlobs.Find(branchB, 100))
	assert.Nil(t, persistedBlobs.Find(branchB, 105))
	assert.Nil(t, persistedBlobs.Find([]byte{99}, 100))
}
