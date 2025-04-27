package owe_drahn

import (
	"context"
	"encoding/json"
	"fmt"
	"gameserver/games/owe_drahn/database"
	"gameserver/internal/client"
	"gameserver/internal/game"
	"gameserver/internal/room"
	"gameserver/internal/router"
	"testing"
)

// TestMessage represents the structure for game messages
type TestMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func TestOweDrahnIntegration(t *testing.T) {
	testCtx := context.Background()

	// Set up the complete system with real components and mock DB
	registry := game.NewRegistry()
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	registry.RegisterGame(g)
	clientManager := client.NewManager()
	roomManager := room.NewRoomManager(registry)
	testRouter := router.NewRouter(testCtx, clientManager, roomManager, registry)

	// Create mock clients
	client1 := client.NewClientMock("player1")
	client2 := client.NewClientMock("player2")

	// Client1 creating a room
	testRouter.HandleMessage(client1, []byte(`{"type":"join_room","data":{"gameType":"owedrahn","playerName":"tester-1"}}`))

	messages := client1.GetSentMessages()
	if len(messages) == 0 {
		t.Fatalf("No messages received after room creation")
	}

	// Extract room ID from response
	createResponse := messages[len(messages)-1]
	if createResponse.Success != true {
		t.Fatalf("createResponse was not successful")
	}

	if createResponse.Type != "join_room_result" {
		t.Fatalf("Expected 'join_room_result' message, got: %v", createResponse.Type)
	}

	data, ok := createResponse.Data.(*router.JoinResponse)
	if !ok {
		t.Fatalf("Invalid data in response")
	}

	roomID := data.RoomID

	// Clear messages before next step
	client1.ClearMessages()
	client2.ClearMessages()

	// Second player joins the room
	joinMessage := fmt.Sprintf(`{"type":"join_room","data":{"roomId":"%s", "playerName":"tester-2"}}`, roomID)
	testRouter.HandleMessage(client2, []byte(joinMessage))

	// Verify both clients received appropriate messages
	client1Messages := client1.GetSentMessages()
	if len(client1Messages) == 0 {
		t.Errorf("Player 1 didn't receive notification about player 2 joining")
	}

	client2Messages := client2.GetSentMessages()
	if len(client2Messages) == 0 {
		t.Errorf("Player 2 didn't receive join confirmation")
	}

	// Clear messages before game actions
	client1.ClearMessages()
	client2.ClearMessages()

	// Get the game room
	testRoom, err := roomManager.GetRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	// Both players set ready to start the game
	readyMessage := `{"type":"ready","data":true}`
	testRouter.HandleMessage(client1, []byte(readyMessage))
	testRouter.HandleMessage(client2, []byte(readyMessage))

	// Check if game started
	state := testRoom.State().(*GameState)
	if !state.Started {
		t.Fatalf("Game should have started after both players were ready")
	}

	// Clear messages
	client1.ClearMessages()
	client2.ClearMessages()

	// Set Player 1 as the current turn for testing
	state.CurrentTurn = client1.ID()
	testRoom.SetState(state)

	// Test rolling dice
	testRouter.HandleMessage(client1, []byte(`{"type":"roll"}`))

	// Verify both players received game update
	client1Messages = client1.GetSentMessages()
	if len(client1Messages) == 0 {
		t.Errorf("Player 1 didn't receive game state update after rolling")
	}

	client2Messages = client2.GetSentMessages()
	if len(client2Messages) == 0 {
		t.Errorf("Player 2 didn't receive game state update after rolling")
	}

	// Check if the dice roll event was broadcasted
	foundRollEvent := false
	for _, msg := range client1Messages {
		if msg.Type == "rolledDice" {
			foundRollEvent = true
			break
		}
	}
	if !foundRollEvent {
		t.Errorf("rolledDice event not found in messages")
	}

	// Clear messages
	client1.ClearMessages()
	client2.ClearMessages()

	// Get updated state
	state = testRoom.State().(*GameState)

	// Test losing life functionality
	// set current value to a value that would kill the player
	state.CurrentValue = 15
	state.CurrentTurn = client1.ID()
	testRoom.SetState(state)

	testRouter.HandleMessage(client1, []byte(`{"type":"loseLife"}`))

	// Verify state update
	state = testRoom.State().(*GameState)
	player1 := state.Players[client1.ID()]

	if player1.Life != 5 {
		t.Errorf("Expected player 1 to have 5 lives after losing one, got %d", player1.Life)
	}

	if state.CurrentValue != 0 {
		t.Errorf("Expected current value to be reset to 0, got %d", state.CurrentValue)
	}

	if !player1.IsChoosing {
		t.Errorf("Player should be in choosing state after losing life")
	}

	foundLostLifeEvent := false
	client1Messages = client1.GetSentMessages()
	for _, msg := range client1Messages {
		if msg.Type == "lostLife" {
			foundLostLifeEvent = true
			break
		}
	}
	if !foundLostLifeEvent {
		t.Errorf("lostLife event not found in messages")
	}

	// Test choosing next player
	client1.ClearMessages()
	client2.ClearMessages()

	chooseNextMsg := fmt.Sprintf(`{"type":"chooseNextPlayer","data":{"nextPlayerId":"%s"}}`, client2.ID())
	testRouter.HandleMessage(client1, []byte(chooseNextMsg))

	state = testRoom.State().(*GameState)

	if state.CurrentTurn != client2.ID() {
		t.Errorf("Expected current turn to switch to player 2, still on %s", state.CurrentTurn)
	}

	if player1.IsChoosing {
		t.Errorf("Player 1 should no longer be in choosing state")
	}
}
