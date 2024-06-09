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

type StoreSettings struct {
	StoreInterval int    `env:"STORE_INTERVAL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
}

func NewStoreSettings(storeInterval int, filepath string, restore bool) *StoreSettings {
	return &StoreSettings{
		StoreInterval: storeInterval,
		FilePath:      filepath,
		Restore:       restore,
	}
}

func parseFlags() (*mflag.RunAddress, *StoreSettings, error) {
	addr := mflag.NewRunAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "Net address host:port")

	var (
		storeInterval int
		filepath      string
		restore       bool
	)
	flag.IntVar(&storeInterval, "i", 300, "Store interval in seconds")
	flag.StringVar(&filepath, "f", "/tmp/metrics-db.json", "Filepath where metrics will be saved")
	flag.BoolVar(&restore, "r", true, "Whether server needs to restore metrics from file or not")
	flag.Parse()

	storageSettings := NewStoreSettings(storeInterval, filepath, restore)

	if envAddr := os.Getenv(EnvAddress); envAddr != "" {
		if err := addr.Set(envAddr); err != nil {
			return nil, nil, fmt.Errorf("cannot parse ADDRESS env var: %w", err)
		}
	}

	if err := env.Parse(storageSettings); err != nil {
		return nil, nil, err
	}
	return addr, storageSettings, nil
}
