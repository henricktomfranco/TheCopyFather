package guardrails

import (
	"strings"
)

// GetTemperatureForStyle returns the tuned temperature for a given rewrite style.
// Low temperatures (0.1 - 0.35) drastically reduce hallucinations for factual / editorial tasks.
func GetTemperatureForStyle(style string) float64 {
	switch strings.ToLower(style) {
	case "grammar":
		// Deterministic error fixing - zero creativity / zero hallucination
		return 0.15
	case "short", "summarize", "bullets", "insights":
		// Factual distillation / extraction - very low entropy
		return 0.25
	case "formal", "standard":
		// Professional tone - balanced and faithful
		return 0.35
	case "paraphrase":
		// Varied phrasing, maintaining strict meaning
		return 0.45
	case "expand":
		// Elaboration with controlled creativity
		return 0.50
	case "casual":
		// Conversational, friendly
		return 0.55
	case "creative":
		// Expressive writing
		return 0.70
	case "sliders":
		// Slider default baseline
		return 0.35
	default:
		return 0.35
	}
}

// CalculateMaxTokens bounds generation length based on input size and style.
// This prevents runaway models from spamming thousands of tokens when looping.
func CalculateMaxTokens(inputText, style string) int {
	words := len(strings.Fields(inputText))
	if words == 0 {
		return 512
	}

	var multiplier float64
	switch strings.ToLower(style) {
	case "short", "summarize", "bullets":
		multiplier = 1.5
	case "grammar":
		multiplier = 2.0
	case "expand":
		multiplier = 4.5
	default:
		multiplier = 3.0
	}

	// ~1.3 tokens per word + generous headroom
	tokens := int(float64(words)*1.3*multiplier) + 300
	if tokens < 256 {
		tokens = 256
	}
	// Cap upper limit to 3072 instead of unbounded 4096 to prevent runaway generation
	if tokens > 3072 {
		tokens = 3072
	}
	return tokens
}

// Common generation guardrail parameters
const (
	DefaultFrequencyPenalty = 0.35  // Penalizes repeated tokens based on frequency
	DefaultPresencePenalty  = 0.20  // Encourages token diversity
	DefaultRepeatPenalty    = 1.15  // Llama.cpp / Ollama repeat penalty
	DefaultRepeatLastN      = 64    // Window of tokens to penalize repeats
	DefaultTopP             = 0.90  // Nucleus sampling
)

// GetDefaultStopSequences returns standard stop sequences to cut off runaway loops
func GetDefaultStopSequences() []string {
	return []string{
		"\n\n\n\n",
		"<|im_end|>",
		"<|endoftext|>",
		"<|eot_id|>",
		"<end_of_turn>",
		"### Input:",
		"### Original:",
	}
}

// AntiHallucinationGuardrails provides strict system-level instructions preventing fabricated facts
const AntiHallucinationGuardrails = `
CRITICAL INTEGRITY & ANTI-HALLUCINATION GUARDRAILS:
- STRICT ZERO-FABRICATION: Do NOT invent, assume, fabricate, or extrapolate any names, dates, numbers, URLs, emails, locations, statistics, or background details not explicitly in the input text.
- 100% FAITHFUL REWRITE: Preserve the original meaning and factual substance faithfully. Do not add outside information, speculative claims, or imaginary context.
- TASK BOUNDARY: You are a text rewriter, NOT a conversational partner. Do NOT answer questions inside the input text, do NOT write a reply to the text, and do NOT continue the story. Rewrite ONLY what was given.
- NO REPETITION LOOPS: Never repeat words, phrases, or sentences. Output each thought once and stop immediately.`
