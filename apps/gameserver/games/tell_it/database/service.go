package database

import (
	"context"
	"fmt"
	"gameserver/games/tell_it/models"
	"gameserver/internal/database/sql"
	"github.com/rs/zerolog/log"
	"time"
)

// DatabaseService handles database operations for the tell_it game
type DatabaseService struct {
	db *sql.Client
}

// NewDatabaseService creates a new instance of the database service
func NewDatabaseService(db *sql.Client) *DatabaseService {
	return &DatabaseService{
		db: db,
	}
}

// StoreStories stores multiple stories in the database using a transaction
func (s *DatabaseService) StoreStories(ctx context.Context, roomName string, stories []models.StoryData) error {
	log.Info().Str("room", roomName).Int("count", len(stories)).Msg("storing stories")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Str("room", roomName).Msg("failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, story := range stories {
		dbStory := models.DBStory{
			Text:      story.Text,
			Author:    story.Author,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := tx.Create(ctx, "stories", &dbStory)
		if err != nil {
			log.Error().Err(err).Str("room", roomName).Msg("failed to store story")
			return fmt.Errorf("failed to store story: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Str("room", roomName).Msg("failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetStories retrieves all stories
func (s *DatabaseService) GetStories(ctx context.Context) ([]models.DBStory, error) {
	log.Info().Msg("retrieving stories")

	var stories []models.DBStory


	err := s.db.Query(ctx, "SELECT * FROM stories", &stories)
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve stories")
		return nil, fmt.Errorf("failed to retrieve stories: %w", err)
	}

	return stories, nil
}

// Close closes the database connection
func (s *DatabaseService) Close() error {
	return s.db.Close()
}
