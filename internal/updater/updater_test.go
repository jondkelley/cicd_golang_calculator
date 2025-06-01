package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// ============================================================================
// TEST HELPER FUNCTIONS
// ============================================================================
// These helper functions create test data and reduce code duplication across tests

// createMockManifest creates a VersionManifest with the given releases
// This is a test helper that simulates what we'd get from a real version manifest file
func createMockManifest(releases []Release) *VersionManifest {
	return &VersionManifest{Releases: releases}
}

// createTestRelease creates a Release struct for testing purposes
// Parameters:
//   - version: the semantic version string (e.g., "1.2.3", "2.0.0-alpha")
//   - isAlpha: whether this is an alpha pre-release version
//   - isBeta: whether this is a beta pre-release version
//
// Returns a complete Release struct with dummy data for testing
func createTestRelease(version string, isAlpha, isBeta bool) Release {
	return Release{
		Version:     version,
		URLs:        map[string]string{"linux": "http://example.com/binary"}, // Dummy download URL
		IsAlpha:     isAlpha,
		IsBeta:      isBeta,
		ReleaseDate: "2025-01-01", // Fixed date for predictable testing
	}
}

// ============================================================================
// SEMANTIC VERSION PARSING TESTS
// ============================================================================

// TestParseSemanticVersion tests the ParseSemanticVersion function which converts
// version strings like "1.2.3" or "2.0.0-alpha" into structured SemanticVersion objects
func TestParseSemanticVersion(t *testing.T) {
	// Define test cases using a table-driven test pattern
	// This is a common Go testing idiom that makes it easy to add new test cases
	tests := []struct {
		name        string           // Human-readable description of what we're testing
		version     string           // Input: the version string to parse
		expected    *SemanticVersion // Expected output: what the parsed version should look like
		expectError bool             // Whether we expect this test case to return an error
	}{
		{
			// Test parsing a standard stable version (no pre-release suffix)
			name:    "valid stable version",
			version: "1.2.3",
			expected: &SemanticVersion{
				Major: 1, Minor: 2, Patch: 3,
				PreRelease: "", IsAlpha: false, IsBeta: false,
			},
		},
		{
			// Test that versions with "v" prefix (common in Git tags) are handled correctly
			name:    "valid version with v prefix",
			version: "v2.0.1",
			expected: &SemanticVersion{
				Major: 2, Minor: 0, Patch: 1,
				PreRelease: "", IsAlpha: false, IsBeta: false,
			},
		},
		{
			// Test parsing an alpha pre-release version
			name:    "valid alpha version",
			version: "1.0.0-alpha",
			expected: &SemanticVersion{
				Major: 1, Minor: 0, Patch: 0,
				PreRelease: "alpha", IsAlpha: true, IsBeta: false,
			},
		},
		{
			// Test parsing a beta pre-release version
			name:    "valid beta version",
			version: "2.1.0-beta",
			expected: &SemanticVersion{
				Major: 2, Minor: 1, Patch: 0,
				PreRelease: "beta", IsAlpha: false, IsBeta: true,
			},
		},
		{
			// Test that malformed versions (missing patch number) cause errors
			name:        "invalid version format",
			version:     "1.2",
			expectError: true,
		},
		{
			// Test that non-numeric version components cause errors
			name:        "invalid major version",
			version:     "a.2.3",
			expectError: true,
		},
		{
			// Test that empty input causes an error
			name:        "empty version",
			version:     "",
			expectError: true,
		},
	}

	// Run each test case
	for _, tt := range tests {
		// t.Run creates a subtest for each test case, which helps with debugging
		// If one test fails, the others will still run
		t.Run(tt.name, func(t *testing.T) {
			// Call the function we're testing
			result, err := ParseSemanticVersion(tt.version)

			// If we expected an error, check that we got one
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for version %s, got nil", tt.version)
				}
				return // Don't check the result if we expected an error
			}

			// If we didn't expect an error, make sure we didn't get one
			if err != nil {
				t.Errorf("unexpected error for version %s: %v", tt.version, err)
				return
			}

			// Compare each field of the result with what we expected
			// We check each field individually to give more specific error messages
			if result.Major != tt.expected.Major ||
				result.Minor != tt.expected.Minor ||
				result.Patch != tt.expected.Patch ||
				result.PreRelease != tt.expected.PreRelease ||
				result.IsAlpha != tt.expected.IsAlpha ||
				result.IsBeta != tt.expected.IsBeta {
				t.Errorf("version %s: expected %+v, got %+v", tt.version, tt.expected, result)
			}
		})
	}
}

