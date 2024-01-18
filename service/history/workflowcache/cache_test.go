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

package workflowcache

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/quotas"
)

const (
	testDomainID    = "B59344B2-4166-462D-9CBD-22B25D2A7B1B"
	testWorkflowID  = "8ED9219B-36A2-4FD0-B9EA-6298A0F2ED1A"
	testWorkflowID2 = "F6E31C3D-3E54-4530-BDBE-68AEBA475473"
)

// TestWfCache_AllowSingleWorkflow tests that the cache will use the correct rate limiter for internal and external requests.
func TestWfCache_AllowSingleWorkflow(t *testing.T) {
	// The external rate limiter will allow the first request, but not the second.
	ctrl := gomock.NewController(t)
	externalLimiter := quotas.NewMockLimiter(ctrl)
	externalLimiter.EXPECT().Allow().Return(true).Times(1)
	externalLimiter.EXPECT().Allow().Return(false).Times(1)

	externalLimiterFactory := quotas.NewMockLimiterFactory(ctrl)
	externalLimiterFactory.EXPECT().GetLimiter(testDomainID).Return(externalLimiter).Times(1)

	// The internal rate limiter will allow the second request, but not the first.
	internalLimiter := quotas.NewMockLimiter(ctrl)
	internalLimiter.EXPECT().Allow().Return(false).Times(1)
	internalLimiter.EXPECT().Allow().Return(true).Times(1)

	internalLimiterFactory := quotas.NewMockLimiterFactory(ctrl)
	internalLimiterFactory.EXPECT().GetLimiter(testDomainID).Return(internalLimiter).Times(1)

	wfCache := New(Params{
		// The cache TTL is set to 1 minute, so all requests will hit the cache
		TTL:                    time.Minute,
		MaxCount:               1_000,
		ExternalLimiterFactory: externalLimiterFactory,
		InternalLimiterFactory: internalLimiterFactory,
	})

	assert.True(t, wfCache.AllowExternal(testDomainID, testWorkflowID))
	assert.False(t, wfCache.AllowExternal(testDomainID, testWorkflowID))

	assert.False(t, wfCache.AllowInternal(testDomainID, testWorkflowID))
	assert.True(t, wfCache.AllowInternal(testDomainID, testWorkflowID))
}

// TestWfCache_AllowMultipleWorkflow tests that the cache will use the correct rate limiter for different workflows.
func TestWfCache_AllowMultipleWorkflow(t *testing.T) {
	ctrl := gomock.NewController(t)

	// The external rate limiter for wf1 will allow the first request, but not the second.
	externalLimiterWf1 := quotas.NewMockLimiter(ctrl)
	externalLimiterWf1.EXPECT().Allow().Return(true).Times(1)
	externalLimiterWf1.EXPECT().Allow().Return(false).Times(1)

	// The external rate limiter for wf2 will allow the second request, but not the first.
	externalLimiterWf2 := quotas.NewMockLimiter(ctrl)
	externalLimiterWf2.EXPECT().Allow().Return(false).Times(1)
	externalLimiterWf2.EXPECT().Allow().Return(true).Times(1)

	externalLimiterFactory := quotas.NewMockLimiterFactory(ctrl)
	externalLimiterFactory.EXPECT().GetLimiter(testDomainID).Return(externalLimiterWf1).Times(1)
	externalLimiterFactory.EXPECT().GetLimiter(testDomainID).Return(externalLimiterWf2).Times(1)

	// We do not expect calls to the internal rate limiters, but they will still be created.
	internalLimiterWf1 := quotas.NewMockLimiter(ctrl)
	internalLimiterWf2 := quotas.NewMockLimiter(ctrl)

	internalLimiterFactory := quotas.NewMockLimiterFactory(ctrl)
	internalLimiterFactory.EXPECT().GetLimiter(testDomainID).Return(internalLimiterWf1).Times(1)
	internalLimiterFactory.EXPECT().GetLimiter(testDomainID).Return(internalLimiterWf2).Times(1)

	wfCache := New(Params{
		TTL:                    time.Minute,
		MaxCount:               1_000,
		ExternalLimiterFactory: externalLimiterFactory,
		InternalLimiterFactory: internalLimiterFactory,
	})

	assert.True(t, wfCache.AllowExternal(testDomainID, testWorkflowID))
	assert.False(t, wfCache.AllowExternal(testDomainID, testWorkflowID2))

	assert.False(t, wfCache.AllowExternal(testDomainID, testWorkflowID))
	assert.True(t, wfCache.AllowExternal(testDomainID, testWorkflowID2))
}
