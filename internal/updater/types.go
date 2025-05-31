// Package updater handles automatic updates for the calculator application.
package updater

// Release represents a single software release with version information and download URLs
type Release struct {
	Version     string            `json:"version"`
	URLs        map[string]string `json:"urls"`
	IsAlpha     bool              `json:"isAlpha"`
	IsBeta      bool              `json:"isBeta"`
	ReleaseDate string            `json:"releaseDate"`
}

// VersionManifest contains a collection of software releases
type VersionManifest struct {
	Releases []Release `json:"releases"`
}
