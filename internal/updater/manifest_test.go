// Add this to internal/updater/manifest_test.go
package updater

import (
	"fmt"
	"os"
	"testing"
)

func TestFindLatestEligibleRelease(t *testing.T) {
	// Sample releases that mirror your real manifest structure
	testReleases := []Release{
		{Version: "v1.1.5", IsAlpha: false, IsBeta: false},
		{Version: "v1.1.4", IsAlpha: false, IsBeta: false},
		{Version: "v0.1.3", IsAlpha: false, IsBeta: false},
		{Version: "v0.1.2", IsAlpha: false, IsBeta: false},
		{Version: "v0.0.12-beta", IsAlpha: false, IsBeta: true},
		{Version: "v0.0.11-beta", IsAlpha: false, IsBeta: true},
		{Version: "v0.0.10-beta", IsAlpha: false, IsBeta: true},
		{Version: "v0.0.12-alpha", IsAlpha: true, IsBeta: false},
		{Version: "v0.0.11-alpha", IsAlpha: true, IsBeta: false},
		{Version: "v0.0.10-alpha", IsAlpha: true, IsBeta: false},
		{Version: "v0.0.8-alpha", IsAlpha: true, IsBeta: false},
		{Version: "v0.0.1-alpha", IsAlpha: true, IsBeta: false},
	}

	tests := []struct {
		name           string
		currentVersion string
		envVars        map[string]string
		expected       string
		description    string
	}{
		{
			name:           "BugReproduction_AlphaUser_ShouldGetNewestAlpha",
			currentVersion: "v0.0.10-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "v0.0.12-alpha",
			description:    "THIS IS THE BUG!!!!!: Alpha user on v0.0.10-alpha should get v0.0.12-alpha, not v0.0.1-alpha",
		},
		{
			name:           "AlphaUser_NewerAlphaAvailable",
			currentVersion: "v0.0.8-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "v0.0.12-alpha",
			description:    "Alpha user should always get the newest alpha when available",
		},
		{
			name:           "AlphaUser_NoAlphaEnvVar_NoUpdate",
			currentVersion: "v0.0.10-alpha",
			envVars:        map[string]string{},
			expected:       "",
			description:    "Alpha user without CALC_ALLOW_ALPHA should get NO update (channel isolation)",
		},
		{
			name:           "StableUser_ShouldGetNewerStable",
			currentVersion: "v0.1.2",
			envVars:        map[string]string{},
			expected:       "v1.1.5",
			description:    "Stable user should get newer stable version",
		},
		{
			name:           "StableUser_WithAlphaEnabled_ShouldStillGetStable",
			currentVersion: "v0.1.2",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "v1.1.5",
			description:    "Stable user should only get stable versions (channel isolation)",
		},
		{
			name:           "BetaUser_ShouldGetNewerBeta",
			currentVersion: "v0.0.10-beta",
			envVars:        map[string]string{"CALC_ALLOW_BETA": "1"},
			expected:       "v0.0.12-beta",
			description:    "Beta user should get newer beta version only",
		},
		{
			name:           "BetaUser_NoNewerBeta_NoUpdate",
			currentVersion: "v0.0.12-beta",
			envVars:        map[string]string{"CALC_ALLOW_BETA": "1"},
			expected:       "",
			description:    "Beta user with no newer beta should get NO update (channel isolation)",
		},
		{
			name:           "BetaUser_NoBetaEnvVar_NoUpdate",
			currentVersion: "v0.0.10-beta",
			envVars:        map[string]string{},
			expected:       "",
			description:    "Beta user without CALC_ALLOW_BETA should get NO update (channel isolation)",
		},
		{
			name:           "LatestStable_NoUpdate",
			currentVersion: "v1.1.5",
			envVars:        map[string]string{},
			expected:       "",
			description:    "User on latest stable should get no update",
		},
		{
			name:           "LatestAlpha_NoUpdate",
			currentVersion: "v0.0.12-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "",
			description:    "User on latest alpha should get NO update (stays in alpha channel)",
		},
		{
			name:           "StableUser_WithBetaEnabled_ShouldOnlyGetStable",
			currentVersion: "v0.1.2",
			envVars:        map[string]string{"CALC_ALLOW_BETA": "1"},
			expected:       "v1.1.5",
			description:    "Stable user should only get stable versions even with beta enabled",
		},
		{
			name:           "EdgeCase_VeryOldVersion",
			currentVersion: "v0.0.1-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "v0.0.12-alpha",
			description:    "Very old alpha should get newest alpha",
		},
		{
			name:           "EdgeCase_UnparsableVersion",
			currentVersion: "invalid-version",
			envVars:        map[string]string{},
			expected:       "",
			description:    "Invalid current version should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			// Run the function
			result := FindLatestEligibleRelease(testReleases, tt.currentVersion)

			// Check result
			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected no update (nil), but got: %s", result.Version)
				}
			} else {
				if result == nil {
					t.Errorf("Expected update to %s, but got nil", tt.expected)
				} else if result.Version != tt.expected {
					t.Errorf("Expected update to %s, but got %s", tt.expected, result.Version)
				}
			}

			// Extra logging for the main bug case
			if tt.name == "BugReproduction_AlphaUser_ShouldGetNewestAlpha" {
				if result != nil {
					t.Logf("BUG TEST: Current=%s, Got=%s, Expected=%s",
						tt.currentVersion, result.Version, tt.expected)
					if result.Version == "v0.0.1-alpha" {
						t.Logf("BUG CONFIRMED: Got older alpha version instead of newer one!")
					} else if result.Version == tt.expected {
						t.Logf("BUG FIXED: Got correct newer alpha version!")
					}
				}
			}
		})
	}
}

