package main

import (
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	unsanitizedComment := "Uncrewed, yes. <a href=\"https:&#x2F;&#x2F;en.wikipedia.org&#x2F;wiki&#x2F;Boeing_Orbital_Flight_Test_2\" rel=\"nofollow\">https:&#x2F;&#x2F;en.wikipedia.org&#x2F;wiki&#x2F;Boeing_Orbital_Flight_Test_2</a>"
	expectedComment := "Uncrewed, yes. https://en.wikipedia.org/wiki/Boeing_Orbital_Flight_Test_2"
	sanitizedComment := sanitize(unsanitizedComment)

	if sanitizedComment != expectedComment {
		t.Errorf("Expected %q, got %q", expectedComment, sanitizedComment)
	}
}

func TestSanitizeKorean(t *testing.T) {
	// Test Korean text with HTML tags
	input := "<p>안녕하세요! <strong>테스트</strong>입니다.</p>"
	result := sanitize(input)

	if !strings.Contains(result, "안녕하세요") {
		t.Errorf("Expected Korean text to be preserved, got %q", result)
	}
	if strings.Contains(result, "<p>") || strings.Contains(result, "<strong>") {
		t.Errorf("Expected HTML tags to be removed, got %q", result)
	}
}

func TestWrapTextWithRuneWidth(t *testing.T) {
	// Test English text wrapping
	text := "This is a test string that should be wrapped at a certain width"
	lines := wrapTextWithRuneWidth(text, 30, "  ")

	if len(lines) < 2 {
		t.Errorf("Expected multiple lines, got %d", len(lines))
	}

	for _, line := range lines {
		if !strings.HasPrefix(line, "  ") {
			t.Errorf("Expected line to start with indent, got %q", line)
		}
	}
}

func TestWrapTextWithRuneWidthKorean(t *testing.T) {
	// Test Korean text wrapping - Korean characters are 2 cells wide
	text := "한글 테스트 문자열입니다 긴 문장을 줄바꿈하는 테스트"
	lines := wrapTextWithRuneWidth(text, 30, "")

	if len(lines) == 0 {
		t.Error("Expected at least one line")
	}

	// Each line should contain Korean text
	for _, line := range lines {
		if line == "" {
			t.Error("Unexpected empty line")
		}
	}
}

func TestWrapTextWithRuneWidthMixed(t *testing.T) {
	// Test mixed English and Korean text
	text := "안녕 hello 세계 world 테스트 test"
	lines := wrapTextWithRuneWidth(text, 20, "| ")

	if len(lines) == 0 {
		t.Error("Expected at least one line")
	}

	for _, line := range lines {
		if !strings.HasPrefix(line, "| ") {
			t.Errorf("Expected line to start with '| ', got %q", line)
		}
	}
}

func TestWrapTextWithRuneWidthEmpty(t *testing.T) {
	lines := wrapTextWithRuneWidth("", 60, "")

	if len(lines) != 0 {
		t.Errorf("Expected empty result for empty input, got %v", lines)
	}
}

func TestWrapTextWithRuneWidthSingleWord(t *testing.T) {
	lines := wrapTextWithRuneWidth("hello", 60, ">> ")

	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	if lines[0] != ">> hello" {
		t.Errorf("Expected '>> hello', got %q", lines[0])
	}
}

func TestFormatComments(t *testing.T) {
	comments := []Comment{
		{
			Author: "user1",
			Body:   "This is a test comment.",
			Depth:  0,
			Time:   "1시간전",
			ID:     "123",
		},
		{
			Author: "user2",
			Body:   "This is a reply.",
			Depth:  1,
			Time:   "30분전",
			ID:     "124",
		},
	}

	lines := formatComments(comments)

	if len(lines) == 0 {
		t.Error("Expected formatted lines")
	}

	// Check that author names appear
	found := false
	for _, line := range lines {
		if strings.Contains(line, "user1") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find author 'user1' in formatted comments")
	}

	// Check that nested comment has more indentation
	var depth0Indent, depth1Indent int
	for _, line := range lines {
		if strings.Contains(line, "user1") {
			depth0Indent = len(line) - len(strings.TrimLeft(line, " |"))
		}
		if strings.Contains(line, "user2") {
			depth1Indent = len(line) - len(strings.TrimLeft(line, " |"))
		}
	}
	if depth1Indent <= depth0Indent {
		t.Error("Expected depth 1 comment to have more indentation than depth 0")
	}
}

func TestFormatCommentsEmpty(t *testing.T) {
	lines := formatComments([]Comment{})

	if len(lines) != 0 {
		t.Errorf("Expected empty result for empty comments, got %v", lines)
	}
}

func TestFormatCommentsDeletedComment(t *testing.T) {
	comments := []Comment{
		{
			Author: "",
			Body:   "",
			Depth:  0,
			Time:   "",
			ID:     "123",
		},
	}

	lines := formatComments(comments)

	found := false
	for _, line := range lines {
		if strings.Contains(line, "[삭제됨]") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected deleted comment indicator '[삭제됨]'")
	}
}
