package main

import (
	"testing"
)

// TestAdd tests the add function
func TestAdd(t *testing.T) {
	result := add(2, 3)
	if result != 5 {
		t.Errorf("add(2, 3) = %d; want 5", result)
	}
}

// TestSubtract tests the Subtract function
func TestSubtract(t *testing.T) {
	result := Subtract(5, 3)
	if result != 2 {
		t.Errorf("Subtract(5, 3) = %d; want 2", result)
	}
}

// TestMultiply tests the multiply function
func TestMultiply(t *testing.T) {
	result := multiply(2, 3)
	if result != 6 {
		t.Errorf("multiply(2, 3) = %d; want 6", result)
	}
}

// TestDivide tests the Divide function
func TestDivide(t *testing.T) {
	t.Run("valid division", func(t *testing.T) {
		result := Divide(6, 3)
		if result != 2 {
			t.Errorf("Divide(6, 3) = %d; want 2", result)
		}
	})

	t.Run("division by zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Divide(6, 0) did not panic")
			}
		}()
		Divide(6, 0) // This should panic
	})
}
