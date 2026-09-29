package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"textrewriter/internal/config"
	"textrewriter/internal/engine"
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

func (s *SettingsService) GetEngineStatus() engine.Status {
	if s.app.engine == nil {
		return engine.Status{Hardware: "CPU", ModelLoaded: "Qwen3-1.7B", ThinkingOff: true, StreamingOn: true}
	}
	return s.app.engine.GetStatus()
}

func (s *SettingsService) StartEmbeddedEngine() error {
	if s.app.engine == nil {
		return fmt.Errorf("engine not initialized")
	}
	return s.app.engine.Start()
}

func (s *SettingsService) StopEmbeddedEngine() error {
	if s.app.engine == nil {
		return nil
	}
	return s.app.engine.Stop()
}

func (s *SettingsService) OpenEngineFolder() error {
	path := engine.GetDefaultEngineDir()
	_ = os.MkdirAll(path, 0755)
	return exec.Command("explorer", path).Start()
}

func (s *SettingsService) OpenModelsFolder() error {
	path := engine.GetDefaultModelsDir()
	_ = os.MkdirAll(path, 0755)
	return exec.Command("explorer", path).Start()
}

func (s *SettingsService) CheckSetupStatus() engine.SetupStatus {
	return engine.GetDownloader().CheckSetupStatus()
}

func (s *SettingsService) StartAutoDownload() error {
	downloader := engine.GetDownloader()
	go func() {
		err := downloader.DownloadAll(context.Background(), func(status engine.SetupStatus) {
			runtime.EventsEmit(s.app.ctx, "setup:status", status)
		})
		if err != nil {
			runtime.EventsEmit(s.app.ctx, "setup:error", err.Error())
			return
		}
		if s.app.engine != nil {
			_ = s.app.engine.Start()
		}
		runtime.EventsEmit(s.app.ctx, "setup:completed", true)
	}()
	return nil
}

func (s *SettingsService) CancelAutoDownload() {
	engine.GetDownloader().Cancel()
}

