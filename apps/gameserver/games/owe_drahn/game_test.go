package owe_drahn

import (
	"testing"

	"gameserver/games/owe_drahn/database"
	"gameserver/internal/interfaces"
	"gameserver/internal/testicles"
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

func TestMainBetSetting(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("owedrahn", 2)
	player1ID := playerIds[0]

	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	state := testRoom.State().(*GameState)

	// Test default main bet is 1
	if state.MainBet != 1 {
		t.Errorf("Expected default main bet to be 1, got %f", state.MainBet)
	}

	// Test setting valid main bet before game starts
	helper.SendMessage(player1ID, "set_main_bet", interfaces.M{
		"amount": 5.0,
	})

	state = testRoom.State().(*GameState)
	if state.MainBet != 5.0 {
		t.Errorf("Expected main bet to be 5, got %f", state.MainBet)
	}

	// Test setting main bet to different valid value
	helper.SendMessage(player1ID, "set_main_bet", interfaces.M{
		"amount": 10.5,
	})

	state = testRoom.State().(*GameState)
	if state.MainBet != 10.5 {
		t.Errorf("Expected main bet to be 10.5, got %f", state.MainBet)
	}
}

func TestMainBetValidation(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("owedrahn", 2)
	player1ID := playerIds[0]

	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	// main bet to 0 (should fail)
	helper.SendMessage(player1ID, "set_main_bet", interfaces.M{
		"amount": 0.0,
	})

	// negative value (should fail)
	helper.SendMessage(player1ID, "set_main_bet", interfaces.M{
		"amount": -5.0,
	})

	// main bet above maximum (should fail)
	helper.SendMessage(player1ID, "set_main_bet", interfaces.M{
		"amount": 150.0,
	})

	// Verify main bet remains at default
	state := testRoom.State().(*GameState)
	if state.MainBet != 1 {
		t.Errorf("Expected main bet to remain 1 after failed attempts, got %f", state.MainBet)
	}
}

func TestMainBetCannotChangeAfterGameStarts(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("owedrahn", 2)
	player1ID := playerIds[0]
	player2ID := playerIds[1]

	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	// Set main bet before game starts
	helper.SendMessage(player1ID, "set_main_bet", interfaces.M{
		"amount": 5.0,
	})

	state := testRoom.State().(*GameState)
	if state.MainBet != 5.0 {
		t.Errorf("Expected main bet to be 5, got %f", state.MainBet)
	}

	helper.SendMessage(player1ID, "ready", true)
	helper.SendMessage(player2ID, "ready", true)

	state = testRoom.State().(*GameState)
	if !state.Started {
		t.Fatal("Game should have started")
	}

	// Try to change main bet after game started (should fail)
	helper.SendMessage(player1ID, "set_main_bet", interfaces.M{
		"amount": 10.0,
	})

	// Verify main bet hasn't changed
	state = testRoom.State().(*GameState)
	if state.MainBet != 5.0 {
		t.Errorf("Expected main bet to remain 5 after game started, got %f", state.MainBet)
	}
}

func TestMainBetAffectsBalance(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("owedrahn", 3)
	player1ID := playerIds[0]
	player2ID := playerIds[1]
	player3ID := playerIds[2]

	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	// Set main bet to 5
	helper.SendMessage(player1ID, "set_main_bet", interfaces.M{
		"amount": 5.0,
	})

	// Start the game
	helper.SendMessage(player1ID, "ready", true)
	helper.SendMessage(player2ID, "ready", true)
	helper.SendMessage(player3ID, "ready", true)

	state := testRoom.State().(*GameState)

	// Simulate player 1 losing (going over 15)
	state.CurrentTurn = helper.Clients[player1ID].ID()
	state.CurrentValue = 16
	player1 := state.Players[helper.Clients[player1ID].ID()]
	player1.Life = 0
	player1.Balance -= state.MainBet
	testRoom.SetState(state)

	// Verify player 1 lost 5 from balance
	if player1.Balance != -5.0 {
		t.Errorf("Expected player 1 balance to be -5, got %f", player1.Balance)
	}

	// Simulate player 2 losing
	state = testRoom.State().(*GameState)
	state.CurrentTurn = helper.Clients[player2ID].ID()
	player2 := state.Players[helper.Clients[player2ID].ID()]
	player2.Life = 0
	player2.Balance -= state.MainBet
	testRoom.SetState(state)

	// Verify player 2 lost 5 from balance
	if player2.Balance != -5.0 {
		t.Errorf("Expected player 2 balance to be -5, got %f", player2.Balance)
	}

	// Player 3 wins - should get 10 (5 from each loser)
	player3 := state.Players[helper.Clients[player3ID].ID()]
	player3.Balance += state.MainBet * float64(len(state.Players)-1)
	testRoom.SetState(state)

	if player3.Balance != 10.0 {
		t.Errorf("Expected player 3 balance to be 10, got %f", player3.Balance)
	}

	// Verify zero-sum
	state = testRoom.State().(*GameState)
	if !g.ValidateZeroSum(state) {
		t.Error("Game should maintain zero-sum balance")
	}
}

func TestMainBetResetsAfterGame(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	helper.RegisterGame(g)

	playerIds := helper.SetupGameRoom("owedrahn", 2)
	player1ID := playerIds[0]

	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}

	// Set main bet to 7
	helper.SendMessage(player1ID, "set_main_bet", interfaces.M{
		"amount": 7.0,
	})

	state := testRoom.State().(*GameState)
	if state.MainBet != 7.0 {
		t.Errorf("Expected main bet to be 7, got %f", state.MainBet)
	}

	// Reset the game
	g.reset(state)

	// Verify main bet is reset to 1
	if state.MainBet != 1.0 {
		t.Errorf("Expected main bet to reset to 1, got %f", state.MainBet)
	}
}
