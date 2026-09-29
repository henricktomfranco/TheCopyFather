package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"textrewriter/internal/guardrails"
)

// OpenAICompatibleClient handles communication with OpenAI-compatible APIs (e.g., NVIDIA NIM, LM Studio, embedded llama.cpp)
type OpenAICompatibleClient struct {
	baseURL          string
	model            string
	apiKey           string
	httpClient       *http.Client
	disableStreaming bool
	disableThinking  bool
}

// NewOpenAICompatibleClient creates a new OpenAI-compatible API client
func NewOpenAICompatibleClient(baseURL, model, apiKey string) *OpenAICompatibleClient {
	return NewOpenAICompatibleClientWithOptions(baseURL, model, apiKey, false)
}

// NewOpenAICompatibleClientWithOptions creates a new OpenAI-compatible API client with additional options
func NewOpenAICompatibleClientWithOptions(baseURL, model, apiKey string, disableStreaming bool) *OpenAICompatibleClient {
	return NewOpenAICompatibleClientWithAllOptions(baseURL, model, apiKey, disableStreaming, false)
}

// NewOpenAICompatibleClientWithAllOptions creates a new OpenAI-compatible API client with all options
func NewOpenAICompatibleClientWithAllOptions(baseURL, model, apiKey string, disableStreaming, disableThinking bool) *OpenAICompatibleClient {
	if baseURL == "" {
		baseURL = "https://integrate.api.nvidia.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &OpenAICompatibleClient{
		baseURL:          baseURL,
		model:            model,
		apiKey:           apiKey,
		httpClient:       &http.Client{Timeout: 120 * time.Second},
		disableStreaming: disableStreaming,
		disableThinking:  disableThinking,
	}
}

// SetDisableThinking configures whether reasoning/thinking tags are stripped
func (c *OpenAICompatibleClient) SetDisableThinking(disable bool) {
	c.disableThinking = disable
}

