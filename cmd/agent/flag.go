package main

import (
	"flag"
	"time"

	mflag "github.com/itallix/go-metrics/internal/flag"
)

type IntervalSettings struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func parseFlags() (*mflag.RunAddress, *IntervalSettings) {
	addr := mflag.NewRunAddress()
	intervalSettings := &IntervalSettings{}
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")
	flag.DurationVar(&intervalSettings.PollInterval, "p", 2*time.Second, "Poll interval in seconds")
	flag.DurationVar(&intervalSettings.ReportInterval, "r", 10*time.Second, "Report interval in seconds")
	flag.Parse()
	return addr, intervalSettings
}
