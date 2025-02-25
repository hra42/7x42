package websocket

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/ai"
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

// ChatMessageRaw is used for unmarshaling when the chatId might be a string
type ChatMessageRaw struct {
	ChatID    interface{} `json:"chatId"`
	Content   string      `json:"content"`
	Role      string      `json:"role"`
	Timestamp time.Time   `json:"timestamp"`
}

type Manager struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	aiService  *ai.Service
	mu         sync.RWMutex
}

func NewManager(aiService *ai.Service) *Manager {
	return &Manager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		aiService:  aiService,
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

	m.register <- client
	go m.handlePingPong(client)

	defer func() {
		m.unregister <- client
	}()

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

			switch msg.Type {
			case "chat_message":
				var rawChatMsg ChatMessageRaw
				if err := json.Unmarshal(msg.Content, &rawChatMsg); err != nil {
					log.Printf("Error unmarshaling chat message: %v", err)
					continue
				}

				// Convert chatId to uint regardless of whether it's a string or number
				var chatID uint
				switch v := rawChatMsg.ChatID.(type) {
				case float64:
					chatID = uint(v)
				case int:
					chatID = uint(v)
				case string:
					if id, err := strconv.ParseUint(v, 10, 32); err == nil {
						chatID = uint(id)
					} else {
						log.Printf("Invalid chat ID format: %v", v)
						continue
					}
				default:
					log.Printf("Unexpected chatId type: %T", rawChatMsg.ChatID)
					continue
				}

				// Create a proper ChatMessage with the converted chatID
				chatMsg := ChatMessage{
					ChatID:    chatID,
					Content:   rawChatMsg.Content,
					Role:      rawChatMsg.Role,
					Timestamp: rawChatMsg.Timestamp,
				}

				if err := m.aiService.HandleChatMessage(c, chatMsg.ChatID, chatMsg.Content, client.UserID); err != nil {
					log.Printf("Error handling chat message: %v", err)
					errorMsg := map[string]interface{}{
						"type":    "error",
						"content": err.Error(),
					}
					if err := client.Conn.WriteJSON(errorMsg); err != nil {
						log.Printf("Error sending error message: %v", err)
					}
				}
			case "ping":
				if err := client.Conn.WriteJSON(map[string]string{"type": "pong"}); err != nil {
					log.Printf("Error sending pong: %v", err)
				}
			}
		}
	}
}

func (m *Manager) handlePingPong(client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			client.mu.Lock()
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				client.mu.Unlock()
				return
			}
			client.mu.Unlock()
		}
	}
}
