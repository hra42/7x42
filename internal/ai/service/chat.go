package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/ai/openrouter"
	"github.com/hra42/7x42/internal/models"
	"gorm.io/gorm"
)

// HandleChatMessage processes a chat message and generates a response
func (s *Service) HandleChatMessage(wsConn *websocket.Conn, chatID uint, content string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Save user message
	if err := s.saveUserMessage(ctx, chatID, content, userID); err != nil {
		return fmt.Errorf("failed to save user message: %w", err)
	}

	// Get or create chat
	chat, err := s.getOrCreateChat(ctx, chatID, content, userID)
	if err != nil {
		return fmt.Errorf("failed to get or create chat: %w", err)
	}

	// Convert messages to OpenRouter format
	messages := s.convertMessagesToOpenRouterFormat(chat.Messages)

	// Generate response using streaming if WebSocket is available
	if wsConn != nil {
		return s.streamResponse(ctx, wsConn, content, messages)
	}

	// Fall back to non-streaming response if WebSocket is not available
	return s.generateResponse(ctx, chatID, content, messages)
}

// saveUserMessage saves the user's message to the database
func (s *Service) saveUserMessage(ctx context.Context, chatID uint, content string, userID string) error {
	userMsg := &models.Message{
		ChatID:    uint64(chatID),
		Content:   content,
		Role:      "user",
		Timestamp: time.Now(),
		Metadata:  models.MessageMetadata{},
	}

	return s.chatRepo.CreateMessage(ctx, userMsg)
}

// getOrCreateChat gets an existing chat or creates a new one if it doesn't exist
func (s *Service) getOrCreateChat(ctx context.Context, chatID uint, content string, userID string) (*models.Chat, error) {
	if chatID != 0 {
		chat, err := s.chatRepo.GetChat(ctx, uint64(chatID))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		if chat != nil {
			return chat, nil
		}
	}

	// Create new chat if chatID is 0 or chat not found
	title := content
	if len(title) > 30 {
		title = title[:30]
	}

	chat := &models.Chat{
		Title:       title,
		LastMessage: time.Now(),
		UserID:      userID,
	}

	if err := s.chatRepo.CreateChat(ctx, chat); err != nil {
		return nil, err
	}

	return chat, nil
}

// streamResponse streams an AI response through a WebSocket connection
func (s *Service) streamResponse(ctx context.Context, wsConn *websocket.Conn, content string, messages []openrouter.ChatMessage) error {
	if err := s.openRouter.StreamResponse(ctx, wsConn, content, messages); err != nil {
		log.Printf("Error streaming response: %v", err)

		// Try fallback to non-streaming response
		response, fallbackErr := s.openRouter.GenerateResponse(ctx, content, messages)
		if fallbackErr != nil {
			return fmt.Errorf("failed to generate response (fallback): %w", fallbackErr)
		}

		// Send fallback response as a complete message
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

// generateResponse generates a non-streaming AI response
func (s *Service) generateResponse(ctx context.Context, chatID uint, content string, messages []openrouter.ChatMessage) error {
	response, err := s.openRouter.GenerateResponse(ctx, content, messages)
	if err != nil {
		return fmt.Errorf("failed to generate response: %w", err)
	}

	// Save AI response to database
	aiMsg := &models.Message{
		ChatID:    uint64(chatID),
		Content:   response,
		Role:      "assistant",
		Timestamp: time.Now(),
		Metadata: models.MessageMetadata{
			Model: s.config.Model,
		},
	}

	return s.chatRepo.CreateMessage(ctx, aiMsg)
}

// convertMessagesToOpenRouterFormat converts database messages to OpenRouter format
func (s *Service) convertMessagesToOpenRouterFormat(messages []models.Message) []openrouter.ChatMessage {
	// Limit to last 10 messages to avoid context length issues
	startIdx := 0
	if len(messages) > 10 {
		startIdx = len(messages) - 10
	}

	result := make([]openrouter.ChatMessage, 0, len(messages)-startIdx)
	for i := startIdx; i < len(messages); i++ {
		result = append(result, openrouter.ChatMessage{
			Role:    messages[i].Role,
			Content: messages[i].Content,
		})
	}

	return result
}
