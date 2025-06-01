package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	t.Run("successful copy", func(t *testing.T) {
		// Create source file
		srcPath := filepath.Join(tempDir, "source.txt")
		srcContent := "Hello, World!"
		err := os.WriteFile(srcPath, []byte(srcContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Copy file
		dstPath := filepath.Join(tempDir, "destination.txt")
		err = CopyFile(srcPath, dstPath)
		if err != nil {
			t.Fatalf("CopyFile failed: %v", err)
		}

		// Verify destination file exists and has correct content
		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}

		if string(dstContent) != srcContent {
			t.Errorf("Content mismatch: expected %q, got %q", srcContent, string(dstContent))
		}
	})

	t.Run("preserve permissions", func(t *testing.T) {
		// Create source file with specific permissions
		srcPath := filepath.Join(tempDir, "source_perms.txt")
		err := os.WriteFile(srcPath, []byte("test content"), 0755)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Copy file
		dstPath := filepath.Join(tempDir, "destination_perms.txt")
		err = CopyFile(srcPath, dstPath)
		if err != nil {
			t.Fatalf("CopyFile failed: %v", err)
		}

		// Check permissions are preserved
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			t.Fatalf("Failed to stat source file: %v", err)
		}

		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("Failed to stat destination file: %v", err)
		}

		if srcInfo.Mode() != dstInfo.Mode() {
			t.Errorf("Permissions not preserved: source %v, destination %v",
				srcInfo.Mode(), dstInfo.Mode())
		}
	})

	t.Run("copy large file", func(t *testing.T) {
		// Create a larger source file
		srcPath := filepath.Join(tempDir, "large_source.txt")
		largeContent := make([]byte, 10*1024*1024) // 10 MB
		for i := range largeContent {
			largeContent[i] = byte(i % 256)
		}
		err := os.WriteFile(srcPath, largeContent, 0644)
		if err != nil {
			t.Fatalf("Failed to create large source file: %v", err)
		}

		// Copy file
		dstPath := filepath.Join(tempDir, "large_destination.txt")
		err = CopyFile(srcPath, dstPath)
		if err != nil {
			t.Fatalf("CopyFile failed for large file: %v", err)
		}

		// Verify content matches
		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read large destination file: %v", err)
		}

		if len(dstContent) != len(largeContent) {
			t.Errorf("Size mismatch: expected %d, got %d", len(largeContent), len(dstContent))
		}

		for i, b := range dstContent {
			if b != largeContent[i] {
				t.Errorf("Content mismatch at byte %d: expected %d, got %d", i, largeContent[i], b)
				break
			}
		}
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		// Create source file
		srcPath := filepath.Join(tempDir, "overwrite_source.txt")
		srcContent := "new content"
		err := os.WriteFile(srcPath, []byte(srcContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Create existing destination file with different content
		dstPath := filepath.Join(tempDir, "overwrite_destination.txt")
		err = os.WriteFile(dstPath, []byte("old content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create existing destination file: %v", err)
		}

		// Copy file (should overwrite)
		err = CopyFile(srcPath, dstPath)
		if err != nil {
			t.Fatalf("CopyFile failed when overwriting: %v", err)
		}

		// Verify destination has new content
		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read overwritten destination file: %v", err)
		}

		if string(dstContent) != srcContent {
			t.Errorf("Overwrite failed: expected %q, got %q", srcContent, string(dstContent))
		}
	})
}

func TestCopyFileErrors(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("source file does not exist", func(t *testing.T) {
		srcPath := filepath.Join(tempDir, "nonexistent.txt")
		dstPath := filepath.Join(tempDir, "destination.txt")

		err := CopyFile(srcPath, dstPath)
		if err == nil {
			t.Error("Expected error when source file doesn't exist, got nil")
		}
	})

	t.Run("destination directory does not exist", func(t *testing.T) {
		// Create source file
		srcPath := filepath.Join(tempDir, "source.txt")
		err := os.WriteFile(srcPath, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Try to copy to non-existent directory
		dstPath := filepath.Join(tempDir, "nonexistent", "destination.txt")

		err = CopyFile(srcPath, dstPath)
		if err == nil {
			t.Error("Expected error when destination directory doesn't exist, got nil")
		}
	})

	t.Run("source is directory", func(t *testing.T) {
		// Create source directory
		srcPath := filepath.Join(tempDir, "source_dir")
		err := os.Mkdir(srcPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create source directory: %v", err)
		}

		dstPath := filepath.Join(tempDir, "destination.txt")

		err = CopyFile(srcPath, dstPath)
		if err == nil {
			t.Error("Expected error when source is directory, got nil")
		}
	})
}
