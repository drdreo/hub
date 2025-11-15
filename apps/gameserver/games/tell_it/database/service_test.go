package database

import (
	"context"
	"gameserver/games/tell_it/models"
	"os"
	"testing"
	"time"
)

func TestDatabaseService_StoreAndGetStories(t *testing.T) {
	// Use in-memory SQLite for testing
	os.Setenv("DATABASE_URL", "file::memory:?cache=shared")
	defer os.Unsetenv("DATABASE_URL")

	ctx := context.Background()
	factory := NewDatabaseFactory("development")

	service, err := factory.CreateDatabaseService(ctx)
	if err != nil {
		t.Fatalf("Failed to create database service: %v", err)
	}
	defer service.Close()

	// Test data
	roomName := "test-room-" + time.Now().Format("20060102150405")
	stories := []models.StoryData{
		{
			Text:   "Once upon a time there was a brave knight",
			Author: "Alice",
		},
		{
			Text:   "In a galaxy far far away",
			Author: "Bob",
		},
	}

	// Store stories
	err = service.StoreStories(ctx, roomName, stories)
	if err != nil {
		t.Errorf("Failed to store stories: %v", err)
	}

	// Retrieve stories
	retrieved, err := service.GetStories(ctx)
	if err != nil {
		t.Errorf("Failed to retrieve stories: %v", err)
	}

	// Verify count
	if len(retrieved) != len(stories) {
		t.Errorf("Expected %d stories, got %d", len(stories), len(retrieved))
	}

	// Verify content
	for i, story := range retrieved {
		if story.Text != stories[i].Text {
			t.Errorf("Story %d: expected text %s, got %s", i, stories[i].Text, story.Text)
		}
		if story.Author != stories[i].Author {
			t.Errorf("Story %d: expected author %s, got %s", i, stories[i].Author, story.Author)
		}
	}
}
