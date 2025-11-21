package owe_drahn

import (
	"context"
	"encoding/json"
	"errors"
	"gameserver/games/owe_drahn/database"
	"gameserver/games/owe_drahn/models"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"github.com/rs/zerolog/log"
	"math/rand"
	"sort"
	"time"
)

type Game struct {
	dbService database.Database
}

type GameState struct {
	Ctx         context.Context
	Players     map[string]*Player `json:"players"`
	PlayerOrder []string           `json:"playerOrder"`
	Started     bool               `json:"started"`
	CurrentTurn string             `json:"currentTurn"`
	Over        bool               `json:"over"`

	Rolls        []models.Roll `json:"rolls"`
	CurrentValue int           `json:"currentValue"`
	StartedAt    time.Time     `json:"startedAt"`
	FinishedAt   time.Time     `json:"finishedAt"`
}

func (s *GameState) ToMap() interfaces.M {
	return interfaces.M{
		"players":      mapPlayersToArray(s.Players, s.PlayerOrder),
		"started":      s.Started,
		"over":         s.Over,
		"currentValue": s.CurrentValue,
		"currentTurn":  s.CurrentTurn,
	}
}

func (s *GameState) ToDBGame() models.DBGame {
	return models.DBGame{
		Players:    mapPlayersToFormattedPlayers(mapPlayersToArray(s.Players, s.PlayerOrder)),
		StartedAt:  s.StartedAt,
		FinishedAt: s.FinishedAt,
		Rolls:      s.Rolls,
	}
}

type HandshakePayload struct {
	UserID string `json:"uid"`
}

type NextPlayerPayload struct {
	NextPlayerId string `json:"nextPlayerId"`
}

func (g *Game) AddPlayer(id string, name string, state *GameState) {
	state.Players[id] = NewPlayer(id, name)
	state.PlayerOrder = append(state.PlayerOrder, id)
}

func (g *Game) GetPlayer(id string, state *GameState) *Player {
	return state.Players[id]
}

func (g *Game) RemovePlayer(clientId string, room interfaces.Room) {
	state := room.State().(*GameState)
	playerName := state.Players[clientId].Name
	delete(state.Players, clientId)

	// Remove from player order
	for i, id := range state.PlayerOrder {
		if id == clientId {
			state.PlayerOrder = append(state.PlayerOrder[:i], state.PlayerOrder[i+1:]...)
			break
		}
	}

	g.broadcastGameEvent(room, "playerLeft", interfaces.M{
		"username": playerName,
	})
}

// GetPlayersAlive returns all players that are still alive
func (g *Game) GetPlayersAlive(state *GameState) []*Player {
	alivePlayers := make([]*Player, 0)
	for _, player := range state.Players {
		if player.Life > 0 {
			alivePlayers = append(alivePlayers, player)
		}
	}
	return alivePlayers
}

// GetPlayersRegistered returns all players that are registered (Google login)
func (g *Game) GetPlayersRegistered(state *GameState) []*Player {
	registeredPlayers := make([]*Player, 0)
	for _, player := range state.Players {
		if player.UserID != "" {
			registeredPlayers = append(registeredPlayers, player)
		}
	}
	return registeredPlayers
}

func (g *Game) GetCurrentPlayer(state *GameState) *Player {
	return state.Players[state.CurrentTurn]
}

func (g *Game) IsPlayersTurn(id string, state *GameState) bool {
	return state.CurrentTurn == id
}

func (g *Game) IsPlayerConnected(id string, state *GameState) bool {
	return state.Players[id].IsConnected
}

func (g *Game) IsEveryoneReady(state *GameState) bool {
	for _, player := range state.Players {
		if !player.IsReady {
			return false
		}
	}
	return true
}

// IsPlayer checks if the given client id is an actual player
func (g *Game) IsPlayer(id string, state *GameState) bool {
	for _, player := range state.Players {
		if player.ID == id {
			return true
		}
	}
	return false
}

func (g *Game) HasPlayers(state *GameState) bool {
	return len(state.Players) > 0
}

func (g *Game) Reset(state *GameState) {
	log.Info().Msg("resetting game")

	state.Started = false
	state.Over = false
	state.CurrentValue = 0
	state.CurrentTurn = ""
	for _, player := range state.Players {
		player.Reset()
	}
	state.Rolls = make([]models.Roll, 0)
}

func (g *Game) start(state *GameState) {
	state.Started = true
	state.StartedAt = time.Now()

	// Randomly select the starting player
	playerIDs := make([]string, 0, len(state.Players))
	for id := range state.Players {
		playerIDs = append(playerIDs, id)
	}
	state.CurrentTurn = playerIDs[rand.Intn(len(playerIDs))]

	log.Debug().Str("currentTurn", state.CurrentTurn).Msg("starting game")
}

