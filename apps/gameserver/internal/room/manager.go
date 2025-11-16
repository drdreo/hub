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
	rooms            map[string]interfaces.Room
	mu               sync.RWMutex
	gameRegistry     interfaces.GameRegistry
	cleanupInterval  time.Duration
	cleanupTicker    *time.Ticker
	cleanupStop      chan struct{}
	onRoomListChange func(gameType string)
}

// RoomManagerOption is a functional option for configuring RoomManager
type RoomManagerOption func(*RoomManager)

// WithCleanupInterval sets a custom cleanup interval
func WithCleanupInterval(interval time.Duration) RoomManagerOption {
	return func(rm *RoomManager) {
		rm.cleanupInterval = interval
	}
}

// WithRoomListChangeCallback sets a callback to be called when room lists change
func WithRoomListChangeCallback(callback func(gameType string)) RoomManagerOption {
	return func(rm *RoomManager) {
		rm.onRoomListChange = callback
	}
}

func (rm *RoomManager) SetRoomListChangeCallback(callback func(gameType string)) {
	rm.onRoomListChange = callback
}

// NewRoomManager creates a new room manager
func NewRoomManager(registry interfaces.GameRegistry, opts ...RoomManagerOption) *RoomManager {
	rm := &RoomManager{
		rooms:           make(map[string]interfaces.Room),
		gameRegistry:    registry,
		cleanupInterval: 5 * time.Minute, // Default: 5 minutes
		cleanupStop:     make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(rm)
	}

	rm.startCleanup()

	return rm
}

// startCleanup starts the background cleanup routine
func (m *RoomManager) startCleanup() {
	m.cleanupTicker = time.NewTicker(m.cleanupInterval)
	go func() {
		for {
			select {
			case <-m.cleanupTicker.C:
				m.Cleanup()
			case <-m.cleanupStop:
				m.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// Stop gracefully stops the room manager and its cleanup routine
func (m *RoomManager) Stop() {
	close(m.cleanupStop)

	// Close all remaining rooms
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, room := range m.rooms {
		room.Close()
		delete(m.rooms, id)
	}

	log.Info().Msg("room manager stopped")
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

	// Notify about room list change
	if m.onRoomListChange != nil {
		m.onRoomListChange(createOptions.GameType)
	}

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

	log.Info().Str("roomId", roomID).Msg("removing room")

	var gameType string
	if room, exists := m.rooms[roomID]; exists {
		gameType = room.GameType()
		room.Close()
		delete(m.rooms, roomID)
	}

	m.mu.Unlock()

	// Notify about room list change after releasing the lock
	if gameType != "" && m.onRoomListChange != nil {
		m.onRoomListChange(gameType)
	}
}

// ListRooms returns a list of all active rooms
func (m *RoomManager) ListRooms() []interfaces.Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return slices.Collect(maps.Values(m.rooms))
}

// Cleanup periodically checks for and removes empty rooms
// Returns a slice of game types that had rooms cleaned up
func (m *RoomManager) Cleanup() []string {
	// First find empty rooms with read lock
	m.mu.RLock()
	type roomInfo struct {
		id       string
		gameType string
	}
	roomsToCleanup := make([]roomInfo, 0)
	for id, room := range m.rooms {
		if len(room.Clients()) == 0 {
			roomsToCleanup = append(roomsToCleanup, roomInfo{id: id, gameType: room.GameType()})
		}
	}
	m.mu.RUnlock()

	// Early return if nothing to cleanup
	if len(roomsToCleanup) == 0 {
		log.Debug().Msg("cleanup completed: no empty rooms found")
		return nil
	}

	log.Info().
		Int("count", len(roomsToCleanup)).
		Msg("cleaning up empty rooms")

	// Then remove rooms incrementally with brief write locks
	// Track which game types were affected
	affectedGameTypes := make(map[string]bool)
	cleanedCount := 0
	for _, info := range roomsToCleanup {
		m.mu.Lock()

		// Double-check room is still empty (could have changed between locks)
		if room, exists := m.rooms[info.id]; exists {
			if len(room.Clients()) == 0 {
				room.Close()
				delete(m.rooms, info.id)
				affectedGameTypes[info.gameType] = true
				cleanedCount++
			}
		}

		m.mu.Unlock()
	}

	log.Info().
		Int("cleaned", cleanedCount).
		Int("checked", len(roomsToCleanup)).
		Msg("cleanup completed")

	// Convert affected game types map to slice and notify about changes
	gameTypes := make([]string, 0, len(affectedGameTypes))
	for gameType := range affectedGameTypes {
		gameTypes = append(gameTypes, gameType)

		// Notify about room list change for this game type
		if m.onRoomListChange != nil {
			m.onRoomListChange(gameType)
		}
	}

	return gameTypes
}

var (
	ErrRoomNotFound = errors.New("room not found")
)
