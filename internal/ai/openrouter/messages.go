package openrouter

import (
	"context"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/hra42/7x42/internal/models"
)

// prepareMessages formats the messages for the API request
func (c *Client) prepareMessages(prompt string, chatHistory []ChatMessage) []ChatMessage {
	if len(chatHistory) == 0 {
		return []ChatMessage{
			{Role: "user", Content: prompt},
		}
	}

	// Append the new prompt to existing chat history
	return append(chatHistory, ChatMessage{Role: "user", Content: prompt})
}

// extractChatID extracts the chat ID from the prompt if available
func (c *Client) extractChatID(prompt string) uint64 {
	chatIDMatches := regexp.MustCompile(`chat:(\d+)`).FindStringSubmatch(prompt)
	if len(chatIDMatches) > 1 {
		if id, err := strconv.ParseUint(chatIDMatches[1], 10, 64); err == nil {
			return id
		}
	}
	return 0
}

func (c *Client) saveMessage(ctx context.Context, chatID uint64, content string, processingTime int) {
	message := &models.Message{
		ChatID:    chatID,
		Content:   content,
		Role:      "assistant",
		Timestamp: time.Now(),
		Metadata: models.MessageMetadata{
			Model:       c.config.Model,
			TokenCount:  len(content) / 4,
			ProcessTime: processingTime,
		},
	}

	// Use messageRepo instead of chatRepo
	if err := c.messageRepo.CreateMessage(ctx, message); err != nil {
		log.Printf("Failed to save AI response: %v", err)
	}
}
