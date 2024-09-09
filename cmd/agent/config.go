package main

import (
	"flag"

	"github.com/caarlos0/env"

	"github.com/itallix/go-metrics/internal/model"
)

func parseConfig() (*model.AgentConfig, error) {
	var cfg model.AgentConfig
	flag.StringVar(&cfg.ConfigPath, "config", "", "Path to config file")
	flag.StringVar(&cfg.ConfigPath, "c", "", "Path to config file (shorthand)")
	flag.StringVar(&cfg.ServerURL, "a", "localhost:8080", "Net address host:port")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "Key that will be used to calculate hash")
	flag.IntVar(&cfg.RateLimit, "l", 3, "Max number of concurrent requests to the server")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to public key that will be used for payload encryption")
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
