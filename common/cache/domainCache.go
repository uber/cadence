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

package cache

import (
	"sync"
	"sync/atomic"
	"time"

	workflow "github.com/uber/cadence/.gen/go/shared"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/errors"
	"github.com/uber/cadence/common/persistence"

	"github.com/uber-common/bark"
)

const (
	domainCacheInitialSize     = 10 * 1024
	domainCacheMaxSize         = 16 * 1024
	domainCacheTTL             = time.Hour
	domainEntryRefreshInterval = 10 * time.Second

	domainCacheLocked   int32 = 0
	domainCacheReleased int32 = 1

	domainCacheInitialized int32 = 0
	domainCacheStarted     int32 = 1
	domainCacheStopped     int32 = 2
)

type (
	// function to be called when the domain cache is changed
	// the callback function will be called within the domain cache entry lock
	// make sure the callback function will not call domain cache again
	// in case of deadlock
	callbackFn func(prevDomain *DomainCacheEntry, nextDomain *DomainCacheEntry)

	// DomainCache is used the cache domain information and configuration to avoid making too many calls to cassandra.
	// This cache is mainly used by frontend for resolving domain names to domain uuids which are used throughout the
	// system.  Each domain entry is kept in the cache for one hour but also has an expiry of 10 seconds.  This results
	// in updating the domain entry every 10 seconds but in the case of a cassandra failure we can still keep on serving
	// requests using the stale entry from cache upto an hour
	DomainCache interface {
		common.Daemon
		RegisterDomainChangeCallback(shard int, fn callbackFn)
		UnregisterDomainChangeCallback(shard int)
		GetDomain(name string) (*DomainCacheEntry, error)
		GetDomainByID(id string) (*DomainCacheEntry, error)
		GetDomainID(name string) (string, error)
	}

	domainCache struct {
		status          int32
		shutdownChan    chan struct{}
		cacheNameToID   Cache
		cacheByID       Cache
		metadataMgr     persistence.MetadataManager
		clusterMetadata cluster.Metadata
		timeSource      common.TimeSource
		logger          bark.Logger

		sync.RWMutex
		callbacks map[int]callbackFn
	}

	// DomainCacheEntry contains the info and config for a domain
	DomainCacheEntry struct {
		clusterMetadata cluster.Metadata

		sync.RWMutex
		info              *persistence.DomainInfo
		config            *persistence.DomainConfig
		replicationConfig *persistence.DomainReplicationConfig
		configVersion     int64
		failoverVersion   int64
		isGlobalDomain    bool
		expiry            time.Time
	}
)

// NewDomainCache creates a new instance of cache for holding onto domain information to reduce the load on persistence
func NewDomainCache(metadataMgr persistence.MetadataManager, clusterMetadata cluster.Metadata, logger bark.Logger) DomainCache {
	opts := &Options{}
	opts.InitialCapacity = domainCacheInitialSize
	opts.TTL = domainCacheTTL

	return &domainCache{
		status:          domainCacheInitialized,
		shutdownChan:    make(chan struct{}),
		cacheNameToID:   New(domainCacheMaxSize, opts),
		cacheByID:       New(domainCacheMaxSize, opts),
		metadataMgr:     metadataMgr,
		clusterMetadata: clusterMetadata,
		timeSource:      common.NewRealTimeSource(),
		logger:          logger,
		callbacks:       make(map[int]callbackFn),
	}
}

func newDomainCacheEntry(clusterMetadata cluster.Metadata) *DomainCacheEntry {
	return &DomainCacheEntry{clusterMetadata: clusterMetadata}
}

// Start start the background refresh of domain
func (c *domainCache) Start() {
	if !atomic.CompareAndSwapInt32(&c.status, domainCacheInitialized, domainCacheStarted) {
		return
	}
	go c.refreshLoop()
}

// Start start the background refresh of domain
func (c *domainCache) Stop() {
	if !atomic.CompareAndSwapInt32(&c.status, domainCacheStarted, domainCacheStopped) {
		return
	}
	close(c.shutdownChan)
}

func (c *domainCache) refreshLoop() {
	timer := time.NewTimer(domainEntryRefreshInterval)
	defer timer.Stop()
	for {
		select {
		case <-c.shutdownChan:
			return
		case <-timer.C:
			timer.Reset(domainEntryRefreshInterval)
			err := c.refreshDomains()
			if err != nil {
				c.logger.Errorf("Error refreshing domain cache: %v", err)
			}
		}
	}
}

