package rewriter

import (
	"strings"
)

// TextType represents the type of text being processed
type TextType string

const (
	TextTypeEmail   TextType = "email"
	TextTypeChat    TextType = "chat"
	TextTypeNormal  TextType = "normal"
	TextTypeCode    TextType = "code"
	TextTypeList    TextType = "list"
	TextTypeUnknown TextType = "unknown"
)

// TextTypeInfo contains display information for each text type
type TextTypeInfo struct {
	Type        string
	Label       string
	Icon        string
	Description string
}

var TextTypeInfoMap = map[TextType]TextTypeInfo{
	TextTypeEmail:   {Type: "email", Label: "Email", Icon: "📧", Description: "Formal correspondence with greeting and signature"},
	TextTypeChat:    {Type: "chat", Label: "Chat/Message", Icon: "💬", Description: "Casual conversation or instant messages"},
	TextTypeNormal:  {Type: "normal", Label: "Normal Text", Icon: "📝", Description: "General prose or paragraphs"},
	TextTypeCode:    {Type: "code", Label: "Code", Icon: "💻", Description: "Programming code or technical content"},
	TextTypeList:    {Type: "list", Label: "List/Points", Icon: "•••", Description: "Bullet points or numbered items"},
	TextTypeUnknown: {Type: "unknown", Label: "Unknown", Icon: "❓", Description: "Could not determine text type"},
}

// AllTextTypes returns all available text types
func AllTextTypes() []TextType {
	return []TextType{
		TextTypeEmail,
		TextTypeChat,
		TextTypeNormal,
		TextTypeCode,
		TextTypeList,
	}
}

