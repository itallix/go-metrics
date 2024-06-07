package sync

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/service"
)

type FileSyncer struct {
	filepath  string
	metricSrv service.MetricService
}

func NewFileSyncer(metricSrv service.MetricService, filepath string) *FileSyncer {
	return &FileSyncer{
		filepath:  filepath,
		metricSrv: metricSrv,
	}
}

func (s *FileSyncer) sync() error {
	logger.Log().Infof("Saving metrics to file %s", s.filepath)
	file, err := os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	metrics := s.metricSrv.GetMetrics()
	if err = encoder.Encode(metrics); err != nil {
		return err
	}
	return nil
}

func (s *FileSyncer) Start(ctx context.Context, cfg *Config) {
	if cfg.restore {
		if err := s.load(); err != nil {
			logger.Log().Errorf("Error loading metrics from file: %v", err)
		}
	}
	if cfg.interval == 0 {
		go func() {
			for range cfg.syncCh {
				if err := s.sync(); err != nil {
					logger.Log().Errorf("Error syncing to the file: %v", err)
				}
			}
		}()
	} else {
		go func() {
			tickerStore := time.NewTicker(time.Duration(cfg.interval) * time.Second)
			defer tickerStore.Stop()
			for {
				select {
				case <-tickerStore.C:
					if err := s.sync(); err != nil {
						logger.Log().Errorf("Error syncing to the file: %v", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (s *FileSyncer) load() error {
	logger.Log().Infof("Loading metrics from file %s...", s.filepath)
	file, err := os.OpenFile(s.filepath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	var metrics []model.Metrics
	if err = decoder.Decode(&metrics); err != nil {
		return err
	}
	s.metricSrv.Write(metrics)
	logger.Log().Infof("Metrics has been successfully loaded from file %s.", s.filepath)
	return nil
}
