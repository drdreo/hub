package dicegame

import (
	"encoding/json"
	"gameserver/internal/client"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
)

type DiceGameBot struct {
	*client.BotClient
	game      *DiceGame
	hasRolled bool
}

func NewDiceGameBot(id string, game *DiceGame, reg interfaces.GameRegistry) *DiceGameBot {
	bot := &DiceGameBot{
		BotClient: client.NewBotClient(id, reg),
		game:      game,
		hasRolled: false,
	}
	bot.BotClient.SetMessageHandler(bot.handleMessage)
	return bot
}

func (b *DiceGameBot) handleMessage(message *protocol.Response) {
	time.Sleep(1 * time.Second) // Simulate thinking time

	switch message.Type {
	case "game_state":
		gameState, ok := b.getGameState(message)
		if !ok || !gameState.Started || !b.isBotTurn(gameState) {
			return
		}
		b.makeNextMove(gameState)

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

func (b *DiceGameBot) isBotTurn(state *GameState) bool {
	if state.CurrentTurn == b.ID() {
		return true
	}
	return false
}

func (b *DiceGameBot) makeNextMove(state *GameState) {
	// Add a small delay to simulate thinking
	time.Sleep(500 * time.Millisecond)

	log.Debug().Ints("dice", state.Dice).Msg("current dice")

	// If we still have all dice and havent selected any dice yet, roll the dice
	if !b.hasRolled {
		if err := b.sendAction("roll", nil); err == nil{
			b.hasRolled = true
		}
		return
	}

	// If we have selected dice, set them aside
	if len(state.SelectedDice) > 0 {
		// Decide whether to end turn based on risk assessment
		endTurn := b.shouldEndTurn(state)
		b.sendAction("set_aside", map[string]bool{"endTurn": endTurn})
		if endTurn {
			b.hasRolled = false
		}
		return
	}

	// Find scoring dice to select
	scoringIndexes := b.findScoringDiceIndexes(state)
	log.Debug().Ints("scoringIndexes", scoringIndexes).Msg("found scoring dice")
	if len(scoringIndexes) > 0 {
		// Select the first scoring die
		b.sendAction("select", map[string]int{"diceIndex": scoringIndexes[0]})
		return
	}

	// No scoring dice left, end turn
	b.endTurn()
}

func (b *DiceGameBot) endTurn() {
	b.sendAction("end_turn", nil)
	b.hasRolled = false
}

func (b *DiceGameBot) findScoringDiceIndexes(state *GameState) []int {
	// Collect indices of dice that are either 1 or 5
	return slices.Collect(func(yield func(int) bool) {
		for dieIdx, die := range state.Dice {
			if die == 1 || die == 5 {
				if !yield(dieIdx) {
					return
				}
			}
		}
	})
}

func (b *DiceGameBot) shouldEndTurn(state *GameState) bool {
	// Simple logic - end turn if we have accumulated some points
	return len(state.SetAside) > 0
}

func (b *DiceGameBot) sendAction(action string, payload interface{}) error {
	messageData, _ := json.Marshal(payload)
	if err := b.BotClient.SendMessage(action, messageData); err != nil {
		log.Error().Err(err).Str("action", action).Msg("failed to send action")
		return err
	}
	return nil
}
