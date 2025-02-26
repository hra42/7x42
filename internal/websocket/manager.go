package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/ai"
)

const (
	// DefaultPingInterval is the default interval for sending ping messages
	DefaultPingInterval = 30 * time.Second

	// DefaultIdleTimeout is the default timeout for idle connections
	DefaultIdleTimeout = 5 * time.Minute

	// MaxMessageSize is the maximum message size in bytes
	MaxMessageSize = 1024 * 1024 // 1MB
)

// Manager manages WebSocket connections
type Manager struct {
	// clients is a map of all connected clients
	clients map[*Client]bool

	// register is a channel for registering new clients
	register chan *Client

	// unregister is a channel for unregistering clients
	unregister chan *Client

	// broadcast is a channel for broadcasting messages to all clients
	broadcast chan []byte

	// aiService is the AI service for handling chat messages
	aiService *ai.Service

	// mu protects the manager's fields during concurrent access
	mu sync.RWMutex

	// pingInterval is the interval for sending ping messages
	pingInterval time.Duration

	// idleTimeout is the timeout for idle connections
	idleTimeout time.Duration

	// running indicates if the manager is running
	running bool

	// done is a channel for signaling when the manager is done
	done chan struct{}
}

// ManagerConfig holds configuration for the WebSocket manager
type ManagerConfig struct {
	PingInterval time.Duration
	IdleTimeout  time.Duration
}

// NewManager creates a new WebSocket manager
func NewManager(aiService *ai.Service, config ...*ManagerConfig) *Manager {
	m := &Manager{
		clients:      make(map[*Client]bool),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan []byte),
		aiService:    aiService,
		pingInterval: DefaultPingInterval,
		idleTimeout:  DefaultIdleTimeout,
		done:         make(chan struct{}),
	}

	// Apply config if provided
	if len(config) > 0 && config[0] != nil {
		if config[0].PingInterval > 0 {
			m.pingInterval = config[0].PingInterval
		}
		if config[0].IdleTimeout > 0 {
			m.idleTimeout = config[0].IdleTimeout
		}
	}

	return m
}

// Start starts the WebSocket manager
func (m *Manager) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	go m.run()
	go m.pingClients()
	go m.cleanIdleConnections()

	log.Println("WebSocket manager started")
}

// Stop gracefully stops the WebSocket manager
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.done)

	// Close all client connections
	for client := range m.clients {
		client.Close()
		delete(m.clients, client)
	}

	log.Println("WebSocket manager stopped")
}

// run processes WebSocket events
func (m *Manager) run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client] = true
			m.mu.Unlock()
			log.Printf("Client connected. Total clients: %d", m.countClients())

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				client.Close()
			}
			m.mu.Unlock()
			log.Printf("Client disconnected. Total clients: %d", m.countClients())

		case message := <-m.broadcast:
			m.mu.RLock()
			for client := range m.clients {
				if err := client.SendText(string(message)); err != nil {
					log.Printf("Error broadcasting to client: %v", err)
					m.unregister <- client
				}
			}
			m.mu.RUnlock()

		case <-m.done:
			return
		}
	}
}

// pingClients sends ping messages to all clients periodically
func (m *Manager) pingClients() {
	ticker := time.NewTicker(m.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.RLock()
			for client := range m.clients {
				if err := client.SendPing(); err != nil {
					log.Printf("Error sending ping to client: %v", err)
					m.unregister <- client
				}
			}
			m.mu.RUnlock()

		case <-m.done:
			return
		}
	}
}

// cleanIdleConnections closes idle connections
func (m *Manager) cleanIdleConnections() {
	ticker := time.NewTicker(m.idleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.RLock()
			for client := range m.clients {
				if client.IsIdle(m.idleTimeout) {
					log.Printf("Closing idle connection for user %s", client.UserID)
					m.unregister <- client
				}
			}
			m.mu.RUnlock()

		case <-m.done:
			return
		}
	}
}

// HandleConnection handles a new WebSocket connection
func (m *Manager) HandleConnection(conn *websocket.Conn, userID string) {
	// Create a new client
	client := NewClient(conn, userID)

	// Set read limit to prevent malicious messages
	conn.SetReadLimit(MaxMessageSize)

	// Register the client
	m.register <- client

	// Handle client messages
	m.handleClientMessages(client)

	// Unregister the client when done
	m.unregister <- client
}

// handleClientMessages processes messages from a client
func (m *Manager) handleClientMessages(client *Client) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in handleClientMessages: %v", r)
		}
	}()

	for {
		messageType, payload, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		client.UpdateActivity()

		// Handle different message types
		switch messageType {
		case websocket.TextMessage:
			if err := m.handleTextMessage(client, payload); err != nil {
				log.Printf("Error handling text message: %v", err)
				if err := client.SendJSON(NewErrorMessage(err.Error(), "message_error")); err != nil {
					log.Printf("Error sending error message: %v", err)
				}
			}

		case websocket.PingMessage:
			if err := client.SendText(`{"type":"pong"}`); err != nil {
				log.Printf("Error sending pong: %v", err)
			}

		case websocket.CloseMessage:
			return
		}
	}
}

// handleTextMessage processes text messages from a client
func (m *Manager) handleTextMessage(client *Client, payload []byte) error {
	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return NewError("unmarshal", ErrInvalidMessage, "invalid_format")
	}

	switch msg.Type {
	case TypeChatMessage:
		return m.handleChatMessage(client, msg.Content)

	case TypePing:
		return client.SendJSON(map[string]string{"type": "pong"})

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleChatMessage processes chat messages
func (m *Manager) handleChatMessage(client *Client, content json.RawMessage) error {
	var rawChatMsg ChatMessageRaw
	if err := json.Unmarshal(content, &rawChatMsg); err != nil {
		return NewError("unmarshal", ErrInvalidMessage, "invalid_chat_format")
	}

	chatID, err := ParseChatID(rawChatMsg.ChatID)
	if err != nil {
		return NewError("parse_chat_id", ErrInvalidChatID, "invalid_chat_id")
	}

	chatMsg := ChatMessage{
		ChatID:    chatID,
		Content:   rawChatMsg.Content,
		Role:      rawChatMsg.Role,
		Timestamp: rawChatMsg.Timestamp,
	}

	// Process the chat message with the AI service
	if err := m.aiService.HandleChatMessage(client.Conn, chatMsg.ChatID, chatMsg.Content, client.UserID); err != nil {
		return NewError("ai_service", err, "ai_service_error")
	}

	return nil
}

// BroadcastToUser broadcasts a message to a specific user
func (m *Manager) BroadcastToUser(userID string, message interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for client := range m.clients {
		if client.UserID == userID {
			if err := client.SendJSON(message); err != nil {
				log.Printf("Error sending message to user %s: %v", userID, err)
			}
		}
	}
}

// Broadcast broadcasts a message to all clients
func (m *Manager) Broadcast(message interface{}) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	m.broadcast <- jsonMessage
}

// countClients returns the number of connected clients
func (m *Manager) countClients() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.clients)
}

// GetClientCount returns the number of connected clients
func (m *Manager) GetClientCount() int {
	return m.countClients()
}

// GetClientByUserID returns a client by user ID
func (m *Manager) GetClientByUserID(userID string) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for client := range m.clients {
		if client.UserID == userID {
			return client
		}
	}

	return nil
}
