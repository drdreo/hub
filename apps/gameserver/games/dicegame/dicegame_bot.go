package dicegame

import (
	"gameserver/internal/client"
	"gameserver/internal/protocol"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
)

type DiceGameBot struct {
	*client.BotClient
	game *DiceGame
}

func NewDiceGameBot(id string, game *DiceGame) *DiceGameBot {
	bot := &DiceGameBot{
		BotClient: client.NewBotClient(id),
		game:      game,
	}
	bot.BotClient.SetMessageHandler(bot.handleMessage)
	return bot
}

func (b *DiceGameBot) handleMessage(message *protocol.Response) {
	time.Sleep(1 * time.Second) // Simulate thinking time

	switch message.Type {
	case "game_state":
		gameState, _ := b.getGameState(message)
		if !b.isBotTurn(gameState) {
			return
		}
		log.Error().Str("botId", b.ID()).Msg("I HAVE NO CLUE WHAT IM SUPPOSED TO DO YET")
		b.decideTurn(gameState)
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

func (b *DiceGameBot) decideTurn(state *GameState) {
	if !state.Started {
		return
	}

	// Roll the dice
	b.sendAction("roll", nil)
	time.Sleep(1 * time.Second) // Simulate thinking time

	// Select and set aside dice until no scoring dice are left
	for {
		scoringIndexes := b.findScoringDie(state)
		if len(scoringIndexes) == 0 {
			// No scoring dice left, end turn
			b.sendAction("end_turn", nil)
			break
		}

		// Select the scoring die
		for _, scoringIndex := range scoringIndexes {
			selectPayload := map[string]int{"diceIndex": scoringIndex}
			b.sendAction("select", selectPayload)
		}

		// Simulate thinking time
		time.Sleep(1 * time.Second)

		b.sendAction("set_aside", map[string]bool{"endTurn": true})
		// Simulate thinking time
		time.Sleep(1 * time.Second)
	}
}

func (b *DiceGameBot) findScoringDie(state *GameState) []int {
	// Collect indices of dice that are either 1 or 5
	return slices.Collect(func(yield func(int) bool) {
		for _, die := range state.Dice {
			if die == 1 || die == 5 {
				if !yield(die) {
					return
				}
			}
		}
	})
}

func (b *DiceGameBot) sendAction(action string, data interface{}) {
	b.BotClient.Send(protocol.NewSuccessResponse(action, data))
}
