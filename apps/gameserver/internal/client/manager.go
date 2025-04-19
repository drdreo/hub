package client

import (
	"gameserver/internal/interfaces"
	"github.com/rs/zerolog/log"
	"sync"
)

// Manager manages all connected WebSocket clients
type Manager struct {
	// All connected clients, does NOT include bots for now
	clients map[string]interfaces.Client
	// Clients organized by game type interest, does NOT include bots for now
	clientsByGameType map[string]map[string]interfaces.Client
	mutex             sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients:           make(map[string]interfaces.Client),
		clientsByGameType: make(map[string]map[string]interfaces.Client),
	}
}

// RegisterClient adds a client to the manager
func (m *Manager) RegisterClient(client interfaces.Client, gameType string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	log.Debug().Str("clientID", client.ID()).Str("gameType", gameType).Msg("registering client")

	m.clients[client.ID()] = client

	if gameType != "" {
		// If this is the first client interested in this game type, initialize the map
		if _, exists := m.clientsByGameType[gameType]; !exists {
			m.clientsByGameType[gameType] = make(map[string]interfaces.Client)
		}

		m.clientsByGameType[gameType][client.ID()] = client
	}
}

// UnregisterClient removes a client from the manager
func (m *Manager) UnregisterClient(client interfaces.Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	log.Debug().Str("clientID", client.ID()).Msg("unregistering client")
	clientID := client.ID()

	// Remove from all game type interests
	for gameType, clients := range m.clientsByGameType {
		delete(clients, clientID)
		// Clean up empty maps
		if len(clients) == 0 {
			delete(m.clientsByGameType, gameType)
		}
	}

	delete(m.clients, clientID)
}

// GetClientsByGameType returns all clients interested in a specific game type
func (m *Manager) GetClientsByGameType(gameType string) []interfaces.Client {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	gameTypeClients, exists := m.clientsByGameType[gameType]
	if !exists {
		return []interfaces.Client{}
	}

	clients := make([]interfaces.Client, 0, len(gameTypeClients))
	for _, client := range gameTypeClients {
		clients = append(clients, client)
	}

	return clients
}
