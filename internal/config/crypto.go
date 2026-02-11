package config

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	crypt32                = syscall.NewLazyDLL("crypt32.dll")
	procCryptProtectData   = crypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
)

// DATA_BLOB represents the CRYPT_DATA_BLOB structure
type DATA_BLOB struct {
	CbData uint32
	PbData *byte
}

// EncryptString encrypts a string using Windows DPAPI
func EncryptString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	data := []byte(plaintext)
	blobIn := DATA_BLOB{
		CbData: uint32(len(data)),
		PbData: &data[0],
	}

	var blobOut DATA_BLOB

	// CRYPTPROTECT_UI_FORBIDDEN = 0x1
	// This prevents the system from displaying a dialog box if the master key is not available
	ret, _, err := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(&blobIn)),
		0,            // Description (optional)
		0,            // Optional entropy
		0,            // Reserved
		0,            // Prompt struct
		uintptr(0x1), // CRYPTPROTECT_UI_FORBIDDEN
		uintptr(unsafe.Pointer(&blobOut)),
	)

	if ret == 0 {
		return "", fmt.Errorf("CryptProtectData failed: %v", err)
	}

	// Copy the encrypted data
	encrypted := make([]byte, blobOut.CbData)
	copy(encrypted, (*[1 << 30]byte)(unsafe.Pointer(blobOut.PbData))[:blobOut.CbData:blobOut.CbData])

	// Free the memory allocated by CryptProtectData
	localFree := syscall.NewLazyDLL("kernel32.dll").NewProc("LocalFree")
	localFree.Call(uintptr(unsafe.Pointer(blobOut.PbData)))

	// Encode to base64 for storage
	return string(encrypted), nil
}

// DecryptString decrypts a string using Windows DPAPI
func DecryptString(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	data := []byte(encrypted)
	blobIn := DATA_BLOB{
		CbData: uint32(len(data)),
		PbData: &data[0],
	}

	var blobOut DATA_BLOB

	// CRYPTPROTECT_UI_FORBIDDEN = 0x1
	ret, _, err := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(&blobIn)),
		0,            // Description (optional)
		0,            // Optional entropy
		0,            // Reserved
		0,            // Prompt struct
		uintptr(0x1), // CRYPTPROTECT_UI_FORBIDDEN
		uintptr(unsafe.Pointer(&blobOut)),
	)

	if ret == 0 {
		return "", fmt.Errorf("CryptUnprotectData failed: %v", err)
	}

	// Copy the decrypted data
	decrypted := make([]byte, blobOut.CbData)
	copy(decrypted, (*[1 << 30]byte)(unsafe.Pointer(blobOut.PbData))[:blobOut.CbData:blobOut.CbData])

	// Free the memory allocated by CryptUnprotectData
	localFree := syscall.NewLazyDLL("kernel32.dll").NewProc("LocalFree")
	localFree.Call(uintptr(unsafe.Pointer(blobOut.PbData)))

	return string(decrypted), nil
}

// EncryptAPIKey encrypts an API key for storage
func EncryptAPIKey(apiKey string) (string, error) {
	return EncryptString(apiKey)
}

// DecryptAPIKey decrypts a stored API key
func DecryptAPIKey(encryptedKey string) (string, error) {
	return DecryptString(encryptedKey)
}
