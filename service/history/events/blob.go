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
	"bytes"

	"github.com/uber/cadence/common/persistence"
)

type (
	// PersistedBlob is a wrapper on persistence.DataBlob with additional field indicating what was persisted.
	// Additional fields are used as an identification key among other blobs.
	PersistedBlob struct {
		persistence.DataBlob

		BranchToken  []byte
		FirstEventID int64
	}
	// PersistedBlobs is a slice of PersistedBlob
	PersistedBlobs []PersistedBlob
)

// Find searches for persisted event blob. Returns nil when not found.
func (blobs PersistedBlobs) Find(branchToken []byte, firstEventID int64) *persistence.DataBlob {
	// Linear search is ok here, as we will only have 1-2 persisted blobs per transaction
	for _, blob := range blobs {
		if bytes.Equal(blob.BranchToken, branchToken) && blob.FirstEventID == firstEventID {
			return &blob.DataBlob
		}
	}
	return nil
}
