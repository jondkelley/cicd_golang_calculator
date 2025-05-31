// calculator_test.go
package calculator

import (
	"math"
	"testing"
)

func floatEquals(a, b float64, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestAdd(t *testing.T) {
	calc := New()
	result := calc.Add(2, 3)
	if result != 5 {
		t.Errorf("Expected 5, got %f", result)
	}
}

func TestSubtract(t *testing.T) {
	calc := New()
	result := calc.Subtract(5, 3)
	if result != 2 {
		t.Errorf("Expected 2, got %f", result)
	}
}

func TestMultiply(t *testing.T) {
	calc := New()
	result := calc.Multiply(3, 4)
	if result != 12 {
		t.Errorf("Expected 12, got %f", result)
	}
}

func TestDivide(t *testing.T) {
	calc := New()

	result, err := calc.Divide(10, 2)
	if err != nil || result != 5 {
		t.Errorf("Expected 5, got %f (err: %v)", result, err)
	}

	_, err = calc.Divide(10, 0)
	if err == nil {
		t.Error("Expected division by zero error")
	}
}

func TestMod(t *testing.T) {
	calc := New()

	result, err := calc.Mod(10, 3)
	if err != nil || result != 1 {
		t.Errorf("Expected 1, got %d (err: %v)", result, err)
	}

	_, err = calc.Mod(10, 0)
	if err == nil {
		t.Error("Expected modulus by zero error")
	}
}

func TestModFloat(t *testing.T) {
	calc := New()

	result, err := calc.ModFloat(10.5, 3.0)
	if err != nil || !floatEquals(result, 1.5, 1e-9) {
		t.Errorf("Expected 1.5, got %f (err: %v)", result, err)
	}

	_, err = calc.ModFloat(10.0, 0.0)
	if err == nil {
		t.Error("Expected modulus by zero error")
	}
}

func TestPower(t *testing.T) {
	calc := New()
	result := calc.Power(2, 3)
	if result != 8 {
		t.Errorf("Expected 8, got %f", result)
	}

	result = calc.Power(2, -2)
	if !floatEquals(result, 0.25, 1e-9) {
		t.Errorf("Expected 0.25, got %f", result)
	}
}

func TestSqrt(t *testing.T) {
	calc := New()
	result, err := calc.Sqrt(16)
	if err != nil || result != 4 {
		t.Errorf("Expected 4, got %f (err: %v)", result, err)
	}

	_, err = calc.Sqrt(-1)
	if err == nil {
		t.Error("Expected square root of negative number error")
	}
}

func TestFloatingPointDivision(t *testing.T) {
	calc := New()
	result, err := calc.Divide(1.0, 3.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !floatEquals(result, 0.333333, 1e-5) {
		t.Errorf("Expected ~0.33333, got %f", result)
	}
}

func TestAddEdgeCases(t *testing.T) {
	calc := New()

	if result := calc.Add(0, 0); result != 0 {
		t.Errorf("Expected 0 + 0 = 0, got %f", result)
	}

	if result := calc.Add(-5, 5); result != 0 {
		t.Errorf("Expected -5 + 5 = 0, got %f", result)
	}
}

func TestSubtractEdgeCases(t *testing.T) {
	calc := New()

	if result := calc.Subtract(0, 0); result != 0 {
		t.Errorf("Expected 0 - 0 = 0, got %f", result)
	}

	if result := calc.Subtract(-5, -5); result != 0 {
		t.Errorf("Expected -5 - (-5) = 0, got %f", result)
	}
}

func TestMultiplyEdgeCases(t *testing.T) {
	calc := New()

	if result := calc.Multiply(0, math.MaxFloat64); result != 0 {
		t.Errorf("Expected 0 * MaxFloat64 = 0, got %f", result)
	}
}

func TestDivideEdgeCases(t *testing.T) {
	calc := New()

	if result, err := calc.Divide(0, 1); err != nil || result != 0 {
		t.Errorf("Expected 0 / 1 = 0, got %f (err: %v)", result, err)
	}

	if result, err := calc.Divide(1, math.Inf(1)); err != nil || result != 0 {
		t.Errorf("Expected 1 / +Inf = 0, got %f (err: %v)", result, err)
	}
}

func TestModNegativeOperands(t *testing.T) {
	calc := New()

	if result, err := calc.Mod(-10, 3); err != nil || result != -1 {
		t.Errorf("Expected -10 %% 3 = -1, got %d (err: %v)", result, err)
	}

	if result, err := calc.Mod(10, -3); err != nil || result != 1 {
		t.Errorf("Expected 10 %% -3 = 1, got %d (err: %v)", result, err)
	}
}

func TestModFloatEdgeCases(t *testing.T) {
	calc := New()

	if result, err := calc.ModFloat(-10.5, 3.0); err != nil || !floatEquals(result, -1.5, 1e-9) {
		t.Errorf("Expected -10.5 %% 3.0 = -1.5, got %f (err: %v)", result, err)
	}

	if result, err := calc.ModFloat(10.5, -3.0); err != nil || !floatEquals(result, 1.5, 1e-9) {
		t.Errorf("Expected 10.5 %% -3.0 = 1.5, got %f (err: %v)", result, err)
	}
}

func TestPowerEdgeCases(t *testing.T) {
	calc := New()

	if result := calc.Power(0, 0); result != 1 {
		t.Errorf("Expected 0^0 = 1 (by convention), got %f", result)
	}

	if result := calc.Power(-2, 3); result != -8 {
		t.Errorf("Expected -2^3 = -8, got %f", result)
	}

	if result := calc.Power(-2, 2); result != 4 {
		t.Errorf("Expected -2^2 = 4, got %f", result)
	}

	if result := calc.Power(math.Inf(1), 1); !math.IsInf(result, 1) {
		t.Errorf("Expected +Inf^1 = +Inf, got %f", result)
	}
}

func TestSqrtEdgeCases(t *testing.T) {
	calc := New()

	if result, err := calc.Sqrt(0); err != nil || result != 0 {
		t.Errorf("Expected sqrt(0) = 0, got %f (err: %v)", result, err)
	}
}
