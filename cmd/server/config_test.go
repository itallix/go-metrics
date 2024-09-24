package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfig_Parse(t *testing.T) {
	tests := []struct {
		name              string
		giveArgs          []string
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
		{
			name:              "WithArgs",
			giveArgs:          []string{"-a", "localhost:8081", "-i", "400", "-f", "filepath", "-d", "dsn", "-k", "key", "-crypto-key", "cryptoKey"},
			wantAddr:          "localhost:8081",
			wantFilepath:      "filepath",
			wantStoreInterval: 400,
			wantRestore:       true,
			wantKey:           "key",
			wantDSN:           "dsn",
			wantCryptoKey:     "cryptoKey",
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
