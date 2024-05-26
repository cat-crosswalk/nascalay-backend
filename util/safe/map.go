package safe

import "sync"

type Map[K comparable, V any] struct {
	m   map[K]V
	mux sync.RWMutex
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{m: make(map[K]V)}
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	value, ok = m.m[key]
	return
}

func (m *Map[K, V]) Store(key K, value V) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.m[key] = value
}

func (m *Map[K, V]) Delete(key K) {
	m.mux.Lock()
	defer m.mux.Unlock()
	delete(m.m, key)
}
