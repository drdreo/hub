package tell_it

import (
	"gameserver/games/tell_it/database"
	"gameserver/internal/protocol"
	"gameserver/internal/testicles"
	"testing"
)

func TestGame_SubmitText_CreateNewStory(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 2)
	player1ID := playerIds[0]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	if len(state.Stories) != 0 {
		t.Errorf("Expected 0 stories initially, got %d", len(state.Stories))
	}

	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story that is."})

	state = room.State().(*GameState)
	if len(state.Stories) != 1 {
		t.Errorf("Expected 1 story after first submit, got %d", len(state.Stories))
	}
}

// Test: should create multiple stories (one per user)
func TestGame_SubmitText_CreateMultipleStories(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 3)
	player1ID := playerIds[0]
	player2ID := playerIds[1]
	player3ID := playerIds[2]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	if len(state.Stories) != 0 {
		t.Errorf("Expected 0 stories initially, got %d", len(state.Stories))
	}

	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story that is."})
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "First story that is."})
	helper.SendMessage(player3ID, "submit_text", map[string]string{"text": "First story that is."})

	state = room.State().(*GameState)
	if len(state.Stories) != 3 {
		t.Errorf("Expected 3 stories after all users submit, got %d", len(state.Stories))
	}
}

// Test: should create 1 story per user (with 2 users)
func TestGame_SubmitText_OneStoryPerUser(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 2)
	player1ID := playerIds[0]
	player2ID := playerIds[1]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	if len(state.Stories) != 0 {
		t.Errorf("Expected 0 stories initially, got %d", len(state.Stories))
	}

	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story that is."})
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "First story that is."})

	state = room.State().(*GameState)
	if len(state.Stories) != 2 {
		t.Errorf("Expected 2 stories, got %d", len(state.Stories))
	}
}

// Test: should return error if user can't submit new text (no story queued)
func TestGame_SubmitText_ErrorWhenNoStoryQueued(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 2)
	player1ID := playerIds[0]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story that is."})

	helper.ClearAllMessages()

	// Second submit by same user should send error (no story queued yet)
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "Second story that is."})

	// Verify error message was sent back to client
	client := helper.Clients[player1ID]
	messages := client.GetSentMessages()

	hasError := false
	for _, msg := range messages {
		if !msg.Success && msg.Type == "submit_text" {
			hasError = true
			break
		}
	}

	if !hasError {
		t.Error("Expected error response when user tries to submit without queued story")
	}
}

func TestGame_SubmitText_CircleStories(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 3)
	player1ID := playerIds[0]
	player2ID := playerIds[1]
	player3ID := playerIds[2]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	// First round - everyone creates a story
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story that is."})
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "First story that is."})
	helper.SendMessage(player3ID, "submit_text", map[string]string{"text": "First story that is."})

	state = room.State().(*GameState)

	// Check that user1 has user3's story
	story := state.Users[player1ID].GetCurrentStory()
	if story == nil {
		t.Fatal("Expected user1 to have a story")
	}
	if story.OwnerID != player3ID {
		t.Errorf("Expected user1 to have user3's story, got story from %s", story.OwnerID)
	}

	// Check user1 has 1 story in queue
	if len(state.Users[player1ID].StoryQueue) != 1 {
		t.Errorf("Expected user1 to have 1 story queued, got %d", len(state.Users[player1ID].StoryQueue))
	}

	// Second round
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "Second story that is 1"})
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "Second story that is 2"})
	helper.SendMessage(player3ID, "submit_text", map[string]string{"text": "Second story that is 3"})

	// Third round
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "Third story that is 1"})
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "Third story that is 2"})
	helper.SendMessage(player3ID, "submit_text", map[string]string{"text": "Third story that is 3"})

	state = room.State().(*GameState)

	// After 3 rounds, user1 should have their own story back
	story = state.Users[player1ID].GetCurrentStory()
	if story == nil {
		t.Fatal("Expected user1 to have a story after 3 rounds")
	}
	if story.OwnerID != player1ID {
		t.Errorf("Expected user1 to have their own story back, got story from %s", story.OwnerID)
	}

	// Check user1 still has 1 story in queue
	if len(state.Users[player1ID].StoryQueue) != 1 {
		t.Errorf("Expected user1 to have 1 story queued, got %d", len(state.Users[player1ID].StoryQueue))
	}
}

