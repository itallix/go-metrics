package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestCollectRuntimeMetrics(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	agent := newAgent(client, "")

	assert.Empty(t, agent.Gauges)

	err := agent.collectRuntime()
	require.NoError(t, err)

	for _, key := range RuntimeMetrics {
		_, exists := agent.Gauges[key]
		assert.Truef(t, exists, "Expected key %s is missing in the map", key)
	}
	assert.Equal(t, int64(1), agent.Counter)
}

func TestCollectExtraMetrics(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	agent := newAgent(client, "")

	assert.Empty(t, agent.Gauges)

	err := agent.collectExtra()
	require.NoError(t, err)

	metrics := []string{"TotalMemory", "FreeMemory", "CPUutilization1"}

	for _, key := range metrics {
		_, exists := agent.Gauges[key]
		assert.Truef(t, exists, "Expected key %s is missing in the map", key)
	}
}
