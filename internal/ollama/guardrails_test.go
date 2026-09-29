package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatible_GuardrailParameters(t *testing.T) {
	var capturedReq OpenAIRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := OpenAIResponse{
			Choices: []Choice{
				{
					Message: Message{
						Role:    "assistant",
						Content: "Corrected sentence.",
					},
					FinishReason: "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(server.URL, "test-model", "")
	res, err := client.GenerateRewrite(context.Background(), "This is a test sentence with mistakes.", "grammar", "System prompt")
	if err != nil {
		t.Fatalf("GenerateRewrite failed: %v", err)
	}

	if res != "Corrected sentence." {
		t.Errorf("Unexpected result: %q", res)
	}

	// Verify guardrail parameters were set in request
	if capturedReq.Temperature > 0.25 {
		t.Errorf("Expected low temperature for grammar, got %f", capturedReq.Temperature)
	}
	if capturedReq.FrequencyPenalty <= 0 {
		t.Errorf("Expected positive FrequencyPenalty, got %f", capturedReq.FrequencyPenalty)
	}
	if capturedReq.PresencePenalty <= 0 {
		t.Errorf("Expected positive PresencePenalty, got %f", capturedReq.PresencePenalty)
	}
	if len(capturedReq.Stop) == 0 {
		t.Errorf("Expected stop sequences to be populated")
	}
	if capturedReq.MaxTokens <= 0 || capturedReq.MaxTokens > 3072 {
		t.Errorf("Expected reasonable MaxTokens, got %d", capturedReq.MaxTokens)
	}
}

func TestOpenAICompatible_RepetitionLoopCleanedInRewrite(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OpenAIResponse{
			Choices: []Choice{
				{
					Message: Message{
						Role:    "assistant",
						Content: "The meeting is at noon. Please attend attend attend attend attend",
					},
					FinishReason: "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(server.URL, "test-model", "")
	res, err := client.GenerateRewrite(context.Background(), "input", "standard", "System prompt")
	if err != nil {
		t.Fatalf("GenerateRewrite failed: %v", err)
	}

	// The repetition loop should be detected and cleaned
	expected := "The meeting is at noon. Please attend"
	if res != expected {
		t.Errorf("Expected cleaned text %q, got %q", expected, res)
	}
}

func TestOpenAICompatible_StreamingLoopBreak(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected flusher")
		}

		// Send normal chunks
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"Good morning. \"}}]}\n\n"))
		flusher.Flush()

		// Send repeating chunks in a loop
		for i := 0; i < 10; i++ {
			w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"repeat \"}}]}\n\n"))
			flusher.Flush()
		}

		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(server.URL, "test-model", "")
	ch, err := client.GenerateStream(context.Background(), "input", "standard", "System prompt")
	if err != nil {
		t.Fatalf("GenerateStream failed: %v", err)
	}

	var chunks []string
	isDone := false
	for chunk := range ch {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		if chunk.Response != "" {
			chunks = append(chunks, chunk.Response)
		}
		if chunk.Done {
			isDone = true
			break
		}
	}

	if !isDone {
		t.Error("Expected stream to terminate with Done: true")
	}

	// Verify stream was cut short before receiving 10 repeats
	fullReceived := strings.Join(chunks, "")
	repeatCount := strings.Count(fullReceived, "repeat")
	if repeatCount > 4 {
		t.Errorf("Loop breaker should have stopped stream, but received %d repeats: %q", repeatCount, fullReceived)
	}
}
