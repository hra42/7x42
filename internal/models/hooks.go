package models

import (
	"time"

	"gorm.io/gorm"
)

// RegisterHooks registers all GORM hooks for models
func RegisterHooks(db *gorm.DB) {
	// Chat hooks
	db.Callback().Create().Before("gorm:create").Register("chats:before_create", beforeCreateChat)
	db.Callback().Update().Before("gorm:update").Register("chats:before_update", beforeUpdateChat)

	// Message hooks
	db.Callback().Create().Before("gorm:create").Register("messages:before_create", beforeCreateMessage)
	db.Callback().Create().After("gorm:create").Register("messages:after_create", afterCreateMessage)
}

// Chat hooks
func beforeCreateChat(db *gorm.DB) {
	if chat, ok := db.Statement.Dest.(*Chat); ok {
		if chat.LastMessage.IsZero() {
			chat.LastMessage = time.Now()
		}
	}
}

func beforeUpdateChat(db *gorm.DB) {
	// Add any common update logic here
}

// Message hooks
func beforeCreateMessage(db *gorm.DB) {
	if message, ok := db.Statement.Dest.(*Message); ok {
		if message.Timestamp.IsZero() {
			message.Timestamp = time.Now()
		}
	}
}

func afterCreateMessage(db *gorm.DB) {
	if message, ok := db.Statement.Dest.(*Message); ok {
		// Update the chat's LastMessage time when a message is created
		db.Model(&Chat{}).
			Where("id = ?", message.ChatID).
			Update("last_message", message.Timestamp)
	}
}
