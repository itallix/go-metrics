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

type AgentConfig struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

func parseFlags() (*mflag.RunAddress, *AgentConfig, error) {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")

	var cfg AgentConfig
	flag.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "Key that will be used to calculate hash")
	flag.IntVar(&cfg.RateLimit, "l", 3, "Max number of requests to the server")
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
