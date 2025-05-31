// main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jondkelley/cicd_golang_calculator/internal/calculator"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime" // to put golang version inside
	"strconv"
	"strings"
	"syscall"
	"time"
)

var version = "0.0.0-local"

const manifestURL = "https://raw.githubusercontent.com/jondkelley/cicd_golang_calculator/main/version.json"

// Release represents a single software release with version information and download URLs
type Release struct {
	Version     string            `json:"version"`
	URLs        map[string]string `json:"urls"`
	IsAlpha     bool              `json:"isAlpha"`
	ReleaseDate string            `json:"releaseDate"`
}

// VersionManifest contains a collection of software releases
type VersionManifest struct {
	Releases []Release `json:"releases"`
}

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--version" || arg == "-v" {
			fmt.Printf("cicd_golang_calculator %s (%s)\n", version, runtime.Version())
			return
		}
	}

	fmt.Printf("cicd_golang_calculator %s (%s)\n", version, runtime.Version())
	checkForUpdate()
	fmt.Println("Interactive Calculator - Enter expressions like:")
	fmt.Println(`  3 + 4`)
	fmt.Println(`  5 * 6`)
	fmt.Println(`  10 / 2`)
	fmt.Println(`  sqrt(16)`)
	fmt.Println("Supported operators: + - * / % ^ sqrt()")
	fmt.Println("Type Ctrl+C to exit.")

	// Setup signal handling to exit on Ctrl+C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nExiting.")
		os.Exit(0)
	}()

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

func checkForUpdate() {
	fmt.Print("Checking for updates... ")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(manifestURL)
	if err != nil {
		fmt.Println("🚨 WARNING: no version.json manifest found in project repository.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("🚨 WARNING: no version.json manifest found in project repository (status code = %d)\n", resp.StatusCode)
		return
	}

	var manifest VersionManifest
	err = json.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		fmt.Println("malformed version.json manifest found.")
		return
	}

	if len(manifest.Releases) == 0 {
		fmt.Println("no releases found in manifest.")
		return
	}

	// Find the latest eligible release
	latestRelease := findLatestEligibleRelease(manifest.Releases)
	if latestRelease == nil {
		fmt.Println("everything is up to date!")
		return
	}

	// Check if we're already on the latest version
	if latestRelease.Version == version {
		fmt.Println("you are running latest version.")
		return
	}

	// Show update prompt
	alphaWarning := ""
	if latestRelease.IsAlpha {
		alphaWarning = " (ALPHA RELEASE)"
	}

	fmt.Printf("new version %s%s available, would you like to update? Y[es]/N[o]: ", latestRelease.Version, alphaWarning)
	var input string
	fmt.Scanln(&input)

	if input == "y" || input == "Y" || input == "yes" || input == "Yes" {
		fmt.Printf("Updating current version from %s to %s\n", version, latestRelease.Version)
		updateBinary(*latestRelease)
	} else {
		fmt.Println("update cancelled.")
	}
}

// findLatestEligibleRelease returns the latest release that the user is allowed to install
func findLatestEligibleRelease(releases []Release) *Release {
	allowAlpha := os.Getenv("CALC_ALLOW_ALPHA") != ""

	for _, release := range releases {
		// Skip alpha releases unless explicitly allowed
		if release.IsAlpha && !allowAlpha {
			continue
		}

		// Return the first (latest) eligible release
		return &release
	}

	return nil
}

