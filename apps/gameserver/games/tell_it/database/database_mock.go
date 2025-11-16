package database

import (
	"context"
	"gameserver/games/tell_it/models"
)

// DatabaseServiceMock provides a mock implementation for testing
type DatabaseServiceMock struct{}

func (m *DatabaseServiceMock) StoreStories(ctx context.Context, roomName string, stories []models.StoryDTO) error {
	return nil
}

func (m *DatabaseServiceMock) GetStories(ctx context.Context) ([]models.DBStory, error) {
	return nil, nil
}

func (m *DatabaseServiceMock) Close() error {
	return nil
}
