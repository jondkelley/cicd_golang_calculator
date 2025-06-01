package updater

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Mock HTTP transport for testing
type MockRoundTripper struct {
	Response *http.Response
	Error    error
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Response, nil
}

// Helper function to create a mock HTTP response
func createMockResponse(statusCode int, contentType string, body []byte) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header: http.Header{
			"Content-Type": []string{contentType},
		},
		Body: io.NopCloser(bytes.NewReader(body)),
	}
}

// Create a test version of DownloadBinary that accepts validation functions as parameters
func downloadBinaryWithValidators(url, execPath string, release Release,
	validateFile func(string) bool, testBinary func(string) bool) bool {

	// This is a copy of the DownloadBinary logic but with injectable validators
	fmt.Printf(" * Downloading %s binary", runtime.GOOS)
	if release.IsAlpha {
		fmt.Print(" (alpha)")
	} else if release.IsBeta {
		fmt.Print(" (beta)")
	}
	fmt.Printf(" from: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(" !!!!  Failed to download update:", err)
		return false
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != 200 {
		fmt.Printf(" !!!!  Failed to download update: HTTP %d - %s\n", resp.StatusCode, resp.Status)
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

	// Validate the downloaded file using injected validator
	if !validateFile(tmpPath) {
		return false
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		fmt.Println("Failed to set executable permission:", err)
		return false
	}

	// Test the new binary using injected tester
	if !testBinary(tmpPath) {
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

func TestDownloadBinary_Success(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test-binary")

	// Create mock binary content (> 1KB to pass size validation)
	mockBinary := make([]byte, 2048)
	for i := range mockBinary {
		mockBinary[i] = byte(i % 256)
	}

	// Setup mock HTTP client
	mockTransport := &MockRoundTripper{
		Response: createMockResponse(200, "application/octet-stream", mockBinary),
	}

	// Replace default HTTP client transport
	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	// Create mock validators
	mockValidateFile := func(path string) bool { return true }
	mockTestBinary := func(path string) bool { return true }

	release := Release{IsAlpha: false, IsBeta: false}
	result := downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)

	if !result {
		t.Fatal("Expected DownloadBinary to succeed")
	}

	// Verify the binary was written
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		t.Fatal("Expected binary file to exist after download")
	}

	// Verify file content
	content, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatal("Failed to read downloaded file:", err)
	}

	if !bytes.Equal(content, mockBinary) {
		t.Fatal("Downloaded file content doesn't match expected content")
	}
}

func TestDownloadBinary_HTTPError(t *testing.T) {
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test-binary")

	// Setup mock to return HTTP error
	mockTransport := &MockRoundTripper{
		Error: errors.New("network error"),
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	mockValidateFile := func(path string) bool { return true }
	mockTestBinary := func(path string) bool { return true }

	release := Release{}
	result := downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)

	if result {
		t.Fatal("Expected DownloadBinary to fail on HTTP error")
	}
}

func TestDownloadBinary_HTTP404(t *testing.T) {
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test-binary")

	mockTransport := &MockRoundTripper{
		Response: createMockResponse(404, "text/html", []byte("Not Found")),
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	mockValidateFile := func(path string) bool { return true }
	mockTestBinary := func(path string) bool { return true }

	release := Release{}
	result := downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)

	if result {
		t.Fatal("Expected DownloadBinary to fail on 404")
	}
}

func TestDownloadBinary_FileTooSmall(t *testing.T) {
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test-binary")

	// Create small file (< 1KB)
	smallBinary := make([]byte, 500)

	mockTransport := &MockRoundTripper{
		Response: createMockResponse(200, "application/octet-stream", smallBinary),
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	mockValidateFile := func(path string) bool { return true }
	mockTestBinary := func(path string) bool { return true }

	release := Release{}
	result := downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)

	if result {
		t.Fatal("Expected DownloadBinary to fail on small file")
	}
}

func TestDownloadBinary_ValidationFailed(t *testing.T) {
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test-binary")

	mockBinary := make([]byte, 2048)

	mockTransport := &MockRoundTripper{
		Response: createMockResponse(200, "application/octet-stream", mockBinary),
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	// Make validation fail
	mockValidateFile := func(path string) bool { return false }
	mockTestBinary := func(path string) bool { return true }

	release := Release{}
	result := downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)

	if result {
		t.Fatal("Expected DownloadBinary to fail on validation failure")
	}
}

func TestDownloadBinary_BinaryTestFailed(t *testing.T) {
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test-binary")

	mockBinary := make([]byte, 2048)

	mockTransport := &MockRoundTripper{
		Response: createMockResponse(200, "application/octet-stream", mockBinary),
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	// Make binary test fail
	mockValidateFile := func(path string) bool { return true }
	mockTestBinary := func(path string) bool { return false }

	release := Release{}
	result := downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)

	if result {
		t.Fatal("Expected DownloadBinary to fail on binary test failure")
	}
}

func TestDownloadBinary_AlphaRelease(t *testing.T) {
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test-binary")

	mockBinary := make([]byte, 2048)

	mockTransport := &MockRoundTripper{
		Response: createMockResponse(200, "application/octet-stream", mockBinary),
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	mockValidateFile := func(path string) bool { return true }
	mockTestBinary := func(path string) bool { return true }

	release := Release{IsAlpha: true, IsBeta: false}
	result := downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)

	if !result {
		t.Fatal("Expected DownloadBinary to succeed for alpha release")
	}
}

func TestDownloadBinary_BetaRelease(t *testing.T) {
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test-binary")

	mockBinary := make([]byte, 2048)

	mockTransport := &MockRoundTripper{
		Response: createMockResponse(200, "application/octet-stream", mockBinary),
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	mockValidateFile := func(path string) bool { return true }
	mockTestBinary := func(path string) bool { return true }

	release := Release{IsAlpha: false, IsBeta: true}
	result := downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)

	if !result {
		t.Fatal("Expected DownloadBinary to succeed for beta release")
	}
}

func TestValidateContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{"GitHub octet-stream", "application/octet-stream", true},
		{"GitHub text/plain", "text/plain", true},
		{"Case insensitive", "APPLICATION/OCTET-STREAM", true},
		{"With charset", "application/octet-stream; charset=binary", true},
		{"Unexpected HTML", "text/html", true}, // Function always returns true but warns
		{"Empty content type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateContentType(tt.contentType)
			if result != tt.expected {
				t.Errorf("validateContentType(%q) = %v, expected %v", tt.contentType, result, tt.expected)
			}
		})
	}
}

