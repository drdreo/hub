package game

import (
    "encoding/json"
    "errors"
    "github.com/drdreo/hub/gameserver/internal/room"
    "maps"
    "slices"
    "sync"
)

type Client interface {
    ID() string
    Send(message []byte) error
    Room() room.Room
    SetRoom(room room.Room)
    Close()
}

// Game defines the interface for game implementations
type Game interface {
    Type() string
    HandleMessage(client Client, room room.Room, msgType string, payload []byte)
    InitializeRoom(room room.Room, options json.RawMessage) error
    OnClientJoin(client Client, room room.Room)
    OnClientLeave(client Client, room room.Room)
}

// Registry manages game registrations
type Registry struct {
    games map[string]Game
    mu    sync.RWMutex
}

// NewRegistry creates a new game registry
func NewRegistry() *Registry {
    return &Registry{
        games: make(map[string]Game),
    }
}

// RegisterGame adds a game to the registry
func (r *Registry) RegisterGame(game Game) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.games[game.Type()] = game
}

// GetGame retrieves a game by type
func (r *Registry) GetGame(gameType string) (Game, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    game, exists := r.games[gameType]
    if !exists {
        return nil, errors.New("game type not registered")
    }

    return game, nil
}

// HasGame checks if a game type is registered
func (r *Registry) HasGame(gameType string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()

    _, exists := r.games[gameType]
    return exists
}

// HandleMessage routes a message to the appropriate game handler
func (r *Registry) HandleMessage(client Client, msgType string, payload []byte) error {
    room := client.Room()
    if room == nil {
        return errors.New("client not in a room")
    }

    gameType := room.GameType()
    game, err := r.GetGame(gameType)
    if err != nil {
        return err
    }

    game.HandleMessage(client, room, msgType, payload)
    return nil
}

// InitializeRoom initializes a room with game-specific state
func (r *Registry) InitializeRoom(room room.Room, options json.RawMessage) error {
    gameType := room.GameType()
    game, err := r.GetGame(gameType)
    if err != nil {
        return err
    }

    return game.InitializeRoom(room, options)
}

// HandleClientJoin notifies the game when a client joins
func (r *Registry) HandleClientJoin(client Client, room room.Room) error {
    gameType := room.GameType()
    game, err := r.GetGame(gameType)
    if err != nil {
        return err
    }

    game.OnClientJoin(client, room)
    return nil
}

// HandleClientLeave notifies the game when a client leaves
func (r *Registry) HandleClientLeave(client Client, room room.Room) error {
    gameType := room.GameType()
    game, err := r.GetGame(gameType)
    if err != nil {
        return err
    }

    game.OnClientLeave(client, room)
    return nil
}

// ListGames returns a list of all registered game types
func (r *Registry) ListGames() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()

    return slices.Collect(maps.Keys(r.games))
}
