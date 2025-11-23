package owe_drahn

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rs/zerolog/log"

	"gameserver/games/owe_drahn/database"
	"gameserver/games/owe_drahn/models"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
)

type GameConfig struct {
	Stage          interfaces.Environment
	CredentialsDir string
}

func NewGame(dbService database.Database) *Game {
	return &Game{
		dbService: dbService,
	}
}

func RegisterGame(ctx context.Context, r interfaces.GameRegistry, config GameConfig) error {
	dbInitCtx, dbInitCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbInitCancel()

	dbFactory := database.NewDatabaseFactory(config.Stage, config.CredentialsDir)
	dbService, err := dbFactory.CreateDatabaseService(dbInitCtx)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize database service")
		return err
	}

	g := NewGame(dbService)
	r.RegisterGame(g)
	return nil
}

// Type returns the game type
func (g *Game) Type() string {
	return "owedrahn"
}

// InitializeRoom sets up a new room with the initial game state
func (g *Game) InitializeRoom(ctx context.Context, room interfaces.Room, options json.RawMessage) error {
	// Create initial game state
	state := GameState{
		Ctx:         ctx,
		Players:     make(map[string]*Player),
		PlayerOrder: make([]string, 0),
		Started:     false,
		CurrentTurn: "",
		SideBets:    make([]*models.SideBet, 0),
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

func (g *Game) OnBotAdd(client interfaces.Client, room interfaces.Room, reg interfaces.GameRegistry) (interfaces.Client, string, error) {
	return nil, "", errors.New("game does not support bots")
}

func (g *Game) OnClientLeave(client interfaces.Client, room interfaces.Room) {
	state := room.State().(*GameState)
	player := g.GetPlayer(client.ID(), state)
	if player == nil {
		log.Error().Msg("player not found")
		return
	}
	if !state.Started {
		log.Info().Str("clientID", client.ID()).Msg("game has not started yet. removing player")
		g.RemovePlayer(client.ID(), room)
	} else {
		player.IsConnected = false

		log.Info().Str("clientID", client.ID()).Msg("Someone left")
		g.broadcastGameEvent(room, "playerLeft", interfaces.M{
			"username": player.Name,
		})
	}
	room.SetState(state)
	g.broadcastGameState(room)
}

// OnClientReconnect handles reconnecting a client to the game
func (g *Game) OnClientReconnect(client interfaces.Client, room interfaces.Room, oldClientID string) error {
	state := room.State().(*GameState)

	// Check if the old client ID was a player in this game
	oldPlayer, exists := state.Players[oldClientID]
	if !exists {
		return errors.New("no player found with provided ID")
	}

	// Update player ID and state map references
	oldPlayer.ID = client.ID()
	state.Players[client.ID()] = oldPlayer
	delete(state.Players, oldClientID)

	// Update the new player ID in the player order
	for i, id := range state.PlayerOrder {
		if id == oldClientID {
			state.PlayerOrder[i] = client.ID()
			break
		}
	}

	// If it was this player's turn, update the current turn
	if state.CurrentTurn == oldClientID {
		state.CurrentTurn = client.ID()
	}

	// patch all sideBet ids too
	for i, bet := range state.SideBets {
		if bet.ChallengerID == oldClientID {
			state.SideBets[i].ChallengerID = client.ID()
		}
		if bet.OpponentID == oldClientID {
			state.SideBets[i].OpponentID = client.ID()
		}
	}

	room.SetState(state)
	g.broadcastGameState(room)
	return nil
}

func (g *Game) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, payload []byte) error {
	state := room.State().(*GameState)
	log.Debug().Str("type", msgType).Bytes("payload", payload).Msg("handling message")

	// Handle pre-game messages that don't require turn validation
	switch msgType {
	case "handshake":
		err := g.handleHandshake(client, state, payload)
		if err != nil {
			log.Error().Err(err).Msg("ready failed")
			return err
		}
		g.broadcastGameState(room)
		return nil
	case "ready":
		err := g.handleReady(client, state, payload)
		if err != nil {
			log.Error().Err(err).Msg("ready failed")
			return err
		}

		g.broadcastGameState(room)
		return nil

	case "sidebet_propose":
		betId, err := g.handleProposeSideBet(client, state, payload)
		if err != nil {
			log.Error().Err(err).Msg("bet proposal failed")
			return err
		}
		bets := state.SideBets
		g.broadcastGameEvent(room, "sidebet_propose_result", interfaces.M{
			"bets":  bets,
			"betId": betId,
		})
		return nil
	case "sidebet_accept":
		betId, err := g.handleAcceptSideBet(client, state, payload)
		if err != nil {
			log.Error().Err(err).Msg("bet accept failed")
			return err
		}
		bets := state.SideBets
		g.broadcastGameEvent(room, "sidebet_accept_result", interfaces.M{
			"bets":  bets,
			"betId": betId,
		})
		return nil
	case "sidebet_decline":
		betId, err := g.handleDeclineSideBet(client, state, payload)
		if err != nil {
			log.Error().Err(err).Msg("bet decline failed")
			return err
		}
		bets := state.SideBets
		g.broadcastGameEvent(room, "sidebet_decline_result", interfaces.M{
			"bets":  bets,
			"betId": betId,
		})
		return nil

	case "sidebet_cancel":
		err := g.handleCancelSideBet(state, payload)
		if err != nil {
			log.Error().Err(err).Msg("bet cancel failed")
			return err
		}
		bets := state.SideBets
		g.broadcastGameEvent(room, "sidebet_cancel_result", bets)
		return nil
	}

	// Handle in-game actions
	// Validate it's the player's turn
	if state.CurrentTurn != client.ID() {
		log.Warn().Str("clientId", client.ID()).Str("currentTurn", state.CurrentTurn).Msg(ErrNotYourTurn.Error())
		return ErrNotYourTurn
	}

	// current turn action handling
	var err error
	switch msgType {
	case "roll":
		if err = g.handleRoll(client.Room()); err != nil {
			log.Error().Err(err).Msg("roll failed")
			return ErrRollFailed
		}
	case "loseLife":
		g.handleLoseLife(client, state)
	case "chooseNextPlayer":
		if err = g.handleChooseNextPlayer(state, payload); err != nil {
			log.Error().Err(err).Msg("chooseNextPlayer failed")
			return ErrNextPlayerInvalid
		}
	default:
		return errors.New("unknown message type: " + msgType)
	}

	g.broadcastGameState(room)
	return nil
}

func (g *Game) broadcastGameEvent(room interfaces.Room, eventName string, payload interface{}) {
	msg := protocol.NewSuccessResponse(eventName, payload)
	room.Broadcast(msg)
}

// broadcastGameState sends the current game state to all clients in the room
func (g *Game) broadcastGameState(room interfaces.Room) {
	state := room.State().(*GameState)
	g.broadcastGameEvent(room, "game_state", state.ToDTO())
}

func (g *Game) broadcastPlayerUpdate(room interfaces.Room, players map[string]*Player, playersOrder []string, currentTurn string, updateUI bool) {
	g.broadcastGameEvent(room, "playerUpdate", interfaces.M{
		"currentTurn": currentTurn,
		"players":     mapPlayersToArray(players, playersOrder),
		"updateUI":    updateUI,
	})
}

var (
	ErrNotYourTurn       = errors.New("not your turn")
	ErrRollFailed        = errors.New("roll failed")
	ErrNextPlayerInvalid = errors.New("next player is invalid")
)
