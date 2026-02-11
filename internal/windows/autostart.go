package windows

import (
	"golang.org/x/sys/windows/registry"
)

const (
	autostartKey = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	appName      = "The Copy Father"
)

// SetAutoStart configures whether the application should start automatically with Windows
func SetAutoStart(enabled bool, exePath string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, autostartKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if enabled {
		// Add to startup
		err = key.SetStringValue(appName, exePath)
	} else {
		// Remove from startup
		err = key.DeleteValue(appName)
	}

	return err
}

// IsAutoStartEnabled checks if auto-start is currently enabled
func IsAutoStartEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, autostartKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appName)
	return err == nil
}
