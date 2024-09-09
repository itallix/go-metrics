package main

import (
	"flag"

	"github.com/caarlos0/env"
)

// AgentConfig describes customization settings for the agent.
type AgentConfig struct {
	ServerURL      string `env:"ADDRESS" json:"address"` // Address of the server to send metrics to.
	PollInterval   int    `env:"POLL_INTERVAL"`          // How often to collect metrics from the system.
	ReportInterval int    `env:"REPORT_INTERVAL"`        // How often to report collected metrics to the server.
	Key            string `env:"KEY"`                    // Secret for a hash function.
	RateLimit      int    `env:"RATE_LIMIT"`             // Limits number of concurrent requests to the server.
	CryptoKey      string `env:"CRYPTO_KEY"`             // Public key used to encrypt request payload.
}

func parseConfig() (*AgentConfig, error) {
	var cfg AgentConfig
	flag.StringVar(&cfg.ServerURL, "a", "localhost:8080", "Net address host:port")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "Key that will be used to calculate hash")
	flag.IntVar(&cfg.RateLimit, "l", 3, "Max number of concurrent requests to the server")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Path to public key that will be used for payload encryption")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
