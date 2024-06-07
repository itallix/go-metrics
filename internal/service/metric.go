package service

import (
	"errors"

	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/storage"
)

type MetricService interface {
	UpdateOne(metric *model.Metrics) error
	UpdateBatch(metrics []model.Metrics) error
	Read(metric *model.Metrics) error
	Write(metrics []model.Metrics)
	GetCounters() map[string]int64
	GetGauges() map[string]float64
	GetMetrics() []model.Metrics
}

type MetricServiceImpl struct {
	counters storage.Storage[int64]
	gauges   storage.Storage[float64]
	syncCh   chan int
}

func NewMetricServiceImpl(counters storage.Storage[int64], gauges storage.Storage[float64],
	syncCh chan int) *MetricServiceImpl {
	return &MetricServiceImpl{
		counters: counters,
		gauges:   gauges,
		syncCh:   syncCh,
	}
}

var (
	errMetricNotSupported = errors.New("metric type is not supported")
	errMetricNotFound     = errors.New("metric is not found")
)

func (s *MetricServiceImpl) UpdateOne(metric *model.Metrics) error {
	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return errMetricNotSupported
		}
		val := s.counters.Update(metric.ID, *metric.Delta)
		metric.Delta = &val

	case model.Gauge:
		if metric.Value == nil {
			return errMetricNotSupported
		}
		val := s.gauges.Set(metric.ID, *metric.Value)
		metric.Value = &val
	default:
		return errMetricNotFound
	}
	if s.syncCh != nil {
		s.syncCh <- 1
	}
	return nil
}

func (s *MetricServiceImpl) UpdateBatch(metrics []model.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case model.Counter:
			if metric.Delta == nil {
				return errMetricNotSupported
			}
			val := s.counters.Update(metric.ID, *metric.Delta)
			metric.Delta = &val

		case model.Gauge:
			if metric.Value == nil {
				return errMetricNotSupported
			}
			val := s.gauges.Set(metric.ID, *metric.Value)
			metric.Value = &val
		default:
			return errMetricNotFound
		}
	}
	if s.syncCh != nil {
		s.syncCh <- 1
	}
	return nil
}

func (s *MetricServiceImpl) Read(metric *model.Metrics) error {
	switch metric.MType {
	case model.Counter:
		val, ok := s.counters.Get(metric.ID)
		if !ok {
			return errMetricNotFound
		}
		metric.Delta = &val
		return nil
	case model.Gauge:
		val, ok := s.gauges.Get(metric.ID)
		if !ok {
			return errMetricNotFound
		}
		metric.Value = &val
		return nil
	default:
		return errMetricNotFound
	}
}

func (s *MetricServiceImpl) Write(metrics []model.Metrics) {
	for _, m := range metrics {
		switch m.MType {
		case model.Counter:
			s.counters.Set(m.ID, *m.Delta)
		case model.Gauge:
			s.gauges.Set(m.ID, *m.Value)
		}
	}
}

func (s *MetricServiceImpl) GetCounters() map[string]int64 {
	return s.counters.Copy()
}

func (s *MetricServiceImpl) GetGauges() map[string]float64 {
	return s.gauges.Copy()
}

func (s *MetricServiceImpl) GetMetrics() []model.Metrics {
	var metrics []model.Metrics
	for k, v := range s.GetCounters() {
		cv := v
		c := model.NewCounter(k, &cv)
		metrics = append(metrics, *c)
	}
	for k, v := range s.GetGauges() {
		gv := v
		g := model.NewGauge(k, &gv)
		metrics = append(metrics, *g)
	}
	return metrics
}
