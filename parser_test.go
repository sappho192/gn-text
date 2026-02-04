package main

import (
	"os"
	"strings"
	"testing"
)

func TestParseGeekNewsRSS(t *testing.T) {
	xmlContent, err := os.ReadFile("testdata/geeknews_feed.xml")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	articles, err := parseGeekNewsRSS(string(xmlContent))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(articles) != 2 {
		t.Fatalf("Expected 2 articles, got %d", len(articles))
	}

	// Test first article
	if articles[0].Title != "AI 코딩 도구가 개발자 학습을 방해한다, Anthropic 연구 발견" {
		t.Errorf("Expected title containing 'AI 코딩 도구', got %q", articles[0].Title)
	}
	if articles[0].CommentsLink != "https://news.hada.io/topic?id=26364" {
		t.Errorf("Expected CommentsLink 'https://news.hada.io/topic?id=26364', got %q", articles[0].CommentsLink)
	}
	if articles[0].Comments != "" {
		t.Errorf("Expected empty Comments field, got %q", articles[0].Comments)
	}
	if articles[0].Domain != "news.hada.io" {
		t.Errorf("Expected Domain 'news.hada.io', got %q", articles[0].Domain)
	}

	// Test second article
	if articles[1].Title != "Todd C. Miller – 30년 넘게 Sudo를 유지보수한 개발자" {
		t.Errorf("Expected title containing 'Sudo', got %q", articles[1].Title)
	}
	if articles[1].CommentsLink != "https://news.hada.io/topic?id=26363" {
		t.Errorf("Expected CommentsLink 'https://news.hada.io/topic?id=26363', got %q", articles[1].CommentsLink)
	}
}

func TestParseGeekNewsRSS_EmptyFeed(t *testing.T) {
	emptyFeed := `<?xml version='1.0' encoding='UTF-8'?>
<feed xmlns='http://www.w3.org/2005/Atom'>
<title>GeekNews</title>
</feed>`

	articles, err := parseGeekNewsRSS(emptyFeed)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(articles) != 0 {
		t.Errorf("Expected 0 articles, got %d", len(articles))
	}
}

func TestParseGeekNewsRSS_InvalidXML(t *testing.T) {
	invalidXML := "not valid xml at all"

	_, err := parseGeekNewsRSS(invalidXML)
	if err == nil {
		t.Error("Expected error for invalid XML, got nil")
	}
}

func TestParseGeekNewsComments(t *testing.T) {
	htmlContent, err := os.ReadFile("testdata/geeknews_topic_comments.html")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	comments, err := parseGeekNewsComments(string(htmlContent))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(comments) == 0 {
		t.Fatal("Expected comments, got none")
	}

	// Test first comment (depth 0)
	firstComment := comments[0]
	if firstComment.Author != "kuthia" {
		t.Errorf("Expected author 'kuthia', got %q", firstComment.Author)
	}
	if firstComment.Depth != 0 {
		t.Errorf("Expected depth 0, got %d", firstComment.Depth)
	}
	if firstComment.ID != "50523" {
		t.Errorf("Expected ID '50523', got %q", firstComment.ID)
	}
	if firstComment.Body == "" {
		t.Error("Expected non-empty body")
	}
	// Check time
	if firstComment.Time != "5시간전" {
		t.Errorf("Expected time '5시간전', got %q", firstComment.Time)
	}

	// Test second comment (depth 1 - reply)
	if len(comments) > 1 {
		secondComment := comments[1]
		if secondComment.Depth != 1 {
			t.Errorf("Expected depth 1 for second comment, got %d", secondComment.Depth)
		}
		if secondComment.Author != "gracefullight" {
			t.Errorf("Expected author 'gracefullight', got %q", secondComment.Author)
		}
	}
}

func TestParseGeekNewsComments_Empty(t *testing.T) {
	emptyHTML := `<div id='comment_thread' class='comment_thread'></div>`

	comments, err := parseGeekNewsComments(emptyHTML)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(comments) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(comments))
	}
}

func TestExtractTopicID(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://news.hada.io/topic?id=26364", "26364"},
		{"https://news.hada.io/topic?id=12345&other=param", "12345"},
		{"https://news.hada.io/topic?go=comments&id=26364", "26364"},
		{"https://news.hada.io/topic", ""},
		{"invalid url", ""},
		{"", ""},
	}

	for _, test := range tests {
		result := extractTopicID(test.url)
		if result != test.expected {
			t.Errorf("extractTopicID(%q) = %q, expected %q", test.url, result, test.expected)
		}
	}
}

func TestParseGeekNewsTopicContent(t *testing.T) {
	htmlContent, err := os.ReadFile("testdata/geeknews_topic_full.html")
	if err != nil {
		t.Fatalf("Failed to read test fixture: %v", err)
	}

	content, err := parseGeekNewsTopicContent(string(htmlContent))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test title
	if content.Title == "" {
		t.Error("Expected non-empty title")
	}
	if !strings.Contains(content.Title, "AI 코딩 도구") {
		t.Errorf("Expected title to contain 'AI 코딩 도구', got %q", content.Title)
	}

	// Test external link
	if content.ExternalLink == "" {
		t.Error("Expected non-empty external link")
	}
	if !strings.Contains(content.ExternalLink, "anthropic.com") {
		t.Errorf("Expected external link to contain 'anthropic.com', got %q", content.ExternalLink)
	}

	// Test body
	if content.Body == "" {
		t.Error("Expected non-empty body")
	}
	if !strings.Contains(content.Body, "핵심") {
		t.Errorf("Expected body to contain '핵심', got %q", content.Body)
	}

	// Test author
	if content.Author == "" {
		t.Error("Expected non-empty author")
	}

	// Test points
	if content.Points == "" {
		t.Error("Expected non-empty points")
	}
}

func TestParseGeekNewsTopicContent_NoBody(t *testing.T) {
	// Test with minimal HTML (no topic_contents)
	minimalHTML := `<div class="topictitle"><a href="https://example.com"><h1>Test Title</h1></a></div>`

	content, err := parseGeekNewsTopicContent(minimalHTML)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if content.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got %q", content.Title)
	}

	if content.Body != "" {
		t.Errorf("Expected empty body, got %q", content.Body)
	}
}

func TestExtractDomainFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com/path", "example.com"},
		{"https://www.example.com/path?query=1", "www.example.com"},
		{"http://sub.domain.org:8080/", "sub.domain.org:8080"},
		{"invalid url", ""},
		{"", ""},
	}

	for _, test := range tests {
		result := extractDomainFromURL(test.url)
		if result != test.expected {
			t.Errorf("extractDomainFromURL(%q) = %q, expected %q", test.url, result, test.expected)
		}
	}
}
