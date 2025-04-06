package dicegame

import (
	"errors"
	"gameserver/internal/interfaces"
	"maps"
	"math/rand"
	"slices"
	"sort"

	"github.com/rs/zerolog/log"
)

// DiceGame implements the game interface
type DiceGame struct{}

type Player struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Score      int    `json:"score"`
	TurnScore  int    `json:"turnScore"`
	RoundScore int    `json:"roundScore"`
}

type GameState struct {
	Players      map[string]*Player `json:"players"`
	Started      bool               `json:"started"`
	CurrentTurn  string             `json:"currentTurn"`
	Winner       string             `json:"winner"`
	Dice         []int              `json:"dice"`
	SelectedDice []int              `json:"selectedDice"`
	SetAside     []int              `json:"setAside"`
	TargetScore  int                `json:"targetScore"`
}

type SelectActionPayload struct {
	DiceIndex int `json:"diceIndex,omitempty"`
}

type SetAsideActionPayload struct {
	DiceIndex []int `json:"diceIndex,omitempty"`
}

func (g *DiceGame) AddPlayer(id string, name string, state *GameState) {
	state.Players[id] = &Player{
		ID:    id,
		Name:  name,
		Score: 0,
	}
}

func (g *DiceGame) RollDice(state *GameState) {
	for i := range state.Dice {
		state.Dice[i] = rand.Intn(6) + 1
	}
}

