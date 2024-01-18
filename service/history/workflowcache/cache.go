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
	"errors"
	"time"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/quotas"
)

// WFCache is a per workflow cache used for workflow specific in memory data
type WFCache interface {
	AllowExternal(domainID string, workflowID string) bool
	AllowInternal(domainID string, workflowID string) bool
}

type wfCache struct {
	lru                    cache.Cache
	externalLimiterFactory quotas.LimiterFactory
	internalLimiterFactory quotas.LimiterFactory
	logger                 log.Logger
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
	Logger                 log.Logger
}

// New creates a new WFCache
func New(params Params) WFCache {
	return &wfCache{
		lru: cache.New(&cache.Options{
			TTL:      params.TTL,
			Pin:      false,
			MaxCount: params.MaxCount,
		}),
		externalLimiterFactory: params.ExternalLimiterFactory,
		internalLimiterFactory: params.InternalLimiterFactory,
		logger:                 params.Logger,
	}
}

// AllowExternal returns true if the rate limiter for this domain/workflow allows an external request
func (c *wfCache) AllowExternal(domainID string, workflowID string) bool {
	// Locking is not needed because both getCacheItem and the rate limiter are thread safe
	value, err := c.getCacheItem(domainID, workflowID)
	if err != nil {
		c.logError(domainID, workflowID, err)
		// If we can't get the cache item, we should allow the request through
		return true
	}
	return value.externalRateLimiter.Allow()
}

// AllowInternal returns true if the rate limiter for this domain/workflow allows an internal request
func (c *wfCache) AllowInternal(domainID string, workflowID string) bool {
	// Locking is not needed because both getCacheItem and the rate limiter are thread safe
	value, err := c.getCacheItem(domainID, workflowID)
	if err != nil {
		c.logError(domainID, workflowID, err)
		// If we can't get the cache item, we should allow the request through
		return true
	}
	return value.internalRateLimiter.Allow()
}

func (c *wfCache) getCacheItem(domainID string, workflowID string) (*cacheValue, error) {
	// The underlying lru cache is thread safe, so there is no need to lock
	key := cacheKey{
		domainID:   domainID,
		workflowID: workflowID,
	}

	value, ok := c.lru.Get(key).(*cacheValue)

	if ok {
		return value, nil
	}

	value = &cacheValue{
		externalRateLimiter: c.externalLimiterFactory.GetLimiter(domainID),
		internalRateLimiter: c.internalLimiterFactory.GetLimiter(domainID),
	}
	// PutIfNotExist is thread safe, and will either return the value that was already in the cache or the value we just created
	// another thread might have inserted a value between the Get and PutIfNotExist, but that is ok
	// it should never return an error as we do not use Pin
	valueInterface, err := c.lru.PutIfNotExist(key, value)
	value, ok = valueInterface.(*cacheValue)

	// This should never happen, either the value was already in the cache or we just inserted it
	if !ok {
		return nil, errors.New("Failed to insert new value into cache")
	}

	return value, err
}

func (c *wfCache) logError(domainID string, workflowID string, err error) {
	c.logger.Error("Unexpected error from workflow cache",
		tag.Error(err),
		tag.WorkflowDomainID(domainID),
		tag.WorkflowID(workflowID),
		tag.WorkflowIDCacheSize(c.lru.Size()),
	)
}
