package main

import (
	"fmt"
	"gameserver/games/tictactoe"
	"gameserver/internal/client"
	"gameserver/internal/game"
	"gameserver/internal/room"
	"gameserver/internal/router"
	"testing"
)

func TestGameFlowIntegration(t *testing.T) {
	// Set up the complete system with real components
	registry := game.NewRegistry()
	tictactoe.RegisterTicTacToeGame(registry)
	roomManager := room.NewRoomManager(registry)
	testRouter := router.NewRouter(roomManager, registry)

	// Create mock clients
	client1 := client.NewClientMock("player1")
	client2 := client.NewClientMock("player2")

	// Client1 creating a room
	testRouter.HandleMessage(client1, []byte(`{"type":"join_room","data":{"gameType":"tictactoe", "playerName": "tester-1"}}`))

	messages := client1.GetSentMessages()
	if len(messages) == 0 {
		t.Fatalf("No messages received after room creation")
	}

	if messages[0].Success == false {
		t.Errorf("Expected join_room of tester-1 to be sucessfull")
	}

	// Extract room ID from response
	createResponse := messages[len(messages)-1]
	if createResponse.Success != true {
		t.Fatalf("createResponse was not successful")
	}

	if createResponse.Type != "join_room_result" {
		t.Fatalf("Expected 'join_room_result' message, got: %v", createResponse.Type)
	}

	data, ok := createResponse.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Invalid data in response")
	}

	responseGameType, ok := data["gameType"].(string)
	if !ok || responseGameType != "tictactoe" {
		t.Fatalf("Expected 'tictactoe' game type in response")
	}

	roomID, ok := data["roomId"].(string)
	if !ok || roomID == "" {
		t.Fatalf("Invalid or missing roomId in response")
	}

	// Clear messages before next step
	client1.ClearMessages()
	client2.ClearMessages()

	// Second player joins the room
	joinMessage := fmt.Sprintf(`{"type":"join_room","data":{"roomId":"%s", "playerName": "tester-2"}}`, roomID)
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

	if client2Messages[0].Success == false {
		t.Errorf("Expected join_room of player2 to be sucessfull")
	}

	// Clear messages before game moves
	client1.ClearMessages()
	client2.ClearMessages()

	// GAME ACTIONS ----------------------

	// Since first player is random in tictactoe, reset to client1
	state := client1.Room().State().(tictactoe.GameState)
	state.CurrentTurn = client1.ID()
	client1.Room().SetState(state)

	// Make game moves
	testRouter.HandleMessage(client1, []byte(`{"type":"make_move","data":{"action":"make_move","position":0}}`))

	// Verify both players received game update
	if len(client1.GetSentMessages()) == 0 {
		t.Errorf("Player 1 didn't receive game state update after move")
	}

	if len(client2.GetSentMessages()) == 0 {
		t.Errorf("Player 2 didn't receive game state update after move")
	}

	// Verify game state
	testRoom, err := roomManager.GetRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state = testRoom.State().(tictactoe.GameState)
	board := state.Board
	if board[0][0] != "X" {
		t.Errorf("Expected 'X' at position 0, got %v", board[0])
	}
}
