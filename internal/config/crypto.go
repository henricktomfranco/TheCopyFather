package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// keyFileName is the name of the file that stores the encryption key
const keyFileName = "encryption_key.bin"

// getKeyPath returns the path to the encryption key file
func getKeyPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = os.Getenv("USERPROFILE")
		if appData == "" {
			appData = "."
		}
	}
	configDir := filepath.Join(appData, "TheCopyfather")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, keyFileName)
}

// generateOrLoadKey generates a new encryption key or loads an existing one from disk
// The key is derived from a combination of machine-specific data and random bytes
func generateOrLoadKey() ([]byte, error) {
	keyPath := getKeyPath()
	
	// Try to read existing key
	keyData, err := os.ReadFile(keyPath)
	if err == nil {
		// Key exists, use it
		return keyData, nil
	}
	
	// Generate new key - use SHA256 hash of machine ID + random bytes for consistent but secure key
	// This provides a balance between security and usability (key persists across restarts)
	machineID := getMachineID()
	hash := sha256.New()
	hash.Write([]byte(machineID))
	hash.Write([]byte("TheCopyfather_Salt_"))
	
	// Add random bytes for additional entropy
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	hash.Write(randomBytes)
	
	key := hash.Sum(nil)
	
	// Save the key for future use
	// We save the hashed key, not the raw machine ID + random bytes
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to save encryption key: %w", err)
	}
	
	return key, nil
}

// getMachineID returns a machine-specific identifier
// On Windows, this uses the COMPUTERNAME environment variable
// On other platforms, it uses HOSTNAME
func getMachineID() string {
	// Try Windows-specific first
	if computerName := os.Getenv("COMPUTERNAME"); computerName != "" {
		return computerName
	}
	// Try Unix-like systems
	if hostName := os.Getenv("HOSTNAME"); hostName != "" {
		return hostName
	}
	// Fallback to user name
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	// Last resort - use a fixed string with process ID
	return fmt.Sprintf("default_%d", os.Getpid())
}

// EncryptString encrypts a string using AES-GCM
func EncryptString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key, err := generateOrLoadKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a string using AES-GCM
func DecryptString(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	key, err := generateOrLoadKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptAPIKey encrypts an API key for storage
func EncryptAPIKey(apiKey string) (string, error) {
	return EncryptString(apiKey)
}

// DecryptAPIKey decrypts a stored API key
func DecryptAPIKey(encryptedKey string) (string, error) {
	return DecryptString(encryptedKey)
}
