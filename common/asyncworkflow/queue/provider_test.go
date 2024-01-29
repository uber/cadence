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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/asyncworkflow/queue/provider"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/messaging"
)

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
