package memory

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
)

// Config defines start parameters for the sync process.
type Config struct {
	filepath string
	interval int
	restore  bool
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

func toMetrics(counters *ConcurrentMap[int64], gauges *ConcurrentMap[float64]) []*model.Metrics {
	metrics := make([]*model.Metrics, counters.Len()+gauges.Len())
	i := 0
	for k, v := range counters.Copy() {
		metrics[i] = model.NewCounter(k, &v)
		i++
	}
	for k, v := range gauges.Copy() {
		metrics[i] = model.NewGauge(k, &v)
		i++
	}
	return metrics
}

func (s *FileSyncer) sync(filepath string) error {
	logger.Log().Infof("Saving metrics to file %s", filepath)
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	metrics := toMetrics(s.counters, s.gauges)
	if err = encoder.Encode(metrics); err != nil {
		return err
	}
	logger.Log().Info("Metrics has been successfully saved.")
	return nil
}

func (s *FileSyncer) Start(ctx context.Context, wg *sync.WaitGroup) {
	if s.config.restore {
		if err := s.load(s.config.filepath); err != nil {
			logger.Log().Errorf("Error loading metrics from file: %v", err)
		}
	}
	handleSync := func(close bool) {
		if close {
			logger.Log().Info("Syncing storage with the filesystem due to graceful shutdown...")
		}
		if err := s.sync(s.config.filepath); err != nil {
			logger.Log().Errorf("Error syncing to the file: %v", err)
		}
	}
	wg.Add(1)
	if s.config.interval == 0 {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-s.syncCh:
					handleSync(false)
				case <-ctx.Done():
					handleSync(true)
					return
				}
			}
		}()
	} else {
		go func() {
			defer wg.Done()
			tickerStore := time.NewTicker(time.Duration(s.config.interval) * time.Second)
			defer tickerStore.Stop()
			for {
				select {
				case <-tickerStore.C:
					handleSync(false)
				case <-ctx.Done():
					handleSync(true)
					return
				}
			}
		}()
	}
}

// load initializes storage with metric values that have been read from file.
func (s *FileSyncer) load(filepath string) error {
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
