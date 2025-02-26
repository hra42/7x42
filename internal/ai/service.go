package ai

import (
	"github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/ai/service"
	"gorm.io/gorm"
)

type Service struct {
	service *service.Service
}

// Config holds the configuration for the AI service
type Config struct {
	DB            *gorm.DB
	OpenRouterKey string
	Model         string
	Temperature   float64
	MaxTokens     int
}

func NewService(db *gorm.DB) (*Service, error) {
	svc, err := service.NewFromDB(db)
	if err != nil {
		return nil, err
	}
	return &Service{
		service: svc,
	}, nil
}

func NewServiceWithConfig(config Config) (*Service, error) {
	// Create service directly using the configuration
	svc, err := service.New(service.Config{
		DB:            config.DB,
		OpenRouterKey: config.OpenRouterKey,
		Model:         config.Model,
		Temperature:   config.Temperature,
		MaxTokens:     config.MaxTokens,
	})

	if err != nil {
		return nil, err
	}

	return &Service{
		service: svc,
	}, nil
}

func (s *Service) HandleChatMessage(wsConn *websocket.Conn, chatID uint, content string, userID string) error {
	return s.service.HandleChatMessage(wsConn, chatID, content, userID)
}