// OpenAIRequest represents the request body for OpenAI-compatible APIs
type OpenAIRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
	RepeatPenalty    float64   `json:"repeat_penalty,omitempty"`
	Stop             []string  `json:"stop,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI-compatible APIs
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a single choice in the response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIStreamResponse represents a single chunk from a streaming response
type OpenAIStreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a single choice in a streaming response
type StreamChoice struct {
	Index        int     `json:"index"`
	Delta        Message `json:"delta"`
	FinishReason string  `json:"finish_reason"`
}

// GenerateRewrite generates a rewrite of the given text using the specified style
func (c *OpenAICompatibleClient) GenerateRewrite(ctx context.Context, text, style, systemPrompt string) (string, error) {
	// Sanitize the input text
	sanitizedText := sanitizeInput(text)

	finalSystemPrompt := systemPrompt
	if c.disableThinking {
		finalSystemPrompt += "\n\nCRITICAL: Do not output reasoning, thoughts, internal explanations, or <think> tags. Provide ONLY the final rewritten text directly."
	}

	// Build messages for OpenAI-compatible API
	messages := []Message{
		{
			Role:    "system",
			Content: finalSystemPrompt,
		},
		{
			Role:    "user",
			Content: sanitizedText,
		},
	}

	reqBody := OpenAIRequest{
		Model:            c.model,
		Messages:         messages,
		Temperature:      guardrails.GetTemperatureForStyle(style),
		TopP:             guardrails.DefaultTopP,
		MaxTokens:        guardrails.CalculateMaxTokens(text, style),
		FrequencyPenalty: guardrails.DefaultFrequencyPenalty,
		PresencePenalty:  guardrails.DefaultPresencePenalty,
		RepeatPenalty:    guardrails.DefaultRepeatPenalty,
		Stop:             guardrails.GetDefaultStopSequences(),
		Stream:           false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	var result OpenAIResponse
	err = retryWithBackoff(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return formatOpenAIConnectionError(err, c.baseURL)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errStr := string(body)
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, errStr)
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned in response")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if c.disableThinking {
		content = stripThinking(content)
	}

	// Guardrail: Detect and clean any repetition loops
	repResult := guardrails.DetectAndCleanRepetition(content)
	if repResult.HasLoop {
		content = repResult.CleanText
	}

	if content == "" {
		return "", fmt.Errorf("empty response from API")
	}
	return content, nil
}

// GenerateStream generates a rewrite and streams the response chunk by chunk
func (c *OpenAICompatibleClient) GenerateStream(ctx context.Context, text, style, systemPrompt string) (<-chan ClientStreamResponse, error) {
	// When streaming is disabled, fall back to non-streaming
	if c.disableStreaming {
		return c.generateNonStreaming(ctx, text, style, systemPrompt)
	}

	// Sanitize the input text
	sanitizedText := sanitizeInput(text)

	finalSystemPrompt := systemPrompt
	if c.disableThinking {
		finalSystemPrompt += "\n\nCRITICAL: Do not output reasoning, thoughts, internal explanations, or <think> tags. Provide ONLY the final rewritten text directly."
	}

	// Build messages for OpenAI-compatible API
	messages := []Message{
		{
			Role:    "system",
			Content: finalSystemPrompt,
		},
		{
			Role:    "user",
			Content: sanitizedText,
		},
	}

	reqBody := OpenAIRequest{
		Model:            c.model,
		Messages:         messages,
		Temperature:      guardrails.GetTemperatureForStyle(style),
		TopP:             guardrails.DefaultTopP,
		MaxTokens:        guardrails.CalculateMaxTokens(text, style),
		FrequencyPenalty: guardrails.DefaultFrequencyPenalty,
		PresencePenalty:  guardrails.DefaultPresencePenalty,
		RepeatPenalty:    guardrails.DefaultRepeatPenalty,
		Stop:             guardrails.GetDefaultStopSequences(),
		Stream:           true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create channel for streaming responses
	outputChan := make(chan ClientStreamResponse, 100)

	go func() {
		defer close(outputChan)

		var lastErr error
		for attempt := 0; attempt < MaxRetries; attempt++ {
			if attempt > 0 {
				delay := time.Duration(1<<uint(attempt-1)) * time.Second
				select {
				case <-ctx.Done():
					outputChan <- ClientStreamResponse{Error: ctx.Err()}
					return
				case <-time.After(delay):
				}
			}

			req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
			if err != nil {
				lastErr = fmt.Errorf("failed to create request: %w", err)
				continue
			}

			req.Header.Set("Content-Type", "application/json")
			if c.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+c.apiKey)
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				lastErr = formatOpenAIConnectionError(err, c.baseURL)
				if isLocalConnectionRefused(err, c.baseURL) {
					outputChan <- ClientStreamResponse{Error: lastErr}
					return
				}
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				errStr := string(body)
				lastErr = fmt.Errorf("API error (status %d): %s", resp.StatusCode, errStr)
				// Don't retry on client errors (4xx)
				if strings.Contains(lastErr.Error(), "status 4") {
					outputChan <- ClientStreamResponse{Error: lastErr}
					return
				}
				continue
			}

			// Process SSE streaming response (OpenAI format: "data: {...}\n\n")
			scanner := bufio.NewScanner(resp.Body)
			inThinking := false
			thinkBuffer := ""
			var accumulatedText strings.Builder

			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					return
				}

				var chunk OpenAIStreamResponse
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					outputChan <- ClientStreamResponse{Error: fmt.Errorf("failed to decode stream chunk: %w", err)}
					return
				}

				if len(chunk.Choices) > 0 {
					content := chunk.Choices[0].Delta.Content
					if content != "" {
						if c.disableThinking {
							if inThinking {
								thinkBuffer += content
								if endIdx := strings.Index(strings.ToLower(thinkBuffer), "</think>"); endIdx != -1 {
									content = thinkBuffer[endIdx+8:]
									inThinking = false
									thinkBuffer = ""
								} else {
									content = ""
								}
							} else {
								if startIdx := strings.Index(strings.ToLower(content), "<think>"); startIdx != -1 {
									before := content[:startIdx]
									thinkBuffer = content[startIdx:]
									inThinking = true
									if endIdx := strings.Index(strings.ToLower(thinkBuffer), "</think>"); endIdx != -1 {
										content = before + thinkBuffer[endIdx+8:]
										inThinking = false
										thinkBuffer = ""
									} else {
										content = before
									}
								}
							}
						}

						if content != "" {
							accumulatedText.WriteString(content)

							// Guardrail: Active loop detection during streaming
							repCheck := guardrails.DetectAndCleanRepetition(accumulatedText.String())
							if repCheck.HasLoop {
								// Repetition loop detected! Terminate stream immediately.
								outputChan <- ClientStreamResponse{
									Done: true,
								}
								return
							}

							outputChan <- ClientStreamResponse{
								Response: content,
								Done:     chunk.Choices[0].FinishReason != "",
							}
						}
					}
				}

				if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
					return
				}
			}
			if err := scanner.Err(); err != nil {
				outputChan <- ClientStreamResponse{Error: fmt.Errorf("stream read error: %w", err)}
				return
			}
		}

		// If we exhausted retries
		if lastErr != nil {
			outputChan <- ClientStreamResponse{Error: fmt.Errorf("failed after %d attempts: %w", MaxRetries, lastErr)}
		}
	}()

	return outputChan, nil
}

// HealthCheck checks if the OpenAI-compatible API is reachable
func (c *OpenAICompatibleClient) HealthCheck() error {
	req, err := http.NewRequest("GET", c.baseURL+"/models", nil)
	if err != nil {
		return err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

// GetVersion returns the model name (since OpenAI-compatible APIs don't have a /version endpoint)
func (c *OpenAICompatibleClient) GetVersion() string {
	return c.model
}

// ModelInfo represents information about an available model for OpenAI-compatible APIs
type OpenAIModelInfo struct {
	ID         string        `json:"id"`
	Object     string        `json:"object"`
	Created    int64         `json:"created"`
	OwnedBy    string        `json:"owned_by"`
	Permission []interface{} `json:"permission"`
	Root       interface{}   `json:"root,omitempty"`
	Parent     interface{}   `json:"parent,omitempty"`
}

// OpenAIListModelsResponse represents the response from listing models for OpenAI-compatible APIs
type OpenAIListModelsResponse struct {
	Object string            `json:"object"`
	Data   []OpenAIModelInfo `json:"data"`
}

// GetAvailableModels returns a list of available models from the OpenAI-compatible API
func (c *OpenAICompatibleClient) GetAvailableModels() ([]string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list models: status %d", resp.StatusCode)
	}

	var result OpenAIListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(result.Data))
	for i, model := range result.Data {
		models[i] = model.ID
	}

	return models, nil
}

// generateNonStreaming wraps a non-streaming GenerateRewrite call into the streaming channel interface
func (c *OpenAICompatibleClient) generateNonStreaming(ctx context.Context, text, style, systemPrompt string) (<-chan ClientStreamResponse, error) {
	outputChan := make(chan ClientStreamResponse, 1)
	go func() {
		defer close(outputChan)
		result, err := c.GenerateRewrite(ctx, text, style, systemPrompt)
		if err != nil {
			outputChan <- ClientStreamResponse{Error: err}
			return
		}
		outputChan <- ClientStreamResponse{Response: result, Done: true}
	}()
	return outputChan, nil
}

func isLocalConnectionRefused(err error, baseURL string) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	isLocal := strings.Contains(baseURL, "127.0.0.1") || strings.Contains(baseURL, "localhost") || strings.Contains(baseURL, "8085")
	isRefused := strings.Contains(errStr, "connectex") || strings.Contains(errStr, "actively refused") || strings.Contains(errStr, "connection refused")
	return isLocal && isRefused
}

func formatOpenAIConnectionError(err error, baseURL string) error {
	if isLocalConnectionRefused(err, baseURL) {
		return fmt.Errorf("embedded llama.cpp engine is not running (port 8085 refused connection). Please download the model via the Welcome screen or check Settings")
	}
	return fmt.Errorf("failed to connect to API at %s: %w", baseURL, err)
}

func formatOpenAIError(statusCode int, errStr, model, baseURL string, connErr error) error {
	if connErr != nil {
		if isLocalConnectionRefused(connErr, baseURL) {
			return fmt.Errorf("embedded llama.cpp engine is not running (port 8085 refused connection). Please download the model via the Welcome screen or check Settings")
		}
		return fmt.Errorf("cannot connect to API at %s. Please check your internet connection or server URL", baseURL)
	}
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return fmt.Errorf("API key invalid or unauthorized (status %d). Please check your API key in Settings", statusCode)
	}
	if statusCode == http.StatusNotFound {
		return fmt.Errorf("model '%s' not found or invalid endpoint at %s", model, baseURL)
	}
	if statusCode == http.StatusTooManyRequests {
		return fmt.Errorf("rate limit exceeded (429). Please wait a moment before retrying")
	}
	return fmt.Errorf("API error (status %d): %s", statusCode, errStr)
}

// stripThinking removes reasoning or <think>...</think> blocks from text
func stripThinking(text string) string {
	for {
		start := strings.Index(strings.ToLower(text), "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(strings.ToLower(text), "</think>")
		if end != -1 {
			text = text[:start] + text[end+8:]
		} else {
			text = text[:start]
			break
		}
	}
	return strings.TrimSpace(text)
}
