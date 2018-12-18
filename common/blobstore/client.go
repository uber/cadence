package blobstore

import (
	"context"
	"errors"
	"io"
)

// CompressionType defines the type of compression used for a blob
type CompressionType int

const (
	// NoCompression indicates that blob is not compressed
	NoCompression CompressionType = iota
)

// Blob defines a blob
type Blob struct {
	Body            io.Reader
	CompressionType CompressionType
	Tags            map[string]string
}

// BucketMetadataResponse contains information relating to a bucket's configuration
type BucketMetadataResponse struct {
	Owner         string
	RetentionDays int
}

type Client interface {
	UploadBlob(ctx context.Context, bucket string, path string, blob *Blob) error
	DownloadBlob(ctx context.Context, bucket string, path string) (*Blob, error)
	BucketMetadata(ctx context.Context, bucket string) (*BucketMetadataResponse, error)
}

type nopClient struct{}

func (c *nopClient) UploadBlob(ctx context.Context, bucket string, path string, blob *Blob) error {
	return errors.New("not implemented")
}

func (c *nopClient) DownloadBlob(ctx context.Context, bucket string, path string) (*Blob, error) {
	return nil, errors.New("not implemented")
}

func (c *nopClient) BucketMetadata(ctx context.Context, bucket string) (*BucketMetadataResponse, error) {
	return nil, errors.New("not implemented")
}

// NewNopClient creates a nop client
func NewNopClient() Client {
	return &nopClient{}
}
