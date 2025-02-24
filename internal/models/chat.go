package models

import (
	"gorm.io/gorm"
	"time"
)

type Chat struct {
	gorm.Model
	Title       string    `gorm:"type:varchar(255);not null"`
	Messages    []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
	LastMessage time.Time `gorm:"index"`
	UserID      string    `gorm:"type:varchar(255);index"` // For future authentication
}

type Message struct {
	gorm.Model
	Content   string    `gorm:"type:text;not null"`
	Role      string    `gorm:"type:varchar(20);not null;check:role IN ('user', 'assistant')"`
	ChatID    uint      `gorm:"index;not null"`
	Timestamp time.Time `gorm:"index;not null;default:CURRENT_TIMESTAMP"`
	Metadata  string    `gorm:"type:jsonb"` // For storing additional AI-related metadata
}

// Implement GORM hooks for automatic timestamp management
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	m.Timestamp = time.Now()
	return nil
}

func (c *Chat) BeforeCreate(tx *gorm.DB) error {
	c.LastMessage = time.Now()
	return nil
}
