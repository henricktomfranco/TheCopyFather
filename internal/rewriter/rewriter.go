package rewriter

import (
	"context"
	"fmt"
	"strings"
	"textrewriter/internal/config"
	"textrewriter/internal/ollama"
)

// Rewriter handles text rewriting operations
type Rewriter struct {
	client *ollama.Client
	config *config.Config
}

// RewriteOption represents a single rewrite option
type RewriteOption struct {
	Style string `json:"style"`
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

// RewriteStyles contains all available rewrite styles
var RewriteStyles = []string{
	"grammar",
	"paraphrase",
	"standard",
	"formal",
	"casual",
	"creative",
	"short",
	"expand",
}

// AnalysisStyles contains all available analysis styles
var AnalysisStyles = []string{
	"summarize",
	"bullets",
	"insights",
}

// StyleInfo contains metadata about each style
var StyleInfo = map[string]struct {
	Label       string
	Icon        string
	Description string
}{
	"grammar":    {"Grammar & Spelling", "🛡️", "Corrects errors and improves flow"},
	"paraphrase": {"Paraphrase", "🔄", "Rewrite with different structure"},
	"standard":   {"Standard", "📝", "Balanced and natural"},
	"formal":     {"Formal", "📢", "Professional tone"},
	"casual":     {"Casual", "💬", "Conversational"},
	"creative":   {"Creative", "✨", "Expressive and vivid"},
	"short":      {"Short", "📏", "Concise and brief"},
	"expand":     {"Expand", "📖", "More detail and depth"},
	// Analysis styles
	"summarize": {"TL;DR Summary", "📋", "Concise 2-3 sentence summary"},
	"bullets":   {"Key Points", "•••", "Extract 3-5 main bullet points"},
	"insights":  {"Key Insights", "💡", "Important facts and arguments"},
}

// New creates a new Rewriter instance
func New(client *ollama.Client, cfg *config.Config) *Rewriter {
	return &Rewriter{
		client: client,
		config: cfg,
	}
}

// GenerateRewrites is deprecated. Use GenerateSingleRewrite or GenerateSingleRewriteWithFormatting instead.
// This method generated all styles simultaneously which was inefficient.
// Kept for backward compatibility but will be removed in a future version.

// GenerateSingleRewrite generates a rewrite for a specific style
func (r *Rewriter) GenerateSingleRewrite(ctx context.Context, text, style string) (RewriteOption, error) {
	if !isValidStyle(style) {
		return RewriteOption{}, fmt.Errorf("invalid style: %s", style)
	}

	systemPrompt := r.config.GetPrompt(style, "normal")
	rewritten, err := r.client.GenerateRewrite(ctx, text, style, systemPrompt)

	if err != nil {
		return RewriteOption{
			Style: style,
			Text:  "",
			Error: err.Error(),
		}, err
	}

	return RewriteOption{
		Style: style,
		Text:  cleanResponse(rewritten),
	}, nil
}

// GenerateAnalysis is deprecated. Use GenerateSingleAnalysis or GenerateSingleAnalysisWithFormatting instead.
// This method generated all analysis styles simultaneously which was inefficient.
// Kept for backward compatibility but will be removed in a future version.

// GenerateSingleAnalysis generates an analysis for a specific style
func (r *Rewriter) GenerateSingleAnalysis(ctx context.Context, text, style string) (RewriteOption, error) {
	if !isValidAnalysisStyle(style) {
		return RewriteOption{}, fmt.Errorf("invalid analysis style: %s", style)
	}

	systemPrompt := r.config.GetPrompt(style, "normal")
	rewritten, err := r.client.GenerateRewrite(ctx, text, style, systemPrompt)

	if err != nil {
		return RewriteOption{
			Style: style,
			Text:  "",
			Error: err.Error(),
		}, err
	}

	return RewriteOption{
		Style: style,
		Text:  cleanResponse(rewritten),
	}, nil
}

// GenerateSingleRewriteWithFormatting generates a rewrite with optional formatting
func (r *Rewriter) GenerateSingleRewriteWithFormatting(ctx context.Context, text, style string, enableFormatting bool) (RewriteOption, error) {
	if !isValidStyle(style) {
		return RewriteOption{}, fmt.Errorf("invalid style: %s", style)
	}

	var systemPrompt string
	if enableFormatting {
		systemPrompt = r.config.GetPrompt(style, "normal")
	} else {
		// Check for custom prompt first, fall back to plain text prompt
		if customPrompt := r.config.GetCustomPrompt(style, "normal"); customPrompt != "" {
			systemPrompt = customPrompt
		} else {
			systemPrompt = getPlainTextPrompt(style)
		}
	}

	rewritten, err := r.client.GenerateRewrite(ctx, text, style, systemPrompt)

	if err != nil {
		return RewriteOption{
			Style: style,
			Text:  "",
			Error: err.Error(),
		}, err
	}

	result := cleanResponse(rewritten)

	// If formatting is disabled, also strip any markdown that might have been added
	if !enableFormatting {
		result = stripMarkdownFormatting(result)
	}

	return RewriteOption{
		Style: style,
		Text:  result,
	}, nil
}

// GenerateSingleAnalysisWithFormatting generates an analysis with optional formatting
func (r *Rewriter) GenerateSingleAnalysisWithFormatting(ctx context.Context, text, style string, enableFormatting bool) (RewriteOption, error) {
	if !isValidAnalysisStyle(style) {
		return RewriteOption{}, fmt.Errorf("invalid analysis style: %s", style)
	}

	var systemPrompt string
	if enableFormatting {
		systemPrompt = r.config.GetPrompt(style, "normal")
	} else {
		// Check for custom prompt first, fall back to plain text prompt
		if customPrompt := r.config.GetCustomPrompt(style, "normal"); customPrompt != "" {
			systemPrompt = customPrompt
		} else {
			systemPrompt = getPlainTextPrompt(style)
		}
	}

	rewritten, err := r.client.GenerateRewrite(ctx, text, style, systemPrompt)

	if err != nil {
		return RewriteOption{
			Style: style,
			Text:  "",
			Error: err.Error(),
		}, err
	}

	result := cleanResponse(rewritten)

	// If formatting is disabled, also strip any markdown that might have been added
	if !enableFormatting {
		result = stripMarkdownFormatting(result)
	}

	return RewriteOption{
		Style: style,
		Text:  result,
	}, nil
}

// GenerateRewriteWithTextType generates a rewrite for a specific style and text type
func (r *Rewriter) GenerateRewriteWithTextType(ctx context.Context, text, style string, textType TextType, enableFormatting bool) (RewriteOption, error) {
	if !isValidStyle(style) {
		return RewriteOption{}, fmt.Errorf("invalid style: %s", style)
	}

	systemPrompt := r.getPromptForTextType(style, textType, enableFormatting)
	rewritten, err := r.client.GenerateRewrite(ctx, text, style, systemPrompt)

	if err != nil {
		return RewriteOption{
			Style: style,
			Text:  "",
			Error: err.Error(),
		}, err
	}

	result := cleanResponse(rewritten)

	// If formatting is disabled, also strip any markdown that might have been added
	if !enableFormatting {
		result = stripMarkdownFormatting(result)
	}

	return RewriteOption{
		Style: style,
		Text:  result,
	}, nil
}

// GenerateAnalysisWithTextType generates an analysis for a specific style and text type
func (r *Rewriter) GenerateAnalysisWithTextType(ctx context.Context, text, style string, textType TextType, enableFormatting bool) (RewriteOption, error) {
	if !isValidAnalysisStyle(style) {
		return RewriteOption{}, fmt.Errorf("invalid analysis style: %s", style)
	}

	systemPrompt := r.getPromptForTextType(style, textType, enableFormatting)
	rewritten, err := r.client.GenerateRewrite(ctx, text, style, systemPrompt)

	if err != nil {
		return RewriteOption{
			Style: style,
			Text:  "",
			Error: err.Error(),
		}, err
	}

	result := cleanResponse(rewritten)

	// If formatting is disabled, also strip any markdown that might have been added
	if !enableFormatting {
		result = stripMarkdownFormatting(result)
	}

	return RewriteOption{
		Style: style,
		Text:  result,
	}, nil
}

// cleanResponse sanitizes the AI response by removing thinking tags,
// markdown code blocks, conversational fillers, and extra whitespace.
func cleanResponse(text string) string {
	// 0. Remove <system-reminder>...</system-reminder> blocks entirely
	for {
		start := strings.Index(text, "<system-reminder>")
		if start == -1 {
			break
		}
		end := strings.Index(text[start+len("<system-reminder>"):], "</system-reminder>")
		if end == -1 {
			text = text[:start]
			break
		}
		// Remove the entire block including tags
		blockEnd := start + len("<system-reminder>") + end + len("</system-reminder>")
		text = text[:start] + text[blockEnd:]
	}

	// 1. Remove <input>...</input> blocks - extract content between tags
	for {
		start := strings.Index(text, "<input>")
		if start == -1 {
			break
		}
		end := strings.Index(text[start+len("<input>"):], "</input>")
		if end == -1 {
			text = text[:start]
			break
		}
		text = text[start+len("<input>") : start+len("<input>")+end]
	}

	// 2. Remove <output>...</output> blocks - extract content between tags
	for {
		start := strings.Index(text, "<output>")
		if start == -1 {
			break
		}
		end := strings.Index(text[start+len("<output>"):], "</output>")
		if end == -1 {
			text = text[:start]
			break
		}
		text = text[start+len("<output>") : start+len("<output>")+end]
	}

	// 3. Remove XML-like tags but KEEP the content inside them
	// This handles cases where AI wraps content in <email>, <greeting>, <body>, etc.
	// We strip the tags but preserve the text between them
	for {
		start := strings.Index(text, "<")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		// Only remove if it looks like an XML tag (not markdown like **bold** or *italic*)
		tagContent := text[start : start+end+1]
		isXMLTag := strings.HasPrefix(tagContent, "</") ||
			(len(tagContent) > 2 &&
				tagContent[1] != ' ' &&
				tagContent[1] != '*' &&
				tagContent[1] != '#' &&
				tagContent[1] != '-' &&
				tagContent[1] != '_' &&
				tagContent[1] != '`' &&
				tagContent[1] != '~' &&
				tagContent[1] != '[')

		if isXMLTag {
			text = text[:start] + text[start+end+1:]
		} else {
			break
		}
	}

	// 2. Remove markdown code blocks (```...```)
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		newlineIndex := strings.Index(text, "\n")
		if newlineIndex != -1 {
			text = text[newlineIndex+1:]
		} else {
			text = strings.TrimPrefix(text, "```")
		}
		if strings.HasSuffix(strings.TrimSpace(text), "```") {
			text = strings.TrimSpace(text)
			text = strings.TrimSuffix(text, "```")
		}
	}

	// 3. Remove conversational prefixes
	prefixes := []string{
		"here is the rewritten text:",
		"here is the requested text:",
		"here is the text:",
		"sure, here is",
		"sure!",
		"sure,",
		"okay,",
		"okay!",
		"rewrite:",
		"rewritten text:",
		"answer:",
		"here you go:",
		"here is",
		"i've rewritten",
		"i have rewritten",
		"below is",
		"the rewritten text is:",
		"result:",
		"output:",
		"i've analyzed",
		"i have analyzed",
	}

	lowerText := strings.ToLower(text)
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.TrimSpace(lowerText), prefix) {
			idx := len(prefix)
			if len(text) >= idx {
				text = text[idx:]
				lowerText = strings.ToLower(text)
			}
		}
	}

	// 4. Remove wrapping quotes
	text = strings.TrimSpace(text)
	if len(text) >= 2 && strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
		text = text[1 : len(text)-1]
	}
	if len(text) >= 2 && strings.HasPrefix(text, "'") && strings.HasSuffix(text, "'") {
		text = text[1 : len(text)-1]
	}

	// 5. Remove common suffixes
	suffixes := []string{
		"let me know if you need anything else.",
		"let me know if you have any questions.",
		"hope this helps!",
		"hope this helps.",
		"is there anything else you'd like me to help with?",
		"feel free to ask if you need more assistance.",
		"does this meet your needs?",
		"would you like me to change anything?",
		"please let me know if you'd like any adjustments.",
	}

	lowerText = strings.ToLower(text)
	for _, suffix := range suffixes {
		if strings.HasSuffix(strings.TrimSpace(lowerText), suffix) {
			idx := len(text) - len(suffix)
			if idx > 0 {
				text = text[:idx]
				lowerText = strings.ToLower(text)
			}
		}
	}

	// 6. Final cleanup
	return strings.TrimSpace(text)
}

