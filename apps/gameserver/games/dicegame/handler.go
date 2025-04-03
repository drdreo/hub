package dicegame

import (
	"encoding/json"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"math/rand"

	"github.com/rs/zerolog/log"
)

func NewDiceGame() *DiceGame {
	return &DiceGame{}
}

func RegisterDiceGame(r interfaces.GameRegistry) {
	g := NewDiceGame()
	r.RegisterGame(g)
}

// Type returns the game type
func (g *DiceGame) Type() string {
	return "dicegame"
}

// InitializeRoom sets up a new room with the initial game state
func (g *DiceGame) InitializeRoom(room interfaces.Room, options json.RawMessage) error {
	// Create initial game state
	state := GameState{
		Players:      make(map[string]*Player),
		Dice:         make([]int, 6),
		SelectedDice: make([]int, 0),
		SetAside:     make([]int, 0),
		CurrentTurn:  "",
		Winner:       "",
		TargetScore:  3000,
		TurnScore:    0,
		RoundScore:   0,
	}

	room.SetState(&state)
	return nil
}

func (g *DiceGame) OnClientJoin(client interfaces.Client, room interfaces.Room, options interfaces.CreateRoomOptions) {
	state := room.State().(*GameState)

	// Only allow 2 players
	if len(state.Players) >= 2 {
		client.Send(protocol.NewErrorResponse("error", "Game is full"))
		return
	}

	g.AddPlayer(client.ID(), options.PlayerName, state)

	// If we now have 2 players, start the game
	// TODO: revert to 2 players limit
	if len(state.Players) == 1 {
		// Randomly select first player
		playerIDs := make([]string, 0, len(state.Players))
		for id := range state.Players {
			playerIDs = append(playerIDs, id)
		}
		state.CurrentTurn = playerIDs[rand.Intn(len(playerIDs))]
	}

	room.SetState(state)

	// Broadcast updated state to all clients
	broadcastGameState(room)
}

func (g *DiceGame) OnClientLeave(client interfaces.Client, room interfaces.Room) {
	state := room.State().(*GameState)
	if state.CurrentTurn == client.ID() {
		g.EndTurn(state)
	}
	room.SetState(state)
}

// OnClientReconnect handles reconnecting a client to the game
func (g *DiceGame) OnClientReconnect(client interfaces.Client, room interfaces.Room, oldClientID string) {
	state := room.State().(*GameState)

	// Check if the old client ID was a player in this game
	playerInfo, exists := state.Players[oldClientID]
	if !exists {
		client.Send(protocol.NewErrorResponse("error", "No player found with the provided ID"))
		return
	}

	// Replace the old client ID with the new one, maintaining the same player info
	delete(state.Players, oldClientID)
	state.Players[client.ID()] = playerInfo

	// If it was this player's turn, update the current turn
	if state.CurrentTurn == oldClientID {
		state.CurrentTurn = client.ID()
	}

	// Update the winner reference if applicable
	if state.Winner == oldClientID {
		state.Winner = client.ID()
	}

	// Update state
	room.SetState(state)

	// Broadcast updated state to all clients
	broadcastGameState(room)
}

func (g *DiceGame) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, payload []byte) {
	state := room.State().(*GameState)
	// Validate it's the player's turn
	if state.CurrentTurn != client.ID() {
		client.Send(protocol.NewErrorResponse("error", "Not your turn"))
		return
	}

	switch msgType {
	case "roll":
		g.handleRoll(room)
	case "select":
		var action SelectActionPayload
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &action); err != nil {
				// Only log the error, but continue with default empty action
				log.Error().Str("error", err.Error()).Msg("invalid select payload")
				client.Send(protocol.NewErrorResponse("error", "Invalid select payload"))
				return
			}
		}
		g.handleSelect(room, action)
	case "set_aside":
		var action SetAsideActionPayload
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &action); err != nil {
				// Only log the error, but continue with default empty action
				log.Error().Str("error", err.Error()).Msg("invalid set aside payload")
				client.Send(protocol.NewErrorResponse("error", "Invalid set aside payload"))
				return
			}
		}
		g.handleSetAside(room, action)
	case "end_turn":
		g.handleEndTurn(room)
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
