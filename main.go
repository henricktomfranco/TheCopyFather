package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	"textrewriter/internal/config"
	"textrewriter/internal/ollama"
	"textrewriter/internal/rewriter"
	win "textrewriter/internal/windows"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// Constants for magic numbers
const (
	// MinClipboardTextLength is the minimum length of clipboard text to trigger processing
	// This prevents triggering on small selections like single words or characters
	MinClipboardTextLength = 10

	// ClipboardReadDelay is the delay after simulating copy to allow clipboard update
	ClipboardReadDelay = 150 * time.Millisecond

	// ClipboardRetryDelay is the additional delay when retrying clipboard read
	ClipboardRetryDelay = 100 * time.Millisecond

	// WindowHideDelay is the delay to ensure focus returns to original app
	WindowHideDelay = 200 * time.Millisecond

	// ClipboardSetDelay is the delay after setting clipboard before paste
	ClipboardSetDelay = 150 * time.Millisecond

	// TextPreviewLength is the length of text preview for logging
	TextPreviewLength = 50
)

// App struct
type App struct {
	ctx              context.Context
	config           *config.Config
	ollamaClient     *ollama.Client
	rewriter         *rewriter.Rewriter
	hotkeyManager    *win.HotkeyManager
	trayManager      *win.TrayManager
	clipboardManager *win.ClipboardManager
	quitting         bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize config
	a.config = config.Load()

	// Initialize Ollama client
	a.ollamaClient = ollama.NewClient(a.config.ServerURL, a.config.Model, a.config.APIKey)

	// Initialize rewriter
	a.rewriter = rewriter.New(a.ollamaClient, a.config)

	// Initialize Windows components
	a.initWindowsComponents()
}

func (a *App) initWindowsComponents() {
	// Initialize clipboard manager
	a.clipboardManager = win.NewClipboardManager()

	// Initialize hotkey manager
	a.hotkeyManager = win.NewHotkeyManager()
	err := a.hotkeyManager.Register(a.config.Hotkey, func() {
		a.onHotkeyTriggered()
	})
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to register hotkey: %v", err))
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Hotkey Error",
			Message: fmt.Sprintf("Failed to register hotkey '%s'. It might be in use by another app.\nError: %v", a.config.Hotkey, err),
		})
	} else {
		runtime.LogInfo(a.ctx, fmt.Sprintf("Successfully registered hotkey: %s", a.config.Hotkey))
	}

	// Initialize system tray
	a.trayManager = win.NewTrayManager()
	a.trayManager.OnShowSettings(func() {
		// Emit event to show settings
		runtime.EventsEmit(a.ctx, "window:showsettings")
	})
	a.trayManager.OnExit(func() {
		a.Quit()
	})
	a.trayManager.Start()

	// Start clipboard monitoring if enabled
	if a.config.MonitorClipboard {
		a.clipboardManager.Start(func(text string) {
			if len(text) > MinClipboardTextLength { // Only trigger for substantial text
				a.onTextSelected(text)
			}
		})
	}
}

func (a *App) onHotkeyTriggered() {
	runtime.LogInfo(a.ctx, "Hotkey triggered!")

	// Save current clipboard content
	oldText, err := a.clipboardManager.GetText()
	if err != nil {
		oldText = ""
	}

	// Simulate Ctrl+C to copy selected text
	if err := win.SimulateCopy(); err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("SimulateCopy failed: %v", err))
		// If simulation fails, just try reading clipboard directly
		text, err := a.clipboardManager.GetText()
		if err == nil && text != "" {
			a.onTextSelected(text)
		}
		return
	}

	// Read the copied text from clipboard
	// Increased delay to ensure clipboard is updated even on slower apps
	time.Sleep(ClipboardReadDelay)
	text, err := a.clipboardManager.GetText()
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to get clipboard text: %v", err))
		// Restore original clipboard
		a.clipboardManager.SetText(oldText)
		return
	}

	// Note: We intentionally do NOT restore the old clipboard here.
	// The user selected text and pressed the hotkey to copy it, so the
	// newly copied text should remain in the clipboard for their use.
	// If we restored oldText, it would undo their copy operation.

	if text != "" {
		runtime.LogInfo(a.ctx, fmt.Sprintf("Hotkey captured text length: %d", len(text)))
		a.onTextSelected(text)
	} else {
		// If empty, maybe the user didn't have text selected?
		// Try one more time with a slightly longer delay
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
	// Position window near cursor if enabled
	if a.config.PopupPositionMode == "cursor" {
		x, y, err := win.GetCursorPosition()
		if err == nil {
			// Offset slightly so cursor doesn't block content
			windowX := x + 20
			windowY := y - 100

			// Ensure window stays on screen (basic bounds check)
			if windowX < 0 {
				windowX = 10
			}
			if windowY < 0 {
				windowY = 10
			}

			runtime.WindowSetPosition(a.ctx, int(windowX), int(windowY))
		}
	}

	// Ensure the window is visible and active
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)

	// Trigger the popup with the selected text
	runtime.EventsEmit(a.ctx, "text:selected", text)
}