// stripMarkdownFormatting removes markdown bold markers and normalizes spacing
// Kept for formatting toggle logic, but cleaner response handling is preferred
func stripMarkdownFormatting(text string) string {
	// Remove **bold** markers
	result := strings.ReplaceAll(text, "**", "")

	// Normalize multiple line breaks to single line breaks
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}

	return strings.TrimSpace(result)
}

// getPlainTextPrompt returns a simplified prompt without formatting instructions (generic)
func getPlainTextPrompt(style string) string {
	plainPrompts := map[string]string{
		"grammar": `You are an expert editor and proofreader with exceptional attention to detail.

TASK: Fix all grammar, spelling, punctuation, and awkward phrasing while preserving the original meaning and structure.

RULES:
- Analyze the input type and adapt.
- For CODE: Only fix comments.
- Improve flow and correctness.
- Never change code logic or syntax.
OUTPUT: Return ONLY the corrected text. Nothing before or after.`,

		"paraphrase": `You are an expert writer.

TASK: Rewrite the text using different words/structure but keep the meaning.

RULES:
- Analyze the input type.
- Keep the original tone.
- For CODE: Comments only.
OUTPUT: Return ONLY the rewritten text. Nothing before or after.`,

		"standard": `You are a professional writer.

TASK: Rewrite the text to be clear, natural, and balanced.

RULES:
- Improve clarity and flow.
- Remove awkward phrasing.
OUTPUT: Return ONLY the rewritten text. Nothing before or after.`,

		"formal": `You are a professional communication expert.

TASK: Rewrite the text to be formal and polite.

RULES:
- Use precise, formal language.
- Avoid contractions/slang.
OUTPUT: Return ONLY the formal text. Nothing before or after.`,

		"casual": `You are a friendly, casual writer.

TASK: Rewrite the text to be friendly and conversational.

RULES:
- Use natural language/contractions.
- Make it sound like a friend.
OUTPUT: Return ONLY the casual text. Nothing before or after.`,

		"creative": `You are a creative writer.

TASK: Rewrite the text to be expressive and vivid.

RULES:
- Use strong verbs and evocative language.
OUTPUT: Return ONLY the creative text. Nothing before or after.`,

		"short": `You are a concise editor.

TASK: Shorten the text by removing unnecessary words.

RULES:
- Keep the core message.
- Make it punchy.
OUTPUT: Return ONLY the shortened text. Nothing before or after.`,

		"expand": `You are an expert writer.

TASK: Expand the text by adding details and context.

RULES:
- Elaborate on key points.
- Make it more comprehensive.
OUTPUT: Return ONLY the expanded text. Nothing before or after.`,

		"summarize": `You are a summarizer.

TASK: Provide a concise summary.

RULES:
- Identify main points.
- Condense into a brief overview.
OUTPUT: Return ONLY the summary. Nothing before or after.`,

		"bullets": `You are an analyst.

TASK: Extract key points as a bullet list.

RULES:
- Identify important info.
- Format as clean bullets.
OUTPUT: Return ONLY the bullet list. Nothing before or after.`,

		"insights": `You are a strategic analyst.

TASK: Extract key insights and implications.

RULES:
- Identify the "so what?".
- Analyze tone and intent.
OUTPUT: Return ONLY the insights. Nothing before or after.`,
	}

	if prompt, ok := plainPrompts[style]; ok {
		return prompt
	}
	return "Rewrite the text appropriately. Return only the rewritten text."
}

