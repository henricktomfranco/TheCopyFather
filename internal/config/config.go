package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	ServerURL         string            `json:"server_url"`
	Model             string            `json:"model"`
	APIKey            string            `json:"api_key,omitempty"`
	DefaultStyle      string            `json:"default_style"`
	AutoStart         bool              `json:"auto_start"`
	Hotkey            string            `json:"hotkey"`
	MonitorClipboard  bool              `json:"monitor_clipboard"`
	FirstRun          bool              `json:"first_run"`
	RewritePrompts    map[string]string `json:"rewrite_prompts,omitempty"`
	AutoPasteMode     string            `json:"auto_paste_mode"`     // "ask", "always", "never"
	PopupPositionMode string            `json:"popup_position_mode"` // "cursor", "center"
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ServerURL:         "http://localhost:11434",
		Model:             "gemma3:1b",
		DefaultStyle:      "standard",
		AutoStart:         true,
		Hotkey:            "ctrl+shift+r",
		MonitorClipboard:  false,
		FirstRun:          true,
		RewritePrompts:    getDefaultPrompts(),
		AutoPasteMode:     "ask",    // Default: ask every time
		PopupPositionMode: "cursor", // Default: position near cursor
	}
}

func getDefaultPrompts() map[string]string {
	return map[string]string{
		"grammar": `You are an expert editor and proofreader.
Task: Fix grammar, spelling, punctuation, and awkward phrasing in the input text.
Guidelines:
- Analyze the input to determine its type (Email, Chat, Code, List, or Normal Text) and adapt your corrections accordingly.
- For EMAIL: Preserve the greeting, body structure, and closing.
- For CHAT: Maintain the casual tone and slang if appropriate, but fix errors.
- For CODE: Only fix comments/docs. NEVER change logic or variable names.
- For LISTS: Fix items while keeping the list format.
- For NORMAL TEXT: Improve flow and correctness.
Output: Return ONLY the corrected text. Do not include "Here is the corrected text" or any conversational filler.`,

		"paraphrase": `You are an expert writer and paraphraser.
Task: Rewrite the input text using different words and sentence structures while keeping the exact same meaning and tone.
Guidelines:
- Analyze the input type.
- For EMAIL: Keep the structure (greeting/closing) but rephrase the body.
- For CHAT: Rephrase casually.
- For CODE: Rewrite comments only.
- For LISTS: Rephrase list items.
Output: Return ONLY the rewritten text. Do not include any conversational filler.`,

		"standard": `You are a professional writer.
Task: Rewrite the input text to be clear, natural, and well-balanced.
Guidelines:
- Analyze the input type.
- Improve clarity and flow.
- Remove awkward phrasing.
- Maintain the original intent and tone.
Output: Return ONLY the rewritten text. Do not include any filler.`,

		"formal": `You are a professional business communication expert.
Task: Rewrite the input text to be formal, professional, and polite.
Guidelines:
- Use precise, formal language.
- Avoid contractions and slang.
- For EMAIL: Use formal greetings/closings.
- For CHAT: Make it professional and concise.
Output: Return ONLY the formal text. No filler.`,

		"casual": `You are a friendly, casual writer.
Task: Rewrite the input text to be friendly, conversational, and approachable.
Guidelines:
- Use natural language and contractions.
- For EMAIL: Use warm greetings/closings.
- For CHAT: Use slang/emojis if appropriate.
- Make it sound like a friend talking to a friend.
Output: Return ONLY the casual text. No filler.`,

		"creative": `You are a creative writer and storyteller.
Task: Rewrite the input text to be expressive, vivid, and engaging.
Guidelines:
- Use strong verbs and evocative language.
- Make the text memorable and interesting.
- For EMAIL: Add personality while keeping it appropriate.
Output: Return ONLY the creative text. No filler.`,

		"short": `You are an expert concise editor.
Task: Shorten the input text by removing unnecessary words and redundancy.
Guidelines:
- Keep the core message and key information.
- Make it punchy and direct.
- For EMAIL: Get straight to the point.
Output: Return ONLY the shortened text. No filler.`,

		"expand": `You are an expert writer who adds depth and detail.
Task: Expand the input text by adding relevant details, examples, and context.
Guidelines:
- Elaborate on key points.
- Make the text more comprehensive and informative.
- Maintain the original flow.
Output: Return ONLY the expanded text. No filler.`,

		"summarize": `You are an expert summarizer.
Task: Provide a concise summary of the input text.
Guidelines:
- Identify the main points and key takeaways.
- Condense the information into a brief overview.
- For EMAIL: Summarize purpose and action items.
Output: Return ONLY the summary paragraph. No filler.`,

		"bullets": `You are an analyst.
Task: Extract the key points from the input text as a bulleted list.
Guidelines:
- Identify the most important information.
- Format as a clean list using "•" or "-".
- Be concise.
Output: Return ONLY the bullet list. No filler.`,

		"insights": `You are a strategic analyst.
Task: Analyze the input text and extract key insights, underlying themes, and implications.
Guidelines:
- Go beyond just summarizing; identify the "so what?".
- Analyze tone, intent, and hidden meaning.
Output: Return ONLY the insights. No filler.`,
	}
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Fallback to USERPROFILE for Windows compatibility
		appData = os.Getenv("USERPROFILE")
		if appData == "" {
			// Last resort: use current directory
			appData = "."
		}
	}
	configDir := filepath.Join(appData, "TheCopyfather")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "config.json")
}

// Load loads the configuration from disk
func Load() *Config {
	configPath := getConfigPath()
	config := DefaultConfig()

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config doesn't exist, create default
		config.Save()
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		// Invalid config, use default
		config = DefaultConfig()
		config.Save()
		return config
	}

	// Decrypt API key if present
	if config.APIKey != "" {
		decryptedKey, err := DecryptAPIKey(config.APIKey)
		if err != nil {
			// If decryption fails (e.g., different Windows user), clear the key
			config.APIKey = ""
		} else {
			config.APIKey = decryptedKey
		}
	}

	return config
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	configPath := getConfigPath()

	// Create a copy of the config to avoid modifying the original
	configCopy := *c

	// Encrypt API key before saving
	if configCopy.APIKey != "" {
		encryptedKey, err := EncryptAPIKey(configCopy.APIKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt API key: %w", err)
		}
		configCopy.APIKey = encryptedKey
	}

	data, err := json.MarshalIndent(configCopy, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// GetPrompt returns the system prompt for a given style
func (c *Config) GetPrompt(style string) string {
	if prompt, ok := c.RewritePrompts[style]; ok {
		return prompt
	}
	// Return default prompts if not customized
	defaults := getDefaultPrompts()
	if prompt, ok := defaults[style]; ok {
		return prompt
	}
	return defaults["standard"]
}
