package game

import (
	"encoding/json"
	"errors"
	"gameserver/internal/interfaces"
	"github.com/rs/zerolog/log"
	"maps"
	"slices"
	"sync"
)

// Registry manages game registrations
type Registry struct {
	games map[string]interfaces.Game
	mu    sync.RWMutex
}

// NewRegistry creates a new game registry
func NewRegistry() *Registry {
	log.Debug().Msg("game registry created")
	return &Registry{
		games: make(map[string]interfaces.Game),
	}
}

// RegisterGame adds a game to the registry
func (r *Registry) RegisterGame(game interfaces.Game) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.games[game.Type()] = game
	log.Debug().Str("type", game.Type()).Msg("game registered")
}

// GetGame retrieves a game by type
func (r *Registry) GetGame(gameType string) (interfaces.Game, error) {
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
func (r *Registry) HandleMessage(client interfaces.Client, msgType string, data []byte) error {
	room := client.Room()
	if room == nil {
		return errors.New("client not in a room")
	}

	gameType := room.GameType()
	game, err := r.GetGame(gameType)
	if err != nil {
		return err
	}

	game.HandleMessage(client, room, msgType, data)
	return nil
}

// InitializeRoom initializes a room with game-specific state
func (r *Registry) InitializeRoom(room interfaces.Room, options json.RawMessage) error {
	gameType := room.GameType()
	game, err := r.GetGame(gameType)
	if err != nil {
		return err
	}

	return game.InitializeRoom(room, options)
}

// HandleClientJoin notifies the game when a client joins
func (r *Registry) HandleClientJoin(client interfaces.Client, room interfaces.Room, options interfaces.CreateRoomOptions) error {
	gameType := room.GameType()
	game, err := r.GetGame(gameType)
	if err != nil {
		return err
	}

	// Join the room
	if err = room.Join(client); err != nil {
		log.Error().Err(err).Str("id", room.ID()).Msg("failed to join room")
		return err
	}

	game.OnClientJoin(client, room, options)
	return nil
}

func (r *Registry) HandleAddBot(client interfaces.Client, room interfaces.Room) error {
	gameType := room.GameType()
	game, err := r.GetGame(gameType)
	if err != nil {
		return err
	}

	botClient, err := game.OnBotAdd(client, client.Room(), r)
	if err != nil {
		return err
	}

	return r.HandleClientJoin(botClient, client.Room(), interfaces.CreateRoomOptions{
		PlayerName: "Bot-1",
	})
}

// HandleClientLeave notifies the game when a client leaves
func (r *Registry) HandleClientLeave(client interfaces.Client, room interfaces.Room) error {
	gameType := room.GameType()
	game, err := r.GetGame(gameType)
	if err != nil {
		return err
	}

	game.OnClientLeave(client, room)
	return nil
}

// HandleClientReconnect notifies the game when a client leaves
func (r *Registry) HandleClientReconnect(client interfaces.Client, room interfaces.Room, oldClientId string) error {
	gameType := room.GameType()
	game, err := r.GetGame(gameType)
	if err != nil {
		return err
	}

	game.OnClientReconnect(client, room, oldClientId)
	return nil
}

// ListGames returns a list of all registered game types
func (r *Registry) ListGames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return slices.Collect(maps.Keys(r.games))
}
