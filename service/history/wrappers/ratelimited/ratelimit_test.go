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
