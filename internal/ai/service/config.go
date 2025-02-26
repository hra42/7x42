package service

import (
	"errors"
	"os"
	"strconv"

	"github.com/hra42/7x42/internal/ai/openrouter"
)

// ValidateConfig checks if the service configuration is valid
func ValidateConfig(config Config) error {
	if config.DB == nil {
		return errors.New("database connection is required")
	}

	if config.OpenRouterKey == "" {
		return errors.New("OpenRouter API key is required")
	}

	return nil
}

// CreateOpenRouterConfig creates an OpenRouter client configuration from service config
func CreateOpenRouterConfig(config Config) openrouter.Config {
	return openrouter.Config{
		APIKey:      config.OpenRouterKey,
		Model:       config.Model,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}
}

// LoadConfigFromEnv loads service configuration from environment variables
func LoadConfigFromEnv() Config {
	return Config{
		OpenRouterKey: os.Getenv("OPENROUTER_API_KEY"),
		Model:         getEnvWithDefault("OPENROUTER_MODEL", "google/gemini-2.0-flash-001"),
		Temperature:   getEnvAsFloat("OPENROUTER_TEMPERATURE", 0.7),
		MaxTokens:     getEnvAsInt("OPENROUTER_MAX_TOKENS", 1000),
	}
}

// Helper functions for environment variables
func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsFloat parses an environment variable as a float64
func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

// getEnvAsInt parses an environment variable as an integer
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
