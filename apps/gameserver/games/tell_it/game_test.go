package tell_it

import (
	"testing"
)

func newTestGame() *Game {
	return &Game{
		dbService: nil, // For these tests, we don't need the database service
	}
}

func TestGame_AddUser(t *testing.T) {
	game := newTestGame()
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
	game := newTestGame()
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
	game := newTestGame()
	state := &GameState{
		Users:        make(map[string]*User),
		UserOrder:    make([]string, 0),
		Started:      false,
		GameStatus:   GameStatusWaiting,
		FinishVotes:  nil,
		RestartVotes: nil,
	}

	game.StartGame(state)

	if !state.Started {
		t.Error("Expected Started to be true")
	}

	if state.GameStatus != GameStatusStarted {
		t.Errorf("Expected GameStatus to be 'started', got '%s'", state.GameStatus)
	}

	if state.FinishVotes == nil {
		t.Error("Expected FinishVotes to be initialized")
	}

	if state.RestartVotes == nil {
		t.Error("Expected RestartVotes to be initialized")
	}
}

func TestGame_IsUserStoryOwner(t *testing.T) {
	game := newTestGame()
	state := &GameState{
		Users:     make(map[string]*User),
		UserOrder: make([]string, 0),
		Stories:   make([]*Story, 0),
	}

	// Add users
	game.AddUser("user1", "Alice", state)
	game.AddUser("user2", "Bob", state)

	// Initially, no one is a story owner
	if game.isUserStoryOwner("user1", state) {
		t.Error("Expected user1 to not be a story owner initially")
	}

	// Create stories
	story1 := NewStory("user1")
	story1.AddText("Once upon a time")
	story1.AddText("there was a knight")

	state.Stories = append(state.Stories, story1)

	if !game.isUserStoryOwner("user1", state) {
		t.Error("Expected user1 to be a story owner after submitting")
	}

	// User2 should not be a story owner yet
	if game.isUserStoryOwner("user2", state) {
		t.Error("Expected user2 to not be a story owner yet")
	}

	story2 := NewStory("user2")
	story2.AddText("In a galaxy far away")

	state.Stories = append(state.Stories, story2)

	if !game.isUserStoryOwner("user2", state) {
		t.Error("Expected user2 to be a story owner after submitting")
	}
}

func TestGame_ReconnectUser(t *testing.T) {
	game := newTestGame()
	state := &GameState{
		Users:        make(map[string]*User),
		UserOrder:    make([]string, 0),
		Stories:      make([]*Story, 0),
		FinishVotes:  make(map[string]bool),
		RestartVotes: make(map[string]bool),
	}

	// Add users
	game.AddUser("oldID1", "Alice", state)
	game.AddUser("user2", "Bob", state)
	game.AddUser("user3", "Charlie", state)

	// Set up some game state
	story := NewStory("oldID1")
	story.AddText("test story")
	state.Stories = append(state.Stories, story)
	state.Users["oldID1"].Disconnected = true
	state.FinishVotes["oldID1"] = true
	state.RestartVotes["oldID1"] = true

	// Add kick votes from oldID1
	state.Users["user2"].KickVotes = append(state.Users["user2"].KickVotes, "oldID1")

	// Reconnect the user with a new ID
	err := game.ReconnectUser("oldID1", "newID1", state)
	if err != nil {
		t.Fatalf("ReconnectUser failed: %v", err)
	}

	// Verify old ID is removed from Users map
	if _, exists := state.Users["oldID1"]; exists {
		t.Error("Expected old ID to be removed from Users map")
	}

	// Verify new ID exists in Users map
	user, exists := state.Users["newID1"]
	if !exists {
		t.Fatal("Expected new ID to exist in Users map")
	}

	// Verify user data is preserved
	if user.Name != "Alice" {
		t.Errorf("Expected user name to be 'Alice', got '%s'", user.Name)
	}

	// Verify disconnected status is cleared
	if user.Disconnected {
		t.Error("Expected Disconnected to be false after reconnect")
	}

	// Verify user ID is updated
	if user.ID != "newID1" {
		t.Errorf("Expected user ID to be 'newID1', got '%s'", user.ID)
	}

	// Verify UserOrder is updated
	found := false
	for _, id := range state.UserOrder {
		if id == "newID1" {
			found = true
		}
		if id == "oldID1" {
			t.Error("Old ID should not be in UserOrder")
		}
	}
	if !found {
		t.Error("New ID should be in UserOrder")
	}

	// Verify FinishVotes is updated
	if _, exists := state.FinishVotes["oldID1"]; exists {
		t.Error("Old ID should not be in FinishVotes")
	}
	if !state.FinishVotes["newID1"] {
		t.Error("New ID should be in FinishVotes")
	}

	// Verify RestartVotes is updated
	if _, exists := state.RestartVotes["oldID1"]; exists {
		t.Error("Old ID should not be in RestartVotes")
	}
	if !state.RestartVotes["newID1"] {
		t.Error("New ID should be in RestartVotes")
	}

	// Verify KickVotes are updated
	found = false
	for _, voterID := range state.Users["user2"].KickVotes {
		if voterID == "newID1" {
			found = true
		}
		if voterID == "oldID1" {
			t.Error("Old ID should not be in KickVotes")
		}
	}
	if !found {
		t.Error("New ID should be in KickVotes")
	}

	// Verify Stories are updated
	found = false
	for _, story := range state.Stories {
		if story.OwnerID == "newID1" {
			found = true
			if story.Texts[0] != "test story" {
				t.Error("Story content should remain the same")
			}
		}
		if story.OwnerID == "oldID1" {
			t.Error("Old ID should not be in Stories")
		}
	}
	if !found {
		t.Error("New ID should be in Stories")
	}

	// Test error case: user not found
	err = game.ReconnectUser("nonexistent", "newID", state)
	if err == nil {
		t.Error("Expected error when reconnecting nonexistent user")
	}
}
