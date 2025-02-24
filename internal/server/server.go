package server

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
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
	app := fiber.New(fiber.Config{
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
	s.app.Get("/health", s.handleHealthCheck)

	api := s.app.Group("/api")
	v1 := api.Group("/v1")
	chat := v1.Group("/chat")
	chat.Get("/", s.handleListChats)

	s.app.Static("/", "./web/static")
	s.app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./web/templates/chat.html")
	})

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

func (s *Server) handleHealthCheck(c *fiber.Ctx) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error",
			"error":  err.Error(),
		})
	}

	err = sqlDB.Ping()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error",
			"error":  err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleListChats(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"chats": []string{},
	})
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