func (g *DiceGame) SetAsideDice(indices []int, state *GameState) bool {
	// Handle empty indices
	if len(indices) == 0 {
		log.Error().Msg("SetAsideDice called with empty indices, no action taken")
		return false
	}

	// Validate indices
	for _, idx := range indices {
		if idx < 0 || idx >= len(state.Dice) {
			log.Error().Int("index", idx).Int("dice_length", len(state.Dice)).Msg("Invalid dice index")
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
	log.Debug().Ints("dice", dice).Msg("Calculating score for dice")

	score := 0
	usedDiceCount := 0

	// Make a copy of dice that we can modify
	remainingDice := make([]int, len(dice))
	copy(remainingDice, dice)

	// Check for runs first
	if len(remainingDice) >= 6 && containsRun(remainingDice, 1, 6) {
		score = 1500
		usedDiceCount += 6
		remainingDice = removeRun(remainingDice, 1, 6)
		log.Debug().Msg("Found 1-6 run: +1500")
	} else if len(remainingDice) >= 5 && containsRun(remainingDice, 1, 5) {
		score += 500
		usedDiceCount += 5
		remainingDice = removeRun(remainingDice, 1, 5)
		log.Debug().Msg("Found 1-5 run: +500")
	} else if len(remainingDice) >= 5 && containsRun(remainingDice, 2, 6) {
		score += 750
		usedDiceCount += 5
		remainingDice = removeRun(remainingDice, 2, 6)
		log.Debug().Msg("Found 2-6 run: +750")
	}

	// Count occurrences for remaining dice
	counts := make(map[int]int)
	for _, die := range remainingDice {
		counts[die]++
	}
	log.Debug().Interface("counts", counts).Msg("Dice counts")

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
			usedDiceCount += count
			// Remove these dice from further consideration
			counts[num] = 0
			log.Debug().Int("num", num).Int("count", count).Int("baseScore", baseScore).Msg("Found three or more of a kind")
		}
	}

	// Check for individual 1s and 5s from remaining dice
	if counts[1] > 0 {
		score += counts[1] * 100
		usedDiceCount += counts[1]
		log.Debug().Int("count", counts[1]).Msg("Found ones: +100 each")
	}

	if counts[5] > 0 {
		score += counts[5] * 50
		usedDiceCount += counts[5]
		log.Debug().Int("count", counts[5]).Msg("Found fives: +50 each")
	}

	// Check if all dice are used in valid combinations
	valid := usedDiceCount == len(dice)

	log.Debug().Int("final_score", score).Bool("valid", valid).Msg("Final score calculation")
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

// Helper function to remove run dice from the slice
func removeRun(dice []int, start, end int) []int {
	result := make([]int, 0)
	runDice := make(map[int]bool)

	for i := start; i <= end; i++ {
		runDice[i] = true
	}

	// Add one occurrence of each number in the run
	usedRun := make(map[int]bool)

	for _, die := range dice {
		if runDice[die] && !usedRun[die] {
			usedRun[die] = true
			continue
		}
		result = append(result, die)
	}

	return result
}

func (g *DiceGame) EndTurn(state *GameState) {
	log.Info().Msg("Ending turn")
	// Add turn score to player's total score
	if player, exists := state.Players[state.CurrentTurn]; exists {
		player.Score += player.RoundScore
		player.TurnScore = 0
		player.RoundScore = 0
	}

	// Reset turn-specific variables
	state.SetAside = make([]int, 0)
	state.Dice = make([]int, 6)
	state.SelectedDice = make([]int, 0)

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
	log.Debug().Str("room", room.ID()).Msg("rolling dice")

	state := room.State().(*GameState)
	// reset state
	state.SelectedDice = make([]int, 0)

	if len(state.Dice) == 0 {
		state.Dice = make([]int, 6)
	}
	g.RollDice(state)
	score, valid := g.CalculateScore(state.Dice)
	// the first roll can be invalid but still be scoreable
	if score == 0 && !valid {
		g.EndTurn(state)
	}

	room.SetState(state)
}

func (g *DiceGame) handleSelect(room interfaces.Room, payload SelectActionPayload) error {
	log.Debug().Str("room", room.ID()).Any("diceIndex", payload.DiceIndex).Msg("selecting dice")

	state := room.State().(*GameState)

	// validate that the payload is in bounds
	for sD := range payload.DiceIndex {
		if sD >= len(state.Dice) {
			return errors.New("Invalid dice selection")
		}
	}
	// Create a temporary selection to test if it's valid
	tempSelected := make([]int, len(state.SelectedDice))
	copy(tempSelected, state.SelectedDice)

	// if we already have that dice selected, remove it
	if slices.Contains(tempSelected, payload.DiceIndex) {
		newSelected := make([]int, 0, len(tempSelected)-1)
		for _, idx := range tempSelected {
			if idx != payload.DiceIndex {
				newSelected = append(newSelected, idx)
			}
		}
		tempSelected = newSelected
	} else {
		tempSelected = append(tempSelected, payload.DiceIndex)
	}

	/**
	** dice: [1,2,2,5,6,6]
	** selectedIdx: [1,4] --> 2,6
	** selected dice: [2,6]
	**/

	selectedDice := make([]int, 0)
	// populate dice from temporary selection
	for _, idx := range tempSelected {
		selectedDice = append(selectedDice, state.Dice[idx])
	}

	// Only proceed if there are valid selections
	if len(selectedDice) == 0 {
		return errors.New("No valid dice indices to select")
	}

	log.Debug().Str("room", room.ID()).Any("dice", selectedDice).Msg("selected dice")

	score, _ := g.CalculateScore(selectedDice)

	// we update the state's selected dice even if invalid, maybe stupid
	state.SelectedDice = tempSelected
	if player, exists := state.Players[state.CurrentTurn]; exists {
		player.TurnScore = score
	}

	room.SetState(state)
	return nil
}

func (g *DiceGame) handleSetAside(room interfaces.Room) {
	log.Debug().Str("room", room.ID()).Msg("setting dice aside")

	state := room.State().(*GameState)
	selectedDice := state.SelectedDice
	// TODO: reorder, check score first and if valid allow setting aside
	if g.SetAsideDice(selectedDice, state) {
		score, valid := g.CalculateScore(state.SetAside)
		if valid {
			if player, exists := state.Players[state.CurrentTurn]; exists {
				player.TurnScore = 0
				player.RoundScore = score
				state.SelectedDice = make([]int, 0)
			}
		}

		// auto-reroll when all dice were successfully removed
		if len(state.Dice) == 0 {
			g.RollDice(state)
		}
	}
	room.SetState(state)
}

func (g *DiceGame) handleEndTurn(room interfaces.Room) {
	log.Debug().Str("room", room.ID()).Msg("ending turn")

	state := room.State().(*GameState)
	g.EndTurn(state)
	room.SetState(state)
}