// RetryRewrite generates a new rewrite for a specific style
func (a *App) RetryRewrite(text, style string) rewriter.RewriteOption {
	option, _ := a.rewriter.GenerateSingleRewrite(a.ctx, text, style)
	return option
}

// ApplyRewrite applies the rewritten text by copying it to clipboard
func (a *App) ApplyRewrite(text string) error {
	return a.clipboardManager.SetRichText(text, text)
}

// ApplyRewriteAndPaste applies the rewritten text by copying it to clipboard and pasting it
func (a *App) ApplyRewriteAndPaste(text string) error {
	// First hide the window to return focus to original app
	runtime.WindowHide(a.ctx)

	// Longer delay to ensure focus returns to original application
	time.Sleep(WindowHideDelay)

	// Set clipboard text (with rich formatting)
	if err := a.clipboardManager.SetRichText(text, text); err != nil {
		return err
	}

	// Additional delay after setting clipboard
	time.Sleep(ClipboardSetDelay)

	// Simulate paste operation
	if err := win.SimulatePaste(); err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to paste: %v", err))
		return err
	}

	// Log for debugging
	previewLen := TextPreviewLength
	if len(text) < TextPreviewLength {
		previewLen = len(text)
	}
	runtime.LogInfo(a.ctx, fmt.Sprintf("Pasted text: %s", text[:previewLen]))

	return nil
}

// GetCursorPosition returns the current cursor position
func (a *App) GetCursorPosition() (map[string]int32, error) {
	x, y, err := win.GetCursorPosition()
	if err != nil {
		return nil, err
	}
	return map[string]int32{"x": x, "y": y}, nil
}

// GetSettings returns the current settings
func (a *App) GetSettings() *config.Config {
	return a.config
}

// SaveSettings saves new settings
func (a *App) SaveSettings(newConfig *config.Config) error {
	// Update config
	a.config = newConfig

	// Save to disk
	if err := a.config.Save(); err != nil {
		return err
	}

	// Reinitialize Ollama client with new settings
	a.ollamaClient = ollama.NewClient(a.config.ServerURL, a.config.Model, a.config.APIKey)
	a.rewriter = rewriter.New(a.ollamaClient, a.config)

	// Update hotkey if changed
	if a.config.Hotkey != "" {
		a.hotkeyManager.Stop()
		a.hotkeyManager = win.NewHotkeyManager()
		if err := a.hotkeyManager.Register(a.config.Hotkey, func() {
			a.onHotkeyTriggered()
		}); err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("Failed to register hotkey after settings change: %v", err))
			runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Type:    runtime.ErrorDialog,
				Title:   "Hotkey Error",
				Message: fmt.Sprintf("Failed to register hotkey '%s': %v", a.config.Hotkey, err),
			})
		}
	}

	// Update clipboard monitor if changed
	if a.clipboardManager != nil {
		a.clipboardManager.Stop() // Always stop existing monitor
		if a.config.MonitorClipboard {
			// Restart if enabled
			a.clipboardManager = win.NewClipboardManager() // Re-initialize to be safe
			a.clipboardManager.Start(func(text string) {
				if len(text) > MinClipboardTextLength { // Only trigger for substantial text
					a.onTextSelected(text)
				}
			})
		}
	}

	// Update auto-start setting
	exePath, err := os.Executable()
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to get executable path: %v", err))
	} else {
		if err := win.SetAutoStart(a.config.AutoStart, exePath); err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("Failed to update auto-start setting: %v", err))
		}
	}

	return nil
}

// GetAvailableModels returns available Ollama models
func (a *App) GetAvailableModels() ([]string, error) {
	return a.ollamaClient.GetAvailableModels()
}

// TestConnection tests the Ollama connection with custom parameters
func (a *App) TestConnection(serverURL, model, apiKey string) (string, error) {
	client := ollama.NewClient(serverURL, model, apiKey)
	err := client.HealthCheck()
	if err != nil {
		return "", err
	}
	return client.GetVersion(), nil
}

// GetRewriteStyles returns available rewrite styles
func (a *App) GetRewriteStyles() []string {
	return rewriter.RewriteStyles
}

// GetAnalysisStyles returns available analysis styles
func (a *App) GetAnalysisStyles() []string {
	return rewriter.AnalysisStyles
}

// RetryAnalysis generates a new analysis for a specific style
func (a *App) RetryAnalysis(text, style string) rewriter.RewriteOption {
	option, _ := a.rewriter.GenerateSingleAnalysis(a.ctx, text, style)
	return option
}

