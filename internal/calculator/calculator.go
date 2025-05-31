// calculator.go
package calculator

import (
	"errors"
	"math"
)

type Calculator struct{}

func New() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}

func (c *Calculator) Subtract(a, b float64) float64 {
	return a - b
}

func (c *Calculator) Multiply(a, b float64) float64 {
	return a * b
}

func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0.0 {
		return 0.0, errors.New("division by zero")
	}
	return a / b, nil
}

func (c *Calculator) Mod(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("modulus by zero")
	}
	return a % b, nil
}

// ModFloat performs modulus operation on float64
func (c *Calculator) ModFloat(a, b float64) (float64, error) {
	if b == 0.0 {
		return 0.0, errors.New("modulus by zero")
	}
	return math.Mod(a, b), nil
}

func (c *Calculator) Power(base, exponent float64) float64 {
	return math.Pow(base, exponent)
}

func (c *Calculator) Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0.0, errors.New("square root of negative number")
	}
	return math.Sqrt(a), nil
}
