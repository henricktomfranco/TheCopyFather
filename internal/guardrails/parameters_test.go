package guardrails

import (
	"testing"
)

func TestGetTemperatureForStyle(t *testing.T) {
	tests := []struct {
		style       string
		expectedMax float64
		expectedMin float64
	}{
		{"grammar", 0.20, 0.10},
		{"short", 0.30, 0.20},
		{"summarize", 0.30, 0.20},
		{"formal", 0.40, 0.30},
		{"standard", 0.40, 0.30},
		{"paraphrase", 0.50, 0.40},
		{"casual", 0.60, 0.50},
		{"creative", 0.75, 0.65},
	}

	for _, tt := range tests {
		temp := GetTemperatureForStyle(tt.style)
		if temp < tt.expectedMin || temp > tt.expectedMax {
			t.Errorf("Style %s temp %f not in range [%f, %f]", tt.style, temp, tt.expectedMin, tt.expectedMax)
		}
	}
}

func TestCalculateMaxTokens(t *testing.T) {
	// Empty input
	if tokens := CalculateMaxTokens("", "standard"); tokens < 256 {
		t.Errorf("Expected at least 256 tokens for empty input, got %d", tokens)
	}

	// Short input (10 words)
	tokensShort := CalculateMaxTokens("one two three four five six seven eight nine ten", "short")
	if tokensShort < 256 || tokensShort > 600 {
		t.Errorf("Unexpected token limit for short input: %d", tokensShort)
	}

	// Huge input should be capped
	hugeInput := ""
	for i := 0; i < 5000; i++ {
		hugeInput += "word "
	}
	tokensHuge := CalculateMaxTokens(hugeInput, "expand")
	if tokensHuge > 3072 {
		t.Errorf("Tokens for huge input exceeded cap: %d", tokensHuge)
	}
}
