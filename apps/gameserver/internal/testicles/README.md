# Testicles - Testing Utilities Package

A collection of testing utilities for the gameserver project.

## Float Comparison Helpers

The package provides helper functions for comparing floating-point numbers, which is necessary due to floating-point precision issues in computers.

### Functions

#### Float64 Comparison

-   `FloatEquals(a, b float64) bool` - Checks if two float64 values are equal within default epsilon (1e-9)
-   `FloatEqualsWithEpsilon(a, b, epsilon float64) bool` - Checks if two float64 values are equal within custom epsilon
-   `AssertFloatEquals(t *testing.T, expected, actual float64, msgAndArgs ...interface{})` - Asserts equality with default epsilon
-   `AssertFloatEqualsWithEpsilon(t *testing.T, expected, actual, epsilon float64, msgAndArgs ...interface{})` - Asserts equality with custom epsilon

#### Float32 Comparison

-   `Float32Equals(a, b float32) bool` - Checks if two float32 values are equal within default epsilon (1e-6)
-   `Float32EqualsWithEpsilon(a, b, epsilon float32) bool` - Checks if two float32 values are equal within custom epsilon
-   `AssertFloat32Equals(t *testing.T, expected, actual float32, msgAndArgs ...interface{})` - Asserts equality with default epsilon
-   `AssertFloat32EqualsWithEpsilon(t *testing.T, expected, actual, epsilon float32, msgAndArgs ...interface{})` - Asserts equality with custom epsilon

### Usage Examples

```go
package mypackage

import (
    "testing"
    "gameserver/internal/testicles"
)

func TestCalculation(t *testing.T) {
    result := 0.1 + 0.2 // May not exactly equal 0.3 due to floating-point precision

    // Using boolean check
    if !testicles.FloatEquals(result, 0.3) {
        t.Error("calculation failed")
    }

    // Using assert (recommended)
    testicles.AssertFloatEquals(t, 0.3, result)

    // With custom error message
    testicles.AssertFloatEquals(t, 0.3, result, "addition of 0.1 and 0.2")

    // With custom epsilon for less precision
    testicles.AssertFloatEqualsWithEpsilon(t, 1.0, 1.05, 0.1, "values should be close")
}

func TestFloat32Calculation(t *testing.T) {
    var result float32 = 1.0 / 3.0

    // Using boolean check
    if testicles.Float32Equals(result, 0.333333) {
        t.Log("values are equal within epsilon")
    }

    // Using assert with custom epsilon
    testicles.AssertFloat32EqualsWithEpsilon(t, 0.333333, result, 0.00001)
}
```

## Why Use Float Comparison Helpers?

Direct comparison of floating-point numbers can lead to unexpected test failures:

```go
// ❌ BAD - May fail due to floating-point precision
if 0.1 + 0.2 == 0.3 {
    // This might be false!
}

// ✅ GOOD - Uses epsilon comparison
if testicles.FloatEquals(0.1 + 0.2, 0.3) {
    // This will be true
}
```

## Test Helper

The `TestHelper` struct provides utilities for setting up integration tests with game rooms and clients.

### Example

```go
func TestGameFeature(t *testing.T) {
    helper := testicles.NewTestHelper(t)
    g := NewGame(dbServiceMock)
    helper.RegisterGame(g)

    // Setup game room with 2 players
    playerIds := helper.SetupGameRoom("owedrahn", 2)

    // Send messages
    helper.SendMessage(playerIds[0], "roll", nil)

    // Verify messages received
    helper.AssertMessageReceived(playerIds[0], "rolledDice")
}
```
