package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/mattn/go-runewidth"
	"jaytaylor.com/html2text"
)

const geekNewsBaseURL = "https://news.hada.io/"

func fetchWebpage(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func sanitize(input string) string {
	sanitized, _ := html2text.FromString(input)
	return sanitized
}

// fetchGeekNewsComments fetches and formats comments for a GeekNews topic
func fetchGeekNewsComments(topicID string) []string {
	// Fetch the full topic page (not just comments) to get body content
	topicURL := geekNewsBaseURL + "topic?id=" + topicID
	html, err := fetchWebpage(topicURL)
	if err != nil {
		return []string{"페이지를 불러오는데 실패했습니다: " + err.Error()}
	}

	var lines []string

	// Parse and display topic content (body)
	topicContent, err := parseGeekNewsTopicContent(html)
	if err == nil && topicContent.Body != "" {
		lines = append(lines, formatTopicContent(topicContent)...)
		lines = append(lines, "")
		lines = append(lines, "───────────────────────────────────────────────────────────")
		lines = append(lines, "")
	}

	// Parse comments
	comments, err := parseGeekNewsComments(html)
	if err != nil {
		lines = append(lines, "댓글을 파싱하는데 실패했습니다: "+err.Error())
		return lines
	}

	if len(comments) == 0 {
		lines = append(lines, "아직 댓글이 없습니다. 오른쪽 화살표 또는 'l' 키를 눌러 기사를 읽어보세요.")
		return lines
	}

	lines = append(lines, formatComments(comments)...)
	return lines
}

// formatTopicContent formats the topic content (title, meta, body) for display
func formatTopicContent(content *TopicContent) []string {
	var lines []string
	maxWidth := 60

	// Title
	if content.Title != "" {
		lines = append(lines, "[yellow]"+content.Title+"[-]")
		lines = append(lines, "")
	}

	// Meta info (author, time, points)
	var meta []string
	if content.Author != "" {
		meta = append(meta, content.Author)
	}
	if content.Time != "" {
		meta = append(meta, content.Time)
	}
	if content.Points != "" {
		meta = append(meta, content.Points+"P")
	}
	if len(meta) > 0 {
		lines = append(lines, "[gray]"+strings.Join(meta, " · ")+"[-]")
		lines = append(lines, "")
	}

	// Body content
	if content.Body != "" {
		// Split body into paragraphs and wrap each
		paragraphs := strings.Split(content.Body, "\n\n")
		for _, paragraph := range paragraphs {
			paragraph = strings.TrimSpace(paragraph)
			if paragraph == "" {
				continue
			}
			// Handle single newlines as line breaks within paragraph
			subLines := strings.Split(paragraph, "\n")
			for _, subLine := range subLines {
				subLine = strings.TrimSpace(subLine)
				if subLine == "" {
					continue
				}
				wrappedLines := wrapTextWithRuneWidth(subLine, maxWidth, "")
				lines = append(lines, wrappedLines...)
			}
			lines = append(lines, "")
		}
	}

	return lines
}

// formatComments formats a list of comments for display
func formatComments(comments []Comment) []string {
	var lines []string
	maxWidth := 60

	for _, comment := range comments {
		// Create indentation based on depth (limit visual depth to 4 for readability)
		visualDepth := min(comment.Depth, 4)
		indent := strings.Repeat("   ", visualDepth*2) + "| "

		// Add author line with time
		authorLine := indent + comment.Author
		if comment.Time != "" {
			authorLine += " (" + comment.Time + ")"
		}
		authorLine += " 님:"
		lines = append(lines, authorLine)

		// Process comment body
		if comment.Body == "" {
			lines = append(lines, indent+"[삭제됨]")
		} else {
			// Split into paragraphs and wrap each
			paragraphs := strings.Split(comment.Body, "\n\n")
			for _, paragraph := range paragraphs {
				paragraph = strings.TrimSpace(paragraph)
				if paragraph == "" {
					continue
				}
				wrappedLines := wrapTextWithRuneWidth(paragraph, maxWidth, indent)
				lines = append(lines, wrappedLines...)
				lines = append(lines, indent)
			}
			// Remove trailing empty indent line
			if len(lines) > 0 && lines[len(lines)-1] == indent {
				lines = lines[:len(lines)-1]
			}
		}

		// Add depth indicator for deeply nested comments
		if comment.Depth > 10 {
			lines = append(lines, indent+"[...]")
		}

		lines = append(lines, "  ")
	}

	return lines
}

// wrapTextWithRuneWidth wraps text using runewidth for accurate CJK character width
func wrapTextWithRuneWidth(text string, maxWidth int, indent string) []string {
	words := strings.Fields(text)
	var lines []string
	var currentLine strings.Builder
	indentWidth := runewidth.StringWidth(indent)
	effectiveWidth := maxWidth - indentWidth

	for _, word := range words {
		wordWidth := runewidth.StringWidth(word)
		currentWidth := runewidth.StringWidth(currentLine.String())

		if currentWidth+wordWidth+1 > effectiveWidth && currentWidth > 0 {
			// Current word doesn't fit, start new line
			lines = append(lines, indent+currentLine.String())
			currentLine.Reset()
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		lines = append(lines, indent+currentLine.String())
	}

	return lines
}

// fetchExternalLink fetches a topic page and extracts the external article link
func fetchExternalLink(topicURL string) (string, error) {
	html, err := fetchWebpage(topicURL)
	if err != nil {
		return "", err
	}

	link, _, err := parseGeekNewsTopicLink(html)
	if err != nil {
		return "", err
	}

	return link, nil
}
