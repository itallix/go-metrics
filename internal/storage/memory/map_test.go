package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentMap_UpdateValue(t *testing.T) {
	m := NewConcurrentMap[int64]()
	m.Inc("counter0", 123)
	m.Inc("counter0", 132)
	val, ok := m.Get("counter0")
	assert.True(t, ok)
	assert.Equal(t, int64(255), val)
}

func TestConcurrentMap_SetValue(t *testing.T) {
	m := NewConcurrentMap[float64]()
	m.Set("gauge0", 123.983)
	m.Set("gauge0", 132.625)
	val, ok := m.Get("gauge0")
	assert.True(t, ok)
	assert.InEpsilon(t, 132.625, val, 0.00001)
}

func TestConcurrentMap_ValueDoesntExist(t *testing.T) {
	m := NewConcurrentMap[int64]()
	m.Inc("counter0", 123)

	val, ok := m.Get("metric")

	assert.False(t, ok)
	assert.Zero(t, val)
}

func TestConcurrentMap_Copy(t *testing.T) {
	m1 := NewConcurrentMap[int64]()
	m1.Inc("counter0", 123)

	m2 := m1.Copy()
	m2["counter0"] = 234

	val, ok := m1.Get("counter0")
	assert.Equal(t, int64(123), val)
	assert.True(t, ok)
	assert.Equal(t, int64(234), m2["counter0"])
}
