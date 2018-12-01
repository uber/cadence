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
	"code.uber.internal/devexp/cadence-server/go-build/.go/src/gb2/src/github.com/pkg/errors"
	"context"
	"io"
	"time"
)

// CompressionType defines the type of compression used for a blob
type CompressionType int

const (

	// NoCompression indicates that blob was not compressed
	NoCompression CompressionType = iota
)

// BucketPolicy defines the policies that can be applied at bucket level
type BucketPolicy struct {
	Prefix        *string
	Owner         *string
	ReadACL       []string
	CreateACL     []string
	DeleteACL     []string
	AdminACL      []string
	RetentionDays *int32
}

// BlobMetadata defines metadata for a blob
type BlobMetadata struct {
	CreatedAt       *time.Time
	Owner           *string
	Size            *int64
	Tags            map[string]string
	CompressionType CompressionType
}

// Blob defines a blob
type Blob struct {
	Size            *int64 // optional
	Body            io.Reader
	CompressionType CompressionType
}

// ListItem is a listing of a single blob
type ListItem struct {
	IsBlob       *bool
	Name         *string
	BlobMetadata BlobMetadata
}

// UploadBlobRequest is request for UploadBlob
type UploadBlobRequest struct {
	Path  *string
	Index *bool
	Blob  *Blob
	Tags  map[string]string
}

// ListByPrefixRequest is request for ListByPrefix
type ListByPrefixRequest struct {
	Prefix   *string
	ListFrom *string
	Detailed *bool
	Limit    *int
}

// Client is the interface to interact with archival store
type Client interface {
	CreateBucket(ctx context.Context, request *BucketPolicy) error
	UpdateBucket(ctx context.Context, request *BucketPolicy) error
	GetBucketPolicy(ctx context.Context, prefix *string) (*BucketPolicy, error)

	UploadBlob(ctx context.Context, request *UploadBlobRequest) error
	DownloadBlob(ctx context.Context, path *string) (*Blob, error)
	GetBlobMetadata(ctx context.Context, path *string) (*BlobMetadata, error)
	ListByPrefix(ctx context.Context, request *ListByPrefixRequest) ([]*ListItem, error)
}

type nopClient struct{}

func (c *nopClient) CreateBucket(ctx context.Context, request *BucketPolicy) error {
	return errors.New("not implemented")
}

func (c *nopClient) UpdateBucket(ctx context.Context, request *BucketPolicy) error {
	return errors.New("not implemented")
}

func (c *nopClient) GetBucketPolicy(ctx context.Context, prefix *string) (*BucketPolicy, error) {
	return nil, errors.New("not implemented")
}

func (c *nopClient) UploadBlob(ctx context.Context, request *UploadBlobRequest) error {
	return errors.New("not implemented")
}

func (c *nopClient) DownloadBlob(ctx context.Context, path *string) (*Blob, error) {
	return nil, errors.New("not implemented")
}

func (c *nopClient) GetBlobMetadata(ctx context.Context, path *string) (*BlobMetadata, error) {
	return nil, errors.New("not implemented")
}

func (c *nopClient) ListByPrefix(ctx context.Context, request *ListByPrefixRequest) ([]*ListItem, error) {
	return nil, errors.New("not implemented")
}

// NewNopClient creates a nop client
func NewNopClient() Client {
	return &nopClient{}
}
