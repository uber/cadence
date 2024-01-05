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

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

func TestWorkflowCheckforValidation(t *testing.T) {
	testCases := []struct {
		name         string
		domainID     string
		domainStatus int
		expectError  bool
	}{
		{
			name:         "ActiveDomain",
			domainID:     "activeDomainID",
			domainStatus: 0, // Assuming 0 represents an active domain
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := log.NewNoop()
			metricsClient := metrics.NewNoopMetricsClient()
			mockDomainCache := cache.NewMockDomainCache(ctrl)

			checker := &checkerImpl{
				logger:        logger,
				metricsClient: metricsClient,
				dc:            mockDomainCache,
			}

			// Create a real DomainCacheEntry with the necessary DomainInfo
			domainInfo := &persistence.DomainInfo{Status: tc.domainStatus}
			domainEntry := &cache.DomainCacheEntry{Info: domainInfo}

			// Mock GetDomainByID to return the manually created DomainCacheEntry
			mockDomainCache.EXPECT().GetDomainByID(tc.domainID).Return(domainEntry, nil).AnyTimes()

			err := checker.WorkflowCheckforValidation("testWorkflowID", tc.domainID, "testDomainName", "testRunID")
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
