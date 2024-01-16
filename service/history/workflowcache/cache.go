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
	"sync"
	"time"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/quotas"
)

// WFCache is a per workflow cache used for workflow specific in memory data
type WFCache interface {
	AllowExternal(domainID string, workflowID string) bool
	AllowInternal(domainID string, workflowID string) bool
}

type wfCache struct {
	muc                    sync.Mutex
	lru                    cache.Cache
	externalLimiterFactory quotas.LimiterFactory
	internalLimiterFactory quotas.LimiterFactory
}

type cacheKey struct {
	domainID   string
	workflowID string
}

type cacheValue struct {
	externalRateLimiter quotas.Limiter
	internalRateLimiter quotas.Limiter
}

// Params is the parameters for a new WFCache
type Params struct {
	TTL                    time.Duration
	MaxCount               int
	ExternalLimiterFactory quotas.LimiterFactory
	InternalLimiterFactory quotas.LimiterFactory
}

// New creates a new WFCache
func New(params Params) WFCache {
	return &wfCache{
		lru: cache.New(&cache.Options{
			TTL:      params.TTL,
			Pin:      true,
			MaxCount: params.MaxCount,
		}),
		externalLimiterFactory: params.ExternalLimiterFactory,
		internalLimiterFactory: params.InternalLimiterFactory,
	}
}

// AllowExternal returns true if the rate limiter for this domain/workflow allows an external request
func (c *wfCache) AllowExternal(domainID string, workflowID string) bool {
	c.muc.Lock()
	defer c.muc.Unlock()

	value := c.getCacheItem(domainID, workflowID)
	return value.externalRateLimiter.Allow()
}

// AllowInternal returns true if the rate limiter for this domain/workflow allows an internal request
func (c *wfCache) AllowInternal(domainID string, workflowID string) bool {
	c.muc.Lock()
	defer c.muc.Unlock()

	value := c.getCacheItem(domainID, workflowID)
	return value.internalRateLimiter.Allow()
}

func (c *wfCache) getCacheItem(domainID string, workflowID string) *cacheValue {
	key := cacheKey{
		domainID:   domainID,
		workflowID: workflowID,
	}

	value, ok := c.lru.Get(key).(*cacheValue)
	if !ok {
		value = &cacheValue{
			externalRateLimiter: c.externalLimiterFactory.GetLimiter(domainID),
			internalRateLimiter: c.internalLimiterFactory.GetLimiter(domainID),
		}
		c.lru.PutIfNotExist(key, value)
	}

	return value
}
