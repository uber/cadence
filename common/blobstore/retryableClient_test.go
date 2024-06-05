package blobstore

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uber/cadence/common/backoff"
)

func TestRetryableClient_Put(t *testing.T) {
	mockClient := new(MockClient)
	policy := backoff.NewExponentialRetryPolicy(0)
	client := NewRetryableClient(mockClient, policy)

	req := &PutRequest{}
	resp := &PutResponse{}
	mockClient.On("Put", mock.Anything, req).Return(resp, nil).Once()

	result, err := client.Put(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp, result)

	mockClient.AssertExpectations(t)
}

func TestRetryableClient_Get(t *testing.T) {
	mockClient := new(MockClient)
	policy := backoff.NewExponentialRetryPolicy(0)
	client := NewRetryableClient(mockClient, policy)

	req := &GetRequest{}
	resp := &GetResponse{}
	mockClient.On("Get", mock.Anything, req).Return(resp, nil).Once()

	result, err := client.Get(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp, result)

	mockClient.AssertExpectations(t)
}

func TestRetryableClient_Exists(t *testing.T) {
	mockClient := new(MockClient)
	policy := backoff.NewExponentialRetryPolicy(0)
	client := NewRetryableClient(mockClient, policy)

	req := &ExistsRequest{}
	resp := &ExistsResponse{}
	mockClient.On("Exists", mock.Anything, req).Return(resp, nil).Once()

	result, err := client.Exists(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp, result)

	mockClient.AssertExpectations(t)
}

func TestRetryableClient_Delete(t *testing.T) {
	mockClient := new(MockClient)
	policy := backoff.NewExponentialRetryPolicy(0)
	client := NewRetryableClient(mockClient, policy)

	req := &DeleteRequest{}
	resp := &DeleteResponse{}
	mockClient.On("Delete", mock.Anything, req).Return(resp, nil).Once()

	result, err := client.Delete(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp, result)

	mockClient.AssertExpectations(t)
}

func TestRetryableClient_RetryOnError(t *testing.T) {
	mockClient := new(MockClient)
	policy := backoff.NewExponentialRetryPolicy(1) // Adjusting the retry interval to ensure retry is attempted
	client := NewRetryableClient(mockClient, policy)

	req := &PutRequest{}
	resp := &PutResponse{}
	retryableError := errors.New("retryable error")

	mockClient.On("Put", mock.Anything, req).Return(nil, retryableError).Once()
	mockClient.On("IsRetryableError", retryableError).Return(true).Once()
	mockClient.On("Put", mock.Anything, req).Return(resp, nil).Once()

	result, err := client.Put(context.Background(), req)
	assert.NoError(t, err, "Expected no error on successful retry")
	assert.Equal(t, resp, result, "Expected the response to match")

	mockClient.AssertExpectations(t)
}

func TestRetryableClient_NotRetryOnError(t *testing.T) {
	mockClient := new(MockClient)
	policy := backoff.NewExponentialRetryPolicy(0)
	client := NewRetryableClient(mockClient, policy)

	req := &PutRequest{}
	nonRetryableError := errors.New("non-retryable error")

	mockClient.On("Put", mock.Anything, req).Return(nil, nonRetryableError).Once()
	mockClient.On("IsRetryableError", nonRetryableError).Return(false).Once()

	result, err := client.Put(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)

	mockClient.AssertExpectations(t)
}
