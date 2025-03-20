package client

import (
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"github.com/rs/zerolog/log"
	"sync"
)

type ClientMock struct {
	id       string
	room     interfaces.Room
	mu       sync.Mutex
	messages []*protocol.Response
}

func (m *ClientMock) ID() string {
	return m.id
}

func (m *ClientMock) Send(message *protocol.Response) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, message)

	log.Info().Fields(message).Str("clientId", m.id).Msg("Send()")
	return nil
}

func (m *ClientMock) Room() interfaces.Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.room
}

func (m *ClientMock) SetRoom(room interfaces.Room) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.room = room
}

func (m *ClientMock) Close() {
	log.Info().Str("clientId", m.id).Msg("Close()")
}

func (m *ClientMock) GetSentMessages() []*protocol.Response {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages
}

func (m *ClientMock) ClearMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]*protocol.Response, 0)
}

func NewClientMock(id string) *ClientMock {
	return &ClientMock{
		id:       id,
		messages: make([]*protocol.Response, 0),
	}
}
