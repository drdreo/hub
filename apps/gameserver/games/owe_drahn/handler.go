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

	g.broadcastGameState(room)
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
	oldPlayer, exists := state.Players[oldClientID]
	if !exists {
		client.Send(protocol.NewErrorResponse("error", "No player found with the provided ID"))
		return
	}

	// Update player ID and state map references
	oldPlayer.ID = client.ID()
	state.Players[client.ID()] = oldPlayer
	delete(state.Players, oldClientID)

	// If it was this player's turn, update the current turn
	if state.CurrentTurn == oldClientID {
		state.CurrentTurn = client.ID()
	}

	room.SetState(state)

	// tell the new client the game state
	msg := protocol.NewSuccessResponse("game_state", state)
	client.Send(msg)
}

func (g *Game) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, payload []byte) error {
	state := room.State().(*GameState)

	log.Debug().Str("type", msgType).Bytes("payload", payload).Msg("handling message")

	// TODO: revisit during game actions vs. non handling
	if msgType == "handshake" {
		g.handleHandshake(client, state, payload)
	} else if msgType == "ready" {
		g.handleReady(client, state, payload)
	} else {
		// Validate it's the player's turn
		if state.CurrentTurn != client.ID() {
			log.Warn().Str("clientId", client.ID()).Str("currentTurn", state.CurrentTurn).Msg(ErrNotYourTurn.Error())
			return ErrNotYourTurn
		}

		// current turn action handling
		switch msgType {
		case "roll":
			if err := g.handleRoll(client, state); err != nil {
				log.Error().Err(err).Msg("roll failed")
				return ErrRollFailed
			}
			break
		case "loseLife":
			g.handleLoseLife(client, state)
			g.broadcastGameState(room)
			break
		case "chooseNextPlayer":
			if err := g.handleChooseNextPlayer(client, state, payload); err != nil {
				log.Error().Err(err).Msg("chooseNextPlayer failed")
				return ErrNextPlayerInvalid
			}

			break
		default:
			return errors.New("unknown message type: " + msgType)
		}
	}
	return nil
}

func (g *Game) broadcastGameEvent(room interfaces.Room, eventName string, payload interfaces.M) {
	msg := protocol.NewSuccessResponse(eventName, payload)
	room.Broadcast(msg)
}

// broadcastGameState sends the current game state to all clients in the room
func (g *Game) broadcastGameState(room interfaces.Room) {
	state := room.State().(*GameState)
	g.broadcastGameEvent(room, "game_state", state.ToMap())
}

func (g *Game) broadcastPlayerUpdate(room interfaces.Room, players map[string]*Player, currentTurn string, updateUI bool) {
	g.broadcastGameEvent(room, "playerUpdate", interfaces.M{
		"currentTurn": currentTurn,
		"players":     mapPlayersToSortedSlice(players),
		"updateUI":    updateUI,
	})
}

var (
	ErrNotYourTurn       = errors.New("not your turn")
	ErrRollFailed        = errors.New("roll failed")
	ErrNextPlayerInvalid = errors.New("next player is invalid")
)
