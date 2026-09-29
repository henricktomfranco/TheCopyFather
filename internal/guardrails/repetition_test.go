package guardrails

import (
	"testing"
)

func TestDetectAndCleanRepetition_NormalText(t *testing.T) {
	texts := []string{
		"Hello, please make sure you send the file before end of day.",
		"This is a perfectly normal paragraph with rich vocabulary and varying sentence structures.",
		"- Bullet point 1\n- Bullet point 2\n- Bullet point 3",
		"He said that that was true.", // "that that" is only 2 repeats, not a loop
		"No, no, that's not what I meant.", // "no, no" is only 2 repeats
	}

	for _, text := range texts {
		res := DetectAndCleanRepetition(text)
		if res.HasLoop {
			t.Errorf("Expected no loop for normal text: %q, got loop type %s, snippet: %q", text, res.LoopType, res.LoopSnippet)
		}
		if res.CleanText != text {
			t.Errorf("Expected CleanText == original, got: %q", res.CleanText)
		}
	}
}

func TestDetectAndCleanRepetition_WordLoopAtTail(t *testing.T) {
	input := "Please submit your timesheet timesheet timesheet timesheet"
	res := DetectAndCleanRepetition(input)
	if !res.HasLoop {
		t.Fatalf("Expected loop to be detected for %q", input)
	}
	if res.LoopType != "word" {
		t.Errorf("Expected loop type 'word', got %q", res.LoopType)
	}
	expected := "Please submit your timesheet"
	if res.CleanText != expected {
		t.Errorf("Expected clean text %q, got %q", expected, res.CleanText)
	}
}

func TestDetectAndCleanRepetition_WordLoopMidText(t *testing.T) {
	input := "We should definitely definitely definitely meet tomorrow at noon."
	res := DetectAndCleanRepetition(input)
	if !res.HasLoop {
		t.Fatalf("Expected loop to be detected for %q", input)
	}
	expected := "We should definitely meet tomorrow at noon."
	if res.CleanText != expected {
		t.Errorf("Expected clean text %q, got %q", expected, res.CleanText)
	}
}

func TestDetectAndCleanRepetition_ShortWordLoop(t *testing.T) {
	// Short word (<= 2 chars) requires >= 4 repeats
	input := "I need to to to to go there."
	res := DetectAndCleanRepetition(input)
	if !res.HasLoop {
		t.Fatalf("Expected loop for 4 repeats of 'to'")
	}
	expected := "I need to go there."
	if res.CleanText != expected {
		t.Errorf("Expected %q, got %q", expected, res.CleanText)
	}

	// 2 or 3 repeats of 'to' should not be flagged as a loop
	noLoop := "I need to to go there."
	res2 := DetectAndCleanRepetition(noLoop)
	if res2.HasLoop {
		t.Errorf("Did not expect loop for 2 repeats of 'to'")
	}
}

func TestDetectAndCleanRepetition_PhraseLoopAtTail(t *testing.T) {
	input := "The draft is ready. Please let me know your thoughts. Please let me know your thoughts."
	res := DetectAndCleanRepetition(input)
	if !res.HasLoop {
		t.Fatalf("Expected phrase loop to be detected")
	}
	if res.LoopType != "phrase" {
		t.Errorf("Expected loop type 'phrase', got %q", res.LoopType)
	}
	expected := "The draft is ready. Please let me know your thoughts."
	if res.CleanText != expected {
		t.Errorf("Expected clean text %q, got %q", expected, res.CleanText)
	}
}

func TestDetectAndCleanRepetition_TwoWordPhraseLoop(t *testing.T) {
	input := "I appreciate your help. Thank you thank you thank you"
	res := DetectAndCleanRepetition(input)
	if !res.HasLoop {
		t.Fatalf("Expected 2-word phrase loop with 3 repeats")
	}
	expected := "I appreciate your help. Thank you"
	if res.CleanText != expected {
		t.Errorf("Expected %q, got %q", expected, res.CleanText)
	}
}

func TestDetectAndCleanRepetition_LineLoop(t *testing.T) {
	input := "This is a detailed analysis of the quarterly budget.\nThis is a detailed analysis of the quarterly budget.\nLet's review the expenses."
	res := DetectAndCleanRepetition(input)
	if !res.HasLoop {
		t.Fatalf("Expected line loop to be detected")
	}
	expected := "This is a detailed analysis of the quarterly budget.\nLet's review the expenses."
	if res.CleanText != expected {
		t.Errorf("Expected %q, got %q", expected, res.CleanText)
	}
}

func TestDetectAndCleanRepetition_CharRunaway(t *testing.T) {
	input := "Processing......................."
	res := DetectAndCleanRepetition(input)
	if !res.HasLoop {
		t.Fatalf("Expected char runaway loop to be detected")
	}
	expected := "Processing..."
	if res.CleanText != expected {
		t.Errorf("Expected %q, got %q", expected, res.CleanText)
	}
}
