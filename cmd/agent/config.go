package main

import (
	"flag"

	"github.com/caarlos0/env"

	"github.com/itallix/go-metrics/internal/model"
)

const (
	defaultServerURL      = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultRateLimit      = 3
)

func parseConfig() (*model.AgentConfig, error) {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&configPath, "c", "", "Path to config file (shorthand)")

	serverURL := flag.String("a", defaultServerURL, "Net address host:port")
	pollInterval := flag.Int("p", defaultPollInterval, "Poll interval in seconds")
	reportInterval := flag.Int("r", defaultReportInterval, "Report interval in seconds")
	key := flag.String("k", "", "Key that will be used to calculate hash")
	rateLimit := flag.Int("l", defaultRateLimit, "Max number of concurrent requests to the server")
	cryptoKey := flag.String("crypto-key", "", "Path to public key that will be used for payload encryption")
	flag.Parse()

	cfg := model.AgentConfig{
		ServerURL:      defaultServerURL,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		RateLimit:      defaultRateLimit,
	}
	if configPath != "" {
		err := model.ParseFileConfig(configPath, &cfg)
		if err != nil {
			return nil, err
		}
	}
	if *serverURL != defaultServerURL {
		cfg.ServerURL = *serverURL
	}
	if *pollInterval != defaultPollInterval {
		cfg.PollInterval = *pollInterval
	}
	if *reportInterval != defaultReportInterval {
		cfg.ReportInterval = *reportInterval
	}
	if *rateLimit != defaultRateLimit {
		cfg.RateLimit = *rateLimit
	}
	if *key != "" {
		cfg.Key = *key
	}
	if *cryptoKey != "" {
		cfg.CryptoKey = *cryptoKey
	}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
