package memory

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
)

type Config struct {
	filepath string
	interval int
	restore  bool
}

type Syncer interface {
	Start(ctx context.Context, cfg *Config)
}

func NewConfig(filepath string, interval int, restore bool) *Config {
	return &Config{
		filepath: filepath,
		interval: interval,
		restore:  restore,
	}
}

type FileSyncer struct {
	config   *Config
	counters *ConcurrentMap[int64]
	gauges   *ConcurrentMap[float64]
	syncCh   chan int
}

func NewFileSyncer(config *Config, counters *ConcurrentMap[int64], gauges *ConcurrentMap[float64],
	syncCh chan int) *FileSyncer {
	return &FileSyncer{
		config:   config,
		counters: counters,
		gauges:   gauges,
		syncCh:   syncCh,
	}
}

func ToMetrics(counters *ConcurrentMap[int64], gauges *ConcurrentMap[float64]) []model.Metrics {
	var metrics []model.Metrics
	for k, v := range counters.Copy() {
		cv := v
		c := model.NewCounter(k, &cv)
		metrics = append(metrics, *c)
	}
	for k, v := range gauges.Copy() {
		gv := v
		g := model.NewGauge(k, &gv)
		metrics = append(metrics, *g)
	}
	return metrics
}

func (s *FileSyncer) sync() error {
	filepath := s.config.filepath
	logger.Log().Infof("Saving metrics to file %s", filepath)
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	metrics := ToMetrics(s.counters, s.gauges)
	if err = encoder.Encode(metrics); err != nil {
		return err
	}
	logger.Log().Info("Metrics has been successfully saved.")
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
			for range s.syncCh {
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

// load initializes storage with metric values that have been read from file.
func (s *FileSyncer) load() error {
	filepath := s.config.filepath
	logger.Log().Infof("Loading metrics from file %s...", filepath)
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	var metrics []model.Metrics
	if err = decoder.Decode(&metrics); err != nil {
		return err
	}
	counters := make(map[string]int64)
	gauges := make(map[string]float64)
	for _, m := range metrics {
		switch m.MType {
		case model.Counter:
			counters[m.ID] = *m.Delta
		case model.Gauge:
			gauges[m.ID] = *m.Value
		}
	}
	s.counters.Init(counters)
	s.gauges.Init(gauges)
	logger.Log().Infof("Metrics has been successfully loaded from file %s.", filepath)
	return nil
}
