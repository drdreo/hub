package database

import (
	"context"
	"gameserver/games/owe_drahn/models"
	"gameserver/games/owe_drahn/utils"
	"gameserver/internal/database/firestore"
	"github.com/rs/zerolog/log"

	googlefirestore "cloud.google.com/go/firestore"
)

// Database defines the methods required for database operations
type Database interface {
	StoreGame(ctx context.Context, state models.DBGame) error
	GetUserStats(ctx context.Context, uid string) (*models.PlayerStats, error)
	GetAllGames(ctx context.Context) ([]models.DBGame, error)
	GetUser(ctx context.Context, uid string) (*models.DBUser, error)
}

// DatabaseService handles database operations for the owe_drahn game
type DatabaseService struct {
	db firestore.Database
}

// NewDatabaseService creates a new instance of the database service
func NewDatabaseService(db firestore.Database) *DatabaseService {
	return &DatabaseService{
		db: db,
	}
}

// StoreGame stores a game in the database and updates player statistics
func (s *DatabaseService) StoreGame(ctx context.Context, state models.DBGame) error {
	log.Info().Msg("storing game state")
	// Store the game
	if err := s.db.Create(ctx, "games", state); err != nil {
		log.Error().Err(err).Msg("failed to store game state")
		return err
	}

	// Update stats for each player
	for _, player := range state.Players {
		if player.UID == "" {
			continue
		}
		if err := s.updateUserStats(ctx, player, state); err != nil {
			log.Error().Str("playerID", player.UID).Err(err).Msg("Failed to update player stats")
		}
	}

	return nil
}

// GetUserStats retrieves a player's statistics
func (s *DatabaseService) GetUserStats(ctx context.Context, uid string) (*models.PlayerStats, error) {
	var user models.DBUser
	err := s.db.Get(ctx, "users", uid, &user)
	if err != nil {
		log.Error().Msgf("Error getting player stats: %v", err)
		return nil, err
	}

	return &user.Stats, nil
}

// updateUserStats updates a player's statistics with data from a new game
func (s *DatabaseService) updateUserStats(ctx context.Context, player *models.FormattedPlayer, game models.DBGame) error {
	uid := player.UID
	newStats := utils.ExtractPlayerStats(uid, game)

	var user models.DBUser
	err := s.db.Get(ctx, "users", uid, &user)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user document")
		return err
	}

	user.Stats = MergeStats(user.Stats, newStats)

	updates := []googlefirestore.Update{
		{Path: "stats", Value: user.Stats},
	}

	log.Debug().Str("username", player.Username).Str("uid", uid).Msg("updating user stats ")
	if err = s.db.Update(ctx, "users", uid, updates); err != nil {
		log.Error().Err(err).Msg("failed to update user stats")
		return err
	}

	return nil
}

// GetAllGames retrieves all games from the database
func (s *DatabaseService) GetAllGames(ctx context.Context) ([]models.DBGame, error) {
	var games []models.DBGame
	err := s.db.Query(ctx, "games", nil, &games)
	if err != nil {
		log.Error().Err(err).Msg("failed to query games")
		return nil, err
	}

	return games, nil
}

// GetUser retrieves a user by UID
func (s *DatabaseService) GetUser(ctx context.Context, uid string) (*models.DBUser, error) {
	var user models.DBUser
	err := s.db.Get(ctx, "users", uid, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// MergeStats combines existing player stats with new stats from a game
func MergeStats(oldStats models.PlayerStats, newStats models.PlayerStatAggregation) models.PlayerStats {
	stats := oldStats
	stats.PerfectRoll += newStats.PerfectRoll
	stats.LuckiestRoll += newStats.LuckiestRoll
	stats.WorstRoll += newStats.WorstRoll
	stats.Rolled21 += newStats.Rolled21
	stats.MaxLifeLoss += newStats.MaxLifeLoss

	if newStats.Won {
		stats.Wins++
	}
	stats.TotalGames++

	for i := 0; i < 6; i++ {
		stats.RolledDice[i] += newStats.RolledDice[i]
	}

	return stats
}

// Helper function to check if an error is a "not found" error
func isNotFoundError(err error) bool {
	return err != nil && (err.Error() == "document not found" || err.Error() == "not found" ||
		err.Error() == "rpc error: code = NotFound desc = " ||
		err.Error() == "firestore: document doesn't exist")
}
