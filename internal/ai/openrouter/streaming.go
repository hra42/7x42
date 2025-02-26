package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/websocket/v2"
)

// streamAPIResponse handles streaming the API response through WebSockets
func (c *Client) streamAPIResponse(ctx context.Context, wsConn *websocket.Conn, messages []ChatMessage) (string, error) {
	requestBody := map[string]interface{}{
		"model":       c.config.Model,
		"messages":    messages,
		"temperature": c.config.Temperature,
		"max_tokens":  c.config.MaxTokens,
		"stream":      true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setRequestHeaders(req)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned non-200 status: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return c.processStreamResponse(resp.Body, wsConn)
}

// processStreamResponse handles the streaming response data
func (c *Client) processStreamResponse(responseBody io.ReadCloser, wsConn *websocket.Conn) (string, error) {
	reader := bufio.NewReaderSize(responseBody, 32*1024)
	var buffer string
	var fullContent string

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fullContent, fmt.Errorf("error reading stream: %w", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		buffer += string(line)

		// Skip non-data lines
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		data := bytes.TrimPrefix(line, []byte("data: "))

		// Check for stream end
		if bytes.Equal(bytes.TrimSpace(data), []byte("[DONE]")) {
			break
		}

		// Parse the streaming response
		var streamResponse streamResponse
		if err := json.Unmarshal(data, &streamResponse); err != nil {
			continue
		}

		// Process content if available
		if len(streamResponse.Choices) > 0 && streamResponse.Choices[0].Delta.Content != "" {
			content := streamResponse.Choices[0].Delta.Content
			fullContent += content

			// Send the content chunk via WebSocket
			if err := wsConn.WriteJSON(map[string]interface{}{
				"type": "chat_message",
				"content": map[string]interface{}{
					"content":   content,
					"role":      "assistant",
					"timestamp": time.Now(),
				},
			}); err != nil {
				return fullContent, fmt.Errorf("failed to send message chunk: %w", err)
			}
		}
	}

	return fullContent, nil
}