func TestGame_SubmitText_NoStoryAfterSubmitting(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 2)
	player1ID := playerIds[0]
	player2ID := playerIds[1]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	// First round
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story of 1"})
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "First story of 2"})

	// Second round - user1 submits
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "Second story of 1"})

	state = room.State().(*GameState)

	// User1 should have no story now (submitted and no new one queued yet)
	story := state.Users[player1ID].GetCurrentStory()
	if story != nil {
		t.Errorf("Expected user1 to have no story after submitting, but got story from %s", story.OwnerID)
	}
}

// Test: should not switch stories during a round
func TestGame_SubmitText_NoStorySwitching(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 2)
	player1ID := playerIds[0]
	player2ID := playerIds[1]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	// First round
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story of 1"})
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "First story of 2"})

	state = room.State().(*GameState)

	// User1 should have user2's story
	story := state.Users[player1ID].GetCurrentStory()
	if story == nil {
		t.Fatal("Expected user1 to have a story")
	}
	if story.GetLatestText() != "First story of 2" {
		t.Errorf("Expected user1 to have 'First story of 2', got '%s'", story.GetLatestText())
	}

	// User2 should have user1's story
	story = state.Users[player2ID].GetCurrentStory()
	if story == nil {
		t.Fatal("Expected user2 to have a story")
	}
	if story.GetLatestText() != "First story of 1" {
		t.Errorf("Expected user2 to have 'First story of 1', got '%s'", story.GetLatestText())
	}

	// User2 submits second round
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "Second story of 2"})

	state = room.State().(*GameState)

	// User1 should have 2 stories
	if len(state.Users[player1ID].StoryQueue) != 2 {
		t.Errorf("Expected user1 to have 2 stories, got %d", len(state.Users[player1ID].StoryQueue))
	}

	// User1's story should remain the same (not switched)
	story = state.Users[player1ID].GetCurrentStory()
	if story == nil {
		t.Fatal("Expected user1 to still have a story")
	}
	if story.GetLatestText() != "First story of 2" {
		t.Errorf("Expected user1 to have stil have same story with 'First story of 2', got '%s'", story.GetLatestText())
	}

	// User2 should have no story now (already submitted)
	story = state.Users[player2ID].GetCurrentStory()
	if story != nil {
		t.Errorf("Expected user2 to have no story after submitting, but got story")
	}
}

// Test: should queue multiple stories when users are at different paces
func TestGame_SubmitText_MultipleStoriesQueued(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 3)
	player1ID := playerIds[0]
	player2ID := playerIds[1]
	player3ID := playerIds[2]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	// First round
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story of 1"})
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "First story of 2"})
	helper.SendMessage(player3ID, "submit_text", map[string]string{"text": "First story of 3"})

	state = room.State().(*GameState)

	// User1 should have 1 story queued
	if len(state.Users[player1ID].StoryQueue) != 1 {
		t.Errorf("Expected user1 to have 1 story queued, got %d", len(state.Users[player1ID].StoryQueue))
	}

	// Second round - user2 and user3 submit (user1 doesn't)
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "Second story of 2"})
	helper.SendMessage(player3ID, "submit_text", map[string]string{"text": "Second story of 3"})

	state = room.State().(*GameState)

	// User1 should now have 2 stories queued
	if len(state.Users[player1ID].StoryQueue) != 2 {
		t.Errorf("Expected user1 to have 2 stories queued, got %d", len(state.Users[player1ID].StoryQueue))
	}

	// Third round - user3 submits again (user1 still hasn't)
	helper.SendMessage(player3ID, "submit_text", map[string]string{"text": "Third story of 3"})

	state = room.State().(*GameState)

	// User1 should now have 3 stories queued
	if len(state.Users[player1ID].StoryQueue) != 3 {
		t.Errorf("Expected user1 to have 3 stories queued, got %d", len(state.Users[player1ID].StoryQueue))
	}
}

