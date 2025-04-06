package dicegame

import (
	"testing"
)

func TestCalculateScore(t *testing.T) {
	diceGame := &DiceGame{}

	testCases := []struct {
		name     string
		dice     []int
		expected int
		valid    bool
	}{
		// Empty dice should return 0 and false
		{
			name:     "empty dice",
			dice:     []int{},
			expected: 0,
			valid:    false,
		},
		// Valid single dice combinations
		{
			name:     "single 1",
			dice:     []int{1},
			expected: 100,
			valid:    true,
		},
		{
			name:     "single 5",
			dice:     []int{5},
			expected: 50,
			valid:    true,
		},
		// Invalid single dice
		{
			name:     "single 2",
			dice:     []int{2},
			expected: 0,
			valid:    false,
		},
		{
			name:     "single 3",
			dice:     []int{3},
			expected: 0,
			valid:    false,
		},
		{
			name:     "single 4",
			dice:     []int{4},
			expected: 0,
			valid:    false,
		},
		{
			name:     "single 6",
			dice:     []int{6},
			expected: 0,
			valid:    false,
		},
		// Multiple 1s and 5s
		{
			name:     "multiple 1s",
			dice:     []int{1, 1},
			expected: 200,
			valid:    true,
		},
		{
			name:     "multiple 5s",
			dice:     []int{5, 5},
			expected: 100,
			valid:    true,
		},
		{
			name:     "mixed 1s and 5s",
			dice:     []int{1, 5},
			expected: 150,
			valid:    true,
		},
		// Three of a kind
		{
			name:     "three 1s",
			dice:     []int{1, 1, 1},
			expected: 1000,
			valid:    true,
		},
		{
			name:     "three 2s",
			dice:     []int{2, 2, 2},
			expected: 200,
			valid:    true,
		},
		{
			name:     "three 3s",
			dice:     []int{3, 3, 3},
			expected: 300,
			valid:    true,
		},
		{
			name:     "three 4s",
			dice:     []int{4, 4, 4},
			expected: 400,
			valid:    true,
		},
		{
			name:     "three 5s",
			dice:     []int{5, 5, 5},
			expected: 500,
			valid:    true,
		},
		{
			name:     "three 6s",
			dice:     []int{6, 6, 6},
			expected: 600,
			valid:    true,
		},
		// More than three of a kind
		{
			name:     "four 1s",
			dice:     []int{1, 1, 1, 1},
			expected: 2000, // Double the score of three 1s
			valid:    true,
		},
		{
			name:     "four 4s",
			dice:     []int{4, 4, 4, 4},
			expected: 800,
			valid:    true,
		},
		{
			name:     "four 5s",
			dice:     []int{5, 5, 5, 5},
			expected: 1000,
			valid:    true,
		},
		{
			name:     "six 1s",
			dice:     []int{1, 1, 1, 1, 1, 1},
			expected: 8000,
			valid:    true,
		},
		{
			name:     "six 3s",
			dice:     []int{3, 3, 3, 3, 3, 3},
			expected: 2400,
			valid:    true,
		},
		{
			name:     "six 5s",
			dice:     []int{5, 5, 5, 5, 5, 5},
			expected: 4000,
			valid:    true,
		},
		// Runs
		{
			name:     "run 1-5",
			dice:     []int{1, 2, 3, 4, 5},
			expected: 500,
			valid:    true,
		},
		{
			name:     "run 1-5 + 1",
			dice:     []int{1, 2, 3, 4, 5, 1},
			expected: 600,
			valid:    true,
		},
		{
			name:     "run 1-5 + 5",
			dice:     []int{1, 2, 3, 4, 5, 5},
			expected: 550,
			valid:    true,
		},
		{
			name:     "run 1-5 + 3",
			dice:     []int{1, 2, 3, 4, 5, 3},
			expected: 500,
			valid:    false,
		},
		{
			name:     "run 2-6",
			dice:     []int{2, 3, 4, 5, 6},
			expected: 750,
			valid:    true,
		},
		{
			name:     "run 2-6 + 5",
			dice:     []int{5, 2, 3, 4, 5, 6},
			expected: 800,
			valid:    true,
		},
		{
			name:     "run 2-6 + 3",
			dice:     []int{3, 2, 3, 4, 5, 6},
			expected: 750,
			valid:    false,
		},
		{
			name:     "run 1-6",
			dice:     []int{1, 2, 3, 4, 5, 6},
			expected: 1500,
			valid:    true,
		},
		{
			name:     "scrambled run 1-6",
			dice:     []int{6, 4, 2, 1, 3, 5},
			expected: 1500,
			valid:    true,
		},
		// Mixed valid combinations
		{
			name:     "three 1s plus a 5",
			dice:     []int{1, 1, 1, 5},
			expected: 1050, // 1000 for three 1s + 50 for one 5
			valid:    true,
		},
		{
			name:     "three 2s plus a 1",
			dice:     []int{2, 2, 2, 1},
			expected: 300, // 200 for three 2s + 100 for one 1
			valid:    true,
		},
		{
			name:     "scrambled run 1-6",
			dice:     []int{6, 4, 2, 1, 3, 5},
			expected: 1500,
			valid:    true,
		},
		// Invalid combinations
		{
			name:     "two 1s and one 3",
			dice:     []int{1, 1, 3},
			expected: 200,   // only count the valid 1s
			valid:    false, // invalid because of the 3
		},
		{
			name:     "two 2s and one 3",
			dice:     []int{2, 2, 3},
			expected: 0,
			valid:    false, // No valid scoring combination
		},
		{
			name:     "two 3s and one 4",
			dice:     []int{3, 3, 4},
			expected: 0,
			valid:    false,
		},
		{
			name:     "mixed non-scoring dice",
			dice:     []int{2, 3, 4, 6},
			expected: 0,
			valid:    false,
		},
		// Edge cases
		{
			name:     "incomplete run 1-4",
			dice:     []int{1, 2, 3, 4},
			expected: 100, // only count the 1
			valid:    false,
		},
		{
			name:     "almost three of a kind plus scoring dice",
			dice:     []int{2, 2, 1},
			expected: 100, // only count the 1
			valid:    false,
		},
		// Complex cases
		{
			name:     "multiple three of a kind",
			dice:     []int{1, 1, 1, 2, 2, 2},
			expected: 1200, // 1000 for three 1s + 200 for three 2s
			valid:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score, valid := diceGame.CalculateScore(tc.dice)
			if score != tc.expected {
				t.Errorf("Expected score %d, got %d for dice %v", tc.expected, score, tc.dice)
			}
			if valid != tc.valid {
				t.Errorf("Expected valid=%v, got valid=%v for dice %v", tc.valid, valid, tc.dice)
			}
		})
	}
}
