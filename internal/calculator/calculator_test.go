// calculator_test.go
package calculator

import (
	"math"
	"testing"
)

// floatEquals is a helper function to compare floating-point numbers with a tolerance
// to handle precision issues that occur with floating-point arithmetic
func floatEquals(a, b float64, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

// =============================================================================
// BASIC ARITHMETIC OPERATIONS TESTS
// These tests verify the core mathematical operations work correctly with
// standard positive integer inputs
// =============================================================================

// TestAdd verifies that basic addition works correctly
// Tests: 2 + 3 = 5
func TestAdd(t *testing.T) {
	calc := New()
	result := calc.Add(2, 3)
	if result != 5 {
		t.Errorf("Expected 5, got %f", result)
	}
}

// TestSubtract verifies that basic subtraction works correctly
// Tests: 5 - 3 = 2
func TestSubtract(t *testing.T) {
	calc := New()
	result := calc.Subtract(5, 3)
	if result != 2 {
		t.Errorf("Expected 2, got %f", result)
	}
}

// TestMultiply verifies that basic multiplication works correctly
// Tests: 3 * 4 = 12
func TestMultiply(t *testing.T) {
	calc := New()
	result := calc.Multiply(3, 4)
	if result != 12 {
		t.Errorf("Expected 12, got %f", result)
	}
}

// =============================================================================
// DIVISION AND MODULUS OPERATIONS TESTS
// These tests verify division and modulus operations, including proper error
// handling for division/modulus by zero scenarios
// =============================================================================

// TestDivide verifies division works correctly and handles division by zero
// Tests both successful division (10 / 2 = 5) and error case (division by zero)
func TestDivide(t *testing.T) {
	calc := New()

	// Test normal division case
	result, err := calc.Divide(10, 2)
	if err != nil || result != 5 {
		t.Errorf("Expected 5, got %f (err: %v)", result, err)
	}

	// Test division by zero should return an error
	_, err = calc.Divide(10, 0)
	if err == nil {
		t.Error("Expected division by zero error")
	}
}

// TestMod verifies integer modulus operation works correctly and handles modulus by zero
// Tests both successful modulus (10 % 3 = 1) and error case (modulus by zero)
func TestMod(t *testing.T) {
	calc := New()

	// Test normal modulus case
	result, err := calc.Mod(10, 3)
	if err != nil || result != 1 {
		t.Errorf("Expected 1, got %d (err: %v)", result, err)
	}

	// Test modulus by zero should return an error
	_, err = calc.Mod(10, 0)
	if err == nil {
		t.Error("Expected modulus by zero error")
	}
}

// TestModFloat verifies floating-point modulus operation works correctly
// Tests both successful float modulus (10.5 % 3.0 = 1.5) and error case (modulus by zero)
// Uses floatEquals helper due to floating-point precision considerations
func TestModFloat(t *testing.T) {
	calc := New()

	// Test normal floating-point modulus case
	result, err := calc.ModFloat(10.5, 3.0)
	if err != nil || !floatEquals(result, 1.5, 1e-9) {
		t.Errorf("Expected 1.5, got %f (err: %v)", result, err)
	}

	// Test floating-point modulus by zero should return an error
	_, err = calc.ModFloat(10.0, 0.0)
	if err == nil {
		t.Error("Expected modulus by zero error")
	}
}

// =============================================================================
// ADVANCED MATHEMATICAL OPERATIONS TESTS
// These tests verify power and square root operations work correctly
// =============================================================================

// TestPower verifies exponentiation works for both positive and negative exponents
// Tests: 2^3 = 8 and 2^(-2) = 0.25
func TestPower(t *testing.T) {
	calc := New()

	// Test positive exponent
	result := calc.Power(2, 3)
	if result != 8 {
		t.Errorf("Expected 8, got %f", result)
	}

	// Test negative exponent (should produce fractional result)
	result = calc.Power(2, -2)
	if !floatEquals(result, 0.25, 1e-9) {
		t.Errorf("Expected 0.25, got %f", result)
	}
}

// TestSqrt verifies square root calculation and proper error handling for negative inputs
// Tests both successful sqrt(16) = 4 and error case (sqrt of negative number)
func TestSqrt(t *testing.T) {
	calc := New()

	// Test normal square root case
	result, err := calc.Sqrt(16)
	if err != nil || result != 4 {
		t.Errorf("Expected 4, got %f (err: %v)", result, err)
	}

	// Test square root of negative number should return an error
	_, err = calc.Sqrt(-1)
	if err == nil {
		t.Error("Expected square root of negative number error")
	}
}

// =============================================================================
// FLOATING-POINT PRECISION TESTS
// These tests verify that floating-point arithmetic handles precision correctly
// =============================================================================

// TestFloatingPointDivision verifies that division resulting in repeating decimals
// is handled correctly with appropriate precision tolerance
// Tests: 1/3 ≈ 0.333333 (within tolerance)
func TestFloatingPointDivision(t *testing.T) {
	calc := New()
	result, err := calc.Divide(1.0, 3.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Use tolerance-based comparison for floating-point results
	if !floatEquals(result, 0.333333, 1e-5) {
		t.Errorf("Expected ~0.33333, got %f", result)
	}
}

// =============================================================================
// EDGE CASE TESTS - ADDITION
// These tests verify addition handles special cases like zero values and
// negative numbers correctly
// =============================================================================

// TestAddEdgeCases verifies addition works correctly with edge cases
// Tests zero addition and negative number cancellation
func TestAddEdgeCases(t *testing.T) {
	calc := New()

	// Test adding zero to zero
	if result := calc.Add(0, 0); result != 0 {
		t.Errorf("Expected 0 + 0 = 0, got %f", result)
	}

	// Test adding negative and positive numbers that cancel out
	if result := calc.Add(-5, 5); result != 0 {
		t.Errorf("Expected -5 + 5 = 0, got %f", result)
	}
}

// =============================================================================
// EDGE CASE TESTS - SUBTRACTION
// These tests verify subtraction handles special cases correctly
// =============================================================================

// TestSubtractEdgeCases verifies subtraction works correctly with edge cases
// Tests zero subtraction and subtracting negative numbers
func TestSubtractEdgeCases(t *testing.T) {
	calc := New()

	// Test subtracting zero from zero
	if result := calc.Subtract(0, 0); result != 0 {
		t.Errorf("Expected 0 - 0 = 0, got %f", result)
	}

	// Test subtracting a negative number from itself (double negative)
	if result := calc.Subtract(-5, -5); result != 0 {
		t.Errorf("Expected -5 - (-5) = 0, got %f", result)
	}
}

// =============================================================================
// EDGE CASE TESTS - MULTIPLICATION
// These tests verify multiplication handles special cases like zero and infinity
// =============================================================================

// TestMultiplyEdgeCases verifies multiplication works correctly with edge cases
// Tests multiplication by zero (should always result in zero)
func TestMultiplyEdgeCases(t *testing.T) {
	calc := New()

	// Test that zero times any number (even very large) equals zero
	if result := calc.Multiply(0, math.MaxFloat64); result != 0 {
		t.Errorf("Expected 0 * MaxFloat64 = 0, got %f", result)
	}
}

// =============================================================================
// EDGE CASE TESTS - DIVISION
// These tests verify division handles special cases like zero and infinity
// =============================================================================

// TestDivideEdgeCases verifies division works correctly with edge cases
// Tests dividing zero and dividing by infinity
func TestDivideEdgeCases(t *testing.T) {
	calc := New()

	// Test dividing zero by any non-zero number should equal zero
	if result, err := calc.Divide(0, 1); err != nil || result != 0 {
		t.Errorf("Expected 0 / 1 = 0, got %f (err: %v)", result, err)
	}

	// Test dividing by positive infinity should equal zero
	if result, err := calc.Divide(1, math.Inf(1)); err != nil || result != 0 {
		t.Errorf("Expected 1 / +Inf = 0, got %f (err: %v)", result, err)
	}
}

// =============================================================================
// EDGE CASE TESTS - MODULUS WITH NEGATIVE OPERANDS
// These tests verify modulus operations work correctly with negative numbers
// =============================================================================

// TestModNegativeOperands verifies integer modulus handles negative operands correctly
// Tests the mathematical behavior of modulus with negative numbers
func TestModNegativeOperands(t *testing.T) {
	calc := New()

	// Test negative dividend: -10 % 3 = -1 (result has same sign as dividend)
	if result, err := calc.Mod(-10, 3); err != nil || result != -1 {
		t.Errorf("Expected -10 %% 3 = -1, got %d (err: %v)", result, err)
	}

	// Test negative divisor: 10 % -3 = 1 (result has same sign as dividend)
	if result, err := calc.Mod(10, -3); err != nil || result != 1 {
		t.Errorf("Expected 10 %% -3 = 1, got %d (err: %v)", result, err)
	}
}

// =============================================================================
// EDGE CASE TESTS - FLOATING-POINT MODULUS WITH NEGATIVE OPERANDS
// These tests verify floating-point modulus operations work correctly with negative numbers
// =============================================================================

// TestModFloatEdgeCases verifies floating-point modulus handles negative operands correctly
// Tests the mathematical behavior of floating-point modulus with negative numbers
func TestModFloatEdgeCases(t *testing.T) {
	calc := New()

	// Test negative dividend: -10.5 % 3.0 = -1.5 (result has same sign as dividend)
	if result, err := calc.ModFloat(-10.5, 3.0); err != nil || !floatEquals(result, -1.5, 1e-9) {
		t.Errorf("Expected -10.5 %% 3.0 = -1.5, got %f (err: %v)", result, err)
	}

	// Test negative divisor: 10.5 % -3.0 = 1.5 (result has same sign as dividend)
	if result, err := calc.ModFloat(10.5, -3.0); err != nil || !floatEquals(result, 1.5, 1e-9) {
		t.Errorf("Expected 10.5 %% -3.0 = 1.5, got %f (err: %v)", result, err)
	}
}

// =============================================================================
// EDGE CASE TESTS - POWER OPERATIONS
// These tests verify power operations handle mathematical edge cases correctly
// =============================================================================

// TestPowerEdgeCases verifies exponentiation works correctly with edge cases
// Tests special mathematical cases like 0^0, negative bases, and infinity
func TestPowerEdgeCases(t *testing.T) {
	calc := New()

	// Test 0^0 = 1 (mathematical convention, though technically undefined)
	if result := calc.Power(0, 0); result != 1 {
		t.Errorf("Expected 0^0 = 1 (by convention), got %f", result)
	}

	// Test negative base with odd exponent: -2^3 = -8
	if result := calc.Power(-2, 3); result != -8 {
		t.Errorf("Expected -2^3 = -8, got %f", result)
	}

	// Test negative base with even exponent: -2^2 = 4
	if result := calc.Power(-2, 2); result != 4 {
		t.Errorf("Expected -2^2 = 4, got %f", result)
	}

	// Test infinity to any positive power should remain infinity
	if result := calc.Power(math.Inf(1), 1); !math.IsInf(result, 1) {
		t.Errorf("Expected +Inf^1 = +Inf, got %f", result)
	}
}

// =============================================================================
// EDGE CASE TESTS - SQUARE ROOT OPERATIONS
// These tests verify square root operations handle edge cases correctly
// =============================================================================

// TestSqrtEdgeCases verifies square root works correctly with edge cases
// Tests the boundary case of sqrt(0) = 0
func TestSqrtEdgeCases(t *testing.T) {
	calc := New()

	// Test square root of zero should equal zero
	if result, err := calc.Sqrt(0); err != nil || result != 0 {
		t.Errorf("Expected sqrt(0) = 0, got %f (err: %v)", result, err)
	}
}
