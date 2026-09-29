package rewriter

import (
	"strings"
	"testing"
)

// TestComputeDiff tests the diff computation between two strings
func TestComputeDiff(t *testing.T) {
	// Create a mock rewriter (we don't need a real client for diff tests)
	r := &Rewriter{}

	testCases := []struct {
		name     string
		original string
		rewritten string
		hasDiff  bool
	}{
		{
			name:     "identical strings",
			original: "Hello, world!",
			rewritten: "Hello, world!",
			hasDiff:  false,
		},
		{
			name:     "single character change",
			original: "Hello, world!",
			rewritten: "Hello, World!",
			hasDiff:  true,
		},
		{
			name:     "added text",
			original: "Hello",
			rewritten: "Hello, world!",
			hasDiff:  true,
		},
		{
			name:     "removed text",
			original: "Hello, world!",
			rewritten: "Hello",
			hasDiff:  true,
		},
		{
			name:     "empty strings",
			original: "",
			rewritten: "",
			hasDiff:  false,
		},
		{
			name:     "one empty string",
			original: "",
			rewritten: "Hello",
			hasDiff:  true,
		},
		{
			name:     "complex changes",
			original: "This is the original text.",
			rewritten: "This is the modified text.",
			hasDiff:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := r.ComputeDiff(tc.original, tc.rewritten)

			if result.HasDiff != tc.hasDiff {
				t.Errorf("Expected HasDiff to be %v, got %v", tc.hasDiff, result.HasDiff)
			}

			// Check that HTML is generated
			if result.HTML == "" && tc.hasDiff {
				t.Error("Expected HTML to be generated for diffs")
			}
		})
	}
}

// TestAnalyzeTone tests the tone analysis function
func TestAnalyzeTone(t *testing.T) {
	testCases := []struct {
		name           string
		text           string
		expectedTone   string
		expectedToneLabel string
	}{
		{
			name:           "formal text",
			text:           "Therefore, we must accordingly proceed with the aforementioned plan notwithstanding the circumstances.",
			expectedTone:   "formal",
			expectedToneLabel: "Formal",
		},
		{
			name:           "casual text",
			text:           "OMG that's so cool lol",
			expectedTone:   "casual",
			expectedToneLabel: "Casual",
		},
		{
			name:           "neutral text",
			text:           "This is a normal sentence with regular words.",
			expectedTone:   "neutral",
			expectedToneLabel: "Neutral",
		},
		{
			name:           "inquisitive text",
			text:           "What is this? How does it work? Why is it like this?",
			expectedTone:   "inquisitive",
			expectedToneLabel: "Inquisitive",
		},
		{
			name:           "urgent text",
			text:           "We need to fix this ASAP! This is urgent and must be done immediately.",
			expectedTone:   "urgent",
			expectedToneLabel: "Urgent",
		},
		{
			name:           "empty text",
			text:           "",
			expectedTone:   "neutral",
			expectedToneLabel: "Neutral",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := AnalyzeTone(tc.text)

			if result.Tone != tc.expectedTone {
				t.Errorf("Expected tone '%s', got '%s'", tc.expectedTone, result.Tone)
			}
			if result.ToneLabel != tc.expectedToneLabel {
				t.Errorf("Expected tone label '%s', got '%s'", tc.expectedToneLabel, result.ToneLabel)
			}

			// Check that word count is correct
			if result.WordCount != len(strings.Fields(tc.text)) {
				t.Errorf("Expected word count %d, got %d", len(strings.Fields(tc.text)), result.WordCount)
			}
		})
	}
}