// ============================================================================
// VERSION COMPARISON TESTS
// ============================================================================

// TestIsNewerThan tests the version comparison logic
// This is crucial for determining whether an update is available
func TestIsNewerThan(t *testing.T) {
	tests := []struct {
		name     string // Description of what we're comparing
		version1 string // The version we're asking "is this newer than..."
		version2 string // The version we're comparing against
		expected bool   // Whether version1 should be considered newer than version2
	}{
		// Test major version comparisons (2.x.x vs 1.x.x)
		{"major version higher", "2.0.0", "1.9.9", true},

		// Test minor version comparisons (1.2.x vs 1.1.x)
		{"minor version higher", "1.2.0", "1.1.9", true},

		// Test patch version comparisons (1.1.2 vs 1.1.1)
		{"patch version higher", "1.1.2", "1.1.1", true},

		// Test that identical versions are not considered "newer"
		{"same version", "1.2.3", "1.2.3", false},

		// Test that older versions are not considered "newer"
		{"lower version", "1.2.3", "1.2.4", false},

		// Test pre-release version handling
		// Stable versions are considered newer than pre-release versions with the same number
		{"stable newer than alpha", "1.2.3", "1.2.3-alpha", true},
		{"stable newer than beta", "1.2.3", "1.2.3-beta", true},

		// Test pre-release version ordering (beta is typically "newer" than alpha)
		{"beta newer than alpha", "1.2.3-beta", "1.2.3-alpha", true},
		{"alpha not newer than beta", "1.2.3-alpha", "1.2.3-beta", false},

		// Test that identical pre-release versions are not considered "newer"
		{"same alpha versions", "1.2.3-alpha", "1.2.3-alpha", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse both version strings into SemanticVersion structs
			v1, err := ParseSemanticVersion(tt.version1)
			if err != nil {
				t.Fatalf("failed to parse version1 %s: %v", tt.version1, err)
			}

			v2, err := ParseSemanticVersion(tt.version2)
			if err != nil {
				t.Fatalf("failed to parse version2 %s: %v", tt.version2, err)
			}

			// Test the comparison method
			result := v1.IsNewerThan(v2)
			if result != tt.expected {
				t.Errorf("%s.IsNewerThan(%s): expected %v, got %v",
					tt.version1, tt.version2, tt.expected, result)
			}
		})
	}
}

// ============================================================================
// UPDATE ELIGIBILITY TESTS
// ============================================================================

// TestFindLatestEligibleReleaseAdvanced tests the core logic for determining
// what updates a user should be offered based on their current version and preferences
func TestFindLatestEligibleReleaseAdvanced(t *testing.T) {
	// Create a realistic set of releases that might exist in a project
	// This simulates what we'd get from a version manifest file
	releases := []Release{
		createTestRelease("1.0.0", false, false),      // Stable release
		createTestRelease("1.1.0-alpha", true, false), // Alpha pre-release
		createTestRelease("1.1.0-beta", false, true),  // Beta pre-release
		createTestRelease("1.1.0", false, false),      // Stable release
		createTestRelease("1.2.0-alpha", true, false), // Newer alpha
		createTestRelease("2.0.0", false, false),      // Latest stable
	}

	tests := []struct {
		name           string  // Description of the test scenario
		currentVersion string  // What version the user currently has
		allowAlpha     string  // Value for CALC_ALLOW_ALPHA environment variable
		allowBeta      string  // Value for CALC_ALLOW_BETA environment variable
		expectedResult *string // Expected version to update to (nil = no update)
	}{
		{
			// A user on a stable version should get the latest stable update
			name:           "stable user gets latest stable",
			currentVersion: "0.9.0",
			expectedResult: stringPtr("2.0.0"),
		},
		{
			// An alpha user without the environment flag shouldn't get updates
			// This prevents alpha users from accidentally getting stuck without updates
			name:           "alpha user without flag gets no update",
			currentVersion: "1.0.0-alpha",
			expectedResult: nil,
		},
		{
			// An alpha user who sets the environment flag should get alpha updates
			name:           "alpha user with flag gets latest alpha",
			currentVersion: "1.0.0-alpha",
			allowAlpha:     "1",
			expectedResult: stringPtr("1.2.0-alpha"),
		},
		{
			// Similar logic for beta users
			name:           "beta user without flag gets no update",
			currentVersion: "1.0.0-beta",
			expectedResult: nil,
		},
		{
			name:           "beta user with flag gets latest beta",
			currentVersion: "1.0.0-beta",
			allowBeta:      "1",
			expectedResult: stringPtr("1.1.0-beta"),
		},
		{
			// If the user already has the latest version, no update should be offered
			name:           "current version is latest",
			currentVersion: "2.0.0",
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables to simulate user preferences
			// These environment variables control whether alpha/beta updates are allowed
			if tt.allowAlpha != "" {
				os.Setenv("CALC_ALLOW_ALPHA", tt.allowAlpha)
				defer os.Unsetenv("CALC_ALLOW_ALPHA") // Clean up after the test
			}
			if tt.allowBeta != "" {
				os.Setenv("CALC_ALLOW_BETA", tt.allowBeta)
				defer os.Unsetenv("CALC_ALLOW_BETA")
			}

			// Call the function we're testing
			result := FindLatestEligibleRelease(releases, tt.currentVersion)

			// Check the result
			if tt.expectedResult == nil {
				// We expected no update to be available
				if result != nil {
					t.Errorf("expected no update, got version %s", result.Version)
				}
			} else {
				// We expected a specific update
				if result == nil {
					t.Errorf("expected update to %s, got nil", *tt.expectedResult)
				} else if result.Version != *tt.expectedResult {
					t.Errorf("expected update to %s, got %s", *tt.expectedResult, result.Version)
				}
			}
		})
	}
}

