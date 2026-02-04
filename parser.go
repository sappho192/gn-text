package main

import (
	"encoding/xml"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Article represents a news article from GeekNews
type Article struct {
	Title        string
	Link         string // External article URL (may be empty if not available)
	Comments     string // Comment count as string (empty for RSS-based list)
	CommentsLink string // GeekNews topic URL
	Domain       string // Extracted from Link or "news.hada.io" if Link is topic URL
}

// Comment represents a comment from GeekNews
type Comment struct {
	Author string
	Body   string // HTML converted to plain text
	Depth  int    // Nesting level (0-based)
	Time   string // Display as-is from GeekNews
	ID     string // Comment ID
}

// AtomFeed represents the GeekNews Atom feed structure
type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []AtomEntry `xml:"entry"`
}

// AtomEntry represents a single entry in the Atom feed
type AtomEntry struct {
	Title   string     `xml:"title"`
	Links   []AtomLink `xml:"link"`
	ID      string     `xml:"id"`
	Updated string     `xml:"updated"`
	Author  AtomAuthor `xml:"author"`
	Content string     `xml:"content"`
}

// AtomLink represents a link element in Atom
type AtomLink struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

// AtomAuthor represents an author in Atom
type AtomAuthor struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

// parseGeekNewsRSS parses the GeekNews Atom feed and returns articles
func parseGeekNewsRSS(xmlData string) ([]Article, error) {
	var feed AtomFeed
	if err := xml.Unmarshal([]byte(xmlData), &feed); err != nil {
		return nil, err
	}

	var articles []Article
	for _, entry := range feed.Entries {
		// Find the alternate link (topic page URL)
		var topicURL string
		for _, link := range entry.Links {
			if link.Rel == "alternate" || link.Rel == "" {
				topicURL = link.Href
				break
			}
		}

		// If no alternate link found, use the ID which is also the topic URL
		if topicURL == "" {
			topicURL = entry.ID
		}

		article := Article{
			Title:        entry.Title,
			Link:         "", // External link not available in RSS, will be fetched on demand
			Comments:     "", // Not available in RSS
			CommentsLink: topicURL,
			Domain:       "news.hada.io",
		}

		articles = append(articles, article)
	}

	return articles, nil
}

// parseGeekNewsComments parses the GeekNews comments HTML and returns comments
func parseGeekNewsComments(htmlContent string) ([]Comment, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var comments []Comment
	depthRegex := regexp.MustCompile(`--depth:\s*(\d+)`)

	doc.Find("#comment_thread .comment_row").Each(func(i int, s *goquery.Selection) {
		// Extract depth from style attribute
		depth := 0
		if style, exists := s.Attr("style"); exists {
			if matches := depthRegex.FindStringSubmatch(style); len(matches) > 1 {
				depth, _ = strconv.Atoi(matches[1])
			}
		}

		// Limit depth to 10 as per spec
		if depth > 10 {
			// Skip deeply nested comments, will show [...] indicator
			return
		}

		// Extract author
		author := s.Find(".commentinfo a[href^='/user?id=']").First().Text()

		// Extract time
		time := s.Find(".commentinfo a[href^='comment?id=']").Text()

		// Extract comment ID from element ID (e.g., "cid50523" -> "50523")
		var commentID string
		if id, exists := s.Attr("id"); exists {
			commentID = strings.TrimPrefix(id, "cid")
		}

		// Extract body and sanitize
		bodyHTML, _ := s.Find(".commentTD .comment_contents").Html()
		body := sanitize(bodyHTML)

		comment := Comment{
			Author: author,
			Body:   body,
			Depth:  depth,
			Time:   time,
			ID:     commentID,
		}

		comments = append(comments, comment)
	})

	return comments, nil
}

// parseGeekNewsTopicLink extracts the external article link from a topic page
func parseGeekNewsTopicLink(htmlContent string) (string, string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", "", err
	}

	// Try to find external link in topic header
	// Selector: .topictitle.link > a or .topictitle > a
	var externalLink string
	var title string

	// First try .topictitle.link > a
	linkSel := doc.Find(".topictitle.link > a").First()
	if linkSel.Length() > 0 {
		externalLink, _ = linkSel.Attr("href")
		title = linkSel.Find("h1").Text()
		if title == "" {
			title = linkSel.Text()
		}
	}

	// If not found, try alternative selectors
	if externalLink == "" {
		linkSel = doc.Find(".topictitle > a").First()
		if linkSel.Length() > 0 {
			externalLink, _ = linkSel.Attr("href")
			title = linkSel.Find("h1").Text()
			if title == "" {
				title = linkSel.Text()
			}
		}
	}

	return externalLink, title, nil
}

// extractDomainFromURL extracts the domain from a URL
func extractDomainFromURL(link string) string {
	if link == "" {
		return ""
	}
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}
	return u.Host
}

// extractTopicID extracts the topic ID from a GeekNews URL
func extractTopicID(topicURL string) string {
	u, err := url.Parse(topicURL)
	if err != nil {
		return ""
	}
	return u.Query().Get("id")
}
