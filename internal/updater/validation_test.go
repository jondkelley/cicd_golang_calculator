package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDownloadedFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid executable file", func(t *testing.T) {
		// Create a file with ELF magic bytes
		validExecPath := filepath.Join(tempDir, "valid_exec")
		elfHeader := []byte{0x7F, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00}
		err := os.WriteFile(validExecPath, elfHeader, 0755)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := ValidateDownloadedFile(validExecPath)
		if !result {
			t.Error("Expected ValidateDownloadedFile to return true for valid executable")
		}
	})

	t.Run("invalid file", func(t *testing.T) {
		// Create a file with random bytes (not executable)
		invalidPath := filepath.Join(tempDir, "invalid_file")
		invalidData := []byte{0x12, 0x34, 0x56, 0x78}
		err := os.WriteFile(invalidPath, invalidData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := ValidateDownloadedFile(invalidPath)
		if result {
			t.Error("Expected ValidateDownloadedFile to return false for invalid file")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		nonexistentPath := filepath.Join(tempDir, "nonexistent")
		result := ValidateDownloadedFile(nonexistentPath)
		if result {
			t.Error("Expected ValidateDownloadedFile to return false for nonexistent file")
		}
	})
}

func TestIsValidExecutable(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		header   []byte
		expected bool
	}{
		{
			name:     "ELF executable",
			header:   []byte{0x7F, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: true,
		},
		{
			name:     "PE executable (Windows)",
			header:   []byte{'M', 'Z', 0x90, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00},
			expected: true,
		},
		{
			name:     "Mach-O 32-bit",
			header:   []byte{0xFE, 0xED, 0xFA, 0xCE, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x20, 0x00},
			expected: true,
		},
		{
			name:     "Mach-O 64-bit",
			header:   []byte{0xFE, 0xED, 0xFA, 0xCF, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x20, 0x00},
			expected: true,
		},
		{
			name:     "Text file",
			header:   []byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!', 0x0A, 0x00, 0x00, 0x00},
			expected: false,
		},
		{
			name:     "Empty file",
			header:   []byte{},
			expected: false,
		},
		{
			name:     "Short file",
			header:   []byte{0x7F, 'E'},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, tt.name)
			err := os.WriteFile(testFile, tt.header, 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result := IsValidExecutable(testFile)
			if result != tt.expected {
				t.Errorf("IsValidExecutable() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCheckELF(t *testing.T) {
	tests := []struct {
		name     string
		header   []byte
		expected bool
	}{
		{"valid ELF", []byte{0x7F, 'E', 'L', 'F'}, true},
		{"invalid magic", []byte{0x7F, 'E', 'L', 'X'}, false},
		{"too short", []byte{0x7F, 'E', 'L'}, false},
		{"empty", []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkELF(tt.header)
			if result != tt.expected {
				t.Errorf("checkELF() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCheckMachO(t *testing.T) {
	tests := []struct {
		name     string
		header   []byte
		expected bool
	}{
		{"Mach-O 32-bit big endian", []byte{0xFE, 0xED, 0xFA, 0xCE}, true},
		{"Mach-O 32-bit little endian", []byte{0xCE, 0xFA, 0xED, 0xFE}, true},
		{"Mach-O 64-bit big endian", []byte{0xFE, 0xED, 0xFA, 0xCF}, true},
		{"Mach-O 64-bit little endian", []byte{0xCF, 0xFA, 0xED, 0xFE}, true},
		{"invalid magic", []byte{0xFE, 0xED, 0xFA, 0xFF}, false},
		{"too short", []byte{0xFE, 0xED, 0xFA}, false},
		{"empty", []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkMachO(tt.header)
			if result != tt.expected {
				t.Errorf("checkMachO() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCheckPE(t *testing.T) {
	tests := []struct {
		name     string
		header   []byte
		expected bool
	}{
		{"valid PE", []byte{'M', 'Z'}, true},
		{"invalid magic", []byte{'M', 'X'}, false},
		{"too short", []byte{'M'}, false},
		{"empty", []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkPE(tt.header)
			if result != tt.expected {
				t.Errorf("checkPE() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
