package dicegame

import (
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
	valid := false

	// Track which dice have been used in combinations
	usedDice := make(map[int]bool)

	// Count occurrences of each number
	counts := make(map[int]int)
	for _, die := range dice {
		counts[die]++
	}
	log.Debug().Interface("counts", counts).Msg("Dice counts")

	// First check for three of a kind and beyond
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
			// Mark all dice in the three of a kind as used
			for i := 0; i < count; i++ {
				usedDice[num] = true
			}
			log.Debug().Int("num", num).Int("count", count).Int("baseScore", baseScore).Msg("Found three or more of a kind")
		}
	}

	// Then check for runs (only if we haven't used the dice in three of a kind)
	if len(dice) >= 5 {
		// Check for 1-5 run
		if containsRun(dice, 1, 5) {
			score += 500
			valid = true
			// Mark all dice in the run as used
			for i := 1; i <= 5; i++ {
				usedDice[i] = true
			}
			log.Debug().Msg("Found 1-5 run: +500")
		}
		// Check for 2-6 run
		if containsRun(dice, 2, 6) {
			score += 750
			valid = true
			// Mark all dice in the run as used
			for i := 2; i <= 6; i++ {
				usedDice[i] = true
			}
			log.Debug().Msg("Found 2-6 run: +750")
		}
		// Check for 1-6 run
		if containsRun(dice, 1, 6) {
			score += 1500
			valid = true
			// Mark all dice in the run as used
			for i := 1; i <= 6; i++ {
				usedDice[i] = true
			}
			log.Debug().Msg("Found 1-6 run: +1500")
		}
	}

	// Finally check for individual 1s and 5s (only if not used in combinations)
	for num, count := range counts {
		if !usedDice[num] {
			if num == 1 {
				score += count * 100
				valid = true
				log.Debug().Int("count", count).Msg("Found ones: +100 each")
			} else if num == 5 {
				score += count * 50
				valid = true
				log.Debug().Int("count", count).Msg("Found fives: +50 each")
			}
		}
	}

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

func (g *DiceGame) EndTurn(state *GameState) {
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
	_, valid := g.CalculateScore(state.Dice)
	if !valid {
		g.EndTurn(state)
	}

	room.SetState(state)
}

func (g *DiceGame) handleSelect(room interfaces.Room, payload SelectActionPayload) {
	log.Debug().Str("room", room.ID()).Any("diceIndex", payload.DiceIndex).Msg("selecting dice")

	state := room.State().(*GameState)

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
		log.Error().Msg("No valid dice indices to select")
		return
	}

	log.Debug().Str("room", room.ID()).Any("dice", selectedDice).Msg("selected dice")

	score, _ := g.CalculateScore(selectedDice)

	// we update the state's selected dice even if invalid, maybe stupid
	state.SelectedDice = tempSelected
	if player, exists := state.Players[state.CurrentTurn]; exists {
		player.TurnScore = score
	}

	room.SetState(state)
}

func (g *DiceGame) handleSetAside(room interfaces.Room, payload SetAsideActionPayload) {
	log.Debug().Str("room", room.ID()).Msg("setting dice aside")

	// Handle case where DiceIndex is empty
	if len(payload.DiceIndex) == 0 {
		log.Error().Msg("handleSelect called with empty DiceIndex, ignoring selection")
		return
	}

	state := room.State().(*GameState)
	// TODO: reorder, check score first and if valid allow setting aside
	if g.SetAsideDice(payload.DiceIndex, state) {
		score, valid := g.CalculateScore(state.SetAside)
		if valid {
			if player, exists := state.Players[state.CurrentTurn]; exists {
				player.RoundScore = score
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
