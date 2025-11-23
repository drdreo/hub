package owe_drahn

import (
	"gameserver/games/owe_drahn/database"
	"gameserver/games/owe_drahn/models"
	"gameserver/internal/interfaces"
	"gameserver/internal/testicles"
	"testing"
)

func TestOweDrahnSideBet_ProposeToSelf(t *testing.T) {
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
	// Player 1 tries to propose a side bet to themselves
	payload := interfaces.M{
		"opponentId": helper.Clients[player1ID].ID(),
		"amount":     3,
	}
	helper.SendMessage(player1ID, "proposeSideBet", payload)
	state = testRoom.State().(*GameState)
	// Verify bet was NOT created
	if len(state.SideBets) != 0 {
		t.Errorf("Expected 0 side bets when proposing to self, got %d", len(state.SideBets))
	}
}

func TestOweDrahnSideBet_ProposeInvalidAmount(t *testing.T) {
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
	state := testRoom.State().(*GameState)

	testCases := []struct {
		amount      int
		description string
	}{
		{0, "zero"},
		{-5, "negative"},
		{-1, "negative one"},
	}
	for _, tc := range testCases {
		payload := interfaces.M{
			"opponentId": helper.Clients[player2ID].ID(),
			"amount":     tc.amount,
		}
		helper.SendMessage(player1ID, "proposeSideBet", payload)
		state = testRoom.State().(*GameState)
		if len(state.SideBets) != 0 {
			t.Errorf("Expected 0 side bets for %s amount %d, got %d", tc.description, tc.amount, len(state.SideBets))
		}
	}
}

func TestOweDrahnSideBet_ProposeValidBet(t *testing.T) {
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
	state := testRoom.State().(*GameState)

	// Player 1 proposes a valid bet to Player 2
	payload := interfaces.M{
		"opponentId": helper.Clients[player2ID].ID(),
		"amount":     5,
	}
	helper.SendMessage(player1ID, "proposeSideBet", payload)
	state = testRoom.State().(*GameState)

	// Verify bet was created
	if len(state.SideBets) != 1 {
		t.Fatalf("Expected 1 side bet, got %d", len(state.SideBets))
	}

	bet := state.SideBets[0]
	if bet.ChallengerID != player1ID {
		t.Errorf("Expected challenger ID %s, got %s", player1ID, bet.ChallengerID)
	}
	if bet.OpponentID != player2ID {
		t.Errorf("Expected opponent ID %s, got %s", player2ID, bet.OpponentID)
	}
	if bet.Amount != 5 {
		t.Errorf("Expected bet amount 5, got %f", bet.Amount)
	}
	if bet.Status != models.BetStatusPending {
		t.Errorf("Expected bet status Pending, got %d", bet.Status)
	}
}

func TestOweDrahnSideBet_ProposeInvalidOpponent(t *testing.T) {
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

	// Player 1 tries to propose a bet to a non-existent player
	payload := interfaces.M{
		"opponentId": "non-existent-player-id",
		"amount":     5,
	}
	helper.SendMessage(player1ID, "proposeSideBet", payload)
	state = testRoom.State().(*GameState)

	// Verify bet was NOT created
	if len(state.SideBets) != 0 {
		t.Errorf("Expected 0 side bets when proposing to invalid opponent, got %d", len(state.SideBets))
	}
}

func TestOweDrahnSideBet_AcceptValid(t *testing.T) {
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
	state := testRoom.State().(*GameState)

	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet1",
		ChallengerID: helper.Clients[player1ID].ID(),
		OpponentID:   helper.Clients[player2ID].ID(),
		Amount:       5,
		Status:       models.BetStatusPending,
	})
	// Player 2 accepts the bet
	payload := interfaces.M{
		"betId": "bet1",
	}
	helper.SendMessage(player2ID, "acceptSideBet", payload)
	state = testRoom.State().(*GameState)

	if len(state.SideBets) != 1 {
		t.Fatalf("Expected 1 side bet, got %d", len(state.SideBets))
	}
	bet := state.SideBets[0]
	if bet.Status != models.BetStatusAccepted {
		t.Errorf("Expected bet status Accepted, got %d", bet.Status)
	}
}

