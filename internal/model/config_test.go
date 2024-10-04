package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAgentConfig(t *testing.T) {
	wantConfig := &AgentConfig{
		ServerURL:      "localhost:8080",
		PollInterval:   1,
		ReportInterval: 1,
		CryptoKey:      "/path/to/key.pem",
	}

	var cfg AgentConfig
	err := ParseFileConfig("../../test_data/agentConfig.json", &cfg)
	require.NoError(t, err)

	assert.Equal(t, wantConfig.ServerURL, cfg.ServerURL)
	assert.Equal(t, wantConfig.PollInterval, cfg.PollInterval)
	assert.Equal(t, wantConfig.ReportInterval, cfg.ReportInterval)
	assert.Equal(t, wantConfig.Key, cfg.Key)
	assert.Equal(t, wantConfig.RateLimit, cfg.RateLimit)
	assert.Equal(t, wantConfig.CryptoKey, cfg.CryptoKey)
}

func TestParseServerConfig(t *testing.T) {
	wantConfig := &ServerConfig{
		Address:       "localhost:8080",
		Restore:       true,
		StoreInterval: 1,
		FilePath:      "/path/to/file.db",
		CryptoKey:     "/path/to/key.pem",
		TrustedSubnet: "192.168.2.0/24",
	}

	var cfg ServerConfig
	err := ParseFileConfig("../../test_data/serverConfig.json", &cfg)
	require.NoError(t, err)

	assert.Equal(t, wantConfig.Address, cfg.Address)
	assert.Equal(t, wantConfig.FilePath, cfg.FilePath)
	assert.Equal(t, wantConfig.StoreInterval, cfg.StoreInterval)
	assert.Equal(t, wantConfig.Restore, cfg.Restore)
	assert.Equal(t, wantConfig.Key, cfg.Key)
	assert.Equal(t, wantConfig.DatabaseDSN, cfg.DatabaseDSN)
	assert.Equal(t, wantConfig.CryptoKey, cfg.CryptoKey)
	assert.Equal(t, wantConfig.TrustedSubnet, cfg.TrustedSubnet)
}
