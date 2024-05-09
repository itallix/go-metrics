package storage_test

import (
	"testing"

	"github.com/itallix/go-metrics/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestStorageTest_UpdateValue(t *testing.T) {
	memStorage := storage.NewMemStorage[int]()
	memStorage.Update("counter0", 123)
	memStorage.Update("counter0", 132)
	val, ok := memStorage.Get("counter0")
	assert.True(t, ok)
	assert.Equal(t, 255, val)
}

func TestStorageTest_SetValue(t *testing.T) {
	memStorage := storage.NewMemStorage[float64]()
	memStorage.Set("gauge0", 123.983)
	memStorage.Set("gauge0", 132.625)
	val, ok := memStorage.Get("gauge0")
	assert.True(t, ok)
	assert.InEpsilon(t, 132.625, val, 0.00001)
}

func TestStorageTest_ValueDoesntExist(t *testing.T) {
	memStorage := storage.NewMemStorage[int]()
	memStorage.Update("counter0", 123)

	val, ok := memStorage.Get("metric")

	assert.False(t, ok)
	assert.Zero(t, val)
}
