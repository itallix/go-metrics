package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"

	mflag "github.com/itallix/go-metrics/internal/flag"
)

const (
	EnvAddress = "ADDRESS"
)

// ServerConfig describes customization settings for the server.
type ServerConfig struct {
	StoreInterval int    `env:"STORE_INTERVAL"`    // How often to store metrics in the file.
	FilePath      string `env:"FILE_STORAGE_PATH"` // Location where to store metrics.
	Restore       bool   `env:"RESTORE"`           // Should the metrics be loaded from the file on start?
	DatabaseDSN   string `env:"DATABASE_DSN"`      // DB connection string, example: postgresql://username:password@hostname:port/database_name.
	Key           string `env:"KEY"`               // Secret for a hash function.
}

func parseFlags() (*mflag.RunAddress, *ServerConfig, error) {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")

	var cfg ServerConfig
	flag.IntVar(&cfg.StoreInterval, "i", 300, "Store interval in seconds")
	flag.StringVar(&cfg.FilePath, "f", "/tmp/metrics-db.json", "Filepath where metrics will be saved")
	flag.BoolVar(&cfg.Restore, "r", true, "Whether server needs to restore metrics from file or not")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&cfg.Key, "k", "", "Key that will be used to calculate hash")
	flag.Parse()

	if envAddr := os.Getenv(EnvAddress); envAddr != "" {
		if err := addr.Set(envAddr); err != nil {
			return nil, nil, fmt.Errorf("cannot parse ADDRESS env var: %w", err)
		}
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, nil, err
	}
	return addr, &cfg, nil
}
