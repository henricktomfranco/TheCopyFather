package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestCheckForUpdates tests the CheckForUpdates function
func TestCheckForUpdates(t *testing.T) {
	// Create a mock HTTP server for GitHub API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a request to the releases endpoint
		if r.URL.Path == "/repos/TheCopyFather/TheCopyFather/releases/latest" {
			release := GitHubRelease{
				TagName: "v2.0.0",
				Name:    "Release 2.0.0",
				Assets: []Asset{
					{Name: "thecopyfather.exe", URL: "https://github.com/TheCopyFather/TheCopyFather/releases/download/v2.0.0/thecopyfather.exe"},
				},
			}
			json.NewEncoder(w).Encode(release)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Override ReleaseAPIURL for testing
	originalReleaseAPIURL := ReleaseAPIURL
	ReleaseAPIURL = server.URL + "/repos/TheCopyFather/TheCopyFather/releases/latest"
	defer func() { ReleaseAPIURL = originalReleaseAPIURL }()

	// Test case 1: Update available
	t.Run("update available", func(t *testing.T) {
		info := CheckForUpdates("v1.0.0")
		if info.Error != "" {
			t.Fatalf("CheckForUpdates failed: %v", info.Error)
		}

		if !info.Available {
			t.Error("Expected update to be available")
		}
		if info.LatestVersion != "2.0.0" {
			t.Errorf("Expected latest version '2.0.0', got '%s'", info.LatestVersion)
		}
		if info.CurrentVersion != "v1.0.0" {
			t.Errorf("Expected current version 'v1.0.0', got '%s'", info.CurrentVersion)
		}
		if info.DownloadURL == "" {
			t.Error("Expected download URL to be set")
		}
	})

	// Test case 2: No update available (same version)
	t.Run("no update available", func(t *testing.T) {
		info := CheckForUpdates("2.0.0")
		if info.Error != "" {
			t.Fatalf("CheckForUpdates failed: %v", info.Error)
		}

		if info.Available {
			t.Error("Expected no update to be available")
		}
		if info.LatestVersion != "2.0.0" {
			t.Errorf("Expected latest version '2.0.0', got '%s'", info.LatestVersion)
		}
	})

	// Test case 3: Server error
	t.Run("server error", func(t *testing.T) {
		// Create a server that returns an error
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer errorServer.Close()

		originalReleaseAPIURL := ReleaseAPIURL
		ReleaseAPIURL = errorServer.URL + "/repos/TheCopyFather/TheCopyFather/releases/latest"
		defer func() { ReleaseAPIURL = originalReleaseAPIURL }()

		info := CheckForUpdates("v1.0.0")
		if info.Error == "" {
			t.Error("Expected error from server, got nil")
		}
	})
}

// TestFetchLatestRelease tests the fetchLatestRelease function
func TestFetchLatestRelease(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := GitHubRelease{
			TagName: "v1.5.0",
			Name:    "Release 1.5.0",
			Assets: []Asset{
				{Name: "thecopyfather.exe", URL: "https://example.com/thecopyfather.exe"},
			},
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Override ReleaseAPIURL for testing
	originalReleaseAPIURL := ReleaseAPIURL
	ReleaseAPIURL = server.URL + "/repos/TheCopyFather/TheCopyFather/releases/latest"
	defer func() { ReleaseAPIURL = originalReleaseAPIURL }()

	release, err := fetchLatestRelease()
	if err != nil {
		t.Fatalf("fetchLatestRelease failed: %v", err)
	}

	if release.TagName != "v1.5.0" {
		t.Errorf("Expected tag name 'v1.5.0', got '%s'", release.TagName)
	}
	if len(release.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(release.Assets))
	}
	if release.Assets[0].Name != "thecopyfather.exe" {
		t.Errorf("Expected asset name 'thecopyfather.exe', got '%s'", release.Assets[0].Name)
	}
}

// TestGetDownloadURL tests the getDownloadURL function
func TestGetDownloadURL(t *testing.T) {
	assets := []Asset{
		{Name: "readme.txt", URL: "https://example.com/readme.txt"},
		{Name: "thecopyfather.exe", URL: "https://example.com/thecopyfather.exe"},
		{Name: "TheCopyFather.exe", URL: "https://example.com/TheCopyFather.exe"},
		{Name: "config.json", URL: "https://example.com/config.json"},
	}

	// Test with lowercase filename
	url := getDownloadURL(assets)
	if url != "https://example.com/thecopyfather.exe" {
		t.Errorf("Expected URL 'https://example.com/thecopyfather.exe', got '%s'", url)
	}

	// Test with uppercase filename
	assetsUpper := []Asset{
		{Name: "TheCopyFather.exe", URL: "https://example.com/TheCopyFather.exe"},
	}
	url = getDownloadURL(assetsUpper)
	if url != "https://example.com/TheCopyFather.exe" {
		t.Errorf("Expected URL 'https://example.com/TheCopyFather.exe', got '%s'", url)
	}

	// Test with no matching asset
	assetsNoMatch := []Asset{
		{Name: "readme.txt", URL: "https://example.com/readme.txt"},
	}
	url = getDownloadURL(assetsNoMatch)
	if url != "" {
		t.Errorf("Expected empty URL, got '%s'", url)
	}
	
	// Test with empty assets
	url = getDownloadURL([]Asset{})
	if url != "" {
		t.Errorf("Expected empty URL for empty assets, got '%s'", url)
	}
}

// TestGetUpdatePaths tests the GetUpdatePaths function
func TestGetUpdatePaths(t *testing.T) {
	// Save original function and restore after test
	originalGetExePath := getExePath
	originalGetTempDir := getTempDir
	defer func() {
		getExePath = originalGetExePath
		getTempDir = originalGetTempDir
	}()

	// Mock the functions
	getExePath = func() (string, error) {
		return "/path/to/thecopyfather.exe", nil
	}
	getTempDir = func() string {
		return "/tmp"
	}

	currentExe, tempDir, err := GetUpdatePaths()
	if err != nil {
		t.Fatalf("GetUpdatePaths failed: %v", err)
	}

	if currentExe != "/path/to/thecopyfather.exe" {
		t.Errorf("Expected current exe '/path/to/thecopyfather.exe', got '%s'", currentExe)
	}
	if tempDir != "/tmp" {
		t.Errorf("Expected temp dir '/tmp', got '%s'", tempDir)
	}
}

// TestDownloadUpdate tests the DownloadUpdate function
func TestDownloadUpdate(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock executable content"))
	}))
	defer server.Close()

	// Create a temporary directory for the download
	tempDir, err := os.MkdirTemp("", "updater_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock getTempDir to return our temp directory
	originalGetTempDir := getTempDir
	getTempDir = func() string {
		return tempDir
	}
	defer func() { getTempDir = originalGetTempDir }()

	// Test downloading from the mock server
	tempFile, err := DownloadUpdate(server.URL)
	if err != nil {
		t.Fatalf("DownloadUpdate failed: %v", err)
	}

	// Check that the file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Expected downloaded file to exist")
	}

	// Check that the file contains the expected content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != "mock executable content" {
		t.Errorf("Expected file content 'mock executable content', got '%s'", string(content))
	}

	// Test with a server that returns an error
	t.Run("server error", func(t *testing.T) {
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}))
		defer errorServer.Close()

		_, err := DownloadUpdate(errorServer.URL)
		if err == nil {
			t.Error("Expected error from server, got nil")
		}
	})
}

