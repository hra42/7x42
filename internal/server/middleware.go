package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberws "github.com/gofiber/websocket/v2"
	"log"
	"os"
	"time"
)

// setupMiddleware configures the middleware for the server
func (s *Server) setupMiddleware() {
	// Logger middleware
	s.app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))

	// Recover middleware
	s.app.Use(recover.New())

	// CORS middleware
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
	}))

	// Serve static files with improved logging
	s.app.Use("/static", func(c *fiber.Ctx) error {
		// Log static file requests in development
		if os.Getenv("GO_ENV") != "production" {
			log.Printf("Static file requested: %s", c.Path())
		}
		return c.Next()
	})

	// Set up static file serving with proper configuration
	s.app.Static("/static", "./web/static", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 24 * time.Hour,
		MaxAge:        86400,
	})
}

// WebSocketMiddleware is a middleware that upgrades the connection to WebSocket
func WebSocketMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}