func (s *SettingsService) TestConnection(serverURL, model, apiKey string, useOpenAICompatible bool) (string, error) {
	if !useOpenAICompatible && (serverURL == "" || serverURL == "embedded") {
		if s.app.engine == nil {
			return "", fmt.Errorf("embedded engine not initialized")
		}
		status := s.app.engine.GetStatus()
		if !status.BinaryFound {
			return "", fmt.Errorf("%s", status.Error)
		}
		if !status.ModelFound {
			return "", fmt.Errorf("%s", status.Error)
		}
		if !status.IsRunning {
			if err := s.app.engine.Start(); err != nil {
				return "", err
			}
			status = s.app.engine.GetStatus()
		}
		return fmt.Sprintf("Embedded llama.cpp (CPU) Ready - %s (Threads: %d, Context: %d)", status.ModelLoaded, status.Threads, status.ContextSize), nil
	}

	if useOpenAICompatible {
		if serverURL == "" {
			serverURL = "https://integrate.api.nvidia.com/v1"
		}
		if model == "" {
			model = "mistralai/mistral-7b-instruct"
		}
		client := ollama.NewOpenAICompatibleClientWithAllOptions(serverURL, model, apiKey, false, s.app.cfg.DisableThinking)
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
	if s.app.cfg.ProviderMode == "embedded" {
		return []string{"Qwen3-1.7B"}, nil
	}
	return s.app.ollamaClient.GetAvailableModels()
}

// ============================================================================
// REWRITE SERVICE
// ============================================================================

func (r *RewriteService) ensureEngineReady() error {
	if r.app.cfg.ProviderMode == "embedded" || (r.app.cfg.ProviderMode == "" && !r.app.cfg.UseOpenAICompatible) {
		if r.app.engine == nil {
			return fmt.Errorf("embedded AI engine is not initialized")
		}
		status := r.app.engine.GetStatus()
		if !status.BinaryFound {
			return fmt.Errorf("llama-server.exe is missing. Please download it from Settings or the Welcome screen")
		}
		if !status.ModelFound {
			return fmt.Errorf("AI model file (Qwen3-1.7B) is missing. Please download it from Settings or the Welcome screen")
		}
		if !status.IsRunning {
			if err := r.app.engine.Start(); err != nil {
				return fmt.Errorf("failed to start embedded AI engine: %w", err)
			}
		}
	}
	return nil
}

func (r *RewriteService) RetryRewrite(text, style string) rewriter.RewriteOption {
	if err := r.ensureEngineReady(); err != nil {
		return rewriter.RewriteOption{Error: err.Error()}
	}
	option, _ := r.app.rewriter.GenerateSingleRewrite(r.app.ctx, text, style)
	return option
}

func (r *RewriteService) RetryAnalysis(text, style string) rewriter.RewriteOption {
	if err := r.ensureEngineReady(); err != nil {
		return rewriter.RewriteOption{Error: err.Error()}
	}
	option, _ := r.app.rewriter.GenerateSingleAnalysis(r.app.ctx, text, style)
	return option
}

func (r *RewriteService) RetryRewriteWithFormatting(text, style string, enableFormatting bool) rewriter.RewriteOption {
	if err := r.ensureEngineReady(); err != nil {
		return rewriter.RewriteOption{Error: err.Error()}
	}
	option, _ := r.app.rewriter.GenerateSingleRewriteWithFormatting(r.app.ctx, text, style, enableFormatting)
	return option
}

func (r *RewriteService) RetryAnalysisWithFormatting(text, style string, enableFormatting bool) rewriter.RewriteOption {
	if err := r.ensureEngineReady(); err != nil {
		return rewriter.RewriteOption{Error: err.Error()}
	}
	option, _ := r.app.rewriter.GenerateSingleAnalysisWithFormatting(r.app.ctx, text, style, enableFormatting)
	return option
}

func (r *RewriteService) RetryRewriteWithTextType(text, style, textType string, enableFormatting bool) rewriter.RewriteOption {
	if err := r.ensureEngineReady(); err != nil {
		return rewriter.RewriteOption{Error: err.Error()}
	}
	option, _ := r.app.rewriter.GenerateRewriteWithTextType(r.app.ctx, text, style, rewriter.TextType(textType), enableFormatting)
	return option
}

func (r *RewriteService) RetryAnalysisWithTextType(text, style, textType string, enableFormatting bool) rewriter.RewriteOption {
	if err := r.ensureEngineReady(); err != nil {
		return rewriter.RewriteOption{Error: err.Error()}
	}
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

		if err := r.ensureEngineReady(); err != nil {
			runtime.EventsEmit(r.app.ctx, "stream:error:"+requestID, err.Error())
			return
		}

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

		if err := r.ensureEngineReady(); err != nil {
			runtime.EventsEmit(r.app.ctx, "stream:error:"+requestID, err.Error())
			return
		}

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

		if err := r.ensureEngineReady(); err != nil {
			runtime.EventsEmit(r.app.ctx, "stream:error:"+requestID, err.Error())
			return
		}

		streamChan, err := r.app.rewriter.GenerateStreamAnalysisWithTextType(streamCtx, text, style, rewriter.TextType(textType), enableFormatting)
		if err != nil {
			runtime.EventsEmit(r.app.ctx, "stream:error:"+requestID, err.Error())
			return
		}
		r.app.streamChunksWithRateLimit(requestID, streamChan)
	}()
}

func (r *RewriteService) RetryRewriteWithSliders(text string, formality, length int, textType string, enableFormatting bool) rewriter.RewriteOption {
	if err := r.ensureEngineReady(); err != nil {
		return rewriter.RewriteOption{Error: err.Error()}
	}
	option, _ := r.app.rewriter.GenerateRewriteWithSliders(r.app.ctx, text, formality, length, rewriter.TextType(textType), enableFormatting)
	return option
}

func (r *RewriteService) StreamRewriteWithSliders(requestID, text string, formality, length int, textType string, enableFormatting bool) {
	r.app.registerStream(requestID)
	go func() {
		defer r.app.unregisterStream(requestID)
		streamCtx, cancel := context.WithCancel(r.app.ctx)
		r.app.setStreamCancel(requestID, cancel)
		defer cancel()

		if err := r.ensureEngineReady(); err != nil {
			runtime.EventsEmit(r.app.ctx, "stream:error:"+requestID, err.Error())
			return
		}

		streamChan, err := r.app.rewriter.GenerateStreamWithSliders(streamCtx, text, formality, length, rewriter.TextType(textType), enableFormatting)
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
