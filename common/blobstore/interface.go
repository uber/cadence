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
	"errors"
	"strings"
)

const (
	// KeySeparator separates pieces of a key
	KeySeparator = "_"
)

var (
	// ErrBlobNotExists indicates that requested blob does not exist
	ErrBlobNotExists = errors.New("requested blob does not exist")
	// ErrBucketNotExists indicates that requested bucket does not exist
	ErrBucketNotExists = errors.New("requested bucket does not exist")
)

type (
	// Blob defines a blob
	Blob struct {
		Body []byte
		Tags map[string]string
	}

	// BucketMetadataResponse contains information relating to a bucket's configuration
	BucketMetadataResponse struct {
		Owner         string
		RetentionDays int
	}
)

// Client is used to operate on blobs in a blobstore
type Client interface {
	Upload(ctx context.Context, bucket string, key string, blob *Blob) error
	Download(ctx context.Context, bucket string, key string) (*Blob, error)
	Exists(ctx context.Context, bucket string, key string) (bool, error)
	Delete(ctx context.Context, bucket string, key string) (bool, error)
	ListByPrefix(ctx context.Context, bucket string, prefix string) ([]string, error)
	BucketMetadata(ctx context.Context, bucket string) (*BucketMetadataResponse, error)
}

// ConstructKey constructs a blob key based upon name pieces
func ConstructKey(pieces ...string) (string, error) {
	for _, p := range pieces {
		if strings.Contains(p, KeySeparator) {
			return "", errors.New("key pieces cannot contain underscore")
		}
	}
	return strings.Join(pieces, KeySeparator), nil
}
