package router

import (
	"encoding/json"
	"github.com/drdreo/hub/gameserver/internal/room"

	"github.com/drdreo/hub/gameserver/internal/protocol"
)

type Client interface {
	ID() string
	Send(message []byte) error
	Room() room.Room
	SetRoom(room room.Room)
	Close()
}

type RoomManager interface {
	CreateRoom(gameType string, options json.RawMessage) (room.Room, error)
	GetRoom(roomID string) (room.Room, error)
	RemoveRoom(roomID string)
}

type GameRegistry interface {
	HandleMessage(client Client, msgType string, payload []byte) error
	HandleClientJoin(client Client, room room.Room) error
	HandleClientLeave(client Client, room room.Room) error
}

// Router handles WebSocket message routing
type Router struct {
	roomManager  RoomManager
	gameRegistry GameRegistry
}

// NewRouter creates a new message router
func NewRouter(roomManager RoomManager, gameRegistry GameRegistry) *Router {
	return &Router{
		roomManager:  roomManager,
		gameRegistry: gameRegistry,
	}
}

// HandleMessage processes an incoming message from a client
func (r *Router) HandleMessage(client Client, messageData []byte) {
	var message protocol.Message
	if err := json.Unmarshal(messageData, &message); err != nil {
		client.Send(protocol.NewErrorResponse("error", "Invalid message format"))
		return
	}

	switch message.Type {
	case "create_room":
		r.handleCreateRoom(client, message)
	case "join_room":
		r.handleJoinRoom(client, message)
	case "leave_room":
		r.handleLeaveRoom(client)
	case "game_action":
		r.handleGameAction(client, message)
	default:
		// Forward to game-specific handler
		if client.Room() != nil {
			err := r.gameRegistry.HandleMessage(client, message.Type, message.Payload)
			if err != nil {
				client.Send(protocol.NewErrorResponse("error", err.Error()))
			}
		} else {
			client.Send(protocol.NewErrorResponse("error", "Unknown message type"))
		}
	}
}

// handleCreateRoom creates a new game room
func (r *Router) handleCreateRoom(client Client, msg protocol.Message) {
	var createOptions struct {
		GameType string          `json:"gameType"`
		Options  json.RawMessage `json:"options,omitempty"`
	}

	if err := json.Unmarshal(msg.Payload, &createOptions); err != nil {
		client.Send(protocol.NewErrorResponse("create_room_result", "Invalid options format"))
		return
	}

	room, err := r.roomManager.CreateRoom(createOptions.GameType, createOptions.Options)
	if err != nil {
		client.Send(protocol.NewErrorResponse("create_room_result", err.Error()))
		return
	}

	// Join the room
	if err := room.Join(client); err != nil {
		client.Send(protocol.NewErrorResponse("create_room_result", err.Error()))
		return
	}

	// Notify game about client join
	r.gameRegistry.HandleClientJoin(client, room)

	response := map[string]interface{}{
		"roomId":   room.ID(),
		"gameType": room.GameType(),
	}
	client.Send(protocol.NewSuccessResponse("create_room_result", response))
}

// handleJoinRoom joins an existing room
func (r *Router) handleJoinRoom(client Client, msg protocol.Message) {
	var joinOptions struct {
		RoomID string `json:"roomId"`
	}

	if err := json.Unmarshal(msg.Payload, &joinOptions); err != nil {
		client.Send(protocol.NewErrorResponse("join_room_result", "Invalid options format"))
		return
	}

	room, err := r.roomManager.GetRoom(joinOptions.RoomID)
	if err != nil {
		client.Send(protocol.NewErrorResponse("join_room_result", err.Error()))
		return
	}

	// Join the room
	if err := room.Join(client); err != nil {
		client.Send(protocol.NewErrorResponse("join_room_result", err.Error()))
		return
	}

	// Notify game about client join
	r.gameRegistry.HandleClientJoin(client, room)

	// Send success response
	response := map[string]interface{}{
		"roomId":   room.ID(),
		"gameType": room.GameType(),
		"clients":  len(room.Clients()),
	}
	client.Send(protocol.NewSuccessResponse("join_room_result", response))
}

// handleLeaveRoom leaves the current room
func (r *Router) handleLeaveRoom(client Client) {
	room := client.Room()
	if room == nil {
		client.Send(protocol.NewErrorResponse("leave_room_result", "Client not in a room"))
		return
	}

	// Notify game about client leave
	r.gameRegistry.HandleClientLeave(client, room)

	// Leave the room
	roomID := room.ID()
	room.Leave(client)
	client.SetRoom(nil)

	response := map[string]string{
		"roomId": roomID,
	}
	client.Send(protocol.NewSuccessResponse("leave_room_result", response))
}

// handleGameAction forwards a game-specific action to the game handler
func (r *Router) handleGameAction(client Client, msg protocol.Message) {
	if client.Room() == nil {
		client.Send(protocol.NewErrorResponse("game_action_result", "Client not in a room"))
		return
	}

	if err := r.gameRegistry.HandleMessage(client, "game_action", msg.Payload); err != nil {
		client.Send(protocol.NewErrorResponse("game_action_result", err.Error()))
		return
	}
}