func (g *Game) handleRoll(room interfaces.Room) error {
	state := room.State().(*GameState)
	player := g.GetCurrentPlayer(state)

	dice := random(1, 6)
	// Rule of 3, doesn't count
	if dice != 3 {
		state.CurrentValue += dice
	}
	state.Rolls = append(state.Rolls, models.Roll{
		Player: player.ToFormattedPlayer(),
		Dice:   dice,
		Total:  state.CurrentValue,
	})

	// check player death
	total := state.CurrentValue
	if total > 15 {
		player.Life = 0
		state.CurrentValue = 0
		player.Score -= 1
	}

	if player.IsChoosing {
		player.IsChoosing = false
		log.Error().Msg(" How the fuck?! Player is choosing, but should not be.")
	}

	room.SetState(state)
	g.broadcastGameEvent(room, "rolledDice", interfaces.M{
		"dice":   dice,
		"player": player.ToFormattedPlayer(),
		"total":  total,
	})

	return g.setNextPlayer(room, state)
}

/**
 * When a Player "draht owe", he can choose who starts next.
 * Only let the player choose next if:
 *  1. it's his turn
 *  2. He is choosing. Is set after he "drahs owe"
 *  3. Chosen Player is still alive. (prevent choosing of already lost players)
 */
func (g *Game) handleChooseNextPlayer(client interfaces.Client, state *GameState, payload []byte) error {
	var nextPlayerData NextPlayerPayload
	if err := json.Unmarshal(payload, &nextPlayerData); err != nil {
		return errors.New("invalid nextPlayer format")
	}

	currentPlayer := g.GetCurrentPlayer(state)
	nextPlayer := state.Players[nextPlayerData.NextPlayerId]
	if nextPlayer == nil {
		return errors.New("next player is invalid")
	}

	if currentPlayer.IsChoosing && nextPlayer.Life > 0 {
		log.Info().Str("currentTurn", state.CurrentTurn).Str("nextPlayerId", nextPlayerData.NextPlayerId).Msg("choosing next player")
		currentPlayer.IsChoosing = false
		state.CurrentTurn = nextPlayerData.NextPlayerId
	}

	//g.broadcastPlayerUpdate(client.Room(), state.Players, state.CurrentTurn, true)
	//g.broadcastGameState(client.Room())
	return nil
}

/**
 * The next-player algorithm.
 * Always chooses the next player in the array order. If last, start at first.
 *
 * Determines if the game is over, when no players are left alive.
 */
