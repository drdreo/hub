package dicegame

import (
	"context"
	"encoding/json"
	"errors"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"time"

	// 	"math/rand"

	"github.com/rs/zerolog/log"
)

const BustedAnimationDelay = 4 * time.Second // (~3-4 seconds)

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
func (g *DiceGame) InitializeRoom(ctx context.Context, room interfaces.Room, options json.RawMessage) error {
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
		g.start(state)
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
		g.EndTurn(room, state)
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

	room.SetState(state)

	// tell the new client the game state
	msg := protocol.NewSuccessResponse("game_state", state)
	client.Send(msg)
}

func (g *DiceGame) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, payload []byte) error {
	state := room.State().(*GameState)
	// Validate it's the player's turn
	if state.CurrentTurn != client.ID() {
		client.Send(protocol.NewErrorResponse("error", ErrNotYourTurn.Error()))
		return ErrNotYourTurn
	}

	log.Debug().Str("type", msgType).Bytes("payload", payload).Msg("handling message")

	switch msgType {
	case "roll":
		if busted := g.handleRoll(room); busted {
			log.Info().Msgf("%s busted on roll", state.Players[state.CurrentTurn].Name)

			// using `clientId + busted message` since bot filters if it was its bust or not
			bustedMsg := protocol.NewErrorResponse("error", client.ID()+ErrBusted.Error())

			// if the bot busted, tell them immediately
			if client.IsBot() {
				room.BroadcastTo(bustedMsg, client)
			}

			bustedPlayer := state.Players[state.CurrentTurn]

			// Schedule the turn end after a delay to allow for animations
			time.AfterFunc(BustedAnimationDelay, func() {
				room.Broadcast(bustedMsg)
				bustedPlayer.TurnScore = 0
				bustedPlayer.RoundScore = 0
				g.handleEndTurn(room)
				broadcastGameState(room)
			})
		}
	case "select":
		var action SelectActionPayload
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &action); err != nil {
				log.Error().Str("error", err.Error()).Msg(ErrSelectPayloadInvalid.Error())
				return ErrSelectPayloadInvalid
			}
		}

		log.Debug().Int("diceIndex", action.DiceIndex).Int("length", len(state.Dice)).Msg("select coming in")

		if err := g.handleSelect(room, action); err != nil {
			log.Error().Int("diceIndex", action.DiceIndex).Int("length", len(state.Dice)).Err(err).Msg(ErrSelectInvalid.Error())
			return ErrSelectInvalid
		}
	case "set_aside":
		var action SetAsideActionPayload
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &action); err != nil {
				log.Error().Err(err).Msg(ErrSetAsidePayloadInvalid.Error())
				return ErrSetAsidePayloadInvalid
			}
		}

		log.Info().Bool("endTurn", action.EndTurn).Ints("selectedDice", state.SelectedDice).Ints("dice", state.Dice).Msg("setting dice aside")

		if err := g.handleSetAside(room, action.EndTurn); err != nil {
			log.Error().Err(err).Msg(ErrSetAsideInvalid.Error())
			return ErrSetAsideInvalid
		}
	//case "end_turn":
	//	g.handleEndTurn(room)
	default:
		client.Send(protocol.NewErrorResponse("error", "Unknown message type: "+msgType))
	}

	broadcastGameState(room)
	return nil
}

// broadcastGameState sends the current game state to all clients in the room
func broadcastGameState(room interfaces.Room) {
	state := room.State()

	msg := protocol.NewSuccessResponse("game_state", state)
	room.Broadcast(msg)
}

var (
	ErrNotYourTurn            = errors.New("not your turn")
	ErrBusted                 = errors.New("busted")
	ErrSelectPayloadInvalid   = errors.New("select payload invalid")
	ErrSelectInvalid          = errors.New("select invalid")
	ErrSetAsidePayloadInvalid = errors.New("set aside payload invalid ")
	ErrSetAsideInvalid        = errors.New("set aside invalid")
)
