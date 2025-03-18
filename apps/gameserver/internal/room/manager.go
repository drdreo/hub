package room

import (
	"encoding/json"
	"errors"
	"gameserver/internal/interfaces"
	"github.com/rs/zerolog/log"
	"maps"
	"slices"
	"sync"
)

// RoomManager handles the creation and tracking of game rooms
type RoomManager struct {
	rooms        map[string]interfaces.Room
	mu           sync.RWMutex
	gameRegistry interfaces.GameRegistry
}

// NewRoomManager creates a new room manager
func NewRoomManager(registry interfaces.GameRegistry) *RoomManager {
	return &RoomManager{
		rooms:        make(map[string]interfaces.Room),
		gameRegistry: registry,
	}
}

// CreateRoom creates a new game room
func (m *RoomManager) CreateRoom(gameType string, options json.RawMessage) (interfaces.Room, error) {
	// Verify game type exists
	if !m.gameRegistry.HasGame(gameType) {
		return nil, errors.New("unknown game type")
	}

	room := NewRoom(gameType)
	log.Info().Str("id", room.ID()).Str("type", room.GameType()).Msg("room created")

	// Initialize with game-specific settings
	if err := m.gameRegistry.InitializeRoom(room, options); err != nil {
		return nil, err
	}

	// Store room
	m.mu.Lock()
	m.rooms[room.ID()] = room
	m.mu.Unlock()

	log.Debug().Msg("room stored")

	return room, nil
}

// GetRoom retrieves a room by ID
func (m *RoomManager) GetRoom(roomID string) (interfaces.Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.rooms[roomID]
	if !exists {
		return nil, errors.New("room not found")
	}

	return room, nil
}

// RemoveRoom removes a room
func (m *RoomManager) RemoveRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, exists := m.rooms[roomID]; exists {
		room.Close()
		delete(m.rooms, roomID)
	}
}

// ListRooms returns a list of all active rooms
func (m *RoomManager) ListRooms() []interfaces.Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return slices.Collect(maps.Values(m.rooms))
}

// ListRoomsByGameType returns rooms of a specific game type
func (m *RoomManager) ListRoomsByGameType(gameType string) []interfaces.Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var roomList []interfaces.Room
	for _, room := range m.rooms {
		if room.GameType() == gameType {
			roomList = append(roomList, room)
		}
	}

	return roomList
}

// Cleanup periodically checks for and removes empty rooms
func (m *RoomManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, room := range m.rooms {
		if len(room.Clients()) == 0 {
			room.Close()
			delete(m.rooms, id)
		}
	}
}
