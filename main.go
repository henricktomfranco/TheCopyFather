package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"sync"
	"time"

	"textrewriter/internal/config"
	"textrewriter/internal/ollama"
	"textrewriter/internal/rewriter"
	"textrewriter/internal/updater"
	win "textrewriter/internal/windows"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsWindows "github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// Version is set at build time via -ldflags "-X main.Version=x.y.z"
var Version = "dev"

// Constants for magic numbers
const (
	MinClipboardTextLength = 10
	ClipboardReadDelay     = 150 * time.Millisecond
	ClipboardRetryDelay    = 100 * time.Millisecond
	WindowHideDelay        = 200 * time.Millisecond
	ClipboardSetDelay      = 150 * time.Millisecond
	TextPreviewLength      = 50
)

// App struct — internal state only. Public frontend-facing methods live on
// SettingsService, RewriteService, and UpdateService (see services.go).
// Only clipboard helpers and Quit remain on App itself.
type App struct {
	ctx               context.Context
	cfg               *config.Config
	ollamaClient      ollama.AIClient
	rewriter          *rewriter.Rewriter
	hotkeyManager     *win.HotkeyManager
	trayManager       *win.TrayManager
	clipboardManager  *win.ClipboardManager
	quitting          bool
	streamingRequests map[string]context.CancelFunc
	streamingMu       sync.RWMutex
}

func NewApp() *App {
	return &App{}
}

// ============================================================================
// LIFECYCLE
// ============================================================================

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.cfg = config.Load()

	// Set version from build if not yet saved (first run after build)
	if a.cfg.CurrentVersion == "" {
		a.cfg.CurrentVersion = Version
	}

	if a.cfg.AutoUpdateEnabled {
		go a.checkForUpdates()
	}

	a.createClients()
	a.initWindowsComponents()
}

func (a *App) domReady(ctx context.Context) {
	// Restore saved window position and size
	if a.cfg.WindowWidth > 0 && a.cfg.WindowHeight > 0 {
		runtime.WindowSetSize(ctx, a.cfg.WindowWidth, a.cfg.WindowHeight)
	}
	if a.cfg.WindowX != 0 || a.cfg.WindowY != 0 {
		runtime.WindowSetPosition(ctx, a.cfg.WindowX, a.cfg.WindowY)
	}
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.quitting {
		return false
	}
	runtime.WindowHide(ctx)
	return true
}

func (a *App) shutdown(ctx context.Context) {
	a.streamingMu.RLock()
	for _, cancel := range a.streamingRequests {
		cancel()
	}
	a.streamingMu.RUnlock()

	// Persist window position and size
	if w, h := runtime.WindowGetSize(ctx); w > 0 && h > 0 {
		a.cfg.WindowWidth = w
		a.cfg.WindowHeight = h
	}
	x, y := runtime.WindowGetPosition(ctx)
	a.cfg.WindowX = x
	a.cfg.WindowY = y

	if a.hotkeyManager != nil {
		a.hotkeyManager.Stop()
	}
	if a.clipboardManager != nil {
		a.clipboardManager.Stop()
	}
	if a.trayManager != nil {
		a.trayManager.Stop()
	}
	if a.cfg != nil {
		a.cfg.Save()
	}
}

// ============================================================================
// PUBLIC CLIPBOARD / WINDOW HELPERS
// ============================================================================

func (a *App) ApplyRewrite(text string) error {
	return a.clipboardManager.SetRichText(text, text)
}

func (a *App) ApplyRewriteAndPaste(text string) error {
	runtime.WindowHide(a.ctx)
	time.Sleep(WindowHideDelay)

	if err := a.clipboardManager.SetRichText(text, text); err != nil {
		return err
	}
	time.Sleep(ClipboardSetDelay)

	if err := win.SimulatePaste(); err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to paste: %v", err))
		return err
	}

	previewLen := TextPreviewLength
	if len(text) < TextPreviewLength {
		previewLen = len(text)
	}
	runtime.LogInfo(a.ctx, fmt.Sprintf("Pasted text: %s", text[:previewLen]))
	return nil
}

func (a *App) GetCursorPosition() (map[string]int32, error) {
	x, y, err := win.GetCursorPosition()
	if err != nil {
		return nil, err
	}
	return map[string]int32{"x": x, "y": y}, nil
}

func (a *App) Quit() {
	a.quitting = true
	a.shutdown(a.ctx)
	runtime.Quit(a.ctx)
}

