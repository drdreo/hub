package protocol

import (
	"encoding/json"
)

// Message represents the standard message format
type Message struct {
	Type     string          `json:"type"`
	RoomID   string          `json:"roomId,omitempty"`
	GameType string          `json:"gameType,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// Response represents a standard response format
type Response struct {
	Type    string      `json:"type"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(responseType string, data interface{}) []byte {
	resp := Response{
		Type:    responseType,
		Success: true,
		Data:    data,
	}

	bytes, _ := json.Marshal(resp)
	return bytes
}

// NewErrorResponse creates a new error response
func NewErrorResponse(responseType string, errorMsg string) []byte {
	resp := Response{
		Type:    responseType,
		Success: false,
		Error:   errorMsg,
	}

	bytes, _ := json.Marshal(resp)
	return bytes
}
