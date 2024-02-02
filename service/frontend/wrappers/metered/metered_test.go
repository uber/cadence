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

package metered

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/frontend/api"
	"github.com/uber/cadence/service/frontend/config"
)

func TestContextMetricsTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockHandler := api.NewMockHandler(ctrl)
	mockHandler.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.CountWorkflowExecutionsResponse{}, nil).Times(1)
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	testScope := tally.NewTestScope("test", nil)
	metricsClient := metrics.NewClient(testScope, metrics.Frontend)
	handler := NewAPIHandler(mockHandler, testlogger.New(t), metricsClient, mockDomainCache, nil)

	tag := metrics.TransportTag("grpc")
	ctx := metrics.TagContext(context.Background(), tag)
	handler.CountWorkflowExecutions(ctx, nil)

	snapshot := testScope.Snapshot()
	for _, counter := range snapshot.Counters() {
		if counter.Name() == "test.cadence_requests" {
			assert.Equal(t, tag.Value(), counter.Tags()[tag.Key()])
			return
		}
	}
	assert.Fail(t, "counter not found")
}

func TestSignalMetricHasSignalName(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockHandler := api.NewMockHandler(ctrl)
	mockHandler.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	testScope := tally.NewTestScope("test", nil)
	metricsClient := metrics.NewClient(testScope, metrics.Frontend)
	handler := NewAPIHandler(mockHandler, testlogger.New(t), metricsClient, mockDomainCache, &config.Config{EmitSignalNameMetricsTag: dynamicconfig.GetBoolPropertyFnFilteredByDomain(true)})

	signalRequest := &types.SignalWorkflowExecutionRequest{
		SignalName: "test_signal",
	}
	handler.SignalWorkflowExecution(context.Background(), signalRequest)

	expectedMetrics := make(map[string]bool)
	expectedMetrics["test.cadence_requests"] = false

	snapshot := testScope.Snapshot()
	for _, counter := range snapshot.Counters() {
		if _, ok := expectedMetrics[counter.Name()]; ok {
			expectedMetrics[counter.Name()] = true
		}
		if val, ok := counter.Tags()["signalName"]; ok {
			assert.Equal(t, val, "test_signal")
		} else {
			assert.Fail(t, "Couldn't find signalName tag")
		}
	}
	assert.True(t, expectedMetrics["test.cadence_requests"])
}
