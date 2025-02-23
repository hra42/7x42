package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	app *fiber.App
}

// New creates a new server instance with configured middleware
func New() *Server {
	app := fiber.New(fiber.Config{
		// Enable server-side WebSocket compression
		EnableTrustedProxyCheck: true,
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Create server instance
	s := &Server{
		app: app,
	}

	// Setup routes
	s.setupRoutes()

	return s
}

// Listen starts the server
func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

// setupRoutes configures all the routes for the server
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// API routes group
	api := s.app.Group("/api")

	// Version 1 routes
	v1 := api.Group("/v1")

	// Chat endpoints will be added here
	v1.Get("/chat", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Chat endpoint placeholder",
		})
	})

	// Serve static files
	s.app.Static("/", "./web/static")

	// Serve templates
	s.app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./web/templates/chat.html")
	})
}