func TestSemanticVersionComparison(t *testing.T) {
	tests := []struct {
		version1        string
		version2        string
		v1ShouldBeNewer bool
		description     string
	}{
		{"v0.0.12-alpha", "v0.0.10-alpha", true, "v0.0.12-alpha should be newer than v0.0.10-alpha"},
		{"v0.0.12-alpha", "v0.0.1-alpha", true, "v0.0.12-alpha should be newer than v0.0.1-alpha"},
		{"v0.0.10-alpha", "v0.0.1-alpha", true, "v0.0.10-alpha should be newer than v0.0.1-alpha"},
		{"v1.1.5", "v0.0.12-alpha", true, "v1.1.5 stable should be newer than v0.0.12-alpha"},
		{"v0.0.12-beta", "v0.0.12-alpha", true, "v0.0.12-beta should be newer than v0.0.12-alpha (beta > alpha)"},
		{"v1.0.0", "v0.0.12-alpha", true, "v1.0.0 stable should be newer than v0.0.12-alpha"},
		{"v0.0.1-alpha", "v0.0.10-alpha", false, "v0.0.1-alpha should NOT be newer than v0.0.10-alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			v1, err1 := ParseSemanticVersion(tt.version1)
			v2, err2 := ParseSemanticVersion(tt.version2)

			if err1 != nil || err2 != nil {
				t.Fatalf("Failed to parse versions: %v, %v", err1, err2)
			}

			result := v1.IsNewerThan(v2)
			if result != tt.v1ShouldBeNewer {
				t.Errorf("Expected %s.IsNewerThan(%s) = %t, got %t",
					tt.version1, tt.version2, tt.v1ShouldBeNewer, result)
			}
		})
	}
}

func TestEdgeCases_VersionParsing(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		releases       []Release
		envVars        map[string]string
		expected       string
		shouldError    bool
		description    string
	}{
		{
			name:           "InvalidCurrentVersion_ShouldReturnNil",
			currentVersion: "not-a-version",
			releases:       []Release{{Version: "v1.0.0", IsAlpha: false, IsBeta: false}},
			envVars:        map[string]string{},
			expected:       "",
			shouldError:    true,
			description:    "Invalid current version should return nil",
		},
		{
			name:           "CurrentVersionMissingV_ShouldStillWork",
			currentVersion: "1.0.0",
			releases:       []Release{{Version: "v1.1.0", IsAlpha: false, IsBeta: false}},
			envVars:        map[string]string{},
			expected:       "v1.1.0",
			shouldError:    false,
			description:    "Current version without 'v' prefix should still work",
		},
		{
			name:           "ReleaseVersionMissingV_ShouldBeSkipped",
			currentVersion: "v1.0.0",
			releases: []Release{
				{Version: "1.1.0", IsAlpha: false, IsBeta: false},  // Invalid - no 'v'
				{Version: "v1.2.0", IsAlpha: false, IsBeta: false}, // Valid
			},
			envVars:     map[string]string{},
			expected:    "v1.2.0",
			shouldError: false,
			description: "Releases with invalid versions should be skipped",
		},
		{
			name:           "MalformedSemanticVersion_ShouldBeSkipped",
			currentVersion: "v1.0.0",
			releases: []Release{
				{Version: "v1.1", IsAlpha: false, IsBeta: false},     // Invalid - only major.minor
				{Version: "v1.1.0.1", IsAlpha: false, IsBeta: false}, // Invalid - too many parts
				{Version: "v1.2.0", IsAlpha: false, IsBeta: false},   // Valid
			},
			envVars:     map[string]string{},
			expected:    "v1.2.0",
			shouldError: false,
			description: "Malformed semantic versions should be skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			result := FindLatestEligibleRelease(tt.releases, tt.currentVersion)

			if tt.shouldError {
				if result != nil {
					t.Errorf("Expected nil due to error, but got: %s", result.Version)
				}
			} else if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected no update, but got: %s", result.Version)
				}
			} else {
				if result == nil {
					t.Errorf("Expected %s, but got nil", tt.expected)
				} else if result.Version != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result.Version)
				}
			}
		})
	}
}

