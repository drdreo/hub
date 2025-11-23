package testicles

import (
	"testing"
)

func TestFloatEquals(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		expected bool
	}{
		{
			name:     "exact equal values",
			a:        1.0,
			b:        1.0,
			expected: true,
		},
		{
			name:     "very close values within epsilon",
			a:        1.0000000001,
			b:        1.0,
			expected: true,
		},
		{
			name:     "different values outside epsilon",
			a:        1.0001,
			b:        1.0,
			expected: false,
		},
		{
			name:     "negative values equal",
			a:        -5.5,
			b:        -5.5,
			expected: true,
		},
		{
			name:     "zero values",
			a:        0.0,
			b:        0.0,
			expected: true,
		},
		{
			name:     "zero and very small value",
			a:        0.0,
			b:        1e-10,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FloatEquals(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("FloatEquals(%f, %f) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestFloatEqualsWithEpsilon(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		epsilon  float64
		expected bool
	}{
		{
			name:     "within custom epsilon",
			a:        1.05,
			b:        1.0,
			epsilon:  0.1,
			expected: true,
		},
		{
			name:     "outside custom epsilon",
			a:        1.15,
			b:        1.0,
			epsilon:  0.1,
			expected: false,
		},
		{
			name:     "exact match with any epsilon",
			a:        42.0,
			b:        42.0,
			epsilon:  0.001,
			expected: true,
		},
		{
			name:     "larger epsilon allows more difference",
			a:        1.5,
			b:        1.0,
			epsilon:  1.0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FloatEqualsWithEpsilon(tt.a, tt.b, tt.epsilon)
			if result != tt.expected {
				t.Errorf("FloatEqualsWithEpsilon(%f, %f, %e) = %v, expected %v", tt.a, tt.b, tt.epsilon, result, tt.expected)
			}
		})
	}
}

func TestFloat32Equals(t *testing.T) {
	tests := []struct {
		name     string
		a        float32
		b        float32
		expected bool
	}{
		{
			name:     "exact equal values",
			a:        1.0,
			b:        1.0,
			expected: true,
		},
		{
			name:     "very close values within epsilon",
			a:        1.000001,
			b:        1.0,
			expected: true,
		},
		{
			name:     "different values outside epsilon",
			a:        1.001,
			b:        1.0,
			expected: false,
		},
		{
			name:     "negative values equal",
			a:        -3.14,
			b:        -3.14,
			expected: true,
		},
		{
			name:     "zero values",
			a:        0.0,
			b:        0.0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Float32Equals(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Float32Equals(%f, %f) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestFloat32EqualsWithEpsilon(t *testing.T) {
	tests := []struct {
		name     string
		a        float32
		b        float32
		epsilon  float32
		expected bool
	}{
		{
			name:     "within custom epsilon",
			a:        1.05,
			b:        1.0,
			epsilon:  0.1,
			expected: true,
		},
		{
			name:     "outside custom epsilon",
			a:        1.15,
			b:        1.0,
			epsilon:  0.1,
			expected: false,
		},
		{
			name:     "exact match with any epsilon",
			a:        9.99,
			b:        9.99,
			epsilon:  0.001,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Float32EqualsWithEpsilon(tt.a, tt.b, tt.epsilon)
			if result != tt.expected {
				t.Errorf("Float32EqualsWithEpsilon(%f, %f, %e) = %v, expected %v", tt.a, tt.b, tt.epsilon, result, tt.expected)
			}
		})
	}
}

func TestAssertFloatEquals(t *testing.T) {
	// Test that passes
	t.Run("should pass for equal values", func(t *testing.T) {
		mockT := &testing.T{}
		AssertFloatEquals(mockT, 1.0, 1.0000000001)
		if mockT.Failed() {
			t.Error("AssertFloatEquals should not fail for nearly equal values")
		}
	})

	// Test with custom message
	t.Run("should include custom message on failure", func(t *testing.T) {
		mockT := &testing.T{}
		AssertFloatEquals(mockT, 1.0, 2.0, "custom error message")
		if !mockT.Failed() {
			t.Error("AssertFloatEquals should fail for different values")
		}
	})
}

func TestAssertFloat32Equals(t *testing.T) {
	// Test that passes
	t.Run("should pass for equal values", func(t *testing.T) {
		mockT := &testing.T{}
		AssertFloat32Equals(mockT, float32(1.0), float32(1.000001))
		if mockT.Failed() {
			t.Error("AssertFloat32Equals should not fail for nearly equal values")
		}
	})

	// Test with custom message
	t.Run("should include custom message on failure", func(t *testing.T) {
		mockT := &testing.T{}
		AssertFloat32Equals(mockT, float32(1.0), float32(2.0), "custom error message")
		if !mockT.Failed() {
			t.Error("AssertFloat32Equals should fail for different values")
		}
	})
}
