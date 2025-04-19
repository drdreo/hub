package router

import (
	"encoding/json"
	"errors"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"gameserver/internal/session"
	"github.com/rs/zerolog/log"
)

// Router handles WebSocket message routing
type Router struct {
	clientManager interfaces.ClientManager
	roomManager   interfaces.RoomManager
	gameRegistry  interfaces.GameRegistry
}

// ReconnectPayload the reconnect message
type ReconnectPayload struct {
	ClientID string `json:"clientId"`
	RoomID   string `json:"roomId"`
}

type RoomListInfo struct {
	RoomId      string `json:"roomId"`
	PlayerCount int    `json:"playerCount"`
	GameStarted bool   `json:"started"`
}

// NewRouter creates a new message router
func NewRouter(clientManager interfaces.ClientManager, roomManager interfaces.RoomManager, gameRegistry interfaces.GameRegistry) *Router {
	log.Debug().Msg("creating new router")

	return &Router{
		clientManager: clientManager,
		roomManager:   roomManager,
		gameRegistry:  gameRegistry,
	}
}

// HandleMessage processes an incoming message from a client
func (r *Router) HandleMessage(client interfaces.Client, messageData []byte) {
	var message protocol.Message
	if err := json.Unmarshal(messageData, &message); err != nil {
		log.Error().Err(err).Msg(ErrMessageInvalid.Error())

		client.Send(protocol.NewErrorResponse("error", ErrMessageInvalid.Error()))
		return
	}

	switch message.Type {
	case "join_room":
		r.handleJoinRoom(client, message.Data)
	case "leave_room":
		r.handleLeaveRoom(client)
	case "reconnect":
		r.handleReconnect(client, message.Data)
	case "game_action":
		r.handleGameAction(client, message.Data)
	case "add_bot":
		r.handleAddBot(client)
	case "get_room_list":
		r.handleGetRoomList(client, message.Data)
	default:
		// Forward to game-specific handler
		if client.Room() != nil {
			err := r.gameRegistry.HandleMessage(client, message.Type, message.Data)
			if err != nil {
				client.Send(protocol.NewErrorResponse("error", err.Error()))
			}
		} else {
			client.Send(protocol.NewErrorResponse("error", "Unknown message type: "+message.Type))
		}
	}
}

// handleCreateRoom creates a new game room
func (r *Router) handleCreateRoom(createOptions interfaces.CreateRoomOptions) (interfaces.Room, error) {
	if createOptions.GameType == "" {
		return nil, ErrGameTypeRequired
	}

	log.Debug().Fields(createOptions).Msg("client creating room")

	room, err := r.roomManager.CreateRoom(createOptions)
	if err != nil {
		return nil, err
	}

	return room, nil
}

// handleJoinRoom joins an existing room
func (r *Router) handleJoinRoom(client interfaces.Client, data json.RawMessage) {
	// prevent multi-room joining
	if client.Room() != nil {
		client.Send(protocol.NewErrorResponse("join_room_result", ErrClientAlreadyInRoom.Error()))
		return
	}

	var joinOptions interfaces.CreateRoomOptions

	if err := json.Unmarshal(data, &joinOptions); err != nil {
		client.Send(protocol.NewErrorResponse("join_room_result", ErrGameOptionsInvalid.Error()))
		return
	}

	if len(joinOptions.PlayerName) == 0 {
		client.Send(protocol.NewErrorResponse("join_room_result", ErrPlayerNameRequired.Error()))
		return
	}

	log.Debug().Fields(joinOptions).Msg("client joining room")

	var room interfaces.Room
	if joinOptions.RoomID == nil {
		log.Info().Msg("Room id not provided, creating new room")
		cr, err := r.handleCreateRoom(joinOptions)
		room = cr
		if err != nil {
			log.Error().Err(err).Msg("failed to create room")
			client.Send(protocol.NewErrorResponse("join_room_result", err.Error()))
			return
		}
	}

	if room == nil {
		log.Info().Str("id", *joinOptions.RoomID).Msg("Room id provided, getting room")
		tr, err := r.roomManager.GetRoom(*joinOptions.RoomID)
		room = tr
		if err != nil {
			log.Info().Str("id", *joinOptions.RoomID).Msg("Room not found, creating new room with provided id")
			tr, err = r.handleCreateRoom(joinOptions)
			room = tr
			if err != nil {
				log.Error().Err(err).Str("id", *joinOptions.RoomID).Msg("failed to create new room with provided id")
				client.Send(protocol.NewErrorResponse("join_room_result", err.Error()))
				return
			}
		}
	}

	err := r.gameRegistry.HandleClientJoin(client, room, joinOptions)
	if err != nil {
		client.Send(protocol.NewErrorResponse("join_room_result", err.Error()))
		return
	}

	response := map[string]interface{}{
		"clientId": client.ID(),
		"roomId":   room.ID(),
	}

	log.Info().Str("roomID", room.ID()).Msg("client joined room")

	client.Send(protocol.NewSuccessResponse("join_room_result", response))
	r.broadCastRoomListChange(room.GameType())
}

