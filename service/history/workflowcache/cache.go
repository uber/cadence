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

type WFCache interface {
	Allow(domainID string, workflowID string) bool
}

type wfCache struct {
	muc            sync.Mutex
	lru            cache.Cache
	rateLimiterRPS func() float64
}

type cacheKey struct {
	domainID   string
	workflowID string
}

type cacheValue struct {
	rateLimiter *quotas.DynamicRateLimiter
}

type Params struct {
	TTL      time.Duration
	MaxCount int
	MaxRPS   func() float64
}

func New(params Params) WFCache {
	return &wfCache{
		lru: cache.New(&cache.Options{
			TTL:      params.TTL,
			Pin:      true,
			MaxCount: params.MaxCount,
		}),
		rateLimiterRPS: params.MaxRPS,
	}
}

func (c *wfCache) Allow(domainID string, workflowID string) bool {
	key := cacheKey{
		domainID:   domainID,
		workflowID: workflowID,
	}

	c.muc.Lock()
	defer c.muc.Unlock()

	value, ok := c.lru.Get(key).(*cacheValue)
	if !ok {
		value = &cacheValue{
			rateLimiter: quotas.NewDynamicRateLimiter(c.rateLimiterRPS),
		}
		c.lru.PutIfNotExist(key, value)
	}

	return value.rateLimiter.Allow()
}
