package main

import (
	"fmt"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/itallix/go-metrics/internal/grpc/proto"
	"github.com/itallix/go-metrics/internal/model"
)

type GRPCMetricsClient struct {
	pb.MetricsClient
	conn *grpc.ClientConn
}

func NewGrpcClient() (*GRPCMetricsClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			retry.UnaryClientInterceptor(
				retry.WithMax(3),
				retry.WithBackoff(retry.BackoffExponential(1*time.Second)),
				retry.WithCodes(codes.Unavailable, codes.Internal),
			),
		),
	}
	conn, err := grpc.NewClient(":"+model.GRPCPort, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new grpc client: %w", err)
	}

	client := pb.NewMetricsClient(conn)

	return &GRPCMetricsClient{
		conn:          conn,
		MetricsClient: client,
	}, nil
}

func (c *GRPCMetricsClient) Close() {
	c.conn.Close()
}
