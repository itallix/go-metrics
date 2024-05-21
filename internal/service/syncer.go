package service

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
)

type Saver interface {
	Save()
}

type Loader interface {
	Load()
}

type Syncer struct {
	metricSrv MetricService
	interval  int
	filepath  string
	ctx       context.Context
}

func NewSyncer(ctx context.Context, metricSrv MetricService, interval int, filepath string) *Syncer {
	return &Syncer{
		ctx:       ctx,
		metricSrv: metricSrv,
		interval:  interval,
		filepath:  filepath,
	}
}

func (s *Syncer) Save() {
	if s.filepath != "" {
		tickerStore := time.NewTicker(time.Duration(s.interval) * time.Second)
		defer tickerStore.Stop()

		for {
			select {
			case <-tickerStore.C:
				logger.Log().Infof("Saving metrics to file %s", s.filepath)
				file, err := os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					logger.Log().Errorf("Cannot create or open file: %v", file)
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
					logger.Log().Errorf("Issue encoding metric: %v", err)
				}
			case <-s.ctx.Done():
				return
			}
		}
	} else {
		logger.Log().Info("Filepath is empty, cannot save metrics")
	}
}

func (s *Syncer) Load() {
	logger.Log().Infof("Loading metrics from file %s", s.filepath)
	file, err := os.OpenFile(s.filepath, os.O_RDONLY, 0666)
	if err != nil {
		logger.Log().Errorf("Cannot open file: %v", file)
	}
	decoder := json.NewDecoder(file)
	var metrics []model.Metrics
	if err = decoder.Decode(&metrics); err != nil {
		logger.Log().Infof("Issue decoding metrics: %v", err)
	}
	s.metricSrv.Write(metrics)
}
