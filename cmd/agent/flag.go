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

// AgentConfig describes customization settings for the agent.
type AgentConfig struct {
	PollInterval   int    `env:"POLL_INTERVAL"`   // How often to collect metrics from the system.
	ReportInterval int    `env:"REPORT_INTERVAL"` // How often to report collected metrics to the server.
	Key            string `env:"KEY"`             // Secret for a hash function.
	RateLimit      int    `env:"RATE_LIMIT"`      // Limits number of concurrent requests to the server.
	CryptoKey      string `env:"CRYPTO_KEY"`      // Public key used to encrypt request payload.
}

func parseFlags() (*mflag.RunAddress, *AgentConfig, error) {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")

	var cfg AgentConfig
	flag.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "Key that will be used to calculate hash")
	flag.IntVar(&cfg.RateLimit, "l", 3, "Max number of concurrent requests to the server")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Path to public key that will be used for payload encryption")
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
