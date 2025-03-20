package interfaces

import (
	"encoding/json"
	"gameserver/internal/protocol"
)


// Client represents a connected websocket client
type Client interface {
	ID() string
	Send(message *protocol.Response) error
	Room() Room
	SetRoom(room Room)
	Close()
}

type Room interface {
	ID() string
	GameType() string
	Join(client Client) error
	Leave(client Client)
	Broadcast(message *protocol.Response, exclude ...Client)
	BroadcastTo(message *protocol.Response, clients ...Client)
	Clients() map[string]Client
	State() interface{}
	SetState(state interface{})
	Close()
}

type RoomManager interface {
	CreateRoom(gameType string, options json.RawMessage) (Room, error)
	GetRoom(roomID string) (Room, error)
	RemoveRoom(roomID string)
}

// Game defines the interface for game implementations
type Game interface {
	Type() string
	HandleMessage(client Client, room Room, msgType string, data []byte)
	InitializeRoom(room Room, options json.RawMessage) error
	OnClientJoin(client Client, room Room)
	OnClientLeave(client Client, room Room)
	OnClientReconnect(client Client, room Room, oldClientId string)
}

type GameRegistry interface {
	RegisterGame(game Game)
	GetGame(gameType string) (Game, error)
	HasGame(gameType string) bool
	InitializeRoom(room Room, options json.RawMessage) error
	HandleMessage(client Client, msgType string, data []byte) error
	HandleClientJoin(client Client, room Room) error
	HandleClientLeave(client Client, room Room) error
	HandleClientReconnect(client Client, room Room, oldClientId string) error
}
