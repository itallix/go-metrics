package main

import (
	"flag"
	mflag "github.com/itallix/go-metrics/internal/flag"
	"time"
)

var (
	PollInterval   time.Duration
	ReportInterval time.Duration
)

func parseFlags() *mflag.RunAddress {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")
	flag.DurationVar(&PollInterval, "p", 2*time.Second, "Poll interval in seconds")
	flag.DurationVar(&ReportInterval, "r", 10*time.Second, "Report interval in seconds")
	flag.Parse()
	return addr
}