func (g *Game) setNextPlayer(room interfaces.Room, state *GameState) error {
	if len(state.PlayerOrder) == 0 {
		return errors.New("no players while trying to set next player")
	}

	// start of the game, nobodys turn
	// If no current turn is set, start with the first player
	if state.CurrentTurn == "" {
		state.CurrentTurn = state.PlayerOrder[0]
		return nil
	}

	// Find the current player's position in our ordered list
	currentIndex := -1
	for i, id := range state.PlayerOrder {
		if id == state.CurrentTurn {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return errors.New("could not find current player")
	}

	// Find the next alive player
	alivePlayers := g.GetPlayersAlive(state)

	if len(alivePlayers) <= 1 {
		winner := alivePlayers[0]
		winner.Score += len(state.Players) - 1 // add winnings to the winner, -1 for their own bet
		g.gameOver(room, winner.Name, state)
	} else {
		// Find the next player who is still alive
		nextIndex := currentIndex
		for {
			nextIndex = (nextIndex + 1) % len(state.PlayerOrder)
			nextPlayerID := state.PlayerOrder[nextIndex]

			// Check if the player is alive
			if state.Players[nextPlayerID].Life > 0 {
				log.Info().Str("currentTurn", state.CurrentTurn).Str("nextPlayerID", nextPlayerID).Msg("setting next player")
				state.CurrentTurn = nextPlayerID
				break
			}

			// Failsafe - if we've checked all players and come back to the starting point
			if nextIndex == currentIndex {
				// This shouldn't happen if there are alive players
				return errors.New("Critical! Could not find next player")
			}
		}
	}

	if !state.Over {
		//g.broadcastPlayerUpdate(room, state.Players, state.CurrentTurn, false)
		//g.broadcastGameState(room)
	}

	return nil
}

// setNextPlayerRandom selects a random player to be the next player.
// Should only be used at the start when we have no other current player yet.
func (g *Game) setNextPlayerRandom(room interfaces.Room, state *GameState) {
	if len(state.PlayerOrder) == 0 {
		log.Error().Msg("no players while trying to set next random player")
		return
	}
	randomIdx := random(0, len(state.PlayerOrder)-1)
	state.CurrentTurn = state.PlayerOrder[randomIdx]

	//g.broadcastPlayerUpdate(room, state.Players, state.CurrentTurn, false)
}

func (g *Game) getSortedPlayerIDs(state *GameState) []string {
	// Create an ordered slice of player IDs
	playerIDs := make([]string, 0, len(state.Players))
	for id := range state.Players {
		playerIDs = append(playerIDs, id)
	}

	// Sort the player IDs to ensure a consistent order
	sort.Strings(playerIDs)

	return playerIDs
}

func (g *Game) gameOver(room interfaces.Room, winner string, state *GameState) {
	log.Info().Str("winner", winner).Msg("game over")
	state.Over = true
	state.FinishedAt = time.Now()

	g.broadcastGameEvent(room, "gameOver", interfaces.M{
		"winner": winner,
	})

	g.dbService.StoreGame(state.Ctx, state.ToDBGame())
	// restart after 5s
	time.AfterFunc(5*time.Second, func() {
		g.Reset(state)
		g.broadcastGameEvent(room, "gameInit", state.ToMap())
	})
}

// SetStatsOnPlayer connects the player and sets the stats.
func (g *Game) SetStatsOnPlayer(clientId string, userId string, stats *models.PlayerStats, state *GameState) {
	log.Info().Str("clientId", clientId).Str("userId", userId).Msg("setting registered user data")

	player := g.GetPlayer(clientId, state)
	player.UserID = userId
	player.Stats = stats
}

func (g *Game) handleReady(client interfaces.Client, state *GameState, payload []byte) {
	var ready bool
	if err := json.Unmarshal(payload, &ready); err != nil {
		client.Send(protocol.NewErrorResponse("error", "Invalid ready format"))
		return
	}

	log.Debug().Str("clientID", client.ID()).Bool("ready", ready).Msg("player sends ready")

	player := state.Players[client.ID()]
	player.IsReady = ready

	if g.IsEveryoneReady(state) {
		g.start(state)

		g.broadcastGameEvent(client.Room(), "gameStarted", nil)
		g.setNextPlayerRandom(client.Room(), state)
		// reset everyones ready state for UI purposes
		for _, p := range state.Players {
			p.IsReady = false
		}
	}

	//g.broadcastPlayerUpdate(client.Room(), state.Players, state.CurrentTurn, true)
}

func (g *Game) handleLoseLife(client interfaces.Client, state *GameState) {
	log.Debug().Str("clientID", client.ID()).Msg("player loses life")

	player := g.GetCurrentPlayer(state)
	player.Life -= 1
	player.IsChoosing = true
	state.CurrentValue = 0

	g.broadcastGameEvent(client.Room(), "lostLife", interfaces.M{
		"player": player.ToFormattedPlayer(),
	})
}

//	handleHandshake
//
// When a client loads the game page, they send a handshake event.
// We connect the Client back to the Player if it was one.
func (g *Game) handleHandshake(client interfaces.Client, state *GameState, payload []byte) {
	p := g.GetPlayer(client.ID(), state)
	if p == nil {
		return
	}

	var handshake HandshakePayload
	if err := json.Unmarshal(payload, &handshake); err != nil {
		client.Send(protocol.NewErrorResponse("error", "Invalid handshake format"))
		return
	}
	log.Debug().Str("clientId", client.ID()).Str("userId", handshake.UserID).Msg("handshake")
	if handshake.UserID != "" {
		if userStats, err := g.dbService.GetUserStats(state.Ctx, handshake.UserID); err == nil {
			g.SetStatsOnPlayer(client.ID(), handshake.UserID, userStats, state)
		} else {
			log.Error().Err(err).Msg("error getting user stats")
		}
	}
	p.IsConnected = true
}

// random returns a random number between min and max (inclusive)
func random(min int, max int) int {
	return rand.Intn(max-min+1) + min
}

func mapPlayersToArray(players map[string]*Player, playerOrder []string) []*Player {
	result := make([]*Player, 0, len(players))
	for _, id := range playerOrder {
		if player, exists := players[id]; exists {
			result = append(result, player)
		} else {
			log.Error().Str("id", id).Strs("order", playerOrder).Msg("player not found in player order")
		}
	}
	return result
}

func mapPlayersToFormattedPlayers(players []*Player) []*models.FormattedPlayer {
	dbPlayers := make([]*models.FormattedPlayer, 0, len(players))
	for _, player := range players {
		dbPlayers = append(dbPlayers, player.ToFormattedPlayer())
	}

	return dbPlayers
}
