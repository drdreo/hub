package client

import (
	"context"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"sync"

	"github.com/rs/zerolog/log"
)

type BotClient struct {
	id             string
	room           interfaces.Room
	gameRegistry   interfaces.GameRegistry
	messages       []*protocol.Response
	messageHandler func(*protocol.Response)
	cancelCtx      context.Context
	cancelFunc     context.CancelFunc
	mu             sync.Mutex
}

func NewBotClient(id string, reg interfaces.GameRegistry) *BotClient {
	ctx, cancel := context.WithCancel(context.Background())
	bot := &BotClient{
		id:           id,
		messages:     make([]*protocol.Response, 0),
		gameRegistry: reg,
		cancelCtx:    ctx,
		cancelFunc:   cancel,
	}
	bot.messageHandler = bot.defaultMessageHandler
	return bot
}

func (b *BotClient) ID() string {
	return b.id
}

func (b *BotClient) IsBot() bool {
	return true
}

func (b *BotClient) Context() context.Context {
	return b.cancelCtx
}

// Send - sends a message TO the bot
func (b *BotClient) Send(message *protocol.Response) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// For history keeping. Concerned about memory consumption
	// b.messages = append(b.messages, message)
	log.Debug().Fields(message).Str("clientId", b.id).Msgf("BotClient Receives(%s)", message.Type)

	if message.Success == false {
		log.Error().Str("err", message.Error).Str("clientId", b.id).Msg("bot received error")
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
	roomId := "no-room"
	if b.room != nil && !b.room.IsClosed() {
		roomId = b.room.ID()
	}
	log.Info().
		Str("clientId", b.id).
		Str("roomId", roomId).
		Msg("BotClient shutting down and cleaning up resources")

	// Cancel all goroutines
	b.cancelFunc()

	// Clear message queue to help with garbage collection
	b.mu.Lock()
	b.messages = make([]*protocol.Response, 0)
	b.messageHandler = b.defaultMessageHandler
	b.mu.Unlock()
}

func (b *BotClient) defaultMessageHandler(message *protocol.Response) {
	log.Info().Str("botId", b.ID()).Msg("Default message handler (no action taken)")
}

func (b *BotClient) SetMessageHandler(handler func(*protocol.Response)) {
	b.messageHandler = handler
}
