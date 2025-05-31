// main_test.go
package main

import (
	"math"
	"testing"

	"github.com/jondkelley/cicd_golang_calculator/internal/calculator"
)

func TestEvaluateExpression(t *testing.T) {
	calc := calculator.New()

	tests := []struct {
		expr     string
		expected float64
		err      bool
	}{
		{"2 + 2", 4, false},
		{"5 - 3", 2, false},
		{"3 * 4", 12, false},
		{"8 / 2", 4, false},
		{"sqrt(16)", 4, false},
		{"10 % 3", 1, false},
		{"2 ^ 3", 8, false},
		{"sqrt(-1)", 0, true},
		{"10 / 0", 0, true},
		{"abc", 0, true},
		{"5 /", 0, true},
	}

	for _, test := range tests {
		result, err := evaluateExpression(calc, test.expr)
		if test.err && err == nil {
			t.Errorf("Expected error for %q, got none", test.expr)
		}
		if !test.err && math.Abs(result-test.expected) > 1e-6 {
			t.Errorf("Expected %v for %q, got %v", test.expected, test.expr, result)
		}
	}
}
