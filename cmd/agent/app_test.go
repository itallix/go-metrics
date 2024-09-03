package main

import (
	"context"
	"runtime"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/itallix/go-metrics/internal/model"
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

func TestSendMetrics(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	httpmock.RegisterResponder("POST", "/updates/",
		httpmock.NewStringResponder(200, `{"status":"success"}`))
	agent, err := newAgent(client, "")
	require.NoError(t, err)
	assert.Empty(t, agent.Gauges)

	err = agent.collectRuntime()
	require.NoError(t, err)

	jobs := make(chan []model.Metrics, 1)
	results := make(chan error, 1)

	go func() {
		jobs <- agent.metrics()
	}()
	go func() {
		agent.send(context.Background(), jobs, results)
	}()

	err = <-results
	require.NoError(t, err)

	info := httpmock.GetCallCountInfo()
	assert.Equal(t, 1, info["POST /updates/"])

	close(jobs)
	close(results)
}
