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

package queue

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/asyncworkflow/queue/provider"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type noopConsumer struct{}

func (n *noopConsumer) Start() error {
	return nil
}

func (n *noopConsumer) Stop() {}

func (n *noopConsumer) Messages() <-chan messaging.Message {
	return nil
}

// mockProducerConstructor is a mock function for producer constructor
func mockProducerConstructor(cfg *config.YamlNode, params *provider.Params) (messaging.Producer, error) {
	// Mock implementation
	return nil, nil
}

// mockConsumerConstructor is a mock function for consumer constructor
func mockConsumerConstructor(cfg *config.YamlNode, params *provider.Params) (messaging.Consumer, error) {
	// Mock implementation
	return nil, nil
}

func TestNewAsyncQueueProvider(t *testing.T) {
	// Mock the provider registration functions
	provider.RegisterAsyncQueueProducerProvider("validType", mockProducerConstructor)
	provider.RegisterAsyncQueueConsumerProvider("validType", mockConsumerConstructor)

	tests := []struct {
		name          string
		cfg           map[string]config.AsyncWorkflowQueueProvider
		expectError   bool
		errorContains string
	}{
		{
			name: "Successful Initialization",
			cfg: map[string]config.AsyncWorkflowQueueProvider{
				"testQueue": {Type: "validType", Config: &config.YamlNode{}},
			},
			expectError: false,
		},
		{
			name: "Unregistered Queue Type",
			cfg: map[string]config.AsyncWorkflowQueueProvider{
				"testQueue": {Type: "invalidType", Config: &config.YamlNode{}},
			},
			expectError:   true,
			errorContains: "not registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAsyncQueueProvider(tt.cfg, &provider.Params{})
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAsyncQueueProducer(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		mockSetup   func(*cache.MockDomainCache, *cache.MockCache)
		producers   map[string]messaging.Producer
		expectError bool
	}{
		{
			name:   "Success case - predefined queue",
			domain: "test-domain",
			mockSetup: func(mockDomainCache *cache.MockDomainCache, mockCache *cache.MockCache) {
				mockDomainCache.EXPECT().GetDomain("test-domain").Return(cache.NewGlobalDomainCacheEntryForTest(
					nil,
					&persistence.DomainConfig{
						AsyncWorkflowConfig: types.AsyncWorkflowConfiguration{
							PredefinedQueueName: "testQueue",
						},
					},
					nil,
					0,
				), nil)
			},
			producers: map[string]messaging.Producer{
				"testQueue": messaging.NewNoopProducer(),
			},
			expectError: false,
		},
		{
			name:   "Success case - domain cache hit",
			domain: "test-domain",
			mockSetup: func(mockDomainCache *cache.MockDomainCache, mockCache *cache.MockCache) {
				mockDomainCache.EXPECT().GetDomain("test-domain").Return(cache.NewGlobalDomainCacheEntryForTest(
					nil,
					&persistence.DomainConfig{
						AsyncWorkflowConfig: types.AsyncWorkflowConfiguration{
							PredefinedQueueName: "",
						},
					},
					nil,
					0,
				), nil)
				mockCache.EXPECT().Get("test-domain").Return(messaging.NewNoopProducer())
			},
			expectError: false,
		},
		{
			name:   "Error case - failed to get domain entry from cache",
			domain: "test-domain",
			mockSetup: func(mockDomainCache *cache.MockDomainCache, mockCache *cache.MockCache) {
				mockDomainCache.EXPECT().GetDomain("test-domain").Return(nil, errors.New("failed to get domain entry from cache"))
			},
			producers: map[string]messaging.Producer{
				"testQueue": messaging.NewNoopProducer(),
			},
			expectError: true,
		},
		{
			name:   "Error case - predefined queue not found",
			domain: "test-domain",
			mockSetup: func(mockDomainCache *cache.MockDomainCache, mockCache *cache.MockCache) {
				mockDomainCache.EXPECT().GetDomain("test-domain").Return(cache.NewGlobalDomainCacheEntryForTest(
					nil,
					&persistence.DomainConfig{
						AsyncWorkflowConfig: types.AsyncWorkflowConfiguration{
							PredefinedQueueName: "not-found",
						},
					},
					nil,
					0,
				), nil)
			},
			producers: map[string]messaging.Producer{
				"testQueue": messaging.NewNoopProducer(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)

			mockDomainCache := cache.NewMockDomainCache(mockCtrl)
			mockCache := cache.NewMockCache(mockCtrl)
			tt.mockSetup(mockDomainCache, mockCache)

			provider := &providerImpl{
				domainCache:   mockDomainCache,
				producerCache: mockCache,
				producers:     tt.producers,
			}

			producer, err := provider.GetAsyncQueueProducer(tt.domain)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NotNil(t, producer)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAsyncQueueConsumer(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		mockSetup   func(*cache.MockDomainCache)
		consumers   map[string]messaging.Consumer
		expectError bool
	}{
		{
			name:   "Success case - predefined queue",
			domain: "test-domain",
			mockSetup: func(mockDomainCache *cache.MockDomainCache) {
				mockDomainCache.EXPECT().GetDomain("test-domain").Return(cache.NewGlobalDomainCacheEntryForTest(
					nil,
					&persistence.DomainConfig{
						AsyncWorkflowConfig: types.AsyncWorkflowConfiguration{
							PredefinedQueueName: "testQueue",
						},
					},
					nil,
					0,
				), nil)
			},
			consumers: map[string]messaging.Consumer{
				"testQueue": &noopConsumer{},
			},
			expectError: false,
		},
		{
			name:   "Error case - failed to get domain entry from cache",
			domain: "test-domain",
			mockSetup: func(mockDomainCache *cache.MockDomainCache) {
				mockDomainCache.EXPECT().GetDomain("test-domain").Return(nil, errors.New("failed to get domain entry from cache"))
			},
			consumers: map[string]messaging.Consumer{
				"testQueue": &noopConsumer{},
			},
			expectError: true,
		},
		{
			name:   "Error case - predefined queue not found",
			domain: "test-domain",
			mockSetup: func(mockDomainCache *cache.MockDomainCache) {
				mockDomainCache.EXPECT().GetDomain("test-domain").Return(cache.NewGlobalDomainCacheEntryForTest(
					nil,
					&persistence.DomainConfig{
						AsyncWorkflowConfig: types.AsyncWorkflowConfiguration{
							PredefinedQueueName: "not-found",
						},
					},
					nil,
					0,
				), nil)
			},
			consumers: map[string]messaging.Consumer{
				"testQueue": &noopConsumer{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)

			mockDomainCache := cache.NewMockDomainCache(mockCtrl)
			tt.mockSetup(mockDomainCache)

			provider := &providerImpl{
				domainCache: mockDomainCache,
				consumers:   tt.consumers,
			}

			producer, err := provider.GetAsyncQueueConsumer(tt.domain)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NotNil(t, producer)
				assert.NoError(t, err)
			}
		})
	}
}
