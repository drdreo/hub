package dicegame

import (
	"encoding/json"
	"errors"
	"gameserver/internal/client"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
)

const BOT_DELAY = 2000 // 2 second delay for bot actions

type DiceGameBot struct {
	*client.BotClient
	game                *DiceGame
	myTurn              bool
	busted              bool
	combinationToSelect []int
}

func NewDiceGameBot(id string, game *DiceGame, reg interfaces.GameRegistry) *DiceGameBot {
	bot := &DiceGameBot{
		BotClient:           client.NewBotClient(id, reg),
		game:                game,
		myTurn:              false,
		busted:              false,
		combinationToSelect: make([]int, 0),
	}
	bot.SetMessageHandler(bot.handleMessage)
	return bot
}

func (b *DiceGameBot) handleMessage(message *protocol.Response) {
	// First check if we should still process this message
	if b.Context().Err() != nil {
		log.Debug().Str("botId", b.ID()).Msg("Context canceled, ignoring message")
		return
	}

	switch message.Type {
	case "game_state":
		gameState, ok := b.getGameState(message)
		if !ok || !gameState.Started {
			return
		}

		b.checkBotTurn(gameState)

		if !b.myTurn {
			// do nothing
			log.Debug().Msg("not my turn, chilling")
			return
		}
		// Don't make moves if we're busted
		if b.busted {
			log.Debug().Msg("i am busted, chilling")
			return
		}

		b.makeNextMove(gameState)

	case "busted":
		data, _ := message.Data.(*BustedResponse)

		// Check for bust and if it includes the bots ID
		if b.myTurn && data.ClientID == b.ID() {
			log.Warn().Msg("Bot detected bust and will wait for turn to end")
			b.busted = true
		}
	case "error":

	default:
		log.Warn().Str("type", message.Type).Str("botId", b.ID()).Msg("bot did not handle message")
	}
}

func (b *DiceGameBot) getGameState(message *protocol.Response) (*GameState, bool) {
	gameState, ok := message.Data.(*GameState)
	if !ok {
		log.Error().Str("type", message.Type).Str("botId", b.ID()).Msg("bot could not handle data")
	}

	return gameState, ok
}

func (b *DiceGameBot) makeNextMove(state *GameState) {
	// Add a small delay to simulate thinking
	time.Sleep(BOT_DELAY * time.Millisecond)
	log.Debug().Msg("deciding on next move")

	if err := b.checkRoomStatus(); err != nil {
		// cancel move processing if the room is invalid (closed, ...)
		log.Debug().Msg("room is not okay: " + err.Error())
		return
	}

	log.Debug().Ints("dice", state.Dice).Msg("current dice")

	// 1. Check roll condition
	if b.shouldRoll(state) {
		b.sendAction("roll", nil)
		return
	}

	// 2. Select some dice
	scoringIdx := b.findScoringDiceIdx(state)
	log.Debug().Int("scoringIdx", scoringIdx).Msg("found scoring dice")
	// we still have scoring dice left and havent set aside too many yet
	if scoringIdx != -1 && len(state.SetAside) <= 3 {
		log.Debug().Int("scoringIdx", scoringIdx).Msg("selecting dice")
		b.sendAction("select", map[string]int{"diceIndex": scoringIdx})
		return
	}

	// 3. Set dice aside
	if len(state.SelectedDice) > 0 {
		log.Debug().Ints("selectedDice", state.SelectedDice).Msg("setting dice aside")

		// Decide whether to end turn based on risk assessment
		endTurn := b.shouldEndTurn(state)
		b.sendAction("set_aside", map[string]bool{"endTurn": endTurn})
		return
	}

	log.Warn().
		Ints("dice", state.Dice).
		Ints("selectedIdx", state.SelectedDice).
		Msg("Bot should already have ended turn, but did not. Maybe we couldnt find a combination?")
}

