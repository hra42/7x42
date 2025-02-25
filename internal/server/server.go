package server

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	fiberwebsocket "github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/websocket"
	"gorm.io/gorm"
)

type Server struct {
	app       *fiber.App
	db        *gorm.DB
	wsManager *websocket.Manager
}

type Config struct {
	DB *gorm.DB
}

func New(config *Config) *Server {
	// Setup template engine
	viewEngine := html.New("./web/templates", ".html")

	// Create new Fiber app with custom error handler
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

	// Initialize websocket manager
	wsManager := websocket.NewManager()
	wsManager.Start()

	s := &Server{
		app:       app,
		db:        config.DB,
		wsManager: wsManager,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Add global middleware
	s.app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	s.app.Use(recover.New())
	s.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
	}))
}

func (s *Server) setupRoutes() {
	// Static files
	s.app.Static("/static", "./web/static")

	// Health check
	s.app.Get("/health", s.handleHealthCheck)

	// Main routes
	s.app.Get("/", s.handleIndex)
	s.app.Get("/chat", s.handleChat)
	s.app.Get("/settings", s.handleSettings)

	// API routes
	api := s.app.Group("/api")
	v1 := api.Group("/v1")
	chat := v1.Group("/chat")
	chat.Get("/", s.handleListChats)
	chat.Post("/", s.handleCreateChat)
	chat.Get("/:id", s.handleGetChat)
	chat.Post("/:id/messages", s.handleSendMessage)

	// WebSocket setup
	s.app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberwebsocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket endpoint
	s.app.Get("/ws/:userId", fiberwebsocket.New(func(c *fiberwebsocket.Conn) {
		userId := c.Params("userId")
		s.wsManager.HandleConnection(c, userId)
	}))
}

func (s *Server) handleIndex(c *fiber.Ctx) error {
	return c.Render("base", fiber.Map{
		"Title": "7x42 Home",
	})
}

func (s *Server) handleChat(c *fiber.Ctx) error {
	// Render the base template with chat content
	return c.Render("base", fiber.Map{
		"Title":   "7x42 Chat",
		"Content": "chat", // This tells the template to use the chat.html content
	})
}

func (s *Server) handleSettings(c *fiber.Ctx) error {
	return c.Render("base", fiber.Map{
		"Title":   "7x42 Settings",
		"Content": "settings", // This tells the template to use the settings.html content
	})
}

func (s *Server) handleHealthCheck(c *fiber.Ctx) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": "database connection error",
		})
	}

	err = sqlDB.Ping()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": "database ping failed",
		})
	}

	return c.JSON(fiber.Map{
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "healthy",
	})
}

func (s *Server) handleListChats(c *fiber.Ctx) error {
	// This would normally fetch chats from the database
	return c.JSON(fiber.Map{
		"chats": []string{},
	})
}

func (s *Server) handleCreateChat(c *fiber.Ctx) error {
	// Implementation for creating a new chat
	return c.JSON(fiber.Map{
		"status": "created",
		"chatId": 123,
	})
}

func (s *Server) handleGetChat(c *fiber.Ctx) error {
	chatId := c.Params("id")
	// Implementation for getting a specific chat
	return c.JSON(fiber.Map{
		"id":    chatId,
		"title": "Sample Chat",
	})
}

func (s *Server) handleSendMessage(c *fiber.Ctx) error {
	chatId := c.Params("id")
	// Implementation for sending a message to a chat
	return c.JSON(fiber.Map{
		"status":    "sent",
		"chatId":    chatId,
		"messageId": 456,
	})
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
