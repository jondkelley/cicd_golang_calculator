// Package updater handles automatic updates for the calculator application.
// It checks for new versions from a remote manifest and can download and install updates.
package updater

import (
	"fmt"
	"os"
	"runtime"
)

// CheckForUpdate checks for available updates and prompts the user to update if a new version is found
func CheckForUpdate(currentVersion, buildTime string) {
	fmt.Print("Checking for updates... ")

	manifest, err := FetchVersionManifest()
	if err != nil {
		fmt.Printf("🚨 WARNING: %v\n", err)
		return
	}

	latestRelease := FindLatestEligibleRelease(manifest.Releases, currentVersion)
	if latestRelease == nil {
		fmt.Println("everything is up to date!")
		return
	}

	if latestRelease.Version == currentVersion {
		fmt.Println("you are running latest version.")
		return
	}

	promptAndUpdate(*latestRelease, currentVersion, buildTime)
}

// promptAndUpdate shows the update prompt and handles the update process
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

// UpdateBinary downloads and installs the new binary version
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