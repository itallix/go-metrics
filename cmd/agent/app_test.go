package main

import (
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestCollectRuntimeMetrics(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	agent, err := newAgent(client, "")
	require.NoError(t, err)

	assert.Empty(t, agent.Gauges)

	err = agent.collectRuntime()
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
	agent, err := newAgent(client, "")
	require.NoError(t, err)
	assert.Empty(t, agent.Gauges)

	err = agent.collectExtra()
	require.NoError(t, err)

	for _, key := range []string{"TotalMemory", "FreeMemory"} {
		_, exists := agent.Gauges[key]
		assert.Truef(t, exists, "Expected key %s is missing in the map", key)
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		metricName := "CPUutilization" + strconv.Itoa(i)
		_, exists := agent.Gauges[metricName]
		assert.Truef(t, exists, "Expected key %s is missing in the map", metricName)
	}
}
