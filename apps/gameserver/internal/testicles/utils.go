package testicles

import (
	"encoding/json"
	"gameserver/internal/protocol"
)

// CreateMessage creates a protocol.Message with the given type and data
func CreateMessage(messageType string, data interface{}) []byte {
	var jsonData []byte

	if data == nil {
		jsonData = []byte("{}")
	} else if rawJSON, ok := data.(json.RawMessage); ok {
		jsonData = rawJSON
	} else {
		jsonData, _ = json.Marshal(data)
	}

	message := protocol.Message{
		Type: messageType,
		Data: json.RawMessage(jsonData),
	}

	msgData, _ := json.Marshal(message)
	return msgData
}

// GetMessageField extracts a field from a response's data
func GetMessageField(response *protocol.Response, fieldName string) (interface{}, bool) {
	if data, ok := response.Data.(map[string]interface{}); ok {
		value, exists := data[fieldName]
		return value, exists
	}
	return nil, false
}
