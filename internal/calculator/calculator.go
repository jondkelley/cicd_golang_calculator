// Package calculator provides basic arithmetic operations with proper error handling.
// It supports addition, subtraction, multiplication, division, modulus, power, and square root operations.
package calculator

import (
	"errors"
	"math"
)

// Calculator represents a calculator that can perform basic arithmetic operations
type Calculator struct{}

// New creates and returns a new Calculator instance
func New() *Calculator {
	return &Calculator{}
}

// Add performs addition of two float64 numbers and returns the result
func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}

// Subtract performs subtraction of two float64 numbers and returns the result
func (c *Calculator) Subtract(a, b float64) float64 {
	return a - b
}

// Multiply performs multiplication of two float64 numbers and returns the result
func (c *Calculator) Multiply(a, b float64) float64 {
	return a * b
}

// Divide performs division of two float64 numbers and returns the result with error handling for division by zero
func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0.0 {
		return 0.0, errors.New("division by zero")
	}
	return a / b, nil
}

// Mod performs modulus operation on two integers and returns the result with error handling for modulus by zero
func (c *Calculator) Mod(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("modulus by zero")
	}
	return a % b, nil
}

// ModFloat performs modulus operation on float64 numbers and returns the result with error handling for modulus by zero
func (c *Calculator) ModFloat(a, b float64) (float64, error) {
	if b == 0.0 {
		return 0.0, errors.New("modulus by zero")
	}
	return math.Mod(a, b), nil
}

// Power performs exponentiation of base raised to the power of exponent and returns the result
func (c *Calculator) Power(base, exponent float64) float64 {
	return math.Pow(base, exponent)
}

// Sqrt performs square root operation on a float64 number and returns the result with error handling for negative numbers
func (c *Calculator) Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0.0, errors.New("square root of negative number")
	}
	return math.Sqrt(a), nil
}
