package owe_drahn

import (
	"gameserver/games/owe_drahn/database"
	"gameserver/internal/interfaces"
	"gameserver/internal/testicles"
	"testing"
)

func TestOweDrahnIntegration(t *testing.T) {
	// Set up the test helper with all components
	helper := testicles.NewTestHelper(t)

	// Set up mock DB and register the game
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	helper.RegisterGame(g)

	// Setup the game room with two players
	playerIds := helper.SetupGameRoom("owedrahn", 2)
	player1ID := playerIds[0]
	player2ID := playerIds[1]

	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	// Both players set ready to start the game
	helper.SendMessage(player1ID, "ready", true)
	helper.SendMessage(player2ID, "ready", true)

	// Check if game started
	state := testRoom.State().(*GameState)
	if !state.Started {
		t.Fatalf("Game should have started after both players were ready")
	}

	helper.ClearAllMessages()

	// Set Player 1 as the current turn for testing
	state.CurrentTurn = helper.Clients[player1ID].ID()
	testRoom.SetState(state)

	// Test rolling dice
	helper.SendMessage(player1ID, "roll", nil)
	helper.AssertClientsReceivedMessages([]string{player1ID, player2ID})

	// Check if the dice roll event was broadcasted
	if !helper.VerifyMessageReceived(player1ID, "rolledDice") {
		t.Errorf("rolledDice event not found in messages")
	}

	helper.ClearAllMessages()

	state = testRoom.State().(*GameState)

	// Test losing life functionality
	// set current value to a value that would kill the player
	state.CurrentValue = 15
	state.CurrentTurn = helper.Clients[player1ID].ID()
	testRoom.SetState(state)

	helper.SendMessage(player1ID, "loseLife", nil)

	// Verify state update
	state = testRoom.State().(*GameState)
	player1 := state.Players[helper.Clients[player1ID].ID()]

	if player1.Life != 5 {
		t.Errorf("Expected player 1 to have 5 lives after losing one, got %d", player1.Life)
	}

	if state.CurrentValue != 0 {
		t.Errorf("Expected current value to be reset to 0, got %d", state.CurrentValue)
	}

	if !player1.IsChoosing {
		t.Errorf("Player should be in choosing state after losing life")
	}

	if !helper.VerifyMessageReceived(player1ID, "lostLife") {
		t.Errorf("lostLife event not found in messages")
	}

	// Test choosing next player
	helper.ClearAllMessages()

	helper.SendMessage(player1ID, "chooseNextPlayer", interfaces.M{
		"nextPlayerId": helper.Clients[player2ID].ID(),
	})

	state = testRoom.State().(*GameState)

	if state.CurrentTurn != helper.Clients[player2ID].ID() {
		t.Errorf("Expected current turn to switch to player 2, still on %s", state.CurrentTurn)
	}

	if player1.IsChoosing {
		t.Errorf("Player 1 should no longer be in choosing state")
	}
}
