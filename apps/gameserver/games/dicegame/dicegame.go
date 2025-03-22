package dicegame

import (
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"maps"
	"math/rand"
	"slices"
	"sort"

	"github.com/rs/zerolog/log"
)

const (
	TargetScore = 3000
)

// DiceGame implements the game interface
type DiceGame struct{}

type Player struct {
	ID    string
	Score int
}

type GameState struct {
	Players     map[string]*Player
	CurrentTurn string
	Winner      string
	Dice        []int
	SetAside    []int
	TurnScore   int
	RoundScore  int
}

type ActionPayload struct {
	PlayerID  string `json:"playerId"`
	DiceIndex []int  `json:"diceIndex,omitempty"`
}

func (g *DiceGame) AddPlayer(id string, state *GameState) {
	state.Players[id] = &Player{
		ID:    id,
		Score: 0,
	}
}

func (g *DiceGame) RollDice(state *GameState) {
	for i := range state.Dice {
		state.Dice[i] = rand.Intn(6) + 1
	}
}

func (g *DiceGame) SetAsideDice(indices []int, state *GameState) bool {
	// Validate indices
	for _, idx := range indices {
		if idx < 0 || idx >= len(state.Dice) {
			return false
		}
	}

	// Move selected dice to setAside
	for _, idx := range indices {
		state.SetAside = append(state.SetAside, state.Dice[idx])
	}

	// Remove selected dice from main dice pool
	newDice := make([]int, 0)
	for i, die := range state.Dice {
		selected := false
		for _, idx := range indices {
			if i == idx {
				selected = true
				break
			}
		}
		if !selected {
			newDice = append(newDice, die)
		}
	}
	state.Dice = newDice

	return true
}

func (g *DiceGame) CalculateScore(dice []int) (int, bool) {
	if len(dice) == 0 {
		return 0, false
	}

	// Sort dice for easier combination checking
	sort.Ints(dice)

	score := 0
	valid := false

	// Check for runs
	if len(dice) >= 5 {
		// Check for 1-5 run
		if containsRun(dice, 1, 5) {
			score += 500
			valid = true
		}
		// Check for 2-6 run
		if containsRun(dice, 2, 6) {
			score += 750
			valid = true
		}
		// Check for 1-6 run
		if containsRun(dice, 1, 6) {
			score += 1500
			valid = true
		}
	}

	// Count occurrences of each number
	counts := make(map[int]int)
	for _, die := range dice {
		counts[die]++
	}

	// Check for three of a kind and beyond
	for num, count := range counts {
		if count >= 3 {
			baseScore := num * 100
			if num == 1 {
				baseScore = 1000
			}
			// Double score for each additional die beyond three
			for i := 3; i < count; i++ {
				baseScore *= 2
			}
			score += baseScore
			valid = true
		}
	}

	// Check for individual 1s and 5s
	for num, count := range counts {
		if num == 1 {
			score += count * 100
			valid = true
		} else if num == 5 {
			score += count * 50
			valid = true
		}
	}

	return score, valid
}

func containsRun(dice []int, start, end int) bool {
	if len(dice) < end-start+1 {
		return false
	}

	// Create a map to track found numbers
	found := make(map[int]bool)
	for _, die := range dice {
		if die >= start && die <= end {
			found[die] = true
		}
	}

	// Check if all numbers in the range are present
	for i := start; i <= end; i++ {
		if !found[i] {
			return false
		}
	}
	return true
}

func (g *DiceGame) EndTurn(state *GameState) {
	// Add turn score to player's total score
	if player, exists := state.Players[state.CurrentTurn]; exists {
		player.Score += state.TurnScore
	}

	// Reset turn-specific variables
	state.TurnScore = 0
	state.SetAside = make([]int, 0)
	state.Dice = make([]int, 6)

	// Switch to next player
	players := slices.Collect(maps.Values(state.Players))
	for idx, player := range players {
		if player.ID == state.CurrentTurn {
			nextIndex := (idx + 1) % len(state.Players)
			newPlayerId := players[nextIndex].ID

			log.Info().Msgf("Switching turn from %s to %s", state.CurrentTurn, newPlayerId)
			state.CurrentTurn = newPlayerId
			break
		}
	}

}

func (g *DiceGame) handleRoll(room interfaces.Room) {
	state := room.State().(*GameState)

	if len(state.Dice) == 0 {
		state.Dice = make([]int, 6)
	}
	g.RollDice(state)
	_, valid := g.CalculateScore(state.Dice)
	if !valid {
		state.TurnScore = 0
		g.EndTurn(state)
	}

	room.SetState(state)
}

func (g *DiceGame) handleSelect(room interfaces.Room, payload ActionPayload) {
	state := room.State().(*GameState)
	selectedDice := make([]int, 0)
	for _, idx := range payload.DiceIndex {
		if idx >= 0 && idx < len(state.Dice) {
			selectedDice = append(selectedDice, state.Dice[idx])
		}
	}
	score, valid := g.CalculateScore(append(state.SetAside, selectedDice...))
	if valid {
		room.Broadcast(protocol.NewSuccessResponse("temp_score", interfaces.M{
			"score": score,
		}))
	}

	room.SetState(state)
}

func (g *DiceGame) handleSetAside(room interfaces.Room, payload ActionPayload) {
	state := room.State().(*GameState)
	if g.SetAsideDice(payload.DiceIndex, state) {
		score, valid := g.CalculateScore(state.SetAside)
		if valid {
			state.TurnScore = score
		}
	}
	room.SetState(state)
}

func (g *DiceGame) handleEndTurn(room interfaces.Room) {
	state := room.State().(*GameState)
	g.EndTurn(state)
	room.SetState(state)
}
