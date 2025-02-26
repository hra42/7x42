package server

import (
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/repository"
	"github.com/hra42/7x42/internal/server/handlers"
)

// setupRoutes configures the routes for the server
func (s *Server) setupRoutes() {
	// Create repositories
	chatRepo := repository.NewChatRepository(s.db)
	messageRepo := repository.NewMessageRepository(s.db)

	// Create handlers
	healthHandler := handlers.NewHealthHandler(s.db)
	chatHandler := handlers.NewChatHandler(chatRepo, messageRepo)
	pageHandler := handlers.NewPageHandler()
	wsHandler := handlers.NewWebSocketHandler(s.wsManager)

	// Health routes
	s.app.Get("/health", healthHandler.Check)

	// Page routes
	s.app.Get("/", pageHandler.Index)
	s.app.Get("/chat", pageHandler.Chat)
	s.app.Get("/settings", pageHandler.Settings)

	// API routes
	api := s.app.Group("/api")
	v1 := api.Group("/v1")

	// Chat routes
	chat := v1.Group("/chat")
	chat.Get("/", chatHandler.List)
	chat.Post("/", chatHandler.Create)
	chat.Get("/:id", chatHandler.Get)
	chat.Put("/:id", chatHandler.Update)
	chat.Delete("/:id", chatHandler.Delete)
	chat.Post("/:id/messages", chatHandler.SendMessage)
	chat.Get("/:id/messages", chatHandler.ListMessages)

	// WebSocket routes
	s.app.Use("/ws", WebSocketMiddleware())
	s.app.Get("/ws/:userId", fiberws.New(wsHandler.HandleConnection))
}
