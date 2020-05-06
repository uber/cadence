package blobstore

import (
	"context"

	"github.com/uber/cadence/common/backoff"
)

type (
	retryableClient struct {
		client Client
		policy backoff.RetryPolicy
	}
)

// NewRetryableClient constructs a blobstorre client which retries transient errors.
func NewRetryableClient(client Client, policy backoff.RetryPolicy) Client {
	return &retryableClient{
		client: client,
		policy: policy,
	}
}

func (c *retryableClient) Put(ctx context.Context, req *PutRequest) (*PutResponse, error) {
	var resp *PutResponse
	var err error
	op := func() error {
		resp, err = c.client.Put(ctx, req)
		return err
	}
	err = backoff.Retry(op, c.policy, c.client.IsRetryableError)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *retryableClient) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	var resp *GetResponse
	var err error
	op := func() error {
		resp, err = c.client.Get(ctx, req)
		return err
	}
	err = backoff.Retry(op, c.policy, c.client.IsRetryableError)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *retryableClient) Exists(ctx context.Context, req *ExistsRequest) (*ExistsResponse, error) {
	var resp *ExistsResponse
	var err error
	op := func() error {
		resp, err = c.client.Exists(ctx, req)
		return err
	}
	err = backoff.Retry(op, c.policy, c.client.IsRetryableError)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *retryableClient) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	var resp *DeleteResponse
	var err error
	op := func() error {
		resp, err = c.client.Delete(ctx, req)
		return err
	}
	err = backoff.Retry(op, c.policy, c.client.IsRetryableError)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *retryableClient) IsRetryableError(err error) bool {
	return c.client.IsRetryableError(err)
}