func TestOweDrahnSideBet_AcceptWrongPlayer(t *testing.T) {
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
	state := testRoom.State().(*GameState)

	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet1",
		ChallengerID: helper.Clients[player1ID].ID(),
		OpponentID:   helper.Clients[player2ID].ID(),
		Amount:       5,
		Status:       models.BetStatusPending,
	})
	// Player 3 tries to accept a bet they're not part of
	payload := interfaces.M{
		"betId": "bet1",
	}
	helper.SendMessage(player3ID, "acceptSideBet", payload)
	state = testRoom.State().(*GameState)
	// Verify bet status is still Pending
	bet := state.SideBets[0]
	if bet.Status != models.BetStatusPending {
		t.Errorf("Expected bet status to remain Pending when wrong player accepts, got %d", bet.Status)
	}
}
func TestOweDrahnSideBet_DeclineValid(t *testing.T) {
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
	state := testRoom.State().(*GameState)

	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet1",
		ChallengerID: helper.Clients[player1ID].ID(),
		OpponentID:   helper.Clients[player2ID].ID(),
		Amount:       5,
		Status:       models.BetStatusPending,
	})
	// Player 2 declines the bet
	payload := interfaces.M{
		"betId": "bet1",
	}
	helper.SendMessage(player2ID, "declineSideBet", payload)
	state = testRoom.State().(*GameState)
	// Verify bet status changed to Declined
	if len(state.SideBets) != 1 {
		t.Fatalf("Expected 1 side bet, got %d", len(state.SideBets))
	}
	bet := state.SideBets[0]
	if bet.Status != models.BetStatusDeclined {
		t.Errorf("Expected bet status Declined, got %d", bet.Status)
	}
}
func TestOweDrahnSideBet_ResolveChallengerWins(t *testing.T) {
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
	state := testRoom.State().(*GameState)
	player1 := state.Players[helper.Clients[player1ID].ID()]
	player2 := state.Players[helper.Clients[player2ID].ID()]
	player1.Balance = 10
	player2.Balance = 10

	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet1",
		ChallengerID: player1.ID,
		OpponentID:   player2.ID,
		Amount:       5,
		Status:       models.BetStatusAccepted,
	})
	// Player 2 loses all their lives (opponent loses)
	player2.Life = 0

	g.resolveSideBets(state)

	// Verify balances
	testicles.AssertFloatEquals(t, player1.Balance, 15, "expected player1 balance")
	testicles.AssertFloatEquals(t, player2.Balance, 5, "expected player2 balance")

	if state.SideBets[0].Status != models.BetStatusResolved {
		t.Errorf("Expected bet status Resolved, got %d", state.SideBets[0].Status)
	}

	if !g.ValidateZeroSum(state) {
		t.Error("Balance should maintain zero-sum after bet resolution")
	}
}

