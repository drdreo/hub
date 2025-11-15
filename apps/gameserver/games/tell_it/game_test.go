package tell_it

import (
	"context"
	"gameserver/games/tell_it/models"
	"testing"
)

// MockDatabase implements the database.Database interface for testing
type MockDatabase struct {
	StoredStories map[string][]models.StoryData
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		StoredStories: make(map[string][]models.StoryData),
	}
}

func (m *MockDatabase) StoreStories(ctx context.Context, roomName string, stories []models.StoryData) error {
	m.StoredStories[roomName] = stories
	return nil
}

func (m *MockDatabase) GetStories(ctx context.Context, roomName string) ([]models.DBStory, error) {
	return nil, nil
}

func (m *MockDatabase) Close() error {
	return nil
}

func TestNewStory(t *testing.T) {
	story := NewStory("user1")

	if story.OwnerID != "user1" {
		t.Errorf("Expected OwnerID to be 'user1', got '%s'", story.OwnerID)
	}

	if len(story.Texts) != 0 {
		t.Errorf("Expected empty Texts array, got length %d", len(story.Texts))
	}
}

func TestStory_AddText(t *testing.T) {
	story := NewStory("user1")
	story.AddText("Once upon a time")
	story.AddText("there was a brave knight")

	if len(story.Texts) != 2 {
		t.Errorf("Expected 2 texts, got %d", len(story.Texts))
	}

	if story.Texts[0] != "Once upon a time" {
		t.Errorf("Expected first text to be 'Once upon a time', got '%s'", story.Texts[0])
	}
}

func TestStory_GetLatestText(t *testing.T) {
	story := NewStory("user1")

	// Test empty story
	if story.GetLatestText() != "" {
		t.Errorf("Expected empty string for empty story, got '%s'", story.GetLatestText())
	}

	story.AddText("First text")
	story.AddText("Second text")

	if story.GetLatestText() != "Second text" {
		t.Errorf("Expected 'Second text', got '%s'", story.GetLatestText())
	}
}