// TestInstallUpdate tests the InstallUpdate function
func TestInstallUpdate(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "updater_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock current executable
	currentExePath := filepath.Join(tempDir, "current.exe")
	if err := os.WriteFile(currentExePath, []byte("current content"), 0755); err != nil {
		t.Fatalf("Failed to create mock current exe: %v", err)
	}

	// Create a mock temp file with new content
	tempFilePath := filepath.Join(tempDir, "update.exe")
	if err := os.WriteFile(tempFilePath, []byte("new content"), 0755); err != nil {
		t.Fatalf("Failed to create mock temp file: %v", err)
	}

	// Note: We can't actually test the full InstallUpdate function
	// because it restarts the application, but we can test the file replacement
	t.Run("file replacement", func(t *testing.T) {
		// Read the content before replacement
		originalContent, _ := os.ReadFile(currentExePath)
		if string(originalContent) != "current content" {
			t.Fatalf("Expected original content 'current content', got '%s'", string(originalContent))
		}

		// Manually replace the file (simulating what InstallUpdate does)
		if err := os.Rename(tempFilePath, currentExePath); err != nil {
			t.Fatalf("Failed to replace file: %v", err)
		}

		// Check that the file was replaced
		newContent, err := os.ReadFile(currentExePath)
		if err != nil {
			t.Fatalf("Failed to read replaced file: %v", err)
		}

		if string(newContent) != "new content" {
			t.Errorf("Expected new content 'new content', got '%s'", string(newContent))
		}
	})
}

// TestGetLatestVersionFromString tests the GetLatestVersionFromString function
func TestGetLatestVersionFromString(t *testing.T) {
	testCases := []struct {
		versionString string
		expected      string
	}{
		{"v1.0.0", "1.0.0"},
		{"v2.5.1", "2.5.1"},
		{"1.2.3", "1.2.3"},
		{"release-3.0.0", "3.0.0"},
		{"v1.0.0-beta", "1.0.0-beta"},
		{"invalid", ""},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.versionString, func(t *testing.T) {
			result := GetLatestVersionFromString(tc.versionString)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
