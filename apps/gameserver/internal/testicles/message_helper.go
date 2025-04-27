package testicles

import (
	"encoding/json"
	"gameserver/internal/protocol"
)

// TestMessage represents a generic game message structure
type TestMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// CreateGameMessage creates a JSON message for game actions
func CreateGameMessage(msgType string, actionData interface{}) []byte {
	msg := TestMessage{
		Type: msgType,
	}

	if actionData != nil {
		dataBytes, _ := json.Marshal(actionData)
		msg.Data = dataBytes
	}

	msgBytes, _ := json.Marshal(msg)
	return msgBytes
}

// ExtractField extracts a field from a response message
func ExtractField(response *protocol.Response, field string) (interface{}, bool) {
	if response.Data == nil {
		return nil, false
	}

	// Handle different data types
	switch data := response.Data.(type) {
	case map[string]interface{}:
		value, exists := data[field]
		return value, exists
	case *map[string]interface{}:
		value, exists := (*data)[field]
		return value, exists
	}

	// Try to convert to map if it's a struct
	dataMap := make(map[string]interface{})
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, false
	}

	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		return nil, false
	}

	value, exists := dataMap[field]
	return value, exists
}

// FindMessageByType finds a message of a specific type in a list of responses
func FindMessageByType(messages []*protocol.Response, msgType string) (*protocol.Response, bool) {
	for _, msg := range messages {
		if msg.Type == msgType {
			return msg, true
		}
	}
	return nil, false
}

// ExtractJoinRoomResponseData extracts the data from a join_room_result response
func ExtractJoinRoomResponseData(response *protocol.Response) (string, bool) {
	if response.Type != "join_room_result" || !response.Success {
		return "", false
	}

	// Extract the roomID field from the response data
	roomID, ok := GetMessageField(response, "roomId")
	if !ok {
		// If direct field extraction doesn't work, try JSON approach
		dataMap := make(map[string]interface{})
		dataBytes, err := json.Marshal(response.Data)
		if err != nil {
			return "", false
		}

		if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
			return "", false
		}

		roomID, ok = dataMap["roomId"]
		if !ok {
			return "", false
		}
	}

	// Convert the roomID to a string
	roomIDStr, ok := roomID.(string)
	if !ok {
		return "", false
	}

	return roomIDStr, true
}
