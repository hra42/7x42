package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/models"
	"github.com/hra42/7x42/internal/repository"
	"gorm.io/gorm"
)

type Service struct {
	openRouter *OpenRouterClient
	chatRepo   *repository.ChatRepository
}

func NewService(db *gorm.DB) (*Service, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENROUTER_API_KEY environment variable is not set")
	}

	// Default model if not specified
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "openai/gpt-4o"
	}

	openRouter, err := NewOpenRouterClient(Config{
		APIKey:      apiKey,
		Model:       model,
		Temperature: 0.7,
		MaxTokens:   4000,
		MaxRetries:  3,
		RetryDelay:  time.Second * 2,
		BaseURL:     "https://openrouter.ai/api/v1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRouter client: %w", err)
	}

	chatRepo := repository.NewChatRepository(db)
	openRouter.SetChatRepository(chatRepo)

	if err := openRouter.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenRouter client: %w", err)
	}

	return &Service{
		openRouter: openRouter,
		chatRepo:   chatRepo,
	}, nil
}

func (s *Service) HandleChatMessage(wsConn *websocket.Conn, chatID uint, content string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Save user message
	userMsg := &models.Message{
		Content:   content,
		Role:      "user",
		ChatID:    uint64(chatID),
		Timestamp: time.Now(),
		Metadata:  models.MessageMetadata{},
	}

	if err := s.chatRepo.CreateMessage(ctx, userMsg); err != nil {
		return fmt.Errorf("failed to save user message: %w", err)
	}

	// Get or create chat
	chat, err := s.chatRepo.GetChat(ctx, uint64(chatID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create a new chat if not found
			chat = &models.Chat{
				// Remove the ID field as it's automatically handled by GORM
				Title:       content[:minInt(30, len(content))],
				UserID:      userID,
				LastMessage: time.Now(),
			}
			if err := s.chatRepo.CreateChat(ctx, chat); err != nil {
				return fmt.Errorf("failed to create chat: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get chat history: %w", err)
		}
	}

	// Convert chat messages to the format expected by OpenRouter
	messages := convertMessagesToOpenRouterFormat(chat.Messages)

	// Stream response if WebSocket is provided
	if wsConn != nil {
		if err := s.openRouter.StreamResponse(ctx, wsConn, content, messages); err != nil {
			log.Printf("Error streaming response: %v", err)
			// Fallback to non-streaming response
			response, fallbackErr := s.openRouter.GenerateResponse(ctx, content, messages)
			if fallbackErr != nil {
				return fmt.Errorf("failed to generate response (fallback): %w", fallbackErr)
			}

			// Save AI response
			aiMsg := &models.Message{
				Content:   response,
				Role:      "assistant",
				ChatID:    uint64(chatID),
				Timestamp: time.Now(),
				Metadata:  models.MessageMetadata{},
			}
			if err := s.chatRepo.CreateMessage(ctx, aiMsg); err != nil {
				return fmt.Errorf("failed to save AI response: %w", err)
			}

			// Send complete message
			if err := wsConn.WriteJSON(map[string]interface{}{
				"type": "chat_message",
				"content": map[string]interface{}{
					"content":   response,
					"role":      "assistant",
					"timestamp": time.Now(),
				},
			}); err != nil {
				return fmt.Errorf("failed to send complete message: %w", err)
			}
		}
		return nil
	}

	// Non-streaming response if no WebSocket is provided
	response, err := s.openRouter.GenerateResponse(ctx, content, messages)
	if err != nil {
		return fmt.Errorf("failed to generate response: %w", err)
	}

	// Save AI response
	aiMsg := &models.Message{
		Content:   response,
		Role:      "assistant",
		ChatID:    uint64(chatID),
		Timestamp: time.Now(),
		Metadata:  models.MessageMetadata{},
	}
	if err := s.chatRepo.CreateMessage(ctx, aiMsg); err != nil {
		return fmt.Errorf("failed to save AI response: %w", err)
	}

	return nil
}

// Helper function to convert our models.Message to the format expected by OpenRouter
func convertMessagesToOpenRouterFormat(messages []models.Message) []Message {
	// Limit to last 10 messages to avoid context length issues
	startIdx := 0
	if len(messages) > 10 {
		startIdx = len(messages) - 10
	}

	result := make([]Message, 0, len(messages)-startIdx)
	for i := startIdx; i < len(messages); i++ {
		msg := messages[i]
		result = append(result, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
