package session

import (
	"testing"
)

func TestSessionStore(t *testing.T) {
	store := NewStore(2) // 10-second expiry for quick testicles

	t.Run("store and retrieve session", func(t *testing.T) {
		// Test storing and retrieving session data
		data := SessionData{ClientID: "test1", RoomID: "room1"}
		store.StoreSession("test1", data)

		retrieved, exists := store.GetSession("test1")
		if exists != true {
			t.Error("session is not set")
		}
		if retrieved.RoomID != "room1" {
			t.Errorf("wrong session data retrieved, expected 'room1', got '%s'", retrieved.RoomID)
		}
	})

	t.Run("get non-existent session", func(t *testing.T) {
		_, exists := store.GetSession("non-existent")
		if exists {
			t.Error("expected non-existent session to not exist")
		}
	})

	t.Run("overwrite existing session", func(t *testing.T) {
		// Store initial session
		initialData := SessionData{ClientID: "test2", RoomID: "room1"}
		store.StoreSession("test2", initialData)

		// Overwrite with new data
		newData := SessionData{ClientID: "test2", RoomID: "room2"}
		store.StoreSession("test2", newData)

		// Verify new data
		retrieved, exists := store.GetSession("test2")
		if !exists {
			t.Error("session should exist after overwrite")
		}
		if retrieved.RoomID != "room2" {
			t.Errorf("wrong session data retrieved after overwrite, expected 'room2', got '%s'", retrieved.RoomID)
		}
	})

	// t.Run("session expiry", func(t *testicles.T) {
	// 	// Store session
	// 	data := SessionData{ClientID: "test3", RoomID: "room1"}
	// 	store.StoreSession("test3", data)

	// 	// Wait for expiry (2 seconds)
	// 	time.Sleep(3 * time.Second)

	// 	// Verify session is expired
	// 	_, exists := store.GetSession("test3")
	// 	if exists {
	// 		t.Error("session should be expired")
	// 	}
	// })
}
