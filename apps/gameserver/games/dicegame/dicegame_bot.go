package dicegame

import (
	"gameserver/internal/client"
	"gameserver/internal/protocol"
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
	// TODO bot logic
	//	b.sendAction("bank", nil)
}

//func (b *DiceGameBot) sendAction(action string, data interface{}) {
//	// Construct action message
//	actionMsg := &protocol.Request{
//		Type: action,
//		Data: data,
//	}
//
//	// Send through game's HandleAction method (assuming it exists)
//	actionJSON, _ := json.Marshal(actionMsg)
//	b.game.HandleAction(b.ID(), actionJSON)
//}
