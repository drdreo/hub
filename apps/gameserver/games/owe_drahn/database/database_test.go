package database

import (
	"testing"

	"gameserver/games/owe_drahn/models"
)

func TestMergeStatsInitializesMissingRolledDice(t *testing.T) {
	result := MergeStats(models.PlayerStats{}, models.PlayerStatAggregation{
		RolledDice:   []int{1, 2, 3, 4, 5, 6},
		Won:          true,
		PerfectRoll:  1,
		LuckiestRoll: 2,
		WorstRoll:    1,
		Rolled21:     1,
		MaxLifeLoss:  1,
	})

	if len(result.RolledDice) != diceBucketCount {
		t.Fatalf("expected rolled dice length %d, got %d", diceBucketCount, len(result.RolledDice))
	}

	for i, expected := range []int{1, 2, 3, 4, 5, 6} {
		if result.RolledDice[i] != expected {
			t.Fatalf("rolledDice[%d] = %d, want %d", i, result.RolledDice[i], expected)
		}
	}

	if result.Wins != 1 {
		t.Fatalf("expected wins to increment, got %d", result.Wins)
	}

	if result.TotalGames != 1 {
		t.Fatalf("expected total games to increment, got %d", result.TotalGames)
	}

	if result.PerfectRoll != 1 || result.LuckiestRoll != 2 || result.WorstRoll != 1 || result.Rolled21 != 1 || result.MaxLifeLoss != 1 {
		t.Fatalf("unexpected stat aggregates: %+v", result)
	}
}