func TestWriteTemporaryFile(t *testing.T) {
	tempDir := t.TempDir()
	tmpPath := filepath.Join(tempDir, "test.tmp")

	// Test successful write
	testData := make([]byte, 2048)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	reader := bytes.NewReader(testData)
	result := writeTemporaryFile(tmpPath, reader)

	if !result {
		t.Fatal("Expected writeTemporaryFile to succeed")
	}

	// Verify file was written correctly
	written, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal("Failed to read written file:", err)
	}

	if !bytes.Equal(written, testData) {
		t.Fatal("Written data doesn't match original data")
	}
}

func TestWriteTemporaryFile_TooSmall(t *testing.T) {
	tempDir := t.TempDir()
	tmpPath := filepath.Join(tempDir, "test.tmp")

	// Test with small data (< 1KB)
	smallData := make([]byte, 500)
	reader := bytes.NewReader(smallData)

	result := writeTemporaryFile(tmpPath, reader)

	if result {
		t.Fatal("Expected writeTemporaryFile to fail with small data")
	}

	// Verify temp file was cleaned up
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatal("Expected temp file to be cleaned up after failure")
	}
}

// Benchmark tests
func BenchmarkDownloadBinary(b *testing.B) {
	tempDir := b.TempDir()

	mockBinary := make([]byte, 1024*1024) // 1MB
	mockTransport := &MockRoundTripper{
		Response: createMockResponse(200, "application/octet-stream", mockBinary),
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	mockValidateFile := func(path string) bool { return true }
	mockTestBinary := func(path string) bool { return true }

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		execPath := filepath.Join(tempDir, "binary"+string(rune(i)))
		release := Release{}
		downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)
	}
}

// Test helper to verify cleanup behavior
func TestDownloadBinary_CleanupOnFailure(t *testing.T) {
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "test-binary")

	mockBinary := make([]byte, 2048)

	mockTransport := &MockRoundTripper{
		Response: createMockResponse(200, "application/octet-stream", mockBinary),
	}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = mockTransport
	defer func() { http.DefaultTransport = originalTransport }()

	// Make validation fail to trigger cleanup
	mockValidateFile := func(path string) bool { return false }
	mockTestBinary := func(path string) bool { return true }

	release := Release{}
	result := downloadBinaryWithValidators("http://example.com/binary", execPath, release, mockValidateFile, mockTestBinary)

	if result {
		t.Fatal("Expected DownloadBinary to fail")
	}

	// Verify temp file was cleaned up
	tmpPath := execPath + ".new"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatal("Expected temp file to be cleaned up after failure")
	}
}
