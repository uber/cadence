// Copyright (c) 2021 Uber Technologies, Inc.
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

package collection

import (
	"container/list"
	"sync"
)

type (
	// OrderedMap is an interface for a dictionary which during iteration
	// return all values in their insertion order
	OrderedMap interface {
		// Get returns the value for the given key
		Get(key interface{}) (interface{}, bool)
		// Contains returns true if the key exist and false otherwise
		Contains(key interface{}) bool
		// Put records the mapping from given key to value
		Put(key interface{}, value interface{})
		// Remove deletes the key from the map
		Remove(key interface{})
		// Iter returns an iterator to the map which iterates over map entries
		// in insertion order
		Iter() MapIterator
		// Len returns the number of items in the map
		Len() int
	}

	orderedMapValue struct {
		value interface{}

		listElement *list.Element
	}

	orderedMap struct {
		items map[interface{}]orderedMapValue
		list  *list.List
	}

	concurrentOrderedMap struct {
		sync.RWMutex

		orderedMap *orderedMap
	}
)

// NewOrderedMap creates a new OrderedMap
// implementation has O(1) complexity for all methods
func NewOrderedMap() OrderedMap {
	return &orderedMap{
		items: make(map[interface{}]orderedMapValue),
		list:  list.New(),
	}
}

func (m *orderedMap) Get(key interface{}) (interface{}, bool) {
	if mapValue, ok := m.items[key]; ok {
		return mapValue.value, true
	}

	return nil, false
}

func (m *orderedMap) Contains(key interface{}) bool {
	_, ok := m.items[key]
	return ok
}

func (m *orderedMap) Put(key interface{}, value interface{}) {
	// remove existing key if there's one
	m.Remove(key)

	listElement := m.list.PushBack(&MapEntry{
		Key:   key,
		Value: value,
	})
	m.items[key] = orderedMapValue{
		value:       value,
		listElement: listElement,
	}
}

func (m *orderedMap) Remove(key interface{}) {
	mapValue, ok := m.items[key]
	if !ok {
		return
	}

	m.list.Remove(mapValue.listElement)
	delete(m.items, key)
}

func (m *orderedMap) Iter() MapIterator {
	iterator := newMapIterator()

	go orderedMapIteratorLoop(nil, iterator, m.list)

	return iterator
}

func (m *orderedMap) Len() int {
	return len(m.items)
}

// NewConcurrentOrderedMap creates a new thread safe OrderedMap
func NewConcurrentOrderedMap() OrderedMap {
	return &concurrentOrderedMap{
		orderedMap: NewOrderedMap().(*orderedMap),
	}
}

func (m *concurrentOrderedMap) Get(key interface{}) (interface{}, bool) {
	m.RLock()
	defer m.RUnlock()

	return m.orderedMap.Get(key)
}

func (m *concurrentOrderedMap) Contains(key interface{}) bool {
	m.RLock()
	defer m.RUnlock()

	return m.orderedMap.Contains(key)
}

func (m *concurrentOrderedMap) Put(key interface{}, value interface{}) {
	m.Lock()
	defer m.Unlock()

	m.orderedMap.Put(key, value)
}

func (m *concurrentOrderedMap) Remove(key interface{}) {
	m.Lock()
	defer m.Unlock()

	m.orderedMap.Remove(key)
}

// Iter returns an iterator to the map.
// NOTE that any modification to the map during iteration
// will lead to a dead lock.
func (m *concurrentOrderedMap) Iter() MapIterator {
	iterator := newMapIterator()

	go orderedMapIteratorLoop(&m.RWMutex, iterator, m.orderedMap.list)

	return iterator
}

func (m *concurrentOrderedMap) Len() int {
	m.RLock()
	defer m.RUnlock()

	return m.orderedMap.Len()
}

func orderedMapIteratorLoop(
	mutex *sync.RWMutex,
	iterator *mapIteratorImpl,
	list *list.List,
) {
	if mutex != nil {
		mutex.RLock()
		defer mutex.RUnlock()
	}

	for element := list.Front(); element != nil; element = element.Next() {
		select {
		case iterator.dataCh <- element.Value.(*MapEntry):
			// no-op
		case <-iterator.stopCh:
			close(iterator.dataCh)
			return
		}
	}
	close(iterator.dataCh)
}
