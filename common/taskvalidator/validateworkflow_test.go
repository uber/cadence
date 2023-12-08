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

package taskvalidator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

// MockMetricsScope implements the metrics.Scope interface for testing purposes.
type MockMetricsScope struct{}

func (s *MockMetricsScope) IncCounter(counter int) {}

func TestWorkflowCheckforValidation(t *testing.T) {
	// Create a mock logger and metrics client
	logger := log.NewNoop()
	metricsClient := metrics.NewNoopMetricsClient()

	// Create an instance of checkerImpl with the mock logger and metrics client
	checker := NewWfChecker(logger, metricsClient)

	// Define test inputs
	workflowID := "testWorkflowID"
	domainID := "testDomainID"
	runID := "testRunID"

	// Call the method being tested
	err := checker.WorkflowCheckforValidation(workflowID, domainID, runID)

	// Assert that the method returned no error
	assert.NoError(t, err)

	// Add additional assertions as needed based on the expected behavior of the method
}
