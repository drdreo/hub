package models

import (
	"time"
)

// UserDTO represents a client-side user's overview
type UserDTO struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Disconnected  bool     `json:"disconnected"`
	AFK           bool     `json:"afk"`
	KickVotes     []string `json:"kickVotes"`
	QueuedStories int      `json:"queuedStories"`
}

// RoomConfig represents room configuration
type RoomConfig struct {
	SpectatorsAllowed bool `json:"spectatorsAllowed"`
	IsPublic          bool `json:"isPublic"`
	MinUsers          int  `json:"minUsers"`
	MaxUsers          int  `json:"maxUsers"`
	AFKDelay          int  `json:"afkDelay"` // milliseconds
}

// DBStory represents a DB stored story
type DBStory struct {
	ID        *int64    `json:"id,omitempty" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Text      string    `json:"text" db:"text"` // Serialized story (all texts joined)
	Author    string    `json:"author" db:"author"`
}

// StoryDTO represents a story with author information
type StoryDTO struct {
	Text   string     `json:"text"`
	Author string     `json:"author"`
	Stats  StoryStats `json:"stats"`
}

type StoryStats struct {
	AvgReadingTime float64 `json:"avgReadingTime"` // average story reading time in seconds, calculation: (words / 200) * 60
	Words          int     `json:"words"`
	Turns          int     `json:"turns"`
}
