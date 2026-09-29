package config

import (
	"os"
	"strings"
	"testing"
)

// TestDefaultConfig tests that DefaultConfig returns a valid configuration
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check required fields
	if cfg.ServerURL == "" {
		t.Error("DefaultConfig: ServerURL should not be empty")
	}
	if cfg.Model == "" {
		t.Error("DefaultConfig: Model should not be empty")
	}
	if cfg.DefaultStyle == "" {
		t.Error("DefaultConfig: DefaultStyle should not be empty")
	}

	// Check default values
	if cfg.ServerURL != "http://localhost:11434" {
		t.Errorf("DefaultConfig: Expected ServerURL to be 'http://localhost:11434', got '%s'", cfg.ServerURL)
	}
	if cfg.Model != "SmolLM2-360M" {
		t.Errorf("DefaultConfig: Expected Model to be 'SmolLM2-360M', got '%s'", cfg.Model)
	}
	if cfg.AutoStart != true {
		t.Error("DefaultConfig: Expected AutoStart to be true")
	}
	if cfg.MonitorClipboard != false {
		t.Error("DefaultConfig: Expected MonitorClipboard to be false")
	}

	// Check auto-update defaults
	if cfg.AutoUpdateEnabled != true {
		t.Error("DefaultConfig: Expected AutoUpdateEnabled to be true")
	}
	if cfg.CurrentVersion != "" {
		t.Errorf("DefaultConfig: Expected CurrentVersion to be empty (set at build time), got '%s'", cfg.CurrentVersion)
	}
	if cfg.UpdateChannel != "stable" {
		t.Errorf("DefaultConfig: Expected UpdateChannel to be 'stable', got '%s'", cfg.UpdateChannel)
	}

	// Check OpenAI-compatible defaults
	if cfg.UseOpenAICompatible != false {
		t.Error("DefaultConfig: Expected UseOpenAICompatible to be false")
	}
	if cfg.OpenAIBaseURL != "https://integrate.api.nvidia.com/v1" {
		t.Errorf("DefaultConfig: Expected OpenAIBaseURL to be 'https://integrate.api.nvidia.com/v1', got '%s'", cfg.OpenAIBaseURL)
	}
	if cfg.OpenAIModel != "mistralai/mistral-7b-instruct" {
		t.Errorf("DefaultConfig: Expected OpenAIModel to be 'mistralai/mistral-7b-instruct', got '%s'", cfg.OpenAIModel)
	}
}