// ============================================================================
// INTERNAL — SETTINGS (called by SettingsService)
// ============================================================================

func (a *App) createClients() {
	if a.cfg.UseOpenAICompatible {
		client := ollama.NewOpenAICompatibleClientWithOptions(a.cfg.OpenAIBaseURL, a.cfg.OpenAIModel, a.cfg.OpenAIAPIKey, a.cfg.DisableStreaming)
		a.ollamaClient = client
		a.rewriter = rewriter.New(client, a.cfg)
	} else {
		client := ollama.NewClientWithOptions(a.cfg.ServerURL, a.cfg.Model, a.cfg.APIKey, a.cfg.DisableStreaming)
		a.ollamaClient = client
		a.rewriter = rewriter.New(client, a.cfg)
	}
}

func (a *App) saveSettings(newConfig *config.Config) error {
	a.cfg = newConfig
	if err := a.cfg.Save(); err != nil {
		return err
	}

	a.createClients()

	if a.cfg.Hotkey != "" && a.cfg.Hotkey != a.hotkeyManager.CurrentHotkey() {
		a.hotkeyManager.Stop()
		a.hotkeyManager = win.NewHotkeyManager()
		if err := a.hotkeyManager.Register(a.cfg.Hotkey, func() {
			a.onHotkeyTriggered()
		}); err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("Failed to register hotkey after settings change: %v", err))
			runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Type:    runtime.ErrorDialog,
				Title:   "Hotkey Error",
				Message: fmt.Sprintf("Failed to register hotkey '%s': %v", a.cfg.Hotkey, err),
			})
		}
	}

	if a.clipboardManager != nil {
		a.clipboardManager.Stop()
		if a.cfg.MonitorClipboard {
			a.clipboardManager = win.NewClipboardManager()
			a.clipboardManager.Start(func(text string) {
				if len(text) > MinClipboardTextLength {
					a.onTextSelected(text)
				}
			})
		}
	}

	exePath, err := os.Executable()
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to get executable path: %v", err))
	} else {
		if err := win.SetAutoStart(a.cfg.AutoStart, exePath); err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("Failed to update auto-start setting: %v", err))
		}
	}

	return nil
}

// ============================================================================
// INTERNAL — WINDOWS / HOTKEY / CLIPBOARD
// ============================================================================

func (a *App) initWindowsComponents() {
	a.clipboardManager = win.NewClipboardManager()

	a.hotkeyManager = win.NewHotkeyManager()
	err := a.hotkeyManager.Register(a.cfg.Hotkey, func() {
		a.onHotkeyTriggered()
	})
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to register hotkey: %v", err))
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Hotkey Error",
			Message: fmt.Sprintf("Failed to register hotkey '%s'. It might be in use by another app.\nError: %v", a.cfg.Hotkey, err),
		})
	} else {
		runtime.LogInfo(a.ctx, fmt.Sprintf("Successfully registered hotkey: %s", a.cfg.Hotkey))
	}

	a.trayManager = win.NewTrayManager()
	a.trayManager.OnShowSettings(func() {
		runtime.WindowUnminimise(a.ctx)
		runtime.WindowShow(a.ctx)
		runtime.EventsEmit(a.ctx, "window:showsettings")
	})
	a.trayManager.OnExit(func() {
		a.Quit()
	})
	a.trayManager.Start()

	if a.cfg.MonitorClipboard {
		a.clipboardManager.Start(func(text string) {
			if len(text) > MinClipboardTextLength {
				a.onTextSelected(text)
			}
		})
	}
}

func (a *App) onHotkeyTriggered() {
	runtime.LogInfo(a.ctx, "Hotkey triggered!")

	oldText, err := a.clipboardManager.GetText()
	if err != nil {
		oldText = ""
	}

	if err := win.SimulateCopy(); err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("SimulateCopy failed: %v", err))
		text, err := a.clipboardManager.GetText()
		if err == nil && text != "" {
			a.onTextSelected(text)
		}
		return
	}

	time.Sleep(ClipboardReadDelay)
	text, err := a.clipboardManager.GetText()
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to get clipboard text: %v", err))
		a.clipboardManager.SetText(oldText)
		return
	}

	if text != "" {
		runtime.LogInfo(a.ctx, fmt.Sprintf("Hotkey captured text length: %d", len(text)))
		a.onTextSelected(text)
	} else {
		time.Sleep(ClipboardRetryDelay)
		text, err = a.clipboardManager.GetText()
		if err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("Failed to read clipboard on retry: %v", err))
			return
		}
		if text != "" {
			a.onTextSelected(text)
			if err := a.clipboardManager.SetText(oldText); err != nil {
				runtime.LogError(a.ctx, fmt.Sprintf("Failed to restore clipboard: %v", err))
			}
		} else {
			runtime.LogWarning(a.ctx, "Hotkey triggered but no text was captured")
		}
	}
}