func TestEdgeCases_VersionBoundaries(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		releases       []Release
		envVars        map[string]string
		expected       string
		description    string
	}{
		{
			name:           "ExactSameVersion_ShouldNotUpdate",
			currentVersion: "v1.0.0-alpha",
			releases:       []Release{{Version: "v1.0.0-alpha", IsAlpha: true, IsBeta: false}},
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "",
			description:    "Exact same version should not trigger update",
		},
		{
			name:           "OlderVersionAvailable_ShouldNotUpdate",
			currentVersion: "v1.0.0-alpha",
			releases: []Release{
				{Version: "v0.9.0-alpha", IsAlpha: true, IsBeta: false},
				{Version: "v0.5.0-alpha", IsAlpha: true, IsBeta: false},
			},
			envVars:     map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:    "",
			description: "Only older versions available should not trigger update",
		},
		{
			name:           "MajorVersionZero_ShouldWork",
			currentVersion: "v0.0.1",
			releases:       []Release{{Version: "v0.0.2", IsAlpha: false, IsBeta: false}},
			envVars:        map[string]string{},
			expected:       "v0.0.2",
			description:    "Major version 0 should work correctly",
		},
		{
			name:           "LargeVersionNumbers_ShouldWork",
			currentVersion: "v999.999.999",
			releases:       []Release{{Version: "v1000.0.0", IsAlpha: false, IsBeta: false}},
			envVars:        map[string]string{},
			expected:       "v1000.0.0",
			description:    "Large version numbers should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			result := FindLatestEligibleRelease(tt.releases, tt.currentVersion)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected no update, but got: %s", result.Version)
				}
			} else {
				if result == nil {
					t.Errorf("Expected %s, but got nil", tt.expected)
				} else if result.Version != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result.Version)
				}
			}
		})
	}
}

func TestEdgeCases_ChannelFlags(t *testing.T) {
	releases := []Release{
		{Version: "v1.0.0", IsAlpha: false, IsBeta: false},
		{Version: "v0.9.0-beta", IsAlpha: false, IsBeta: true},
		{Version: "v0.8.0-alpha", IsAlpha: true, IsBeta: false},
	}

	tests := []struct {
		name           string
		currentVersion string
		envVars        map[string]string
		expected       string
		description    string
	}{
		{
			name:           "AlphaFlagEmpty_ShouldBeFalse",
			currentVersion: "v0.7.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": ""},
			expected:       "",
			description:    "Empty CALC_ALLOW_ALPHA should be treated as false",
		},
		{
			name:           "BetaFlagEmpty_ShouldBeFalse",
			currentVersion: "v0.7.0-beta",
			envVars:        map[string]string{"CALC_ALLOW_BETA": ""},
			expected:       "",
			description:    "Empty CALC_ALLOW_BETA should be treated as false",
		},
		{
			name:           "AlphaFlagZero_ShouldBeFalse",
			currentVersion: "v0.7.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "0"},
			expected:       "",
			description:    "CALC_ALLOW_ALPHA=0 should be treated as false",
		},
		{
			name:           "AlphaFlagFalse_ShouldBeFalse",
			currentVersion: "v0.7.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "false"},
			expected:       "",
			description:    "CALC_ALLOW_ALPHA=false should be treated as false",
		},
		{
			name:           "AlphaFlagAnyNonEmptyValue_ShouldBeTrue",
			currentVersion: "v0.7.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "yes"},
			expected:       "v0.8.0-alpha",
			description:    "Any non-empty CALC_ALLOW_ALPHA should be treated as true",
		},
		{
			name:           "BothFlagsSet_AlphaUserShouldStayAlpha",
			currentVersion: "v0.7.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			expected:       "v0.8.0-alpha",
			description:    "Alpha user should stay in alpha channel even with beta flag set",
		},
		{
			name:           "BothFlagsSet_BetaUserShouldStayBeta",
			currentVersion: "v0.7.0-beta",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			expected:       "v0.9.0-beta",
			description:    "Beta user should stay in beta channel even with alpha flag set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			result := FindLatestEligibleRelease(releases, tt.currentVersion)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected no update, but got: %s", result.Version)
				}
			} else {
				if result == nil {
					t.Errorf("Expected %s, but got nil", tt.expected)
				} else if result.Version != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result.Version)
				}
			}
		})
	}
}

func TestEdgeCases_ReleaseListOrdering(t *testing.T) {
	// Test that the function works regardless of release list ordering
	tests := []struct {
		name           string
		currentVersion string
		releases       []Release
		envVars        map[string]string
		expected       string
		description    string
	}{
		{
			name:           "ReleasesInReverseOrder_ShouldFindLatest",
			currentVersion: "v1.0.0",
			releases: []Release{
				{Version: "v1.1.0", IsAlpha: false, IsBeta: false},
				{Version: "v1.3.0", IsAlpha: false, IsBeta: false}, // Latest (should be picked)
				{Version: "v1.2.0", IsAlpha: false, IsBeta: false},
			},
			envVars:     map[string]string{},
			expected:    "v1.3.0",
			description: "Should find latest regardless of list order",
		},
		{
			name:           "ReleasesRandomOrder_ShouldFindLatest",
			currentVersion: "v0.1.0-alpha",
			releases: []Release{
				{Version: "v0.3.0-alpha", IsAlpha: true, IsBeta: false},
				{Version: "v0.1.5-alpha", IsAlpha: true, IsBeta: false},
				{Version: "v0.5.0-alpha", IsAlpha: true, IsBeta: false}, // Latest
				{Version: "v0.2.0-alpha", IsAlpha: true, IsBeta: false},
			},
			envVars:     map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:    "v0.5.0-alpha",
			description: "Should find latest alpha regardless of list order",
		},
		{
			name:           "DuplicateVersions_ShouldHandleGracefully",
			currentVersion: "v1.0.0",
			releases: []Release{
				{Version: "v1.1.0", IsAlpha: false, IsBeta: false},
				{Version: "v1.1.0", IsAlpha: false, IsBeta: false}, // Duplicate
				{Version: "v1.2.0", IsAlpha: false, IsBeta: false},
			},
			envVars:     map[string]string{},
			expected:    "v1.2.0",
			description: "Should handle duplicate versions gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			result := FindLatestEligibleRelease(tt.releases, tt.currentVersion)

			if result == nil {
				t.Errorf("Expected %s, but got nil", tt.expected)
			} else if result.Version != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result.Version)
			}
		})
	}
}

