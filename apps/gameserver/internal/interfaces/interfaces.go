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
	IsClosed() bool
	Join(client Client) error
	Leave(client Client)
	Broadcast(message *protocol.Response, exclude ...Client)
	BroadcastTo(message *protocol.Response, clients ...Client)
	Clients() map[string]Client
	State() interface{}
	SetState(state interface{})
	Close()
}

type CreateRoomOptions struct {
	GameType   string          `json:"gameType"`
	PlayerName string          `json:"playerName"`
	RoomID     *string         `json:"roomId,omitempty"`
	Options    json.RawMessage `json:"options,omitempty"`
}

type RoomManager interface {
	CreateRoom(createOptions CreateRoomOptions) (Room, error)
	GetRoom(roomID string) (Room, error)
	RemoveRoom(roomID string)
}

// Game defines the interface for game implementations
type Game interface {
	Type() string
	HandleMessage(client Client, room Room, msgType string, data []byte)
	InitializeRoom(room Room, options json.RawMessage) error
	OnClientJoin(client Client, room Room, options CreateRoomOptions)
	OnClientLeave(client Client, room Room)
	OnClientReconnect(client Client, room Room, oldClientId string)
	OnBotAdd(client Client, room Room) (Client, error)
}

type GameRegistry interface {
	RegisterGame(game Game)
	GetGame(gameType string) (Game, error)
	HasGame(gameType string) bool
	InitializeRoom(room Room, options json.RawMessage) error
	HandleMessage(client Client, msgType string, data []byte) error
	HandleClientJoin(client Client, room Room, options CreateRoomOptions) error
	HandleClientLeave(client Client, room Room) error
	HandleClientReconnect(client Client, room Room, oldClientId string) error
    HandleAddBot(client Client, room Room) error
}

type M map[string]interface{}
