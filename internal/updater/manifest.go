package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const manifestURL = "https://raw.githubusercontent.com/jondkelley/cicd_golang_calculator/main/version.json"

// normalizeVersion removes the "v" prefix from version strings for comparison
func normalizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

// FetchVersionManifest retrieves the version manifest from the remote URL
func FetchVersionManifest() (*VersionManifest, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("no version.json manifest found in project repository")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("no version.json manifest found in project repository (status code = %d)", resp.StatusCode)
	}

	var manifest VersionManifest
	err = json.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		return nil, fmt.Errorf("malformed version.json manifest found")
	}

	if len(manifest.Releases) == 0 {
		return nil, fmt.Errorf("no releases found in manifest")
	}

	return &manifest, nil
}

// SemanticVersion represents a parsed semantic version
type SemanticVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	IsAlpha    bool
	IsBeta     bool
}

// ParseSemanticVersion parses a version string into a SemanticVersion struct
func ParseSemanticVersion(version string) (*SemanticVersion, error) {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	sv := &SemanticVersion{}

	// Check for pre-release suffixes
	if strings.Contains(version, "-alpha") {
		sv.IsAlpha = true
		sv.PreRelease = "alpha"
		version = strings.TrimSuffix(version, "-alpha")
	} else if strings.Contains(version, "-beta") {
		sv.IsBeta = true
		sv.PreRelease = "beta"
		version = strings.TrimSuffix(version, "-beta")
	}

	// Split version into parts
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	var err error
	sv.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	sv.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	sv.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return sv, nil
}

// IsNewerThan compares two semantic versions
// Returns true if sv is newer than other
func (sv *SemanticVersion) IsNewerThan(other *SemanticVersion) bool {
	// Compare major version
	if sv.Major != other.Major {
		return sv.Major > other.Major
	}

	// Compare minor version
	if sv.Minor != other.Minor {
		return sv.Minor > other.Minor
	}

	// Compare patch version
	if sv.Patch != other.Patch {
		return sv.Patch > other.Patch
	}

	// If versions are equal, handle pre-release comparison
	// Stable releases are considered newer than pre-releases
	if sv.PreRelease == "" && other.PreRelease != "" {
		return true
	}
	if sv.PreRelease != "" && other.PreRelease == "" {
		return false
	}

	// Both are pre-releases or both are stable
	if sv.PreRelease != "" && other.PreRelease != "" {
		// Beta is considered newer than alpha
		if sv.IsBeta && other.IsAlpha {
			return true
		}
		if sv.IsAlpha && other.IsBeta {
			return false
		}
	}

	// Versions are identical
	return false
}

// isEnvVarTrue checks if an environment variable should be considered true
// Empty string, "0", "false", "False", "FALSE" are considered false
// Any other non-empty value is considered true
func isEnvVarTrue(envVar string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(envVar)))
	return value != "" && value != "0" && value != "false"
}

// FindLatestEligibleRelease returns the latest release that the user is allowed to install
// and is newer than the current version, respecting channel preferences
func FindLatestEligibleRelease(releases []Release, currentVersion string) *Release {
	allowAlpha := isEnvVarTrue("CALC_ALLOW_ALPHA")
	allowBeta := isEnvVarTrue("CALC_ALLOW_BETA")

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

	var latestEligible *Release
	var latestSemVer *SemanticVersion

	for _, release := range releases {
		// Parse release version first
		releaseSemVer, err := ParseSemanticVersion(release.Version)
		if err != nil {
			fmt.Printf("Warning: Could not parse release version %s: %v\n", release.Version, err)
			continue
		}

		// Skip if this release is not newer than current version
		if !releaseSemVer.IsNewerThan(currentSemVer) {
			continue
		}

		// Determine release channel
		releaseChannel := "stable"
		if release.IsAlpha {
			releaseChannel = "alpha"
		} else if release.IsBeta {
			releaseChannel = "beta"
		}

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

		if !isEligible {
			continue
		}

		// If this is the first eligible release, or if it's newer than our current best
		if latestEligible == nil || releaseSemVer.IsNewerThan(latestSemVer) {
			// Create a copy to avoid pointer issues with loop variable
			releaseCopy := release
			latestEligible = &releaseCopy
			latestSemVer = releaseSemVer
		}
	}

	return latestEligible
}

// getChannelPriority returns priority score for channel matching
// Lower scores mean higher priority (better match)
func getChannelPriority(currentChannel, releaseChannel string) int {
	if currentChannel == releaseChannel {
		return 0 // Perfect match
	}

	switch currentChannel {
	case "alpha":
		if releaseChannel == "beta" {
			return 1 // Alpha user can upgrade to beta
		}
		return 2 // Alpha user can upgrade to stable as last resort
	case "beta":
		if releaseChannel == "stable" {
			return 1 // Beta user can upgrade to stable
		}
		return 3 // Beta user should not downgrade to alpha
	case "stable":
		return 3 // Stable user should never get alpha/beta
	}
	return 3
}