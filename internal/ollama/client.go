package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles communication with the Ollama API
type Client struct {
	baseURL    string
	model      string
	apiKey     string
	version    string
	httpClient *http.Client
}

// NewClient creates a new Ollama API client
func NewClient(baseURL, model, apiKey string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:    baseURL,
		model:      model,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

// GenerateRequest represents the request body for the generate endpoint
type GenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	System  string                 `json:"system,omitempty"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// GenerateResponse represents the response from the generate endpoint
type GenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Context  []int  `json:"context,omitempty"`
}

// sanitizeInput escapes special XML characters to prevent prompt injection
func sanitizeInput(text string) string {
	// Replace XML special characters
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&apos;")
	return text
}

// MaxTextLength is the maximum length of text that can be processed
const MaxTextLength = 10000

// MaxRetries is the maximum number of retry attempts
const MaxRetries = 3

// retryWithBackoff executes the given function with exponential backoff retry logic
func retryWithBackoff(ctx context.Context, operation func() error) error {
	var lastErr error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		// Don't retry on client errors (4xx)
		if strings.Contains(err.Error(), "status 4") {
			return err
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", MaxRetries, lastErr)
}

// GenerateRewrite generates a rewrite of the given text using the specified style
func (c *Client) GenerateRewrite(ctx context.Context, text, style, systemPrompt string) (string, error) {
	// Validate text length
	if len(text) > MaxTextLength {
		return "", fmt.Errorf("text too long: %d characters (max %d)", len(text), MaxTextLength)
	}

	// Sanitize the input text to prevent XML injection
	sanitizedText := sanitizeInput(text)

	// Wrap text in XML tags to clearly separate instructions from data
	// The system prompt handles text type detection and rewriting instructions
	prompt := fmt.Sprintf("<input>\n%s\n</input>", sanitizedText)

	reqBody := c.buildGenerateRequest(prompt, systemPrompt)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	var result GenerateResponse
	err = retryWithBackoff(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to connect to Ollama: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errStr := string(body)
			if resp.StatusCode == http.StatusMethodNotAllowed {
				return fmt.Errorf("ollama API error (status 405): Method Not Allowed. Hint: Check if your server URL is correct and use https if required. (URL: %s)", req.URL.String())
			}
			return fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, errStr)
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return result.Response, nil
}

func (c *Client) buildGenerateRequest(prompt, systemPrompt string) GenerateRequest {
	actualSystemPrompt := systemPrompt
	actualPrompt := prompt

	// Strategy: Polyfill system prompt for versions < 0.1.14
	if c.isLegacyVersion() && actualSystemPrompt != "" {
		actualPrompt = fmt.Sprintf("%s\n\nUser Text:\n%s", actualSystemPrompt, prompt)
		actualSystemPrompt = ""
	}

	return GenerateRequest{
		Model:  c.model,
		Prompt: actualPrompt,
		System: actualSystemPrompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
	}
}

func (c *Client) isLegacyVersion() bool {
	if c.version == "" || c.version == "0.1.0" {
		return true
	}
	var vMajor, vMinor, vPatch int
	fmt.Sscanf(c.version, "%d.%d.%d", &vMajor, &vMinor, &vPatch)

	// System prompt was introduced in 0.1.14
	if vMajor == 0 && (vMinor < 1 || (vMinor == 1 && vPatch < 14)) {
		return true
	}
	return false
}

// ModelInfo represents information about an available model
type ModelInfo struct {
	Name   string `json:"name"`
	Digest string `json:"digest,omitempty"`
}

// ListModelsResponse represents the response from listing models
type ListModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// GetAvailableModels returns a list of available models from the Ollama server
func (c *Client) GetAvailableModels() ([]string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusMethodNotAllowed {
			return nil, fmt.Errorf("failed to list models: status 405 (Method Not Allowed). Check your server URL and protocol (URL: %s)", req.URL.String())
		}
		return nil, fmt.Errorf("failed to list models: status %d", resp.StatusCode)
	}

	var result ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, model := range result.Models {
		models[i] = model.Name
	}

	return models, nil
}

// VersionResponse represents the response from the version endpoint
type VersionResponse struct {
	Version string `json:"version"`
}

// FetchVersion gets the version from the Ollama server
func (c *Client) FetchVersion() (string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/version", nil)
	if err != nil {
		return "", err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusMethodNotAllowed {
			return "", fmt.Errorf("failed to get version: status 405 (Method Not Allowed). Check your server URL and protocol (URL: %s)", req.URL.String())
		}
		return "", fmt.Errorf("failed to get version: status %d", resp.StatusCode)
	}

	var result VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	c.version = result.Version
	return result.Version, nil
}

// HealthCheck checks if the Ollama server is reachable and fetches its version
func (c *Client) HealthCheck() error {
	_, err := c.FetchVersion()
	if err != nil {
		// Fallback to tags if version endpoint fails (older versions might not have it)
		req, err2 := http.NewRequest("GET", c.baseURL+"/api/tags", nil)
		if err2 != nil {
			return err2
		}

		resp, err2 := c.httpClient.Do(req)
		if err2 != nil {
			return fmt.Errorf("cannot connect to Ollama server: %w", err2)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("ollama server returned status %d", resp.StatusCode)
		}

		c.version = "0.1.0" // Assume older version if /api/version missing
		return nil
	}

	return nil
}

// GetVersion returns the cached version
func (c *Client) GetVersion() string {
	return c.version
}
