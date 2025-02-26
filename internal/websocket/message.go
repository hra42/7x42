package websocket

import (
	"encoding/json"
	"fmt"
	"time"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// TypeChatMessage is a chat message
	TypeChatMessage MessageType = "chat_message"
	// TypeTyping indicates the user is typing
	TypeTyping MessageType = "typing"
	// TypePing is a ping message
	TypePing MessageType = "ping"
	// TypePong is a pong message
	TypePong MessageType = "pong"
	// TypeError is an error message
	TypeError MessageType = "error"
	// TypeSystem is a system message
	TypeSystem MessageType = "system"
)

// Message represents a WebSocket message
type Message struct {
	Type    MessageType     `json:"type"`
	Content json.RawMessage `json:"content"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ChatID    uint      `json:"chatId"`
	Content   string    `json:"content"`
	Role      string    `json:"role"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatMessageRaw is used for parsing chat messages
type ChatMessageRaw struct {
	ChatID    interface{} `json:"chatId"`
	Content   string      `json:"content"`
	Role      string      `json:"role"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// SystemMessage represents a system message
type SystemMessage struct {
	Message string `json:"message"`
}

// NewChatMessage creates a new chat message
func NewChatMessage(chatID uint, content, role string) *Message {
	chatMsg := ChatMessage{
		ChatID:    chatID,
		Content:   content,
		Role:      role,
		Timestamp: time.Now(),
	}

	contentBytes, _ := json.Marshal(chatMsg)

	return &Message{
		Type:    TypeChatMessage,
		Content: contentBytes,
	}
}

// NewTypingMessage creates a new typing message
func NewTypingMessage() *Message {
	return &Message{
		Type: TypeTyping,
	}
}

// NewErrorMessage creates a new error message
func NewErrorMessage(message, code string) *Message {
	errMsg := ErrorMessage{
		Message: message,
		Code:    code,
	}

	contentBytes, _ := json.Marshal(errMsg)

	return &Message{
		Type:    TypeError,
		Content: contentBytes,
	}
}

// NewSystemMessage creates a new system message
func NewSystemMessage(message string) *Message {
	sysMsg := SystemMessage{
		Message: message,
	}

	contentBytes, _ := json.Marshal(sysMsg)

	return &Message{
		Type:    TypeSystem,
		Content: contentBytes,
	}
}

// ParseChatID parses a chat ID from different types
func ParseChatID(rawID interface{}) (uint, error) {
	switch v := rawID.(type) {
	case float64:
		return uint(v), nil
	case int:
		return uint(v), nil
	case string:
		var id uint
		if _, err := fmt.Sscanf(v, "%d", &id); err != nil {
			return 0, err
		}
		return id, nil
	default:
		return 0, fmt.Errorf("invalid chat ID format: %v (type: %T)", rawID, rawID)
	}
}
