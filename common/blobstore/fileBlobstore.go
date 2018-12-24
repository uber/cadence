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
)

type (
	// Config describes the configuration needed to construct a blobstore client backed by file system
	Config struct {
		DataDirectory         string `yaml:"dataDirectory"`
		DynamicBucketCreation bool   `yaml:"dynamicBucketCreation"`
	}

	fileBlobstoreClient struct {
		dataDirectory         string
		defaultArchivalBucket string
		dynamicBucketCreation bool
	}
)

func NewFileBlobstoreClient(dataDirectory string, defaultArchivalBucket string, dynamicBucketCreation bool) (Client, error) {
	// basically we will make the default archival bucket as a directory under dataDir
	return &fileBlobstoreClient{
		dataDirectory:         dataDirectory,
		defaultArchivalBucket: defaultArchivalBucket,
		dynamicBucketCreation: dynamicBucketCreation,
	}, nil
}

func (c *fileBlobstoreClient) UploadBlob(ctx context.Context, bucket string, path string, blob *Blob) error {
	return nil
}

func (c *fileBlobstoreClient) DownloadBlob(ctx context.Context, bucket string, path string) (*Blob, error) {
	return nil, nil
}

func (c *fileBlobstoreClient) BucketMetadata(ctx context.Context, bucket string) (*BucketMetadataResponse, error) {
	return nil, nil
}

// what do the contents of these files look like?
// it would be great if the files were self describing and sorta readable
// we should convert all paths such that they cannot have slahses, the basic idea is all these files will be flat always
// first line could be a serialized map of string->string
// then you need a line break
// then you have the content of the file

// I should also write unit tests for this file

// on file based config there should be dataDir, createNonExistant dirs

// the files will go under /dataDir and just be a bunch of flat files under this

// wait no its reasonable that you have /dataDir/bucketName/... all the stuff in that bucket
// logically each directory is its own bucket and within that bucket is a list of flat files basically
// so when you start up you have to say if you want to create dirs
// if you do not want dirs created than if /dataDir/defaultBucket does not exist on constructor then its an error
// if you do want dirs created than /dataDir/defaultBucket will attempt to be created if not already exists in constructor
// if you do not want dirs created and a bucket is given that does not already exist return error
// otherwise dynamically create new buckets....
// whenever a bucket is created a file will be added to it which describes its metadata
// if a bucket is created by hand then uploaded to, its possible it will not have metadata file. In this case we can lazily write these metadata files...

// sounds super reasonable :)
