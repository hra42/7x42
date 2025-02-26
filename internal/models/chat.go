package models

import (
	"time"

	"gorm.io/gorm"
)

// Chat represents a conversation between a user and the AI
type Chat struct {
	gorm.Model
	Title       string    `gorm:"type:varchar(255);not null"`
	Messages    []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
	LastMessage time.Time `gorm:"index"`
	UserID      string    `gorm:"type:varchar(255);index"`
}

// BeforeCreate is a GORM hook that sets default values before creating a chat
func (c *Chat) BeforeCreate(tx *gorm.DB) error {
	// Set default LastMessage time if not set
	if c.LastMessage.IsZero() {
		c.LastMessage = time.Now()
	}
	return nil
}

// AddMessage adds a new message to the chat and updates LastMessage time
func (c *Chat) AddMessage(message *Message) {
	c.Messages = append(c.Messages, *message)
	c.LastMessage = message.Timestamp
}

// GetRecentMessages returns the most recent n messages from the chat
func (c *Chat) GetRecentMessages(limit int) []Message {
	if len(c.Messages) <= limit {
		return c.Messages
	}
	return c.Messages[len(c.Messages)-limit:]
}

// Summary returns a brief summary of the chat
func (c *Chat) Summary() map[string]interface{} {
	var lastMessageContent string
	if len(c.Messages) > 0 {
		lastMsg := c.Messages[len(c.Messages)-1]
		if len(lastMsg.Content) > 50 {
			lastMessageContent = lastMsg.Content[:50] + "..."
		} else {
			lastMessageContent = lastMsg.Content
		}
	}

	return map[string]interface{}{
		"id":           c.ID,
		"title":        c.Title,
		"lastMessage":  c.LastMessage,
		"messageCount": len(c.Messages),
		"preview":      lastMessageContent,
	}
}
