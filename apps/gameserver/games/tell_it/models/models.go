package models

import (
	"time"
)

// Story represents a single story in the game
type Story struct {
	ID       string    `json:"id" db:"id"`
	OwnerID  string    `json:"ownerId" db:"owner_id"`
	Texts    []string  `json:"texts" db:"texts"`
	RoomName string    `json:"roomName" db:"room_name"`
	Created  time.Time `json:"created" db:"created"`
}

// User represents a user in the game
type User struct {
	ID            string   `json:"id" db:"id"`
	Name          string   `json:"name" db:"name"`
	Disconnected  bool     `json:"disconnected" db:"disconnected"`
	AFK           bool     `json:"afk" db:"afk"`
	KickVotes     []string `json:"kickVotes" db:"kick_votes"`
	QueuedStories int      `json:"queuedStories" db:"queued_stories"`
}

// Room represents a game room
type Room struct {
	Name              string    `json:"name" db:"name"`
	Started           bool      `json:"started" db:"started"`
	StartTime         time.Time `json:"startTime" db:"start_time"`
	GameStatus        string    `json:"gameStatus" db:"game_status"` // "waiting", "started", "ended"
	UserCount         int       `json:"userCount" db:"user_count"`
	IsPublic          bool      `json:"isPublic" db:"is_public"`
	SpectatorsAllowed bool      `json:"spectatorsAllowed" db:"spectators_allowed"`
}

// DBStory represents a stored story with all its details
type DBStory struct {
	ID        int64     `json:"id,omitempty" db:"id"`
	Text      string    `json:"text" db:"text"` // Serialized story (all texts joined)
	Author    string    `json:"author" db:"author"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// RoomConfig represents room configuration
type RoomConfig struct {
	SpectatorsAllowed bool `json:"spectatorsAllowed"`
	IsPublic          bool `json:"isPublic"`
	MinUsers          int  `json:"minUsers"`
	MaxUsers          int  `json:"maxUsers"`
	AFKDelay          int  `json:"afkDelay"` // milliseconds
}

// StoryData represents a story with author information
type StoryData struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

// UserOverview represents a user's overview information
type UserOverview struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Disconnected  bool     `json:"disconnected"`
	AFK           bool     `json:"afk"`
	KickVotes     []string `json:"kickVotes"`
	QueuedStories int      `json:"queuedStories"`
}
