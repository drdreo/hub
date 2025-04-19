package room

import "gameserver/internal/interfaces"

// RoomManagerMock is a simple mock implementation of RoomManager for testing
type RoomManagerMock struct{}

// CreateRoom mocks the room creation
func (m *RoomManagerMock) CreateRoom(options interfaces.CreateRoomOptions) (interfaces.Room, error) {
	return nil, nil
}

// GetRoom mocks getting a room
func (m *RoomManagerMock) GetRoom(roomID string) (interfaces.Room, error) {
	return nil, nil
}

// RemoveRoom mocks room removal
func (m *RoomManagerMock) RemoveRoom(roomID string) {
	// Do nothing in the mock
}

// GetAllRoomsByGameType mocks getting rooms by game type
func (m *RoomManagerMock) GetAllRoomsByGameType(gameType string) []interfaces.Room {
	return nil
}

// NewRoomManagerMock creates a new mock room manager
func NewRoomManagerMock() interfaces.RoomManager {
	return &RoomManagerMock{}
}
