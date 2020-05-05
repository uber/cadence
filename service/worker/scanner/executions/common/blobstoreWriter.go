package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/pagination"
)

type (
	blobstoreWriter struct {
		writer    pagination.Writer
		uuid      string
		extension string
	}
)

// NewBlobstoreWriter constructs a new blobstore writer
func NewBlobstoreWriter(
	uuid string,
	extension string,
	client blobstore.Client,
	flushThreshold int,
) ExecutionWriter {
	return &blobstoreWriter{
		writer: pagination.NewWriter(
			getBlobstoreWriteFn(uuid, extension, client),
			getBlobstoreShouldFlushFn(flushThreshold),
			0),
	}
}

// Add adds an entity to blobstore writer
func (bw *blobstoreWriter) Add(e interface{}) error {
	return bw.writer.Add(e)
}

// Flush flushes contents of writer to blobstore
func (bw *blobstoreWriter) Flush() error {
	return bw.writer.Flush()
}

// FlushedKeys returns the keys that have been successfully flushed.
// Returns nil if no keys have been flushed.
func (bw *blobstoreWriter) FlushedKeys() *Keys {
	if len(bw.writer.FlushedPages()) == 0 {
		return nil
	}
	return &Keys{
		UUID:      bw.uuid,
		MinPage:   bw.writer.FirstFlushedPage().(int),
		MaxPage:   bw.writer.LastFlushedPage().(int),
		Extension: bw.extension,
	}
}

func getBlobstoreWriteFn(
	uuid string,
	extension string,
	client blobstore.Client,
) pagination.WriteFn {
	return func(page pagination.Page) (pagination.PageToken, error) {
		blobIndex := page.PageToken.(int)
		key := pageNumberToKey(uuid, extension, blobIndex)
		buffer := &bytes.Buffer{}
		for _, e := range page.Entities {
			data, err := json.Marshal(e)
			if err != nil {
				return nil, err
			}
			buffer.Write(data)
			buffer.Write(BlobstoreSeparatorToken)
		}
		req := &blobstore.PutRequest{
			Key: key,
			Blob: blobstore.Blob{
				Body: buffer.Bytes(),
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), BlobstoreTimeout)
		defer cancel()
		if _, err := client.Put(ctx, req); err != nil {
			return nil, err
		}
		return blobIndex + 1, nil
	}
}

func getBlobstoreShouldFlushFn(
	flushThreshold int,
) pagination.ShouldFlushFn {
	return func(page pagination.Page) bool {
		return len(page.Entities) > flushThreshold
	}
}

func pageNumberToKey(uuid string, extension string, pageNum int) string {
	return fmt.Sprintf("%v_%v.%v", uuid, pageNum, extension)
}