// handleLeaveRoom leaves the current room
func (r *Router) handleLeaveRoom(client interfaces.Client) {
	room := client.Room()
	if room == nil {
		log.Warn().Str("id", client.ID()).Msg("client tried to leave room but room is not set")
		client.Send(protocol.NewErrorResponse("leave_room_result", ErrClientWithoutRoom.Error()))
		return
	}
	roomID := room.ID()
	log.Debug().Str("clientID", client.ID()).Str("roomID", roomID).Msg("client leaving room")

	// Notify game about client leave
	err := r.gameRegistry.HandleClientLeave(client, room)
	if err != nil {
		log.Error().Err(err).Msg("failed to notify game about client leave")
		client.Send(protocol.NewErrorResponse("leave_room_result", err.Error()))
		return
	}

	room.Leave(client)
	client.SetRoom(nil)

	log.Info().Str("clientId", client.ID()).Str("roomID", roomID).Msg("client left room")

	client.Send(protocol.NewSuccessResponse("leave_room_result", nil))
	r.broadCastRoomListChange(room.GameType())
}

// handleReconnect tries to reconnect the new socket to an existing room
func (r *Router) handleReconnect(client interfaces.Client, data json.RawMessage) {
	if client.Room() != nil {
		client.Send(protocol.NewErrorResponse("reconnect_result", ErrClientAlreadyInRoom.Error()))
		return
	}

	var recon ReconnectPayload
	if err := json.Unmarshal(data, &recon); err != nil {
		client.Send(protocol.NewErrorResponse("reconnect_result", err.Error()))
		return
	}

	log.Debug().Str("oldClientID", recon.ClientID).Str("newClientID", client.ID()).Msg("client reconnecting to room")

	sessionStore := session.GetSessionStore()
	sessionData, exists := sessionStore.GetSession(recon.ClientID)
	if !exists {
		log.Warn().Str("clientId", recon.ClientID).Msg(ErrSessionInvalid.Error())
		client.Send(protocol.NewErrorResponse("reconnect_result", ErrSessionInvalid.Error()))
		// TODO: maybe auto-remove player from room if doesnt reconnect in a while?
		return
	}

	// Find room (either from session or from request)
	roomID := sessionData.RoomID
	if recon.RoomID != "" {
		roomID = recon.RoomID
	}

	// Get the room
	targetRoom, err := r.roomManager.GetRoom(roomID)
	if err != nil {
		log.Error().Str("room", roomID).Msg("room not found")
		client.Send(protocol.NewErrorResponse("reconnect_result", "Room not found"))
		return
	}

	// Join room and reconnect
	if err = targetRoom.Join(client); err != nil {
		log.Error().Str("room", roomID).Err(err).Msg("client failed to join during reconnect")
		client.Send(protocol.NewErrorResponse("reconnect_result", err.Error()))
		return
	}

	// Handle reconnection at game level
	if err = r.gameRegistry.HandleClientReconnect(client, targetRoom, recon.ClientID); err != nil {
		log.Error().Str("room", roomID).Err(err).Msg("game failed to reconnect client")
		client.Send(protocol.NewErrorResponse("reconnect_result", err.Error()))
		return
	}

	// Remove the old session
	sessionStore.RemoveSession(recon.ClientID)

	response := map[string]interface{}{
		"roomId":   targetRoom.ID(),
		"clientId": client.ID(),
		"gameType": targetRoom.GameType(),
	}

	log.Info().Str("roomId", targetRoom.ID()).Str("gameType", targetRoom.GameType()).Msg("client reconnected")

	client.Send(protocol.NewSuccessResponse("reconnect_result", response))
}