// DetectTextType analyzes text and returns the detected type with confidence
func DetectTextType(text string) (TextType, float64) {
	text = strings.TrimSpace(text)
	if text == "" {
		return TextTypeUnknown, 0
	}

	lowerText := strings.ToLower(text)
	lines := strings.Split(text, "\n")

	score := make(map[TextType]float64)

	// Email detection patterns
	emailPatterns := []struct {
		pattern string
		weight  float64
	}{
		{"dear ", 3.0},
		{"hi ", 1.5},
		{"hello ", 1.5},
		{"to whom it may concern", 4.0},
		{"regards", 3.0},
		{"sincerely", 3.0},
		{"best regards", 3.5},
		{"kind regards", 3.5},
		{"thanks", 1.0},
		{"thank you", 1.5},
		{"subject:", 4.0},
		{"from:", 2.0},
		{"to:", 2.0},
		{"cc:", 2.5},
		{"bcc:", 2.5},
		{"attached", 1.5},
		{"attachment", 1.5},
	}

	for _, p := range emailPatterns {
		if strings.Contains(lowerText, p.pattern) {
			score[TextTypeEmail] += p.weight
		}
	}

	// Check for email structure (greeting at start, signature at end)
	if len(lines) > 2 {
		firstLine := strings.ToLower(strings.TrimSpace(lines[0]))
		lastLine := strings.ToLower(strings.TrimSpace(lines[len(lines)-1]))

		// Check greeting at beginning
		if strings.HasPrefix(firstLine, "dear ") ||
			strings.HasPrefix(firstLine, "hi ") ||
			strings.HasPrefix(firstLine, "hello ") ||
			strings.HasPrefix(firstLine, "to ") {
			score[TextTypeEmail] += 2.0
		}

		// Check signature at end
		if strings.Contains(lastLine, "regards") ||
			strings.Contains(lastLine, "sincerely") ||
			strings.Contains(lastLine, "best") ||
			strings.Contains(lastLine, "thanks") {
			score[TextTypeEmail] += 2.0
		}
	}

	// Chat detection patterns
	chatPatterns := []struct {
		pattern string
		weight  float64
	}{
		{":)", 1.0},
		{":(", 1.0},
		{":d", 1.0},
		{"lol", 1.5},
		{"omg", 1.5},
		{"wtf", 1.5},
		{"btw", 1.0},
		{"imo", 1.0},
		{"imho", 1.0},
		{"tbh", 1.0},
		{"haha", 1.0},
		{"hehe", 1.0},
		{":p", 1.0},
		{"<3", 1.0},
		{"jk", 0.5},
		{"idk", 1.0},
		{"np", 0.5},
		{"ty", 0.5},
		{"tysm", 1.0},
		{"rn", 0.5},
	}

	for _, p := range chatPatterns {
		if strings.Contains(lowerText, p.pattern) {
			score[TextTypeChat] += p.weight
		}
	}

	// Check for chat-like structure (short lines, timestamps)
	shortLineCount := 0
	timestampCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) < 50 && trimmed != "" {
			shortLineCount++
		}
		// Check for timestamps like "10:30 AM" or "[12:45]"
		if strings.Contains(trimmed, ":") &&
			(strings.Contains(trimmed, "am") || strings.Contains(trimmed, "pm") ||
				strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "(")) {
			timestampCount++
		}
	}

	if len(lines) > 0 {
		shortLineRatio := float64(shortLineCount) / float64(len(lines))
		if shortLineRatio > 0.5 {
			score[TextTypeChat] += 2.0
		}
	}
	if timestampCount > 0 {
		score[TextTypeChat] += float64(timestampCount) * 1.5
	}

	// Code detection patterns
	codePatterns := []struct {
		pattern string
		weight  float64
	}{
		{"func ", 3.0},
		{"function", 2.0},
		{"def ", 3.0},
		{"class ", 2.0},
		{"import ", 2.0},
		{"#include", 3.0},
		{"const ", 1.5},
		{"var ", 1.5},
		{"let ", 1.5},
		{"if (", 1.5},
		{"for (", 1.5},
		{"while (", 1.5},
		{"return", 1.0},
		{"{", 0.5},
		{"}", 0.5},
		{"//", 1.0},
		{"/*", 2.0},
		{"*/", 2.0},
		{"console.log", 2.5},
		{"print(", 2.0},
		{"fmt.", 2.0},
		{"public ", 1.5},
		{"private ", 1.5},
	}

	for _, p := range codePatterns {
		if strings.Contains(text, p.pattern) {
			score[TextTypeCode] += p.weight
		}
	}

	// Check indentation patterns for code
	indentedLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			indentedLines++
		}
	}
	if len(lines) > 0 {
		indentRatio := float64(indentedLines) / float64(len(lines))
		if indentRatio > 0.3 {
			score[TextTypeCode] += 3.0
		}
	}

	// List detection
	bulletCount := 0
	numberedCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") ||
			strings.HasPrefix(trimmed, "• ") ||
			strings.HasPrefix(trimmed, "* ") {
			bulletCount++
		}
		// Check for numbered lists like "1.", "2)", etc.
		if len(trimmed) > 2 {
			firstChar := trimmed[0]
			if firstChar >= '0' && firstChar <= '9' {
				if trimmed[1] == '.' || trimmed[1] == ')' || trimmed[1] == ' ' {
					numberedCount++
				}
			}
		}
	}

	if bulletCount >= 2 {
		score[TextTypeList] += float64(bulletCount) * 2.0
	}
	if numberedCount >= 2 {
		score[TextTypeList] += float64(numberedCount) * 1.5
	}

	// Find the type with highest score
	var bestType TextType = TextTypeNormal
	var bestScore float64 = 0.5 // Default baseline for normal text

	for t, s := range score {
		if s > bestScore {
			bestScore = s
			bestType = t
		}
	}

	// Calculate confidence (0-1)
	confidence := normalizeScore(bestScore)

	return bestType, confidence
}

// normalizeScore converts a raw score to a 0-1 confidence value
func normalizeScore(score float64) float64 {
	// Typical scores range from 0-10, so we normalize
	normalized := score / 8.0
	if normalized > 1.0 {
		return 1.0
	}
	if normalized < 0.0 {
		return 0.0
	}
	return normalized
}

// GetTextTypeInfo returns information about a text type
func GetTextTypeInfo(t TextType) (TextTypeInfo, bool) {
	info, ok := TextTypeInfoMap[t]
	if ok && info.Type == "" {
		info.Type = string(t)
	}
	return info, ok
}
