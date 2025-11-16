package tell_it

import "testing"

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
	} else if current.OwnerID != "user2" {
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