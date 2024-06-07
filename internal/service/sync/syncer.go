package sync

import (
	"context"
)

type Config struct {
	interval int
	restore  bool
	syncCh   chan int
}

type Syncer interface {
	Start(ctx context.Context, cfg *Config)
}

func NewConfig(interval int, restore bool, syncCh chan int) *Config {
	return &Config{
		interval: interval,
		restore:  restore,
		syncCh:   syncCh,
	}
}
