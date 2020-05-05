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
	keys []string,
) ExecutionIterator {
	return &blobstoreIterator{
		itr: pagination.NewIterator(0, getBlobstoreFetchPageFn(client, keys)),
	}
}

// Next returns the next Execution
func (i *blobstoreIterator) Next() (*Execution, error) {
	exec, err := i.itr.Next()
	return exec.(*Execution), err
}

// HasNext retursn true if there is a next Execution false otherwise
func (i *blobstoreIterator) HasNext() bool {
	return i.itr.HasNext()
}

func getBlobstoreFetchPageFn(
	client blobstore.Client,
	keys []string,
) pagination.FetchFn {
	return func(token pagination.PageToken) ([]pagination.Entity, pagination.PageToken, error) {
		index := token.(int)
		if index >= len(keys) {
			return nil, nil, nil
		}
		key := keys[index]
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
			var soe ScanOutputEntity
			if err := json.Unmarshal(p, &soe); err != nil {
				return nil, nil, err
			}
			executions = append(executions, &soe.Execution)
		}
		return executions, index + 1, nil
	}
}
