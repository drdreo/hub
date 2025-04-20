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
	Error   string      `json:"error,omitempty"` // TODO: look into using Error error type instead
	Data    interface{} `json:"data,omitempty"`
}

func (r *Response) ToBytes() []byte {
	data, _ := json.Marshal(r)
	return data
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(responseType string, data interface{}) *Response {
	return &Response{
		Type:    responseType,
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(responseType string, errorMsg string) *Response {
	return &Response{
		Type:    responseType,
		Success: false,
		Error:   errorMsg,
	}
}