func TestEdgeCases_PreReleaseComparison(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		releases       []Release
		envVars        map[string]string
		expected       string
		description    string
	}{
		{
			name:           "AlphaToBeta_ShouldNotUpdate",
			currentVersion: "v1.0.0-alpha",
			releases:       []Release{{Version: "v1.0.0-beta", IsAlpha: false, IsBeta: true}},
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			expected:       "",
			description:    "Alpha user should not get beta update (channel isolation)",
		},
		{
			name:           "BetaToAlpha_ShouldNotUpdate",
			currentVersion: "v1.0.0-beta",
			releases:       []Release{{Version: "v1.0.1-alpha", IsAlpha: true, IsBeta: false}},
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			expected:       "",
			description:    "Beta user should not get alpha update (channel isolation)",
		},
		{
			name:           "PreReleaseToStable_ShouldNotUpdate",
			currentVersion: "v1.0.0-alpha",
			releases:       []Release{{Version: "v1.0.0", IsAlpha: false, IsBeta: false}},
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "",
			description:    "Alpha user should not get stable update (channel isolation)",
		},
		{
			name:           "StableToPreRelease_ShouldNotUpdate",
			currentVersion: "v1.0.0",
			releases: []Release{
				{Version: "v1.0.1-alpha", IsAlpha: true, IsBeta: false},
				{Version: "v1.0.1-beta", IsAlpha: false, IsBeta: true},
			},
			envVars:     map[string]string{},
			expected:    "",
			description: "Stable user should not get pre-release updates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			result := FindLatestEligibleRelease(tt.releases, tt.currentVersion)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected no update, but got: %s", result.Version)
					t.Errorf("This violates channel isolation rules!")
				}
			} else {
				if result == nil {
					t.Errorf("Expected %s, but got nil", tt.expected)
				} else if result.Version != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result.Version)
				}
			}
		})
	}
}

// Helper function to clear environment variables
func clearEnvVars() {
	os.Unsetenv("CALC_ALLOW_ALPHA")
	os.Unsetenv("CALC_ALLOW_BETA")
}

// Debug version of FindLatestEligibleRelease with extensive logging
func FindLatestEligibleReleaseDebug(releases []Release, currentVersion string) *Release {
	fmt.Printf("DEBUG: Starting FindLatestEligibleRelease with currentVersion=%s\n", currentVersion)

	allowAlpha := isEnvVarTrue("CALC_ALLOW_ALPHA")
	allowBeta := isEnvVarTrue("CALC_ALLOW_BETA")
	fmt.Printf("DEBUG: allowAlpha=%t, allowBeta=%t\n", allowAlpha, allowBeta)

	currentSemVer, err := ParseSemanticVersion(currentVersion)
	if err != nil {
		fmt.Printf("Warning: Could not parse current version %s: %v\n", currentVersion, err)
		return nil
	}

	// Determine current channel
	currentChannel := "stable"
	if currentSemVer.IsAlpha {
		currentChannel = "alpha"
	} else if currentSemVer.IsBeta {
		currentChannel = "beta"
	}
	fmt.Printf("DEBUG: currentChannel=%s\n", currentChannel)

	var latestEligible *Release
	var latestSemVer *SemanticVersion

	for i, release := range releases {
		fmt.Printf("DEBUG: [%d] Processing release %s (IsAlpha=%t, IsBeta=%t)\n",
			i, release.Version, release.IsAlpha, release.IsBeta)

		// Parse release version first
		releaseSemVer, err := ParseSemanticVersion(release.Version)
		if err != nil {
			fmt.Printf("Warning: Could not parse release version %s: %v\n", release.Version, err)
			continue
		}

		// Skip if this release is not newer than current version
		isNewer := releaseSemVer.IsNewerThan(currentSemVer)
		fmt.Printf("DEBUG: [%d] %s.IsNewerThan(%s) = %t\n",
			i, release.Version, currentVersion, isNewer)
		if !isNewer {
			fmt.Printf("DEBUG: [%d] Skipping %s - not newer than current\n", i, release.Version)
			continue
		}

		// Determine release channel
		releaseChannel := "stable"
		if release.IsAlpha {
			releaseChannel = "alpha"
		} else if release.IsBeta {
			releaseChannel = "beta"
		}
		fmt.Printf("DEBUG: [%d] releaseChannel=%s\n", i, releaseChannel)

		// Apply channel isolation rules
		isEligible := false
		switch currentChannel {
		case "alpha":
			// Alpha users can only get alpha updates if CALC_ALLOW_ALPHA is set
			if releaseChannel == "alpha" && allowAlpha {
				isEligible = true
			}
		case "beta":
			// Beta users can only get beta updates if CALC_ALLOW_BETA is set
			if releaseChannel == "beta" && allowBeta {
				isEligible = true
			}
		case "stable":
			// Stable users can only get stable updates (no flags needed)
			if releaseChannel == "stable" {
				isEligible = true
			}
		}

		fmt.Printf("DEBUG: [%d] isEligible=%t (currentChannel=%s, releaseChannel=%s)\n",
			i, isEligible, currentChannel, releaseChannel)

		if !isEligible {
			fmt.Printf("DEBUG: [%d] Skipping %s - not eligible\n", i, release.Version)
			continue
		}

		// If this is the first eligible release, or if it's newer than our current best
		if latestEligible == nil {
			fmt.Printf("DEBUG: [%d] Setting %s as first eligible release\n", i, release.Version)
			latestEligible = &release
			latestSemVer = releaseSemVer
		} else {
			isNewerThanBest := releaseSemVer.IsNewerThan(latestSemVer)
			fmt.Printf("DEBUG: [%d] %s.IsNewerThan(%s) = %t\n",
				i, release.Version, latestEligible.Version, isNewerThanBest)
			if isNewerThanBest {
				fmt.Printf("DEBUG: [%d] Updating best from %s to %s\n",
					i, latestEligible.Version, release.Version)
				latestEligible = &release
				latestSemVer = releaseSemVer
			}
		}
	}

	if latestEligible != nil {
		fmt.Printf("DEBUG: Final result: %s\n", latestEligible.Version)
	} else {
		fmt.Printf("DEBUG: Final result: nil\n")
	}
	return latestEligible
}

