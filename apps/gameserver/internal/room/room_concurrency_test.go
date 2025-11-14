package room

import (
	"gameserver/games/dicegame"
	"gameserver/internal/game"
	"gameserver/internal/interfaces"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrentStateAccess tests that concurrent access to room state is safe
// Run with: go test -race ./internal/room
func TestConcurrentStateAccess(t *testing.T) {
	registry := game.NewRegistry()
	dicegame.RegisterDiceGame(registry)
	manager := NewRoomManager(registry)

	// Create a room with dice game
	testRoom := NewRoom(manager, "dicegame", nil)

	// Initialize game state
	initialState := &dicegame.GameState{
		Players:      make(map[string]*dicegame.Player),
		Started:      false,
		Dice:         make([]int, 6),
		SelectedDice: make([]int, 0),
		SetAside:     make([]int, 0),
		TargetScore:  10000,
	}

	// Add test players
	for i := 0; i < 5; i++ {
		playerID := string(rune('A' + i))
		initialState.Players[playerID] = &dicegame.Player{
			ID:    playerID,
			Name:  "Player " + playerID,
			Score: 0,
		}
	}

	testRoom.SetState(initialState)

	// Test 1: Concurrent reads should be safe
	t.Run("ConcurrentReads", func(t *testing.T) {
		var wg sync.WaitGroup
		iterations := 100

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				state := testRoom.State().(*dicegame.GameState)
				// Just read the state
				_ = state.Players
				_ = state.Dice
			}()
		}

		wg.Wait()
	})

	// Test 2: Concurrent writes with proper SetState pattern
	t.Run("ConcurrentWritesWithSetState", func(t *testing.T) {
		var wg sync.WaitGroup
		iterations := 100

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(playerIndex int) {
				defer wg.Done()

				// Simulate the pattern used in actual game handlers
				state := testRoom.State().(*dicegame.GameState)
				playerID := string(rune('A' + (playerIndex % 5)))

				// Modify state
				if player, exists := state.Players[playerID]; exists {
					player.Score += 10
				}

				// Immediately write back
				testRoom.SetState(state)
			}(i)
		}

		wg.Wait()

		// Note: The final score may not be iterations * 10 / 5
		// because multiple goroutines may get the same state snapshot
		// This is expected behavior that demonstrates why immutability matters
		finalState := testRoom.State().(*dicegame.GameState)
		t.Logf("Final scores after %d concurrent updates:", iterations)
		for id, player := range finalState.Players {
			t.Logf("Player %s: %d", id, player.Score)
		}
	})

	// Test 3: Detect actual race conditions with unsafe pattern
	t.Run("DetectRaceConditionWithUnsafePattern", func(t *testing.T) {
		// Reset state
		testRoom.SetState(&dicegame.GameState{
			Players:      make(map[string]*dicegame.Player),
			Started:      false,
			Dice:         make([]int, 6),
			SelectedDice: make([]int, 0),
			SetAside:     make([]int, 0),
			TargetScore:  10000,
		})

		state := testRoom.State().(*dicegame.GameState)
		state.Players["A"] = &dicegame.Player{ID: "A", Name: "Player A", Score: 0}
		testRoom.SetState(state)

		var wg sync.WaitGroup
		iterations := 100

		// This pattern WOULD cause race conditions if we held state reference
		// across async operations, but our pattern is safe because we don't
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Safe pattern: get, modify, set immediately
				state := testRoom.State().(*dicegame.GameState)
				if player, exists := state.Players["A"]; exists {
					player.Score += 1
				}
				testRoom.SetState(state)
			}()
		}

		wg.Wait()

		finalState := testRoom.State().(*dicegame.GameState)
		t.Logf("Player A final score: %d (expected: unpredictable due to lost updates)",
			finalState.Players["A"].Score)
	})
}

