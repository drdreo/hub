package client

import (
    "gameserver/internal/interfaces"
    "sync"
)

type ClientMock struct {
    id       string
    room     interfaces.Room
    mu       sync.Mutex
    messages [][]byte
}

func (m *ClientMock) ID() string {
    return m.id
}

func (m *ClientMock) Send(message []byte) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.messages = append(m.messages, message)
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

func (m *ClientMock) Close() {}

func (m *ClientMock) GetSentMessages() [][]byte {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.messages
}

func (m *ClientMock) ClearMessages() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.messages = make([][]byte, 0)
}

func NewClientMock(id string) *ClientMock {
    return &ClientMock{
        id:       id,
        messages: make([][]byte, 0),
    }
}
