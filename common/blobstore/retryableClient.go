package blobstore

import (
	"context"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/blobstore/blob"
)

var _ Client = (*retryableClient)(nil)

type retryableClient struct {
	client      Client
	policy      backoff.RetryPolicy
	isRetryable backoff.IsRetryable
}

// NewRetryableClient creates a new instance of Client with retry policy
func NewRetryableClient(client Client, policy backoff.RetryPolicy, isRetryable backoff.IsRetryable) Client {
	return &retryableClient{
		client:      client,
		policy:      policy,
		isRetryable: isRetryable,
	}
}

func (c *retryableClient) Upload(ctx context.Context, bucket string, key blob.Key, blob *blob.Blob) error {
	op := func() error {
		return c.client.Upload(ctx, bucket, key, blob)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) Download(ctx context.Context, bucket string, key blob.Key) (*blob.Blob, error) {
	var resp *blob.Blob
	op := func() error {
		var err error
		resp, err = c.client.Download(ctx, bucket, key)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) Exists(ctx context.Context, bucket string, key blob.Key) (bool, error) {
	var resp bool
	op := func() error {
		var err error
		resp, err = c.client.Exists(ctx, bucket, key)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) Delete(ctx context.Context, bucket string, key blob.Key) (bool, error) {
	var resp bool
	op := func() error {
		var err error
		resp, err = c.client.Delete(ctx, bucket, key)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ListByPrefix(ctx context.Context, bucket string, prefix string) ([]blob.Key, error) {
	var resp []blob.Key
	op := func() error {
		var err error
		resp, err = c.client.ListByPrefix(ctx, bucket, prefix)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) BucketMetadata(ctx context.Context, bucket string) (*BucketMetadataResponse, error) {
	var resp *BucketMetadataResponse
	op := func() error {
		var err error
		resp, err = c.client.BucketMetadata(ctx, bucket)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}