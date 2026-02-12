package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the application configuration
type Config struct {
	ServerURL         string                       `json:"server_url"`
	Model             string                       `json:"model"`
	APIKey            string                       `json:"api_key,omitempty"`
	DefaultStyle      string                       `json:"default_style"`
	AutoStart         bool                         `json:"auto_start"`
	Hotkey            string                       `json:"hotkey"`
	MonitorClipboard  bool                         `json:"monitor_clipboard"`
	FirstRun          bool                         `json:"first_run"`
	CustomPrompts     map[string]map[string]string `json:"custom_prompts,omitempty"`
	AutoPasteMode     string                       `json:"auto_paste_mode"`
	PopupPositionMode string                       `json:"popup_position_mode"`
	MiniMode          bool                         `json:"mini_mode"`
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
		CustomPrompts:     make(map[string]map[string]string),
		AutoPasteMode:     "ask",
		PopupPositionMode: "cursor",
		MiniMode:          false,
	}
}

// getDefaultPrompts returns the default prompts organized by text type
func getDefaultPrompts() map[string]map[string]string {
	return map[string]map[string]string{
		"email": {
			"grammar": `You are an expert editor.
Task: Fix grammar, spelling, and punctuation in this EMAIL.
Guidelines:
- Preserve greeting, body, and closing.
- Improve flow and professional tone.
Output: Return ONLY the corrected email.`,

			"paraphrase": `You are an expert writer.
Task: Rewrite this EMAIL using different words.
Guidelines:
- Keep the original meaning and structure.
- Maintain professional tone.
Output: Return ONLY the rewritten email.`,

			"standard": `You are a professional writer.
Task: Rewrite this EMAIL to be clear and natural.
Guidelines:
- Improve clarity and flow.
- Maintain professional business tone.
Output: Return ONLY the rewritten email.`,

			"formal": `You are a business communication expert.
Task: Rewrite this EMAIL to be highly formal.
Guidelines:
- Use formal greetings (e.g., 'Dear...') and closings (e.g., 'Sincerely').
- Use precise, professional language.
Output: Return ONLY the formal email.`,

			"casual": `You are a friendly writer.
Task: Rewrite this EMAIL to be warm and casual.
Guidelines:
- Use friendly greetings (e.g., 'Hi') and closings.
- Make it sound approachable but respectful.
Output: Return ONLY the casual email.`,

			"creative": `You are a creative writer.
Task: Rewrite this EMAIL to be engaging and memorable.
Guidelines:
- Use expressive language.
- Keep it appropriate for an email.
Output: Return ONLY the creative email.`,

			"short": `You are a concise editor.
Task: Shorten this EMAIL.
Guidelines:
- Remove unnecessary words.
- Keep the core message, greeting, and closing.
Output: Return ONLY the shortened email.`,

			"expand": `You are an expert writer.
Task: Expand this EMAIL with more detail.
Guidelines:
- Add relevant context and elaboration.
- Maintain professional structure.
Output: Return ONLY the expanded email.`,

			"summarize": `You are an analyst.
Task: Summarize this EMAIL.
Guidelines:
- Identify purpose, action items, and deadlines.
- Be concise.
Output: Return ONLY the summary paragraph.`,

			"bullets": `You are an analyst.
Task: Extract key points from this EMAIL as bullets.
Guidelines:
- List purpose, requests, and deadlines.
Output: Return ONLY the bullet list.`,

			"insights": `You are a strategic analyst.
Task: Analyze this EMAIL for insights.
Guidelines:
- Identify intent, tone, and implicit requests.
Output: Return ONLY the insights.`,
		},
		"chat": {
			"grammar": `You are an editor.
Task: Fix grammar/spelling in this CHAT message.
Guidelines:
- Keep the casual, conversational tone.
- Preserve emojis and slang if appropriate.
Output: Return ONLY the corrected message.`,

			"paraphrase": `You are a writer.
Task: Rewrite this CHAT message using different words.
Guidelines:
- Keep the friendly, conversational vibe.
Output: Return ONLY the rewritten message.`,

			"standard": `You are a writer.
Task: Rewrite this CHAT message to be natural and clear.
Guidelines:
- Make it sound like a natural conversation.
Output: Return ONLY the rewritten message.`,

			"formal": `You are a professional.
Task: Rewrite this CHAT message to be professional.
Guidelines:
- Remove slang and overly casual language.
- Make it polite and concise.
Output: Return ONLY the formal message.`,

			"casual": `You are a friend.
Task: Rewrite this CHAT message to be super casual.
Guidelines:
- Use slang, contractions, and natural chat speak.
Output: Return ONLY the casual message.`,

			"creative": `You are a creative writer.
Task: Rewrite this CHAT message to be fun and expressive.
Guidelines:
- Show personality.
Output: Return ONLY the creative message.`,

			"short": `You are a concise editor.
Task: Shorten this CHAT message.
Guidelines:
- Make it brief and punchy.
- Get straight to the point.
Output: Return ONLY the shortened message.`,

			"expand": `You are a writer.
Task: Add detail to this CHAT message.
Guidelines:
- Add context without being too long-winded.
Output: Return ONLY the expanded message.`,

			"summarize": `You are an analyst.
Task: Summarize this CHAT conversation.
Guidelines:
- Identify key decisions and topics.
Output: Return ONLY the summary.`,

			"bullets": `You are an analyst.
Task: List key points from this CHAT.
Guidelines:
- Extract decisions and action items.
Output: Return ONLY the bullet list.`,

			"insights": `You are an analyst.
Task: Analyze this CHAT for insights.
Guidelines:
- Identify sentiment and key takeaways.
Output: Return ONLY the insights.`,
		},
		"code": {
			"grammar": `You are a tech editor.
Task: Fix grammar in COMMENTS/DOCS only.
Guidelines:
- DO NOT CHANGE CODE LOGIC OR SYNTAX.
- Only fix English text in comments.
Output: Return ONLY the code with corrected comments.`,

			"paraphrase": `You are a tech editor.
Task: Rewrite COMMENTS/DOCS using different words.
Guidelines:
- DO NOT CHANGE CODE LOGIC.
- Keep code exactly as is.
Output: Return ONLY the code with rewritten comments.`,

			"standard": `You are a tech editor.
Task: Improve clarity of COMMENTS/DOCS.
Guidelines:
- DO NOT CHANGE CODE LOGIC.
- Make comments clear and concise.
Output: Return ONLY the code with improved comments.`,

			"formal": `You are a technical writer.
Task: Make COMMENTS/DOCS professional.
Guidelines:
- Use standard technical documentation style.
- DO NOT CHANGE CODE LOGIC.
Output: Return ONLY the code with formal comments.`,

			"casual": `You are a developer buddy.
Task: Make COMMENTS friendly.
Guidelines:
- Use a helpful, conversational tone in comments.
- DO NOT CHANGE CODE LOGIC.
Output: Return ONLY the code with casual comments.`,

			"creative": `You are a creative coder.
Task: Make COMMENTS expressive.
Guidelines:
- Use vivid language in comments.
- DO NOT CHANGE CODE LOGIC.
Output: Return ONLY the code with creative comments.`,

			"short": `You are a concise coder.
Task: Shorten COMMENTS.
Guidelines:
- Remove unnecessary words from comments.
- DO NOT CHANGE CODE LOGIC.
Output: Return ONLY the code with short comments.`,

			"expand": `You are a mentor.
Task: Explain the code in COMMENTS.
Guidelines:
- Add detailed explanations to comments.
- DO NOT CHANGE CODE LOGIC.
Output: Return ONLY the code with expanded comments.`,

			"summarize": `You are a tech lead.
Task: Summarize what this code does.
Guidelines:
- Explain purpose and functionality.
Output: Return ONLY the summary paragraph.`,

			"bullets": `You are a tech lead.
Task: List key features of this code.
Guidelines:
- Extract main functions and operations.
Output: Return ONLY the bullet list.`,

			"insights": `You are a software architect.
Task: Analyze this code.
Guidelines:
- Identify patterns, quality, and design choices.
Output: Return ONLY the insights.`,
		},
		"list": {
			"grammar": `You are an editor.
Task: Fix grammar in this LIST.
Guidelines:
- Preserve list format (bullets/numbers).
- Fix errors in each item.
Output: Return ONLY the corrected list.`,

			"paraphrase": `You are a writer.
Task: Rewrite this LIST using different words.
Guidelines:
- Keep the list structure.
- Rephrase each item.
Output: Return ONLY the rewritten list.`,

			"standard": `You are a writer.
Task: Rewrite this LIST to be clear and natural.
Guidelines:
- Improve flow and consistency.
- Keep list format.
Output: Return ONLY the rewritten list.`,

			"formal": `You are a professional.
Task: Rewrite this LIST to be formal.
Guidelines:
- Use precise, professional language.
- Keep list format.
Output: Return ONLY the formal list.`,

			"casual": `You are a friend.
Task: Rewrite this LIST to be casual.
Guidelines:
- Use friendly language.
- Keep list format.
Output: Return ONLY the casual list.`,

			"creative": `You are a creative writer.
Task: Rewrite this LIST to be expressive.
Guidelines:
- Use vivid language.
- Keep list format.
Output: Return ONLY the creative list.`,

			"short": `You are an editor.
Task: Shorten this LIST.
Guidelines:
- Make items concise.
- Keep list format.
Output: Return ONLY the shortened list.`,

			"expand": `You are a writer.
Task: Expand this LIST.
Guidelines:
- Add detail to each item.
- Keep list format.
Output: Return ONLY the expanded list.`,

			"summarize": `You are an analyst.
Task: Summarize this LIST.
Guidelines:
- Condense the main theme into a paragraph.
Output: Return ONLY the summary.`,

			"bullets": `You are an analyst.
Task: Refine this LIST.
Guidelines:
- Extract the most important points.
Output: Return ONLY the bullet list.`,

			"insights": `You are an analyst.
Task: Analyze this LIST.
Guidelines:
- Identify patterns and key themes.
Output: Return ONLY the insights.`,
		},
		"normal": {
			"grammar": `You are an editor.
Task: Fix grammar/spelling in this TEXT.
Guidelines:
- Improve flow and correctness.
- Keep paragraph structure.
Output: Return ONLY the corrected text.`,

			"paraphrase": `You are a writer.
Task: Rewrite this TEXT using different words.
Guidelines:
- Keep same meaning and tone.
- Maintain structure.
Output: Return ONLY the rewritten text.`,

			"standard": `You are a writer.
Task: Rewrite this TEXT to be clear and natural.
Guidelines:
- Improve clarity and flow.
- Remove awkward phrasing.
Output: Return ONLY the rewritten text.`,

			"formal": `You are a professional.
Task: Rewrite this TEXT to be formal.
Guidelines:
- Use professional, precise language.
- Avoid contractions.
Output: Return ONLY the formal text.`,

			"casual": `You are a friend.
Task: Rewrite this TEXT to be casual.
Guidelines:
- Use conversational language.
- Make it sound friendly.
Output: Return ONLY the casual text.`,

			"creative": `You are a storyteller.
Task: Rewrite this TEXT to be expressive.
Guidelines:
- Use vivid imagery and strong verbs.
Output: Return ONLY the creative text.`,

			"short": `You are an editor.
Task: Shorten this TEXT.
Guidelines:
- Remove unnecessary words.
- Keep key info.
Output: Return ONLY the shortened text.`,

			"expand": `You are a writer.
Task: Expand this TEXT.
Guidelines:
- Add detail and context.
- Elaborate on ideas.
Output: Return ONLY the expanded text.`,

			"summarize": `You are a summarizer.
Task: Summarize this TEXT.
Guidelines:
- Condense into a brief overview.
- Capture main points.
Output: Return ONLY the summary.`,

			"bullets": `You are an analyst.
Task: Convert this TEXT into key points.
Guidelines:
- Extract 3-5 main ideas as bullets.
Output: Return ONLY the bullet list.`,

			"insights": `You are an analyst.
Task: Analyze this TEXT.
Guidelines:
- Identify key themes and arguments.
Output: Return ONLY the insights.`,
		},
	}
}