func TestStrictChannelIsolation(t *testing.T) {
	// This test ensures users CANNOT escape their channel under ANY circumstances
	releases := []Release{
		{Version: "v2.0.0", IsAlpha: false, IsBeta: false},      // Latest stable
		{Version: "v1.9.0-beta", IsAlpha: false, IsBeta: true},  // Latest beta
		{Version: "v1.8.0-alpha", IsAlpha: true, IsBeta: false}, // Latest alpha
		{Version: "v1.0.0", IsAlpha: false, IsBeta: false},      // Older stable
		{Version: "v0.9.0-beta", IsAlpha: false, IsBeta: true},  // Older beta
		{Version: "v0.8.0-alpha", IsAlpha: true, IsBeta: false}, // Older alpha
	}

	isolationTests := []struct {
		name           string
		currentVersion string
		envVars        map[string]string
		expected       string
		description    string
	}{
		// Alpha channel isolation
		{
			name:           "AlphaUser_CannotGetBeta",
			currentVersion: "v0.8.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			expected:       "v1.8.0-alpha", // Should get newer alpha, NOT beta
			description:    "Alpha user should NEVER get beta updates even with beta flag enabled",
		},
		{
			name:           "AlphaUser_CannotGetStable",
			currentVersion: "v0.8.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "v1.8.0-alpha", // Should get newer alpha, NOT stable
			description:    "Alpha user should NEVER get stable updates",
		},
		{
			name:           "AlphaUser_WithoutFlag_GetsNothing",
			currentVersion: "v0.8.0-alpha",
			envVars:        map[string]string{}, // No alpha flag
			expected:       "",
			description:    "Alpha user without CALC_ALLOW_ALPHA should get NO updates",
		},

		// Beta channel isolation
		{
			name:           "BetaUser_CannotGetAlpha",
			currentVersion: "v0.9.0-beta",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			expected:       "v1.9.0-beta", // Should get newer beta, NOT alpha
			description:    "Beta user should NEVER get alpha updates even with alpha flag enabled",
		},
		{
			name:           "BetaUser_CannotGetStable",
			currentVersion: "v0.9.0-beta",
			envVars:        map[string]string{"CALC_ALLOW_BETA": "1"},
			expected:       "v1.9.0-beta", // Should get newer beta, NOT stable
			description:    "Beta user should NEVER get stable updates",
		},
		{
			name:           "BetaUser_WithoutFlag_GetsNothing",
			currentVersion: "v0.9.0-beta",
			envVars:        map[string]string{}, // No beta flag
			expected:       "",
			description:    "Beta user without CALC_ALLOW_BETA should get NO updates",
		},

		// Stable channel isolation
		{
			name:           "StableUser_CannotGetAlpha",
			currentVersion: "v1.0.0",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "v2.0.0", // Should get newer stable, NOT alpha
			description:    "Stable user should NEVER get alpha updates even with alpha flag enabled",
		},
		{
			name:           "StableUser_CannotGetBeta",
			currentVersion: "v1.0.0",
			envVars:        map[string]string{"CALC_ALLOW_BETA": "1"},
			expected:       "v2.0.0", // Should get newer stable, NOT beta
			description:    "Stable user should NEVER get beta updates even with beta flag enabled",
		},
		{
			name:           "StableUser_CannotGetBeta_EvenWithBothFlags",
			currentVersion: "v1.0.0",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			expected:       "v2.0.0", // Should get newer stable, NOT alpha or beta
			description:    "Stable user should NEVER get pre-release updates regardless of flags",
		},
	}

	for _, tt := range isolationTests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			result := FindLatestEligibleRelease(releases, tt.currentVersion)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("CHANNEL ISOLATION VIOLATION: Expected no update, but got: %s", result.Version)
					t.Errorf("This violates strict channel isolation!")
				}
			} else {
				if result == nil {
					t.Errorf("Expected update to %s, but got nil", tt.expected)
				} else if result.Version != tt.expected {
					t.Errorf("CHANNEL ISOLATION VIOLATION: Expected %s, but got %s", tt.expected, result.Version)
					t.Errorf("This indicates user escaped their intended channel!")
				}
			}
		})
	}
}

