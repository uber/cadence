package collection

import (
	"sync"
	"sync/atomic"
)

const (
	// nShards represents the number of shards
	// At any given point of time, there can only
	// be nShards number of concurrent writers to
	// the map at max
	nShards = 32
)

type (
	// HashFunc represents a hash function for string
	HashFunc func(string) uint32

	// ConcurrentMap is a generic interface for any
	// implementation of a dictionary or a key value
	// lookup table that is thread safe.
	ConcurrentMap interface {
		// Get returns the value for the given key
		Get(key string) (interface{}, bool)
		// Contains returns true if the key exist and false otherwise
		Contains(key string) bool
		// Put records the mapping from given key to value
		Put(key string, value interface{})
		// PutIfNotExist records the key value mapping only
		// if the mapping does not already exist
		PutIfNotExist(key string, value interface{}) bool
		// Remove deletes the key from the map
		Remove(key string)
		// Iter returns an iterator to the map
		Iter() MapIterator
		// Size returns the number of items in the map
		Size() int
	}

	// MapIterator represents the interface for
	// map iterators
	MapIterator interface {
		// Close closes the iterator
		// and releases any allocated resources
		Close()
		// Entries returns a channel of MapEntry
		// objects that can be used in a range loop
		Entries() <-chan *MapEntry
	}

	// MapEntry represents a key-value entry within the map
	MapEntry struct {
		// Key represents the key
		Key string
		// Value represents the value
		Value interface{}
	}

	// ShardedConcurrentMap is an implementation of
	// ConcurrentMap that internally uses multiple
	// sharded maps to increase parallelism
	ShardedConcurrentMap struct {
		shards     [nShards]mapShard
		hashfn     HashFunc
		size       int32
		initialCap int
	}

	// mapIteratorImpl represents an iterator type
	// for the concurrent map.
	mapIteratorImpl struct {
		stopCh chan struct{}
		dataCh chan *MapEntry
	}

	// mapShard represents a single instance
	// of thread safe map
	mapShard struct {
		sync.RWMutex
		items map[string]interface{}
	}
)

// NewShardedConcurrentMap returns an instance of ShardedConcurrentMap
//
// ShardedConcurrentMap is a thread safe map that maintains upto nShards
// number of maps internally to allow nShards writers to be acive at the
// same time. This map *does not* use re-entrant locks, so access to the
// map during iterator can cause a dead lock.
//
// @param initialSz
//		The initial size for the map
// @param hashfn
// 		The hash function to use for sharding
func NewShardedConcurrentMap(initialCap int, hashfn HashFunc) ConcurrentMap {
	cmap := new(ShardedConcurrentMap)
	cmap.hashfn = hashfn
	cmap.initialCap = MaxInt(nShards, initialCap/nShards)
	return cmap
}

// Get returns the value corresponding to the key, if it exist
func (cmap *ShardedConcurrentMap) Get(key string) (interface{}, bool) {
	shard := cmap.getShard(key)
	var ok bool
	var value interface{}
	shard.RLock()
	if shard.items != nil {
		value, ok = shard.items[key]
	}
	shard.RUnlock()
	return value, ok
}

// Contains returns true if the key exist and false otherwise
func (cmap *ShardedConcurrentMap) Contains(key string) bool {
	_, ok := cmap.Get(key)
	return ok
}

// Put records the given key value mapping. Overwrites previous values
func (cmap *ShardedConcurrentMap) Put(key string, value interface{}) {
	shard := cmap.getShard(key)
	shard.Lock()
	cmap.lazyInitShard(shard)
	shard.items[key] = value
	atomic.AddInt32(&cmap.size, 1)
	shard.Unlock()
}

// PutIfNotExist records the mapping, if there is no mapping for this key already
// Returns true if the mapping was recorded, false otherwise
func (cmap *ShardedConcurrentMap) PutIfNotExist(key string, value interface{}) bool {
	shard := cmap.getShard(key)
	var ok bool
	shard.Lock()
	cmap.lazyInitShard(shard)
	_, ok = shard.items[key]
	if !ok {
		shard.items[key] = value
		atomic.AddInt32(&cmap.size, 1)
	}
	shard.Unlock()
	return !ok
}

// Remove deletes the given key from the map
func (cmap *ShardedConcurrentMap) Remove(key string) {
	shard := cmap.getShard(key)
	shard.Lock()
	delete(shard.items, key)
	atomic.AddInt32(&cmap.size, -1)
	shard.Unlock()
}

// Close closes the iterator
func (it *mapIteratorImpl) Close() {
	close(it.stopCh)
}

// Entries returns a channel of map entries
func (it *mapIteratorImpl) Entries() <-chan *MapEntry {
	return it.dataCh
}

// Iter returns an iterator to the map. This map
// does not use re-entrant locks, so access or modification
// to the map during iteration can cause a dead lock.
func (cmap *ShardedConcurrentMap) Iter() MapIterator {

	iterator := new(mapIteratorImpl)
	iterator.dataCh = make(chan *MapEntry, 8)
	iterator.stopCh = make(chan struct{})

	go func(iterator *mapIteratorImpl) {
		for i := 0; i < nShards; i++ {
			cmap.shards[i].RLock()
			for k, v := range cmap.shards[i].items {
				entry := &MapEntry{Key: k, Value: v}
				select {
				case iterator.dataCh <- entry:
				case <-iterator.stopCh:
					cmap.shards[i].RUnlock()
					close(iterator.dataCh)
					return
				}
			}
			cmap.shards[i].RUnlock()
		}
		close(iterator.dataCh)
	}(iterator)

	return iterator
}

// Size returns the number of items in the map
func (cmap *ShardedConcurrentMap) Size() int {
	return int(atomic.LoadInt32(&cmap.size))
}

func (cmap *ShardedConcurrentMap) getShard(key string) *mapShard {
	shardIdx := cmap.hashfn(key) % nShards
	return &cmap.shards[shardIdx]
}

func (cmap *ShardedConcurrentMap) lazyInitShard(shard *mapShard) {
	if shard.items == nil {
		shard.items = make(map[string]interface{}, cmap.initialCap)
	}
}
