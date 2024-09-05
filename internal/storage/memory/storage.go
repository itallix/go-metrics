package memory

import (
	"context"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/storage"
)

const (
	DefaultCounterCapacity = 1
	DefaultGaugeCapacity   = 32
)

type MemStorage struct {
	counters *ConcurrentMap[int64]
	gauges   *ConcurrentMap[float64]
	syncCh   chan int
}

func NewMemStorage(ctx context.Context, config *Config) *MemStorage {
	var syncCh chan int
	counters := NewConcurrentMap[int64](DefaultCounterCapacity)
	gauges := NewConcurrentMap[float64](DefaultGaugeCapacity)

	if config != nil {
		if config.filepath == "" {
			logger.Log().Info("Filepath is not defined. Server will proceed in memory mode.")
		} else {
			if config.interval == 0 {
				syncCh = make(chan int)
			}
			syncer := NewFileSyncer(config, counters, gauges, syncCh)
			syncer.Start(ctx)
		}
	}
	return &MemStorage{
		counters: counters,
		gauges:   gauges,
		syncCh:   syncCh,
	}
}

func (m *MemStorage) Update(_ context.Context, metric *model.Metrics) error {
	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return storage.ErrMetricNotSupported
		}
		val := m.counters.Inc(metric.ID, *metric.Delta)
		metric.Delta = &val

	case model.Gauge:
		if metric.Value == nil {
			return storage.ErrMetricNotSupported
		}
		val := m.gauges.Set(metric.ID, *metric.Value)
		metric.Value = &val

	default:
		return storage.ErrMetricNotFound
	}
	if m.syncCh != nil {
		m.syncCh <- 1
	}
	return nil
}

func (m *MemStorage) UpdateBatch(_ context.Context, metrics []model.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case model.Counter:
			if metric.Delta == nil {
				return storage.ErrMetricNotSupported
			}
			val := m.counters.Inc(metric.ID, *metric.Delta)
			metric.Delta = &val

		case model.Gauge:
			if metric.Value == nil {
				return storage.ErrMetricNotSupported
			}
			val := m.gauges.Set(metric.ID, *metric.Value)
			metric.Value = &val

		default:
			return storage.ErrMetricNotFound
		}
	}
	if m.syncCh != nil {
		m.syncCh <- 1
	}
	return nil
}

func (m *MemStorage) Read(_ context.Context, metric *model.Metrics) error {
	switch metric.MType {
	case model.Counter:
		val, ok := m.counters.Get(metric.ID)
		if !ok {
			return storage.ErrMetricNotFound
		}
		metric.Delta = &val
		return nil
	case model.Gauge:
		val, ok := m.gauges.Get(metric.ID)
		if !ok {
			return storage.ErrMetricNotFound
		}
		metric.Value = &val
		return nil
	default:
		return storage.ErrMetricNotFound
	}
}

func (m *MemStorage) GetCounters(_ context.Context) (map[string]int64, error) {
	return m.counters.Copy(), nil
}

func (m *MemStorage) GetGauges(_ context.Context) (map[string]float64, error) {
	return m.gauges.Copy(), nil
}

func (m *MemStorage) Ping(_ context.Context) bool {
	return false
}

func (m *MemStorage) Close() {
	if m.syncCh != nil {
		close(m.syncCh)
	}
}
