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
)

// OpenAICompatibleClient handles communication with OpenAI-compatible APIs (e.g., NVIDIA NIM, LM Studio)
type OpenAICompatibleClient struct {
	baseURL string
	model string
	apiKey string
	httpClient *http.Client
}

// NewOpenAICompatibleClient creates a new OpenAI-compatible API client
func NewOpenAICompatibleClient(baseURL, model, apiKey string) *OpenAICompatibleClient {
	if baseURL == "" {
		baseURL = "https://integrate.api.nvidia.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &OpenAICompatibleClient{
		baseURL: baseURL,
		model: model,
		apiKey: apiKey,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

// OpenAIRequest represents the request body for OpenAI-compatible APIs
type OpenAIRequest struct {
	Model string `json:"model"`
	Messages []Message `json:"messages"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens int `json:"max_tokens,omitempty"`
	Stream bool `json:"stream,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI-compatible APIs
type OpenAIResponse struct {
	ID string `json:"id"`
	Object string `json:"object"`
	Created int64 `json:"created"`
	Model string `json:"model"`
	Choices []Choice `json:"choices"`
	Usage Usage `json:"usage"`
}

// Choice represents a single choice in the response
type Choice struct {
	Index int `json:"index"`
	Message Message `json:"message"`
	FinishReason string `json:"finish_reason"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens int `json:"total_tokens"`
}

// OpenAIStreamResponse represents a single chunk from a streaming response
type OpenAIStreamResponse struct {
	ID string `json:"id"`
	Object string `json:"object"`
	Created int64 `json:"created"`
	Model string `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a single choice in a streaming response
type StreamChoice struct {
	Index int `json:"index"`
	Delta Message `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

// GenerateRewrite generates a rewrite of the given text using the specified style
func (c *OpenAICompatibleClient) GenerateRewrite(ctx context.Context, text, style, systemPrompt string) (string, error) {
	// Sanitize the input text
	sanitizedText := sanitizeInput(text)

	// Build messages for OpenAI-compatible API
	messages := []Message{
		{
			Role: "system",
			Content: systemPrompt,
		},
		{
			Role: "user",
			Content: sanitizedText,
		},
	}

	reqBody := OpenAIRequest{
		Model: c.model,
		Messages: messages,
		Temperature: 0.7,
		MaxTokens: 4096,
		Stream: false,
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
			return fmt.Errorf("failed to connect to API: %w", err)
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
	if content == "" {
		return "", fmt.Errorf("empty response from API")
	}
	return content, nil
}

// GenerateStream generates a rewrite and streams the response chunk by chunk
func (c *OpenAICompatibleClient) GenerateStream(ctx context.Context, text, style, systemPrompt string) (<-chan ClientStreamResponse, error) {
	// Sanitize the input text
	sanitizedText := sanitizeInput(text)

	// Build messages for OpenAI-compatible API
	messages := []Message{
		{
			Role: "system",
			Content: systemPrompt,
		},
		{
			Role: "user",
			Content: sanitizedText,
		},
	}

	reqBody := OpenAIRequest{
		Model: c.model,
		Messages: messages,
		Temperature: 0.7,
		MaxTokens: 4096,
		Stream: true,
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
				lastErr = fmt.Errorf("failed to connect to API: %w", err)
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
						outputChan <- ClientStreamResponse{
							Response: content,
							Done: chunk.Choices[0].FinishReason != "",
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
	ID string `json:"id"`
	Object string `json:"object"`
	Created int64 `json:"created"`
	OwnedBy string `json:"owned_by"`
	Permission []interface{} `json:"permission"`
	Root interface{} `json:"root,omitempty"`
	Parent interface{} `json:"parent,omitempty"`
}

// OpenAIListModelsResponse represents the response from listing models for OpenAI-compatible APIs
type OpenAIListModelsResponse struct {
	Object string `json:"object"`
	Data []OpenAIModelInfo `json:"data"`
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
