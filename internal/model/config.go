package model

import (
	"encoding/json"
	"fmt"
	"os"
)

// ServerConfig describes customization settings for the server.
type ServerConfig struct {
	ConfigPath    string `env:"CONFIG"`                               // Path to json config file.
	Address       string `env:"ADDRESS" json:"address"`               // Address where server will be started.
	StoreInterval int    `env:"STORE_INTERVAL" json:"store_interval"` // How often to store metrics in the file.
	FilePath      string `env:"FILE_STORAGE_PATH" json:"store_file"`  // Location where to store metrics.
	Restore       bool   `env:"RESTORE" json:"restore"`               // Should the metrics be loaded from the file on start?
	DatabaseDSN   string `env:"DATABASE_DSN" json:"database_dsn"`     // DB connection string, example: postgresql://username:password@hostname:port/database_name.
	Key           string `env:"KEY" json:"hash_secret"`               // Secret for a hash function.
	CryptoKey     string `env:"CRYPTO_KEY" json:"crypto_key"`         // Private key used to decrypt request payload.
}

// AgentConfig describes customization settings for the agent.
type AgentConfig struct {
	ConfigPath     string `env:"CONFIG"`                                 // Path to json config file.
	ServerURL      string `env:"ADDRESS" json:"address"`                 // Address of the server to send metrics to.
	PollInterval   int    `env:"POLL_INTERVAL" json:"poll_interval"`     // How often to collect metrics from the system.
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval"` // How often to report collected metrics to the server.
	Key            string `env:"KEY" json:"hash_secret"`                 // Secret for a hash function.
	RateLimit      int    `env:"RATE_LIMIT" json:"rate_limit"`           // Limits number of concurrent requests to the server.
	CryptoKey      string `env:"CRYPTO_KEY" json:"crypto_key"`           // Public key used to encrypt request payload.
}

// ParseFileConfig used to read configuration for server or agent from the specified JSON file.
func ParseFileConfig[T ServerConfig | AgentConfig](path string, cfg *T) error {
	cfgBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot open config file: %w", err)
	}
	if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
		return fmt.Errorf("cannot parse config file: %w", err)
	}
	return nil
}
