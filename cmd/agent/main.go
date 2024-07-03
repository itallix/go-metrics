package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/go-resty/resty/v2"

	"github.com/itallix/go-metrics/internal/model"

	"github.com/itallix/go-metrics/internal/logger"
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

	serverURL, config, err := parseFlags()
	if err != nil {
		logger.Log().Fatalf("Cannot parse flags: %v", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan []model.Metrics, config.RateLimit)
	results := make(chan error, config.RateLimit)

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	defer tickerPoll.Stop()
	reportPoll := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	defer reportPoll.Stop()

	client := resty.New().SetBaseURL("http://"+serverURL.String()).
		SetHeader("Content-Type", "application/json")
	metricsAgent := newAgent(client, config.Key)

	for i := 0; i < config.RateLimit; i++ {
		go metricsAgent.send(ctx, jobs, results)
	}

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
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log().Info("Shutting down agent gracefully...")
	close(jobs)
	close(results)
	cancel()
}
