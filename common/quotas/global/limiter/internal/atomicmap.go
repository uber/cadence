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

package internal

import (
	"sync"
	"sync/atomic"
)

// AtomicMap adds type safety around a sync.Map (which has atomic-like behavior), and:
//   - implicitly constructs values as needed, not relying on zero values
//   - simplifies the API a bit because not all methods are in use.
//     in particular there is no "Store" currently because it is not needed.
//   - tracks length (atomically, so values are only an estimate)
//
// Due to length tracking, this is marginally more costly when modifying contents
// than "just" a type-safe sync.Map.  It should only be used when length is needed.
type AtomicMap[Key comparable, Value any] struct {
	contents sync.Map
	create   func(key Key) Value
	len      int64
}

// NewAtomicMap makes a simplified type-safe [sync.Map] that creates values as needed, and tracks length.
//
// The `create` callback will be called when creating a new value, possibly multiple times,
// without synchronization.
// It must be concurrency safe and should return ASAP to reduce the window for storage races,
// so ideally it should be simple and non-blocking (or consider filling a [sync.Pool] if not).
//
// Due to length tracking, this is marginally more costly when modifying contents
// than "just" a type-safe [sync.Map].  It should only be used when length is needed.
func NewAtomicMap[Key comparable, Value any](create func(key Key) Value) *AtomicMap[Key, Value] {
	return &AtomicMap[Key, Value]{
		contents: sync.Map{},
		create:   create,
		len:      0,
	}
}

// Load will get the current Value for a Key, initializing it if necessary.
func (t *AtomicMap[Key, Value]) Load(key Key) Value {
	val, loaded := t.contents.Load(key)
	if loaded {
		return val.(Value)
	}
	created := t.create(key)
	val, loaded = t.contents.LoadOrStore(key, created)
	if !loaded {
		// stored a new value
		atomic.AddInt64(&t.len, 1)
	}
	return val.(Value)
}

// Try will get the current Value for a key, or return false if it did not exist.
//
// Unlike Load, this will NOT populate the key if it does not exist.
// It just calls [sync.Map.Load] on the underlying map.
func (t *AtomicMap[Key, Value]) Try(key Key) (Value, bool) {
	v, ok := t.contents.Load(key)
	if ok {
		return v.(Value), true
	}
	var zero Value
	return zero, false
}

// Delete removes an entry from the map, and updates the length.
//
// Like the underlying [sync.Map.LoadAndDelete], this can be called concurrently with Range.
func (t *AtomicMap[Key, Value]) Delete(k Key) {
	_, loaded := t.contents.LoadAndDelete(k)
	if loaded {
		atomic.AddInt64(&t.len, -1)
	}
}

// Range calls [sync.Map.Range] on the underlying [sync.Map], and has the same semantics.
//
// This can be used while concurrently modifying the map, and it may result
// in ranging over more or fewer entries than Len would imply.
func (t *AtomicMap[Key, Value]) Range(f func(k Key, v Value) bool) {
	t.contents.Range(func(k, v any) bool {
		return f(k.(Key), v.(Value))
	})
}

// Len is an atomic count of the size of the collection, i.e. it can never be assumed
// to be precise.
//
// In particular, Range may iterate over more or fewer entries.
func (t *AtomicMap[Key, Value]) Len() int {
	return int(atomic.LoadInt64(&t.len))
}
