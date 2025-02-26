package openrouter

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/repository"
	"gorm.io/gorm"
)

// Client handles communication with the OpenRouter API
type Client struct {
	config      Config
	mu          sync.RWMutex
	isReady     bool
	db          *gorm.DB
	chatRepo    *repository.ChatRepository
	messageRepo *repository.MessageRepository // Add this field
	httpClient  *http.Client
}

// New creates a new OpenRouter client with the given configuration
func New(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Set default values
	config.SetDefaults()

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Second * 60,
		},
	}, nil
}

// Initialize prepares the client for use
func (c *Client) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isReady = true
	return nil
}

// SetChatRepository sets the repository for chat operations
func (c *Client) SetChatRepository(repo *repository.ChatRepository) {
	c.chatRepo = repo
}

// SetMessageRepository sets the message repository
func (c *Client) SetMessageRepository(repo *repository.MessageRepository) {
	c.messageRepo = repo
}

// GenerateResponse sends a prompt to OpenRouter and returns the response
func (c *Client) GenerateResponse(ctx context.Context, prompt string, chatHistory []ChatMessage) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isReady {
		return "", errors.New("client not initialized")
	}

	messages := c.prepareMessages(prompt, chatHistory)

	requestBody := map[string]interface{}{
		"model":       c.config.Model,
		"messages":    messages,
		"temperature": c.config.Temperature,
		"max_tokens":  c.config.MaxTokens,
	}

	var response string
	var err error

	// Implement retry logic
	for i := 0; i < c.config.MaxRetries; i++ {
		response, err = c.makeAPIRequest(ctx, requestBody)
		if err == nil {
			return response, nil
		}

		if i < c.config.MaxRetries-1 {
			time.Sleep(c.config.RetryDelay)
		}
	}

	return "", err
}

// StreamResponse streams the AI response through a WebSocket connection
func (c *Client) StreamResponse(ctx context.Context, wsConn *websocket.Conn, prompt string, chatHistory []ChatMessage) error {
	c.mu.RLock()
	if !c.isReady {
		c.mu.RUnlock()
		return errors.New("client not initialized")
	}
	c.mu.RUnlock()

	// Send typing indicator
	if err := wsConn.WriteJSON(map[string]interface{}{
		"type": "typing",
	}); err != nil {
		return err
	}

	messages := c.prepareMessages(prompt, chatHistory)

	// Extract chat ID if present in the prompt
	chatID := c.extractChatID(prompt)

	// Stream the response
	startTime := time.Now()
	fullContent, err := c.streamAPIResponse(ctx, wsConn, messages)
	if err != nil {
		return err
	}

	// Calculate processing metrics
	processingTime := int(time.Since(startTime).Milliseconds())

	// Save the message if we have a chat repository and valid chat ID
	if c.chatRepo != nil && chatID > 0 {
		c.saveMessage(ctx, chatID, fullContent, processingTime)
	}

	// Send completion notification
	return wsConn.WriteJSON(map[string]interface{}{
		"type": "chat_message",
		"metadata": map[string]interface{}{
			"complete":       true,
			"processingTime": processingTime,
		},
	})
}