func (c *domainCache) getIDs() []string {
	ite := c.cacheByID.Iterator()
	ids := []string{}
	defer ite.Close()
	for ite.HasNext() {
		id := ite.Next().Key().(string)
		ids = append(ids, id)
	}
	return ids
}

func (c *domainCache) refreshDomains() error {
	ids := c.getIDs()
	for _, id := range ids {
		_, err := c.GetDomainByID(id)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterDomainChangeCallback set a domain change callback
// WARN: the callback function will be triggered by domain cache when holding the domain cache lock,
// make sure the callback function will not call domain cache again in case of dead lock
func (c *domainCache) RegisterDomainChangeCallback(shard int, fn callbackFn) {
	c.Lock()
	defer c.Unlock()

	c.callbacks[shard] = fn
}

// UnregisterDomainChangeCallback delete a domain failover callback
func (c *domainCache) UnregisterDomainChangeCallback(shard int) {
	c.Lock()
	defer c.Unlock()

	delete(c.callbacks, shard)
}

// GetDomain retrieves the information from the cache if it exists, otherwise retrieves the information from metadata
// store and writes it to the cache with an expiry before returning back
func (c *domainCache) GetDomain(name string) (*DomainCacheEntry, error) {
	if name == "" {
		return nil, &workflow.BadRequestError{Message: "Domain is empty."}
	}
	return c.getDomain(name)
}

// GetDomainByID retrieves the information from the cache if it exists, otherwise retrieves the information from metadata
// store and writes it to the cache with an expiry before returning back
func (c *domainCache) GetDomainByID(id string) (*DomainCacheEntry, error) {
	if id == "" {
		return nil, &workflow.BadRequestError{Message: "DomainID is empty."}
	}
	return c.getDomainByID(id)
}

// GetDomainID retrieves domainID by using GetDomain
func (c *domainCache) GetDomainID(name string) (string, error) {
	entry, err := c.GetDomain(name)
	if err != nil {
		return "", err
	}
	return entry.info.ID, nil
}

func (c *domainCache) loadDomain(name string, id string) (*persistence.GetDomainResponse, error) {
	return c.metadataMgr.GetDomain(&persistence.GetDomainRequest{Name: name, ID: id})
}

func (c *domainCache) updateNameToIDCache(name string, id string) {
	c.cacheNameToID.Put(name, id)
}

func (c *domainCache) updateIDToDomainCache(id string, record *persistence.GetDomainResponse) (*DomainCacheEntry, error) {
	elem, err := c.cacheByID.PutIfNotExist(id, newDomainCacheEntry(c.clusterMetadata))
	if err != nil {
		return nil, err
	}
	entry := elem.(*DomainCacheEntry)
	entry.Lock()
	defer entry.Unlock()

	var prevDomain *DomainCacheEntry
	triggerCallback := c.clusterMetadata.IsGlobalDomainEnabled() &&
		// expiry will be non zero when the entry is initialized / valid
		!entry.expiry.IsZero() &&
		(record.ConfigVersion > entry.configVersion || record.FailoverVersion > entry.failoverVersion)
	// expiry will be non zero when the entry is initialized / valid
	if triggerCallback {
		prevDomain = entry.duplicate()
	}

	entry.info = record.Info
	entry.config = record.Config
	entry.replicationConfig = record.ReplicationConfig
	entry.configVersion = record.ConfigVersion
	entry.failoverVersion = record.FailoverVersion
	entry.isGlobalDomain = record.IsGlobalDomain
	entry.expiry = c.timeSource.Now().Add(domainEntryRefreshInterval)

	nextDomain := entry.duplicate()
	if triggerCallback {
		c.triggerDomainChangeCallback(prevDomain, nextDomain)
	}

	return nextDomain, nil
}

// getDomain retrieves the information from the cache if it exists, otherwise retrieves the information from metadata
// store and writes it to the cache with an expiry before returning back
func (c *domainCache) getDomain(name string) (*DomainCacheEntry, error) {
	id, cacheHit := c.cacheNameToID.Get(name).(string)
	if cacheHit {
		return c.getDomainByID(id)
	}

	record, err := c.loadDomain(name, "")
	if err != nil {
		return nil, err
	}
	id = record.Info.ID
	newEntry, err := c.updateIDToDomainCache(id, record)
	if err != nil {
		return nil, err
	}
	c.updateNameToIDCache(name, id)
	return newEntry, nil
}

// getDomainByID retrieves the information from the cache if it exists, otherwise retrieves the information from metadata
// store and writes it to the cache with an expiry before returning back
func (c *domainCache) getDomainByID(id string) (*DomainCacheEntry, error) {
	now := c.timeSource.Now()
	var result *DomainCacheEntry

	entry, cacheHit := c.cacheByID.Get(id).(*DomainCacheEntry)
	if cacheHit {
		// Found the information in the cache, lets check if it needs to be refreshed before returning back
		entry.RLock()
		if !entry.isExpired(now) {
			result = entry.duplicate()
			entry.RUnlock()
			return result, nil
		}
		// cache expired, need to refresh
		entry.RUnlock()
	}

	record, err := c.loadDomain("", id)
	if err != nil {
		// err updating, use the existing record if record is valid
		// i.e. expiry is set
		if cacheHit {
			entry.RLock()
			defer entry.RUnlock()
			if !entry.expiry.IsZero() {
				return entry.duplicate(), nil
			}
		}
		return nil, err
	}

	newEntry, err := c.updateIDToDomainCache(id, record)
	if err != nil {
		// err updating, use the existing record if record is valid
		// i.e. expiry is set
		if cacheHit {
			entry.RLock()
			defer entry.RUnlock()
			if !entry.expiry.IsZero() {
				return entry.duplicate(), nil
			}
		}
		return nil, err
	}
	return newEntry, nil
}

func (c *domainCache) triggerDomainChangeCallback(prevDomain *DomainCacheEntry, nextDomain *DomainCacheEntry) {
	c.RLock()
	defer c.RUnlock()
	for _, callback := range c.callbacks {
		callback(prevDomain, nextDomain)
	}
}

func (entry *DomainCacheEntry) duplicate() *DomainCacheEntry {
	result := newDomainCacheEntry(entry.clusterMetadata)
	result.info = entry.info
	result.config = entry.config
	result.replicationConfig = entry.replicationConfig
	result.configVersion = entry.configVersion
	result.failoverVersion = entry.failoverVersion
	result.isGlobalDomain = entry.isGlobalDomain
	return result
}

func (entry *DomainCacheEntry) isExpired(now time.Time) bool {
	return entry.expiry.IsZero() || now.After(entry.expiry)
}

// GetInfo return the domain info
func (entry *DomainCacheEntry) GetInfo() *persistence.DomainInfo {
	return entry.info
}

// GetConfig return the domain config
func (entry *DomainCacheEntry) GetConfig() *persistence.DomainConfig {
	return entry.config
}

// GetReplicationConfig return the domain replication config
func (entry *DomainCacheEntry) GetReplicationConfig() *persistence.DomainReplicationConfig {
	return entry.replicationConfig
}

// GetConfigVersion return the domain config version
func (entry *DomainCacheEntry) GetConfigVersion() int64 {
	return entry.configVersion
}

// GetFailoverVersion return the domain failover version
func (entry *DomainCacheEntry) GetFailoverVersion() int64 {
	return entry.failoverVersion
}

// IsGlobalDomain return whether the domain is a global domain
func (entry *DomainCacheEntry) IsGlobalDomain() bool {
	return entry.isGlobalDomain
}

// IsDomainActive return whether the domain is active, i.e. non global domain or global domain which active cluster is the current cluster
func (entry *DomainCacheEntry) IsDomainActive() bool {
	if !entry.isGlobalDomain {
		// domain is not a global domain, meaning domain is always "active" within each cluster
		return true
	}
	return entry.clusterMetadata.GetCurrentClusterName() == entry.replicationConfig.ActiveClusterName
}

// ShouldReplicateEvent return whether the workflows within this domain should be replicated
func (entry *DomainCacheEntry) ShouldReplicateEvent() bool {
	// frontend guarantee that the clusters always contains the active domain, so if the # of clusters is 1
	// then we do not need to send out any events for replication
	return entry.clusterMetadata.GetCurrentClusterName() == entry.replicationConfig.ActiveClusterName &&
		entry.isGlobalDomain && len(entry.replicationConfig.Clusters) > 1
}

// GetDomainNotActiveErr return err if domain is not active, nil otherwise
func (entry *DomainCacheEntry) GetDomainNotActiveErr() error {
	if entry.IsDomainActive() {
		// domain is consider active
		return nil
	}
	return errors.NewDomainNotActiveError(entry.info.Name, entry.clusterMetadata.GetCurrentClusterName(), entry.replicationConfig.ActiveClusterName)
}
