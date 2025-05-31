// main.go
package main

import (
	"bufio"
	"fmt"
	"github.com/jondkelley/cicd_golang_calculator/internal/calculator"
	"github.com/jondkelley/cicd_golang_calculator/internal/updater"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var version = "0.0.0-local"
var buildTime = "unknown"

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--version" || arg == "-v" {
			printVersion()
			return
		}
	}

	fmt.Printf("cicd_golang_calculator %s\n", version)

	// Check for updates using the updater package
	updater.CheckForUpdate(version, buildTime)

	printWelcomeMessage()
	setupSignalHandling()
	runCalculator()
}

func printVersion() {
	fmt.Printf("cicd_golang_calculator %s\n", version)
	fmt.Printf("Built: %s\n", buildTime)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func printWelcomeMessage() {
	fmt.Println("Interactive Calculator - Enter expressions like:")
	fmt.Println(`  3 + 4`)
	fmt.Println(`  5 * 6`)
	fmt.Println(`  10 / 2`)
	fmt.Println(`  sqrt(16)`)
	fmt.Println("Supported operators: + - * / % ^ sqrt()")
	fmt.Println("Type Ctrl+C to exit.")
}

func setupSignalHandling() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nExiting.")
		os.Exit(0)
	}()
}

func runCalculator() {
	calc := calculator.New()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		result, err := evaluateExpression(calc, line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("= %v\n", result)
		}
	}
}

// Expression evaluation logic
var (
	operatorRE = regexp.MustCompile(`^\s*([-+]?[0-9]*\.?[0-9]+)\s*([+\-*/^%])\s*([-+]?[0-9]*\.?[0-9]+)\s*$`)
	sqrtRE     = regexp.MustCompile(`(?i)^sqrt\(\s*([-+]?[0-9]*\.?[0-9]+)\s*\)$`)
)

func evaluateExpression(calc *calculator.Calculator, expr string) (float64, error) {
	if matches := operatorRE.FindStringSubmatch(expr); matches != nil {
		a, _ := strconv.ParseFloat(matches[1], 64)
		op := matches[2]
		b, _ := strconv.ParseFloat(matches[3], 64)

		switch op {
		case "+":
			return calc.Add(a, b), nil
		case "-":
			return calc.Subtract(a, b), nil
		case "*":
			return calc.Multiply(a, b), nil
		case "/":
			return calc.Divide(a, b)
		case "^":
			return calc.Power(a, b), nil
		case "%":
			return calc.ModFloat(a, b)
		default:
			return 0, fmt.Errorf("unsupported operator: %s", op)
		}
	} else if matches := sqrtRE.FindStringSubmatch(expr); matches != nil {
		a, _ := strconv.ParseFloat(matches[1], 64)
		return calc.Sqrt(a)
	}
	return 0, fmt.Errorf("invalid expression format")
}
