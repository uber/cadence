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

package ratelimited

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/service/history/workflowcache"
)

func TestAllowWfID(t *testing.T) {
	tests := []struct {
		ratelimitEnabled     bool
		workflowIDCacheAllow bool
		expected             bool
	}{
		{
			ratelimitEnabled:     true,
			workflowIDCacheAllow: true,
			expected:             true,
		},
		{
			ratelimitEnabled:     true,
			workflowIDCacheAllow: false,
			expected:             false,
		},
		{
			ratelimitEnabled:     false,
			workflowIDCacheAllow: true,
			expected:             true,
		},
		{
			ratelimitEnabled:     false,
			workflowIDCacheAllow: false,
			expected:             true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ratelimitEnabled: %t, workflowIDCacheAllow: %t", tt.ratelimitEnabled, tt.workflowIDCacheAllow), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			workflowIDCacheMock := workflowcache.NewMockWFCache(ctrl)
			workflowIDCacheMock.EXPECT().AllowExternal(testDomainID, testWorkflowID).Return(tt.workflowIDCacheAllow).Times(1)

			domainCacheMock := cache.NewMockDomainCache(ctrl)
			domainCacheMock.EXPECT().GetDomainName(testDomainID).Return(testDomainID, nil).Times(1)

			h := &historyHandler{
				workflowIDCache:                workflowIDCacheMock,
				domainCache:                    domainCacheMock,
				logger:                         log.NewNoop(),
				ratelimitExternalPerWorkflowID: func(domain string) bool { return tt.ratelimitEnabled },
			}

			got := h.allowWfID(testDomainID, testWorkflowID)

			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestAllowWfID_DomainCacheError(t *testing.T) {
	ctrl := gomock.NewController(t)
	domainCacheMock := cache.NewMockDomainCache(ctrl)
	domainCacheMock.EXPECT().GetDomainName(testDomainID).Return("", fmt.Errorf("TEST ERROR")).Times(1)

	loggerMock := &log.MockLogger{}
	loggerMock.On("Error", "Error when getting domain name", mock.Anything).Return().Times(1)

	h := &historyHandler{
		domainCache: domainCacheMock,
		logger:      loggerMock,
	}

	got := h.allowWfID(testDomainID, testWorkflowID)

	// Fail open
	assert.True(t, got)
}
