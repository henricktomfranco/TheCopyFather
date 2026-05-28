package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEncryptDecryptString tests encryption and decryption of strings
func TestEncryptDecryptString(t *testing.T) {
	// Set up a temporary directory for the encryption key
	tempDir, err := os.MkdirTemp("", "crypto_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the key path for testing
	originalAppData := os.Getenv("APPDATA")
	originalUserProfile := os.Getenv("USERPROFILE")
	originalHome := os.Getenv("HOME")

	os.Setenv("APPDATA", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	os.Setenv("HOME", tempDir)
	defer func() {
		os.Setenv("APPDATA", originalAppData)
		os.Setenv("USERPROFILE", originalUserProfile)
		os.Setenv("HOME", originalHome)
	}()

	// Test cases
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"simple text", "hello world", "hello world"},
		{"special characters", "Hello, 世界! 🌍", "Hello, 世界! 🌍"},
		{"long text", "This is a longer text to test encryption and decryption. It should handle longer strings without issues.", "This is a longer text to test encryption and decryption. It should handle longer strings without issues."},
		{"api key", "sk-1234567890abcdef", "sk-1234567890abcdef"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt the string
			encrypted, err := EncryptString(tc.input)
			if err != nil {
				t.Fatalf("Failed to encrypt string: %v", err)
			}

			// Decrypt the string
			decrypted, err := DecryptString(encrypted)
			if err != nil {
				t.Fatalf("Failed to decrypt string: %v", err)
			}

			// Verify the decrypted string matches the input
			if decrypted != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, decrypted)
			}

			// Verify that encrypted string is different from input (for non-empty strings)
			if tc.input != "" && encrypted == tc.input {
				t.Error("Encrypted string should be different from input for non-empty strings")
			}
		})
	}
}

// TestEncryptDecryptAPIKey tests encryption and decryption of API keys
func TestEncryptDecryptAPIKey(t *testing.T) {
	// Set up a temporary directory for the encryption key
	tempDir, err := os.MkdirTemp("", "crypto_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the key path for testing
	originalAppData := os.Getenv("APPDATA")
	originalUserProfile := os.Getenv("USERPROFILE")
	originalHome := os.Getenv("HOME")

	os.Setenv("APPDATA", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	os.Setenv("HOME", tempDir)
	defer func() {
		os.Setenv("APPDATA", originalAppData)
		os.Setenv("USERPROFILE", originalUserProfile)
		os.Setenv("HOME", originalHome)
	}()

	// Test API key encryption/decryption
	apiKey := "sk-test-api-key-1234567890"

	encrypted, err := EncryptAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to encrypt API key: %v", err)
	}

	decrypted, err := DecryptAPIKey(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt API key: %v", err)
	}

	if decrypted != apiKey {
		t.Errorf("Expected API key '%s', got '%s'", apiKey, decrypted)
	}
}

// TestKeyPersistence tests that the encryption key persists across multiple operations
func TestKeyPersistence(t *testing.T) {
	// Set up a temporary directory for the encryption key
	tempDir, err := os.MkdirTemp("", "crypto_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the key path for testing
	originalAppData := os.Getenv("APPDATA")
	originalUserProfile := os.Getenv("USERPROFILE")
	originalHome := os.Getenv("HOME")

	os.Setenv("APPDATA", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	os.Setenv("HOME", tempDir)
	defer func() {
		os.Setenv("APPDATA", originalAppData)
		os.Setenv("USERPROFILE", originalUserProfile)
		os.Setenv("HOME", originalHome)
	}()

	// Encrypt a string
	input := "test persistence"
	encrypted1, err := EncryptString(input)
	if err != nil {
		t.Fatalf("Failed to encrypt string: %v", err)
	}

	// Decrypt the string
	decrypted1, err := DecryptString(encrypted1)
	if err != nil {
		t.Fatalf("Failed to decrypt string: %v", err)
	}

	// Encrypt again (should use the same key)
	encrypted2, err := EncryptString(input)
	if err != nil {
		t.Fatalf("Failed to encrypt string again: %v", err)
	}

	// The encrypted strings should be different (due to random nonce)
	// but both should decrypt to the same value
	decrypted2, err := DecryptString(encrypted2)
	if err != nil {
		t.Fatalf("Failed to decrypt string again: %v", err)
	}

	if decrypted1 != decrypted2 {
		t.Errorf("Expected both decrypted values to match, got '%s' and '%s'", decrypted1, decrypted2)
	}

	// Check that the key file was created
	keyPath := filepath.Join(tempDir, "TheCopyfather", "encryption_key.bin")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("Expected encryption key file to be created")
	}
}

// TestGetMachineID tests the getMachineID function
func TestGetMachineID(t *testing.T) {
	// This test just verifies that getMachineID doesn't panic
	// and returns a non-empty string
	machineID := getMachineID()
	if machineID == "" {
		t.Error("Expected getMachineID to return a non-empty string")
	}
	
	// On most systems, this should return a reasonable identifier
	t.Logf("Machine ID: %s", machineID)
}

// TestGenerateOrLoadKey tests the key generation and loading
func TestGenerateOrLoadKey(t *testing.T) {
	// Set up a temporary directory for the encryption key
	tempDir, err := os.MkdirTemp("", "crypto_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the key path for testing
	originalAppData := os.Getenv("APPDATA")
	originalUserProfile := os.Getenv("USERPROFILE")
	originalHome := os.Getenv("HOME")

	os.Setenv("APPDATA", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	os.Setenv("HOME", tempDir)
	defer func() {
		os.Setenv("APPDATA", originalAppData)
		os.Setenv("USERPROFILE", originalUserProfile)
		os.Setenv("HOME", originalHome)
	}()

	// Generate or load a key
	key1, err := generateOrLoadKey()
	if err != nil {
		t.Fatalf("Failed to generate or load key: %v", err)
	}

	// Key should be 32 bytes (SHA256 hash)
	if len(key1) != 32 {
		t.Errorf("Expected key length to be 32 bytes, got %d", len(key1))
	}

	// Generate or load the key again (should load from file)
	key2, err := generateOrLoadKey()
	if err != nil {
		t.Fatalf("Failed to generate or load key again: %v", err)
	}

	// Both keys should be the same
	if string(key1) != string(key2) {
		t.Error("Expected both generated keys to be the same")
	}
}
