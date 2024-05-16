package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectMetrics(t *testing.T) {
	metrics := newAgent()

	assert.Empty(t, metrics.Gauges)

	metrics.collect()

	for _, key := range metrics.RegisteredMetrics {
		_, exists := metrics.Gauges[key]
		assert.Truef(t, exists, "Expected key %s is missing in the map", key)
	}
	assert.Equal(t, uint64(1), metrics.Counter)
}
