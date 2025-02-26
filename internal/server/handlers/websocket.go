package handlers

import (
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/websocket"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	manager *websocket.Manager
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(manager *websocket.Manager) *WebSocketHandler {
	return &WebSocketHandler{
		manager: manager,
	}
}

// HandleConnection handles a WebSocket connection
func (h *WebSocketHandler) HandleConnection(c *fiberws.Conn) {
	userId := c.Params("userId")
	h.manager.HandleConnection(c, userId)
}
