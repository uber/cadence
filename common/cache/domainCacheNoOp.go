// Copyright (c) 2017 Uber Technologies, Inc.
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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination domainCache_mock.go -self_package github.com/uber/cadence/common/cache

package cache

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

type noOpDomainCache struct {
	status        int32
	shutdownChan  chan struct{}
	cacheNameToID *atomic.Value
	cacheByID     *atomic.Value
	domainManager persistence.DomainManager
	timeSource    clock.TimeSource
	scope         metrics.Scope
	logger        log.Logger

	// refresh lock is used to guarantee at most one
	// coroutine is doing domain refreshment
	refreshLock     sync.Mutex
	lastRefreshTime time.Time
	// This is debug field to emit callback count
	lastCallbackEmitTime time.Time

	callbackLock     sync.Mutex
	prepareCallbacks map[int]PrepareCallbackFn
	callbacks        map[int]CallbackFn
}

func (c *noOpDomainCache) GetCacheSize() (sizeOfCacheByName int64, sizeOfCacheByID int64) {
	return 0, 0
}

func (c *noOpDomainCache) GetAllDomain() map[string]*DomainCacheEntry {
	return nil
}

func (c *noOpDomainCache) RegisterDomainChangeCallback(
	shard int,
	initialNotificationVersion int64,
	prepareCallback PrepareCallbackFn,
	callback CallbackFn,
) {
	return
}

func (c *noOpDomainCache) UnregisterDomainChangeCallback(
	shard int,
) {
	return
}

func (c *noOpDomainCache) GetDomain(
	name string,
) (*DomainCacheEntry, error) {
	return nil, errors.New("Not supported for noOpDomainCache")
}

func (c *noOpDomainCache) GetDomainByID(
	id string,
) (*DomainCacheEntry, error) {

	return nil, errors.New("Not supported for noOpDomainCache")
}

func (c *noOpDomainCache) GetDomainID(
	name string,
) (string, error) {
	return "", errors.New("Not supported for noOpDomainCache")
}

func (c *noOpDomainCache) GetDomainName(
	id string,
) (string, error) {
	return "", errors.New("Not supported for noOpDomainCache")
}

func (c *noOpDomainCache) getDomain(
	name string,
) (*DomainCacheEntry, error) {
	return nil, errors.New("Not supported for noOpDomainCache")
}

func (c *noOpDomainCache) getDomainByID(
	id string,
	deepCopy bool,
) (*DomainCacheEntry, error) {
	return nil, errors.New("Not supported for noOpDomainCache")
}

func newNoOpDomainCache() Cache {
	return &noOpCache{}
}

// NewDomainCache creates a new instance of cache for holding onto domain information to reduce the load on persistence
func NewNoOpDomainCache() DomainCache {

	cache := &noOpDomainCache{}
	cache.cacheNameToID.Store(newNoOpDomainCache())
	cache.cacheByID.Store(newNoOpDomainCache())

	return cache
}

func (c *noOpDomainCache) Start() {
	if !atomic.CompareAndSwapInt32(&c.status, domainCacheInitialized, domainCacheStarted) {
		return
	}

	// initialize the cache by initial scan
	err := c.refreshDomains()
	if err != nil {
		c.logger.Fatal("Unable to initialize domain cache", tag.Error(err))
	}
	go c.refreshLoop()
}

// Stop stops background refresh of domain
func (c *noOpDomainCache) Stop() {
	if !atomic.CompareAndSwapInt32(&c.status, domainCacheStarted, domainCacheStopped) {
		return
	}
	close(c.shutdownChan)
}

func (c *noOpDomainCache) refreshLoop() {
	timer := time.NewTicker(DomainCacheRefreshInterval)
	defer timer.Stop()

	for {
		select {
		case <-c.shutdownChan:
			return
		case <-timer.C:
			for err := c.refreshDomains(); err != nil; err = c.refreshDomains() {
				select {
				case <-c.shutdownChan:
					return
				default:
					c.logger.Error("Error refreshing domain cache", tag.Error(err))
					time.Sleep(DomainCacheRefreshFailureRetryInterval)
				}
			}
		}
	}
}

func (c *noOpDomainCache) refreshDomains() error {
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	return c.refreshDomainsLocked()
}

// this function only refresh the domains in the v2 table
// the domains in the v1 table will be refreshed if cache is stale
func (c *noOpDomainCache) refreshDomainsLocked() error {
	now := c.timeSource.Now()
	if now.Sub(c.lastRefreshTime) < domainCacheMinRefreshInterval {
		return nil
	}
	return errors.New("Not supported for noOpDomainCache")
}

func (c *noOpDomainCache) checkDomainExists(
	name string,
	id string,
) error {
	return errors.New("Not supported for noOpDomainCache")
}

func (c *noOpDomainCache) updateNameToIDCache(
	cacheNameToID Cache,
	name string,
	id string,
) {
	return
}

func (c *noOpDomainCache) updateIDToDomainCache(
	cacheByID Cache,
	id string,
	record *DomainCacheEntry,
) (bool, *DomainCacheEntry, error) {

	return false, nil, errors.New("Not supported for noOpDomainCache")
}
