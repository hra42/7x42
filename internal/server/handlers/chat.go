package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hra42/7x42/internal/models"
	"github.com/hra42/7x42/internal/repository"
	"github.com/hra42/7x42/internal/server/responses"
)

// ChatHandler handles chat-related requests
type ChatHandler struct {
	chatRepo    *repository.ChatRepository
	messageRepo *repository.MessageRepository
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatRepo *repository.ChatRepository, messageRepo *repository.MessageRepository) *ChatHandler {
	return &ChatHandler{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
	}
}

// List handles the list chats endpoint
func (h *ChatHandler) List(c *fiber.Ctx) error {
	userID := GetUserID(c)
	page, pageSize := ParsePagination(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get chats
	chats, err := h.chatRepo.ListChats(ctx, userID, page, pageSize)
	if err != nil {
		return err
	}

	// Get total count
	total, err := h.chatRepo.CountChats(ctx, userID)
	if err != nil {
		return err
	}

	// Format response
	result := make([]fiber.Map, len(chats))
	for i, chat := range chats {
		result[i] = fiber.Map{
			"id":          chat.ID,
			"title":       chat.Title,
			"lastMessage": chat.LastMessage,
			"createdAt":   chat.CreatedAt,
		}
	}

	return responses.JSON(c, fiber.StatusOK, fiber.Map{
		"chats": result,
		"pagination": fiber.Map{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// Get handles the get chat endpoint
func (h *ChatHandler) Get(c *fiber.Ctx) error {
	chatID, err := ParseUint64Param(c, "id")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get chat
	chat, err := h.chatRepo.GetChat(ctx, chatID)
	if err != nil {
		return err
	}

	// Format messages
	messages := make([]fiber.Map, len(chat.Messages))
	for i, msg := range chat.Messages {
		messages[i] = fiber.Map{
			"id":        msg.ID,
			"content":   msg.Content,
			"role":      msg.Role,
			"timestamp": msg.Timestamp,
			"metadata":  msg.Metadata,
		}
	}

	return responses.JSON(c, fiber.StatusOK, fiber.Map{
		"id":          chat.ID,
		"title":       chat.Title,
		"createdAt":   chat.CreatedAt,
		"lastMessage": chat.LastMessage,
		"messages":    messages,
	})
}

// Create handles the create chat endpoint
func (h *ChatHandler) Create(c *fiber.Ctx) error {
	type request struct {
		Title  string `json:"title"`
		UserID string `json:"userId"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	userID := req.UserID
	if userID == "" {
		userID = GetUserID(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create chat
	chat := &models.Chat{
		Title:  req.Title,
		UserID: userID,
	}

	if err := h.chatRepo.CreateChat(ctx, chat); err != nil {
		return err
	}

	return responses.JSON(c, fiber.StatusCreated, fiber.Map{
		"id":        chat.ID,
		"title":     chat.Title,
		"createdAt": chat.CreatedAt,
	})
}

// Update handles the update chat endpoint
func (h *ChatHandler) Update(c *fiber.Ctx) error {
	chatID, err := ParseUint64Param(c, "id")
	if err != nil {
		return err
	}

	type request struct {
		Title string `json:"title"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get chat
	chat, err := h.chatRepo.GetChat(ctx, chatID)
	if err != nil {
		return err
	}

	// Update chat
	chat.Title = req.Title
	if err := h.chatRepo.UpdateChat(ctx, chat); err != nil {
		return err
	}

	return responses.JSON(c, fiber.StatusOK, fiber.Map{
		"id":        chat.ID,
		"title":     chat.Title,
		"updatedAt": chat.UpdatedAt,
	})
}

// Delete handles the delete chat endpoint
func (h *ChatHandler) Delete(c *fiber.Ctx) error {
	chatID, err := ParseUint64Param(c, "id")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Delete chat
	if err := h.chatRepo.DeleteChat(ctx, chatID); err != nil {
		return err
	}

	return responses.JSON(c, fiber.StatusOK, fiber.Map{
		"success": true,
	})
}

// SendMessage handles the send message endpoint
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	chatID, err := ParseUint64Param(c, "id")
	if err != nil {
		return err
	}

	type request struct {
		Content string `json:"content"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create message
	message := &models.Message{
		ChatID:    chatID,
		Content:   req.Content,
		Role:      models.RoleUser,
		Timestamp: time.Now(),
	}

	if err := h.messageRepo.CreateMessage(ctx, message); err != nil {
		return err
	}

	return responses.JSON(c, fiber.StatusCreated, fiber.Map{
		"id":        message.ID,
		"content":   message.Content,
		"role":      message.Role,
		"timestamp": message.Timestamp,
	})
}

// ListMessages handles the list messages endpoint
func (h *ChatHandler) ListMessages(c *fiber.Ctx) error {
	chatID, err := ParseUint64Param(c, "id")
	if err != nil {
		return err
	}

	page, pageSize := ParsePagination(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get messages
	messages, err := h.messageRepo.GetChatMessages(ctx, chatID, page, pageSize)
	if err != nil {
		return err
	}

	// Get total count
	total, err := h.messageRepo.CountChatMessages(ctx, chatID)
	if err != nil {
		return err
	}

	// Format response
	result := make([]fiber.Map, len(messages))
	for i, msg := range messages {
		result[i] = fiber.Map{
			"id":        msg.ID,
			"content":   msg.Content,
			"role":      msg.Role,
			"timestamp": msg.Timestamp,
			"metadata":  msg.Metadata,
		}
	}

	return responses.JSON(c, fiber.StatusOK, fiber.Map{
		"messages": result,
		"pagination": fiber.Map{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}