// TestDetectTextType tests the text type detection function
func TestDetectTextType(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected TextType
	}{
		{
			name:     "email",
			text:     "Dear John, I hope this email finds you well. Best regards, Franco",
			expected: TextTypeEmail,
		},
		{
			name:     "chat",
			text:     "Hey! How's it going?",
			expected: TextTypeChat,
		},
		{
			name:     "code",
			text:     "func main() { fmt.Println(\"Hello, world!\") }",
			expected: TextTypeCode,
		},
		{
			name:     "list",
			text:     "- Item 1\n- Item 2\n- Item 3",
			expected: TextTypeList,
		},
		{
			name:     "normal",
			text:     "This is a normal paragraph of text.",
			expected: TextTypeNormal,
		},
		{
			name:     "empty",
			text:     "",
			expected: TextTypeUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			detectedType, _ := DetectTextType(tc.text)

			if detectedType != tc.expected {
				t.Errorf("Expected text type '%s', got '%s'", tc.expected, detectedType)
			}
		})
	}
}

// TestIsValidStyle tests the style validation function
func TestIsValidStyle(t *testing.T) {
	testCases := []struct {
		style    string
		expected bool
	}{
		{"grammar", true},
		{"paraphrase", true},
		{"standard", true},
		{"formal", true},
		{"casual", true},
		{"creative", true},
		{"short", true},
		{"expand", true},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.style, func(t *testing.T) {
			result := isValidStyle(tc.style)

			if result != tc.expected {
				t.Errorf("Expected isValidStyle('%s') to be %v, got %v", tc.style, tc.expected, result)
			}
		})
	}
}

// TestIsValidAnalysisStyle tests the analysis style validation function
func TestIsValidAnalysisStyle(t *testing.T) {
	testCases := []struct {
		style    string
		expected bool
	}{
		{"summarize", true},
		{"bullets", true},
		{"insights", true},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.style, func(t *testing.T) {
			result := isValidAnalysisStyle(tc.style)

			if result != tc.expected {
				t.Errorf("Expected isValidAnalysisStyle('%s') to be %v, got %v", tc.style, tc.expected, result)
			}
		})
	}
}

// TestGetStyleInfo tests the style info retrieval
func TestGetStyleInfo(t *testing.T) {
	testCases := []struct {
		style    string
		valid    bool
		label    string
	}{
		{"grammar", true, "Grammar & Spelling"},
		{"formal", true, "Formal"},
		{"invalid", false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.style, func(t *testing.T) {
			info, exists := GetStyleInfo(tc.style)

			if exists != tc.valid {
				t.Errorf("Expected GetStyleInfo('%s') exists to be %v, got %v", tc.style, tc.valid, exists)
			}

			if exists && info.Label != tc.label {
				t.Errorf("Expected label '%s', got '%s'", tc.label, info.Label)
			}
		})
	}
}

// TestCleanResponse tests the cleanResponse function
func TestCleanResponse(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes XML tags",
			input:    "<response>Hello, world!</response>",
			expected: "Hello, world!",
		},
		{
			name:     "removes input tags",
			input:    "<input>Test</input>",
			expected: "Test",
		},
		{
			name:     "handles nested tags",
			input:    "<root><child>Nested</child></root>",
			expected: "Nested",
		},
		{
			name:     "handles plain text",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles multiple tags",
			input:    "<tag1>First</tag1><tag2>Second</tag2>",
			expected: "FirstSecond",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cleanResponse(tc.input)

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestRewriteStyles tests that all rewrite styles are valid
func TestRewriteStyles(t *testing.T) {
	for _, style := range RewriteStyles {
		if !isValidStyle(style) {
			t.Errorf("Rewrite style '%s' is not valid", style)
		}
	}
}

// TestAnalysisStyles tests that all analysis styles are valid
func TestAnalysisStyles(t *testing.T) {
	for _, style := range AnalysisStyles {
		if !isValidAnalysisStyle(style) {
			t.Errorf("Analysis style '%s' is not valid", style)
		}
	}
}

func TestSlidersValidation(t *testing.T) {
	r := &Rewriter{}
	// Empty text validation
	_, err := r.GenerateRewriteWithSliders(nil, "", 50, 50, TextTypeNormal, false)
	if err == nil {
		t.Error("Expected error for empty text, got nil")
	}

	// Stream empty text validation
	_, err = r.GenerateStreamWithSliders(nil, "", 50, 50, TextTypeNormal, false)
	if err == nil {
		t.Error("Expected error for streaming empty text, got nil")
	}
}
