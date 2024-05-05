package storage

import "sync"

type Storage interface {
	Set(name string, value float64)
	Get(name string) (float64, bool)
	Delete(name string)
}

type MemStorage struct {
	metrics map[string]float64
	mu      sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]float64),
	}
}

func (m *MemStorage) Set(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[name] = value
}

func (m *MemStorage) Get(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.metrics[name]
	return value, ok
}

func (m *MemStorage) Delete(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.metrics, name)
}