func TestForbiddenCrossChannelUpgrades(t *testing.T) {
	// This test explicitly verifies that ALL cross-channel upgrades are BLOCKED
	releases := []Release{
		{Version: "v3.0.0", IsAlpha: false, IsBeta: false},      // Highest stable
		{Version: "v2.9.0-beta", IsAlpha: false, IsBeta: true},  // Highest beta
		{Version: "v2.8.0-alpha", IsAlpha: true, IsBeta: false}, // Highest alpha
	}

	forbiddenUpgrades := []struct {
		name           string
		currentVersion string
		envVars        map[string]string
		description    string
	}{
		// Alpha → Beta (FORBIDDEN)
		{
			name:           "Alpha_To_Beta_Forbidden",
			currentVersion: "v1.0.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			description:    "Alpha user should NEVER upgrade to beta",
		},
		// Alpha → Stable (FORBIDDEN)
		{
			name:           "Alpha_To_Stable_Forbidden",
			currentVersion: "v1.0.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			description:    "Alpha user should NEVER upgrade to stable",
		},
		// Beta → Alpha (FORBIDDEN)
		{
			name:           "Beta_To_Alpha_Forbidden",
			currentVersion: "v1.0.0-beta",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			description:    "Beta user should NEVER downgrade to alpha",
		},
		// Beta → Stable (FORBIDDEN)
		{
			name:           "Beta_To_Stable_Forbidden",
			currentVersion: "v1.0.0-beta",
			envVars:        map[string]string{"CALC_ALLOW_BETA": "1"},
			description:    "Beta user should NEVER upgrade to stable",
		},
		// Stable → Alpha (FORBIDDEN)
		{
			name:           "Stable_To_Alpha_Forbidden",
			currentVersion: "v1.0.0",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			description:    "Stable user should NEVER downgrade to alpha",
		},
		// Stable → Beta (FORBIDDEN)
		{
			name:           "Stable_To_Beta_Forbidden",
			currentVersion: "v1.0.0",
			envVars:        map[string]string{"CALC_ALLOW_BETA": "1"},
			description:    "Stable user should NEVER downgrade to beta",
		},
	}

	for _, tt := range forbiddenUpgrades {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			result := FindLatestEligibleRelease(releases, tt.currentVersion)

			// For forbidden cross-channel upgrades, we expect either:
			// 1. No update (nil result) - indicating proper channel isolation
			// 2. Same-channel update only - never cross-channel

			if result != nil {
				// If we got an update, verify it's NOT a cross-channel upgrade
				currentSemVer, _ := ParseSemanticVersion(tt.currentVersion)
				resultSemVer, _ := ParseSemanticVersion(result.Version)

				currentChannel := getChannelName(currentSemVer)
				resultChannel := getChannelName(resultSemVer)

				if currentChannel != resultChannel {
					t.Errorf("FORBIDDEN CROSS-CHANNEL UPGRADE DETECTED!")
					t.Errorf("Current: %s (%s channel)", tt.currentVersion, currentChannel)
					t.Errorf("Result: %s (%s channel)", result.Version, resultChannel)
					t.Errorf("This violates channel isolation policy!")
				}
			}
		})
	}
}

func TestChannelIsolationWithNoEligibleUpdates(t *testing.T) {
	// Test channel isolation when there are NO eligible updates in the user's channel
	releases := []Release{
		{Version: "v2.0.0", IsAlpha: false, IsBeta: false},     // Only stable available
		{Version: "v1.0.0-beta", IsAlpha: false, IsBeta: true}, // Only beta available
		// No alpha releases available
	}

	tests := []struct {
		name           string
		currentVersion string
		envVars        map[string]string
		expected       string
		description    string
	}{
		{
			name:           "AlphaUser_NoAlphaAvailable_ShouldNotGetBetaOrStable",
			currentVersion: "v0.5.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"},
			expected:       "", // Should get nothing, not cross channels
			description:    "Alpha user should get nothing when no alpha updates available",
		},
		{
			name:           "BetaUser_OnLatestBeta_ShouldNotGetStable",
			currentVersion: "v1.0.0-beta",
			envVars:        map[string]string{"CALC_ALLOW_BETA": "1"},
			expected:       "", // Should get nothing, not cross to stable
			description:    "Beta user on latest beta should not get stable updates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			result := FindLatestEligibleRelease(releases, tt.currentVersion)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("CHANNEL ISOLATION VIOLATION: Expected no update (respecting channel isolation), but got: %s", result.Version)
					t.Errorf("Users should stay in their channel even when no updates are available in that channel")
				}
			}
		})
	}
}

