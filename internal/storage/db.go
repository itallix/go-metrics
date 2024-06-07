package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DBTimeoutInSeconds = 3

type Config struct {
	DSN string
}

type DB interface {
	Ping(ctx context.Context) error
	Close() error
	CreateTablesIfNeeded(ctx context.Context) error
	WriteMetrics(ctx context.Context, metrics []model.Metrics) error
	LoadMetrics(ctx context.Context) ([]model.Metrics, error)
}

type PgxDB struct {
	pool *pgxpool.Pool
}

func NewPgxDB(ctx context.Context, dsn string) (*PgxDB, error) {
	pool, err := initPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	return &PgxDB{
		pool: pool,
	}, nil
}

func initPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	return pool, nil
}

func (db *PgxDB) Ping(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping the DB: %w", err)
	}
	return nil
}

func (db *PgxDB) Close() error {
	db.pool.Close()
	return nil
}

func (db *PgxDB) CreateTablesIfNeeded(ctx context.Context) error {
	c, cancel := context.WithTimeout(ctx, DBTimeoutInSeconds*time.Second)
	defer cancel()

	tx, err := db.pool.Begin(c)
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

func (db *PgxDB) WriteMetrics(ctx context.Context, metrics []model.Metrics) error {
	c, cancel := context.WithTimeout(ctx, DBTimeoutInSeconds*time.Second)
	defer cancel()

	batch := &pgx.Batch{}
	queryCounter := `INSERT INTO counters(id, delta) VALUES($1, $2)
		ON CONFLICT(id)
		DO UPDATE SET delta = EXCLUDED.delta`
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

	br := db.pool.SendBatch(c, batch)

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

func (db *PgxDB) LoadMetrics(ctx context.Context) ([]model.Metrics, error) {
	c, cancel := context.WithTimeout(ctx, DBTimeoutInSeconds*time.Second)
	defer cancel()

	var (
		metrics []model.Metrics
		metric  model.Metrics
	)
	rows, err := db.pool.Query(c, `SELECT * FROM gauges`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&metric.ID, &metric.Value); err != nil {
			return nil, err
		}
		metric.MType = model.Gauge
		metrics = append(metrics, metric)
	}

	rows, err = db.pool.Query(c, `SELECT * FROM counters`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&metric.ID, &metric.Delta); err != nil {
			return nil, err
		}
		metric.MType = model.Counter
		metrics = append(metrics, metric)
	}

	return metrics, nil
}
