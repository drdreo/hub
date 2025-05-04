package room

import (
	"context"
	"errors"
	"gameserver/internal/interfaces"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// RoomManager handles the creation and tracking of game rooms
type RoomManager struct {
	rooms        map[string]interfaces.Room
	mu           sync.RWMutex
	gameRegistry interfaces.GameRegistry
}

// NewRoomManager creates a new room manager
func NewRoomManager(registry interfaces.GameRegistry) *RoomManager {
	rm := &RoomManager{
		rooms:        make(map[string]interfaces.Room),
		gameRegistry: registry,
	}

	// Run cleanup every 5 minutes as a safety net
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			rm.Cleanup()
		}
	}()

	return rm
}

// CreateRoom creates a new game room
func (m *RoomManager) CreateRoom(ctx context.Context, createOptions interfaces.CreateRoomOptions) (interfaces.Room, error) {
	// Verify game type exists
	if !m.gameRegistry.HasGame(createOptions.GameType) {
		return nil, errors.New("unknown game type")
	}

	room := NewRoom(m, createOptions.GameType, createOptions.RoomID)
	log.Info().Str("id", room.ID()).Str("type", room.GameType()).Msg("room created")

	// Initialize with game-specific settings
	if err := m.gameRegistry.InitializeRoom(ctx, room, createOptions.Options); err != nil {
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
		return nil, ErrRoomNotFound
	}

	return room, nil
}

// GetAllRoomsByGameType retrieves a room by ID
func (m *RoomManager) GetAllRoomsByGameType(gameType string) []interfaces.Room {
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

// RemoveRoom removes a room
func (m *RoomManager) RemoveRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Info().Str("roomId", roomID).Msg("removing room")

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

// Cleanup periodically checks for and removes empty rooms
func (m *RoomManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Info().Msg("checking for cleanup")

	for id, room := range m.rooms {
		if len(room.Clients()) == 0 {
			room.Close()
			delete(m.rooms, id)
		}
	}
}

var (
	ErrRoomNotFound = errors.New("room not found")
)
