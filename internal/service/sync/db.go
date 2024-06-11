package sync

import (
	"context"
	"errors"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"time"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/service"
	"github.com/itallix/go-metrics/internal/storage"
)

type DBSyncer struct {
	db          storage.DB
	metricSrv   service.MetricService
	retryDelays []time.Duration
}

func NewDBSyncer(metricSrv service.MetricService, db storage.DB) *DBSyncer {
	return &DBSyncer{
		db:          db,
		metricSrv:   metricSrv,
		retryDelays: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

func (s *DBSyncer) sync(ctx context.Context) error {
	logger.Log().Info("Saving metrics to db...")
	metrics := s.metricSrv.GetMetrics()

	var err error
	for _, delay := range s.retryDelays {
		err = s.db.WriteMetrics(ctx, metrics)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr); pgerrcode.IsConnectionException(pgErr.Code) {
				logger.Log().Errorf("Failed to connect to DB, retrying after %v...", delay)
				time.Sleep(delay)
				continue
			}
		}
		break
	}

	if err != nil {
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

	var (
		err     error
		metrics []model.Metrics
	)
	for _, delay := range s.retryDelays {
		metrics, err = s.db.LoadMetrics(ctx)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr); pgerrcode.IsConnectionException(pgErr.Code) {
				logger.Log().Errorf("Failed to connect to DB, retrying after %v...", delay)
				time.Sleep(delay)
				continue
			}
		}
		break
	}

	if err != nil {
		return err
	}

	s.metricSrv.Write(metrics)
	logger.Log().Info("Metrics has been successfully loaded from db.")
	return nil
}