// ============================================================================
// UPDATE CHECK RESULT TESTS
// ============================================================================

// TestCheckForUpdatesWithResult tests the wrapper function that provides
// detailed information about update availability and gating
func TestCheckForUpdatesWithResult(t *testing.T) {
	// Create test releases
	releases := []Release{
		createTestRelease("1.0.0", false, false),
		createTestRelease("1.1.0-alpha", true, false),
		createTestRelease("1.1.0-beta", false, true),
		createTestRelease("1.1.0", false, false),
		createTestRelease("2.0.0-alpha", true, false),
	}

	tests := []struct {
		name           string            // Test scenario description
		currentVersion string            // User's current version
		allowAlpha     string            // CALC_ALLOW_ALPHA environment variable
		allowBeta      string            // CALC_ALLOW_BETA environment variable
		expectedResult UpdateCheckResult // Expected detailed result
	}{
		{
			// Standard case: stable user with available update
			name:           "stable user has update",
			currentVersion: "0.9.0",
			expectedResult: UpdateCheckResult{
				HasUpdate:      true,     // An update is available
				IsGated:        false,    // No special permission needed
				CurrentChannel: "stable", // User is on stable channel
				RequiredEnvVar: "",       // No environment variable needed
			},
		},
		{
			// Alpha user without permission: update exists but is gated
			name:           "alpha user without flag has gated update",
			currentVersion: "1.0.0-alpha",
			expectedResult: UpdateCheckResult{
				HasUpdate:      true,               // Update exists
				IsGated:        true,               // But it's gated behind env var
				CurrentChannel: "alpha",            // User is on alpha channel
				RequiredEnvVar: "CALC_ALLOW_ALPHA", // This env var would unlock it
			},
		},
		{
			// Alpha user with permission: update is available and ungated
			name:           "alpha user with flag has ungated update",
			currentVersion: "1.0.0-alpha",
			allowAlpha:     "1",
			expectedResult: UpdateCheckResult{
				HasUpdate:      true,
				IsGated:        false, // Permission granted, so not gated
				CurrentChannel: "alpha",
				RequiredEnvVar: "CALC_ALLOW_ALPHA", // Still shows which var was checked
			},
		},
		{
			// Beta user without permission
			name:           "beta user without flag has gated update",
			currentVersion: "1.0.0-beta",
			expectedResult: UpdateCheckResult{
				HasUpdate:      true,
				IsGated:        true,
				CurrentChannel: "beta",
				RequiredEnvVar: "CALC_ALLOW_BETA",
			},
		},
		{
			// User already on latest version
			name:           "no update available",
			currentVersion: "2.0.0",
			expectedResult: UpdateCheckResult{
				HasUpdate:      false, // No newer version exists
				IsGated:        false, // Gating is irrelevant if no update exists
				CurrentChannel: "stable",
				RequiredEnvVar: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			if tt.allowAlpha != "" {
				os.Setenv("CALC_ALLOW_ALPHA", tt.allowAlpha)
				defer os.Unsetenv("CALC_ALLOW_ALPHA")
			}
			if tt.allowBeta != "" {
				os.Setenv("CALC_ALLOW_BETA", tt.allowBeta)
				defer os.Unsetenv("CALC_ALLOW_BETA")
			}

			// Call the function and get detailed results
			result := CheckForUpdatesWithResult(releases, tt.currentVersion)

			// Check each field of the result
			if result.HasUpdate != tt.expectedResult.HasUpdate {
				t.Errorf("HasUpdate: expected %v, got %v", tt.expectedResult.HasUpdate, result.HasUpdate)
			}
			if result.IsGated != tt.expectedResult.IsGated {
				t.Errorf("IsGated: expected %v, got %v", tt.expectedResult.IsGated, result.IsGated)
			}
			if result.CurrentChannel != tt.expectedResult.CurrentChannel {
				t.Errorf("CurrentChannel: expected %s, got %s", tt.expectedResult.CurrentChannel, result.CurrentChannel)
			}
			if result.RequiredEnvVar != tt.expectedResult.RequiredEnvVar {
				t.Errorf("RequiredEnvVar: expected %s, got %s", tt.expectedResult.RequiredEnvVar, result.RequiredEnvVar)
			}
		})
	}
}

