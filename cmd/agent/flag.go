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

func NewIntervalSettings(pollInterval int64, reportInterval int64) *IntervalSettings {
	return &IntervalSettings{
		PollInterval:   time.Duration(pollInterval) * time.Second,
		ReportInterval: time.Duration(reportInterval) * time.Second,
	}
}

func parseFlags() (*mflag.RunAddress, *IntervalSettings) {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")
	var pollInterval, reportInterval int64
	flag.Int64Var(&pollInterval, "p", 2, "Poll interval in seconds")
	flag.Int64Var(&reportInterval, "r", 10, "Report interval in seconds")
	flag.Parse()
	return addr, NewIntervalSettings(pollInterval, reportInterval)
}