// TestConfigSaveAndLoad tests saving and loading configuration
func TestConfigSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set the config directory to our temp dir
	originalHome := os.Getenv("HOME")
	originalAppData := os.Getenv("APPDATA")
	originalUserProfile := os.Getenv("USERPROFILE")

	os.Setenv("HOME", tempDir)
	os.Setenv("APPDATA", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("APPDATA", originalAppData)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

	// Create a test config
	cfg := &Config{
		ServerURL: "http://test-server:8080",
		Model: "test-model",
		APIKey: "test-api-key",
		DefaultStyle: "standard",
		AutoStart: false,
		Hotkey: "ctrl+shift+t",
		MonitorClipboard: true,
		FirstRun: false,
		CustomPrompts: map[string]map[string]string{
			"email": {
				"grammar": "Test grammar prompt",
			},
		},
		AutoPasteMode: "ask",
		PopupPositionMode: "cursor",
		MiniMode: true,
		AutoMinimizeOnCopy: false,
		AutoUpdateEnabled: true,
		CurrentVersion: "1.0.0",
		UpdateChannel: "beta",
		UseOpenAICompatible: true,
		OpenAIBaseURL: "https://test-api.example.com/v1",
		OpenAIModel: "test-model",
		OpenAIAPIKey: "test-openai-key",
	}

	// Save the config
	err = cfg.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the config
	loadedCfg := Load()

	// Verify all fields
	if loadedCfg.ServerURL != cfg.ServerURL {
		t.Errorf("Expected ServerURL '%s', got '%s'", cfg.ServerURL, loadedCfg.ServerURL)
	}
	if loadedCfg.Model != cfg.Model {
		t.Errorf("Expected Model '%s', got '%s'", cfg.Model, loadedCfg.Model)
	}
	if loadedCfg.APIKey != cfg.APIKey {
		t.Errorf("Expected APIKey '%s', got '%s'", cfg.APIKey, loadedCfg.APIKey)
	}
	if loadedCfg.DefaultStyle != cfg.DefaultStyle {
		t.Errorf("Expected DefaultStyle '%s', got '%s'", cfg.DefaultStyle, loadedCfg.DefaultStyle)
	}
	if loadedCfg.AutoStart != cfg.AutoStart {
		t.Errorf("Expected AutoStart %v, got %v", cfg.AutoStart, loadedCfg.AutoStart)
	}
	if loadedCfg.Hotkey != cfg.Hotkey {
		t.Errorf("Expected Hotkey '%s', got '%s'", cfg.Hotkey, loadedCfg.Hotkey)
	}
	if loadedCfg.MonitorClipboard != cfg.MonitorClipboard {
		t.Errorf("Expected MonitorClipboard %v, got %v", cfg.MonitorClipboard, loadedCfg.MonitorClipboard)
	}
	if loadedCfg.FirstRun != cfg.FirstRun {
		t.Errorf("Expected FirstRun %v, got %v", cfg.FirstRun, loadedCfg.FirstRun)
	}
	if loadedCfg.AutoPasteMode != cfg.AutoPasteMode {
		t.Errorf("Expected AutoPasteMode '%s', got '%s'", cfg.AutoPasteMode, loadedCfg.AutoPasteMode)
	}
	if loadedCfg.PopupPositionMode != cfg.PopupPositionMode {
		t.Errorf("Expected PopupPositionMode '%s', got '%s'", cfg.PopupPositionMode, loadedCfg.PopupPositionMode)
	}
	if loadedCfg.MiniMode != cfg.MiniMode {
		t.Errorf("Expected MiniMode %v, got %v", cfg.MiniMode, loadedCfg.MiniMode)
	}
	if loadedCfg.AutoMinimizeOnCopy != cfg.AutoMinimizeOnCopy {
		t.Errorf("Expected AutoMinimizeOnCopy %v, got %v", cfg.AutoMinimizeOnCopy, loadedCfg.AutoMinimizeOnCopy)
	}
	// Auto-update fields
	if loadedCfg.AutoUpdateEnabled != cfg.AutoUpdateEnabled {
		t.Errorf("Expected AutoUpdateEnabled %v, got %v", cfg.AutoUpdateEnabled, loadedCfg.AutoUpdateEnabled)
	}
	if loadedCfg.CurrentVersion != cfg.CurrentVersion {
		t.Errorf("Expected CurrentVersion '%s', got '%s'", cfg.CurrentVersion, loadedCfg.CurrentVersion)
	}
	if loadedCfg.UpdateChannel != cfg.UpdateChannel {
		t.Errorf("Expected UpdateChannel '%s', got '%s'", cfg.UpdateChannel, loadedCfg.UpdateChannel)
	}
	// OpenAI-compatible fields
	if loadedCfg.UseOpenAICompatible != cfg.UseOpenAICompatible {
		t.Errorf("Expected UseOpenAICompatible %v, got %v", cfg.UseOpenAICompatible, loadedCfg.UseOpenAICompatible)
	}
	if loadedCfg.OpenAIBaseURL != cfg.OpenAIBaseURL {
		t.Errorf("Expected OpenAIBaseURL '%s', got '%s'", cfg.OpenAIBaseURL, loadedCfg.OpenAIBaseURL)
	}
	if loadedCfg.OpenAIModel != cfg.OpenAIModel {
		t.Errorf("Expected OpenAIModel '%s', got '%s'", cfg.OpenAIModel, loadedCfg.OpenAIModel)
	}
	if loadedCfg.OpenAIAPIKey != cfg.OpenAIAPIKey {
		t.Errorf("Expected OpenAIAPIKey '%s', got '%s'", cfg.OpenAIAPIKey, loadedCfg.OpenAIAPIKey)
	}
}

