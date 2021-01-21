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
	"time"

	"github.com/uber/cadence/common"

	"github.com/uber/cadence/common/metrics"
)

// A Cache is a generalized interface to a cache.  See cache.LRU for a specific
// implementation (bounded cache with LRU eviction)
type Cache interface {
	// Get retrieves an element based on a key, returning nil if the element
	// does not exist
	Get(key interface{}) interface{}

	// Put adds an element to the cache, returning the previous element
	Put(key interface{}, value interface{}) interface{}

	// PutIfNotExist puts a value associated with a given key if it does not exist
	PutIfNotExist(key interface{}, value interface{}) (interface{}, error)

	// Delete deletes an element in the cache
	Delete(key interface{})

	// Release decrements the ref count of a pinned element. If the ref count
	// drops to 0, the element can be evicted from the cache.
	Release(key interface{})

	// Iterator returns the iterator of the cache
	Iterator() Iterator

	// Size returns the number of entries currently stored in the Cache
	Size() int
}

// Options control the behavior of the cache
type Options struct {
	// TTL controls the time-to-live for a given cache entry.  Cache entries that
	// are older than the TTL will not be returned.
	TTL time.Duration

	// InitialCapacity controls the initial capacity of the cache
	InitialCapacity int

	// Pin prevents in-use objects from getting evicted.
	Pin bool

	// RemovedFunc is an optional function called when an element
	// is scheduled for deletion
	RemovedFunc RemovedFunc

	// UpdatedFunc is an optional function called when the cache item was put/deleted
	UpdatedFunc UpdatedFunc

	// MaxCount controls the max capacity of the cache
	// It is required option if MaxSize is not provided
	MaxCount int

	// GetCacheItemSizeFunc is a function called upon adding the item to update the cache size.
	// It returns 0 by default, assuming the cache is just count based
	// It is required option if MaxCount is not provided
	GetCacheItemSizeFunc GetCacheItemSizeFunc

	// MaxSize is an optional and must be set along with GetCacheItemSizeFunc
	// to control the max size in bytes of the cache
	// It is required option if MaxCount is not provided
	MaxSize uint64
}

// SimpleOptions provides options that can be used to configure SimpleCache
type SimpleOptions struct {
	// InitialCapacity controls the initial capacity of the cache
	InitialCapacity int

	// RemovedFunc is an optional function called when an element
	// is scheduled for deletion
	RemovedFunc RemovedFunc
}

// RemovedFunc is a type for notifying applications when an item is
// scheduled for removal from the Cache. If f is a function with the
// appropriate signature and i is the interface{} scheduled for
// deletion, Cache calls go f(i)
type RemovedFunc func(interface{})

// UpdatedFunc is a type for notifying applications when the cache item was put/deleted
type UpdatedFunc func(newSize int)

// Iterator represents the interface for cache iterators
type Iterator interface {
	// Close closes the iterator
	// and releases any allocated resources
	Close()
	// HasNext return true if there is more items to be returned
	HasNext() bool
	// Next return the next item
	Next() Entry
}

// Entry represents a key-value entry within the map
type Entry interface {
	// Key represents the key
	Key() interface{}
	// Value represents the value
	Value() interface{}
	// CreateTime represents the time when the entry is created
	CreateTime() time.Time
}

// GetCacheItemSizeFunc returns the cache item size in bytes
type GetCacheItemSizeFunc func(interface{}) uint64

// DomainMetricsScopeCache represents a interface for mapping domainID and scopeIdx to metricsScope
type DomainMetricsScopeCache interface {
	// Get retrieves metrics scope for a domainID and scopeIdx
	Get(domainID string, scopeIdx int) (metrics.Scope, bool)
	// Put adds metrics scope for a domainID and scopeIdx
	Put(domainID string, scopeIdx int, metricsScope metrics.Scope)

	common.Daemon
}
