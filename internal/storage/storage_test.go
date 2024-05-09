package storage_test

import (
	"testing"

	"github.com/itallix/go-metrics/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestStorageTest_MetricExists(t *testing.T) {
	memStorage := storage.NewMemStorage[int]()
	memStorage.Set("counter0", 123)
	val, ok := memStorage.Get("counter0")
	assert.True(t, ok)
	assert.Equal(t, 123, val)
}

func TestStorageTest_MetricDoesntExist(t *testing.T) {
	memStorage := storage.NewMemStorage[int]()
	memStorage.Set("counter0", 123)

	val, ok := memStorage.Get("metric")

	assert.False(t, ok)
	assert.Zero(t, val)
}
