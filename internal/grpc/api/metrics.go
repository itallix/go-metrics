package api

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	pb "github.com/itallix/go-metrics/internal/grpc/proto"
	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/service"
	"github.com/itallix/go-metrics/internal/storage"
)

type Server struct {
	pb.UnimplementedMetricsServer

	metricsStorage storage.Storage
	hashService    service.HashService
}

func NewServer(metricsStorage storage.Storage, hashService service.HashService) *Server {
	return &Server{
		metricsStorage: metricsStorage,
		hashService:    hashService,
	}
}

func (srv *Server) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var (
		batch    []model.Metrics
		response pb.UpdateMetricsResponse
	)

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		hash := md.Get(model.HashSha256Header)
		if len(hash) > 0 && srv.hashService != nil {
			reqBytes, err := proto.Marshal(in)
			if err != nil {
				return nil, fmt.Errorf("cannot marshall req to bytes: %w", err)
			}
			if !srv.hashService.Matches(reqBytes, hash[0]) {
				return nil, errors.New("hash doesn't match")
			}
		}
	}

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
