package owe_drahn

import (
	"encoding/json"
	"errors"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"github.com/rs/zerolog/log"
	"math/rand"
	"sort"
	"time"
)

type Game struct{}

type Roll struct {
	Player interfaces.M `json:"player"`
	Dice   int          `json:"dice"`
	Total  int          `json:"total"`
}

type GameState struct {
	Players     map[string]*Player `json:"players"`
	Started     bool               `json:"started"`
	CurrentTurn string             `json:"currentTurn"` // TODO: this prop is new, adapt frontend
	Over        bool               `json:"over"`

	Rolls        []Roll    `json:"rolls"`
	CurrentValue int       `json:"currentValue"`
	StartedAt    time.Time `json:"startedAt"`
	FinishedAt   time.Time `json:"finishedAt"`
}

// MovePayload represents a move action from a client
type ReadyPayload struct {
	Ready bool `json:"ready"`
}

func (g *Game) AddPlayer(id string, name string, state *GameState) {
	state.Players[id] = NewPlayer(id, name)
}

// TODO: Needed?
func (g *Game) GetPlayer(id string, state *GameState) *Player {
	return state.Players[id]
}

func (g *Game) RemovePlayer(id string, state *GameState) {
	delete(state.Players, id)
	// TODO: broadcast `playerLeft` to all players
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

func (g *Game) Init(state *GameState) {
	state.Started = false
	state.Over = false
	state.CurrentValue = 0
	state.Rolls = make([]Roll, 0)
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
}

func (g *Game) RollDice(playerId string, state *GameState) error {
	player := g.GetPlayer(playerId, state)
	if player == nil {
		return errors.New("player not found")
	}

	if player.ID != state.CurrentTurn {
		return errors.New("not your turn")
	}

	dice := random(0, 6)
	// Rule of 3, doesn't count
	if dice != 3 {
		state.CurrentValue += dice
	}
	state.Rolls = append(state.Rolls, Roll{
		Player: player.ToFormattedPlayer(),
		Dice:   dice,
		Total:  state.CurrentValue,
	})

	// check player death
	if state.CurrentValue > 15 {
		player.Life = 0
		state.CurrentValue = 0
	}

	// TODO: send `rolledDice` to all players, {dice, player, total}

	if player.IsChoosing {
		player.IsChoosing = false
		log.Error().Msg(" How the fuck?! Player is choosing, but should not be.")
	}

	return g.setNextPlayer(state)
}

/**
 * When a Player "draht owe", he can choose who starts next.
 * Only let the player choose next if:
 *  1. it's his turn
 *  2. He is choosing. Is set after he "drahs owe"
 *  3. Chosen Player is still alive. (prevent choosing of already lost players)
 */
func (g *Game) chooseNextPlayer(playerId string, nextPlayerId string, state *GameState) {
	currentPlayer := g.GetCurrentPlayer(state)
	nextPlayer := state.Players[nextPlayerId]
	if nextPlayer == nil {
		// TODO: send NO_PLAYER game error
		//this.sendGameError({
		//code: GameErrorCode.NO_PLAYER,
		//	message: "You are not part of this game!"
		//});
		return
	}

	if currentPlayer.ID == playerId && currentPlayer.IsChoosing && nextPlayer.Life > 0 {
		currentPlayer.IsPlayersTurn = false
		nextPlayer.IsPlayersTurn = true
		currentPlayer.IsChoosing = false
		state.CurrentTurn = nextPlayerId
	}

	// TODO: send this.sendPlayerUpdate(true);
}

/**
 * The next-player algorithm.
 * Always chooses the next player in the array order. If last, start at first.
 *
 * Determines if the game is over, when no players are left alive.
 */
func (g *Game) setNextPlayer(state *GameState) error {
	playerIDs := g.getSortedPlayerIDs(state)
	if len(playerIDs) == 0 {
		return errors.New("no players while trying to set next player")
	}

	// start of the game, nobodys turn
	// If no current turn is set, start with the first player
	if state.CurrentTurn == "" {
		state.CurrentTurn = playerIDs[0]
		g.GetCurrentPlayer(state).IsPlayersTurn = true
		return nil
	}

	// Find the current player's position in our ordered list
	currentIndex := -1
	for i, id := range playerIDs {
		if id == state.CurrentTurn {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return errors.New("could not find current player")
	}

	// unset current players turn
	g.GetCurrentPlayer(state).IsPlayersTurn = false

	// Find the next alive player
	alivePlayers := g.GetPlayersAlive(state)

	if len(alivePlayers) <= 1 {
		winner := alivePlayers[0]
		g.gameOver(winner.Name, state)
	} else {
		// Find the next player who is still alive
		nextIndex := currentIndex
		for {
			nextIndex = (nextIndex + 1) % len(playerIDs)
			nextPlayerID := playerIDs[nextIndex]

			// Check if the player is alive
			if state.Players[nextPlayerID].Life > 0 {
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
		// TODO: send this.sendPlayerUpdate()
	}

	return nil
}

// setNextPlayerRandom selects a random player to be the next player.
// Should only be used at the start when we have no other current player yet.
func (g *Game) setNextPlayerRandom(state *GameState) {
	playerIDs := g.getSortedPlayerIDs(state)
	randomIdx := random(0, len(playerIDs)-1)
	state.CurrentTurn = playerIDs[randomIdx]
	g.GetCurrentPlayer(state).IsPlayersTurn = true

	// TODO: send   this.sendPlayerUpdate();
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

func (g *Game) gameOver(winner string, state *GameState) {
	state.Over = true
	state.FinishedAt = time.Now()
	// TODO: send this.sendGameOver(winner);
	// restart after 5s
	time.AfterFunc(5*time.Second, func() {
		g.Init(state)
		// TODO: send this.sendGameInit();
	})
}

// SetStatsOnPlayer connects the player and sets the stats.
func (g *Game) SetStatsOnPlayer(clientId string, userId string, stats interface{}, state *GameState) error {
	player := g.GetPlayer(clientId, state)
	if player == nil {
		return errors.New("player not found")
	}

	player.IsConnected = true

	if userId != "" {
		log.Info().Str("clientId", clientId).Str("userId", userId).Msg("Setting registered user data")
		player.UserID = userId
		if stats != nil {
			player.Stats = stats
		}
	}
	return nil
}

func (g *Game) handleReady(client interfaces.Client, room interfaces.Room, payload []byte) {
	var ready ReadyPayload
	if err := json.Unmarshal(payload, &ready); err != nil {
		client.Send(protocol.NewErrorResponse("error", "Invalid ready format"))
		return
	}

	state := room.State().(*GameState)
	log.Debug().Str("clientID", client.ID()).Bool("ready", ready.Ready).Msg("player sends ready")

	player := g.GetPlayer(client.ID(), state)
	if player == nil {
		client.Send(protocol.NewErrorResponse("error", "player not found"))
		return
	}

	player.IsReady = ready.Ready

	if g.IsEveryoneReady(state) {
		g.start(state)

		// TODO: send `gameStarted`
		g.setNextPlayerRandom(state)
		// reset everyones ready state for UI purposes
		for _, p := range state.Players {
			p.IsReady = false
		}
	}

	// TODO: send  this.sendPlayerUpdate(true);
}

func (g *Game) handleLoseLife(client interfaces.Client, room interfaces.Room) {
	state := room.State().(*GameState)
	log.Debug().Str("clientID", client.ID()).Msg("player sends rollDice")

	player := g.GetPlayer(client.ID(), state)
	if player == nil {
		client.Send(protocol.NewErrorResponse("error", "player not found"))
		return
	}

	player.Life -= 1
	player.IsChoosing = true
	state.CurrentValue = 0

	// TODO: send `lostLife` {player}
	// TODO sendGameUpdate();
}

// random returns a random number between min and max (inclusive)
func random(min int, max int) int {
	return rand.Intn(max-min+1) + min
}
