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

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/uber/cadence/common/asyncworkflow/queue/provider"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
)

func TestConsumerConstructor(t *testing.T) {
	testCases := []struct {
		name       string
		configYaml string
		wantErr    bool
	}{
		{
			name: "Success case",
			configYaml: `
connection:
  brokers:
    - localhost:9092
  tls:
    enabled: false
  sasl:
    enabled: false
  topic: test-topic
`,
			wantErr: false,
		},
		{
			name: "Invalid yaml",
			configYaml: `
connection:
`,
			wantErr: true,
		},
		{
			name: "Invalid auth config",
			configYaml: `
connection:
  brokers:
    - localhost:9092
  tls:
    enabled: false
  sasl:
    enabled: true
    algorithm: test
  topic: test-topic
`,
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var out *config.YamlNode
			if !tt.wantErr {
				mockBroker := sarama.NewMockBroker(t, 0)
				defer mockBroker.Close()
				mockBroker.SetHandlerByMap(map[string]sarama.MockResponse{
					"MetadataRequest": sarama.NewMockMetadataResponse(t).
						SetBroker(mockBroker.Addr(), mockBroker.BrokerID()).
						SetLeader("test-topic", 0, mockBroker.BrokerID()).
						SetController(mockBroker.BrokerID()),
				})
				var queueCfg *QueueConfig
				err := yaml.Unmarshal([]byte(tt.configYaml), &queueCfg)
				require.NoError(t, err)
				queueCfg.Connection.Brokers = []string{mockBroker.Addr()}
				configYaml, err := yaml.Marshal(queueCfg)
				require.NoError(t, err)
				err = yaml.Unmarshal([]byte(configYaml), &out)
				require.NoError(t, err)
			} else {
				err := yaml.Unmarshal([]byte(tt.configYaml), &out)
				require.NoError(t, err)
			}

			_, err := ConsumerConstructor(out, &provider.Params{
				Logger:        testlogger.New(t),
				MetricsClient: metrics.NewNoopMetricsClient(),
			})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
