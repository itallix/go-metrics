package api

import (
	"context"

	pb "github.com/itallix/go-metrics/internal/grpc/proto"
	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/storage"
)

type Server struct {
	pb.UnimplementedMetricsServer

	metricsStorage storage.Storage
}

func NewServer(metricsStorage storage.Storage) *Server {
	return &Server{
		metricsStorage: metricsStorage,
	}
}

func (srv *Server) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var (
		batch    []model.Metrics
		response pb.UpdateMetricsResponse
	)

	for _, metric := range in.Metrics {
		switch metric.GetMtype() {
		case pb.Metric_M_TYPE_COUNTER:
			batch = append(batch, *model.NewCounter(metric.Id, metric.Delta))
		case pb.Metric_M_TYPE_GAUGE:
			batch = append(batch, *model.NewGauge(metric.Id, metric.Value))
		}
	}

	if err := srv.metricsStorage.UpdateBatch(ctx, batch); err != nil {
		return nil, err
	}

	logger.Log().Info("Successfully saved metrics.")

	response.Metrics = in.GetMetrics()

	return &response, nil
}