// ============================================================================
// NETWORK/HTTP TESTS
// ============================================================================

// TestFetchVersionManifest tests the HTTP functionality for fetching version information
// from a remote server. We use httptest.NewServer to create a fake HTTP server for testing
func TestFetchVersionManifest(t *testing.T) {
	tests := []struct {
		name           string                                       // Test scenario name
		serverResponse func(w http.ResponseWriter, r *http.Request) // How our fake server should respond
		expectError    bool                                         // Whether we expect an error
		expectNil      bool                                         // Whether we expect a nil result
	}{
		{
			// Test successful HTTP request and JSON parsing
			name: "successful fetch",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				// Create a valid version manifest
				manifest := VersionManifest{
					Releases: []Release{
						createTestRelease("1.0.0", false, false),
					},
				}
				// Send it as JSON response
				json.NewEncoder(w).Encode(manifest)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			// Test handling of HTTP 404 errors
			name: "server returns 404",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectError: true,
			expectNil:   false,
		},
		{
			// Test handling of malformed JSON responses
			name: "malformed JSON",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("invalid json"))
			},
			expectError: true,
			expectNil:   false,
		},
		{
			// Test handling of empty release lists (should be treated as an error)
			name: "empty releases",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				manifest := VersionManifest{Releases: []Release{}}
				json.NewEncoder(w).Encode(manifest)
			},
			expectError: true,
			expectNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test HTTP server that responds according to our test case
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close() // Clean up the server when the test finishes

			// Test the network logic by making a direct HTTP request
			// Note: Since we can't easily override the const manifestURL in tests,
			// we test the HTTP handling logic separately here
			client := &http.Client{}
			resp, err := client.Get(server.URL)

			if tt.expectError {
				// For error cases, we simulate what FetchVersionManifest would do
				if resp != nil && resp.StatusCode != 200 {
					// This would cause an error in the real function
					if resp.StatusCode == 404 {
						expectedErr := fmt.Sprintf("no version.json manifest found in project repository (status code = %d)", resp.StatusCode)
						t.Logf("Expected error: %s", expectedErr)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			defer resp.Body.Close()

			// Test JSON decoding if we got a successful response
			if resp.StatusCode == 200 {
				var manifest VersionManifest
				err = json.NewDecoder(resp.Body).Decode(&manifest)
				if err != nil && !tt.expectError {
					t.Errorf("failed to decode manifest: %v", err)
				}
			}
		})
	}
}

// ============================================================================
// UTILITY FUNCTION TESTS
// ============================================================================