// handleGameAction forwards a game-specific action to the game handler
func (r *Router) handleGameAction(client interfaces.Client, data json.RawMessage) {
	if client.Room() == nil {
		client.Send(protocol.NewErrorResponse("game_action_result", ErrClientWithoutRoom.Error()))
		return
	}

	if err := r.gameRegistry.HandleMessage(client, "game_action", data); err != nil {
		client.Send(protocol.NewErrorResponse("game_action_result", err.Error()))
		return
	}
}

// handleAddBot adds a bot to the current room
func (r *Router) handleAddBot(client interfaces.Client) {
	if client.Room() == nil {
		client.Send(protocol.NewErrorResponse("add_bot_result", ErrClientWithoutRoom.Error()))
		return
	}

	err := r.gameRegistry.HandleAddBot(client, client.Room())
	if err != nil {
		client.Send(protocol.NewErrorResponse("add_bot_result", err.Error()))
		return
	}

	log.Info().Str("roomID", client.Room().ID()).Msg("bot added to room")

	client.Send(protocol.NewSuccessResponse("add_bot_result", nil))
}

// getRoomList generates a list of room information for a specific game type
func (r *Router) getRoomList(gameType string) []RoomListInfo {
	rooms := r.roomManager.GetAllRoomsByGameType(gameType)

	roomList := make([]RoomListInfo, 0)
	for _, room := range rooms {
		clients := room.Clients()

		// Safely check the Started property from room state
		started := false
		if state := room.State(); state != nil {
			if stateMap, ok := state.(map[string]interface{}); ok {
				if startedVal, exists := stateMap["Started"]; exists {
					started, _ = startedVal.(bool)
				}
			}
		}

		roomInfo := RoomListInfo{
			RoomId:      room.ID(),
			PlayerCount: len(clients),
			GameStarted: started,
		}
		roomList = append(roomList, roomInfo)
	}

	return roomList
}

func (r *Router) broadCastRoomListChange(gameType string) {
	// find all clients that are connected to a certain game type and inform them of the room list change
	roomList := r.getRoomList(gameType)
	response := protocol.NewSuccessResponse("room_list_update", roomList)

	gameClients := r.clientManager.GetClientsByGameType(gameType)
	r.BroadcastTo(response, gameClients)
}

// handleGetRoomList sends the current room list for a game type to the requesting client
func (r *Router) handleGetRoomList(client interfaces.Client, data json.RawMessage) {
	var request struct {
		GameType string `json:"gameType"`
	}

	if err := json.Unmarshal(data, &request); err != nil {
		client.Send(protocol.NewErrorResponse("get_room_list_result", "Invalid request format"))
		return
	}

	if request.GameType == "" {
		client.Send(protocol.NewErrorResponse("get_room_list_result", "Game type is required"))
		return
	}

	roomList := r.getRoomList(request.GameType)
	response := protocol.NewSuccessResponse("get_room_list_result", roomList)
	client.Send(response)
}

// BroadcastTo sends a message to specific clients
func (r *Router) BroadcastTo(message *protocol.Response, clients []interfaces.Client) {
	for _, client := range clients {
		client.Send(message)
	}
}

// Error definitions
var (
	ErrClientAlreadyInRoom = errors.New("client is already in a room")
	ErrClientWithoutRoom   = errors.New("client is not in a room")
	ErrSessionInvalid      = errors.New("session expired or not found")
	ErrMessageInvalid      = errors.New("invalid message format")

	ErrGameTypeRequired   = errors.New("game type is required")
	ErrGameOptionsInvalid = errors.New("game options are invalid")
	ErrPlayerNameRequired = errors.New("player name is required")
)
