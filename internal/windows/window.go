package windows

import (
	"runtime"
	"unsafe"
)

var (
	procSetWindowPos = user32.NewProc("SetWindowPos")
	procGetCursorPos = user32.NewProc("GetCursorPos")
)

const (
	HWND_TOPMOST   = ^uintptr(0) // -1
	HWND_NOTOPMOST = ^uintptr(1) // -2
	SWP_NOSIZE     = 0x0001
	SWP_NOMOVE     = 0x0002
	SWP_SHOWWINDOW = 0x0040
)

// GetCursorPosition returns the current cursor position
func GetCursorPosition() (x, y int32, err error) {
	var pt struct {
		X, Y int32
	}

	ret, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return 0, 0, err
	}

	return pt.X, pt.Y, nil
}

// SetWindowTopMost makes a window stay on top
func SetWindowTopMost(hwnd uintptr) error {
	ret, _, err := procSetWindowPos.Call(
		hwnd,
		HWND_TOPMOST,
		0, 0, 0, 0,
		SWP_NOMOVE|SWP_NOSIZE|SWP_SHOWWINDOW,
	)

	if ret == 0 {
		return err
	}

	return nil
}

// EnableMainThread ensures code runs on the main thread
func EnableMainThread() {
	runtime.LockOSThread()
}
