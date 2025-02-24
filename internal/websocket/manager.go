// internal/websocket/manager.go
package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

type Client struct {
	Conn    *websocket.Conn
	UserID  string
	IsAlive bool
	mu      sync.Mutex
}

type Message struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

type ChatMessage struct {
	ChatID    uint      `json:"chatId"`
	Content   string    `json:"content"`
	Role      string    `json:"role"`
	Timestamp time.Time `json:"timestamp"`
}

type Manager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (m *Manager) Start() {
	go m.run()
}

func (m *Manager) run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client] = true
			m.mu.Unlock()
			log.Printf("Client connected. Total clients: %d", len(m.clients))

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				client.Conn.Close()
			}
			m.mu.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(m.clients))

		case message := <-m.broadcast:
			m.mu.RLock()
			for client := range m.clients {
				client.mu.Lock()
				if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("Error broadcasting to client: %v", err)
					client.IsAlive = false
					m.unregister <- client
				}
				client.mu.Unlock()
			}
			m.mu.RUnlock()
		}
	}
}

func (m *Manager) HandleConnection(c *websocket.Conn, userID string) {
	client := &Client{
		Conn:    c,
		UserID:  userID,
		IsAlive: true,
	}

	// Register new client
	m.register <- client

	// Start ping/pong
	go m.handlePingPong(client)

	// Handle incoming messages
	for {
		messageType, payload, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			var msg Message
			if err := json.Unmarshal(payload, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			// Handle different message types
			switch msg.Type {
			case "chat_message":
				var chatMsg ChatMessage
				if err := json.Unmarshal(msg.Content, &chatMsg); err != nil {
					log.Printf("Error unmarshaling chat message: %v", err)
					continue
				}
				// Broadcast the message to all clients
				m.broadcast <- payload
			}
		}
	}

	// Unregister client on disconnect
	m.unregister <- client
}

func (m *Manager) handlePingPong(client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	client.Conn.SetPongHandler(func(string) error {
		client.mu.Lock()
		client.IsAlive = true
		client.mu.Unlock()
		return nil
	})

	for range ticker.C {
		client.mu.Lock()
		if !client.IsAlive {
			client.mu.Unlock()
			m.unregister <- client
			return
		}

		client.IsAlive = false
		client.mu.Unlock()

		if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}