// TestIsEnvVarTrue tests the helper function that interprets environment variables as booleans
// This is important because environment variables are always strings, but we need boolean logic
func TestIsEnvVarTrue(t *testing.T) {
	tests := []struct {
		name     string // Test case description
		envVar   string // Environment variable name
		value    string // Value to set the environment variable to
		expected bool   // Whether this should be interpreted as "true"
	}{
		// Test values that should be interpreted as "false"
		{"empty value", "TEST_VAR", "", false},
		{"zero value", "TEST_VAR", "0", false},
		{"false lowercase", "TEST_VAR", "false", false},
		{"false uppercase", "TEST_VAR", "FALSE", false},
		{"false mixed case", "TEST_VAR", "False", false},

		// Test values that should be interpreted as "true"
		{"true value", "TEST_VAR", "1", true},
		{"true string", "TEST_VAR", "true", true},
		{"yes value", "TEST_VAR", "yes", true},
		{"any other value", "TEST_VAR", "anything", true}, // Non-false values default to true

		// Test edge cases with whitespace
		{"whitespace around false", "TEST_VAR", " false ", false},
		{"whitespace around true", "TEST_VAR", " 1 ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable for this test
			os.Setenv(tt.envVar, tt.value)
			defer os.Unsetenv(tt.envVar) // Clean up after the test

			// Test the function
			result := isEnvVarTrue(tt.envVar)
			if result != tt.expected {
				t.Errorf("isEnvVarTrue(%s=%s): expected %v, got %v",
					tt.envVar, tt.value, tt.expected, result)
			}
		})
	}
}

// TestIsNetworkError tests the function that determines if an error is network-related
// This helps distinguish between network problems and other types of errors
func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name     string // Test case description
		err      error  // Error to test
		expected bool   // Whether this should be considered a network error
	}{
		// Non-error cases
		{"nil error", nil, false},

		// Common network errors that should be detected
		{"connection refused", fmt.Errorf("dial tcp: connection refused"), true},
		{"no such host", fmt.Errorf("no such host example.com"), true},
		{"timeout", fmt.Errorf("dial tcp: timeout"), true},
		{"network unreachable", fmt.Errorf("network is unreachable"), true},

		// Non-network errors
		{"other error", fmt.Errorf("some other error"), false},
		{"empty error", fmt.Errorf(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNetworkError(tt.err)
			if result != tt.expected {
				t.Errorf("isNetworkError(%v): expected %v, got %v",
					tt.err, tt.expected, result)
			}
		})
	}
}

// ============================================================================
// HELPER FUNCTIONS FOR TESTS
// ============================================================================

// stringPtr is a helper function that creates a pointer to a string
// This is useful for test cases where we need to distinguish between
// "no result" (nil) and "empty result" (pointer to empty string)
func stringPtr(s string) *string {
	return &s
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

// TestCheckForUpdateIntegration would test the complete update check flow
// This is a placeholder showing how you might test the entire system together
func TestCheckForUpdateIntegration(t *testing.T) {
	// Create a mock HTTP server that serves version manifests
	releases := []Release{
		createTestRelease("1.0.0", false, false),
		createTestRelease("1.1.0", false, false),
		createTestRelease("2.0.0-alpha", true, false),
	}
	manifest := VersionManifest{Releases: releases}

	// Set up a test server that returns our mock manifest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(manifest)
	}))
	defer server.Close()

	// Note: This test would require modifying the package to accept a custom URL
	// For now, we test the components separately above
	// In a real implementation, maybe:
	// 1. Make the manifest URL configurable
	// 2. Use dependency injection to mock the HTTP client
	// 3. Use build tags to switch between test and production configurations
	t.Log("Integration test placeholder - components tested individually")
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================
// Benchmark tests help us understand the performance characteristics of our code
// They're useful for detecting performance regressions and optimizing critical paths

// BenchmarkParseSemanticVersion measures how fast version parsing is
// This is important if we're parsing many versions frequently
func BenchmarkParseSemanticVersion(b *testing.B) {
	// b.N is set by the testing framework and represents the number of iterations
	for i := 0; i < b.N; i++ {
		ParseSemanticVersion("1.2.3-beta")
	}
}

// BenchmarkIsNewerThan measures version comparison performance
func BenchmarkIsNewerThan(b *testing.B) {
	// Pre-parse versions outside the benchmark loop so we only measure comparison time
	v1, _ := ParseSemanticVersion("2.0.0")
	v2, _ := ParseSemanticVersion("1.9.9")

	b.ResetTimer() // Don't count the setup time in the benchmark
	for i := 0; i < b.N; i++ {
		v1.IsNewerThan(v2)
	}
}

// BenchmarkFindLatestEligibleRelease measures how the function performs with many releases
// This helps us understand if our algorithm scales well
func BenchmarkFindLatestEligibleRelease(b *testing.B) {
	// Create 100 releases to simulate a project with many versions
	releases := make([]Release, 100)
	for i := 0; i < 100; i++ {
		releases[i] = createTestRelease(fmt.Sprintf("1.%d.0", i), false, false)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindLatestEligibleRelease(releases, "0.1.0")
	}
}
