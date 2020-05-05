// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
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

package common

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/pagination"
)

type (
	blobstoreIterator struct {
		itr pagination.Iterator
	}
)

// NewBlobstoreIterator constructs a new iterator backed by blobstore.
func NewBlobstoreIterator(
	client blobstore.Client,
	keys Keys,
) ExecutionIterator {
	return &blobstoreIterator{
		itr: pagination.NewIterator(keys.MinPage, getBlobstoreFetchPageFn(client, keys)),
	}
}

// Next returns the next Execution
func (i *blobstoreIterator) Next() (*Execution, error) {
	exec, err := i.itr.Next()
	return exec.(*Execution), err
}

// HasNext returns true if there is a next Execution false otherwise
func (i *blobstoreIterator) HasNext() bool {
	return i.itr.HasNext()
}

func getBlobstoreFetchPageFn(
	client blobstore.Client,
	keys Keys,
) pagination.FetchFn {
	return func(token pagination.PageToken) ([]pagination.Entity, pagination.PageToken, error) {
		index := token.(int)
		if index > keys.MaxPage {
			return nil, nil, nil
		}
		key := pageNumberToKey(keys.UUID, keys.Extension, index)
		// TODO: add retries here
		ctx, cancel := context.WithTimeout(context.Background(), BlobstoreTimeout)
		defer cancel()
		req := &blobstore.GetRequest{
			Key: key,
		}
		resp, err := client.Get(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		parts := bytes.Split(resp.Blob.Body, BlobstoreSeparatorToken)
		var executions []pagination.Entity
		for _, p := range parts {
			if len(p) == 0 {
				continue
			}
			// TODO: this iterator should be more generic then this, it should not assume this type
			var soe ScanOutputEntity
			if err := json.Unmarshal(p, &soe); err != nil {
				return nil, nil, err
			}
			executions = append(executions, &soe.Execution)
		}
		return executions, index + 1, nil
	}
}
