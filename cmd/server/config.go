package main

import (
	"flag"

	"github.com/caarlos0/env"

	"github.com/itallix/go-metrics/internal/model"
)

func parseConfig() (*model.ServerConfig, error) {
	var cfg model.ServerConfig
	flag.StringVar(&cfg.ConfigPath, "config", "", "Path to config file")
	flag.StringVar(&cfg.ConfigPath, "c", "", "Path to config file (shorthand)")
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

	if cfg.ConfigPath != "" {
		err := model.ParseFileConfig(cfg.ConfigPath, &cfg)
		if err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}
