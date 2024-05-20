package service

import (
	"errors"

	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/storage"
)

type MetricService interface {
	Update(metric *model.Metrics) error
	Read(metric *model.Metrics) error
	GetCounters() map[string]int64
	GetGauges() map[string]float64
}

type MetricServiceImpl struct {
	counters storage.Storage[int64]
	gauges   storage.Storage[float64]
}

func NewMetricService(counters storage.Storage[int64], gauges storage.Storage[float64]) *MetricServiceImpl {
	return &MetricServiceImpl{
		counters: counters,
		gauges:   gauges,
	}
}

var (
	errMetricNotSupported = errors.New("metric type is not supported")
	errMetricNotFound     = errors.New("metric is not found")
)

func (s *MetricServiceImpl) Update(metric *model.Metrics) error {
	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return errMetricNotSupported
		}
		val := s.counters.Update(metric.ID, *metric.Delta)
		metric.Delta = &val
		return nil
	case model.Gauge:
		if metric.Value == nil {
			return errMetricNotSupported
		}
		val := s.gauges.Set(metric.ID, *metric.Value)
		metric.Value = &val
		return nil
	default:
		return errMetricNotFound
	}
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

func (s *MetricServiceImpl) GetCounters() map[string]int64 {
	return s.counters.Copy()
}

func (s *MetricServiceImpl) GetGauges() map[string]float64 {
	return s.gauges.Copy()
}
