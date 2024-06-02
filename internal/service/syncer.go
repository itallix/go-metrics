package service

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
)

type ISyncer interface {
	Start(restore bool)
}

type Syncer struct {
	metricSrv MetricService
	interval  int
	filepath  string
	ctx       context.Context
	syncCh    chan int
}

func NewSyncer(ctx context.Context, metricSrv MetricService, interval int, filepath string, syncCh chan int) *Syncer {
	return &Syncer{
		ctx:       ctx,
		metricSrv: metricSrv,
		interval:  interval,
		filepath:  filepath,
		syncCh:    syncCh,
	}
}

func (s *Syncer) sync() error {
	logger.Log().Infof("Saving metrics to file %s", s.filepath)
	file, err := os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	counters := s.metricSrv.GetCounters()
	gauges := s.metricSrv.GetGauges()
	var metrics []model.Metrics
	for k, v := range counters {
		cv := v
		c := model.NewCounter(k, &cv)
		metrics = append(metrics, *c)
	}
	for k, v := range gauges {
		gv := v
		g := model.NewGauge(k, &gv)
		metrics = append(metrics, *g)
	}
	if err = encoder.Encode(metrics); err != nil {
		return err
	}
	return nil
}

func (s *Syncer) Start(restore bool) {
	if restore {
		if err := s.load(); err != nil {
			logger.Log().Errorf("Error loading metrics from file: %v", err)
		}
	}
	if s.interval == 0 {
		go func() {
			for range s.syncCh {
				if err := s.sync(); err != nil {
					logger.Log().Errorf("Error syncing to the file: %v", err)
				}
			}
		}()
	} else if s.filepath != "" {
		go func() {
			tickerStore := time.NewTicker(time.Duration(s.interval) * time.Second)
			defer tickerStore.Stop()
			for {
				select {
				case <-tickerStore.C:
					if err := s.sync(); err != nil {
						logger.Log().Errorf("Error syncing to the file: %v", err)
					}
				case <-s.ctx.Done():
					return
				}
			}
		}()
	} else {
		logger.Log().Info("Filepath is empty, cannot save metrics")
	}
}

func (s *Syncer) load() error {
	logger.Log().Infof("Loading metrics from file %s", s.filepath)
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
	return nil
}
