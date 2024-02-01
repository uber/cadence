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

package syncmap

import "sync"

// syncmap is a very simple type-safe locked map, with semantics similar to sync.Map,
// but only supports inserting once and getting.
type (
	syncmap[K comparable, V any] struct {
		mut  sync.Mutex
		data map[K]V
	}
	SyncMap[K comparable, V any] interface {
		Get(key K) (value V, ok bool)
		Put(key K, value V) (inserted bool)
	}
)

func New[K comparable, V any]() SyncMap[K, V] {
	return &syncmap[K, V]{
		data: make(map[K]V),
	}
}

func (m *syncmap[K, V]) Get(key K) (value V, ok bool) {
	m.mut.Lock()
	defer m.mut.Unlock()
	value, ok = m.data[key]
	return value, ok
}

func (m *syncmap[K, V]) Put(key K, value V) (inserted bool) {
	m.mut.Lock()
	defer m.mut.Unlock()
	if _, ok := m.data[key]; ok {
		return false
	}
	m.data[key] = value
	return true
}
