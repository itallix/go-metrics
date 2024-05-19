package main

import (
	"flag"
	"fmt"
	"os"

	mflag "github.com/itallix/go-metrics/internal/flag"
)

const (
	EnvAddress = "ADDRESS"
)

func parseFlags() (*mflag.RunAddress, error) {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")
	flag.Parse()

	if envAddr := os.Getenv(EnvAddress); envAddr != "" {
		if err := addr.Set(envAddr); err != nil {
			return nil, fmt.Errorf("cannot parse ADDRESS env var: %w", err)
		}
	}
	return addr, nil
}
