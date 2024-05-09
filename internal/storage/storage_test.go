package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStorageTest_MetricExists(t *testing.T) {
	memStorage := NewMemStorage[int]()
	memStorage.Set("counter0", 123)
	val, ok := memStorage.Get("counter0")
	assert.True(t, ok)
	assert.Equal(t, 123, val)
}

func TestStorageTest_MetricDoesntExist(t *testing.T) {
	memStorage := NewMemStorage[int]()
	memStorage.Set("counter0", 123)

	val, ok := memStorage.Get("metric")

	assert.False(t, ok)
	assert.Zero(t, val)
}