func TestEnvironmentFlagStrictness(t *testing.T) {
	// Test that environment flags are interpreted strictly
	releases := []Release{
		{Version: "v1.1.0-alpha", IsAlpha: true, IsBeta: false},
		{Version: "v1.1.0-beta", IsAlpha: false, IsBeta: true},
	}

	flagTests := []struct {
		name           string
		currentVersion string
		envVars        map[string]string
		expected       string
		description    string
	}{
		// Test various "false" values for alpha flag
		{
			name:           "AlphaFlag_EmptyString_ShouldBeFalse",
			currentVersion: "v1.0.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": ""},
			expected:       "",
			description:    "Empty CALC_ALLOW_ALPHA should deny alpha updates",
		},
		{
			name:           "AlphaFlag_Zero_ShouldBeFalse",
			currentVersion: "v1.0.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "0"},
			expected:       "",
			description:    "CALC_ALLOW_ALPHA=0 should deny alpha updates",
		},
		{
			name:           "AlphaFlag_False_ShouldBeFalse",
			currentVersion: "v1.0.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "false"},
			expected:       "",
			description:    "CALC_ALLOW_ALPHA=false should deny alpha updates",
		},
		{
			name:           "AlphaFlag_FALSE_ShouldBeFalse",
			currentVersion: "v1.0.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "FALSE"},
			expected:       "",
			description:    "CALC_ALLOW_ALPHA=FALSE should deny alpha updates",
		},

		// Test various "true" values for alpha flag
		{
			name:           "AlphaFlag_One_ShouldBeTrue",
			currentVersion: "v1.0.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "1"},
			expected:       "v1.1.0-alpha",
			description:    "CALC_ALLOW_ALPHA=1 should allow alpha updates",
		},
		{
			name:           "AlphaFlag_True_ShouldBeTrue",
			currentVersion: "v1.0.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "true"},
			expected:       "v1.1.0-alpha",
			description:    "CALC_ALLOW_ALPHA=true should allow alpha updates",
		},
		{
			name:           "AlphaFlag_YES_ShouldBeTrue",
			currentVersion: "v1.0.0-alpha",
			envVars:        map[string]string{"CALC_ALLOW_ALPHA": "YES"},
			expected:       "v1.1.0-alpha",
			description:    "CALC_ALLOW_ALPHA=YES should allow alpha updates",
		},

		// Same tests for beta flag
		{
			name:           "BetaFlag_EmptyString_ShouldBeFalse",
			currentVersion: "v1.0.0-beta",
			envVars:        map[string]string{"CALC_ALLOW_BETA": ""},
			expected:       "",
			description:    "Empty CALC_ALLOW_BETA should deny beta updates",
		},
		{
			name:           "BetaFlag_One_ShouldBeTrue",
			currentVersion: "v1.0.0-beta",
			envVars:        map[string]string{"CALC_ALLOW_BETA": "1"},
			expected:       "v1.1.0-beta",
			description:    "CALC_ALLOW_BETA=1 should allow beta updates",
		},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnvVars()

			result := FindLatestEligibleRelease(releases, tt.currentVersion)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected no update due to flag restrictions, but got: %s", result.Version)
				}
			} else {
				if result == nil {
					t.Errorf("Expected update to %s, but got nil", tt.expected)
				} else if result.Version != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result.Version)
				}
			}
		})
	}
}

// Helper function to get channel name from semantic version
func getChannelName(sv *SemanticVersion) string {
	if sv.IsAlpha {
		return "alpha"
	} else if sv.IsBeta {
		return "beta"
	}
	return "stable"
}

