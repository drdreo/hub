package dicegame

import (
	"encoding/json"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"strings"
	"time"

	// 	"math/rand"

	"github.com/rs/zerolog/log"
)

const BustedAnimationDelay = 3 * time.Second // (~2-3 seconds)

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
		Dice:         make([]int, 0, 6),
		SelectedDice: make([]int, 0),
		SetAside:     make([]int, 0),
		Started:      false,
		CurrentTurn:  "",
		Winner:       "",
		TargetScore:  3000,
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
	if len(state.Players) == 2 {
		state.Started = true

		// Randomly select first player
		playerIDs := make([]string, 0, len(state.Players))
		for id := range state.Players {
			playerIDs = append(playerIDs, id)
		}
		// TODO: put back
		// 		state.CurrentTurn = playerIDs[rand.Intn(len(playerIDs))]
		for _, player := range state.Players {
			log.Debug().Str("name", player.Name).Msg("SEE ME")
			if !strings.Contains(player.Name, "Bot") {
				state.CurrentTurn = player.ID
			}
		}
	}

	room.SetState(state)

	// Broadcast updated state to all clients
	broadcastGameState(room)
}

func (g *DiceGame) OnBotAdd(client interfaces.Client, room interfaces.Room, reg interfaces.GameRegistry) (interfaces.Client, error) {
	bot := NewDiceGameBot("bot-1", g, reg)

	return bot.BotClient, nil
}

func (g *DiceGame) OnClientLeave(client interfaces.Client, room interfaces.Room) {
	state := room.State().(*GameState)
	if state.CurrentTurn == client.ID() {
		log.Info().Str("clientId", client.ID()).Msg("client left. Ending players turn")
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
	playerInfo.ID = client.ID()
	state.Players[client.ID()] = playerInfo

	// If it was this player's turn, update the current turn
	if state.CurrentTurn == oldClientID {
		state.CurrentTurn = client.ID()
	}

	// Update the winner reference if applicable
	if state.Winner == oldClientID {
		state.Winner = client.ID()
	}

	room.SetState(state)

	// tell the new client the game state
	msg := protocol.NewSuccessResponse("game_state", state)
	client.Send(msg)
}

func (g *DiceGame) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, payload []byte) {
	state := room.State().(*GameState)
	// Validate it's the player's turn
	if state.CurrentTurn != client.ID() {
		client.Send(protocol.NewErrorResponse("error", "Not your turn"))
		return
	}

	log.Debug().Str("type", msgType).Str("payload", string(payload)).Msg("handling message")

	switch msgType {
	case "roll":
		if busted := g.handleRoll(room); busted {
			log.Info().Msg("Player busted on roll")
			// Schedule the turn end after a delay to allow for animations
			go func(roomID string, playerID string) {
				// Wait for dice animation
				time.Sleep(BustedAnimationDelay)
				bustedMsg := protocol.NewErrorResponse("error", "Busted")
				room.Broadcast(bustedMsg)
				g.handleEndTurn(room)
				broadcastGameState(room)
			}(room.ID(), client.ID())
		}
	case "select":
		var action SelectActionPayload
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &action); err != nil {
				log.Error().Str("error", err.Error()).Msg("invalid select payload")
				client.Send(protocol.NewErrorResponse("error", "Invalid select payload"))
				return
			}
		}

		log.Debug().Int("diceIndex", action.DiceIndex).Int("length", len(state.Dice)).Msg("select coming in")

		if err := g.handleSelect(room, action); err != nil {
			log.Error().Int("diceIndex", action.DiceIndex).Int("length", len(state.Dice)).Msg(err.Error())
			client.Send(protocol.NewErrorResponse("error", "Select failed; "+err.Error()))
		}
	case "set_aside":
		var action SetAsideActionPayload
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &action); err != nil {
				log.Error().Err(err).Msg("invalid set aside payload")
				client.Send(protocol.NewErrorResponse("error", "Invalid set aside payload"))
				return
			}
		}

		log.Info().Bool("endTurn", action.EndTurn).Ints("selectedDice", state.SelectedDice).Ints("dice", state.Dice).Msg("setting dice aside")

		if err := g.handleSetAside(room, action); err != nil {
			log.Error().Err(err).Msg("set aside failed; " + err.Error())
			client.Send(protocol.NewErrorResponse("error", "set aside failed; "+err.Error()))
		}
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
