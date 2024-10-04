package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/service"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}
	defer func() {
		if deferErr := logger.Log().Sync(); deferErr != nil {
			logger.Log().Errorf("Failed to sync logger: %s", deferErr)
		}
	}()

	service.PrintBuildInfo(buildVersion, buildDate, buildCommit, os.Stdout)

	config, err := parseConfig()
	if err != nil {
		logger.Log().Fatalf("Cannot parse flags: %v", err)
	}

	mainCtx := context.Background()
	ctx, cancel := context.WithCancel(mainCtx)

	jobs := make(chan []model.Metrics, config.RateLimit)
	results := make(chan error, config.RateLimit)

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	defer tickerPoll.Stop()
	reportPoll := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	defer reportPoll.Stop()

	var (
		httpClient *resty.Client
		grpcClient *GRPCMetricsClient
	)
	if config.Schema == "http" {
		httpClient = resty.New().SetBaseURL("http://"+config.ServerURL).
			SetHeader("Content-Type", "application/json")
	} else if config.Schema == "grpc" {
		client, err := NewGrpcClient()
		if err != nil {
			logger.Log().Fatalf("Failed to create a new grpc client: %v", err)
		}
		grpcClient = client
	}
	metricsAgent, err := newAgent(httpClient, grpcClient, config.Key, config.CryptoKey)
	if err != nil {
		logger.Log().Fatalf("Failed to instantiate agent: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(config.RateLimit)

	for i := 0; i < config.RateLimit; i++ {
		go metricsAgent.send(mainCtx, &wg, jobs, results)
	}

	// this goroutine monitors results from workers and logs the status of the job
	go func() {
		for err := range results {
			if err != nil {
				logger.Log().Errorf("Error sending metrics %v", err)
			} else {
				logger.Log().Info("Metrics has been successfully sent")
			}
		}
	}()

	go func() {
		for {
			select {
			case <-tickerPoll.C:
				g := new(errgroup.Group)
				g.Go(metricsAgent.collectRuntime)
				g.Go(metricsAgent.collectExtra)

				if err = g.Wait(); err != nil {
					logger.Log().Errorf("Issue collecting metrics: %v", err)
				}
				logger.Log().Infof("Collected metrics: %v", metricsAgent.metrics())
			case <-reportPoll.C:
				logger.Log().Info("Scheduling new job to send metrics...")
				jobs <- metricsAgent.metrics()
			case <-ctx.Done():
				return
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit

	logger.Log().Info("Shutting down agent gracefully...")
	cancel()
	// if there are unsent metrics in agent, schedule the last job with collected metrics
	if metricsAgent.Counter > 0 {
		logger.Log().Info("Scheduling new job to send metrics due to graceful shutdown...")
		jobs <- metricsAgent.metrics()
	}
	close(jobs)
	// waiting for all workers that send metrics to the server to finish
	wg.Wait()
	close(results)
	if grpcClient != nil {
		grpcClient.Close()
	}
}
