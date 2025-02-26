package websocket

import (
	"errors"
	"fmt"
)

// Common WebSocket errors
var (
	ErrClientDisconnected = errors.New("client disconnected")
	ErrInvalidMessage     = errors.New("invalid message format")
	ErrMessageTooLarge    = errors.New("message too large")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrInvalidChatID      = errors.New("invalid chat ID")
)

// WebSocketError represents a WebSocket-specific error
type WebSocketError struct {
	Op   string // Operation that failed
	Err  error  // Original error
	Code string // Error code for the client
}

// NewError creates a new WebSocket error
func NewError(op string, err error, code string) error {
	return &WebSocketError{
		Op:   op,
		Err:  err,
		Code: code,
	}
}

// Error implements the error interface
func (e *WebSocketError) Error() string {
	return fmt.Sprintf("websocket %s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *WebSocketError) Unwrap() error {
	return e.Err
}

// GetCode returns the error code
func (e *WebSocketError) GetCode() string {
	return e.Code
}
