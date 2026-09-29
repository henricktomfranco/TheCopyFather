package guardrails

import (
	"strings"
	"unicode"
)

// RepetitionResult contains the result of repetition analysis
type RepetitionResult struct {
	HasLoop       bool
	CleanText     string
	LoopSnippet   string
	LoopType      string // "char", "word", "phrase", "sentence"
}

// wordToken holds a word and its character span in the original text
type wordToken struct {
	normalized string
	start      int
	end        int
}

// tokenizeWords extracts words with their exact character offsets in the input text
func tokenizeWords(text string) []wordToken {
	var tokens []wordToken
	inWord := false
	start := 0

	for i, r := range text {
		if !unicode.IsSpace(r) {
			if !inWord {
				inWord = true
				start = i
			}
		} else {
			if inWord {
				raw := text[start:i]
				norm := normalizeToken(raw)
				if norm != "" {
					tokens = append(tokens, wordToken{
						normalized: norm,
						start:      start,
						end:        i,
					})
				}
				inWord = false
			}
		}
	}
	if inWord {
		raw := text[start:]
		norm := normalizeToken(raw)
		if norm != "" {
			tokens = append(tokens, wordToken{
				normalized: norm,
				start:      start,
				end:        len(text),
			})
		}
	}

	return tokens
}

// normalizeToken strips surrounding punctuation and converts to lowercase
func normalizeToken(s string) string {
	s = strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r)
	})
	return strings.ToLower(s)
}

// DetectAndCleanRepetition scans the text for degeneration loops.
// If a loop is found, HasLoop is true, and CleanText contains the text with the loop truncated/collapsed.
func DetectAndCleanRepetition(text string) RepetitionResult {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) < 6 {
		return RepetitionResult{HasLoop: false, CleanText: text}
	}

	// 1. Check for character runaway (e.g. "..........", "-------", "aaaaaaa")
	if res, found := checkCharRunaway(text); found {
		return res
	}

	// 2. Tokenize words with character offsets
	tokens := tokenizeWords(text)
	if len(tokens) < 3 {
		return RepetitionResult{HasLoop: false, CleanText: text}
	}

	// 3. Check for multi-word phrase cycles (N-grams of length 2..12)
	if res, found := checkPhraseRepetition(text, tokens); found {
		return res
	}

	// 4. Check for single-word repetition (e.g. "word word word word")
	if res, found := checkWordRepetition(text, tokens); found {
		return res
	}

	// 5. Check for sentence / line repetition
	if res, found := checkLineRepetition(text); found {
		return res
	}

	return RepetitionResult{HasLoop: false, CleanText: text}
}

// checkCharRunaway detects excessive repetition of single characters (e.g., punctuation or letters)
func checkCharRunaway(text string) (RepetitionResult, bool) {
	// Look for runaway identical characters (> 7 for punctuation, > 5 for letters)
	runes := []rune(text)
	n := len(runes)
	if n < 6 {
		return RepetitionResult{}, false
	}

	for i := 0; i < n; {
		r := runes[i]
		if unicode.IsSpace(r) {
			i++
			continue
		}

		j := i + 1
		for j < n && runes[j] == r {
			j++
		}
		count := j - i

		isPunct := unicode.IsPunct(r) || unicode.IsSymbol(r)
		isLetter := unicode.IsLetter(r)

		if (isPunct && count >= 8) || (isLetter && count >= 6) {
			// Found runaway char sequence
			maxKeep := 1
			if isPunct {
				maxKeep = 3
			}
			replacement := strings.Repeat(string(r), maxKeep)
			clean := string(runes[:i]) + replacement + string(runes[j:])
			return RepetitionResult{
				HasLoop:     true,
				CleanText:   clean,
				LoopSnippet: string(runes[i:j]),
				LoopType:    "char",
			}, true
		}

		i = j
	}

	return RepetitionResult{}, false
}

// checkWordRepetition detects single word repeated consecutively
func checkWordRepetition(text string, tokens []wordToken) (RepetitionResult, bool) {
	n := len(tokens)
	if n < 3 {
		return RepetitionResult{}, false
	}

	// Check from the end first (streaming degeneration almost always happens at the tail)
	for i := 0; i < n; i++ {
		norm := tokens[i].normalized
		if norm == "" {
			continue
		}

		// Count consecutive identical normalized words
		count := 1
		j := i + 1
		for j < n && tokens[j].normalized == norm {
			count++
			j++
		}

		// Threshold:
		// Length > 2: 3 consecutive identical words (e.g. "please please please")
		// Length <= 2: 4 consecutive identical words (e.g. "no no no no")
		minThreshold := 3
		if len(norm) <= 2 {
			minThreshold = 4
		}

		if count >= minThreshold {
			// Found word repetition loop!
			// We keep the first occurrence (tokens[i]) and cut out tokens[i+1...j-1]
			keepEnd := tokens[i].end
			cutEnd := tokens[j-1].end

			// If the loop continues to the end of the text, truncate text at keepEnd
			var clean string
			if j >= n {
				clean = strings.TrimRight(text[:keepEnd], " \t\r\n")
			} else {
				// Mid-text repetition: splice out repeats
				clean = text[:keepEnd] + text[cutEnd:]
			}

			return RepetitionResult{
				HasLoop:     true,
				CleanText:   clean,
				LoopSnippet: text[tokens[i+1].start:cutEnd],
				LoopType:    "word",
			}, true
		}
	}

	return RepetitionResult{}, false
}