func (b *DiceGameBot) findScoringDiceIdx(state *GameState) int {
	// 1. select all 1s
	// 2. no 1s left, select all 5s
	// 3. gotta check for multiples
	scoringPriorities := []int{1, 5}

	for _, priority := range scoringPriorities {
		for idx, die := range state.Dice {
			// Don't select dice that are already selected
			if slices.Contains(state.SelectedDice, idx) {
				continue
			}

			if die == priority {
				return idx
			}
		}
	}

	if len(b.combinationToSelect) == 0 {
		b.combinationToSelect = b.detectOtherDiceToSelect(state.Dice)
	}

	if len(b.combinationToSelect) > 0 {
		combinationDie := b.combinationToSelect[0]
		stateDiceIdx := -1
		for idx, die := range state.Dice {
			if die != combinationDie {
				continue
			}

			// if we have already selected that dice, continue search
			if slices.Contains(state.SelectedDice, idx) {
				continue
			}

			stateDiceIdx = idx
			// Found a valid index, so break the loop
			break
		}

		if stateDiceIdx != -1 {
			// remove the selected dice from the combination list
			b.combinationToSelect = slices.Delete(b.combinationToSelect, 0, 1)
			return stateDiceIdx
		}

		// No valid dice found in this round, so clear the combination list to avoid getting stuck
		b.combinationToSelect = []int{}
		return -1
	}

	return -1
}

func (b *DiceGameBot) shouldEndTurn(state *GameState) bool {
	// end turn if we have banked at least 3 dice or are about to bank more than 3 dice
	return len(state.SetAside) >= 3 || len(state.SetAside)+len(state.SelectedDice) >= 3
}

func (b *DiceGameBot) checkBotTurn(state *GameState) {
	log.Debug().Msg("checking bot turn")
	// check if bot is the current turn
	if state.CurrentTurn == b.ID() {
		log.Debug().Msg("detected its my turn")
		b.myTurn = true
	} else if b.myTurn {
		log.Debug().Msg("no longer my turn, cleaning up")
		// Reset flags when it's no longer the bot's turn
		b.myTurn = false
		b.busted = false
		b.combinationToSelect = []int{}
	}
}

func (b *DiceGameBot) sendAction(action string, payload interface{}) error {
	messageData, _ := json.Marshal(payload)
	if err := b.SendMessage(action, messageData); err != nil {
		log.Error().Err(err).Str("action", action).Msg("failed to send action")
		return err
	}
	return nil
}

func (b *DiceGameBot) checkRoomStatus() error {
	// Check context before proceeding
	if b.Context().Err() != nil {
		log.Debug().Str("botId", b.ID()).Msg("Context canceled, not making moves")
		return errors.New("context canceled")
	}

	// Additionally check if the room is still valid
	room := b.Room()
	if room == nil || room.IsClosed() {
		log.Debug().Str("botId", b.ID()).Msg("Room is closed or nil, not making moves")
		return errors.New("room is closed")
	}

	return nil
}

func (b *DiceGameBot) shouldRoll(state *GameState) bool {
	// If we have invalid dice (unrolled), roll the dice
	if b.allDiceInvalid(state.Dice) {
		return true
	}
	return false
}

func (b *DiceGameBot) allDiceInvalid(dice []int) bool {
	for _, die := range dice {
		if die != 0 {
			return false
		}
	}

	return true
}

// checkMultiples checks for three of a kind and beyond.
func (b *DiceGameBot) detectOtherDiceToSelect(stateDice []int) []int {
	// Count occurrences for remaining dice
	var combinations []int
	counts := make(map[int]int)
	for _, die := range stateDice {
		counts[die]++
	}

	// Check for three of a kind and beyond
	for num, count := range counts {
		if count >= 3 {
			log.Debug().Int("num", num).Int("count", count).Msg("Found three or more of a kind")
			for i := 0; i < count; i++ {
				combinations = append(combinations, num)
			}
		}
	}

	return combinations
}
