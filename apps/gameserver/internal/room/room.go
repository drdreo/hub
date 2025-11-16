package room

import (
	"errors"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"maps"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
)

// GameRoom implements the Room interface
type GameRoom struct {
	id       string
	gameType string
	manager  interfaces.RoomManager
	clients  map[string]interfaces.Client
	state    interface{}
	closed   bool
	mu       sync.RWMutex

	closeTimer *time.Timer // handling delayed room closure
}

// NewRoom creates a new game room
func NewRoom(manager interfaces.RoomManager, gameType string, roomId *string) *GameRoom {
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
		manager:  manager,
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
	log.Debug().Str("roomId", room.ID()).Str("clientId", client.ID()).Msg("client joining")

	// First, check if the room is closed
	if room.closed {
		return ErrRoomClosed
	}

	// Leave the old room if it was different
	oldRoom := client.Room()
	if oldRoom != nil && oldRoom.ID() != room.id {
		oldRoom.Leave(client)
	}

	client.SetRoom(room)

	room.clients[client.ID()] = client

	// If this is a human client and we have a pending close timer, cancel it
	if !client.IsBot() && room.closeTimer != nil {
		log.Debug().Str("roomId", room.ID()).Msg("stoping room close timer")
		room.closeTimer.Stop()
		room.closeTimer = nil
	}

	// Notify other clients about the new joiner
	joinMessage := protocol.NewSuccessResponse("client_joined", interfaces.M{
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
		leaveMessage := protocol.NewSuccessResponse("client_left", interfaces.M{
			"clientId": client.ID(),
		})

		room.Broadcast(leaveMessage)
	}

	humanClientExists := room.hasHumanClients()
	// Close room if no human clients remain
	if !humanClientExists {
		room.ScheduleClose()
	}
}

// SendTo sends a message to the specific client with clientId
func (room *GameRoom) SendTo(message *protocol.Response, clientId string) {
	// Send to specific user
	if client, ok := room.clients[clientId]; ok {
		client.Send(message)
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

	return maps.Clone(room.clients)
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

func (room *GameRoom) ScheduleClose() {
	log.Debug().Str("roomId", room.ID()).Msg("room scheduled for closing")
	// If we're already pending closure, don't reset
	if room.closeTimer != nil {
		return
	}

	// Set a timer to close the room after the timeout (e.g., 30 seconds)
	room.closeTimer = time.AfterFunc(30*time.Second, func() {
		log.Debug().Str("roomId", room.ID()).Msg("room checking for closure after timeout")

		room.mu.Lock()
		defer room.mu.Unlock()

		// Check again if a human player has reconnected
		humanExists := room.hasHumanClients()

		// Only close if still no humans
		if !humanExists {
			log.Debug().Str("roomId", room.ID()).Msg("nobody in room")
			room.Close()
			// auto-remove from manager if manager exists
			if room.manager != nil {
				room.manager.RemoveRoom(room.ID())
				// TODO: broadcast new game room list update
			}
		} else {
			log.Debug().Str("roomId", room.ID()).Msg("room had humans again")
		}
		room.closeTimer = nil
	})
}

// Close terminates the room and disconnects all clients
func (room *GameRoom) Close() {
	if room.closed {
		return
	}

	log.Info().Str("roomId", room.ID()).Msg("room closing")
	room.closed = true

	// Notify all clients
	closeMessage := protocol.NewSuccessResponse("room_closed", map[string]interface{}{
		"roomId": room.id,
	})

	room.Broadcast(closeMessage)

	// Explicitly close all bot clients to ensure proper cleanup
	for id, client := range room.clients {
		if client.IsBot() {
			log.Info().Str("roomId", room.ID()).Str("botId", id).Msg("closing bot client")
			client.Close()
		}
	}
}

func (room *GameRoom) hasHumanClients() bool {
	for _, c := range room.clients {
		if !c.IsBot() {
			return true
		}
	}
	return false
}

// Error definitions
var (
	ErrRoomClosed = errors.New("room is closed")
)
