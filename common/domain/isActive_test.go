// Copyright (c) 2022 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package domain

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

func Test_IsActive(t *testing.T) {
	tests := []struct {
		msg              string
		isGlobalDomain   bool
		currentCluster   string
		activeCluster    string
		failoverDeadline *int64
		expectIsActive   bool
		expectedErr      error
	}{
		{
			msg:            "local domain",
			isGlobalDomain: false,
			expectIsActive: true,
		},
		{
			msg:              "global pending active domain",
			isGlobalDomain:   true,
			failoverDeadline: common.Int64Ptr(time.Now().Unix()),
			expectedErr:      &types.DomainNotActiveError{Message: "Domain: test-domain is pending active in cluster: .", DomainName: "test-domain", CurrentCluster: "", ActiveCluster: ""},
		},
		{
			msg:            "global domain on active cluster",
			isGlobalDomain: true,
			currentCluster: "A",
			activeCluster:  "A",
			expectIsActive: true,
		},
		{
			msg:            "global domain on passive cluster",
			isGlobalDomain: true,
			currentCluster: "A",
			activeCluster:  "B",
			expectedErr:    &types.DomainNotActiveError{Message: "Domain: test-domain is active in cluster: B, while current cluster A is a standby cluster.", DomainName: "test-domain", CurrentCluster: "A", ActiveCluster: "B"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			cluster := cluster.NewMetadata(0, "", tt.currentCluster, nil)
			domain := cache.NewDomainCacheEntryForTest(
				&persistence.DomainInfo{Name: "test-domain"},
				nil,
				tt.isGlobalDomain,
				&persistence.DomainReplicationConfig{ActiveClusterName: tt.activeCluster},
				0,
				tt.failoverDeadline,
			)

			isActive, err := IsActive(domain, cluster)

			assert.Equal(t, tt.expectIsActive, isActive)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func Test_GetActiveDomainByID(t *testing.T) {
	nonExistingUUID := uuid.New()
	activeDomainUUID := uuid.New()
	passiveDomainUUID := uuid.New()

	activeDomain := cache.NewGlobalDomainCacheEntryForTest(&persistence.DomainInfo{ID: activeDomainUUID, Name: "active"}, nil, &persistence.DomainReplicationConfig{ActiveClusterName: "A"}, 0)
	passiveDomain := cache.NewGlobalDomainCacheEntryForTest(&persistence.DomainInfo{ID: passiveDomainUUID, Name: "passive"}, nil, &persistence.DomainReplicationConfig{ActiveClusterName: "B"}, 0)

	tests := []struct {
		msg          string
		domainID     string
		expectDomain *cache.DomainCacheEntry
		expectedErr  error
	}{
		{
			msg:         "invalid UUID",
			domainID:    "invalid",
			expectedErr: &types.BadRequestError{Message: "Invalid domain UUID."},
		},
		{
			msg:         "non existing domain",
			domainID:    nonExistingUUID,
			expectedErr: assert.AnError,
		},
		{
			msg:          "active domain",
			domainID:     activeDomainUUID,
			expectDomain: activeDomain,
		},
		{
			msg:          "passive domain",
			domainID:     passiveDomainUUID,
			expectDomain: passiveDomain,
			expectedErr:  &types.DomainNotActiveError{Message: "Domain: passive is active in cluster: B, while current cluster A is a standby cluster.", DomainName: "passive", CurrentCluster: "A", ActiveCluster: "B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cache := cache.NewMockDomainCache(ctrl)
			cache.EXPECT().GetDomainByID(nonExistingUUID).Return(nil, assert.AnError).AnyTimes()
			cache.EXPECT().GetDomainByID(activeDomainUUID).Return(activeDomain, nil).AnyTimes()
			cache.EXPECT().GetDomainByID(passiveDomainUUID).Return(passiveDomain, nil).AnyTimes()

			cluster := cluster.NewMetadata(0, "", "A", nil)

			domain, err := GetActiveDomainByID(cache, cluster, tt.domainID)

			assert.Equal(t, tt.expectDomain, domain)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
