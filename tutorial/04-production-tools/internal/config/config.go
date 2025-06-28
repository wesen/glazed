package config

import (
	"fmt"
	"os"
	"path/filepath"

	"tutorial-04/internal/logging"
	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	App     AppConfig     `yaml:"app" json:"app"`
	Logging logging.Config `yaml:"logging" json:"logging"`
	Server  ServerConfig  `yaml:"server" json:"server"`
	Storage StorageConfig `yaml:"storage" json:"storage"`
}

// AppConfig contains application-specific settings
type AppConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Environment string `yaml:"environment" json:"environment"`
	Debug       bool   `yaml:"debug" json:"debug"`
}

// ServerConfig contains server settings
type ServerConfig struct {
	Host           string `yaml:"host" json:"host"`
	Port           int    `yaml:"port" json:"port"`
	MetricsEnabled bool   `yaml:"metrics_enabled" json:"metrics_enabled"`
	MetricsPort    int    `yaml:"metrics_port" json:"metrics_port"`
}

// StorageConfig contains storage settings
type StorageConfig struct {
	Type      string            `yaml:"type" json:"type"`
	Path      string            `yaml:"path" json:"path"`
	Options   map[string]string `yaml:"options" json:"options"`
	Retention int               `yaml:"retention_days" json:"retention_days"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "log-processor",
			Version:     "1.0.0",
			Environment: "development",
			Debug:       false,
		},
		Logging: logging.Config{
			Level:      "info",
			Format:     "console",
			Output:     "stderr",
			AddCaller:  false,
			TimeFormat: "2006-01-02T15:04:05Z07:00",
		},
		Server: ServerConfig{
			Host:           "localhost",
			Port:           8080,
			MetricsEnabled: true,
			MetricsPort:    9090,
		},
		Storage: StorageConfig{
			Type:      "local",
			Path:      "./data",
			Options:   map[string]string{},
			Retention: 30,
		},
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Load from file if it exists
	if configPath != "" && fileExists(configPath) {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	if env := os.Getenv("LOG_PROCESSOR_ENV"); env != "" {
		config.App.Environment = env
	}
	if debug := os.Getenv("LOG_PROCESSOR_DEBUG"); debug == "true" {
		config.App.Debug = true
	}
	if logLevel := os.Getenv("LOG_PROCESSOR_LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name is required")
	}
	
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	if c.Storage.Type == "" {
		return fmt.Errorf("storage type is required")
	}
	
	if c.Storage.Retention < 0 {
		return fmt.Errorf("storage retention must be non-negative")
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