// Test: story updates should NOT be sent to non-owners in first round
func TestGame_SubmitText_NoStoryUpdateForNonOwnersFirstRound(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 2)
	player1ID := playerIds[0]
	player2ID := playerIds[1]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	helper.ClearAllMessages()

	// User1 submits first story
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story from user1"})

	// User2 should NOT receive a story_update because they haven't submitted yet
	client2 := helper.Clients[player2ID]
	messages := client2.GetSentMessages()

	hasStoryUpdate := false
	for _, msg := range messages {
		if msg.Type == "story_update" {
			hasStoryUpdate = true
			break
		}
	}

	if hasStoryUpdate {
		t.Error("User2 should NOT receive story_update in first round before they've submitted")
	}

	// Verify the story is queued for user2
	state = room.State().(*GameState)
	if len(state.Users[player2ID].StoryQueue) != 1 {
		t.Errorf("Expected user2 to have 1 story queued, got %d", len(state.Users[player2ID].StoryQueue))
	}
}

// Test: story updates SHOULD be sent to owners in subsequent rounds
func TestGame_SubmitText_StoryUpdateForOwnersAfterFirstRound(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 2)
	player1ID := playerIds[0]
	player2ID := playerIds[1]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	// First round - both users submit
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "First story from user1"})
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "First story from user2"})

	helper.ClearAllMessages()

	// Second round - user1 submits
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "Second story from user1"})

	// User2 SHOULD receive a story_update because they are now a story owner
	client2 := helper.Clients[player2ID]
	messages := client2.GetSentMessages()

	hasStoryUpdate := false
	for _, msg := range messages {
		if msg.Type == "story_update" {
			hasStoryUpdate = true
			break
		}
	}

	if !hasStoryUpdate {
		t.Error("User2 should receive story_update in second round since they're a story owner")
	}
}

// Test: multiple users should get updates only after they become owners
func TestGame_SubmitText_MultipleUsersUpdateLogic(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbMock := &database.DatabaseServiceMock{}
	g := NewGame(dbMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("tellit", 3)
	player1ID := playerIds[0]
	player2ID := playerIds[1]
	player3ID := playerIds[2]

	room, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := room.State().(*GameState)
	state.GameStatus = GameStatusStarted
	room.SetState(state)

	helper.ClearAllMessages()

	// User1 submits - no one should get story_update (no one is owner yet)
	helper.SendMessage(player1ID, "submit_text", map[string]string{"text": "Story 1"})

	client2 := helper.Clients[player2ID]
	client3 := helper.Clients[player3ID]

	if hasStoryUpdateMessage(client2.GetSentMessages()) {
		t.Error("User2 should not receive story_update before becoming owner")
	}
	if hasStoryUpdateMessage(client3.GetSentMessages()) {
		t.Error("User3 should not receive story_update before becoming owner")
	}

	helper.ClearAllMessages()

	// User2 submits - only user3 gets the story (user1 already has one queued)
	helper.SendMessage(player2ID, "submit_text", map[string]string{"text": "Story 2"})

	if hasStoryUpdateMessage(client3.GetSentMessages()) {
		t.Error("User3 should not receive story_update before becoming owner")
	}

	helper.ClearAllMessages()

	// User3 submits - now user1 should get an update (user1 is an owner and has queued story)
	helper.SendMessage(player3ID, "submit_text", map[string]string{"text": "Story 3"})

	client1 := helper.Clients[player1ID]
	if !hasStoryUpdateMessage(client1.GetSentMessages()) {
		t.Error("User1 should receive story_update after round completes since they're an owner")
	}
}

// Helper function to check if messages contain story_update
func hasStoryUpdateMessage(messages []*protocol.Response) bool {
	for _, msg := range messages {
		if msg.Type == "story_update" {
			return true
		}
	}
	return false
}
