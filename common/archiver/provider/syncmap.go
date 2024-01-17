package provider

import "sync"

// syncmap is a very simple type-safe wrapper around sync.Map,
// which only supports inserting once and getting.
type syncmap[K comparable, V any] struct {
	mut  sync.Mutex
	data map[K]V
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
