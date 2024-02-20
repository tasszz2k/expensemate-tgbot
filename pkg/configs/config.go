package configs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type (
	AppConfig struct {
		TelegramBot  TelegramBot  `yaml:"telegram_bot"  validate:"required"`
		GoogleApis   GoogleApis   `yaml:"google_apis"   validate:"required"`
		GoogleSheets GoogleSheets `yaml:"google_sheets" validate:"required"`
	}

	TelegramBot struct {
		ApiToken string `yaml:"api_token" validate:"required"`
		Timeout  int    `yaml:"timeout"   validate:"required"`
		Debug    bool   `yaml:"debug"`
	}

	GoogleApis struct {
		Credentials Credentials `yaml:"credentials" validate:"required"`
	}

	Credentials struct {
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

	GoogleSheets struct {
		DatabaseSpreadsheetId string `yaml:"database_spreadsheet_id" validate:"required"`
	}
)

var (
	_, b, _, _           = runtime.Caller(0)
	defaultConfigFile, _ = getConfigFilePath("configs/local.yaml")
	appCfg               AppConfig
)

func init() {
	Load()
}

func Load() {
	configReaderMode := os.Getenv("CONFIG_READER_MODE")

	switch configReaderMode {
	case "secret":
		if err := loadConfigFromSecret(); err != nil {
			panic(err)
		}
	default: //case "file"
		var configPath string
		if configPath = os.Getenv("CONFIG_PATH"); len(configPath) == 0 {
			configPath = defaultConfigFile
		}
		var err error
		if appCfg, err = loadConfigFromFile(configPath); err != nil {
			panic(err)
		}
	}
}

// loadConfigFromSecret decodes the base64 encoded YAML content from the environment variable ENCODED_CONFIG
// to support secret-based configuration for "free deployment" environments
func loadConfigFromSecret() error {
	base64EncodedConfig := os.Getenv("ENCODED_CONFIG")
	if base64EncodedConfig == "" {
		log.Fatal("ENCODED_CONFIG is empty")
	}

	decodedConfig, err := base64.StdEncoding.DecodeString(base64EncodedConfig)
	if err != nil {
		return err
	}

	// Now, unmarshal the decoded YAML content directly into the config struct
	err = yaml.Unmarshal(decodedConfig, &appCfg)
	if err != nil {
		return err
	}

	return nil
}

func loadConfigFromFile(path string) (AppConfig, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return AppConfig{}, fmt.Errorf("config read: open file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	switch ext {
	case "json":
		err = json.NewDecoder(f).Decode(&appCfg)
	case "yaml", "yml":
		err = yaml.NewDecoder(f).Decode(&appCfg)
	default:
		return AppConfig{}, errors.New("config read: invalid file type: " + ext)
	}

	if errors.Is(err, io.EOF) {
		return appCfg, nil
	}
	if err != nil {
		return AppConfig{}, fmt.Errorf("config read: parse file: %w", err)
	}

	return appCfg, nil
}

func Get() AppConfig {
	return appCfg
}

// getConfigFilePath returns the absolute path for the given relative file path.
func getConfigFilePath(relativePath string) (string, error) {
	// Move back to the root directory
	relativePath = "../" + relativePath

	// Navigate up one level to reach the root directory
	rootDir := filepath.Dir(filepath.Dir(b))

	// Construct the absolute path of the config file
	configPath := filepath.Join(rootDir, relativePath)

	// Clean and return the absolute path
	return filepath.Abs(configPath)
}
