package windows

import (
	"fmt"
	"runtime"
	"strings"
	"time"
	"unsafe"
)

var (
	getAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	getForegroundWindow      = user32.NewProc("GetForegroundWindow")
	getWindowRect            = user32.NewProc("GetWindowRect")
	getCursorPos             = user32.NewProc("GetCursorPos")
	registerHotKey           = user32.NewProc("RegisterHotKey")
	unregisterHotKey         = user32.NewProc("UnregisterHotKey")
	getMessage               = user32.NewProc("GetMessageW")
	peekMessage              = user32.NewProc("PeekMessageW")
	translateMessage         = user32.NewProc("TranslateMessage")
	dispatchMessage          = user32.NewProc("DispatchMessageW")
	openClipboard            = user32.NewProc("OpenClipboard")
	closeClipboard           = user32.NewProc("CloseClipboard")
	getClipboardData         = user32.NewProc("GetClipboardData")
	setClipboardData         = user32.NewProc("SetClipboardData")
	emptyClipboard           = user32.NewProc("EmptyClipboard")
	registerClipboardFormatW = user32.NewProc("RegisterClipboardFormatW")
	globalLock               = kernel32.NewProc("GlobalLock")
	globalUnlock             = kernel32.NewProc("GlobalUnlock")
	globalAlloc              = kernel32.NewProc("GlobalAlloc")
	globalFree               = kernel32.NewProc("GlobalFree")
	lstrlenW                 = kernel32.NewProc("lstrlenW")
	sendInput                = user32.NewProc("SendInput")
)

const (
	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002
	MOD_ALT        = 0x0001
	MOD_CONTROL    = 0x0002
	MOD_SHIFT      = 0x0004
	MOD_WIN        = 0x0008
	WM_HOTKEY      = 0x0312
	HOTKEY_ID      = 0x1234
	PM_REMOVE      = 0x0001

	// Input types
	INPUT_KEYBOARD = 1

	// Key codes
	VK_SHIFT   = 0x10
	VK_CONTROL = 0x11
	VK_ALT     = 0x12
	VK_C       = 0x43
	VK_V       = 0x56
)

type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
}

type KEYBDINPUT struct {
	WVK         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// HotkeyManager handles global hotkey registration
type HotkeyManager struct {
	hotkeys      map[string]func()
	running      bool
	stopChan     chan bool
	registerChan chan hotkeyRequest
}

type hotkeyRequest struct {
	combo    string
	callback func()
	errChan  chan error
}

// NewHotkeyManager creates a new hotkey manager
func NewHotkeyManager() *HotkeyManager {
	h := &HotkeyManager{
		hotkeys:      make(map[string]func()),
		stopChan:     make(chan bool),
		registerChan: make(chan hotkeyRequest),
	}
	go h.listen()
	return h
}

// Register registers a global hotkey. It clears any previous registrations for this manager.
func (h *HotkeyManager) Register(combo string, callback func()) error {
	// Clear all existing callbacks before registering new one
	// This ensures we don't have multiple handlers for the same hotkey
	for k := range h.hotkeys {
		delete(h.hotkeys, k)
	}

	errChan := make(chan error)
	h.registerChan <- hotkeyRequest{
		combo:    combo,
		callback: callback,
		errChan:  errChan,
	}
	return <-errChan
}

// Stop unregisters all hotkeys
func (h *HotkeyManager) Stop() {
	if !h.running {
		return
	}
	h.running = false
	close(h.stopChan)
}

func (h *HotkeyManager) listen() {
	// Lock this goroutine to a specific OS thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	h.running = true

	// Small wait to ensure thread is ready to receive messages
	time.Sleep(100 * time.Millisecond)

	for h.running {
		var msg struct {
			HWnd    uintptr
			Message uint32
			WParam  uintptr
			LParam  uintptr
			Time    uint32
			Pt      struct{ X, Y int32 }
		}

		select {
		case <-h.stopChan:
			unregisterHotKey.Call(0, uintptr(HOTKEY_ID))
			return
		case req := <-h.registerChan:
			mods, vk, err := parseHotkey(req.combo)
			if err != nil {
				req.errChan <- err
				continue
			}

			h.hotkeys[req.combo] = req.callback

			// Unregister existing if any (ignore error)
			unregisterHotKey.Call(0, uintptr(HOTKEY_ID))

			// Register with Windows on THIS thread
			ret, _, err := registerHotKey.Call(
				0,
				uintptr(HOTKEY_ID),
				uintptr(mods),
				uintptr(vk),
			)

			if ret == 0 {
				req.errChan <- fmt.Errorf("failed to register hotkey %s: %v", req.combo, err)
			} else {
				fmt.Printf("[HOTKEY] Successfully registered hotkey: %s with mods: %v, vk: %v\n", req.combo, mods, vk)
				req.errChan <- nil
			}
		default:
			// Using PeekMessage with PM_REMOVE to handle thread messages
			ret, _, _ := peekMessage.Call(
				uintptr(unsafe.Pointer(&msg)),
				0,
				0,
				0,
				uintptr(PM_REMOVE),
			)

			if ret != 0 {
				// Log every message to console for deep debugging
				if msg.Message == WM_HOTKEY {
					fmt.Printf("[HOTKEY] EXPLICIT WM_HOTKEY DETECTED: wParam=%v, lParam=%v\n", msg.WParam, msg.LParam)
					if msg.WParam == uintptr(HOTKEY_ID) {
						fmt.Println("[HOTKEY] Match found for HOTKEY_ID. Triggering callbacks...")
						for _, callback := range h.hotkeys {
							go callback()
						}
					} else {
						fmt.Printf("[HOTKEY] wParam %v does not match HOTKEY_ID %v\n", msg.WParam, HOTKEY_ID)
					}
				}
				translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
				dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

// SimulateCopy simulates a Ctrl+C keypress independently
func SimulateCopy() error {
	// Define the input events for Ctrl+C
	inputs := []INPUT{
		// Press Ctrl
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVK: VK_CONTROL, DwFlags: 0}},
		// Press C
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVK: VK_C, DwFlags: 0}},
		// Release C
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVK: VK_C, DwFlags: 2}},
		// Release Ctrl
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVK: VK_CONTROL, DwFlags: 2}},
	}

	ret, _, err := sendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(INPUT{}),
	)

	if ret == 0 {
		return fmt.Errorf("SendInput failed: %v", err)
	}

	// Wait a bit for the copy operation to complete
	time.Sleep(100 * time.Millisecond)

	return nil
}

