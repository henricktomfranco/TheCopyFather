package main

import (
	"context"
	"fmt"

	"textrewriter/internal/config"
	"textrewriter/internal/ollama"
	"textrewriter/internal/rewriter"
	"textrewriter/internal/updater"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ============================================================================
// SERVICE STRUCTS — Each maps to a focused set of frontend-facing methods.
// Only these are exposed via Wails Bind. App's internal methods are unexported.
// ============================================================================

// SettingsService exposes configuration and AI provider management.
type SettingsService struct {
	app *App
}

// RewriteService exposes text rewriting, analysis, streaming, and diff operations.
type RewriteService struct {
	app *App
}

// UpdateService exposes auto-update operations.
type UpdateService struct {
	app *App
}

// ============================================================================
// SETTINGS SERVICE
// ============================================================================

func (s *SettingsService) GetSettings() *config.Config {
	return s.app.cfg
}

func (s *SettingsService) SaveSettings(newConfig *config.Config) error {
	return s.app.saveSettings(newConfig)
}

func (s *SettingsService) TestConnection(serverURL, model, apiKey string, useOpenAICompatible bool) (string, error) {
	if useOpenAICompatible {
		client := ollama.NewOpenAICompatibleClient(serverURL, model, apiKey)
		if err := client.HealthCheck(); err != nil {
			return "", err
		}
		return client.GetVersion(), nil
	}
	client := ollama.NewClient(serverURL, model, apiKey)
	if err := client.HealthCheck(); err != nil {
		return "", err
	}
	return client.GetVersion(), nil
}

func (s *SettingsService) GetAvailableModels() ([]string, error) {
	return s.app.ollamaClient.GetAvailableModels()
}

// ============================================================================
// REWRITE SERVICE
// ============================================================================

func (r *RewriteService) RetryRewrite(text, style string) rewriter.RewriteOption {
	option, _ := r.app.rewriter.GenerateSingleRewrite(r.app.ctx, text, style)
	return option
}

func (r *RewriteService) RetryAnalysis(text, style string) rewriter.RewriteOption {
	option, _ := r.app.rewriter.GenerateSingleAnalysis(r.app.ctx, text, style)
	return option
}

func (r *RewriteService) RetryRewriteWithFormatting(text, style string, enableFormatting bool) rewriter.RewriteOption {
	option, _ := r.app.rewriter.GenerateSingleRewriteWithFormatting(r.app.ctx, text, style, enableFormatting)
	return option
}

func (r *RewriteService) RetryAnalysisWithFormatting(text, style string, enableFormatting bool) rewriter.RewriteOption {
	option, _ := r.app.rewriter.GenerateSingleAnalysisWithFormatting(r.app.ctx, text, style, enableFormatting)
	return option
}

func (r *RewriteService) RetryRewriteWithTextType(text, style, textType string, enableFormatting bool) rewriter.RewriteOption {
	option, _ := r.app.rewriter.GenerateRewriteWithTextType(r.app.ctx, text, style, rewriter.TextType(textType), enableFormatting)
	return option
}

func (r *RewriteService) RetryAnalysisWithTextType(text, style, textType string, enableFormatting bool) rewriter.RewriteOption {
	option, _ := r.app.rewriter.GenerateAnalysisWithTextType(r.app.ctx, text, style, rewriter.TextType(textType), enableFormatting)
	return option
}

func (r *RewriteService) StreamRewriteWithFormatting(requestID, text, style string, enableFormatting bool) {
	r.app.registerStream(requestID)
	go func() {
		defer r.app.unregisterStream(requestID)
		streamCtx, cancel := context.WithCancel(r.app.ctx)
		r.app.setStreamCancel(requestID, cancel)
		defer cancel()

		streamChan, err := r.app.rewriter.GenerateStreamWithFormatting(streamCtx, text, style, enableFormatting)
		if err != nil {
			runtime.EventsEmit(r.app.ctx, "stream:error:"+requestID, err.Error())
			return
		}
		r.app.streamChunksWithRateLimit(requestID, streamChan)
	}()
}

func (r *RewriteService) StreamRewriteWithTextType(requestID, text, style, textType string, enableFormatting bool) {
	r.app.registerStream(requestID)
	go func() {
		defer r.app.unregisterStream(requestID)
		streamCtx, cancel := context.WithCancel(r.app.ctx)
		r.app.setStreamCancel(requestID, cancel)
		defer cancel()

		streamChan, err := r.app.rewriter.GenerateStreamWithTextType(streamCtx, text, style, rewriter.TextType(textType), enableFormatting)
		if err != nil {
			runtime.EventsEmit(r.app.ctx, "stream:error:"+requestID, err.Error())
			return
		}
		r.app.streamChunksWithRateLimit(requestID, streamChan)
	}()
}

func (r *RewriteService) StreamAnalysisWithTextType(requestID, text, style, textType string, enableFormatting bool) {
	r.app.registerStream(requestID)
	go func() {
		defer r.app.unregisterStream(requestID)
		streamCtx, cancel := context.WithCancel(r.app.ctx)
		r.app.setStreamCancel(requestID, cancel)
		defer cancel()

		streamChan, err := r.app.rewriter.GenerateStreamAnalysisWithTextType(streamCtx, text, style, rewriter.TextType(textType), enableFormatting)
		if err != nil {
			runtime.EventsEmit(r.app.ctx, "stream:error:"+requestID, err.Error())
			return
		}
		r.app.streamChunksWithRateLimit(requestID, streamChan)
	}()
}

func (r *RewriteService) CancelStream(requestID string) {
	r.app.cancelStream(requestID)
}

func (r *RewriteService) GetRewriteStyles() []string {
	return rewriter.RewriteStyles
}

func (r *RewriteService) GetAnalysisStyles() []string {
	return rewriter.AnalysisStyles
}

func (r *RewriteService) GetStyleInfo(style string) (rewriter.StyleInfoData, bool) {
	return rewriter.GetStyleInfo(style)
}

func (r *RewriteService) DetectTextType(text string) rewriter.TextTypeDetected {
	detectedType, confidence := rewriter.DetectTextType(text)
	info, _ := rewriter.GetTextTypeInfo(detectedType)
	return rewriter.TextTypeDetected{
		Type:       string(detectedType),
		Label:      info.Label,
		Icon:       info.Icon,
		Confidence: confidence,
	}
}

func (r *RewriteService) GetTextTypes() []rewriter.TextTypeInfo {
	types := rewriter.AllTextTypes()
	result := make([]rewriter.TextTypeInfo, 0, len(types))
	for _, t := range types {
		info, _ := rewriter.GetTextTypeInfo(t)
		result = append(result, rewriter.TextTypeInfo{
			Type:        string(t),
			Label:       info.Label,
			Icon:        info.Icon,
			Description: info.Description,
		})
	}
	return result
}

func (r *RewriteService) GetAllCustomPrompts() map[string]map[string]string {
	return r.app.cfg.GetAllCustomPrompts()
}

func (r *RewriteService) SetCustomPrompt(style, textType, prompt string) error {
	if err := r.app.cfg.SetCustomPrompt(style, textType, prompt); err != nil {
		return err
	}
	return r.app.cfg.Save()
}

func (r *RewriteService) DeleteCustomPrompt(style, textType string) error {
	r.app.cfg.DeleteCustomPrompt(style, textType)
	return r.app.cfg.Save()
}

func (r *RewriteService) ResetAllCustomPrompts() error {
	r.app.cfg.CustomPrompts = make(map[string]map[string]string)
	return r.app.cfg.Save()
}

func (r *RewriteService) GetDefaultPrompt(style, textType string) string {
	defaultConfig := config.DefaultConfig()
	return defaultConfig.GetPrompt(style, textType)
}

func (r *RewriteService) ComputeDiff(original, rewritten string) rewriter.DiffResult {
	return r.app.rewriter.ComputeDiff(original, rewritten)
}

// ============================================================================
// UPDATE SERVICE
// ============================================================================

func (u *UpdateService) DownloadAndInstallUpdate() error {
	if u.app.cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	updateInfo := updater.CheckForUpdates(u.app.cfg.CurrentVersion)
	if updateInfo.Error != "" {
		return fmt.Errorf("update check failed: %s", updateInfo.Error)
	}
	if !updateInfo.Available {
		return fmt.Errorf("no update available")
	}

	runtime.LogInfo(u.app.ctx, fmt.Sprintf("Downloading update from: %s", updateInfo.DownloadURL))
	runtime.EventsEmit(u.app.ctx, "update:downloadstart")

	tempFile, err := updater.DownloadUpdate(updateInfo.DownloadURL)
	if err != nil {
		runtime.EventsEmit(u.app.ctx, "update:downloadfailed", map[string]string{"error": err.Error()})
		return fmt.Errorf("failed to download update: %v", err)
	}

	runtime.LogInfo(u.app.ctx, fmt.Sprintf("Update downloaded to: %s", tempFile))

	_, exePath, err := updater.GetUpdatePaths()
	if err != nil {
		runtime.EventsEmit(u.app.ctx, "update:installfailed", map[string]string{"error": err.Error()})
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	if err := updater.InstallUpdate(tempFile, exePath); err != nil {
		runtime.EventsEmit(u.app.ctx, "update:installfailed", map[string]string{"error": err.Error()})
		return fmt.Errorf("failed to install update: %v", err)
	}

	runtime.LogInfo(u.app.ctx, "Update installed successfully")

	u.app.cfg.CurrentVersion = updateInfo.LatestVersion
	if err := u.app.cfg.Save(); err != nil {
		runtime.LogError(u.app.ctx, fmt.Sprintf("Failed to save updated version: %v", err))
	}

	runtime.EventsEmit(u.app.ctx, "update:success", map[string]string{
		"version": u.app.cfg.CurrentVersion,
	})

	if err := updater.RestartApp(exePath); err != nil {
		runtime.LogError(u.app.ctx, fmt.Sprintf("Failed to restart: %v", err))
		return err
	}

	return nil
}

func (u *UpdateService) SkipUpdate() {
	runtime.EventsEmit(u.app.ctx, "update:skipped")
}
