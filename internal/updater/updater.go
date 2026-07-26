package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitHubRepo is the repository for TheCopyFather.
// Override at build time: -ldflags "-X textrewriter/internal/updater.GitHubRepo=owner/repo"
var GitHubRepo = "yourusername/TheCopyFather"

// ReleaseAPIURL is the GitHub Releases API endpoint
var ReleaseAPIURL = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

// UpdateInfo holds the latest version and download URL
type UpdateInfo struct {
	LatestVersion  string
	CurrentVersion string
	DownloadURL    string
	Available      bool
	Error          string
}

// CheckForUpdates checks if a new version is available
func CheckForUpdates(currentVersion string) *UpdateInfo {
	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		Available:      false,
	}

	release, err := fetchLatestRelease()
	if err != nil {
		info.Error = fmt.Sprintf("Failed to fetch latest release: %v", err)
		return info
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	info.LatestVersion = latestVersion
	info.DownloadURL = getDownloadURL(release.Assets)

	// Compare versions using semver (latest > current)
	if isNewerVersion(latestVersion, currentVersion) {
		info.Available = true
	}

	return info
}

// isNewerVersion returns true if latest is strictly newer than current (e.g. 1.2.0 > 1.1.0)
func isNewerVersion(latest, current string) bool {
	if latest == "" || current == "" || current == "dev" {
		return latest != "" && latest != current
	}
	latestParts := parseSemver(latest)
	currentParts := parseSemver(current)

	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	var res [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		var val int
		fmt.Sscanf(parts[i], "%d", &val)
		res[i] = val
	}
	return res
}

// fetchLatestRelease fetches the latest release from GitHub
func fetchLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", ReleaseAPIURL, nil)
	if err != nil {
		return nil, err
	}

	// GitHub API requires a User-Agent header
	req.Header.Set("User-Agent", "TheCopyFather-Updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status: %s, body: %s", resp.Status, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// Asset represents a GitHub release asset
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// getDownloadURL finds the download URL for the Windows executable
func getDownloadURL(assets []Asset) string {
	for _, asset := range assets {
		if strings.EqualFold(asset.Name, "thecopyfather.exe") ||
			strings.EqualFold(asset.Name, "TheCopyFather.exe") {
			return asset.URL
		}
	}
	return ""
}

// DownloadUpdate downloads the latest executable to a temporary directory
func DownloadUpdate(downloadURL string) (string, error) {
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "thecopyfather_update.exe")

	// Remove any existing temp file
	if _, err := os.Stat(tempFile); err == nil {
		os.Remove(tempFile)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download: HTTP %s", resp.Status)
	}

	out, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %v", err)
	}

	return tempFile, nil
}

// InstallUpdate replaces the old binary with the new one and restarts the app
// Returns the temp file path and current exe path for the caller to handle restart
func GetUpdatePaths() (string, string, error) {
	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current executable path: %v", err)
	}

	// Generate temp file path
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "thecopyfather_update.exe")

	return tempFile, exePath, nil
}

// InstallUpdate replaces the old binary with the new one
func InstallUpdate(tempFile, targetPath string) error {
	// Ensure the temp file exists
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		return fmt.Errorf("update file not found: %s", tempFile)
	}

	// Close the current app (caller should handle this)
	// Replace the old binary
	if err := os.Rename(tempFile, targetPath); err != nil {
		// On Windows, os.Rename fails if target exists, so we delete first
		if os.IsExist(err) {
			if err := os.Remove(targetPath); err != nil {
				return fmt.Errorf("failed to remove old binary: %v", err)
			}
			if err := os.Rename(tempFile, targetPath); err != nil {
				return fmt.Errorf("failed to rename new binary: %v", err)
			}
		} else {
			return fmt.Errorf("failed to replace binary: %v", err)
		}
	}

	return nil
}

// RestartApp restarts the application
func RestartApp(exePath string) error {
	cmd := exec.Command(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to restart: %v", err)
	}

	// Exit the current process
	os.Exit(0)
	return nil
}

// GetLatestVersionFromString extracts the version from a tag string
func GetLatestVersionFromString(tagName string) string {
	// Remove 'v' prefix if present
	version := strings.TrimPrefix(tagName, "v")
	// Extract only the version number (remove any suffix like -beta)
	// This is a simple implementation; you might want to use a regex for more robustness
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}
	return version
}

// getExePath is a variable function for testing
var getExePath = func() (string, error) {
	return os.Executable()
}

// getTempDir is a variable function for testing
var getTempDir = func() string {
	return os.TempDir()
}
