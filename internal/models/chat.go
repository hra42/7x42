package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	Title       string    `gorm:"type:varchar(255);not null"`
	Messages    []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
	LastMessage time.Time `gorm:"index"`
	UserID      string    `gorm:"type:varchar(255);index"`
}

type MessageMetadata struct {
	Model       string `json:"model,omitempty"`
	TokenCount  int    `json:"token_count,omitempty"`
	ProcessTime int    `json:"process_time,omitempty"`
}

// Value implements the driver.Valuer interface for database serialization
func (m MessageMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database deserialization
func (m *MessageMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &m)
}

type Message struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt  `gorm:"index"`
	Content   string          `gorm:"type:text;not null"`
	Role      string          `gorm:"type:varchar(20);not null;check:role IN ('user', 'assistant')"`
	ChatID    uint64          `gorm:"index;not null"`
	Timestamp time.Time       `gorm:"index;not null;default:CURRENT_TIMESTAMP"`
	Metadata  MessageMetadata `gorm:"type:jsonb"`
}

func (m *Message) BeforeCreate() error {
	m.Timestamp = time.Now()
	return nil
}

func (c *Chat) BeforeCreate() error {
	c.LastMessage = time.Now()
	return nil
}
