package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentConfig_Parse(t *testing.T) {
	tests := []struct {
		name          string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
