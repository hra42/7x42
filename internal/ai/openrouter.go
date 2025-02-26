package ai

import (
	"github.com/hra42/7x42/internal/ai/openrouter"
	"github.com/hra42/7x42/internal/repository"
)

// OpenRouterClient is the main interface for interacting with OpenRouter API
type OpenRouterClient struct {
	client *openrouter.Client
}

// NewOpenRouterClient creates a new OpenRouter client
func NewOpenRouterClient(config openrouter.Config) (*OpenRouterClient, error) {
	client, err := openrouter.New(config)
	if err != nil {
		return nil, err
	}

	return &OpenRouterClient{
		client: client,
	}, nil
}

// Initialize prepares the client for use
func (c *OpenRouterClient) Initialize() error {
	return c.client.Initialize()
}

// SetChatRepository sets the repository for chat operations
func (c *OpenRouterClient) SetChatRepository(repo *repository.ChatRepository) {
	c.client.SetChatRepository(repo)
}

// SetMessageRepository sets the message repository for the client
func (c *OpenRouterClient) SetMessageRepository(repo *repository.MessageRepository) {
	c.client.SetMessageRepository(repo)
}
