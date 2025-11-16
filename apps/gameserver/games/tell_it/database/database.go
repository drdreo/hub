package database

import (
	"context"
	"gameserver/games/tell_it/models"
)

// Database defines the methods required for database operations
type Database interface {
	StoreStories(ctx context.Context, roomName string, stories []models.StoryDTO) error
	GetStories(ctx context.Context) ([]models.DBStory, error)
	Close() error
}
