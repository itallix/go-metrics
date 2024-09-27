package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentConfig_Parse(t *testing.T) {
	tests := []struct {
		name          string
		giveArgs      []string
		wantAddr      string
		wantPoll      int
		wantReport    int
		wantKey       string
		wantRateLimit int
		wantCryptoKey string
	}{
		{
			name:          "Default",
			wantAddr:      "localhost:8080",
			wantPoll:      2,
			wantReport:    10,
			wantKey:       "",
			wantRateLimit: 3,
			wantCryptoKey: "",
		},
		{
			name:          "WithArgs",
			giveArgs:      []string{"-a", "localhost:8081", "-p", "4", "-r", "20", "-k", "key", "-l", "5", "-crypto-key", "cryptoKey"},
			wantAddr:      "localhost:8081",
			wantPoll:      4,
			wantReport:    20,
			wantKey:       "key",
			wantRateLimit: 5,
			wantCryptoKey: "cryptoKey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			if tt.giveArgs != nil {
				os.Args = append([]string{os.Args[0]}, tt.giveArgs...)
			}
			cfg, err := parseConfig()
			require.NoError(t, err)

			assert.Equal(t, tt.wantAddr, cfg.ServerURL)
			assert.Equal(t, tt.wantPoll, cfg.PollInterval)
			assert.Equal(t, tt.wantReport, cfg.ReportInterval)
			assert.Equal(t, tt.wantKey, cfg.Key)
			assert.Equal(t, tt.wantRateLimit, cfg.RateLimit)
			assert.Equal(t, tt.wantCryptoKey, cfg.CryptoKey)
		})
	}
}
