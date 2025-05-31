package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const manifestURL = "https://raw.githubusercontent.com/jondkelley/cicd_golang_calculator/main/version.json"

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

// FindLatestEligibleRelease returns the latest release that the user is allowed to install
func FindLatestEligibleRelease(releases []Release) *Release {
	allowAlpha := os.Getenv("CALC_ALLOW_ALPHA") != ""
	allowBeta := os.Getenv("CALC_ALLOW_BETA") != ""

	for _, release := range releases {
		// Skip alpha releases unless allowed
		if release.IsAlpha && !allowAlpha {
			continue
		}

		// Skip beta releases unless allowed
		if release.IsBeta && !allowBeta {
			continue
		}

		// Return the first (latest) eligible release
		return &release
	}

	return nil
}