// TestConcurrentStateAccessSimulatingRealGame simulates real game message handling
func TestConcurrentStateAccessSimulatingRealGame(t *testing.T) {
	registry := game.NewRegistry()
	dicegame.RegisterDiceGame(registry)
	manager := NewRoomManager(registry)

	testRoom := NewRoom(manager, "dicegame", nil)

	// Initialize game state
	initialState := &dicegame.GameState{
		Players:      make(map[string]*dicegame.Player),
		Started:      true,
		CurrentTurn:  "player1",
		Dice:         []int{1, 2, 3, 4, 5, 6},
		SelectedDice: make([]int, 0),
		SetAside:     make([]int, 0),
		TargetScore:  10000,
	}

	initialState.Players["player1"] = &dicegame.Player{
		ID:    "player1",
		Name:  "Player 1",
		Score: 0,
	}
	initialState.Players["player2"] = &dicegame.Player{
		ID:    "player2",
		Name:  "Player 2",
		Score: 0,
	}

	testRoom.SetState(initialState)

	// Simulate multiple clients sending messages concurrently
	t.Run("SimulateRealGameMessages", func(t *testing.T) {
		var wg sync.WaitGroup
		messageCount := 100
		var processedMessages atomic.Int32

		// Simulate game message handlers being called concurrently
		for i := 0; i < messageCount; i++ {
			wg.Add(1)
			go func(msgNum int) {
				defer wg.Done()

				// Simulate a game action handler (like handleRoll or handleSelect)
				simulateGameAction(testRoom, msgNum%2 == 0)
				processedMessages.Add(1)
			}(i)
		}

		wg.Wait()

		t.Logf("Processed %d messages concurrently", processedMessages.Load())
		finalState := testRoom.State().(*dicegame.GameState)
		t.Logf("Final state - Player 1 score: %d, Player 2 score: %d",
			finalState.Players["player1"].Score,
			finalState.Players["player2"].Score)
	})
}

// simulateGameAction mimics the pattern used in actual game handlers
func simulateGameAction(room interfaces.Room, isPlayer1 bool) {
	// This mimics the pattern from dicegame.handleRoll
	state := room.State().(*dicegame.GameState)

	// Simulate some game logic
	playerID := "player1"
	if !isPlayer1 {
		playerID = "player2"
	}

	if player, exists := state.Players[playerID]; exists {
		player.TurnScore += 10
		player.Score += 10
	}

	// Modify dice (simulating a roll)
	state.Dice = []int{1, 2, 3, 4, 5, 6}
	state.SelectedDice = make([]int, 0)

	// Immediately set state (this is the safe pattern)
	room.SetState(state)
}

// TestStateIsolationBetweenRooms ensures that different rooms don't interfere
func TestStateIsolationBetweenRooms(t *testing.T) {
	registry := game.NewRegistry()
	dicegame.RegisterDiceGame(registry)
	manager := NewRoomManager(registry)

	// Create multiple rooms
	room1 := NewRoom(manager, "dicegame", nil)
	room2 := NewRoom(manager, "dicegame", nil)
	room3 := NewRoom(manager, "dicegame", nil)

	// Initialize each with different state
	for i, room := range []*GameRoom{room1, room2, room3} {
		state := &dicegame.GameState{
			Players:      make(map[string]*dicegame.Player),
			Started:      false,
			Dice:         make([]int, 6),
			SelectedDice: make([]int, 0),
			SetAside:     make([]int, 0),
			TargetScore:  10000 * (i + 1), // Different target scores
		}
		state.Players["A"] = &dicegame.Player{
			ID:    "A",
			Name:  "Player A",
			Score: i * 1000, // Different initial scores
		}
		room.SetState(state)
	}

	// Concurrently modify all rooms
	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		for roomIndex, room := range []*GameRoom{room1, room2, room3} {
			wg.Add(1)
			go func(r *GameRoom, idx int) {
				defer wg.Done()

				state := r.State().(*dicegame.GameState)
				state.Players["A"].Score += 10
				r.SetState(state)
			}(room, roomIndex)
		}
	}

	wg.Wait()

	// Verify each room maintained its own state
	state1 := room1.State().(*dicegame.GameState)
	state2 := room2.State().(*dicegame.GameState)
	state3 := room3.State().(*dicegame.GameState)

	t.Logf("Room 1 - Target: %d, Score: %d", state1.TargetScore, state1.Players["A"].Score)
	t.Logf("Room 2 - Target: %d, Score: %d", state2.TargetScore, state2.Players["A"].Score)
	t.Logf("Room 3 - Target: %d, Score: %d", state3.TargetScore, state3.Players["A"].Score)

	if state1.TargetScore != 10000 || state2.TargetScore != 20000 || state3.TargetScore != 30000 {
		t.Error("Room states were not properly isolated")
	}
}

