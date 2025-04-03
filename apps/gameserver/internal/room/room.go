package room

import (
	"errors"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
)

// GameRoom implements the Room interface
type GameRoom struct {
	id       string
	gameType string
	clients  map[string]interfaces.Client
	state    interface{}
	mu       sync.RWMutex
	closed   bool
}

// NewRoom creates a new game room
func NewRoom(gameType string, roomId *string) *GameRoom {
	var id string
	if roomId == nil || len(*roomId) == 0 {
		id = uuid.New().String()
	} else {
		id = *roomId
	}

	return &GameRoom{
		id:       id,
		gameType: gameType,
		clients:  make(map[string]interfaces.Client),
		closed:   false,
	}
}

// ID returns the room's unique ID
func (room *GameRoom) ID() string {
	return room.id
}

// GameType returns the room's game type
func (room *GameRoom) GameType() string {
	return room.gameType
}

// IsClosed returns the room's closed status
func (room *GameRoom) IsClosed() bool {
	return room.closed
}

// Join adds a client to the room
func (room *GameRoom) Join(client interfaces.Client) error {
	room.mu.Lock()
	defer room.mu.Unlock()
	log.Debug().Str("room", room.ID()).Str("client", client.ID()).Msg("client joining")

	// First, check if room is closed (only taking room lock briefly)
	if room.closed {
		return ErrRoomClosed
	}

	// Leave old room
	oldRoom := client.Room()
	if oldRoom != nil && oldRoom.ID() != room.id {
		oldRoom.Leave(client)
	}

	client.SetRoom(room)

	room.clients[client.ID()] = client

	// Notify other clients about the new joiner
	joinMessage := protocol.NewSuccessResponse("client_joined", map[string]interface{}{
		"clientId": client.ID(),
	})

	room.Broadcast(joinMessage, client)

	return nil
}

// Leave removes a client from the room
func (room *GameRoom) Leave(client interfaces.Client) {
	if _, exists := room.clients[client.ID()]; exists {
		delete(room.clients, client.ID())

		// Notify other clients about the departure
		leaveMessage := protocol.NewSuccessResponse("client_left", map[string]interface{}{
			"clientId": client.ID(),
		})

		room.Broadcast(leaveMessage)
	}

	// Close room if empty
	if len(room.clients) == 0 && !room.closed {
		room.Close()
	}
}

// Broadcast sends a message to all clients in the room except excluded ones
func (room *GameRoom) Broadcast(message *protocol.Response, exclude ...interfaces.Client) {
	excludeMap := make(map[string]bool)
	for _, client := range exclude {
		excludeMap[client.ID()] = true
	}

	for _, client := range room.clients {
		if !excludeMap[client.ID()] {
			client.Send(message)
		}
	}
}

// BroadcastTo sends a message to specific clients in the room
func (room *GameRoom) BroadcastTo(message *protocol.Response, clients ...interfaces.Client) {
	for _, client := range clients {
		client.Send(message)
	}
}

// Clients returns a map of clients in the room
func (room *GameRoom) Clients() map[string]interfaces.Client {
	room.mu.RLock()
	defer room.mu.RUnlock()

	// Return a copy to avoid race conditions
	clientsCopy := make(map[string]interfaces.Client)
	for id, client := range room.clients {
		clientsCopy[id] = client
	}

	return clientsCopy
}

// State returns the room's current state
func (room *GameRoom) State() interface{} {
	room.mu.RLock()
	defer room.mu.RUnlock()
	return room.state
}

// SetState updates the room's state
func (room *GameRoom) SetState(state interface{}) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.state = state
}

// Close terminates the room and disconnects all clients
func (room *GameRoom) Close() {
	if room.closed {
		return
	}

	room.closed = true

	// Notify all clients
	closeMessage := protocol.NewSuccessResponse("room_closed", map[string]interface{}{
		"roomId": room.id,
	})

	room.Broadcast(closeMessage)
}

// Error definitions
var (
	ErrRoomClosed = errors.New("room is closed")
)
