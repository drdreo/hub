package client

import (
	"errors"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"sync"

	"github.com/rs/zerolog/log"
)

type BotClient struct {
	id             string
	room           interfaces.Room
	gameRegistry   interfaces.GameRegistry
	mu             sync.Mutex
	messages       []*protocol.Response
	messageHandler func(*protocol.Response)
}

func NewBotClient(id string, reg interfaces.GameRegistry) *BotClient {
	bot := &BotClient{
		id:           id,
		messages:     make([]*protocol.Response, 0),
		gameRegistry: reg,
	}
	bot.messageHandler = bot.defaultMessageHandler
	return bot
}

func (b *BotClient) ID() string {
	return b.id
}

// Send - sends a message TO the bot
func (b *BotClient) Send(message *protocol.Response) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages = append(b.messages, message)
	log.Debug().Fields(message).Str("clientId", b.id).Msgf("BotClient Receives(%s)", message.Type)

	if message.Success == false {
		log.Error().Str("err", message.Error).Str("clientId", b.id).Msg("bot received error")
		return errors.New(message.Error)
	}
	// Process the message asynchronously
	go b.messageHandler(message)
	return nil
}

// SendMessage - sends a message FROM the bot
func (b *BotClient) SendMessage(action string, data []byte) error {
	log.Debug().Bytes("data", data).Str("clientId", b.id).Msgf("BotClient Sends(%s)", action)

	return b.gameRegistry.HandleMessage(b, action, data)
}

func (b *BotClient) Room() interfaces.Room {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.room
}

func (b *BotClient) SetRoom(room interfaces.Room) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.room = room
}

func (b *BotClient) Close() {
	log.Info().Str("clientId", b.id).Msg("BotClient Close()")
}

func (b *BotClient) defaultMessageHandler(message *protocol.Response) {
	log.Info().Str("botId", b.ID()).Msg("Default message handler (no action taken)")
}

func (b *BotClient) SetMessageHandler(handler func(*protocol.Response)) {
	b.messageHandler = handler
}
