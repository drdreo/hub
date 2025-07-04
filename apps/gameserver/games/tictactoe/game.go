package tictactoe

import (
	"context"
	"encoding/json"
	"errors"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"math/rand"

	"github.com/rs/zerolog/log"
)

// TicTacToe implements the game interface
type TicTacToe struct{}

// GameState represents the state of a tic tac toe game
type GameState struct {
	Board       [3][3]string          `json:"board"`
	Players     map[string]PlayerInfo `json:"players"`
	CurrentTurn string                `json:"currentTurn"`
	Winner      string                `json:"winner"`
	GameOver    bool                  `json:"gameOver"`
	DrawGame    bool                  `json:"drawGame"`
}

// PlayerInfo stores player information
type PlayerInfo struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

// MovePayload represents a move action from a client
type MovePayload struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// NewTicTacToe creates a new tic tac toe game
func NewTicTacToe() *TicTacToe {
	return &TicTacToe{}
}

func RegisterTicTacToeGame(r interfaces.GameRegistry) {
	g := NewTicTacToe()
	r.RegisterGame(g)
}

// Type returns the game type
func (g *TicTacToe) Type() string {
	return "tictactoe"
}

// InitializeRoom sets up a new room with the initial game state
func (g *TicTacToe) InitializeRoom(ctx context.Context, room interfaces.Room, options json.RawMessage) error {
	// Create initial game state
	state := GameState{
		Board:       [3][3]string{{"", "", ""}, {"", "", ""}, {"", "", ""}},
		Players:     make(map[string]PlayerInfo),
		CurrentTurn: "",
		Winner:      "",
		GameOver:    false,
		DrawGame:    false,
	}

	room.SetState(state)
	return nil
}

// OnClientJoin handles a client joining the room
func (g *TicTacToe) OnClientJoin(client interfaces.Client, room interfaces.Room, _ interfaces.CreateRoomOptions) {
	state := room.State().(GameState)

	// Only allow 2 players
	if len(state.Players) >= 2 {
		client.Send(protocol.NewErrorResponse("error", "Game is full"))
		return
	}

	// Assign symbol (X for first player, O for second)
	var symbol string
	if len(state.Players) == 0 {
		symbol = "X"
	} else {
		symbol = "O"
	}

	// Add player to game state
	state.Players[client.ID()] = PlayerInfo{
		Symbol: symbol,
		Name:   "Player " + symbol,
	}

	// If we now have 2 players, start the game
	if len(state.Players) == 2 {
		// Randomly select first player
		playerIDs := make([]string, 0, len(state.Players))
		for id := range state.Players {
			playerIDs = append(playerIDs, id)
		}
		state.CurrentTurn = playerIDs[rand.Intn(len(playerIDs))]
	}

	// Update state
	room.SetState(state)

	// Broadcast updated state to all clients
	broadcastGameState(room)

	// Send welcome message
	client.Send(protocol.NewSuccessResponse("joined", interfaces.M{
		"clientId": client.ID(),
		"symbol":   state.Players[client.ID()].Symbol,
		"roomId":   room.ID(),
	}))
}

func (g *TicTacToe) OnBotAdd(client interfaces.Client, room interfaces.Room, reg interfaces.GameRegistry) (interfaces.Client, string, error) {
	return nil, "", errors.New("game does not support bots")
}

// OnClientLeave handles a client leaving the room
func (g *TicTacToe) OnClientLeave(client interfaces.Client, room interfaces.Room) {
	state := room.State().(GameState)

	// Remove player from game
	delete(state.Players, client.ID())

	// If game was in progress, end it
	if !state.GameOver && len(state.Players) < 2 {
		state.GameOver = true
	}

	// Update state
	room.SetState(state)

	// Broadcast to remaining players
	broadcastGameState(room)
}

// OnClientReconnect handles reconnecting a client to the game
func (g *TicTacToe) OnClientReconnect(client interfaces.Client, room interfaces.Room, oldClientID string) error {
	state := room.State().(GameState)

	// Check if the old client ID was a player in this game
	playerInfo, exists := state.Players[oldClientID]
	if !exists {
		return errors.New("no player found with provided ID")
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
		"symbol":   playerInfo.Symbol,
		"roomId":   room.ID(),
	}))

	return nil
}

