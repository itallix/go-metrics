package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectMetrics(t *testing.T) {
	metrics := NewMetrics()

	assert.Empty(t, metrics.Values)

	metrics.collect()

	for _, key := range metrics.RegisteredMetrics {
		_, exists := metrics.Values[key]
		assert.Truef(t, exists, "Expected key %s is missing in the map", key)
	}
}
