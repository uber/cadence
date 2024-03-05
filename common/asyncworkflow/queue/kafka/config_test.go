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

package kafka

import "testing"

func TestQueueConfigID(t *testing.T) {
	tests := []struct {
		name     string
		config   queueConfig
		expected string
	}{
		{
			name: "single broker",
			config: queueConfig{
				Connection: connectionConfig{
					Brokers: []string{"broker1:9092"},
				},
				Topic: "topic1",
			},
			expected: "kafka::topic1/broker1:9092",
		},
		{
			name: "multiple brokers",
			config: queueConfig{
				Connection: connectionConfig{
					Brokers: []string{"broker1:9092", "broker2:9092"},
				},
				Topic: "topic2",
			},
			expected: "kafka::topic2/broker1:9092,broker2:9092",
		},
		{
			name: "no brokers",
			config: queueConfig{
				Connection: connectionConfig{
					Brokers: []string{},
				},
				Topic: "topic3",
			},
			expected: "kafka::topic3/",
		},
		{
			name: "empty topic",
			config: queueConfig{
				Connection: connectionConfig{
					Brokers: []string{"broker1:9092"},
				},
				Topic: "",
			},
			expected: "kafka::/broker1:9092",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ID()
			if got != tt.expected {
				t.Errorf("queueConfig.ID() = %v, want %v", got, tt.expected)
			}
		})
	}
}