// TestGetPrompt tests the GetPrompt method
func TestGetPrompt(t *testing.T) {
	cfg := DefaultConfig()

	// Test getting default prompts
	prompt := cfg.GetPrompt("email", "grammar")
	if prompt == "" {
		t.Error("Expected non-empty prompt for email/grammar")
	}

	// Test with custom prompts
	cfg.CustomPrompts = map[string]map[string]string{
		"email": {
			"grammar": "Custom grammar prompt",
		},
	}
	prompt = cfg.GetPrompt("email", "grammar")
	if !strings.HasPrefix(prompt, "Custom grammar prompt") {
		t.Errorf("Expected custom prompt prefix, got '%s'", prompt)
	}

	// Test fallback to default when custom prompt doesn't exist
	prompt = cfg.GetPrompt("email", "formal")
	if prompt == "" {
		t.Error("Expected default prompt for email/formal when custom doesn't exist")
	}
}

// TestSetCustomPrompt tests setting custom prompts
func TestSetCustomPrompt(t *testing.T) {
	cfg := DefaultConfig()

	// Test setting a valid custom prompt
	err := cfg.SetCustomPrompt("email", "grammar", "Test prompt")
	if err != nil {
		t.Errorf("Unexpected error setting custom prompt: %v", err)
	}

	// Verify the prompt was set
	if cfg.CustomPrompts["email"] == nil {
		t.Error("Expected CustomPrompts['email'] to exist")
	}
	if cfg.CustomPrompts["email"]["grammar"] != "Test prompt" {
		t.Errorf("Expected custom prompt 'Test prompt', got '%s'", cfg.CustomPrompts["email"]["grammar"])
	}

	// Test setting a prompt with invalid style
	err = cfg.SetCustomPrompt("", "grammar", "Test prompt")
	if err == nil {
		t.Error("Expected error for empty style")
	}

	// Test setting a prompt with invalid text type
	err = cfg.SetCustomPrompt("email", "", "Test prompt")
	if err == nil {
		t.Error("Expected error for empty text type")
	}
}

// TestDeleteCustomPrompt tests deleting custom prompts
func TestDeleteCustomPrompt(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CustomPrompts = map[string]map[string]string{
		"email": {
			"grammar": "Test prompt",
			"formal":  "Formal prompt",
		},
	}

	// Delete a custom prompt
	cfg.DeleteCustomPrompt("email", "grammar")

	// Verify it was deleted
	if cfg.CustomPrompts["email"] == nil {
		t.Error("Expected CustomPrompts['email'] to still exist")
	}
	if _, exists := cfg.CustomPrompts["email"]["grammar"]; exists {
		t.Error("Expected 'grammar' prompt to be deleted")
	}
	if cfg.CustomPrompts["email"]["formal"] != "Formal prompt" {
		t.Error("Expected 'formal' prompt to still exist")
	}

	// Delete the last prompt in a style
	cfg.DeleteCustomPrompt("email", "formal")
	if cfg.CustomPrompts["email"] != nil {
		t.Error("Expected CustomPrompts['email'] to be deleted when empty")
	}
}

// TestGetAllCustomPrompts tests getting all custom prompts
func TestGetAllCustomPrompts(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CustomPrompts = map[string]map[string]string{
		"email": {
			"grammar": "Test prompt",
		},
		"chat": {
			"casual": "Casual prompt",
		},
	}

	prompts := cfg.GetAllCustomPrompts()

	if len(prompts) != 2 {
		t.Errorf("Expected 2 styles, got %d", len(prompts))
	}
	if prompts["email"]["grammar"] != "Test prompt" {
		t.Error("Expected email/grammar prompt to match")
	}
	if prompts["chat"]["casual"] != "Casual prompt" {
		t.Error("Expected chat/casual prompt to match")
	}
}

func TestGetSliderPrompt(t *testing.T) {
	cfg := DefaultConfig()

	// Casual + Short
	p1 := cfg.GetSliderPrompt(10, 10, "chat")
	if !strings.Contains(p1, "Highly casual") || !strings.Contains(p1, "Ultra-concise") {
		t.Errorf("Expected casual and short prompt, got: %s", p1)
	}

	// Formal + Expanded
	p2 := cfg.GetSliderPrompt(90, 90, "email")
	if !strings.Contains(p2, "Highly formal") || !strings.Contains(p2, "Comprehensive and detailed") {
		t.Errorf("Expected formal and expanded prompt, got: %s", p2)
	}

	// Balanced
	p3 := cfg.GetSliderPrompt(50, 50, "normal")
	if !strings.Contains(p3, "Balanced") || !strings.Contains(p3, "Maintain roughly the same length") {
		t.Errorf("Expected balanced prompt, got: %s", p3)
	}
}
