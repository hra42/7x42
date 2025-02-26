package websocket

import (
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

// ClientStatus represents the status of a WebSocket client
type ClientStatus int

const (
	// StatusConnected indicates the client is connected and ready
	StatusConnected ClientStatus = iota
	// StatusDisconnecting indicates the client is in the process of disconnecting
	StatusDisconnecting
	// StatusDisconnected indicates the client has disconnected
	StatusDisconnected
	// StatusError indicates the client encountered an error
	StatusError
)

// Client represents a WebSocket client connection
type Client struct {
	// Conn is the WebSocket connection
	Conn *websocket.Conn
	// UserID is the unique identifier for the user
	UserID string
	// Status is the current status of the client
	Status ClientStatus
	// LastActivity is the timestamp of the last activity
	LastActivity time.Time
	// mu protects the client's fields during concurrent access
	mu sync.Mutex
	// ErrorCount tracks the number of consecutive errors
	ErrorCount int
	// Metadata stores additional client information
	Metadata map[string]interface{}
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, userID string) *Client {
	return &Client{
		Conn:         conn,
		UserID:       userID,
		Status:       StatusConnected,
		LastActivity: time.Now(),
		Metadata:     make(map[string]interface{}),
	}
}

// SendJSON sends a JSON message to the client
func (c *Client) SendJSON(data interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Status != StatusConnected {
		return ErrClientDisconnected
	}

	c.LastActivity = time.Now()
	if err := c.Conn.WriteJSON(data); err != nil {
		c.ErrorCount++
		return err
	}

	c.ErrorCount = 0
	return nil
}

// SendText sends a text message to the client
func (c *Client) SendText(message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Status != StatusConnected {
		return ErrClientDisconnected
	}

	c.LastActivity = time.Now()
	if err := c.Conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		c.ErrorCount++
		return err
	}

	c.ErrorCount = 0
	return nil
}

// SendPing sends a ping message to the client
func (c *Client) SendPing() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Status != StatusConnected {
		return ErrClientDisconnected
	}

	if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		c.ErrorCount++
		return err
	}

	return nil
}

// Close closes the client connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Status == StatusDisconnected {
		return nil
	}

	c.Status = StatusDisconnecting

	// Send close message
	closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection closed")
	err := c.Conn.WriteMessage(websocket.CloseMessage, closeMessage)

	// Close connection
	c.Status = StatusDisconnected

	return err
}

// UpdateActivity updates the client's last activity timestamp
func (c *Client) UpdateActivity() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.LastActivity = time.Now()
}

// SetMetadata sets a metadata value for the client
func (c *Client) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Metadata[key] = value
}

// GetMetadata gets a metadata value for the client
func (c *Client) GetMetadata(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.Metadata[key]
	return value, ok
}

// IsIdle checks if the client has been idle for longer than the given duration
func (c *Client) IsIdle(duration time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return time.Since(c.LastActivity) > duration
}
