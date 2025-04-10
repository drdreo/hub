package dicegame

import (
	"encoding/json"
	"fmt"
	"gameserver/internal/client"
	"gameserver/internal/game"
	"gameserver/internal/room"
	"gameserver/internal/router"
	"testing"
)

// TestMessage represents the structure expected for game_action messages
type TestMessage struct {
	Type string `json:"type"`
	Data struct {
		Action    string `json:"action"`
		DiceIndex int    `json:"diceIndex"`
		EndTurn   bool   `json:"endTurn"`
	} `json:"data"`
}

func TestDiceGameIntegration(t *testing.T) {
	// Set up the complete system with real components
	registry := game.NewRegistry()
	RegisterDiceGame(registry)
	roomManager := room.NewRoomManager(registry)
	testRouter := router.NewRouter(roomManager, registry)

	// Create mock clients
	client1 := client.NewClientMock("player1")
	client2 := client.NewClientMock("player2")

	// Client1 creating a room
	testRouter.HandleMessage(client1, []byte(`{"type":"join_room","data":{"gameType":"dicegame","playerName":"tester-1"}}`))

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

	data, ok := createResponse.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Invalid data in response")
	}

	roomID, ok := data["roomId"].(string)
	if !ok || roomID == "" {
		t.Fatalf("Invalid or missing roomId in response")
	}

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

	// Set Player 1 as the current turn
	state := testRoom.State().(*GameState)
	state.CurrentTurn = client1.ID()
	testRoom.SetState(state)

	// Test rolling dice
	msg := TestMessage{Type: "roll"}
	msgBytes, _ := json.Marshal(msg)
	testRouter.HandleMessage(client1, msgBytes)

	// Verify both players received game update
	client1Messages = client1.GetSentMessages()
	if len(client1Messages) == 0 {
		t.Errorf("Player 1 didn't receive game state update after rolling")
	}

	client2Messages = client2.GetSentMessages()
	if len(client2Messages) == 0 {
		t.Errorf("Player 2 didn't receive game state update after rolling")
	}

	// Clear messages
	client1.ClearMessages()
	client2.ClearMessages()

	// Test setting aside dice
	// We'll use a fixed set of dice for testing
	state = testRoom.State().(*GameState)
	state.Dice = []int{1, 2, 3, 4, 5, 6} // Set specific dice values
	testRoom.SetState(state)

	msg = TestMessage{Type: "select"}
	msg.Data.DiceIndex = 0
	msgBytes, _ = json.Marshal(msg)
	testRouter.HandleMessage(client1, msgBytes)

	// Set aside the currently selected dice - first die (index 0)
	msg = TestMessage{Type: "set_aside"}
	msgBytes, _ = json.Marshal(msg)
	testRouter.HandleMessage(client1, msgBytes)

	// Verify state update
	state = testRoom.State().(*GameState)
	if len(state.SetAside) != 1 || state.SetAside[0] != 1 {
		t.Errorf("Expected dice to be set aside, got SetAside=%v", state.SetAside)
	}

	if len(state.Dice) != 5 {
		t.Errorf("Expected 5 dice remaining, got %d", len(state.Dice))
	}

	// End turn and verify turn switches to second player
	msg = TestMessage{Type: "end_turn"}
	msgBytes, _ = json.Marshal(msg)
	testRouter.HandleMessage(client1, msgBytes)

	state = testRoom.State().(*GameState)
	if state.CurrentTurn != client2.ID() {
		t.Errorf("Expected turn to switch to player 2, still on %s", state.CurrentTurn)
	}

	// Verify player score was updated
	if player1Score := state.Players[client1.ID()].Score; player1Score != 100 {
		t.Errorf("Expected player 1 to have 100 points (for setting aside a 1), got %d", player1Score)
	}
}
