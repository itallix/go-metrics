package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentMap_UpdateValue(t *testing.T) {
	memStorage := NewConcurrentMap[int64]()
	memStorage.Inc("counter0", 123)
	memStorage.Inc("counter0", 132)
	val, ok := memStorage.Get("counter0")
	assert.True(t, ok)
	assert.Equal(t, int64(255), val)
}

func TestStorageTest_SetValue(t *testing.T) {
	memStorage := NewConcurrentMap[float64]()
	memStorage.Set("gauge0", 123.983)
	memStorage.Set("gauge0", 132.625)
	val, ok := memStorage.Get("gauge0")
	assert.True(t, ok)
	assert.InEpsilon(t, 132.625, val, 0.00001)
}

func TestStorageTest_ValueDoesntExist(t *testing.T) {
	memStorage := NewConcurrentMap[int64]()
	memStorage.Inc("counter0", 123)

	val, ok := memStorage.Get("metric")

	assert.False(t, ok)
	assert.Zero(t, val)
}
