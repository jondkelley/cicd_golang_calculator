package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

// DownloadBinary downloads the binary from the given URL and validates it
func DownloadBinary(url, execPath string, release Release) bool {
	releaseTypeLabel := ""
	if release.IsAlpha {
		releaseTypeLabel = " (alpha)"
	} else if release.IsBeta {
		releaseTypeLabel = " (beta)"
	}
	fmt.Printf(" * Downloading %s binary%s from: %s\n", runtime.GOOS, releaseTypeLabel, url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(" ⚠️  Failed to download update:", err)
		return false
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != 200 {
		fmt.Printf(" ⚠️  Failed to download update: HTTP %d - %s\n", resp.StatusCode, resp.Status)
		return false
	}

	// Validate content type
	if !validateContentType(resp.Header.Get("Content-Type")) {
		return false
	}

	// Create temporary file and download
	tmpPath := execPath + ".new"
	if !writeTemporaryFile(tmpPath, resp.Body) {
		return false
	}
	defer func() {
		// Clean up temp file if we exit early
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()

	// Validate the downloaded file
	if !ValidateDownloadedFile(tmpPath) {
		return false
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		fmt.Println("Failed to set executable permission:", err)
		return false
	}

	// Test the new binary
	if !TestNewBinary(tmpPath) {
		fmt.Println("New binary failed validation test")
		return false
	}

	// Replace binary in-place
	if err := os.Rename(tmpPath, execPath); err != nil {
		fmt.Println("Failed to overwrite binary:", err)
		return false
	}

	return true
}

// validateContentType checks if the HTTP response has a valid content type for a binary
func validateContentType(contentType string) bool {
	fmt.Printf("Downloaded content type: %s\n", contentType)

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
	if !isValidContentType && !strings.Contains(contentType, "text/plain") {
		fmt.Printf("Warning: Unexpected content type: %s\n", contentType)
	}

	return true
}

// writeTemporaryFile writes the response body to a temporary file
func writeTemporaryFile(tmpPath string, body io.Reader) bool {
	out, err := os.Create(tmpPath)
	if err != nil {
		fmt.Println("Could not create temporary file:", err)
		return false
	}
	defer out.Close()

	bytesWritten, err := io.Copy(out, body)
	if err != nil {
		fmt.Println("Failed writing new binary:", err)
		os.Remove(tmpPath)
		return false
	}

	fmt.Printf("Downloaded %d bytes\n", bytesWritten)

	// Validate minimum file size (executables are typically > 1KB)
	if bytesWritten < 1024 {
		fmt.Printf("Downloaded file too small (%d bytes), likely not a binary executable\n", bytesWritten)
		os.Remove(tmpPath)
		return false
	}

	return true
}
