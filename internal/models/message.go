package models

import (
	"time"

	"gorm.io/gorm"
)

// Role constants for message roles
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

// Message represents a single message in a chat
type Message struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt  `gorm:"index"`
	Content   string          `gorm:"type:text;not null"`
	Role      string          `gorm:"type:varchar(20);not null;check:role IN ('user', 'assistant', 'system')"`
	ChatID    uint64          `gorm:"index;not null"`
	Timestamp time.Time       `gorm:"index;not null;default:CURRENT_TIMESTAMP"`
	Metadata  MessageMetadata `gorm:"type:jsonb"`
}

// BeforeCreate is a GORM hook that sets default values before creating a message
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	// Set default timestamp if not set
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}

	// Validate message role
	switch m.Role {
	case RoleUser, RoleAssistant, RoleSystem:
		// Valid roles
	default:
		m.Role = RoleSystem // Default to system role if invalid
	}

	return nil
}

// IsFromUser returns true if the message is from a user
func (m *Message) IsFromUser() bool {
	return m.Role == RoleUser
}

// IsFromAssistant returns true if the message is from the assistant
func (m *Message) IsFromAssistant() bool {
	return m.Role == RoleAssistant
}

// ToMap converts the message to a map for API responses
func (m *Message) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":        m.ID,
		"content":   m.Content,
		"role":      m.Role,
		"chatId":    m.ChatID,
		"timestamp": m.Timestamp,
		"metadata":  m.Metadata,
	}
}
