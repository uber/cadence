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

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/metrics"
)

func TestGetOrCreateDomain(t *testing.T) {
	wfCache := &wfCache{}

	domainID := "some domain name"
	otherDomainID := "some other domain name"

	domainMaxCount1 := wfCache.getOrCreateDomain(domainID, time.Unix(123, 0))
	domainMaxCount2 := wfCache.getOrCreateDomain(domainID, time.Unix(456, 0))
	otherDomainMaxCount := wfCache.getOrCreateDomain(otherDomainID, time.Unix(789, 0))

	// None of them are nil
	assert.NotNil(t, domainMaxCount1)
	assert.NotNil(t, domainMaxCount2)
	assert.NotNil(t, otherDomainMaxCount)

	// domainMaxCount1 and domainMaxCount2 are the same
	assert.Equal(t, domainMaxCount1, domainMaxCount2)

	// domainMaxCount1/domainMaxCount2 and otherDomainMaxCount are different
	assert.NotEqual(t, domainMaxCount1, otherDomainMaxCount)
	assert.NotEqual(t, domainMaxCount2, otherDomainMaxCount)
}

func TestUpdatePerDomainMaxWFRequestCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	metricsClientMock := metrics.NewMockClient(ctrl)
	metricsMockScope := metrics.NewMockScope(ctrl)

	now := time.Unix(123, 456)
	timeSource := clock.NewMockedTimeSourceAt(now)

	wfc := &wfCache{
		metricsClient: metricsClientMock,
		timeSource:    timeSource,
	}

	domainName := "some domain name"
	value := &cacheValue{}

	for i := 0; i < 5; i++ {
		wfc.updatePerDomainMaxWFRequestCount(domainName, value)
	}

	// Test that the max count is recorded
	domain := wfc.getOrCreateDomain(domainName, now)
	assert.Equal(t, 5, domain.maxCount)

	// Test that different values are independent
	value = &cacheValue{}
	for i := 0; i < 8; i++ {
		wfc.updatePerDomainMaxWFRequestCount(domainName, value)
	}
	domain = wfc.getOrCreateDomain(domainName, now)
	assert.Equal(t, 8, domain.maxCount)

	// Test that the max count is emitted and reset after a second
	metricsClientMock.EXPECT().Scope(metrics.HistoryClientWfIDCacheScope, metrics.DomainTag(domainName)).
		Return(metricsMockScope)
	metricsMockScope.EXPECT().
		UpdateGauge(metrics.WorkflowIDCacheRequestsExternalMaxRequestsPerSecondsGauge, 8.0)
	timeSource.Advance(1100 * time.Millisecond)
	wfc.updatePerDomainMaxWFRequestCount(domainName, value)

	domain = wfc.getOrCreateDomain(domainName, now)
	assert.Equal(t, 1, domain.maxCount)
}