func TestComprehensiveCoverageVerification(t *testing.T) {
	// This test verifies that all critical scenarios are covered
	// It's a meta-test that ensures we haven't missed any edge cases

	t.Run("VerifyBugReproduction", func(t *testing.T) {
		// Ensure the original bug is properly tested
		testReleases := []Release{
			{Version: "v0.0.12-alpha", IsAlpha: true, IsBeta: false},
			{Version: "v0.0.10-alpha", IsAlpha: true, IsBeta: false},
			{Version: "v0.0.1-alpha", IsAlpha: true, IsBeta: false},
		}

		clearEnvVars()
		os.Setenv("CALC_ALLOW_ALPHA", "1")
		defer clearEnvVars()

		result := FindLatestEligibleRelease(testReleases, "v0.0.10-alpha")

		if result == nil {
			t.Error("BUG: Alpha user should get an update")
		} else if result.Version == "v0.0.1-alpha" {
			t.Error("ORIGINAL BUG STILL EXISTS: Getting older alpha instead of newer alpha")
		} else if result.Version != "v0.0.12-alpha" {
			t.Errorf("UNEXPECTED RESULT: Expected v0.0.12-alpha, got %s", result.Version)
		} else {
			t.Log("BUG FIX VERIFIED: Correctly getting newer alpha version")
		}
	})

	t.Run("VerifyAllChannelCombinations", func(t *testing.T) {
		// Verify all possible channel combinations are handled
		releases := []Release{
			{Version: "v2.0.0", IsAlpha: false, IsBeta: false},
			{Version: "v1.9.0-beta", IsAlpha: false, IsBeta: true},
			{Version: "v1.8.0-alpha", IsAlpha: true, IsBeta: false},
		}

		testCombinations := []struct {
			currentChannel string
			currentVersion string
			envVars        map[string]string
		}{
			{"stable", "v1.0.0", map[string]string{}},
			{"beta", "v1.0.0-beta", map[string]string{"CALC_ALLOW_BETA": "1"}},
			{"alpha", "v1.0.0-alpha", map[string]string{"CALC_ALLOW_ALPHA": "1"}},
		}

		for _, combo := range testCombinations {
			clearEnvVars()
			for key, value := range combo.envVars {
				os.Setenv(key, value)
			}

			result := FindLatestEligibleRelease(releases, combo.currentVersion)

			if result != nil {
				resultSemVer, _ := ParseSemanticVersion(result.Version)
				resultChannel := getChannelName(resultSemVer)

				if resultChannel != combo.currentChannel {
					t.Errorf("CHANNEL ISOLATION VIOLATION: %s user got %s update",
						combo.currentChannel, resultChannel)
				}
			}
		}
		clearEnvVars()
	})

	t.Run("VerifySemanticVersioningLogic", func(t *testing.T) {
		// Test that semantic versioning works correctly
		versionPairs := []struct {
			newer string
			older string
		}{
			{"v1.0.0", "v0.9.9"},
			{"v0.1.0", "v0.0.9"},
			{"v0.0.2", "v0.0.1"},
			{"v1.0.0", "v1.0.0-beta"},
			{"v1.0.0-beta", "v1.0.0-alpha"},
			{"v0.0.12-alpha", "v0.0.10-alpha"}, // The critical bug case
		}

		for _, pair := range versionPairs {
			newer, err1 := ParseSemanticVersion(pair.newer)
			older, err2 := ParseSemanticVersion(pair.older)

			if err1 != nil || err2 != nil {
				t.Errorf("Failed to parse versions: %s, %s", pair.newer, pair.older)
				continue
			}

			if !newer.IsNewerThan(older) {
				t.Errorf("SEMANTIC VERSION BUG: %s should be newer than %s", pair.newer, pair.older)
			}

			if older.IsNewerThan(newer) {
				t.Errorf("SEMANTIC VERSION BUG: %s should NOT be newer than %s", pair.older, pair.newer)
			}
		}
	})
}

func TestRegressionPrevention(t *testing.T) {
	// This test specifically prevents regression of the original bug
	// and other critical issues

	t.Run("PreventAlphaVersioningRegression", func(t *testing.T) {
		// Test the exact scenario that was failing
		releases := []Release{
			{Version: "v0.0.12-alpha", IsAlpha: true, IsBeta: false},
			{Version: "v0.0.11-alpha", IsAlpha: true, IsBeta: false},
			{Version: "v0.0.10-alpha", IsAlpha: true, IsBeta: false},
			{Version: "v0.0.1-alpha", IsAlpha: true, IsBeta: false},
		}

		clearEnvVars()
		os.Setenv("CALC_ALLOW_ALPHA", "1")
		defer clearEnvVars()

		result := FindLatestEligibleRelease(releases, "v0.0.10-alpha")

		// This MUST pass to prevent regression
		if result == nil {
			t.Fatal("REGRESSION: Alpha user should get an update")
		}

		if result.Version == "v0.0.1-alpha" {
			t.Fatal("REGRESSION: Original bug has returned - getting older alpha!")
		}

		if result.Version != "v0.0.12-alpha" {
			t.Fatalf("REGRESSION: Expected v0.0.12-alpha, got %s", result.Version)
		}
	})

	t.Run("PreventChannelLeakageRegression", func(t *testing.T) {
		// Ensure channel isolation never breaks
		releases := []Release{
			{Version: "v2.0.0", IsAlpha: false, IsBeta: false},
			{Version: "v1.9.0-beta", IsAlpha: false, IsBeta: true},
			{Version: "v1.8.0-alpha", IsAlpha: true, IsBeta: false},
		}

		criticalTests := []struct {
			current string
			env     map[string]string
		}{
			{"v1.0.0-alpha", map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"}},
			{"v1.0.0-beta", map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"}},
			{"v1.0.0", map[string]string{"CALC_ALLOW_ALPHA": "1", "CALC_ALLOW_BETA": "1"}},
		}

		for _, test := range criticalTests {
			clearEnvVars()
			for key, value := range test.env {
				os.Setenv(key, value)
			}

			result := FindLatestEligibleRelease(releases, test.current)

			if result != nil {
				currentSemVer, _ := ParseSemanticVersion(test.current)
				resultSemVer, _ := ParseSemanticVersion(result.Version)

				currentChannel := getChannelName(currentSemVer)
				resultChannel := getChannelName(resultSemVer)

				if currentChannel != resultChannel {
					t.Errorf("REGRESSION: Channel leakage detected! %s user got %s update",
						currentChannel, resultChannel)
				}
			}
		}
		clearEnvVars()
	})
}
