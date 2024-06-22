package main

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestCollectMetrics(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	metrics := newAgent(client, "")

	assert.Empty(t, metrics.Gauges)

	metrics.collect()

	for _, key := range metrics.RegisteredMetrics {
		_, exists := metrics.Gauges[key]
		assert.Truef(t, exists, "Expected key %s is missing in the map", key)
	}
	assert.Equal(t, int64(1), metrics.Counter)
}
