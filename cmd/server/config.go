package main

import (
	"flag"

	"github.com/caarlos0/env"
)

// ServerConfig describes customization settings for the server.
type ServerConfig struct {
	Address       string `env:"ADDRESS" json:"address"`               // Address where server will be started.
	StoreInterval int    `env:"STORE_INTERVAL" json:"store_interval"` // How often to store metrics in the file.
	FilePath      string `env:"FILE_STORAGE_PATH" json:"store_file"`  // Location where to store metrics.
	Restore       bool   `env:"RESTORE" json:"restore"`               // Should the metrics be loaded from the file on start?
	DatabaseDSN   string `env:"DATABASE_DSN" json:"database_dsn"`     // DB connection string, example: postgresql://username:password@hostname:port/database_name.
	Key           string `env:"KEY" json:"hash_secret"`               // Secret for a hash function.
	CryptoKey     string `env:"CRYPTO_KEY" json:"crypto_key"`         // Private key used to decrypt request payload.
}

func parseConfig() (*ServerConfig, error) {
	var cfg ServerConfig
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Net address host:port")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "Store interval in seconds")
	flag.StringVar(&cfg.FilePath, "f", "/tmp/metrics-db.json", "Filepath where metrics will be saved")
	flag.BoolVar(&cfg.Restore, "r", true, "Whether server needs to restore metrics from file or not")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&cfg.Key, "k", "", "Key that will be used to calculate hash")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Private key that will be used to decrypt the request payload")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
