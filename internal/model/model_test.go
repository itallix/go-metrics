package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModel_NewCounter(t *testing.T) {
	c := int64(64)
	m := NewCounter("c0", &c)

	assert.Equal(t, int64(64), *m.Delta)
	assert.Nil(t, m.Value)
	assert.Equal(t, "c0", m.ID)
	assert.Equal(t, Counter, m.MType)
}

func TestModel_NewGauge(t *testing.T) {
	g := float64(64)
	m := NewGauge("g0", &g)

	assert.Equal(t, float64(64), *m.Value)
	assert.Nil(t, m.Delta)
	assert.Equal(t, "g0", m.ID)
	assert.Equal(t, Gauge, m.MType)
}

func TestModel_String(t *testing.T) {
	c := int64(64)
	mc := NewCounter("c0", &c)
	g := float64(64)
	mg := NewGauge("g0", &g)

	assert.Equal(t, "counter: c0 = 64", mc.String())
	assert.Equal(t, "gauge: g0 = 64.000000", mg.String())
}