func updateBinary(release Release) {
	// Detect current platform
	platform := runtime.GOOS

	url, exists := release.URLs[platform]
	if !exists {
		fmt.Printf("No binary available for platform: %s\n", platform)
		return
	}

	// Create backup before update
	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Could not locate current executable:", err)
		return
	}

	backupPath := execPath + ".bak"
	if err := copyFile(execPath, backupPath); err != nil {
		fmt.Println("Failed to create backup:", err)
		return
	}

	// Download new version
	alphaLabel := ""
	if release.IsAlpha {
		alphaLabel = " (alpha)"
	}
	fmt.Printf(" * Downloading %s binary%s from: %s\n", platform, alphaLabel, url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(" ⚠️  Failed to download update:", err)
		return
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != 200 {
		fmt.Printf(" ⚠️  Failed to download update: HTTP %d - %s\n", resp.StatusCode, resp.Status)
		return
	}

	// Check Content-Type header
	contentType := resp.Header.Get("Content-Type")
	fmt.Printf("Downloaded content type: %s\n", contentType)

	// Validate content type (binary should be application/octet-stream or similar)
	validContentTypes := []string{
		"application/octet-stream",
		"application/x-executable",
		"application/x-binary",
		"binary/octet-stream",
	}

	isValidContentType := false
	for _, validType := range validContentTypes {
		if strings.Contains(contentType, validType) {
			isValidContentType = true
			break
		}
	}

	// GitHub releases might serve as text/plain, so we'll be more lenient
	// but still check file size and magic bytes
	if !isValidContentType && !strings.Contains(contentType, "text/plain") {
		fmt.Printf("Warning: Unexpected content type: %s\n", contentType)
	}

	tmpPath := execPath + ".new"
	out, err := os.Create(tmpPath)
	if err != nil {
		fmt.Println("Could not create temporary file:", err)
		return
	}
	defer out.Close()

	// Copy the response body while keeping track of size
	bytesWritten, err := io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Failed writing new binary:", err)
		os.Remove(tmpPath) // Clean up
		return
	}
	out.Close() // Close file before validation

	fmt.Printf("Downloaded %d bytes\n", bytesWritten)

	// Validate minimum file size (executables are typically > 1KB)
	if bytesWritten < 1024 {
		fmt.Printf("Downloaded file too small (%d bytes), likely not a binary executable\n", bytesWritten)
		os.Remove(tmpPath)
		return
	}

	// Check magic bytes to validate it's an executable
	if !isValidExecutable(tmpPath) {
		fmt.Println("Downloaded file does not appear to be a valid executable")
		os.Remove(tmpPath)
		return
	}

	// Make executable
	err = os.Chmod(tmpPath, 0755)
	if err != nil {
		fmt.Println("Failed to set executable permission:", err)
		os.Remove(tmpPath)
		return
	}

	// Optional: Test run the new binary with --version flag to verify it works
	if !testNewBinary(tmpPath) {
		fmt.Println("New binary failed validation test")
		os.Remove(tmpPath)
		return
	}

	// Replace binary in-place
	err = os.Rename(tmpPath, execPath)
	if err != nil {
		fmt.Println("Failed to overwrite binary:", err)
		os.Remove(tmpPath)
		return
	}

	fmt.Printf("✅ Update complete from %s to %s. Please re-run the application.\n", version, release.Version)
	os.Exit(0)
}

// isValidExecutable checks if the file appears to be a valid executable
func isValidExecutable(filepath string) bool {
	file, err := os.Open(filepath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first few bytes to check magic numbers
	header := make([]byte, 16)
	n, err := file.Read(header)
	if err != nil || n < 4 {
		return false
	}

	// Check for common executable magic bytes
	// ELF (Linux): 0x7F 'E' 'L' 'F'
	if len(header) >= 4 && header[0] == 0x7F && header[1] == 'E' && header[2] == 'L' && header[3] == 'F' {
		return true
	}

	// Mach-O (macOS): 0xFE 0xED 0xFA 0xCE or 0xCE 0xFA 0xED 0xFE
	if len(header) >= 4 {
		if (header[0] == 0xFE && header[1] == 0xED && header[2] == 0xFA && header[3] == 0xCE) ||
			(header[0] == 0xCE && header[1] == 0xFA && header[2] == 0xED && header[3] == 0xFE) {
			return true
		}
	}

	// Mach-O 64-bit: 0xFE 0xED 0xFA 0xCF or 0xCF 0xFA 0xED 0xFE
	if len(header) >= 4 {
		if (header[0] == 0xFE && header[1] == 0xED && header[2] == 0xFA && header[3] == 0xCF) ||
			(header[0] == 0xCF && header[1] == 0xFA && header[2] == 0xED && header[3] == 0xFE) {
			return true
		}
	}

	// PE (Windows): 'M' 'Z' at start, then PE signature later
	if len(header) >= 2 && header[0] == 'M' && header[1] == 'Z' {
		return true
	}

	return false
}

// testNewBinary runs a quick test on the downloaded binary to ensure it works
func testNewBinary(binaryPath string) bool {
	// Try to run the binary with --version flag
	cmd := exec.Command(binaryPath, "--version")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Binary test failed: %v\n", err)
		return false
	}

	fmt.Println("✅ Binary validation test passed")
	return true
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

var operatorRE = regexp.MustCompile(`^\s*([-+]?[0-9]*\.?[0-9]+)\s*([+\-*/^%])\s*([-+]?[0-9]*\.?[0-9]+)\s*$`)
var sqrtRE = regexp.MustCompile(`(?i)^sqrt\(\s*([-+]?[0-9]*\.?[0-9]+)\s*\)$`)

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
