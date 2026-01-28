package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	TelegramBot  TelegramBot  `yaml:"telegram_bot" validate:"required"`
	GoogleAPIs   GoogleAPIs   `yaml:"google_apis" validate:"required"`
	GoogleSheets GoogleSheets `yaml:"google_sheets" validate:"required"`
}

// TelegramBot configuration
type TelegramBot struct {
	APIToken string `yaml:"api_token" validate:"required"`
	Timeout  int    `yaml:"timeout" validate:"required"`
	Debug    bool   `yaml:"debug"`     // Application debug logging (input/output, detailed logs)
	BotDebug bool   `yaml:"bot_debug"` // Telegram Bot API debug (endpoint requests/responses)
}

// GoogleAPIs configuration
type GoogleAPIs struct {
	Credentials Credentials `yaml:"credentials" validate:"required"`
}

// Credentials for Google service account
type Credentials struct {
	Type                string `yaml:"type"`
	ProjectID           string `yaml:"project_id"`
	PrivateKeyID        string `yaml:"private_key_id"`
	PrivateKey          string `yaml:"private_key"`
	ClientEmail         string `yaml:"client_email"`
	ClientID            string `yaml:"client_id"`
	AuthURI             string `yaml:"auth_uri"`
	TokenURI            string `yaml:"token_uri"`
	AuthProviderCertURL string `yaml:"auth_provider_x509_cert_url"`
	ClientCertURL       string `yaml:"client_x509_cert_url"`
}

// GoogleSheets configuration
type GoogleSheets struct {
	DatabaseSpreadsheetID string `yaml:"database_spreadsheet_id" validate:"required"`
}

// Load loads configuration from file or environment
func Load() (*Config, error) {
	mode := os.Getenv("CONFIG_READER_MODE")

	switch mode {
	case "secret":
		return loadFromSecret()
	default:
		path := os.Getenv("CONFIG_PATH")
		if path == "" {
			path = "configs/local.yaml"
		}
		return loadFromFile(path)
	}
}

// loadFromSecret loads config from base64-encoded environment variable
func loadFromSecret() (*Config, error) {
	encoded := os.Getenv("ENCODED_CONFIG")
	if encoded == "" {
		return nil, errors.New("ENCODED_CONFIG environment variable is empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(decoded, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// loadFromFile loads config from a YAML or JSON file
func loadFromFile(path string) (*Config, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	switch ext {
	case "json":
		err = json.NewDecoder(f).Decode(&cfg)
	case "yaml", "yml":
		err = yaml.NewDecoder(f).Decode(&cfg)
	default:
		return nil, fmt.Errorf("unsupported config file type: %s", ext)
	}

	if errors.Is(err, io.EOF) {
		return &cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}
