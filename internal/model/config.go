package model

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ServerConfig describes customization settings for the server.
type ServerConfig struct {
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
	ServerURL      string `env:"ADDRESS" json:"address"`                 // Address of the server to send metrics to.
	PollInterval   int    `env:"POLL_INTERVAL" json:"poll_interval"`     // How often to collect metrics from the system.
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval"` // How often to report collected metrics to the server.
	Key            string `env:"KEY" json:"hash_secret"`                 // Secret for a hash function.
	RateLimit      int    `env:"RATE_LIMIT" json:"rate_limit"`           // Limits number of concurrent requests to the server.
	CryptoKey      string `env:"CRYPTO_KEY" json:"crypto_key"`           // Public key used to encrypt request payload.
}

var re = regexp.MustCompile(`("\w*_interval"):\s*"(\d+)s`)

// ParseFileConfig used to read configuration for server or agent from the specified JSON file.
func ParseFileConfig[T ServerConfig | AgentConfig](path string, cfg *T) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open config file: %w", err)
	}

	// In the Config struct, intervals are represented as integer values (int)
	// In the JSON configuration, intervals may be strings with a suffix (e.g., "5s")
	// Use a scanner to read JSON line by line, then extract and convert
	// numeric values from strings, ignoring unit suffixes
	scanner := bufio.NewScanner(f)
	var jsonBuilder strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) > 2 {
			jsonBuilder.WriteString(fmt.Sprintf("%s: %s,", matches[1], matches[2]))
		} else {
			jsonBuilder.WriteString(line)
		}
	}

	jsonString := jsonBuilder.String()
	if err := json.Unmarshal([]byte(jsonString), &cfg); err != nil {
		return fmt.Errorf("cannot parse config file: %w", err)
	}
	return nil
}
