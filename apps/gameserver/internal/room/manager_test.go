package room

import (
	"gameserver/games/test"
	"gameserver/internal/game"
	"testing"
)

func TestRoomManager(t *testing.T) {
	registry := game.NewRegistry()
	testgame.RegisterTestGame(registry)
	manager := NewRoomManager(registry)

	// Test room creation
	room, err := manager.CreateRoom("testGame", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if room.ID() == "" {
		t.Error("expected non-empty room ID")
	}

	// Test room retrieval
	found, err := manager.GetRoom(room.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.ID() != room.ID() {
		t.Errorf("expected ID %s, got %s", room.ID(), found.ID())
	}

	_, err = manager.GetRoom("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent room, got nil")
	}
}
