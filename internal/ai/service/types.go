package service

import (
	"github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/ai/openrouter"
	"github.com/hra42/7x42/internal/repository"
	"gorm.io/gorm"
)

// Provider defines the interface for AI service providers
type Provider interface {
	Initialize() error
	HandleChatMessage(wsConn *websocket.Conn, chatID uint, content string, userID string) error
}

// Config holds the configuration for the AI service
type Config struct {
	DB            *gorm.DB
	OpenRouterKey string
	Model         string
	Temperature   float64
	MaxTokens     int
}

// Service is the main AI service that coordinates AI providers
type Service struct {
	openRouter  *openrouter.Client
	chatRepo    *repository.ChatRepository
	messageRepo *repository.MessageRepository
	config      Config
}
