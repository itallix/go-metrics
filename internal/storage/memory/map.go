package memory

import "sync"

// ConcurrentMap is a thread-safe map implementation.
// It uses sync.RWMutex to protect read/update operations to standard library map.
type ConcurrentMap[T float64 | int64] struct {
	store map[string]T
	mu    sync.RWMutex
}

func NewConcurrentMap[T float64 | int64](size int) *ConcurrentMap[T] {
	return &ConcurrentMap[T]{
		store: make(map[string]T, size),
	}
}

func (m *ConcurrentMap[T]) Set(name string, value T) T {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[name] = value
	return m.store[name]
}

func (m *ConcurrentMap[T]) Inc(name string, value T) T {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[name] += value
	return m.store[name]
}

func (m *ConcurrentMap[T]) Get(name string) (T, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.store[name]
	return val, ok
}

func (m *ConcurrentMap[T]) Copy() map[string]T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	clone := make(map[string]T, len(m.store))
	for key, value := range m.store {
		clone[key] = value
	}
	return clone
}

func (m *ConcurrentMap[T]) Init(values map[string]T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = values
}

func (m *ConcurrentMap[T]) Len() int {
	return len(m.store)
}
