package owe_drahn

import (
	"encoding/json"
	"errors"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"github.com/rs/zerolog/log"
)

func NewGame() *Game {
	return &Game{}
}

func RegisterGame(r interfaces.GameRegistry) {
	g := NewGame()
	r.RegisterGame(g)
}

// Type returns the game type
func (g *Game) Type() string {
	return "owedrahn"
}

// InitializeRoom sets up a new room with the initial game state
func (g *Game) InitializeRoom(room interfaces.Room, options json.RawMessage) error {
	// Create initial game state
	state := GameState{
		Players:     make(map[string]*Player),
		Started:     false,
		CurrentTurn: "",
	}

	room.SetState(&state)
	return nil
}

func (g *Game) OnClientJoin(client interfaces.Client, room interfaces.Room, options interfaces.CreateRoomOptions) {
	state := room.State().(*GameState)

	// If the game has started, the new client becomes a spectator
	if state.Started {
		client.Send(protocol.NewErrorResponse("error", "game has started"))
		return
	}

	g.AddPlayer(client.ID(), options.PlayerName, state)

	room.SetState(state)

	broadcastGameState(room)
}

func (g *Game) OnBotAdd(client interfaces.Client, room interfaces.Room, reg interfaces.GameRegistry) (interfaces.Client, error) {
	return nil, errors.New("game does not support bots")
}

func (g *Game) OnClientLeave(client interfaces.Client, room interfaces.Room) {
	state := room.State().(*GameState)
	player := g.GetPlayer(client.ID(), state)
	if player == nil {
		log.Error().Msg("player not found")
		return
	}
	player.IsConnected = false
	room.SetState(state)
}

// OnClientReconnect handles reconnecting a client to the game
func (g *Game) OnClientReconnect(client interfaces.Client, room interfaces.Room, oldClientID string) {
	state := room.State().(*GameState)

	// Check if the old client ID was a player in this game
	playerInfo, exists := state.Players[oldClientID]
	if !exists {
		client.Send(protocol.NewErrorResponse("error", "No player found with the provided ID"))
		return
	}

	// Replace the old client ID with the new one, maintaining the same player info
	delete(state.Players, oldClientID)
	playerInfo.ID = client.ID()
	state.Players[client.ID()] = playerInfo

	// If it was this player's turn, update the current turn
	if state.CurrentTurn == oldClientID {
		state.CurrentTurn = client.ID()
	}

	room.SetState(state)

	// tell the new client the game state
	msg := protocol.NewSuccessResponse("game_state", state)
	client.Send(msg)
}

func (g *Game) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, payload []byte) {
	state := room.State().(*GameState)
	// Validate it's the player's turn
	if state.CurrentTurn != client.ID() {
		client.Send(protocol.NewErrorResponse("error", "Not your turn"))
		return
	}

	log.Debug().Str("type", msgType).Str("payload", string(payload)).Msg("handling message")

	switch msgType {
	case "ready":
		g.handleReady(client, room, payload)
		break
	case "loseLife":
		g.handleLoseLife(client, room)
		break
	default:
		client.Send(protocol.NewErrorResponse("error", "Unknown message type: "+msgType))
	}

	broadcastGameState(room)
}

// broadcastGameState sends the current game state to all clients in the room
func broadcastGameState(room interfaces.Room) {
	state := room.State()

	msg := protocol.NewSuccessResponse("game_state", state)
	room.Broadcast(msg)
}
