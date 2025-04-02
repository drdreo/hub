package session

import (
	"gameserver/internal/interfaces"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type SessionData struct {
	ClientID  string
	RoomID    string
	GameType  string
	LeftAt    time.Time
	ExtraData interfaces.M
}

type Store struct {
	sessions      map[string]SessionData
	mu            sync.RWMutex
	expirySeconds int64
}

// Global session store instance
var (
	globalStore *Store
)

// InitGlobalStore initializes the global session store with the given expiry
func InitGlobalStore(expirySeconds int64) {
	globalStore = NewStore(expirySeconds)
}

// GetSessionStore returns the global session store instance
func GetSessionStore() *Store {
	return globalStore
}

func NewStore(expirySeconds int64) *Store {
	store := &Store{
		sessions:      make(map[string]SessionData),
		expirySeconds: expirySeconds,
	}

	log.Debug().Int64("expiry", expirySeconds).Msg("created new session store")

	go store.cleanupRoutine()
	return store
}

func (s *Store) StoreSession(clientID string, data SessionData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data.LeftAt = time.Now()

	s.sessions[clientID] = data
	log.Debug().Str("clientId", clientID).Time("leftAt", data.LeftAt).Msg("session stored")
}

func (s *Store) GetSession(clientID string) (SessionData, bool) {
	// TODO: put locks back
	// s.mu.RLock()
	// defer s.mu.RUnlock()
	data, exists := s.sessions[clientID]
	if !exists {
		return SessionData{}, false
	}

	if time.Since(data.LeftAt).Seconds() > float64(s.expirySeconds) {
		return SessionData{}, false
	}
	return data, true
}

func (s *Store) RemoveSession(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Debug().Str("clientId", clientID).Msg("removing session")
	delete(s.sessions, clientID)
}

func (s *Store) cleanupRoutine() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

func (s *Store) cleanup() {
	log.Info().Msg("check expired sessions and cleanup")

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, session := range s.sessions {
		if now.Sub(session.LeftAt).Seconds() > float64(s.expirySeconds) {
			delete(s.sessions, id)
		}
	}
}
