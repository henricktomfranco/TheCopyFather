package windows

import (
	"os"
	"path/filepath"
)

const (
	appName = "The Copy Father"
)

// SetAutoStart configures whether the application should start automatically with the system
func SetAutoStart(enabled bool, exePath string) error {
	// For Linux, we'll use systemd or desktop autostart entries
	// This is a simplified implementation
	if enabled {
		// Create autostart directory if it doesn't exist
		autostartDir := filepath.Join(os.Getenv("HOME"), ".config", "autostart")
		err := os.MkdirAll(autostartDir, 0755)
		if err != nil {
			return err
		}

		// Create desktop entry
		desktopEntry := filepath.Join(autostartDir, "thecopyfather.desktop")
		content := `[Desktop Entry]
Type=Application
Name=The Copy Father
Exec=` + exePath + `
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
`

		return os.WriteFile(desktopEntry, []byte(content), 0644)
	} else {
		// Remove desktop entry
		autostartDir := filepath.Join(os.Getenv("HOME"), ".config", "autostart")
		desktopEntry := filepath.Join(autostartDir, "thecopyfather.desktop")
		return os.Remove(desktopEntry)
	}
}

// IsAutoStartEnabled checks if auto-start is currently enabled
func IsAutoStartEnabled() bool {
	autostartDir := filepath.Join(os.Getenv("HOME"), ".config", "autostart")
	desktopEntry := filepath.Join(autostartDir, "thecopyfather.desktop")
	_, err := os.Stat(desktopEntry)
	return err == nil
}