// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package filestore

import (
	"context"
	"fmt"
	"github.com/uber/cadence/common/blobstore"
	"path/filepath"
)

const (
	metadataFilename = "metadata"
)

type client struct {
	storeDirectory string
}

// NewFileStoreClient returns a new Client backed by file system
func NewClient(cfg *Config) (blobstore.Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if err := setupDirectories(cfg); err != nil {
		return nil, err
	}
	if err := writeMetadataFiles(cfg); err != nil {
		return nil, err
	}
	return &client{
		storeDirectory: cfg.StoreDirectory,
	}, nil
}

// TODO: it makes sense to have a file_blob.go that knows how to serialize and deserialize blobs to and from files

func (c *client) UploadBlob(_ context.Context, bucket string, filename string, blob *blobstore.Blob) error {
	// check that bucket exists if not return error
	// convert blob into bytes this involves figuring out how to write tags and body into byte array
	// construct blob path from datastoreDir, bucket, filename
	// write blob (this potentially replaces an existing blob)

	return nil
}

func (c *client) DownloadBlob(_ context.Context, bucket string, filename string) (*blobstore.Blob, error) {
	// check that bucket exists if not return error
	// construct blob path from datastoreDir, bucket and fileanme
	// read the whole file into memory, construct blob
	// return result

	return nil, nil
}

func (c *client) BucketMetadata(_ context.Context, bucket string) (*blobstore.BucketMetadataResponse, error) {
	// check if bucket exists if not return error
	// check that metadata file exists if not return error
	// attempt to deserialize metadata file, if failed return error
	// otherwise return metadata

	return nil, nil
}

func setupDirectories(cfg *Config) error {
	var err error
	ensureExists := func(directoryPath string) {
		if err := mkdirAll(directoryPath); err != nil {
			err = fmt.Errorf("failed to ensure directory %v exists: %v", directoryPath, err)
		}
	}

	ensureExists(cfg.StoreDirectory)
	ensureExists(bucketDirectory(cfg.StoreDirectory, cfg.DefaultBucket.Name))
	for _, b := range cfg.CustomBuckets {
		ensureExists(bucketDirectory(cfg.StoreDirectory, b.Name))
	}
	return err
}

func writeMetadataFiles(cfg *Config) error {
	var err error
	writeMetadataFile := func(bucketConfig BucketConfig) {
		path := bucketItemPath(cfg.StoreDirectory, bucketConfig.Name, metadataFilename)
		bytes, err := serialize(&bucketConfig)
		if err != nil {
			err = fmt.Errorf("failed to write metadata file for bucket %v: %v", bucketConfig.Name, err)
		}
		if err := writeFile(path, bytes); err != nil {
			err = fmt.Errorf("failed to write metadata file for bucket %v: %v", bucketConfig.Name, err)
		}
	}

	writeMetadataFile(cfg.DefaultBucket)
	for _, b := range cfg.CustomBuckets {
		writeMetadataFile(b)
	}
	return err
}

func bucketDirectory(storeDirectory string, bucketName string) string {
	return filepath.Join(storeDirectory, bucketName)
}

func bucketItemPath(storeDirectory string, bucketName string, filename string) string {
	return filepath.Join(bucketDirectory(storeDirectory, bucketName), filename)
}