// SimulatePaste simulates a Ctrl+V keypress
func SimulatePaste() error {
	// Define the input events for Ctrl+V
	inputs := []INPUT{
		// Press Ctrl
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVK: VK_CONTROL, DwFlags: 0}},
		// Press V
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVK: VK_V, DwFlags: 0}},
		// Release V
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVK: VK_V, DwFlags: 2}},
		// Release Ctrl
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVK: VK_CONTROL, DwFlags: 2}},
	}

	ret, _, err := sendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(INPUT{}),
	)

	if ret == 0 {
		return fmt.Errorf("SendInput failed: %v", err)
	}

	// Wait a bit for the paste operation to complete
	time.Sleep(100 * time.Millisecond)

	return nil
}

// parseHotkey parses a hotkey string like "ctrl+shift+r"
func parseHotkey(combo string) (mods uint32, vk uint32, err error) {
	// Robust parsing: lowercase and trim
	cleanCombo := strings.ToLower(strings.TrimSpace(combo))
	parts := splitHotkey(cleanCombo)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		switch part {
		case "ctrl", "control":
			mods |= MOD_CONTROL
		case "alt":
			mods |= MOD_ALT
		case "shift":
			mods |= MOD_SHIFT
		case "win", "super", "cmd":
			mods |= MOD_WIN
		default:
			vk = keyToVK(part)
			if vk == 0 {
				return 0, 0, fmt.Errorf("unknown key: %s", part)
			}
		}
	}

	if vk == 0 {
		return 0, 0, fmt.Errorf("no main key specified in hotkey: %s", combo)
	}

	return mods, vk, nil
}

func splitHotkey(combo string) []string {
	// Handle both + and - as separators
	replaced := strings.ReplaceAll(combo, "-", "+")
	return strings.Split(replaced, "+")
}

func keyToVK(key string) uint32 {
	key = strings.TrimSpace(key)
	keyMap := map[string]uint32{
		"a": 0x41, "b": 0x42, "c": 0x43, "d": 0x44, "e": 0x45, "f": 0x46,
		"g": 0x47, "h": 0x48, "i": 0x49, "j": 0x4A, "k": 0x4B, "l": 0x4C,
		"m": 0x4D, "n": 0x4E, "o": 0x4F, "p": 0x50, "q": 0x51, "r": 0x52,
		"s": 0x53, "t": 0x54, "u": 0x55, "v": 0x56, "w": 0x57, "x": 0x58,
		"y": 0x59, "z": 0x5A,
		"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33, "4": 0x34,
		"5": 0x35, "6": 0x36, "7": 0x37, "8": 0x38, "9": 0x39,
		"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73, "f5": 0x74,
		"f6": 0x75, "f7": 0x76, "f8": 0x77, "f9": 0x78, "f10": 0x79,
		"f11": 0x7A, "f12": 0x7B,
		"space": 0x20, "enter": 0x0D, "esc": 0x1B, "tab": 0x09,
		"tick": 0xC0, "`": 0xC0, "grave": 0xC0, "backtick": 0xC0,
	}

	if vk, ok := keyMap[key]; ok {
		return vk
	}
	return 0
}
