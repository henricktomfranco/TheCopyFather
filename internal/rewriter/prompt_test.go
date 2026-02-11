package rewriter

import (
	"strings"
	"testing"
	"textrewriter/internal/config"
)

func TestGetPromptForTextType(t *testing.T) {
	// Setup generic config
	cfg := config.DefaultConfig()
	r := &Rewriter{
		config: cfg,
	}

	testCases := []struct {
		style          string
		textType       TextType
		expectedPhrase string
	}{
		{"formal", TextTypeEmail, "You are a business communication expert"},
		{"casual", TextTypeChat, "You are a friend"},
		{"grammar", TextTypeCode, "You are a tech editor"},
		{"bullets", TextTypeNormal, "Extract 3-5 main ideas"},
	}

	for _, tc := range testCases {
		prompt := r.getPromptForTextType(tc.style, tc.textType, true)
		if !strings.Contains(prompt, tc.expectedPhrase) {
			t.Errorf("Expected prompt for %s/%s to contain '%s', but got: %s",
				tc.style, tc.textType, tc.expectedPhrase, prompt)
		}
	}
}

func TestGetPlainTextPrompt(t *testing.T) {
	testCases := []struct {
		style          string
		expectedPhrase string
	}{
		{"formal", "You are a professional communication expert"},
		{"casual", "You are a friendly, casual writer"},
		{"grammar", "You are an expert editor"},
	}

	for _, tc := range testCases {
		prompt := getPlainTextPrompt(tc.style)
		if !strings.Contains(prompt, tc.expectedPhrase) {
			t.Errorf("Expected plain prompt for %s to contain '%s', but got: %s",
				tc.style, tc.expectedPhrase, prompt)
		}
	}
}
