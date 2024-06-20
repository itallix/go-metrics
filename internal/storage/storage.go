package storage

import (
	"context"
	"errors"

	"github.com/itallix/go-metrics/internal/model"
)

type Storage interface {
	Update(ctx context.Context, metric *model.Metrics) error
	UpdateBatch(ctx context.Context, metrics []model.Metrics) error
	Read(ctx context.Context, metric *model.Metrics) error
	GetCounters(ctx context.Context) map[string]int64
	GetGauges(ctx context.Context) map[string]float64

	Ping(ctx context.Context) bool
	Close()
}

var (
	ErrMetricNotSupported = errors.New("metric type is not supported")
	ErrMetricNotFound     = errors.New("metric is not found")
)
