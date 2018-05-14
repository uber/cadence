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
	domainEntryRefreshInterval = 15 * time.Second
	domainEntryRefreshPageSize = 100

	domainCacheInitialized int32 = 0
	domainCacheStarted     int32 = 1
	domainCacheStopped     int32 = 2
)

type (
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
		metadataMgr     persistence.MetadataManager
		clusterMetadata cluster.Metadata
		timeSource      common.TimeSource
		logger          bark.Logger

		sync.RWMutex
		callbacks    map[int]callbackFn
		nameToDomain map[string]*DomainCacheEntry
		idToName     map[string]string
	}

	// DomainCacheEntry contains the info and config for a domain
	DomainCacheEntry struct {
		clusterMetadata cluster.Metadata

		info              *persistence.DomainInfo
		config            *persistence.DomainConfig
		replicationConfig *persistence.DomainReplicationConfig
		configVersion     int64
		failoverVersion   int64
		isGlobalDomain    bool
	}
)

// NewDomainCache creates a new instance of cache for holding onto domain information to reduce the load on persistence
func NewDomainCache(metadataMgr persistence.MetadataManager, clusterMetadata cluster.Metadata, logger bark.Logger) DomainCache {
	return &domainCache{
		status:          domainCacheInitialized,
		shutdownChan:    make(chan struct{}),
		nameToDomain:    make(map[string]*DomainCacheEntry),
		idToName:        make(map[string]string),
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

// Start start the background scan of domain
func (c *domainCache) Start() {
	if !atomic.CompareAndSwapInt32(&c.status, domainCacheInitialized, domainCacheStarted) {
		return
	}
	c.refreshDomains()
	go c.refreshLoop()
}

// Start start the background scan of domain
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

func (c *domainCache) refreshDomains() error {
	var token []byte
	request := &persistence.ListDomainRequest{
		PageSize: domainEntryRefreshPageSize,
	}

	// get all domains from database
	entries := []*DomainCacheEntry{}
	continuePage := true
	for continuePage {
		request.NextPageToken = token
		response, err := c.metadataMgr.ListDomain(request)
		if err != nil {
			return err
		}
		token = response.NextPageToken
		for _, record := range response.Domains {
			entries = append(entries, c.buildEntryFromRecord(record))
		}
		continuePage = len(token) != 0
	}

	for _, entry := range entries {
		c.updateCache(entry)
	}
	return nil
}

func (c *domainCache) loadDomain(id string, name string) (*DomainCacheEntry, error) {
	response, err := c.metadataMgr.GetDomain(&persistence.GetDomainRequest{Name: name, ID: id})
	// Failed to get domain.  Return stale entry if we have one, otherwise just return error
	if err != nil {
		return nil, err
	}

	return c.buildEntryFromRecord(response), nil
}

func (c *domainCache) buildEntryFromRecord(record *persistence.GetDomainResponse) *DomainCacheEntry {
	newEntry := newDomainCacheEntry(c.clusterMetadata)
	newEntry.info = record.Info
	newEntry.config = record.Config
	newEntry.replicationConfig = record.ReplicationConfig
	newEntry.configVersion = record.ConfigVersion
	newEntry.failoverVersion = record.FailoverVersion
	newEntry.isGlobalDomain = record.IsGlobalDomain
	return newEntry
}

func (c *domainCache) updateCache(newEntry *DomainCacheEntry) *DomainCacheEntry {

	c.Lock()
	oldEntry, ok := c.nameToDomain[newEntry.info.Name]
	if ok {
		if oldEntry.configVersion < newEntry.configVersion {
			c.nameToDomain[newEntry.info.Name] = newEntry
			c.Unlock()
			// do notification
			if c.clusterMetadata.IsGlobalDomainEnabled() {
				c.triggerDomainChangeCallback(oldEntry.duplicate(), newEntry.duplicate())
			}
			return newEntry
		}
		// use the old entry / domain since old has config version >= new
		c.Unlock()
		return oldEntry
	}

	// no prev domain in cache
	c.nameToDomain[newEntry.info.Name] = newEntry
	c.idToName[newEntry.info.ID] = newEntry.info.Name
	c.Unlock()
	return newEntry
}

// RegisterDomainChangeCallback set a domain failover callback, which will be when active domain for a domain changes
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

	c.RLock()
	name, ok := c.idToName[id]
	c.RUnlock()
	if !ok {
		entry, err := c.loadDomain(id, "")
		if err != nil {
			return nil, err
		}
		// the domain name to id relation is immutable
		name = entry.info.Name
		c.updateCache(entry)
	}

	return c.getDomain(name)
}

// GetDomainID retrieves domainID by using GetDomain
func (c *domainCache) GetDomainID(name string) (string, error) {
	entry, err := c.GetDomain(name)
	if err != nil {
		return "", err
	}
	return entry.info.ID, nil
}

// GetDomain retrieves the information from the cache if it exists, otherwise retrieves the information from metadata
// store and writes it to the cache with an expiry before returning back
func (c *domainCache) getDomain(name string) (*DomainCacheEntry, error) {
	var result *DomainCacheEntry

	c.RLock()
	entry, cacheHit := c.nameToDomain[name]
	c.RUnlock()

	if cacheHit {
		// Found the information in the cache, lets check if it needs to be refreshed before returning back
		result = entry.duplicate()
		return result, nil
	}

	entry, err := c.loadDomain("", name)
	if err != nil {
		return nil, err
	}
	entry = c.updateCache(entry)
	result = entry.duplicate()
	return result, nil
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
