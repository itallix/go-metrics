package sync

import (
	"context"
	"time"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/service"
	"github.com/itallix/go-metrics/internal/storage"
)

type DBSyncer struct {
	db        storage.DB
	metricSrv service.MetricService
}

func NewDBSyncer(metricSrv service.MetricService, db storage.DB) *DBSyncer {
	return &DBSyncer{
		db:        db,
		metricSrv: metricSrv,
	}
}

func (s *DBSyncer) sync(ctx context.Context) error {
	logger.Log().Info("Saving metrics to db...")
	metrics := s.metricSrv.GetMetrics()
	if err := s.db.WriteMetrics(ctx, metrics); err != nil {
		return err
	}
	return nil
}

func (s *DBSyncer) Start(ctx context.Context, cfg *Config) {
	if err := s.db.CreateTablesIfNeeded(ctx); err != nil {
		logger.Log().Errorf("Cannot create tables in db: %v. Server will fallback to file or in-memory sync.", err)
		return
	}
	if cfg.restore {
		if err := s.load(ctx); err != nil {
			logger.Log().Errorf("Error loading metrics from db: %v", err)
		}
	}
	if cfg.interval == 0 {
		go func() {
			for range cfg.syncCh {
				if err := s.sync(ctx); err != nil {
					logger.Log().Errorf("Error syncing to the db table: %v", err)
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
					if err := s.sync(ctx); err != nil {
						logger.Log().Errorf("Error syncing to the db table: %v", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (s *DBSyncer) load(ctx context.Context) error {
	logger.Log().Info("Loading metrics from db...")
	metrics, err := s.db.LoadMetrics(ctx)
	if err != nil {
		return err
	}
	s.metricSrv.Write(metrics)
	logger.Log().Info("Metrics has been successfully loaded from db.")
	return nil
}