// getPromptForTextType returns a prompt customized for the specific text type
func (r *Rewriter) getPromptForTextType(style string, textType TextType, enableFormatting bool) string {
	textTypeStr := string(textType)

	if customPrompt := r.config.GetCustomPrompt(style, textTypeStr); customPrompt != "" {
		return customPrompt
	}

	if !enableFormatting {
		return getPlainTextPrompt(style)
	}

	typeInstructions := map[TextType]map[string]string{
		TextTypeEmail: {
			"grammar":    "You are an expert editor specializing in professional email communication.\n\nTASK: Fix all grammar, spelling, punctuation, and awkward phrasing in this email while preserving the original meaning, intent, and structure.\n\nRULES:\n- ONLY fix errors in the text provided - do not add or remove content\n- Preserve the email structure: greeting, body paragraphs, and sign-off\n- If the input lacks a proper greeting or sign-off, add appropriate ones based on context\n- Use **bold** for key terms and important information\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the corrected email as plain text. Nothing before or after.",
			"paraphrase": "You are an expert writer specializing in professional communication.\n\nTASK: Rewrite this email using different words and sentence structures while preserving the exact same meaning and intent.\n\nRULES:\n- Use varied vocabulary and restructured sentences\n- Keep all original information - do not add or remove anything\n- Preserve the email structure: greeting, body paragraphs, and sign-off\n- If the input lacks a proper greeting or sign-off, add appropriate ones based on context\n- Use **bold** for key terms\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the rewritten email as plain text. Nothing before or after.",
			"standard":   "You are a professional writer specializing in clear, effective communication.\n\nTASK: Rewrite this email to be clear, natural, and well-structured while preserving the original meaning.\n\nRULES:\n- Improve clarity, flow, and readability\n- Keep all original information - do not add or remove anything\n- Preserve the email structure: greeting, body paragraphs, and sign-off\n- If the input lacks a proper greeting or sign-off, add appropriate ones based on context\n- Use **bold** for key terms and important points\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the rewritten email as plain text. Nothing before or after.",
			"formal":     "You are a business communication expert specializing in formal correspondence.\n\nTASK: Rewrite this email in a highly formal, professional tone suitable for official communication.\n\nRULES:\n- Use formal greetings (e.g., Dear [Name],) and closings (e.g., Sincerely,)\n- Replace contractions with full forms\n- Use precise, elevated vocabulary\n- Keep all original information - do not add or remove anything\n- If the input lacks a proper greeting or sign-off, add formal ones based on context\n- Use **bold** for key terms and important references\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the formal email as plain text. Nothing before or after.",
			"casual":     "You are a friendly writer who excels at warm, approachable communication.\n\nTASK: Rewrite this email in a warm, casual, and conversational tone while keeping it respectful.\n\nRULES:\n- Use casual greetings (e.g., Hi [Name],) and warm closings (e.g., Best,)\n- Use contractions and natural conversational language\n- Keep all original information - do not add or remove anything\n- If the input lacks a proper greeting or sign-off, add casual ones based on context\n- Use **bold** for key points\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the casual email as plain text. Nothing before or after.",
			"creative":   "You are a creative writer who makes emails engaging and memorable.\n\nTASK: Rewrite this email to be expressive and vivid while maintaining the core message.\n\nRULES:\n- Use expressive language and vivid descriptions\n- Keep all original information - do not add or remove anything\n- Preserve the email structure: greeting, body paragraphs, and sign-off\n- If the input lacks a proper greeting or sign-off, add engaging ones based on context\n- Use **bold** for emphasis and key moments\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the creative email as plain text. Nothing before or after.",
			"short":      "You are a concise editor who specializes in tight, efficient writing.\n\nTASK: Shorten this email by removing unnecessary words while preserving ALL key information.\n\nRULES:\n- Remove redundancy and filler words only\n- Keep all factual information, requests, and important details\n- Preserve the email structure: greeting, body, and sign-off\n- If the input lacks a proper greeting or sign-off, add brief ones based on context\n- Use **bold** for critical information\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the shortened email as plain text. Nothing before or after.",
			"expand":     "You are an expert writer who adds valuable context and detail.\n\nTASK: Expand this email by adding relevant context, elaboration, and helpful detail.\n\nRULES:\n- Add useful context and supporting detail that relates to the original content\n- Do not invent facts, names, or specific details\n- Preserve the email structure: greeting, body paragraphs, and sign-off\n- If the input lacks a proper greeting or sign-off, add appropriate ones based on context\n- Use **bold** for key terms\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the expanded email as plain text. Nothing before or after.",
			"summarize":  "You are a strategic analyst who distills complex information into clear summaries.\n\nTASK: Summarize this email, identifying the purpose, key points, and any required actions.\n\nRULES:\n- Identify the email's primary purpose\n- Highlight action items, deadlines, or decisions needed\n- Keep it to 2-4 sentences maximum\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the summary as plain text. Nothing before or after.",
			"bullets":    "You are an analyst who extracts and organizes key information.\n\nTASK: Extract the key points from this email as a clear, organized bullet list.\n\nRULES:\n- List purpose, requests, deadlines, and action items\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\n- NEVER invent information not present in the original\nOUTPUT: Return ONLY the bullet list as plain text. Nothing before or after.",
			"insights":   "You are a strategic communication analyst who reads between the lines.\n\nTASK: Analyze this email for insights beyond the surface message.\n\nRULES:\n- Identify intent, tone, and implicit requests\n- Assess relationship dynamics\n- NEVER use XML tags, HTML, or any markup in your response\n- NEVER add conversational filler, explanations, or placeholder text\nOUTPUT: Return ONLY the analysis as plain text. Nothing before or after.",
		},
		TextTypeChat: {
			"grammar":    "You are an expert editor specializing in conversational communication.\n\nTASK: Fix grammar, spelling, and punctuation in this chat message while preserving its natural, conversational tone.\n\nRULES:\n- Keep the casual, conversational tone\n- Preserve emojis and slang if appropriate\n- Use **bold** for key points\n- NEVER add conversational filler, explanations, or new content\n- NEVER invent information not present in the original\nOUTPUT: Return ONLY the corrected message. Nothing before or after.",
			"paraphrase": "You are a writer who excels at natural, conversational communication.\n\nTASK: Rewrite this chat message using different words while keeping the same meaning and conversational vibe.\n\nRULES:\n- Keep the casual, friendly feel\n- Preserve any humor or personality\n- Use **bold** for emphasis\n- NEVER add conversational filler, explanations, or new content\n- NEVER invent information not present in the original\nOUTPUT: Return ONLY the rewritten message. Nothing before or after.",
			"standard":   "You are a writer who makes chat messages clear and natural.\n\nTASK: Rewrite this chat message to be clear and natural while keeping it conversational.\n\nRULES:\n- Make it sound like a natural conversation\n- Keep the casual, friendly tone\n- Use **bold** for key points\n- NEVER add conversational filler, explanations, or new content\n- NEVER invent information not present in the original\nOUTPUT: Return ONLY the rewritten message. Nothing before or after.",
			"formal":     "You are a professional who knows how to communicate politely and clearly.\n\nTASK: Rewrite this chat message in a more professional, polished tone.\n\nRULES:\n- Remove slang and overly casual language\n- Use polite, respectful language\n- Use **bold** for important details\n- NEVER add conversational filler, explanations, or new content\n- NEVER invent information not present in the original\nOUTPUT: Return ONLY the formal message. Nothing before or after.",
			"casual":     "You are a friend who writes naturally and casually.\n\nTASK: Rewrite this chat message to sound super casual and relaxed.\n\nRULES:\n- Use slang, contractions, and natural chat speak\n- Sound relaxed and authentic\n- Use **bold** for emphasis\n- NEVER add conversational filler, explanations, or new content\n- NEVER invent information not present in the original\nOUTPUT: Return ONLY the casual message. Nothing before or after.",
			"creative":   "You are a creative writer who brings personality and flair to messages.\n\nTASK: Rewrite this chat message to be fun, expressive, and full of personality.\n\nRULES:\n- Show personality, humor, or creativity\n- Use **bold** for punchy moments\n- NEVER add conversational filler, explanations, or new content\n- NEVER invent information not present in the original\nOUTPUT: Return ONLY the creative message. Nothing before or after.",
			"short":      "You are a concise editor who makes messages brief and punchy.\n\nTASK: Shorten this chat message to be as brief and direct as possible.\n\nRULES:\n- Cut filler words and redundancy\n- Get straight to the point\n- Use **bold** for the most important part\n- NEVER add conversational filler, explanations, or new content\n- NEVER invent information not present in the original\nOUTPUT: Return ONLY the shortened message. Nothing before or after.",
			"expand":     "You are a writer who adds helpful context to messages.\n\nTASK: Expand this chat message by adding relevant context without over-explaining.\n\nRULES:\n- Add useful context and detail\n- Keep it conversational\n- Use **bold** for key information\n- NEVER add conversational filler, explanations, or new content\nOUTPUT: Return ONLY the expanded message. Nothing before or after.",
			"summarize":  "You are an analyst who distills conversations into clear takeaways.\n\nTASK: Summarize this chat conversation, identifying key decisions and topics.\n\nRULES:\n- Identify key decisions and topics\n- Keep it to 2-4 sentences\n- Use **bold** for key takeaways\n- NEVER add conversational filler, explanations, or new content\nOUTPUT: Return ONLY the summary. Nothing before or after.",
			"bullets":    "You are an analyst who extracts key points from conversations.\n\nTASK: Extract the key points from this chat as a bullet list.\n\nRULES:\n- Extract decisions and action items\n- Use **bold** for critical details\n- NEVER add conversational filler, explanations, or new content\nOUTPUT: Return ONLY the bullet list. Nothing before or after.",
			"insights":   "You are a communication analyst who reads between the lines.\n\nTASK: Analyze this chat for insights - identify sentiment and key takeaways.\n\nRULES:\n- Identify sentiment and key takeaways\n- Use **bold** for important observations\n- NEVER add conversational filler, explanations, or new content\nOUTPUT: Return ONLY the analysis. Nothing before or after.",
		},
		TextTypeCode: {
			"grammar":    "You are a technical editor who specializes in code documentation.\n\nTASK: Fix grammar, spelling, and punctuation ONLY in comments and documentation.\n\nRULES:\n- DO NOT CHANGE CODE LOGIC OR SYNTAX\n- Only modify English text in comments\n- Use **bold** for important warnings or notes in comments\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the code with corrected comments. Nothing before or after.",
			"paraphrase": "You are a technical writer who rewrites code documentation.\n\nTASK: Rewrite the comments using different words while keeping the same technical meaning.\n\nRULES:\n- DO NOT CHANGE CODE LOGIC\n- Keep code exactly as is\n- Use **bold** for important technical terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the code with rewritten comments. Nothing before or after.",
			"standard":   "You are a technical editor who improves code documentation clarity.\n\nTASK: Improve the clarity and readability of comments and documentation.\n\nRULES:\n- DO NOT CHANGE CODE LOGIC\n- Make comments clear and concise\n- Use **bold** for important notes\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the code with improved comments. Nothing before or after.",
			"formal":     "You are a technical documentation specialist.\n\nTASK: Rewrite the comments to be formal, precise, and professional.\n\nRULES:\n- DO NOT CHANGE CODE LOGIC\n- Use standard technical documentation style\n- Use **bold** for technical terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the code with formal comments. Nothing before or after.",
			"casual":     "You are a developer buddy who writes helpful, friendly comments.\n\nTASK: Rewrite the comments to be helpful and conversational.\n\nRULES:\n- DO NOT CHANGE CODE LOGIC\n- Use a helpful, friendly tone\n- Use **bold** for tips or gotchas\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the code with casual comments. Nothing before or after.",
			"creative":   "You are a creative coder who makes comments engaging.\n\nTASK: Rewrite the comments to be more expressive and vivid.\n\nRULES:\n- DO NOT CHANGE CODE LOGIC\n- Use vivid language in comments\n- Use **bold** for key concepts\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the code with creative comments. Nothing before or after.",
			"short":      "You are a concise coder who values brevity in documentation.\n\nTASK: Shorten the comments to be brief and direct.\n\nRULES:\n- DO NOT CHANGE CODE LOGIC\n- Remove redundant or unnecessary comments\n- Use **bold** for critical warnings\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the code with short comments. Nothing before or after.",
			"expand":     "You are a mentor who writes thorough, educational code documentation.\n\nTASK: Add detailed, explanatory comments to help readers understand the logic.\n\nRULES:\n- DO NOT CHANGE CODE LOGIC\n- Add detailed explanations\n- Use **bold** for important concepts\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the code with expanded comments. Nothing before or after.",
			"summarize":  "You are a tech lead who explains code clearly.\n\nTASK: Summarize what this code does, its purpose, and its key functionality.\n\nRULES:\n- Explain purpose and functionality\n- Use **bold** for key technical terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the summary. Nothing before or after.",
			"bullets":    "You are a tech lead who breaks down code into clear points.\n\nTASK: List the key features, functions, and operations of this code.\n\nRULES:\n- List main functions and operations\n- Use **bold** for technical details\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the bullet list. Nothing before or after.",
			"insights":   "You are a software architect who evaluates code quality and design.\n\nTASK: Analyze this code for architectural insights and design observations.\n\nRULES:\n- Identify patterns, quality, and design choices\n- Use **bold** for key observations\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the analysis. Nothing before or after.",
		},
		TextTypeList: {
			"grammar":    "You are an editor who specializes in list formatting.\n\nTASK: Fix grammar, spelling, and punctuation while preserving the list format.\n\nRULES:\n- Preserve the list format exactly\n- Fix errors in each item\n- Use **bold** for key terms within items\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the corrected list. Nothing before or after.",
			"paraphrase": "You are a writer who rewrites content while preserving structure.\n\nTASK: Rewrite each item using different words while keeping the same meaning.\n\nRULES:\n- Keep the list structure exactly\n- Rephrase each item\n- Use **bold** for key terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the rewritten list. Nothing before or after.",
			"standard":   "You are a writer who makes lists clear and consistent.\n\nTASK: Rewrite this list to be clear, natural, and well-structured.\n\nRULES:\n- Improve clarity and flow\n- Ensure consistent tone\n- Use **bold** for key terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the rewritten list. Nothing before or after.",
			"formal":     "You are a professional who writes precise, formal lists.\n\nTASK: Rewrite this list in a formal, professional tone.\n\nRULES:\n- Use precise, formal language\n- Ensure parallel structure\n- Use **bold** for key terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the formal list. Nothing before or after.",
			"casual":     "You are a friendly writer who makes lists approachable.\n\nTASK: Rewrite this list in a casual, friendly tone.\n\nRULES:\n- Use conversational language\n- Keep it friendly and easy to read\n- Use **bold** for key points\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the casual list. Nothing before or after.",
			"creative":   "You are a creative writer who brings lists to life.\n\nTASK: Rewrite this list to be expressive and engaging.\n\nRULES:\n- Use vivid language and strong verbs\n- Make each item memorable\n- Use **bold** for emphasis\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the creative list. Nothing before or after.",
			"short":      "You are a concise editor who makes lists tight and efficient.\n\nTASK: Shorten each item while preserving all key information.\n\nRULES:\n- Remove unnecessary words\n- Preserve all items\n- Use **bold** for the most important part\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the shortened list. Nothing before or after.",
			"expand":     "You are a writer who adds valuable detail to lists.\n\nTASK: Expand each item with relevant detail and context.\n\nRULES:\n- Add useful context to each item\n- Make items more comprehensive\n- Use **bold** for key terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the expanded list. Nothing before or after.",
			"summarize":  "You are an analyst who distills lists into their core message.\n\nTASK: Summarize the main theme of this list.\n\nRULES:\n- Identify the overarching theme\n- Keep it to 2-4 sentences\n- Use **bold** for the central idea\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the summary. Nothing before or after.",
			"bullets":    "You are an analyst who refines and prioritizes list content.\n\nTASK: Extract and refine the most important points.\n\nRULES:\n- Extract the most important points\n- Use **bold** for critical details\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the refined bullet list. Nothing before or after.",
			"insights":   "You are an analyst who identifies patterns in lists.\n\nTASK: Analyze this list for patterns, themes, and insights.\n\nRULES:\n- Identify patterns and key themes\n- Use **bold** for key insights\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the analysis. Nothing before or after.",
		},
		TextTypeNormal: {
			"grammar":    "You are an expert editor and proofreader with exceptional attention to detail.\n\nTASK: Fix all grammar, spelling, punctuation, and awkward phrasing while preserving the original meaning.\n\nRULES:\n- Improve flow and correctness\n- Keep paragraph structure\n- Use **bold** for key terms and important concepts\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the corrected text. Nothing before or after.",
			"paraphrase": "You are an expert writer who rewrites content with precision.\n\nTASK: Rewrite the text using different words while preserving the exact same meaning.\n\nRULES:\n- Keep the original tone and structure\n- Use varied vocabulary\n- Use **bold** for key terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the rewritten text. Nothing before or after.",
			"standard":   "You are a professional writer who makes text clear and natural.\n\nTASK: Rewrite the text to be clear, natural, and well-structured.\n\nRULES:\n- Improve clarity and flow\n- Remove awkward phrasing\n- Use **bold** for key terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the rewritten text. Nothing before or after.",
			"formal":     "You are a professional communication expert.\n\nTASK: Rewrite the text in a formal, professional tone.\n\nRULES:\n- Use precise, formal language\n- Avoid contractions and slang\n- Use **bold** for key terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the formal text. Nothing before or after.",
			"casual":     "You are a friendly writer who makes text sound natural.\n\nTASK: Rewrite the text in a casual, friendly tone.\n\nRULES:\n- Use conversational language and contractions\n- Sound like a knowledgeable friend\n- Use **bold** for key points\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the casual text. Nothing before or after.",
			"creative":   "You are a creative writer who transforms text into vivid prose.\n\nTASK: Rewrite the text to be expressive and memorable.\n\nRULES:\n- Use strong verbs and vivid imagery\n- Add personality and flair\n- Use **bold** for emphasis\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the creative text. Nothing before or after.",
			"short":      "You are a concise editor who specializes in tight writing.\n\nTASK: Shorten the text by removing unnecessary words while preserving all key information.\n\nRULES:\n- Remove redundancy and filler\n- Keep the core message\n- Use **bold** for critical information\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the shortened text. Nothing before or after.",
			"expand":     "You are an expert writer who adds valuable context.\n\nTASK: Expand the text by adding relevant detail and context.\n\nRULES:\n- Add useful context and elaboration\n- Maintain the original structure\n- Use **bold** for key terms\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the expanded text. Nothing before or after.",
			"summarize":  "You are a skilled summarizer who distills text into clear overviews.\n\nTASK: Provide a concise summary that captures the main points.\n\nRULES:\n- Identify the central idea and key points\n- Keep it to 2-4 sentences\n- Use **bold** for key takeaways\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the summary. Nothing before or after.",
			"bullets":    "You are an analyst who extracts key information into bullets.\n\nTASK: Convert the text into a bullet list of the most important points.\n\nRULES:\n- Extract 3-7 key ideas\n- Use **bold** for critical details\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the bullet list. Nothing before or after.",
			"insights":   "You are a strategic analyst who identifies non-obvious patterns.\n\nTASK: Analyze the text for key insights, themes, and implications.\n\nRULES:\n- Identify key themes and underlying messages\n- Note implications and significance\n- Use **bold** for key insights\n- NEVER add conversational filler or explanations\nOUTPUT: Return ONLY the analysis. Nothing before or after.",
		},
	}

	if typeMap, ok := typeInstructions[textType]; ok {
		if instruction, ok := typeMap[style]; ok {
			return instruction
		}
	}

	return getPlainTextPrompt(style)
}

// isValidStyle checks if a style name is valid
func isValidStyle(style string) bool {
	for _, s := range RewriteStyles {
		if s == style {
			return true
		}
	}
	return false
}

// isValidAnalysisStyle checks if an analysis style name is valid
func isValidAnalysisStyle(style string) bool {
	for _, s := range AnalysisStyles {
		if s == style {
			return true
		}
	}
	return false
}

// GetStyleInfo returns information about a specific style
func GetStyleInfo(style string) (struct {
	Label       string
	Icon        string
	Description string
}, bool) {
	info, ok := StyleInfo[style]
	return info, ok
}
