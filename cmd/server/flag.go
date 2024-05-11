package main

import (
	"flag"
	"log"
	"os"

	mflag "github.com/itallix/go-metrics/internal/flag"
)

const (
	EnvAddress = "ADDRESS"
)

func parseFlags() *mflag.RunAddress {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")
	flag.Parse()

	if envAddr := os.Getenv(EnvAddress); envAddr != "" {
		if err := addr.Set(envAddr); err != nil {
			log.Fatalf("Cannot parse ADDRESS env var: %s", err)
		}
	}
	return addr
}
