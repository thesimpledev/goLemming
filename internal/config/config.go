// Package config provides configuration management for GoLemming.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds the application configuration.
type Config struct {
	// API settings
	APIKey string `json:"api_key,omitempty"`
	Model  string `json:"model,omitempty"`

	// Agent settings
	MaxIterations     int `json:"max_iterations,omitempty"`
	StabilizationMs   int `json:"stabilization_ms,omitempty"`
	DefaultWaitMs     int `json:"default_wait_ms,omitempty"`
	ScreenshotQuality int `json:"screenshot_quality,omitempty"`

	// Safety settings
	RequireAbsolutePaths bool `json:"require_absolute_paths,omitempty"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Model:                "claude-sonnet-4-20250514",
		MaxIterations:        100,
		StabilizationMs:      500,
		DefaultWaitMs:        500,
		ScreenshotQuality:    80,
		RequireAbsolutePaths: true,
	}
}

// ConfigDir returns the configuration directory path.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".golemming"), nil
}

// ConfigPath returns the configuration file path.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load loads configuration from file and environment variables.
// Environment variables take precedence over file settings.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Try to load from config file
	if err := cfg.loadFromFile(); err != nil {
		// Ignore file not found errors
		if !os.IsNotExist(err) {
			// Log but don't fail
		}
	}

	// Environment variables override file settings
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}

	if model := os.Getenv("GOLEMMING_MODEL"); model != "" {
		cfg.Model = model
	}

	// Validate required fields
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set (set environment variable or run 'golemming' to configure)")
	}

	return cfg, nil
}

// loadFromFile loads configuration from the config file.
func (c *Config) loadFromFile() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

// Save saves the configuration to the config file.
func (c *Config) Save() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with restricted permissions (user-only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// StabilizationDelay returns the stabilization delay as a duration.
func (c *Config) StabilizationDelay() time.Duration {
	return time.Duration(c.StabilizationMs) * time.Millisecond
}

// DefaultWait returns the default wait time as a duration.
func (c *Config) DefaultWait() time.Duration {
	return time.Duration(c.DefaultWaitMs) * time.Millisecond
}
