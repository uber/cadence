// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

func TestRetryableClient_IsRetryableError(t *testing.T) {
	mockClient := new(MockClient)
	client := &retryableClient{
		client: mockClient,
		throttleRetry: backoff.NewThrottleRetry(
			backoff.WithRetryPolicy(backoff.NewExponentialRetryPolicy(0)),
			backoff.WithRetryableError(mockClient.IsRetryableError),
		),
	}

	retryableError := errors.New("retryable error")

	mockClient.On("IsRetryableError", retryableError).Return(true).Once()

	isRetryable := client.IsRetryableError(retryableError)
	assert.True(t, isRetryable, "Expected error to be retryable")

	mockClient.AssertExpectations(t)
}
