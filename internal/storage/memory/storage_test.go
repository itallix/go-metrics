package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/itallix/go-metrics/internal/model"
)

func TestStorage_Update(t *testing.T) {
	ctx := context.Background()
	s := NewMemStorage(ctx, nil, nil)
	c, g := int64(64), float64(64)
	err := s.Update(ctx, model.NewCounter("c0", &c))
	require.NoError(t, err)
	err = s.Update(ctx, model.NewGauge("g0", &g))
	require.NoError(t, err)

	cc, err := s.GetCounters(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(64), cc["c0"])
	gg, err := s.GetGauges(ctx)
	require.NoError(t, err)
	assert.Equal(t, float64(64), gg["g0"])
}

func TestStorage_UpdateBatch(t *testing.T) {
	ctx := context.Background()
	var wg sync.WaitGroup
	s := NewMemStorage(ctx, &wg, nil)
	c, g := int64(64), float64(64)
	metrics := []model.Metrics{*model.NewCounter("c0", &c), *model.NewGauge("g0", &g)}
	err := s.UpdateBatch(ctx, metrics)
	require.NoError(t, err)

	cc, err := s.GetCounters(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(64), cc["c0"])
	gg, err := s.GetGauges(ctx)
	require.NoError(t, err)
	assert.Equal(t, float64(64), gg["g0"])
}

func TestStorage_Read(t *testing.T) {
	ctx := context.Background()
	s := NewMemStorage(ctx, nil, nil)
	c, g := int64(64), float64(64)
	metrics := []model.Metrics{*model.NewCounter("c0", &c), *model.NewGauge("g0", &g)}
	err := s.UpdateBatch(ctx, metrics)
	require.NoError(t, err)

	read := model.Metrics{
		ID:    "c0",
		MType: model.Counter,
	}
	err = s.Read(ctx, &read)

	require.NoError(t, err)
	assert.Equal(t, int64(64), *read.Delta)
	assert.Nil(t, read.Value)
	assert.Equal(t, "c0", read.ID)
	assert.Equal(t, model.Counter, read.MType)

	read = model.Metrics{
		ID:    "g0",
		MType: model.Gauge,
	}
	err = s.Read(ctx, &read)

	require.NoError(t, err)
	assert.Equal(t, float64(64), *read.Value)
	assert.Nil(t, read.Delta)
	assert.Equal(t, "g0", read.ID)
	assert.Equal(t, model.Gauge, read.MType)
}

func TestStorage_Ping(t *testing.T) {
	ctx := context.Background()
	s := NewMemStorage(ctx, nil, nil)

	assert.False(t, s.Ping(ctx))
}

func TestStorage_New(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		interval: 0,
		filepath: "some/path",
	}
	var wg sync.WaitGroup
	s := NewMemStorage(ctx, &wg, &cfg)

	assert.NotNil(t, s.syncCh)
}
