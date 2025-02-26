package repository

import (
	"context"
	"errors"

	"github.com/hra42/7x42/internal/models"
	"gorm.io/gorm"
)

// ChatRepository handles database operations for chat entities
type ChatRepository struct {
	*BaseRepository
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// CreateChat creates a new chat
func (r *ChatRepository) CreateChat(ctx context.Context, chat *models.Chat) error {
	db := r.DB().WithContext(ctx)

	if err := db.Create(chat).Error; err != nil {
		return NewError("create", "chat", err)
	}

	return nil
}

// GetChat retrieves a chat by ID with its messages
func (r *ChatRepository) GetChat(ctx context.Context, id uint64) (*models.Chat, error) {
	var chat models.Chat

	err := r.DB().WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.timestamp ASC")
		}).
		First(&chat, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewError("get", "chat", ErrNotFound)
		}
		return nil, NewError("get", "chat", err)
	}

	return &chat, nil
}

// GetChatByUser retrieves a chat by user ID
func (r *ChatRepository) GetChatByUser(ctx context.Context, userID string, chatID uint64) (*models.Chat, error) {
	var chat models.Chat

	err := r.DB().WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.timestamp ASC")
		}).
		Where("id = ? AND user_id = ?", chatID, userID).
		First(&chat).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewError("get", "chat", ErrNotFound)
		}
		return nil, NewError("get", "chat", err)
	}

	return &chat, nil
}

// UpdateChat updates a chat
func (r *ChatRepository) UpdateChat(ctx context.Context, chat *models.Chat) error {
	result := r.DB().WithContext(ctx).
		Model(chat).
		Updates(map[string]interface{}{
			"title":        chat.Title,
			"last_message": chat.LastMessage,
		})

	if result.Error != nil {
		return NewError("update", "chat", result.Error)
	}

	if result.RowsAffected == 0 {
		return NewError("update", "chat", ErrNotFound)
	}

	return nil
}

// DeleteChat deletes a chat
func (r *ChatRepository) DeleteChat(ctx context.Context, id uint64) error {
	result := r.DB().WithContext(ctx).
		Delete(&models.Chat{}, id)

	if result.Error != nil {
		return NewError("delete", "chat", result.Error)
	}

	if result.RowsAffected == 0 {
		return NewError("delete", "chat", ErrNotFound)
	}

	return nil
}

// ListChats lists chats for a user
func (r *ChatRepository) ListChats(ctx context.Context, userID string, page, pageSize int) ([]models.Chat, error) {
	var chats []models.Chat
	offset := (page - 1) * pageSize

	err := r.DB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("last_message DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&chats).Error

	if err != nil {
		return nil, NewError("list", "chats", err)
	}

	return chats, nil
}

// CountChats counts the number of chats for a user
func (r *ChatRepository) CountChats(ctx context.Context, userID string) (int64, error) {
	var count int64

	err := r.DB().WithContext(ctx).
		Model(&models.Chat{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		return 0, NewError("count", "chats", err)
	}

	return count, nil
}
