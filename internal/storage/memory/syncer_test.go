package memory

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/itallix/go-metrics/internal/model"
)

func TestFileSyncer_SyncDelay(t *testing.T) {
	f, err := os.CreateTemp("", "file_syncer_load_test")
	require.NoError(t, err)
	defer func() {
		os.Remove(f.Name())
	}()

	cfg := NewConfig(f.Name(), 1, false)
	counters := NewConcurrentMap[int64](1)
	counters.Set("c0", 64)
	gauges := NewConcurrentMap[float64](1)
	gauges.Set("g0", 64.0)

	syncer := NewFileSyncer(cfg, counters, gauges, nil)
	syncer.Start(context.Background())

	time.Sleep(2 * time.Second) // to wait syncer with delay=1
	decoder := json.NewDecoder(f)
	metrics := make([]model.Metrics, 2)
	err = decoder.Decode(&metrics)
	require.NoError(t, err)

	assert.Equal(t, 2, len(metrics))
	assert.Equal(t, "c0", metrics[0].ID)
	assert.Equal(t, "g0", metrics[1].ID)
}

func TestFileSyncer_SyncNoDelay(t *testing.T) {
	f, err := os.CreateTemp("", "file_syncer_load_test")
	require.NoError(t, err)
	defer func() {
		os.Remove(f.Name())
	}()

	cfg := NewConfig(f.Name(), 0, false)
	counters := NewConcurrentMap[int64](1)
	counters.Set("c0", 64)
	gauges := NewConcurrentMap[float64](1)
	gauges.Set("g0", 64.0)
	syncCh := make(chan int)
	defer close(syncCh)

	syncer := NewFileSyncer(cfg, counters, gauges, syncCh)
	syncer.Start(context.Background())

	syncCh <- 1

	time.Sleep(100 * time.Millisecond)
	decoder := json.NewDecoder(f)
	metrics := make([]model.Metrics, 2)
	err = decoder.Decode(&metrics)
	require.NoError(t, err)

	assert.Equal(t, 2, len(metrics))
	assert.Equal(t, "c0", metrics[0].ID)
	assert.Equal(t, "g0", metrics[1].ID)
}

func TestFileSyncer_Load(t *testing.T) {
	filepath := "../../../test_data/metrics.json"
	cfg := NewConfig(filepath, 5, true)
	counters := NewConcurrentMap[int64](1)
	gauges := NewConcurrentMap[float64](1)

	syncer := NewFileSyncer(cfg, counters, gauges, nil)
	syncer.Start(context.Background())

	c, exists := counters.Get("C1")
	assert.True(t, exists)
	assert.Equal(t, int64(800), c)

	g, exists := gauges.Get("G2")
	assert.True(t, exists)
	assert.Equal(t, float64(127.452), g)
}
