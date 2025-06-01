// Package updater handles automatic updates for the calculator application.
// It checks for new versions from a remote manifest and can download and install updates.
package updater

import (
	"fmt"
	"os"
	"runtime"
)

// UpdateCheckResult represents the result of checking for updates
type UpdateCheckResult struct {
	HasUpdate      bool
	LatestRelease  *Release
	IsGated        bool   // Update exists but requires env var
	RequiredEnvVar string // Which env var is needed
	CurrentChannel string
}

// CheckForUpdate checks for available updates and provides clear messaging
func CheckForUpdate(currentVersion, buildTime string) {
	fmt.Print("Checking for updates... ")

	manifest, err := FetchVersionManifest()
	if err != nil {
		fmt.Printf("🚨 WARNING: %v\n", err)
		return
	}

	if manifest == nil {
		// FetchVersionManifest should have already printed the warning about no internet
		// Just return gracefully without doing anything else so we don't get panic: runtime error: invalid memory address or nil pointer
		return
	}

	result := CheckForUpdatesWithResult(manifest.Releases, currentVersion)

	if !result.HasUpdate {
		fmt.Println("everything is up to date!")
		return
	}

	if result.IsGated {
		// There IS an update, but it's gated behind an environment variable
		releaseType := ""
		if result.LatestRelease.IsAlpha {
			releaseType = "ALPHA RELEASE"
		} else if result.LatestRelease.IsBeta {
			releaseType = "BETA RELEASE"
		}

		fmt.Printf("new version %s (%s) available!\n", result.LatestRelease.Version, releaseType)
		fmt.Printf("To get %s updates, run: %s=1 ./calc\n",
			result.CurrentChannel, result.RequiredEnvVar)
		return
	}

	promptAndUpdate(*result.LatestRelease, currentVersion, buildTime)
}

// CheckForUpdatesWithResult returns detailed information about update availability
func CheckForUpdatesWithResult(releases []Release, currentVersion string) UpdateCheckResult {
	result := UpdateCheckResult{
		HasUpdate: false,
		IsGated:   false,
	}

	// Parse current version to determine channel
	currentSemVer, err := ParseSemanticVersion(currentVersion)
	if err != nil {
		return result
	}

	// Determine current channel
	currentChannel := "stable"
	requiredEnvVar := ""
	if currentSemVer.IsAlpha {
		currentChannel = "alpha"
		requiredEnvVar = "CALC_ALLOW_ALPHA"
	} else if currentSemVer.IsBeta {
		currentChannel = "beta"
		requiredEnvVar = "CALC_ALLOW_BETA"
	}

	result.CurrentChannel = currentChannel
	result.RequiredEnvVar = requiredEnvVar

	// Check if environment variable is set
	allowAlpha := isEnvVarTrue("CALC_ALLOW_ALPHA")
	allowBeta := isEnvVarTrue("CALC_ALLOW_BETA")

	// Find the latest release in the user's current channel
	var latestInChannel *Release
	var latestInChannelSemVer *SemanticVersion

	for _, release := range releases {
		releaseSemVer, err := ParseSemanticVersion(release.Version)
		if err != nil {
			continue
		}

		// Skip if not newer than current
		if !releaseSemVer.IsNewerThan(currentSemVer) {
			continue
		}

		// Check if this release is in the same channel
		releaseChannel := "stable"
		if release.IsAlpha {
			releaseChannel = "alpha"
		} else if release.IsBeta {
			releaseChannel = "beta"
		}

		if releaseChannel == currentChannel {
			// This is a newer release in the same channel
			if latestInChannel == nil || releaseSemVer.IsNewerThan(latestInChannelSemVer) {
				releaseCopy := release
				latestInChannel = &releaseCopy
				latestInChannelSemVer = releaseSemVer
			}
		}
	}

	// If we found a newer release in the same channel
	if latestInChannel != nil {
		result.HasUpdate = true
		result.LatestRelease = latestInChannel

		// Check if it's gated by environment variables
		switch currentChannel {
		case "alpha":
			result.IsGated = !allowAlpha
		case "beta":
			result.IsGated = !allowBeta
		case "stable":
			result.IsGated = false // Stable updates are never gated
		}
	}

	return result
}

// Rest of the functions remain the same...
func promptAndUpdate(latestRelease Release, currentVersion, buildTime string) {
	releaseTypeWarning := ""
	if latestRelease.IsAlpha {
		releaseTypeWarning = " (ALPHA RELEASE)"
	} else if latestRelease.IsBeta {
		releaseTypeWarning = " (BETA RELEASE)"
	}

	fmt.Printf("new version %s%s available, would you like to update? Y[es]/N[o]: ", latestRelease.Version, releaseTypeWarning)
	var input string
	fmt.Scanln(&input)

	if input == "y" || input == "Y" || input == "yes" || input == "Yes" {
		fmt.Printf("Updating current version from %s to %s\n", currentVersion, latestRelease.Version)
		UpdateBinary(latestRelease, currentVersion, buildTime)
	} else {
		fmt.Println("update cancelled.")
	}
}

func UpdateBinary(release Release, currentVersion, buildTime string) {
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

	backupPath := execPath + "." + currentVersion + ".bak"
	if err := CopyFile(execPath, backupPath); err != nil {
		fmt.Println("Failed to create backup:", err)
		return
	}

	// Download new version
	if !DownloadBinary(url, execPath, release) {
		return
	}

	fmt.Printf("✅ Update complete from %s (built %s) to %s. Please re-run the application.\n", currentVersion, buildTime, release.Version)
	os.Exit(0)
}
