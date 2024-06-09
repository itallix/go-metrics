package storage

import "sync"

type Storage[T float64 | int64] interface {
	Set(name string, value T) T
	Update(name string, value T) T
	Get(name string) (T, bool)
	Copy() map[string]T
}

type MemStorage[T float64 | int64] struct {
	store map[string]T
	mu    sync.RWMutex
}

func NewMemStorage[T float64 | int64]() *MemStorage[T] {
	return &MemStorage[T]{
		store: make(map[string]T),
	}
}

func (m *MemStorage[T]) Set(name string, value T) T {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[name] = value
	return m.store[name]
}

func (m *MemStorage[T]) Update(name string, value T) T {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[name] += value
	return m.store[name]
}

func (m *MemStorage[T]) Get(name string) (T, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.store[name]
	return val, ok
}

func (m *MemStorage[T]) Copy() map[string]T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	clone := make(map[string]T, len(m.store))
	for key, value := range m.store {
		clone[key] = value
	}
	return clone
}
