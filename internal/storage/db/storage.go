package db

import (
	"context"
	"fmt"
	"time"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const TimeoutInSeconds = 3

// var retryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type PgStorage struct {
	pool *pgxpool.Pool
}

func NewPgStorage(ctx context.Context, dsn string) (*PgStorage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	if err = createTablesIfNeeded(ctx, pool); err != nil {
		return nil, err
	}

	return &PgStorage{
		pool: pool,
	}, nil
}

func createTablesIfNeeded(ctx context.Context, pool *pgxpool.Pool) error {
	c, cancel := context.WithTimeout(ctx, TimeoutInSeconds*time.Second)
	defer cancel()

	tx, err := pool.Begin(c)
	defer func() {
		if err != nil {
			_ = tx.Rollback(c)
		}
	}()
	if err != nil {
		return err
	}

	status, err := tx.Exec(c, `CREATE TABLE IF NOT EXISTS gauges(id text primary key, val double precision)`)
	if err != nil {
		return err
	}
	logger.Log().Infof("Table 'gauges' has been successfully created: %s", status)

	status, err = tx.Exec(c, `CREATE TABLE IF NOT EXISTS counters(id text primary key, delta integer)`)
	if err != nil {
		return err
	}
	logger.Log().Infof("Table 'counters' has been successfully created: %s", status)

	if err = tx.Commit(c); err != nil {
		return err
	}
	return nil
}

func (m *PgStorage) Update(ctx context.Context, metric *model.Metrics) error {
	c, cancel := context.WithTimeout(ctx, TimeoutInSeconds*time.Second)
	defer cancel()
	switch metric.MType {
	case model.Counter:
		if metric.Delta == nil {
			return storage.ErrMetricNotSupported
		}
		query := `INSERT INTO counters(id, delta) VALUES($1, $2)
		ON CONFLICT(id)
		DO UPDATE SET delta = counters.delta + EXCLUDED.delta RETURNING delta`
		var newDelta int64
		if err := m.pool.QueryRow(c, query, metric.ID, *metric.Delta).Scan(&newDelta); err != nil {
			return err
		}
		metric.Delta = &newDelta

	case model.Gauge:
		query := `INSERT INTO gauges(id, val) VALUES($1, $2)
		ON CONFLICT(id)
		DO UPDATE SET val = EXCLUDED.val RETURNING val`
		var newVal float64
		if err := m.pool.QueryRow(c, query, metric.ID, *metric.Value).Scan(&newVal); err != nil {
			return err
		}
		metric.Value = &newVal

	default:
		return storage.ErrMetricNotFound
	}
	return nil
}

func (m *PgStorage) UpdateBatch(ctx context.Context, metrics []model.Metrics) error {
	c, cancel := context.WithTimeout(ctx, TimeoutInSeconds*time.Second)
	defer cancel()

	batch := &pgx.Batch{}
	queryCounter := `INSERT INTO counters(id, delta) VALUES($1, $2)
		ON CONFLICT(id)
		DO UPDATE SET delta = counters.delta + EXCLUDED.delta`
	queryGauge := `INSERT INTO gauges(id, val) VALUES($1, $2)
		ON CONFLICT(id)
		DO UPDATE SET val = EXCLUDED.val`

	for _, m := range metrics {
		switch m.MType {
		case model.Counter:
			batch.Queue(queryCounter, m.ID, m.Delta)
		case model.Gauge:
			batch.Queue(queryGauge, m.ID, m.Value)
		}
	}

	br := m.pool.SendBatch(c, batch)

	tag, err := br.Exec()
	if err != nil {
		return err
	}
	defer func() {
		_ = br.Close()
	}()
	logger.Log().Infof("Metrics has been succesfully written to DB: %s", tag)
	return nil
}

func (m *PgStorage) Read(ctx context.Context, metric *model.Metrics) error {
	c, cancel := context.WithTimeout(ctx, TimeoutInSeconds*time.Second)
	defer cancel()

	switch metric.MType {
	case model.Counter:
		if err := m.pool.QueryRow(c, "SELECT delta FROM counters WHERE id = $1", metric.ID).Scan(&metric.Delta); err != nil {
			return err
		}
		return nil
	case model.Gauge:
		if err := m.pool.QueryRow(c, "SELECT val FROM gauges WHERE id = $1", metric.ID).Scan(&metric.Value); err != nil {
			return err
		}
		return nil
	default:
		return storage.ErrMetricNotFound
	}
}

func (m *PgStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	c, cancel := context.WithTimeout(ctx, TimeoutInSeconds*time.Second)
	defer cancel()

	counters := make(map[string]int64)
	rows, err := m.pool.Query(c, "SELECT * FROM counters")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id    string
			delta int64
		)
		if err = rows.Scan(&id, &delta); err != nil {
			return nil, err
		}
		counters[id] = delta
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return counters, nil
}

func (m *PgStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	c, cancel := context.WithTimeout(ctx, TimeoutInSeconds*time.Second)
	defer cancel()

	gauges := make(map[string]float64)
	rows, err := m.pool.Query(c, "SELECT * FROM gauges")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id  string
			val float64
		)
		if err = rows.Scan(&id, &val); err != nil {
			return nil, err
		}
		gauges[id] = val
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return gauges, nil
}

func (m *PgStorage) Ping(ctx context.Context) bool {
	if err := m.pool.Ping(ctx); err != nil {
		return false
	}
	return true
}

func (m *PgStorage) Close() {
	m.pool.Close()
}
