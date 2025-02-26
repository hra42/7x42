package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// MessageMetadata contains additional information about a message
type MessageMetadata struct {
	Model       string `json:"model,omitempty"`
	TokenCount  int    `json:"token_count,omitempty"`
	ProcessTime int    `json:"process_time,omitempty"`
}

// Value implements the driver.Valuer interface for GORM
func (m MessageMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for GORM
func (m *MessageMetadata) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

// AddMetric adds a metric to the metadata
func (m *MessageMetadata) AddMetric(key string, value interface{}) {
	// Convert the metadata to a map
	metadataMap := make(map[string]interface{})
	metadataBytes, _ := json.Marshal(m)
	json.Unmarshal(metadataBytes, &metadataMap)

	// Add the new metric
	metadataMap[key] = value

	// Convert back to metadata
	updatedBytes, _ := json.Marshal(metadataMap)
	json.Unmarshal(updatedBytes, m)
}

// EstimateTokenCount estimates the token count based on content length
// This is a simple estimation; for accurate counts, use a tokenizer
func EstimateTokenCount(content string) int {
	// Rough estimate: 1 token ≈ 4 characters for English text
	return len(content) / 4
}
