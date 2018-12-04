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
	"io"
	"strings"
	"time"
)

// CompressionType defines the type of compression used for a blob
type CompressionType int

const (
	// Separator is used to delimitate parts of a path
	Separator = "/"

	// NoCompression indicates that blob is not compressed
	NoCompression CompressionType = iota
)

type (
	// BucketName defines the name of a bucket
	BucketName string

	// PrefixPath defines a path to a common prefix under which multiple blobs are stored
	PrefixPath string

	// BlobPath defines a path to a blob
	BlobPath string
)

// BucketPolicy defines the policies that can be applied to a bucket
type BucketPolicy struct {
	Owner         string
	ReadACL       []string
	CreateACL     []string
	DeleteACL     []string
	AdminACL      []string
	RetentionDays int32
}

// PrefixMetadata defines metadata on a common prefix under which multiple blobs are stored
type PrefixMetadata struct {
	CreatedAt     time.Time
	Owner         string
	RetentionDays int32
	Listable      bool
}

// BlobMetadata defines metadata for a blob
type BlobMetadata struct {
	CreatedAt       time.Time
	Owner           string
	Size            int64
	Tags            map[string]string
	CompressionType CompressionType
}

// Blob defines a blob
type Blob struct {
	Size            int64 // optional
	Body            io.Reader
	CompressionType CompressionType
	Tags            map[string]string
}

// UploadBlobRequest is request for UploadBlob
type UploadBlobRequest struct {
	PrefixesListable bool
	Blob             *Blob
}

// ListByPrefixRequest is request for ListByPrefix
type ListByPrefixRequest struct {
	ListFrom PrefixPath
	Detailed bool
	Limit    int
}

// ListByPrefixResult is result of ListByPrefix
type ListByPrefixResult struct {
	Prefixes map[PrefixPath]*PrefixMetadata
	Blobs    map[BlobPath]*BlobMetadata
}

// Client is the interface to interact with archival store
type Client interface {
	// BucketExists returns true if bucket already exists, false otherwise
	BucketExists(context.Context, BucketName) (bool, error)
	// CreateBucket creates a new bucket with BucketPolicy
	CreateBucket(context.Context, BucketName, *BucketPolicy) error
	// GetBucketPolicy returns the BucketPolicy for bucket
	GetBucketPolicy(context.Context, BucketName) (*BucketPolicy, error)

	// ListByPrefix returns all blobs and common prefixes under PrefixPath
	ListByPrefix(context.Context, BucketName, PrefixPath, *ListByPrefixRequest) (*ListByPrefixResult, error)
	// CountByPrefix returns count of all blobs and common prefixes under PrefixPath
	CountUnderPrefix(context.Context, BucketName, PrefixPath) (int64, error)

	// UploadBlob uploads a blob to BucketName at location BlobPath
	UploadBlob(context.Context, BucketName, BlobPath, *UploadBlobRequest) error
	// DownloadBlob downloads a blob from BucketName at location BlobPath
	DownloadBlob(context.Context, BucketName, BlobPath) (*Blob, error)
	// GetBlobMetadata gets the metadata for blob from BucketName at location BlobPath
	GetBlobMetadata(context.Context, BucketName, BlobPath) (*BlobMetadata, error)
}

type nopClient struct{}

func (c *nopClient) BucketExists(context.Context, BucketName) (bool, error) {
	return false, errors.New("not implemented")
}

func (c *nopClient) CreateBucket(context.Context, BucketName, *BucketPolicy) error {
	return errors.New("not implemented")
}

func (c *nopClient) GetBucketPolicy(context.Context, BucketName) (*BucketPolicy, error) {
	return nil, errors.New("not implemented")
}

func (c *nopClient) UploadBlob(context.Context, BucketName, BlobPath, *UploadBlobRequest) error {
	return errors.New("not implemented")
}

func (c *nopClient) DownloadBlob(context.Context, BucketName, BlobPath) (*Blob, error) {
	return nil, errors.New("not implemented")
}

func (c *nopClient) GetBlobMetadata(context.Context, BucketName, BlobPath) (*BlobMetadata, error) {
	return nil, errors.New("not implemented")
}

func (c *nopClient) ListByPrefix(context.Context, BucketName, PrefixPath, *ListByPrefixRequest) (*ListByPrefixResult, error) {
	return nil, errors.New("not implemented")
}

func (c *nopClient) CountUnderPrefix(context.Context, BucketName, PrefixPath) (int64, error) {
	return 0, errors.New("not implemented")
}

// NewNopClient creates a nop client
func NewNopClient() Client {
	return &nopClient{}
}

// ConstructBlobPath constructs a blob path from given parts
func ConstructBlobPath(parts ...string) BlobPath {
	return BlobPath(strings.Join(parts, Separator))
}

// ConstructBucketName constructs a bucket name from given parts
func ConstructBucketName(parts ...string) BucketName {
	return BucketName(strings.Join(parts, Separator))
}