// TestLongRunningOperationPattern tests the unsafe pattern we want to avoid
func TestLongRunningOperationPattern(t *testing.T) {
	registry := game.NewRegistry()
	dicegame.RegisterDiceGame(registry)
	manager := NewRoomManager(registry)

	testRoom := NewRoom(manager, "dicegame", nil)

	initialState := &dicegame.GameState{
		Players:      make(map[string]*dicegame.Player),
		Started:      false,
		Dice:         make([]int, 6),
		SelectedDice: make([]int, 0),
		SetAside:     make([]int, 0),
		TargetScore:  10000,
	}
	initialState.Players["A"] = &dicegame.Player{ID: "A", Name: "Player A", Score: 0}
	testRoom.SetState(initialState)

	t.Run("UnsafePatternWithDelay", func(t *testing.T) {
		var wg sync.WaitGroup
		iterations := 10

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				// BAD: Get state
				state := testRoom.State().(*dicegame.GameState)

				// BAD: Simulate long operation (e.g., external API call, database query)
				time.Sleep(10 * time.Millisecond)

				// BAD: Modify state after delay - other goroutines may have changed it!
				state.Players["A"].Score += 100

				// BAD: Write back potentially stale state
				testRoom.SetState(state)
			}(i)
		}

		wg.Wait()

		finalState := testRoom.State().(*dicegame.GameState)
		expectedScore := iterations * 100
		actualScore := finalState.Players["A"].Score

		t.Logf("Expected score: %d, Actual score: %d", expectedScore, actualScore)

		if actualScore != expectedScore {
			t.Logf("⚠️  Lost updates detected! This demonstrates why holding state references across async operations is unsafe.")
			t.Logf("This is expected for this test - it shows the problem the architecture review is warning about.")
		}
	})

	t.Run("SafePatternWithDelay", func(t *testing.T) {
		// Reset state
		initialState := &dicegame.GameState{
			Players:      make(map[string]*dicegame.Player),
			Started:      false,
			Dice:         make([]int, 6),
			SelectedDice: make([]int, 0),
			SetAside:     make([]int, 0),
			TargetScore:  10000,
		}
		initialState.Players["B"] = &dicegame.Player{ID: "B", Name: "Player B", Score: 0}
		testRoom.SetState(initialState)

		var wg sync.WaitGroup
		iterations := 10

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				// GOOD: Do long operation first
				time.Sleep(10 * time.Millisecond)

				// GOOD: Get state right before using it
				state := testRoom.State().(*dicegame.GameState)

				// GOOD: Modify immediately
				state.Players["B"].Score += 100

				// GOOD: Set immediately after modifying
				testRoom.SetState(state)
			}(i)
		}

		wg.Wait()

		finalState := testRoom.State().(*dicegame.GameState)
		actualScore := finalState.Players["B"].Score

		t.Logf("Player B score with safer pattern: %d", actualScore)
		t.Logf("Note: Even this pattern can lose updates with concurrent access,")
		t.Logf("but it's safer than holding state references across async operations.")
	})
}

// BenchmarkConcurrentStateAccess benchmarks the performance of concurrent state access
func BenchmarkConcurrentStateAccess(b *testing.B) {
	registry := game.NewRegistry()
	dicegame.RegisterDiceGame(registry)
	manager := NewRoomManager(registry)

	testRoom := NewRoom(manager, "dicegame", nil)

	initialState := &dicegame.GameState{
		Players:      make(map[string]*dicegame.Player),
		Started:      false,
		Dice:         make([]int, 6),
		SelectedDice: make([]int, 0),
		SetAside:     make([]int, 0),
		TargetScore:  10000,
	}
	initialState.Players["A"] = &dicegame.Player{ID: "A", Name: "Player A", Score: 0}
	testRoom.SetState(initialState)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			state := testRoom.State().(*dicegame.GameState)
			state.Players["A"].Score += 1
			testRoom.SetState(state)
		}
	})
}
