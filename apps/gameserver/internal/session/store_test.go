package session

import "testing"

func TestSessionStore(t *testing.T) {
    store := NewStore(10) // 10-second expiry for quick testing

    // Test storing and retrieving session data
    data := SessionData{ClientID: "test1", RoomID: "room1"}
    store.StoreSession("test1", data)

    retrieved, exists := store.GetSession("test1")
    if exists != true {
        t.Errorf(`Session is not set`)
    }
    if retrieved.RoomID != "room1" {
        t.Errorf(`Wrong session data retrieved`)
    }
}

// TODO: Test session expiry
