package service

import (
	"fmt"

	"github.com/hra42/7x42/internal/ai/openrouter"
	"github.com/hra42/7x42/internal/repository"
	"gorm.io/gorm"
)

// New creates a new AI service with the given configuration
func New(config Config) (*Service, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	chatRepo := repository.NewChatRepository(config.DB)
	messageRepo := repository.NewMessageRepository(config.DB)

	openRouterConfig := CreateOpenRouterConfig(config)
	openRouterClient, err := openrouter.New(openRouterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRouter client: %w", err)
	}

	if err := openRouterClient.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenRouter client: %w", err)
	}

	openRouterClient.SetChatRepository(chatRepo)
	openRouterClient.SetMessageRepository(messageRepo)

	return &Service{
		openRouter:  openRouterClient,
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		config:      config,
	}, nil
}

// NewFromDB creates a new AI service from a database connection
// This is a convenience function for creating a service with environment variables
func NewFromDB(db *gorm.DB) (*Service, error) {
	config := LoadConfigFromEnv()
	config.DB = db

	return New(config)
}
