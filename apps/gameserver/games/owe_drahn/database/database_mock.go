package database

import (
	"context"
	"gameserver/games/owe_drahn/models"
)

// DatabaseServiceMock provides a mock implementation for testing
type DatabaseServiceMock struct{}

func (m *DatabaseServiceMock) StoreGame(ctx context.Context, state models.DBGame) error {
	return nil
}

func (m *DatabaseServiceMock) GetUserStats(ctx context.Context, uid string) (*models.PlayerStats, error) {
	return &models.PlayerStats{}, nil
}

func (m *DatabaseServiceMock) GetAllGames(ctx context.Context) ([]models.DBGame, error) {
	return nil, nil
}

func (m *DatabaseServiceMock) GetUser(ctx context.Context, uid string) (*models.DBUser, error) {
	return nil, nil
}
