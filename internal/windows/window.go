package windows

import (
	"runtime"
	"unsafe"
)

var (
	procSetWindowPos = user32.NewProc("SetWindowPos")
	procGetCursorPos = user32.NewProc("GetCursorPos")
	procGetDC        = user32.NewProc("GetDC")
	procReleaseDC    = user32.NewProc("ReleaseDC")
	procGetDeviceCaps = gdi32.NewProc("GetDeviceCaps")
)

const (
	HWND_TOPMOST   = ^uintptr(0) // -1
	HWND_NOTOPMOST = ^uintptr(1) // -2
	SWP_NOSIZE     = 0x0001
	SWP_NOMOVE     = 0x0002
	SWP_SHOWWINDOW = 0x0040

	// DeviceCap constants for GetDeviceCaps
	HORZRES = 8  // Horizontal width in pixels
	VERTRES = 10 // Vertical height in pixels
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

// GetScreenDimensions returns the screen width and height in pixels
func GetScreenDimensions() (width, height int32, err error) {
	hdc, _, err := procGetDC.Call(0)
	if hdc == 0 {
		return 0, 0, err
	}
	defer procReleaseDC.Call(0, hdc)

	widthRet, _, _ := procGetDeviceCaps.Call(hdc, HORZRES)
	heightRet, _, _ := procGetDeviceCaps.Call(hdc, VERTRES)

	return int32(widthRet), int32(heightRet), nil
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