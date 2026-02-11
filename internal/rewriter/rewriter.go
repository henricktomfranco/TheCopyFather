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

	systemPrompt := r.config.GetPrompt(style)
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

	systemPrompt := r.config.GetPrompt(style)
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
		systemPrompt = r.config.GetPrompt(style)
	} else {
		systemPrompt = getPlainTextPrompt(style)
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
		systemPrompt = r.config.GetPrompt(style)
	} else {
		systemPrompt = getPlainTextPrompt(style)
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
	// 0. Remove </input>...</input> blocks - extract content between tags
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

	// 1. Remove </output>...</output> blocks if present
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

	if prompt, ok := plainPrompts[style]; ok {
		return prompt
	}
	return "Rewrite the text appropriately. Return only the rewritten text."
}

// getPromptForTextType returns a prompt customized for the specific text type
func (r *Rewriter) getPromptForTextType(style string, textType TextType, enableFormatting bool) string {
	// Text type specific instructions
	typeInstructions := map[TextType]map[string]string{
		TextTypeEmail: {
			"grammar":    "You are an expert editor.\nTask: Fix grammar, spelling, and punctuation in this EMAIL.\nGuidelines:\n- Preserve greeting, body, and closing.\n- Improve flow and professional tone.\nOutput: Return ONLY the corrected email.",
			"paraphrase": "You are an expert writer.\nTask: Rewrite this EMAIL using different words.\nGuidelines:\n- Keep the original meaning and structure.\n- Maintain professional tone.\nOutput: Return ONLY the rewritten email.",
			"standard":   "You are a professional writer.\nTask: Rewrite this EMAIL to be clear and natural.\nGuidelines:\n- Improve clarity and flow.\n- Maintain professional business tone.\nOutput: Return ONLY the rewritten email.",
			"formal":     "You are a business communication expert.\nTask: Rewrite this EMAIL to be highly formal.\nGuidelines:\n- Use formal greetings (e.g., 'Dear...') and closings (e.g., 'Sincerely').\n- Use precise, professional language.\nOutput: Return ONLY the formal email.",
			"casual":     "You are a friendly writer.\nTask: Rewrite this EMAIL to be warm and casual.\nGuidelines:\n- Use friendly greetings (e.g., 'Hi') and closings.\n- Make it sound approachable but respectful.\nOutput: Return ONLY the casual email.",
			"creative":   "You are a creative writer.\nTask: Rewrite this EMAIL to be engaging and memorable.\nGuidelines:\n- Use expressive language.\n- Keep it appropriate for an email.\nOutput: Return ONLY the creative email.",
			"short":      "You are a concise editor.\nTask: Shorten this EMAIL.\nGuidelines:\n- Remove unnecessary words.\n- Keep the core message, greeting, and closing.\nOutput: Return ONLY the shortened email.",
			"expand":     "You are an expert writer.\nTask: Expand this EMAIL with more detail.\nGuidelines:\n- Add relevant context and elaboration.\n- Maintain professional structure.\nOutput: Return ONLY the expanded email.",
			"summarize":  "You are an analyst.\nTask: Summarize this EMAIL.\nGuidelines:\n- Identify purpose, action items, and deadlines.\n- Be concise.\nOutput: Return ONLY the summary paragraph.",
			"bullets":    "You are an analyst.\nTask: Extract key points from this EMAIL as bullets.\nGuidelines:\n- List purpose, requests, and deadlines.\nOutput: Return ONLY the bullet list.",
			"insights":   "You are a strategic analyst.\nTask: Analyze this EMAIL for insights.\nGuidelines:\n- Identify intent, tone, and implicit requests.\nOutput: Return ONLY the insights.",
		},
		TextTypeChat: {
			"grammar":    "You are an editor.\nTask: Fix grammar/spelling in this CHAT message.\nGuidelines:\n- Keep the casual, conversational tone.\n- Preserve emojis and slang if appropriate.\nOutput: Return ONLY the corrected message.",
			"paraphrase": "You are a writer.\nTask: Rewrite this CHAT message using different words.\nGuidelines:\n- Keep the friendly, conversational vibe.\nOutput: Return ONLY the rewritten message.",
			"standard":   "You are a writer.\nTask: Rewrite this CHAT message to be natural and clear.\nGuidelines:\n- Make it sound like a natural conversation.\nOutput: Return ONLY the rewritten message.",
			"formal":     "You are a professional.\nTask: Rewrite this CHAT message to be professional.\nGuidelines:\n- Remove slang and overly casual language.\n- Make it polite and concise.\nOutput: Return ONLY the formal message.",
			"casual":     "You are a friend.\nTask: Rewrite this CHAT message to be super casual.\nGuidelines:\n- Use slang, contractions, and natural chat speak.\nOutput: Return ONLY the casual message.",
			"creative":   "You are a creative writer.\nTask: Rewrite this CHAT message to be fun and expressive.\nGuidelines:\n- Show personality.\nOutput: Return ONLY the creative message.",
			"short":      "You are a concise editor.\nTask: Shorten this CHAT message.\nGuidelines:\n- Make it brief and punchy.\n- Get straight to the point.\nOutput: Return ONLY the shortened message.",
			"expand":     "You are a writer.\nTask: Add detail to this CHAT message.\nGuidelines:\n- Add context without being too long-winded.\nOutput: Return ONLY the expanded message.",
			"summarize":  "You are an analyst.\nTask: Summarize this CHAT conversation.\nGuidelines:\n- Identify key decisions and topics.\nOutput: Return ONLY the summary.",
			"bullets":    "You are an analyst.\nTask: List key points from this CHAT.\nGuidelines:\n- Extract decisions and action items.\nOutput: Return ONLY the bullet list.",
			"insights":   "You are an analyst.\nTask: Analyze this CHAT for insights.\nGuidelines:\n- Identify sentiment and key takeaways.\nOutput: Return ONLY the insights.",
		},
		TextTypeCode: {
			"grammar":    "You are a tech editor.\nTask: Fix grammar in COMMENTS/DOCS only.\nGuidelines:\n- DO NOT CHANGE CODE LOGIC OR SYNTAX.\n- Only fix English text in comments.\nOutput: Return ONLY the code with corrected comments.",
			"paraphrase": "You are a tech editor.\nTask: Rewrite COMMENTS/DOCS using different words.\nGuidelines:\n- DO NOT CHANGE CODE LOGIC.\n- Keep code exactly as is.\nOutput: Return ONLY the code with rewritten comments.",
			"standard":   "You are a tech editor.\nTask: Improve clarity of COMMENTS/DOCS.\nGuidelines:\n- DO NOT CHANGE CODE LOGIC.\n- Make comments clear and concise.\nOutput: Return ONLY the code with improved comments.",
			"formal":     "You are a technical writer.\nTask: Make COMMENTS/DOCS professional.\nGuidelines:\n- Use standard technical documentation style.\n- DO NOT CHANGE CODE LOGIC.\nOutput: Return ONLY the code with formal comments.",
			"casual":     "You are a developer buddy.\nTask: Make COMMENTS friendly.\nGuidelines:\n- Use a helpful, conversational tone in comments.\n- DO NOT CHANGE CODE LOGIC.\nOutput: Return ONLY the code with casual comments.",
			"creative":   "You are a creative coder.\nTask: Make COMMENTS expressive.\nGuidelines:\n- Use vivid language in comments.\n- DO NOT CHANGE CODE LOGIC.\nOutput: Return ONLY the code with creative comments.",
			"short":      "You are a concise coder.\nTask: Shorten COMMENTS.\nGuidelines:\n- Remove unnecessary words from comments.\n- DO NOT CHANGE CODE LOGIC.\nOutput: Return ONLY the code with short comments.",
			"expand":     "You are a mentor.\nTask: Explain the code in COMMENTS.\nGuidelines:\n- Add detailed explanations to comments.\n- DO NOT CHANGE CODE LOGIC.\nOutput: Return ONLY the code with expanded comments.",
			"summarize":  "You are a tech lead.\nTask: Summarize what this code does.\nGuidelines:\n- Explain purpose and functionality.\nOutput: Return ONLY the summary paragraph.",
			"bullets":    "You are a tech lead.\nTask: List key features of this code.\nGuidelines:\n- Extract main functions and operations.\nOutput: Return ONLY the bullet list.",
			"insights":   "You are a software architect.\nTask: Analyze this code.\nGuidelines:\n- Identify patterns, quality, and design choices.\nOutput: Return ONLY the insights.",
		},
		TextTypeList: {
			"grammar":    "You are an editor.\nTask: Fix grammar in this LIST.\nGuidelines:\n- Preserve list format (bullets/numbers).\n- Fix errors in each item.\nOutput: Return ONLY the corrected list.",
			"paraphrase": "You are a writer.\nTask: Rewrite this LIST using different words.\nGuidelines:\n- Keep the list structure.\n- Rephrase each item.\nOutput: Return ONLY the rewritten list.",
			"standard":   "You are a writer.\nTask: Rewrite this LIST to be clear and natural.\nGuidelines:\n- Improve flow and consistency.\n- Keep list format.\nOutput: Return ONLY the rewritten list.",
			"formal":     "You are a professional.\nTask: Rewrite this LIST to be formal.\nGuidelines:\n- Use precise, professional language.\n- Keep list format.\nOutput: Return ONLY the formal list.",
			"casual":     "You are a friend.\nTask: Rewrite this LIST to be casual.\nGuidelines:\n- Use friendly language.\n- Keep list format.\nOutput: Return ONLY the casual list.",
			"creative":   "You are a creative writer.\nTask: Rewrite this LIST to be expressive.\nGuidelines:\n- Use vivid language.\n- Keep list format.\nOutput: Return ONLY the creative list.",
			"short":      "You are an editor.\nTask: Shorten this LIST.\nGuidelines:\n- Make items concise.\n- Keep list format.\nOutput: Return ONLY the shortened list.",
			"expand":     "You are a writer.\nTask: Expand this LIST.\nGuidelines:\n- Add detail to each item.\n- Keep list format.\nOutput: Return ONLY the expanded list.",
			"summarize":  "You are an analyst.\nTask: Summarize this LIST.\nGuidelines:\n- Condense the main theme into a paragraph.\nOutput: Return ONLY the summary.",
			"bullets":    "You are an analyst.\nTask: Refine this LIST.\nGuidelines:\n- Extract the most important points.\nOutput: Return ONLY the bullet list.",
			"insights":   "You are an analyst.\nTask: Analyze this LIST.\nGuidelines:\n- Identify patterns and key themes.\nOutput: Return ONLY the insights.",
		},
		TextTypeNormal: {
			"grammar":    "You are an editor.\nTask: Fix grammar/spelling in this TEXT.\nGuidelines:\n- Improve flow and correctness.\n- Keep paragraph structure.\nOutput: Return ONLY the corrected text.",
			"paraphrase": "You are a writer.\nTask: Rewrite this TEXT using different words.\nGuidelines:\n- Keep same meaning and tone.\n- Maintain structure.\nOutput: Return ONLY the rewritten text.",
			"standard":   "You are a writer.\nTask: Rewrite this TEXT to be clear and natural.\nGuidelines:\n- Improve clarity and flow.\n- Remove awkward phrasing.\nOutput: Return ONLY the rewritten text.",
			"formal":     "You are a professional.\nTask: Rewrite this TEXT to be formal.\nGuidelines:\n- Use professional, precise language.\n- Avoid contractions.\nOutput: Return ONLY the formal text.",
			"casual":     "You are a friend.\nTask: Rewrite this TEXT to be casual.\nGuidelines:\n- Use conversational language.\n- Make it sound friendly.\nOutput: Return ONLY the casual text.",
			"creative":   "You are a storyteller.\nTask: Rewrite this TEXT to be expressive.\nGuidelines:\n- Use vivid imagery and strong verbs.\nOutput: Return ONLY the creative text.",
			"short":      "You are an editor.\nTask: Shorten this TEXT.\nGuidelines:\n- Remove unnecessary words.\n- Keep key info.\nOutput: Return ONLY the shortened text.",
			"expand":     "You are a writer.\nTask: Expand this TEXT.\nGuidelines:\n- Add detail and context.\n- Elaborate on ideas.\nOutput: Return ONLY the expanded text.",
			"summarize":  "You are a summarizer.\nTask: Summarize this TEXT.\nGuidelines:\n- Condense into a brief overview.\n- Capture main points.\nOutput: Return ONLY the summary.",
			"bullets":    "You are an analyst.\nTask: Convert this TEXT into key points.\nGuidelines:\n- Extract 3-5 main ideas as bullets.\nOutput: Return ONLY the bullet list.",
			"insights":   "You are an analyst.\nTask: Analyze this TEXT.\nGuidelines:\n- Identify key themes and arguments.\nOutput: Return ONLY the insights.",
		},
	}

	// Get instructions for this text type and style
	if typeMap, ok := typeInstructions[textType]; ok {
		if instruction, ok := typeMap[style]; ok {
			return instruction
		}
	}

	// Fallback to generic prompt
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
