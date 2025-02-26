package openrouter

import (
	"errors"
	"time"
)

// Config holds configuration for the OpenRouter client
type Config struct {
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
	MaxRetries  int
	RetryDelay  time.Duration
	BaseURL     string
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return errors.New("API key is required")
	}
	return nil
}

// SetDefaults sets default values for optional configuration
func (c *Config) SetDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = "https://openrouter.ai/api/v1"
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.RetryDelay == 0 {
		c.RetryDelay = time.Second * 2
	}
	if c.Temperature == 0 {
		c.Temperature = 0.7
	}
	if c.MaxTokens == 0 {
		c.MaxTokens = 1000
	}
}
