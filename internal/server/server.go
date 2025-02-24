package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
	"time"
)

type Server struct {
	app *fiber.App
	db  *gorm.DB
}

type Config struct {
	DB *gorm.DB
}

func New(config *Config) *Server {
	// Initialize Fiber with proper configuration
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  5 * time.Minute,
		// Customize error handling
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Initialize server instance
	s := &Server{
		app: app,
		db:  config.DB,
	}

	// Setup middleware
	s.setupMiddleware()
	// Setup routes
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Add logger middleware
	s.app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} | ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))

	// Add recovery middleware
	s.app.Use(recover.New())

	// Add CORS middleware
	s.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
}

func (s *Server) setupRoutes() {
	// Health check endpoint
	s.app.Get("/health", s.handleHealthCheck)

	// API routes group
	api := s.app.Group("/api")
	v1 := api.Group("/v1")

	// Chat routes will be added here in the WebSocket task
	chat := v1.Group("/chat")
	chat.Get("/", s.handleListChats)
	// More routes will be added later

	// Serve static files
	s.app.Static("/", "./web/static")

	// Serve the main chat interface
	s.app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./web/templates/chat.html")
	})
}

func (s *Server) handleHealthCheck(c *fiber.Ctx) error {
	// Check database connection
	sqlDB, err := s.db.DB()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Database connection error",
		})
	}

	// Ping database
	err = sqlDB.Ping()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Database ping failed",
		})
	}

	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleListChats(c *fiber.Ctx) error {
	// This is a placeholder that will be implemented in the WebSocket task
	return c.JSON(fiber.Map{
		"message": "Chat listing endpoint - to be implemented",
	})
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
