package main

import (
	"flag"

	mflag "github.com/itallix/go-metrics/internal/flag"
)

func parseFlags() *mflag.RunAddress {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")
	flag.Parse()
	return addr
}
