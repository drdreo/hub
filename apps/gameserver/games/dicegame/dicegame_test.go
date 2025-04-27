package dicegame

import (
	"gameserver/internal/interfaces"
	"gameserver/internal/testicles"
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
	// Set up the test helper with all components
	helper := testicles.NewTestHelper(t)
	RegisterDiceGame(helper.Registry)

	// Setup the game room with two players
	playerIds := helper.SetupGameRoom("dicegame", 2)
	player1ID := playerIds[0]
	player2ID := playerIds[1]

	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	// Set Player 1 as the current turn
	state := testRoom.State().(*GameState)
	state.CurrentTurn = player1ID
	testRoom.SetState(state)

	// Test rolling dice
	helper.SendMessage(player1ID, "roll", nil)

	// Verify both players received game update
	helper.AssertClientsReceivedMessages([]string{player1ID, player2ID})

	helper.ClearAllMessages()

	// Test setting aside dice
	// We'll use a fixed set of dice for testing
	state = testRoom.State().(*GameState)
	state.Dice = []int{1, 1, 3, 3, 5, 5} // Set specific dice values
	testRoom.SetState(state)

	helper.SendMessage(player1ID, "select", interfaces.M{
		"diceIndex": 0,
	})

	// Set aside the currently selected dice - first die (index 0)
	helper.SendMessage(player1ID, "set_aside", interfaces.M{
		"endTurn": false,
	})

	// Verify state update
	state = testRoom.State().(*GameState)
	if len(state.SetAside) != 1 || state.SetAside[0] != 1 {
		t.Errorf("Expected dice to be set aside, got SetAside=%v", state.SetAside)
	}

	if len(state.Dice) != 5 {
		t.Errorf("Expected 5 dice remaining, got %d", len(state.Dice))
	}

	state.Dice = []int{1, 3, 3, 5, 5} // Set specific dice values
	testRoom.SetState(state)

	helper.SendMessage(player1ID, "select", interfaces.M{
		"diceIndex": 0,
	})

	// Set aside the currently selected dice - first die (index 0)
	helper.SendMessage(player1ID, "set_aside", interfaces.M{
		"endTurn": true,
	})

	state = testRoom.State().(*GameState)

	if state.CurrentTurn != player2ID {
		t.Errorf("Expected turn to switch to player 2, still on %s", state.CurrentTurn)
	}

	// Verify player score was updated
	if player1Score := state.Players[player1ID].Score; player1Score != 200 {
		t.Errorf("Expected player 1 to have 200 points (for setting aside a 2x 1), got %d", player1Score)
	}
}