func TestOweDrahnSideBet_ResolveOpponentWins(t *testing.T) {
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
	state := testRoom.State().(*GameState)
	player1 := state.Players[helper.Clients[player1ID].ID()]
	player2 := state.Players[helper.Clients[player2ID].ID()]
	player1.Balance = 0
	player2.Balance = 0

	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet1",
		ChallengerID: player1.ID,
		OpponentID:   player2.ID,
		Amount:       5,
		Status:       models.BetStatusAccepted,
	})
	// Player 1 loses all their lives (challenger loses)
	player1.Life = 0

	g.resolveSideBets(state)

	// Verify balances
	testicles.AssertFloatEquals(t, player1.Balance, -5, "expected player1 balance")
	testicles.AssertFloatEquals(t, player2.Balance, 5, "expected player2 balance")

	// Verify bet status
	if state.SideBets[0].Status != models.BetStatusResolved {
		t.Errorf("Expected bet status Resolved, got %d", state.SideBets[0].Status)
	}
	if !g.ValidateZeroSum(state) {
		t.Error("Balance should maintain zero-sum after bet resolution")
	}
}
func TestOweDrahnSideBet_ResolveMultipleBets(t *testing.T) {
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
	state := testRoom.State().(*GameState)
	player1 := state.Players[helper.Clients[player1ID].ID()]
	player2 := state.Players[helper.Clients[player2ID].ID()]
	player3 := state.Players[helper.Clients[player3ID].ID()]

	player1.Balance = 0
	player2.Balance = 0
	player3.Balance = 0

	// Player 1 bets 5 against Player 2
	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet1",
		ChallengerID: player1.ID,
		OpponentID:   player2.ID,
		Amount:       5,
		Status:       models.BetStatusAccepted,
	})
	// Player 1 bets 3 against Player 3
	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet2",
		ChallengerID: player1.ID,
		OpponentID:   player3.ID,
		Amount:       2.5,
		Status:       models.BetStatusAccepted,
	})
	// Player 2 bets 4 against Player 3
	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet3",
		ChallengerID: player2.ID,
		OpponentID:   player3.ID,
		Amount:       4,
		Status:       models.BetStatusAccepted,
	})
	// Player 3 loses (everyone who bet against them wins)
	player3.Life = 0

	g.resolveSideBets(state)

	// Player 1: 0 + 2.5 (won bet2) = 2.5
	testicles.AssertFloatEquals(t, player1.Balance, 2.5, "expected player1 balance")
	// Player 2: 0 + 4 (won bet3) = 4
	testicles.AssertFloatEquals(t, player2.Balance, 4, "expected player2 balance")
	// Player 3: 0 - 3 (lost bet2) - 4 (lost bet3) = -7
	testicles.AssertFloatEquals(t, player3.Balance, -6.5, "expected player3 balance")

	// Verify all bets involving player 3 are resolved
	for i, bet := range state.SideBets {
		if bet.OpponentID == player3.ID || bet.ChallengerID == player3.ID {
			if bet.Status != models.BetStatusResolved {
				t.Errorf("Expected bet %d status Resolved, got %d", i, bet.Status)
			}
		}
	}
	// Verify bet1 (player1 vs player2) is NOT resolved since neither lost
	if state.SideBets[0].Status != models.BetStatusAccepted {
		t.Errorf("Expected bet1 status to remain Accepted, got %d", state.SideBets[0].Status)
	}

	if !g.ValidateZeroSum(state) {
		t.Error("Balance should maintain zero-sum after multiple bet resolutions")
	}
}
func TestOweDrahnSideBet_OnlyResolvesAcceptedBets(t *testing.T) {
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
	state := testRoom.State().(*GameState)
	player1 := state.Players[helper.Clients[player1ID].ID()]
	player2 := state.Players[helper.Clients[player2ID].ID()]
	player1.Balance = 0
	player2.Balance = 0

	// Create bets with different statuses
	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet1",
		ChallengerID: player1.ID,
		OpponentID:   player2.ID,
		Amount:       5,
		Status:       models.BetStatusPending,
	})
	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet2",
		ChallengerID: player1.ID,
		OpponentID:   player2.ID,
		Amount:       3,
		Status:       models.BetStatusDeclined,
	})
	// Player 2 loses
	player2.Life = 0

	g.resolveSideBets(state)

	// Verify balances remain unchanged (no accepted bets)
	testicles.AssertFloatEquals(t, player1.Balance, 0, "expected player1 balance")
	testicles.AssertFloatEquals(t, player2.Balance, 0, "expected player2 balance")

	// Verify bet statuses remain unchanged
	if state.SideBets[0].Status != models.BetStatusPending {
		t.Errorf("Expected bet1 status Pending, got %d", state.SideBets[0].Status)
	}
	if state.SideBets[1].Status != models.BetStatusDeclined {
		t.Errorf("Expected bet2 status Declined, got %d", state.SideBets[1].Status)
	}

	if !g.ValidateZeroSum(state) {
		t.Error("Balance should maintain zero-sum")
	}
}

func TestOweDrahnSideBet_AcceptNonExistentBet(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	helper.RegisterGame(g)
	playerIds := helper.SetupGameRoom("owedrahn", 2)
	player2ID := playerIds[1]
	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}
	state := testRoom.State().(*GameState)

	// Player 2 tries to accept a non-existent bet
	payload := interfaces.M{
		"betId": "non-existent-bet",
	}
	helper.SendMessage(player2ID, "acceptSideBet", payload)
	state = testRoom.State().(*GameState)

	// Verify no bets exist
	if len(state.SideBets) != 0 {
		t.Errorf("Expected 0 side bets, got %d", len(state.SideBets))
	}
}

