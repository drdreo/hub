package room

import (
	"context"
	testgame "gameserver/games/test"
	"gameserver/internal/game"
	"gameserver/internal/interfaces"
	"sync"
	"testing"
)

func TestRoomManager(t *testing.T) {
	registry := game.NewRegistry()
	testgame.RegisterTestGame(registry)
	manager := NewRoomManager(registry, nil)

	testCtx := context.Background()
	t.Run("create and get room by ID", func(t *testing.T) {
		room, err := manager.CreateRoom(testCtx, interfaces.CreateRoomOptions{
			GameType: "testGame",
		})

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
	})

	t.Run("create room with custom ID", func(t *testing.T) {
		roomId := "test-room-1"
		room, err := manager.CreateRoom(testCtx, interfaces.CreateRoomOptions{
			GameType: "testGame",
			RoomID:   &roomId,
		})
		if err != nil {
			t.Errorf("unexpected error creating room: %v", err)
		}
		if room == nil {
			t.Error("expected room to not be nil")
		}
		if room.ID() != roomId {
			t.Errorf("expected room ID to be %s, got %s", roomId, room.ID())
		}
	})

	t.Run("create room with empty ID", func(t *testing.T) {
		emptyID := ""
		room, err := manager.CreateRoom(testCtx, interfaces.CreateRoomOptions{
			GameType: "testGame",
			RoomID:   &emptyID,
		})
		if err != nil {
			t.Errorf("unexpected error creating room: %v", err)
		}
		if room == nil {
			t.Error("expected room to not be nil")
		}
		if room.ID() == "" {
			t.Error("expected room ID to not be empty")
		}
	})

	t.Run("get non-existent room", func(t *testing.T) {
		room, err := manager.GetRoom("non-existent-id")
		if err == nil {
			t.Error("expected error getting non-existent room, got nil")
		}
		if room != nil {
			t.Error("expected room to be nil")
		}
	})

	t.Run("remove non-existent room", func(t *testing.T) {
		// Should not panic
		manager.RemoveRoom("non-existent-id")
	})

	t.Run("cleanup empty rooms", func(t *testing.T) {
		room, err := manager.CreateRoom(testCtx, interfaces.CreateRoomOptions{
			GameType: "testGame",
		})
		if err != nil {
			t.Errorf("unexpected error creating room: %v", err)
		}
		roomID := room.ID()

		_, err = manager.GetRoom(roomID)
		if err != nil {
			t.Errorf("unexpected error getting room: %v", err)
		}

		// Cleanup should remove empty room
		manager.Cleanup()
		_, err = manager.GetRoom(roomID)
		if err == nil {
			t.Error("expected error getting removed room, got nil")
		}
	})

	t.Run("create room with invalid game type", func(t *testing.T) {
		room, err := manager.CreateRoom(testCtx, interfaces.CreateRoomOptions{
			GameType: "invalid-game",
		})
		if err == nil {
			t.Error("expected error creating room with invalid game type, got nil")
		}
		if room != nil {
			t.Error("expected room to be nil")
		}
	})

	t.Run("concurrent room operations", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := manager.CreateRoom(testCtx, interfaces.CreateRoomOptions{
					GameType: "testGame",
				})
				if err != nil {
					t.Errorf("unexpected error in concurrent room creation: %v", err)
				}
			}()
		}
		wg.Wait()

		// Verify all rooms were created
		rooms := manager.ListRooms()
		if len(rooms) != 10 {
			t.Errorf("expected 10 rooms, got %d", len(rooms))
		}
	})

}
