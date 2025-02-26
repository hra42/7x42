package repository

import (
	"context"
	"errors"

	"github.com/hra42/7x42/internal/models"
	"gorm.io/gorm"
)

// MessageRepository handles database operations for message entities
type MessageRepository struct {
	*BaseRepository
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// CreateMessage creates a new message and updates the associated chat's last_message timestamp
func (r *MessageRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	return RunInTransaction(r.DB().WithContext(ctx), func(tx *gorm.DB) error {
		// Create the message
		if err := tx.Create(message).Error; err != nil {
			return NewError("create", "message", err)
		}

		// Update the chat's last_message timestamp
		result := tx.Model(&models.Chat{}).
			Where("id = ?", message.ChatID).
			Update("last_message", message.Timestamp)

		if result.Error != nil {
			return NewError("update", "chat.last_message", result.Error)
		}

		// If no chat was updated, the chat may not exist
		if result.RowsAffected == 0 {
			return NewError("update", "chat.last_message", ErrNotFound)
		}

		return nil
	})
}

// GetMessage retrieves a message by ID
func (r *MessageRepository) GetMessage(ctx context.Context, id uint) (*models.Message, error) {
	var message models.Message

	err := r.DB().WithContext(ctx).First(&message, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewError("get", "message", ErrNotFound)
		}
		return nil, NewError("get", "message", err)
	}

	return &message, nil
}

// GetChatMessages retrieves messages for a chat with pagination
func (r *MessageRepository) GetChatMessages(ctx context.Context, chatID uint64, page, pageSize int) ([]models.Message, error) {
	var messages []models.Message
	offset := (page - 1) * pageSize

	err := r.DB().WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("timestamp DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error

	if err != nil {
		return nil, NewError("list", "messages", err)
	}

	return messages, nil
}

// CountChatMessages counts the number of messages in a chat
func (r *MessageRepository) CountChatMessages(ctx context.Context, chatID uint64) (int64, error) {
	var count int64

	err := r.DB().WithContext(ctx).
		Model(&models.Message{}).
		Where("chat_id = ?", chatID).
		Count(&count).Error

	if err != nil {
		return 0, NewError("count", "messages", err)
	}

	return count, nil
}

// DeleteChatMessages deletes all messages for a chat
func (r *MessageRepository) DeleteChatMessages(ctx context.Context, chatID uint64) error {
	result := r.DB().WithContext(ctx).
		Where("chat_id = ?", chatID).
		Delete(&models.Message{})

	if result.Error != nil {
		return NewError("delete", "messages", result.Error)
	}

	return nil
}
