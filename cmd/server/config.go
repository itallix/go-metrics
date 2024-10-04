package main

import (
	"flag"

	"github.com/caarlos0/env"

	"github.com/itallix/go-metrics/internal/model"
)

const (
	defaultAddress       = "localhost:8080"
	defaultStoreInterval = 300
	defaultFilePath      = "/tmp/metrics-db.json"
	defaultRestore       = true
)

func parseConfig() (*model.ServerConfig, error) {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&configPath, "c", "", "Path to config file (shorthand)")

	addr := flag.String("a", defaultAddress, "Net address host:port")
	storeInterval := flag.Int("i", defaultStoreInterval, "Store interval in seconds")
	filePath := flag.String("f", defaultFilePath, "Filepath where metrics will be saved")
	restore := flag.Bool("r", defaultRestore, "Whether server needs to restore metrics from file or not")
	dsn := flag.String("d", "", "Database connection string")
	key := flag.String("k", "", "Key that will be used to calculate hash")
	cryptoKey := flag.String("crypto-key", "", "Private key that will be used to decrypt the request payload")
	trustedSubnet := flag.String("t", "", "Whitelisted subnet in CIDR format")
	flag.Parse()

	cfg := model.ServerConfig{
		Address:       defaultAddress,
		StoreInterval: defaultStoreInterval,
		FilePath:      defaultFilePath,
		Restore:       defaultRestore,
	}
	if configPath != "" {
		err := model.ParseFileConfig(configPath, &cfg)
		if err != nil {
			return nil, err
		}
	}
	if *addr != defaultAddress {
		cfg.Address = *addr
	}
	if *storeInterval != defaultStoreInterval {
		cfg.StoreInterval = *storeInterval
	}
	if *filePath != defaultFilePath {
		cfg.FilePath = *filePath
	}
	if !*restore {
		cfg.Restore = *restore
	}
	if *dsn != "" {
		cfg.DatabaseDSN = *dsn
	}
	if *key != "" {
		cfg.Key = *key
	}
	if *cryptoKey != "" {
		cfg.CryptoKey = *cryptoKey
	}
	if *trustedSubnet != "" {
		cfg.TrustedSubnet = *trustedSubnet
	}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
