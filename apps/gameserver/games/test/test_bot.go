package testgame

import (
	"gameserver/internal/client"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
)

type Bot struct {
	*client.BotClient
}

func NewBot(id string, game *TestGame, reg interfaces.GameRegistry) *Bot {
	bot := &Bot{
		BotClient: client.NewBotClient(id, reg),
	}
	bot.BotClient.SetMessageHandler(bot.handleMessage)
	return bot
}

func (b *Bot) handleMessage(message *protocol.Response) {
}
