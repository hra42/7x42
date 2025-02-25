package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/hra42/7x42/internal/models"
	"github.com/hra42/7x42/internal/repository"
	"gorm.io/gorm"
)

type Config struct {
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
	MaxRetries  int
	RetryDelay  time.Duration
	BaseURL     string
}

type OpenRouterClient struct {
	config   Config
	mu       sync.RWMutex
	isReady  bool
	db       *gorm.DB
	chatRepo *repository.ChatRepository
}

func NewOpenRouterClient(config Config) (*OpenRouterClient, error) {
	if config.APIKey == "" {
		return nil, errors.New("API key is required")
	}

	return &OpenRouterClient{
		config: config,
	}, nil
}

func (c *OpenRouterClient) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isReady = true
	return nil
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *OpenRouterClient) GenerateResponse(ctx context.Context, prompt string, chatHistory []Message) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isReady {
		return "", errors.New("client not initialized")
	}

	// Prepare messages
	var messages []Message
	if len(chatHistory) == 0 {
		messages = []Message{
			{Role: "user", Content: prompt},
		}
	} else {
		messages = append(chatHistory, Message{Role: "user", Content: prompt})
	}

	// Create the request body
	jsonBody := map[string]interface{}{
		"model":    c.config.Model,
		"messages": messages,
	}

	// Make request with retries
	var response string
	var err error
	for i := 0; i <= c.config.MaxRetries; i++ {
		response, err = c.makeDirectRequest(ctx, jsonBody)
		if err == nil {
			return response, nil
		}
		if i < c.config.MaxRetries {
			log.Printf("Retrying request to OpenRouter (attempt %d/%d): %v", i+1, c.config.MaxRetries, err)
			time.Sleep(c.config.RetryDelay)
		}
	}

	return "", fmt.Errorf("failed after %d retries: %w", c.config.MaxRetries, err)
}

func (c *OpenRouterClient) makeDirectRequest(ctx context.Context, jsonBody map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(jsonBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Title", "7x42 Chat")
	req.Header.Set("HTTP-Referer", "https://7x42.net")

	client := &http.Client{Timeout: time.Second * 60}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned non-200 status: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", errors.New("no response choices returned from the API")
	}

	return result.Choices[0].Message.Content, nil
}

func (c *OpenRouterClient) StreamResponse(ctx context.Context, wsConn *websocket.Conn, prompt string, chatHistory []Message) error {
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
		return fmt.Errorf("failed to send typing indicator: %w", err)
	}

	// Prepare messages
	var messages []Message
	if len(chatHistory) == 0 {
		messages = []Message{
			{Role: "user", Content: prompt},
		}
	} else {
		messages = append(chatHistory, Message{Role: "user", Content: prompt})
	}

	// Create the request body for streaming
	jsonBody := map[string]interface{}{
		"model":    c.config.Model,
		"messages": messages,
		"stream":   true,
	}

	jsonData, err := json.Marshal(jsonBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("X-Title", "7x42 Chat")
	req.Header.Set("HTTP-Referer", "https://7x42.net")

	client := &http.Client{Timeout: time.Second * 120}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned non-200 status: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	startTime := time.Now()
	var fullContent string

	// Use a buffered reader with a larger buffer size to handle large tokens
	reader := bufio.NewReaderSize(resp.Body, 32*1024)
	buffer := ""

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading stream: %w", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		buffer += string(line)

		// Process SSE format
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		data := line[6:] // Remove "data: " prefix

		// Check for end of stream
		if bytes.Equal(bytes.TrimSpace(data), []byte("[DONE]")) {
			break
		}

		// Parse the response chunk
		var streamResponse struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(data, &streamResponse); err != nil {
			// Skip parsing errors
			continue
		}

		if len(streamResponse.Choices) > 0 && streamResponse.Choices[0].Delta.Content != "" {
			content := streamResponse.Choices[0].Delta.Content
			fullContent += content

			// Send the chunk to the WebSocket client
			if err := wsConn.WriteJSON(map[string]interface{}{
				"type": "chat_message",
				"content": map[string]interface{}{
					"content":   content,
					"timestamp": time.Now(),
				},
			}); err != nil {
				return fmt.Errorf("failed to send message chunk: %w", err)
			}
		}
	}

	// Extract chat ID from prompt if available
	var chatID uint64
	chatIDMatches := regexp.MustCompile(`chat:(\d+)`).FindStringSubmatch(prompt)
	if len(chatIDMatches) > 1 {
		if id, err := strconv.ParseUint(chatIDMatches[1], 10, 64); err == nil {
			chatID = id
		}
	}

	// Save the complete message to database
	processingTime := int(time.Since(startTime).Milliseconds())
	if chatID > 0 && c.chatRepo != nil && fullContent != "" {
		message := &models.Message{
			Content:   fullContent,
			Role:      "assistant",
			ChatID:    chatID,
			Timestamp: time.Now(),
			Metadata: models.MessageMetadata{
				Model:       c.config.Model,
				TokenCount:  len(fullContent) / 4, // Rough estimate
				ProcessTime: processingTime,
			},
		}

		if err := c.chatRepo.CreateMessage(ctx, message); err != nil {
			log.Printf("Failed to save AI response: %v", err)
		}
	}

	// Send completion message
	return wsConn.WriteJSON(map[string]interface{}{
		"type": "chat_message",
		"metadata": map[string]interface{}{
			"processingTime": processingTime,
			"complete":       true,
		},
	})
}

func (c *OpenRouterClient) SetChatRepository(repo *repository.ChatRepository) {
	c.chatRepo = repo
}