func (a *App) onTextSelected(text string) {
	if a.cfg.PopupPositionMode == "cursor" {
		x, y, err := win.GetCursorPosition()
		if err == nil {
			windowX := x + 20
			windowY := y - 100
			if windowX < 0 {
				windowX = 10
			}
			if windowY < 0 {
				windowY = 10
			}
			runtime.WindowSetPosition(a.ctx, int(windowX), int(windowY))
		}
	}

	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.EventsEmit(a.ctx, "text:selected", text)
}

// ============================================================================
// INTERNAL — STREAMING
// ============================================================================

func (a *App) registerStream(requestID string) {
	a.streamingMu.Lock()
	defer a.streamingMu.Unlock()
	if a.streamingRequests == nil {
		a.streamingRequests = make(map[string]context.CancelFunc)
	}
}

func (a *App) setStreamCancel(requestID string, cancel context.CancelFunc) {
	a.streamingMu.Lock()
	defer a.streamingMu.Unlock()
	a.streamingRequests[requestID] = cancel
}

func (a *App) unregisterStream(requestID string) {
	a.streamingMu.Lock()
	defer a.streamingMu.Unlock()
	delete(a.streamingRequests, requestID)
}

func (a *App) cancelStream(requestID string) {
	a.streamingMu.RLock()
	cancel, exists := a.streamingRequests[requestID]
	a.streamingMu.RUnlock()
	if exists {
		cancel()
	}
}

func (a *App) streamChunksWithRateLimit(requestID string, streamChan <-chan rewriter.StreamChunk) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var lastText string
	var pending bool
	var errMsg string

	emitPending := func(forceDone bool) {
		if pending {
			pending = false
			if errMsg != "" {
				runtime.EventsEmit(a.ctx, "stream:error:"+requestID, errMsg)
			} else {
				runtime.EventsEmit(a.ctx, "stream:chunk:"+requestID, lastText)
			}
		}
		if forceDone {
			runtime.EventsEmit(a.ctx, "stream:done:"+requestID, true)
		}
	}

	for {
		select {
		case chunk, ok := <-streamChan:
			if !ok {
				emitPending(true)
				return
			}
			if chunk.Error != "" {
				errMsg = chunk.Error
				pending = true
				emitPending(false)
				return
			}
			if chunk.Done {
				// Flush any pending text BEFORE sending done signal
				if chunk.Text != "" {
					lastText = chunk.Text
				}
				emitPending(true)
				return
			}
			lastText = chunk.Text
			pending = true
		case <-ticker.C:
			emitPending(false)
		}
	}
}

// ============================================================================
// INTERNAL — UPDATES
// ============================================================================

func (a *App) checkForUpdates() {
	if a.cfg == nil {
		runtime.LogError(a.ctx, "Config not loaded, skipping update check")
		return
	}

	runtime.LogInfo(a.ctx, fmt.Sprintf("Checking for updates... Current version: %s", a.cfg.CurrentVersion))

	updateInfo := updater.CheckForUpdates(a.cfg.CurrentVersion)
	if updateInfo.Error != "" {
		runtime.LogError(a.ctx, fmt.Sprintf("Update check failed: %s", updateInfo.Error))
		return
	}
	if !updateInfo.Available {
		runtime.LogInfo(a.ctx, "No updates available")
		return
	}

	runtime.LogInfo(a.ctx, fmt.Sprintf("Update available: %s (current: %s)", updateInfo.LatestVersion, updateInfo.CurrentVersion))
	runtime.EventsEmit(a.ctx, "update:available", map[string]string{
		"currentVersion": updateInfo.CurrentVersion,
		"latestVersion":  updateInfo.LatestVersion,
	})
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "The Copyfather",
		Width:  500,
		Height: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 26, G: 26, B: 46, A: 255},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app, // ApplyRewrite, ApplyRewriteAndPaste, GetCursorPosition, Quit
			&SettingsService{app: app},
			&RewriteService{app: app},
			&UpdateService{app: app},
		},
		Windows: &wailsWindows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
