package router

import (
	"encoding/json"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"gameserver/internal/session"
	"github.com/rs/zerolog/log"
)

type RoomManager interface {
	CreateRoom(gameType string, options json.RawMessage) (interfaces.Room, error)
	GetRoom(roomID string) (interfaces.Room, error)
	RemoveRoom(roomID string)
}

// Router handles WebSocket message routing
type Router struct {
	roomManager  RoomManager
	gameRegistry interfaces.GameRegistry
}

// ReconnectPayload the reconnect message
type ReconnectPayload struct {
	ClientID string `json:"clientId"`
	RoomID   string `json:"roomId"`
}

// NewRouter creates a new message router
func NewRouter(roomManager RoomManager, gameRegistry interfaces.GameRegistry) *Router {
	log.Debug().Msg("creating new router")

	return &Router{
		roomManager:  roomManager,
		gameRegistry: gameRegistry,
	}
}

// HandleMessage processes an incoming message from a client
func (r *Router) HandleMessage(client interfaces.Client, messageData []byte) {
	var message protocol.Message
	if err := json.Unmarshal(messageData, &message); err != nil {
		log.Error().Err(err).Msg("Invalid message format")

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
	case "reconnect":
		r.handleReconnect(client, message)
	default:
		// Forward to game-specific handler
		if client.Room() != nil {
			err := r.gameRegistry.HandleMessage(client, message.Type, message.Data)
			if err != nil {
				client.Send(protocol.NewErrorResponse("error", err.Error()))
			}
		} else {
			client.Send(protocol.NewErrorResponse("error", "Unknown message type"))
		}
	}
}

// handleCreateRoom creates a new game room
func (r *Router) handleCreateRoom(client interfaces.Client, msg protocol.Message) {
	var createOptions struct {
		GameType string          `json:"gameType"`
		Options  json.RawMessage `json:"options,omitempty"`
	}

	if err := json.Unmarshal(msg.Data, &createOptions); err != nil {
		client.Send(protocol.NewErrorResponse("create_room_result", "Invalid options format"))
		return
	}

	log.Debug().Fields(createOptions).Msg("handleCreateRoom")

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
func (r *Router) handleJoinRoom(client interfaces.Client, msg protocol.Message) {
	var joinOptions struct {
		RoomID string `json:"roomId"`
	}

	if err := json.Unmarshal(msg.Data, &joinOptions); err != nil {
		client.Send(protocol.NewErrorResponse("join_room_result", "Invalid options format"))
		return
	}

	log.Debug().Fields(joinOptions).Msg("handleJoinRoom")

	room, err := r.roomManager.GetRoom(joinOptions.RoomID)
	if err != nil {
		log.Error().Err(err).Str("id", joinOptions.RoomID).Msg("failed to get room")
		client.Send(protocol.NewErrorResponse("join_room_result", err.Error()))
		return
	}

	// Join the room
	if err := room.Join(client); err != nil {
		log.Error().Err(err).Str("id", room.ID()).Msg("failed to join room")
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

	log.Info().Str("roomID", room.ID()).Msg("client joined room")

	client.Send(protocol.NewSuccessResponse("join_room_result", response))
}

// handleLeaveRoom leaves the current room
func (r *Router) handleLeaveRoom(client interfaces.Client) {
	room := client.Room()
	if room == nil {
		log.Warn().Str("id", client.ID()).Msg("client tried to leave room but room is not set")
		client.Send(protocol.NewErrorResponse("leave_room_result", "Client not in a room"))
		return
	}

	log.Debug().Str("clientID", client.ID()).Msg("handleJoinRoom")

	// Notify game about client leave
	r.gameRegistry.HandleClientLeave(client, room)

	// Leave the room
	roomID := room.ID()
	room.Leave(client)
	client.SetRoom(nil)

	response := map[string]string{
		"roomId": roomID,
	}

	log.Info().Str("roomID", roomID).Msg("client left room")

	client.Send(protocol.NewSuccessResponse("leave_room_result", response))
}

// handleGameAction forwards a game-specific action to the game handler
func (r *Router) handleGameAction(client interfaces.Client, msg protocol.Message) {
	if client.Room() == nil {
		client.Send(protocol.NewErrorResponse("game_action_result", "Client not in a room"))
		return
	}

	if err := r.gameRegistry.HandleMessage(client, "game_action", msg.Data); err != nil {
		client.Send(protocol.NewErrorResponse("game_action_result", err.Error()))
		return
	}
}

// handleReconnect tries to reconnect the new socket to an existing room
func (r *Router) handleReconnect(client interfaces.Client, msg protocol.Message) {
	if client.Room() != nil {
		client.Send(protocol.NewErrorResponse("reconnect_result", "Client is already in a room"))
		return
	}

	var recon ReconnectPayload
	if err := json.Unmarshal(msg.Data, &recon); err != nil {
		client.Send(protocol.NewErrorResponse("reconnect_result", err.Error()))
		return
	}

	log.Debug().Str("oldClientID", recon.ClientID).Str("newClientID", client.ID()).Msg("handleReconnect")

	// Get session store
	sessionStore := session.GetSessionStore()
	sessionData, exists := sessionStore.GetSession(recon.ClientID)
	if !exists {
		client.Send(protocol.NewErrorResponse("reconnect_result", "Session expired or not found"))
		return
	}

	// Find room (either from session or from request)
	roomID := sessionData.RoomID
	if recon.RoomID != "" {
		roomID = recon.RoomID // Optional: override with provided roomID
	}

	// Get the room
	targetRoom, err := r.roomManager.GetRoom(roomID)
	if err != nil {
		client.Send(protocol.NewErrorResponse("reconnect_result", "Room not found"))
		return
	}

	// Join room and reconnect
	if err := targetRoom.Join(client); err != nil {
		client.Send(protocol.NewErrorResponse("reconnect_result", err.Error()))
		return
	}

	// Handle reconnection at game level
	r.gameRegistry.HandleClientReconnect(client, targetRoom, recon.ClientID)

	// Remove the old session
	sessionStore.RemoveSession(recon.ClientID)

	response := map[string]interface{}{
		"roomId":   targetRoom.ID(),
		"gameType": targetRoom.GameType(),
	}

	log.Info().Str("roomId", targetRoom.ID()).Str("gameType", targetRoom.GameType()).Msg("client reconnected")

	client.Send(protocol.NewSuccessResponse("reconnect_result", response))
}