// RetryRewriteWithFormatting generates a new rewrite with optional formatting
func (a *App) RetryRewriteWithFormatting(text, style string, enableFormatting bool) rewriter.RewriteOption {
	option, _ := a.rewriter.GenerateSingleRewriteWithFormatting(a.ctx, text, style, enableFormatting)
	return option
}

// RetryAnalysisWithFormatting generates a new analysis with optional formatting
func (a *App) RetryAnalysisWithFormatting(text, style string, enableFormatting bool) rewriter.RewriteOption {
	option, _ := a.rewriter.GenerateSingleAnalysisWithFormatting(a.ctx, text, style, enableFormatting)
	return option
}

// GetStyleInfo returns information about a specific style
func (a *App) GetStyleInfo(style string) (struct {
	Label       string
	Icon        string
	Description string
}, bool) {
	return rewriter.GetStyleInfo(style)
}

// DetectTextType analyzes text and returns the detected type
func (a *App) DetectTextType(text string) struct {
	Type       string  `json:"type"`
	Label      string  `json:"label"`
	Icon       string  `json:"icon"`
	Confidence float64 `json:"confidence"`
} {
	detectedType, confidence := rewriter.DetectTextType(text)
	info, _ := rewriter.GetTextTypeInfo(detectedType)

	return struct {
		Type       string  `json:"type"`
		Label      string  `json:"label"`
		Icon       string  `json:"icon"`
		Confidence float64 `json:"confidence"`
	}{
		Type:       string(detectedType),
		Label:      info.Label,
		Icon:       info.Icon,
		Confidence: confidence,
	}
}

// GetTextTypes returns all available text types
func (a *App) GetTextTypes() []struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
} {
	types := rewriter.AllTextTypes()
	result := make([]struct {
		Type        string `json:"type"`
		Label       string `json:"label"`
		Icon        string `json:"icon"`
		Description string `json:"description"`
	}, 0, len(types))

	for _, t := range types {
		info, _ := rewriter.GetTextTypeInfo(t)
		result = append(result, struct {
			Type        string `json:"type"`
			Label       string `json:"label"`
			Icon        string `json:"icon"`
			Description string `json:"description"`
		}{
			Type:        string(t),
			Label:       info.Label,
			Icon:        info.Icon,
			Description: info.Description,
		})
	}

	return result
}

// GetAllCustomPrompts returns all custom prompts from config
func (a *App) GetAllCustomPrompts() map[string]map[string]string {
	return a.config.GetAllCustomPrompts()
}

// SetCustomPrompt sets a custom prompt for a specific style and text type
// Returns error if validation fails
func (a *App) SetCustomPrompt(style, textType, prompt string) error {
	if err := a.config.SetCustomPrompt(style, textType, prompt); err != nil {
		return err
	}
	return a.config.Save()
}

// DeleteCustomPrompt removes a custom prompt for a specific style and text type
func (a *App) DeleteCustomPrompt(style, textType string) error {
	a.config.DeleteCustomPrompt(style, textType)
	return a.config.Save()
}

// ResetAllCustomPrompts removes all custom prompts
func (a *App) ResetAllCustomPrompts() error {
	a.config.CustomPrompts = make(map[string]map[string]string)
	return a.config.Save()
}

// GetDefaultPrompt returns the default prompt for a style and text type
func (a *App) GetDefaultPrompt(style, textType string) string {
	defaultConfig := config.DefaultConfig()
	return defaultConfig.GetPrompt(style, textType)
}

// RetryRewriteWithTextType generates a rewrite with specific text type
func (a *App) RetryRewriteWithTextType(text, style, textType string, enableFormatting bool) rewriter.RewriteOption {
	option, _ := a.rewriter.GenerateRewriteWithTextType(a.ctx, text, style, rewriter.TextType(textType), enableFormatting)
	return option
}

// RetryAnalysisWithTextType generates an analysis with specific text type
func (a *App) RetryAnalysisWithTextType(text, style, textType string, enableFormatting bool) rewriter.RewriteOption {
	option, _ := a.rewriter.GenerateAnalysisWithTextType(a.ctx, text, style, rewriter.TextType(textType), enableFormatting)
	return option
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Stop all managers
	if a.hotkeyManager != nil {
		a.hotkeyManager.Stop()
	}
	if a.clipboardManager != nil {
		a.clipboardManager.Stop()
	}
	if a.trayManager != nil {
		a.trayManager.Stop()
	}

	// Save config
	if a.config != nil {
		a.config.Save()
	}
}

// beforeClose is called when the application is about to quit
func (a *App) beforeClose(ctx context.Context) bool {
	if a.quitting {
		return false
	}
	runtime.WindowHide(ctx)
	return true
}

// Quit terminates the application normally
func (a *App) Quit() {
	a.quitting = true
	a.shutdown(a.ctx)
	runtime.Quit(a.ctx)
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "The Copyfather",
		Width:  500,
		Height: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 15, A: 0},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
