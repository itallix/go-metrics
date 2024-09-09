package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfig_Parse(t *testing.T) {
	tests := []struct {
		name              string
		wantAddr          string
		wantFilepath      string
		wantStoreInterval int
		wantRestore       bool
		wantKey           string
		wantDSN           string
		wantCryptoKey     string
	}{
		{
			name:              "Default",
			wantAddr:          "localhost:8080",
			wantFilepath:      "/tmp/metrics-db.json",
			wantStoreInterval: 300,
			wantRestore:       true,
			wantKey:           "",
			wantDSN:           "",
			wantCryptoKey:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseConfig()
			require.NoError(t, err)

			assert.Equal(t, tt.wantAddr, cfg.Address)
			assert.Equal(t, tt.wantFilepath, cfg.FilePath)
			assert.Equal(t, tt.wantStoreInterval, cfg.StoreInterval)
			assert.Equal(t, tt.wantRestore, cfg.Restore)
			assert.Equal(t, tt.wantKey, cfg.Key)
			assert.Equal(t, tt.wantDSN, cfg.DatabaseDSN)
			assert.Equal(t, tt.wantCryptoKey, cfg.CryptoKey)
		})
	}
}
