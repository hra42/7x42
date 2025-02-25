package server

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	fiberwebsocket "github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/ai"
	"github.com/hra42/7x42/internal/models"
	"github.com/hra42/7x42/internal/repository"
	"github.com/hra42/7x42/internal/websocket"
	"gorm.io/gorm"
)

type Server struct {
	app       *fiber.App
	db        *gorm.DB
	wsManager *websocket.Manager
	aiService *ai.Service
}

type Config struct {
	DB        *gorm.DB
	AIService *ai.Service
}

func New(config *Config) *Server {
	viewEngine := html.New("./web/templates", ".html")
	app := fiber.New(fiber.Config{
		Views: viewEngine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Create websocket manager
	wsManager := websocket.NewManager(config.AIService)
	wsManager.Start()

	s := &Server{
		app:       app,
		db:        config.DB,
		wsManager: wsManager,
		aiService: config.AIService,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	s.app.Use(recover.New())
	s.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
}

func (s *Server) setupRoutes() {
	s.app.Static("/static", "./web/static")
	s.app.Get("/health", s.handleHealthCheck)
	s.app.Get("/", s.handleIndex)
	s.app.Get("/chat", s.handleChat)
	s.app.Get("/settings", s.handleSettings)

	api := s.app.Group("/api")
	v1 := api.Group("/v1")
	chat := v1.Group("/chat")
	chat.Get("/", s.handleListChats)
	chat.Post("/", s.handleCreateChat)
	chat.Get("/:id", s.handleGetChat)
	chat.Post("/:id/messages", s.handleSendMessage)

	s.app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberwebsocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	s.app.Get("/ws/:userId", fiberwebsocket.New(func(c *fiberwebsocket.Conn) {
		userId := c.Params("userId")
		s.wsManager.HandleConnection(c, userId)
	}))
}

func (s *Server) handleIndex(c *fiber.Ctx) error {
	return c.Render("base", fiber.Map{
		"Title": "7x42 Chat",
	})
}

func (s *Server) handleChat(c *fiber.Ctx) error {
	return c.Render("chat", fiber.Map{})
}

func (s *Server) handleSettings(c *fiber.Ctx) error {
	return c.SendString("Settings page")
}

func (s *Server) handleHealthCheck(c *fiber.Ctx) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": "error",
		})
	}

	err = sqlDB.Ping()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": "error",
		})
	}

	return c.JSON(fiber.Map{
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "ok",
	})
}

func (s *Server) handleListChats(c *fiber.Ctx) error {
	userID := c.Query("userId", "default-user")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chatRepo := repository.NewChatRepository(s.db)
	chats, err := chatRepo.ListChats(ctx, userID, page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result := make([]fiber.Map, len(chats))
	for i, chat := range chats {
		result[i] = fiber.Map{
			"id":          chat.ID,
			"title":       chat.Title,
			"lastMessage": chat.LastMessage,
		}
	}

	return c.JSON(fiber.Map{
		"chats": result,
		"pagination": fiber.Map{
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

func (s *Server) handleCreateChat(c *fiber.Ctx) error {
	var request struct {
		Title  string `json:"title"`
		UserID string `json:"userId"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if request.UserID == "" {
		request.UserID = "default-user"
	}

	chat := &models.Chat{
		Title:  request.Title,
		UserID: request.UserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chatRepo := repository.NewChatRepository(s.db)
	if err := chatRepo.CreateChat(ctx, chat); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    chat.ID,
		"title": chat.Title,
	})
}

func (s *Server) handleGetChat(c *fiber.Ctx) error {
	chatID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chatRepo := repository.NewChatRepository(s.db)
	chat, err := chatRepo.GetChat(ctx, chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Chat not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	messages := make([]fiber.Map, len(chat.Messages))
	for i, msg := range chat.Messages {
		messages[i] = fiber.Map{
			"content":   msg.Content,
			"role":      msg.Role,
			"timestamp": msg.Timestamp,
		}
	}

	return c.JSON(fiber.Map{
		"id":       chat.ID,
		"title":    chat.Title,
		"messages": messages,
	})
}

func (s *Server) handleSendMessage(c *fiber.Ctx) error {
	chatID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chat ID",
		})
	}

	var request struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	message := &models.Message{
		Content:   request.Content,
		Role:      "user",
		ChatID:    chatID,
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chatRepo := repository.NewChatRepository(s.db)
	if err := chatRepo.CreateMessage(ctx, message); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":        message.ID,
		"content":   message.Content,
		"role":      message.Role,
		"timestamp": message.Timestamp,
	})
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