// getGenericPrompts returns simple generic prompts without text type adaptation
func getGenericPrompts() map[string]string {
	return map[string]string{
		"grammar": `You are an expert editor and proofreader.
Task: Fix grammar, spelling, punctuation, and awkward phrasing.
Guidelines:
- Analyze the input type and adapt.
- For CODE: Only fix comments.
- Improve flow and correctness.
Output: Return ONLY the corrected text. No filler.`,

		"paraphrase": `You are an expert writer.
Task: Rewrite the text using different words/structure but keep the meaning.
Guidelines:
- Analyze the input type.
- Keep the original tone.
- For CODE: Comments only.
Output: Return ONLY the rewritten text. No filler.`,

		"standard": `You are a professional writer.
Task: Rewrite the text to be clear, natural, and balanced.
Guidelines:
- Improve clarity and flow.
- Remove awkward phrasing.
Output: Return ONLY the rewritten text. No filler.`,

		"formal": `You are a professional communication expert.
Task: Rewrite the text to be formal and polite.
Guidelines:
- Use precise, formal language.
- Avoid contractions/slang.
Output: Return ONLY the formal text. No filler.`,

		"casual": `You are a friendly, casual writer.
Task: Rewrite the text to be friendly and conversational.
Guidelines:
- Use natural language/contractions.
- Make it sound like a friend.
Output: Return ONLY the casual text. No filler.`,

		"creative": `You are a creative writer.
Task: Rewrite the text to be expressive and vivid.
Guidelines:
- Use strong verbs and evocative language.
Output: Return ONLY the creative text. No filler.`,

		"short": `You are a concise editor.
Task: Shorten the text by removing unnecessary words.
Guidelines:
- Keep the core message.
- Make it punchy.
Output: Return ONLY the shortened text. No filler.`,

		"expand": `You are an expert writer.
Task: Expand the text by adding details and context.
Guidelines:
- Elaborate on key points.
- Make it more comprehensive.
Output: Return ONLY the expanded text. No filler.`,

		"summarize": `You are a summarizer.
Task: Provide a concise summary.
Guidelines:
- Identify main points.
- Condense into a brief overview.
Output: Return ONLY the summary. No filler.`,

		"bullets": `You are an analyst.
Task: Extract key points as a bullet list.
Guidelines:
- Identify important info.
- Format as clean bullets.
Output: Return ONLY the bullet list. No filler.`,

		"insights": `You are a strategic analyst.
Task: Extract key insights and implications.
Guidelines:
- Identify the "so what?".
- Analyze tone and intent.
Output: Return ONLY the insights. No filler.`,
	}
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = os.Getenv("USERPROFILE")
		if appData == "" {
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
		config.Save()
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		config = DefaultConfig()
		config.Save()
		return config
	}

	if config.CustomPrompts == nil {
		config.CustomPrompts = make(map[string]map[string]string)
	}

	if config.APIKey != "" {
		decryptedKey, err := DecryptAPIKey(config.APIKey)
		if err != nil {
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

	configCopy := *c

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

// GetPrompt returns the prompt for a given style and text type
func (c *Config) GetPrompt(style, textType string) string {
	if c.CustomPrompts != nil {
		if typeMap, ok := c.CustomPrompts[style]; ok {
			if prompt, ok := typeMap[textType]; ok && prompt != "" {
				return prompt
			}
		}
	}

	defaults := getDefaultPrompts()
	if typeMap, ok := defaults[textType]; ok {
		if prompt, ok := typeMap[style]; ok {
			return prompt
		}
	}

	generic := getGenericPrompts()
	if prompt, ok := generic[style]; ok {
		return prompt
	}
	return generic["standard"]
}

// GetCustomPrompt returns a custom prompt for a given style and text type, or empty if not set
func (c *Config) GetCustomPrompt(style, textType string) string {
	if c.CustomPrompts != nil {
		if typeMap, ok := c.CustomPrompts[style]; ok {
			return typeMap[textType]
		}
	}
	return ""
}

// SetCustomPrompt sets a custom prompt for a given style and text type
// Returns error if validation fails. If prompt is empty or whitespace-only, deletes the custom prompt.
func (c *Config) SetCustomPrompt(style, textType, prompt string) error {
	// Validate inputs
	if style == "" {
		return fmt.Errorf("style cannot be empty")
	}
	if textType == "" {
		return fmt.Errorf("text type cannot be empty")
	}

	// Initialize map if needed
	if c.CustomPrompts == nil {
		c.CustomPrompts = make(map[string]map[string]string)
	}

	// If prompt is empty or whitespace-only, treat as delete
	trimmedPrompt := strings.TrimSpace(prompt)
	if trimmedPrompt == "" {
		c.DeleteCustomPrompt(style, textType)
		return nil
	}

	// Initialize style map if needed
	if _, ok := c.CustomPrompts[style]; !ok {
		c.CustomPrompts[style] = make(map[string]string)
	}

	c.CustomPrompts[style][textType] = trimmedPrompt
	return nil
}

// DeleteCustomPrompt removes a custom prompt for a given style and text type
func (c *Config) DeleteCustomPrompt(style, textType string) {
	if c.CustomPrompts != nil {
		if typeMap, ok := c.CustomPrompts[style]; ok {
			delete(typeMap, textType)
			if len(typeMap) == 0 {
				delete(c.CustomPrompts, style)
			}
		}
	}
}

// GetAllCustomPrompts returns all custom prompts
func (c *Config) GetAllCustomPrompts() map[string]map[string]string {
	if c.CustomPrompts == nil {
		return make(map[string]map[string]string)
	}
	return c.CustomPrompts
}

// HasCustomPrompt checks if a custom prompt exists for a given style and text type
func (c *Config) HasCustomPrompt(style, textType string) bool {
	if c.CustomPrompts == nil {
		return false
	}
	if typeMap, ok := c.CustomPrompts[style]; ok {
		_, exists := typeMap[textType]
		return exists
	}
	return false
}
