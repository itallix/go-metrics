package storage

import "sync"

type Storage[T any] interface {
	Set(name string, value T)
	Get(name string) (T, bool)
}

type MemStorage[T any] struct {
	store map[string]T
	mu    sync.RWMutex
}

func NewMemStorage[T any]() *MemStorage[T] {
	return &MemStorage[T]{
		store: make(map[string]T),
	}
}

func (m *MemStorage[T]) Set(name string, value T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[name] = value
}

func (m *MemStorage[T]) Get(name string) (T, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	counter, ok := m.store[name]
	return counter, ok
}
