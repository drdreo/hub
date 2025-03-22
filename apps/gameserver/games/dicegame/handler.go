package dicegame

import (
	"encoding/json"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"math/rand"
)

type Handler struct {
	game *DiceGame
}

func NewHandler() *Handler {
	return &Handler{
		game: NewDiceGame(),
	}
}

func (h *Handler) RegisterDiceGame(r interfaces.GameRegistry) {
	r.RegisterGame(h)
}

func (h *Handler) InitializeRoom(room interfaces.Room, options json.RawMessage) error {
	return nil
}

func (h *Handler) Type() string {
	return "dicegame"
}

func (h *Handler) OnClientJoin(client interfaces.Client, room interfaces.Room) {
	state := room.State().(GameState)

	// Only allow 2 players
	if len(state.Players) >= 2 {
		client.Send(protocol.NewErrorResponse("error", "Game is full"))
		return
	}

	h.game.AddPlayer(client.ID(), state)

	// If we now have 2 players, start the game
	if len(state.Players) == 2 {
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

	client.Send(protocol.NewSuccessResponse("joined", interfaces.M{
		"clientId": client.ID(),
		"roomId":   room.ID(),
	}))
}

func (h *Handler) OnClientLeave(client interfaces.Client, room interfaces.Room) {
	state := room.State().(GameState)
	if state.CurrentTurn == client.ID() {
		h.game.EndTurn(state)
	}
	room.SetState(state)
}

// OnClientReconnect handles reconnecting a client to the game
func (g *Handler) OnClientReconnect(client interfaces.Client, room interfaces.Room, oldClientID string) {
	state := room.State().(GameState)

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

	// Send welcome back message to the reconnected client
	client.Send(protocol.NewSuccessResponse("reconnected", interfaces.M{
		"clientId": client.ID(),
		"roomId":   room.ID(),
	}))
}

func (h *Handler) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, payload []byte) {
	var action GameAction
	if err := json.Unmarshal(payload, &action); err != nil {
		client.Send(protocol.NewErrorResponse("error", "Invalid payload format"))
		return
	}

	h.game.HandleMessage(client, room, msgType, action)
	broadcastGameState(room)
}

// broadcastGameState sends the current game state to all clients in the room
func broadcastGameState(room interfaces.Room) {
	state := room.State()

	msg := protocol.NewSuccessResponse("game_state", state)
	room.Broadcast(msg)
}