func TestStory_Serialize(t *testing.T) {
	story := NewStory("user1")
	story.AddText("Once upon a time")
	story.AddText("there was a brave knight")
	story.AddText("who fought a dragon")

	expected := "Once upon a time. there was a brave knight. who fought a dragon"
	result := story.Serialize()

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestNewUser(t *testing.T) {
	user := NewUser("user1", "Alice")

	if user.ID != "user1" {
		t.Errorf("Expected ID to be 'user1', got '%s'", user.ID)
	}

	if user.Name != "Alice" {
		t.Errorf("Expected Name to be 'Alice', got '%s'", user.Name)
	}

	if user.Disconnected {
		t.Error("Expected Disconnected to be false")
	}

	if user.AFK {
		t.Error("Expected AFK to be false")
	}

	if len(user.StoryQueue) != 0 {
		t.Errorf("Expected empty StoryQueue, got length %d", len(user.StoryQueue))
	}
}

func TestUser_EnqueueDequeue(t *testing.T) {
	user := NewUser("user1", "Alice")
	story1 := NewStory("user2")
	story2 := NewStory("user3")

	user.EnqueueStory(story1)
	user.EnqueueStory(story2)

	if len(user.StoryQueue) != 2 {
		t.Errorf("Expected 2 stories in queue, got %d", len(user.StoryQueue))
	}

	// Dequeue first story
	dequeued, err := user.DequeueStory()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if dequeued.OwnerID != "user2" {
		t.Errorf("Expected dequeued story to be from user2, got '%s'", dequeued.OwnerID)
	}

	if len(user.StoryQueue) != 1 {
		t.Errorf("Expected 1 story in queue after dequeue, got %d", len(user.StoryQueue))
	}

	// Dequeue second story
	dequeued, err = user.DequeueStory()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if dequeued.OwnerID != "user3" {
		t.Errorf("Expected dequeued story to be from user3, got '%s'", dequeued.OwnerID)
	}

	// Try to dequeue from empty queue
	_, err = user.DequeueStory()
	if err == nil {
		t.Error("Expected error when dequeuing from empty queue")
	}
}

func TestUser_GetCurrentStory(t *testing.T) {
	user := NewUser("user1", "Alice")

	// Test empty queue
	if user.GetCurrentStory() != nil {
		t.Error("Expected nil for empty queue")
	}

	story := NewStory("user2")
	user.EnqueueStory(story)

	current := user.GetCurrentStory()
	if current == nil {
		t.Error("Expected story, got nil")
	}

	if current.OwnerID != "user2" {
		t.Errorf("Expected current story to be from user2, got '%s'", current.OwnerID)
	}
}

func TestUser_Reset(t *testing.T) {
	user := NewUser("user1", "Alice")
	user.AFK = true
	user.KickVotes = []string{"user2", "user3"}
	user.EnqueueStory(NewStory("user2"))

	user.Reset()

	if user.AFK {
		t.Error("Expected AFK to be false after reset")
	}

	if len(user.KickVotes) != 0 {
		t.Errorf("Expected empty KickVotes after reset, got %d", len(user.KickVotes))
	}

	if len(user.StoryQueue) != 0 {
		t.Errorf("Expected empty StoryQueue after reset, got %d", len(user.StoryQueue))
	}
}

func TestGame_AddUser(t *testing.T) {
	game := NewGame(NewMockDatabase())
	state := &GameState{
		Users:     make(map[string]*User),
		UserOrder: make([]string, 0),
	}

	game.AddUser("user1", "Alice", state)

	if len(state.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(state.Users))
	}

	if len(state.UserOrder) != 1 {
		t.Errorf("Expected 1 user in order, got %d", len(state.UserOrder))
	}

	user := state.Users["user1"]
	if user == nil {
		t.Fatal("Expected user to exist")
	}

	if user.Name != "Alice" {
		t.Errorf("Expected user name to be 'Alice', got '%s'", user.Name)
	}
}

func TestGame_GetStories(t *testing.T) {
	game := NewGame(NewMockDatabase())
	state := &GameState{
		Users:     make(map[string]*User),
		UserOrder: make([]string, 0),
		Stories:   make([]*Story, 0),
	}

	// Add users
	game.AddUser("user1", "Alice", state)
	game.AddUser("user2", "Bob", state)

	// Create stories
	story1 := NewStory("user1")
	story1.AddText("Once upon a time")
	story1.AddText("there was a knight")

	story2 := NewStory("user2")
	story2.AddText("In a galaxy far away")

	state.Stories = append(state.Stories, story1, story2)

	// Get stories
	stories := game.GetStories(state)

	if len(stories) != 2 {
		t.Errorf("Expected 2 stories, got %d", len(stories))
	}

	if stories[0].Author != "Alice" {
		t.Errorf("Expected first story author to be 'Alice', got '%s'", stories[0].Author)
	}

	if stories[0].Text != "Once upon a time. there was a knight" {
		t.Errorf("Expected serialized text, got '%s'", stories[0].Text)
	}

	if stories[1].Author != "Bob" {
		t.Errorf("Expected second story author to be 'Bob', got '%s'", stories[1].Author)
	}
}

func TestGame_StartGame(t *testing.T) {
	game := NewGame(NewMockDatabase())
	state := &GameState{
		Users:        make(map[string]*User),
		UserOrder:    make([]string, 0),
		Started:      false,
		GameStatus:   "waiting",
		FinishVotes:  nil,
		RestartVotes: nil,
	}

	game.StartGame(state)

	if !state.Started {
		t.Error("Expected Started to be true")
	}

	if state.GameStatus != "started" {
		t.Errorf("Expected GameStatus to be 'started', got '%s'", state.GameStatus)
	}

	if state.FinishVotes == nil {
		t.Error("Expected FinishVotes to be initialized")
	}

	if state.RestartVotes == nil {
		t.Error("Expected RestartVotes to be initialized")
	}
}
