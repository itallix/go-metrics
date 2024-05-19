package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	mflag "github.com/itallix/go-metrics/internal/flag"
)

const (
	EnvAddress        = "ADDRESS"
	EnvReportInterval = "REPORT_INTERVAL"
	EnvPollInterval   = "POLL_INTERVAL"
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

func parseFlags() (*mflag.RunAddress, *IntervalSettings, error) {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")
	var pollInterval, reportInterval int64
	flag.Int64Var(&pollInterval, "p", 2, "Poll interval in seconds")
	flag.Int64Var(&reportInterval, "r", 10, "Report interval in seconds")
	flag.Parse()

	intervalSettings := NewIntervalSettings(pollInterval, reportInterval)

	if envAddr := os.Getenv(EnvAddress); envAddr != "" {
		if err := addr.Set(envAddr); err != nil {
			return nil, nil, fmt.Errorf("cannot parse ADDRESS env var: %w", err)
		}
	}
	if envPollInterval := os.Getenv(EnvPollInterval); envPollInterval != "" {
		pollValue, err := strconv.ParseInt(envPollInterval, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot convert %s env var: %s", EnvPollInterval, envPollInterval)
		}
		intervalSettings.PollInterval = time.Duration(pollValue) * time.Second
	}
	if envReportInterval := os.Getenv(EnvReportInterval); envReportInterval != "" {
		reportValue, err := strconv.ParseInt(envReportInterval, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot convert %s env var: %s", EnvReportInterval, envReportInterval)
		}
		intervalSettings.ReportInterval = time.Duration(reportValue) * time.Second
	}
	return addr, intervalSettings, nil
}