// checkPhraseRepetition checks for consecutive repeating phrases of length 2 to 12 words
func checkPhraseRepetition(text string, tokens []wordToken) (RepetitionResult, bool) {
	n := len(tokens)
	if n < 4 {
		return RepetitionResult{}, false
	}

	// Try phrase lengths from 2 to 12
	maxPhraseLen := 12
	if maxPhraseLen > n/2 {
		maxPhraseLen = n / 2
	}

	// Prioritize tail check for streaming speed
	// Tail check: does tokens end with P P or P P P?
	for pLen := 2; pLen <= maxPhraseLen; pLen++ {
		// Minimum repetitions required to flag as a degenerate loop:
		// For 2-word phrase: 3 repetitions ("thank you thank you thank you")
		// For 3+ word phrase: 2 repetitions ("please let me know please let me know")
		minReps := 2
		if pLen == 2 {
			minReps = 3
		}

		totalTokensNeeded := pLen * minReps
		if n < totalTokensNeeded {
			continue
		}

		// Check at the tail
		tailStart := n - totalTokensNeeded
		isTailLoop := true
		for r := 1; r < minReps; r++ {
			for k := 0; k < pLen; k++ {
				baseWord := tokens[tailStart+k].normalized
				repWord := tokens[tailStart+r*pLen+k].normalized
				if baseWord != repWord {
					isTailLoop = false
					break
				}
			}
			if !isTailLoop {
				break
			}
		}

		if isTailLoop {
			// We have a loop at the tail!
			// Find how many total consecutive repetitions exist ending at n
			// First occurrence starts at tailStart (or earlier)
			firstOccurEnd := tokens[tailStart+pLen-1].end
			clean := strings.TrimRight(text[:firstOccurEnd], " \t\r\n")
			loopStart := tokens[tailStart+pLen].start

			return RepetitionResult{
				HasLoop:     true,
				CleanText:   clean,
				LoopSnippet: text[loopStart:],
				LoopType:    "phrase",
			}, true
		}
	}

	// Full scan for mid-text phrase repetitions (pLen 2..8)
	for pLen := 2; pLen <= 8 && pLen <= n/2; pLen++ {
		minReps := 2
		if pLen == 2 {
			minReps = 3
		}

		for i := 0; i+pLen*minReps <= n; i++ {
			isLoop := true
			for r := 1; r < minReps; r++ {
				for k := 0; k < pLen; k++ {
					if tokens[i+k].normalized != tokens[i+r*pLen+k].normalized {
						isLoop = false
						break
					}
				}
				if !isLoop {
					break
				}
			}

			if isLoop {
				// Count total consecutive repetitions
				reps := minReps
				for i+(reps+1)*pLen <= n {
					matches := true
					for k := 0; k < pLen; k++ {
						if tokens[i+k].normalized != tokens[i+reps*pLen+k].normalized {
							matches = false
							break
						}
					}
					if !matches {
						break
					}
					reps++
				}

				firstOccurEnd := tokens[i+pLen-1].end
				cutEnd := tokens[i+reps*pLen-1].end
				loopStart := tokens[i+pLen].start

				var clean string
				if i+reps*pLen >= n {
					clean = strings.TrimRight(text[:firstOccurEnd], " \t\r\n")
				} else {
					clean = text[:firstOccurEnd] + text[cutEnd:]
				}

				return RepetitionResult{
					HasLoop:     true,
					CleanText:   clean,
					LoopSnippet: text[loopStart:cutEnd],
					LoopType:    "phrase",
				}, true
			}
		}
	}

	return RepetitionResult{}, false
}

// checkLineRepetition detects identical consecutive non-empty lines (excluding short markdown markers)
func checkLineRepetition(text string) (RepetitionResult, bool) {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return RepetitionResult{}, false
	}

	for i := 0; i < len(lines)-1; i++ {
		lineA := strings.TrimSpace(lines[i])
		lineB := strings.TrimSpace(lines[i+1])

		// Ignore empty lines or standard markdown bullet prefixes like "-" or "*"
		if len(lineA) < 15 || lineA == "-" || lineA == "*" || lineA == ">" {
			continue
		}

		if strings.EqualFold(lineA, lineB) {
			// Found repeating line!
			// Count consecutive occurrences
			count := 2
			j := i + 2
			for j < len(lines) && strings.EqualFold(strings.TrimSpace(lines[j]), lineA) {
				count++
				j++
			}

			if count >= 2 {
				// Keep lineA, remove the duplicate lines
				keepLines := append([]string{}, lines[:i+1]...)
				if j < len(lines) {
					keepLines = append(keepLines, lines[j:]...)
				}
				clean := strings.Join(keepLines, "\n")

				return RepetitionResult{
					HasLoop:     true,
					CleanText:   clean,
					LoopSnippet: lineB,
					LoopType:    "sentence",
				}, true
			}
		}
	}

	return RepetitionResult{}, false
}
