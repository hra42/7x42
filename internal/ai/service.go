package ai

import (
	"github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/ai/service"
	"gorm.io/gorm"
)

// Service is the main AI service interface
type Service struct {
	service *service.Service
}

// NewService creates a new AI service
func NewService(db *gorm.DB) (*Service, error) {
	svc, err := service.NewFromDB(db)
	if err != nil {
		return nil, err
	}

	return &Service{
		service: svc,
	}, nil
}

// HandleChatMessage processes a chat message and generates a response
func (s *Service) HandleChatMessage(wsConn *websocket.Conn, chatID uint, content string, userID string) error {
	return s.service.HandleChatMessage(wsConn, chatID, content, userID)
}
