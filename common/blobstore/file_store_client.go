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

package blobstore

import (
	"context"
	"fmt"
	"path/filepath"
)

const (
	metadataFilename = "metadata"
)

type fileStoreClient struct {
	storeDirectory string
}

// NewFileStoreClient returns a new blobstore Client backed by file system
func NewFileStoreClient(cfg *Config) Client {
	cfg.Validate()
	setupDirectories(cfg)
	writeMetadataFiles(cfg)
	return &fileStoreClient{
		storeDirectory: cfg.StoreDirectory,
	}
}

func (c *fileStoreClient) UploadBlob(ctx context.Context, bucket string, filename string, blob *Blob) error {
	return nil
}

func (c *fileStoreClient) DownloadBlob(ctx context.Context, bucket string, filename string) (*Blob, error) {
	return nil, nil
}

func (c *fileStoreClient) BucketMetadata(ctx context.Context, bucket string) (*BucketMetadataResponse, error) {
	return nil, nil
}

func setupDirectories(cfg *Config) {
	ensureExists := func(directoryPath string) {
		if err := createIfNotExists(directoryPath); err != nil {
			panic(fmt.Sprintf("failed to ensure directory %v exists: %v", directoryPath, err))
		}
	}

	ensureExists(cfg.StoreDirectory)
	ensureExists(bucketDirectory(cfg.StoreDirectory, cfg.DefaultBucket.Name))
	for _, b := range cfg.CustomBuckets {
		ensureExists(bucketDirectory(cfg.StoreDirectory, b.Name))
	}
}

func writeMetadataFiles(cfg *Config) {
	writeMetadataFile := func(bucketConfig BucketConfig) {
		path := bucketItemPath(cfg.StoreDirectory, bucketConfig.Name, metadataFilename)
		bytes, err := serialize(&bucketConfig)
		if err != nil {
			panic(fmt.Sprintf("failed to write metadata file for bucket %v: %v", bucketConfig.Name, err))
		}
		if err := writeFile(path, bytes); err != nil {
			panic(fmt.Sprintf("failed to write metadata file for bucket %v: %v", bucketConfig.Name, err))
		}
	}

	writeMetadataFile(cfg.DefaultBucket)
	for _, b := range cfg.CustomBuckets {
		writeMetadataFile(b)
	}
}

func bucketDirectory(storeDirectory string, bucketName string) string {
	return filepath.Join(storeDirectory, bucketName)
}

func bucketItemPath(storeDirectory string, bucketName string, filename string) string {
	return filepath.Join(bucketDirectory(storeDirectory, bucketName), filename)
}
