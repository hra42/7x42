package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/hra42/7x42/internal/ai"
	"github.com/hra42/7x42/internal/server/handlers"
	"github.com/hra42/7x42/internal/websocket"
	"gorm.io/gorm"
)

// Server represents the HTTP server
type Server struct {
	app       *fiber.App
	db        *gorm.DB
	wsManager *websocket.Manager
	aiService *ai.Service
}

// Config holds the server configuration
type Config struct {
	DB        *gorm.DB
	AIService *ai.Service
}

// New creates a new server instance
func New(config *Config) *Server {
	// Initialize template engine
	viewEngine := html.New("./web/templates", ".html")

	// Create Fiber app with custom error handler
	app := fiber.New(fiber.Config{
		Views:        viewEngine,
		ViewsLayout:  "base",
		ErrorHandler: handlers.ErrorHandler,
	})

	// Create WebSocket manager
	wsManager := websocket.NewManager(config.AIService)
	wsManager.Start()

	// Create server instance
	s := &Server{
		app:       app,
		db:        config.DB,
		wsManager: wsManager,
		aiService: config.AIService,
	}

	// Setup middleware and routes
	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// Listen starts the HTTP server
func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	// Stop the WebSocket manager
	s.wsManager.Stop()

	// Shutdown the Fiber app
	return s.app.Shutdown()
}
