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
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/errors"
)

// IsActive return whether the domain is active, i.e. non global domain or global domain which active cluster is the current cluster
// If domain is not active, it also returns an error
func IsActive(domain *cache.DomainCacheEntry, cluster cluster.Metadata) (bool, error) {
	if !domain.IsGlobalDomain() {
		// domain is not a global domain, meaning domain is always "active" within each cluster
		return true, nil
	}

	domainName := domain.GetInfo().Name
	activeCluster := domain.GetReplicationConfig().ActiveClusterName
	currentCluster := cluster.GetCurrentClusterName()

	if domain.IsDomainPendingActive() {
		return false, errors.NewDomainPendingActiveError(domainName, currentCluster)
	}

	if currentCluster != activeCluster {
		return false, errors.NewDomainNotActiveError(domainName, currentCluster, activeCluster)
	}

	return true, nil
}

func GetActiveDomainByID(cache cache.DomainCache, cluster cluster.Metadata, id string) (*cache.DomainCacheEntry, error) {
	if err := common.ValidateDomainUUID(id); err != nil {
		return nil, err
	}

	domain, err := cache.GetDomainByID(id)
	if err != nil {
		return nil, err
	}

	if _, err = IsActive(domain, cluster); err != nil {
		// TODO: currently reapply events API will check if returned domainEntry is nil or not
		// when there's an error.
		return domain, err
	}

	return domain, nil
}
