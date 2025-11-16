package tell_it

import "testing"

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