func TestOweDrahnSideBet_DeclineNonExistentBet(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	helper.RegisterGame(g)
	playerIds := helper.SetupGameRoom("owedrahn", 2)
	player2ID := playerIds[1]
	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}
	state := testRoom.State().(*GameState)

	// Player 2 tries to decline a non-existent bet
	payload := interfaces.M{
		"betId": "non-existent-bet",
	}
	helper.SendMessage(player2ID, "declineSideBet", payload)
	state = testRoom.State().(*GameState)

	// Verify no bets exist
	if len(state.SideBets) != 0 {
		t.Errorf("Expected 0 side bets, got %d", len(state.SideBets))
	}
}

func TestOweDrahnSideBet_DeclineWrongPlayer(t *testing.T) {
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
	state := testRoom.State().(*GameState)

	// Create a pending bet between player 1 and 2
	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet1",
		ChallengerID: helper.Clients[player1ID].ID(),
		OpponentID:   helper.Clients[player2ID].ID(),
		Amount:       5,
		Status:       models.BetStatusPending,
	})

	// Player 3 tries to decline a bet they're not part of
	payload := interfaces.M{
		"betId": "bet1",
	}
	helper.SendMessage(player3ID, "declineSideBet", payload)
	state = testRoom.State().(*GameState)

	// Verify bet status is still Pending
	bet := state.SideBets[0]
	if bet.Status != models.BetStatusPending {
		t.Errorf("Expected bet status to remain Pending when wrong player declines, got %d", bet.Status)
	}
}
func TestOweDrahnSideBet_ResolveBothPlayersAlive(t *testing.T) {
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
	state := testRoom.State().(*GameState)
	player1 := state.Players[helper.Clients[player1ID].ID()]
	player2 := state.Players[helper.Clients[player2ID].ID()]
	player1.Balance = 0
	player2.Balance = 0

	// Create an accepted bet
	state.SideBets = append(state.SideBets, &models.SideBet{
		ID:           "bet1",
		ChallengerID: player1.ID,
		OpponentID:   player2.ID,
		Amount:       5,
		Status:       models.BetStatusAccepted,
	})

	// Both players are still alive
	player1.Life = 3
	player2.Life = 2

	// Resolve side bets (should not resolve anything)
	g.resolveSideBets(state)

	// Verify balances remain unchanged
	testicles.AssertFloatEquals(t, player1.Balance, 0, "expected player1 balance")
	testicles.AssertFloatEquals(t, player2.Balance, 0, "expected player2 balance")

	// Verify bet status remains Accepted
	if state.SideBets[0].Status != models.BetStatusAccepted {
		t.Errorf("Expected bet status to remain Accepted when both players alive, got %d", state.SideBets[0].Status)
	}

	// Verify zero-sum
	if !g.ValidateZeroSum(state) {
		t.Error("Balance should maintain zero-sum")
	}
}

func TestOweDrahnSideBet_ZeroSumValidation(t *testing.T) {
	helper := testicles.NewTestHelper(t)
	dbServiceMock := &database.DatabaseServiceMock{}
	g := NewGame(dbServiceMock)
	helper.RegisterGame(g)
	playerIds := helper.SetupGameRoom("owedrahn", 4)
	testRoom, err := helper.GetRoom()
	if err != nil {
		t.Fatalf("Failed to get room: %v", err)
	}
	state := testRoom.State().(*GameState)
	// Test 1: All zeros
	for _, id := range playerIds {
		state.Players[helper.Clients[id].ID()].Balance = 0
	}
	if !g.ValidateZeroSum(state) {
		t.Error("Expected zero-sum to be valid when all balances are 0")
	}
	// Test 2: Balanced positive and negative
	state.Players[helper.Clients[playerIds[0]].ID()].Balance = 10
	state.Players[helper.Clients[playerIds[1]].ID()].Balance = -5
	state.Players[helper.Clients[playerIds[2]].ID()].Balance = -3
	state.Players[helper.Clients[playerIds[3]].ID()].Balance = -2
	if !g.ValidateZeroSum(state) {
		t.Error("Expected zero-sum to be valid when balances sum to 0")
	}
	// Test 3: Unbalanced (should fail)
	state.Players[helper.Clients[playerIds[0]].ID()].Balance = 10
	state.Players[helper.Clients[playerIds[1]].ID()].Balance = 5
	state.Players[helper.Clients[playerIds[2]].ID()].Balance = -3
	state.Players[helper.Clients[playerIds[3]].ID()].Balance = -2
	if g.ValidateZeroSum(state) {
		t.Error("Expected zero-sum to be invalid when balances don't sum to 0")
	}
}
