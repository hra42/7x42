package repository

import (
	"context"
	"github.com/hra42/7x42/internal/models"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateChat creates a new chat
func (r *ChatRepository) CreateChat(ctx context.Context, chat *models.Chat) error {
	return r.db.WithContext(ctx).Create(chat).Error
}

// GetChat retrieves a chat by ID with its messages
func (r *ChatRepository) GetChat(ctx context.Context, id uint) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.timestamp ASC")
		}).
		First(&chat, id).Error
	return &chat, err
}

// CreateMessage adds a new message to a chat
func (r *ChatRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the message
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		// Update chat's last message timestamp
		return tx.Model(&models.Chat{}).
			Where("id = ?", message.ChatID).
			Update("last_message", message.Timestamp).Error
	})
	return err
}

// GetChatMessages retrieves paginated messages for a chat
func (r *ChatRepository) GetChatMessages(ctx context.Context, chatID uint, page, pageSize int) ([]models.Message, error) {
	var messages []models.Message
	offset := (page - 1) * pageSize

	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("timestamp DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error

	return messages, err
}

// ListChats retrieves a paginated list of chats
func (r *ChatRepository) ListChats(ctx context.Context, userID string, page, pageSize int) ([]models.Chat, error) {
	var chats []models.Chat
	offset := (page - 1) * pageSize

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("last_message DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&chats).Error

	return chats, err
}
