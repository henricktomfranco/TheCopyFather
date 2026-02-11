package windows

import (
	_ "embed" // Added to support //go:embed directive
	"fmt"
	"runtime"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var iconData []byte

// TrayManager handles system tray functionality
type TrayManager struct {
	onShowSettings func()
	onExit         func()
	started        bool
}

// NewTrayManager creates a new tray manager
func NewTrayManager() *TrayManager {
	return &TrayManager{}
}

// OnShowSettings sets the callback for showing settings
func (t *TrayManager) OnShowSettings(callback func()) {
	t.onShowSettings = callback
}

// OnExit sets the callback for exit
func (t *TrayManager) OnExit(callback func()) {
	t.onExit = callback
}

// Start starts the system tray
func (t *TrayManager) Start() {
	if t.started {
		return
	}

	t.started = true

	go func() {
		// Ensure systray runs on main thread
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		systray.Run(t.onReady, t.onExitFunc)
	}()
}

func (t *TrayManager) onReady() {
	// Set tooltip first so something shows up
	systray.SetTooltip("The Copyfather - I'm gonna make you an offer you can't refuse... better writing")

	// Set icon gracefully
	if len(iconData) > 0 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[TRAY] Recovered from SetIcon panic: %v\n", r)
				}
			}()
			// Create a clean copy of the bytes to avoid any weirdness with go:embed memory mapping
			cleanIcon := make([]byte, len(iconData))
			copy(cleanIcon, iconData)
			systray.SetIcon(cleanIcon)
		}()
	}

	// Add menu items
	mSettings := systray.AddMenuItem("Settings", "Open Settings")
	mSeparator := systray.AddMenuItem("", "")
	mSeparator.Disable()
	mExit := systray.AddMenuItem("Exit", "Quit the application")

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mSettings.ClickedCh:
				if t.onShowSettings != nil {
					t.onShowSettings()
				}
			case <-mExit.ClickedCh:
				if t.onExit != nil {
					t.onExit()
				}
				systray.Quit()
				return
			}
		}
	}()
}

func (t *TrayManager) onExitFunc() {
	// Cleanup
}

// Stop stops the system tray
func (t *TrayManager) Stop() {
	if t.started {
		systray.Quit()
		t.started = false
	}
}

// ShowNotification shows a balloon notification
func (t *TrayManager) ShowNotification(title, message string) {
	systray.SetTooltip(message)
}

func getIconBytes() []byte {
	return iconData
}