// HandleMessage processes game-specific messages
func (g *TicTacToe) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, payload []byte) error {
	switch msgType {
	case "make_move":
		g.handleMakeMove(client, room, payload)
	case "restart_game":
		g.handleRestartGame(client, room)
	default:
		client.Send(protocol.NewErrorResponse("error", "Unknown message type: "+msgType))
	}

	return nil
}

// handleMakeMove processes a move from a player
func (g *TicTacToe) handleMakeMove(client interfaces.Client, room interfaces.Room, payload []byte) {
	// Parse move payload
	var move MovePayload
	if err := json.Unmarshal(payload, &move); err != nil {
		client.Send(protocol.NewErrorResponse("error", "Invalid move format"))
		return
	}

	log.Debug().Str("clientID", client.ID()).Msg("player makes a move")

	// Get current game state
	state := room.State().(GameState)

	// Check if it's game over
	if state.GameOver {
		client.Send(protocol.NewErrorResponse("error", "Game is over"))
		return
	}

	// Check if it's the player's turn
	if state.CurrentTurn != client.ID() {
		log.Warn().Str("current", state.CurrentTurn).Str("clientID", client.ID()).Msg("NOT YOUR TURN")
		client.Send(protocol.NewErrorResponse("error", "Not your turn"))
		return
	}

	// Validate move
	if move.Row < 0 || move.Row > 2 || move.Col < 0 || move.Col > 2 {
		client.Send(protocol.NewErrorResponse("error", "Invalid move coordinates"))
		return
	}

	// Check if cell is empty
	if state.Board[move.Row][move.Col] != "" {
		client.Send(protocol.NewErrorResponse("error", "Cell already occupied"))
		return
	}

	// Make the move
	state.Board[move.Row][move.Col] = state.Players[client.ID()].Symbol

	// Check for win condition
	if checkWin(state.Board) {
		state.Winner = client.ID()
		state.GameOver = true

		log.Info().Str("winner", client.ID()).Msg("game over")
	} else if checkDraw(state.Board) {
		state.DrawGame = true
		state.GameOver = true
		log.Info().Msg("game draw")
	} else {
		// Switch turns
		for id := range state.Players {
			if id != client.ID() {
				state.CurrentTurn = id
				break
			}
		}

		log.Debug().Str("current", state.CurrentTurn).Msg("updated current player")
	}

	room.SetState(state)
	broadcastGameState(room)
}

// handleRestartGame resets the game
func (g *TicTacToe) handleRestartGame(client interfaces.Client, room interfaces.Room) {
	log.Debug().Msg("restarting")

	state := room.State().(GameState)

	// Only allow restart if game is over
	if !state.GameOver {
		client.Send(protocol.NewErrorResponse("error", "Cannot restart a game in progress"))
		return
	}

	// Reset the board
	state.Board = [3][3]string{{"", "", ""}, {"", "", ""}, {"", "", ""}}
	state.Winner = ""
	state.GameOver = false
	state.DrawGame = false

	// Randomly select first player
	playerIDs := make([]string, 0, len(state.Players))
	for id := range state.Players {
		playerIDs = append(playerIDs, id)
	}
	state.CurrentTurn = playerIDs[rand.Intn(len(playerIDs))]

	// Update state
	room.SetState(state)

	// Broadcast updated state
	broadcastGameState(room)
}

// checkWin returns true if there's a winning condition on the board
func checkWin(board [3][3]string) bool {
	// Check rows
	for i := 0; i < 3; i++ {
		if board[i][0] != "" && board[i][0] == board[i][1] && board[i][1] == board[i][2] {
			return true
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if board[0][i] != "" && board[0][i] == board[1][i] && board[1][i] == board[2][i] {
			return true
		}
	}

	// Check diagonals
	if board[0][0] != "" && board[0][0] == board[1][1] && board[1][1] == board[2][2] {
		return true
	}
	if board[0][2] != "" && board[0][2] == board[1][1] && board[1][1] == board[2][0] {
		return true
	}

	return false
}

// checkDraw returns true if the board is full
func checkDraw(board [3][3]string) bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] == "" {
				return false
			}
		}
	}
	return true
}

// broadcastGameState sends the current game state to all clients in the room
func broadcastGameState(room interfaces.Room) {
	state := room.State()

	// Create a public view of the game state that hides sensitive info

	msg := protocol.NewSuccessResponse("game_state", state)
	room.Broadcast(msg)
}
