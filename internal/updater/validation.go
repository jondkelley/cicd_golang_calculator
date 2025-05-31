package updater

import (
	"fmt"
	"os"
	"os/exec"
)

// ValidateDownloadedFile checks if the downloaded file is a valid executable
func ValidateDownloadedFile(filepath string) bool {
	if !IsValidExecutable(filepath) {
		fmt.Println("Downloaded file does not appear to be a valid executable")
		return false
	}
	return true
}

// IsValidExecutable checks if the file appears to be a valid executable by examining magic bytes
func IsValidExecutable(filepath string) bool {
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
	return checkELF(header) || checkMachO(header) || checkPE(header)
}

// checkELF checks for ELF (Linux) magic bytes: 0x7F 'E' 'L' 'F'
func checkELF(header []byte) bool {
	return len(header) >= 4 &&
		header[0] == 0x7F &&
		header[1] == 'E' &&
		header[2] == 'L' &&
		header[3] == 'F'
}

// checkMachO checks for Mach-O (macOS) magic bytes
func checkMachO(header []byte) bool {
	if len(header) < 4 {
		return false
	}

	// Mach-O 32-bit: 0xFE 0xED 0xFA 0xCE or 0xCE 0xFA 0xED 0xFE
	if (header[0] == 0xFE && header[1] == 0xED && header[2] == 0xFA && header[3] == 0xCE) ||
		(header[0] == 0xCE && header[1] == 0xFA && header[2] == 0xED && header[3] == 0xFE) {
		return true
	}

	// Mach-O 64-bit: 0xFE 0xED 0xFA 0xCF or 0xCF 0xFA 0xED 0xFE
	if (header[0] == 0xFE && header[1] == 0xED && header[2] == 0xFA && header[3] == 0xCF) ||
		(header[0] == 0xCF && header[1] == 0xFA && header[2] == 0xED && header[3] == 0xFE) {
		return true
	}

	return false
}

// checkPE checks for PE (Windows) magic bytes: 'M' 'Z' at start
func checkPE(header []byte) bool {
	return len(header) >= 2 && header[0] == 'M' && header[1] == 'Z'
}

// TestNewBinary runs a quick test on the downloaded binary to ensure it works
func TestNewBinary(binaryPath string) bool {
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
