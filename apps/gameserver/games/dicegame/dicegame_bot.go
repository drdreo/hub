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
	game   *DiceGame
	myTurn bool
	busted bool
}

func NewDiceGameBot(id string, game *DiceGame, reg interfaces.GameRegistry) *DiceGameBot {
	bot := &DiceGameBot{
		BotClient: client.NewBotClient(id, reg),
		game:      game,
		myTurn:    false,
		busted:    false,
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
			return
		}
		// Don't make moves if we're busted
		if b.busted {
			return
		}

		b.makeNextMove(gameState)

	case "error":
		// Check for bust notification
		if b.myTurn && message.Error == ErrBusted.Error() {
			log.Warn().Msg("Bot detected bust and will wait for turn to end")
			b.busted = true
		}

	default:
		log.Warn().Str("type", message.Type).Str("botId", b.ID()).Msg("bot could not handle data")
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

	if err := b.checkRoomStatus(); err != nil {
		// cancel move processing if the room is invalid (closed, ...)
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
		b.sendAction("select", map[string]int{"diceIndex": scoringIdx})
		return
	}

	// 3. Set dice aside
	if len(state.SelectedDice) > 0 {
		// Decide whether to end turn based on risk assessment
		endTurn := b.shouldEndTurn(state)
		b.sendAction("set_aside", map[string]bool{"endTurn": endTurn})
		return
	}

	log.Warn().Msg("Bot should already have ended turn, but did not. Weird.")
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

	return -1

}

func (b *DiceGameBot) shouldEndTurn(state *GameState) bool {
	// end turn if we have banked at least 2 dice
	return len(state.SetAside) >= 3 || len(state.SelectedDice) >= 3
}

func (b *DiceGameBot) checkBotTurn(state *GameState) {
	// check if bot is the current turn
	if state.CurrentTurn == b.ID() {
		b.myTurn = true
	} else if b.myTurn {
		// Reset flags when it's no longer the bot's turn
		b.myTurn = false
		b.busted = false
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
