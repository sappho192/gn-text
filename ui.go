package main

import (
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gelembjuk/articletext"
	"github.com/rivo/tview"
)

func createArticleList(articles []Article) *tview.List {
	list := tview.NewList().ShowSecondaryText(true).SetSecondaryTextColor(tcell.ColorGray)
	for _, article := range articles {
		// Display title and domain only (no comment count from RSS)
		list.AddItem(article.Title, article.Domain, 0, nil)
	}

	return list
}

func fetchArticles() ([]Article, error) {
	rssContent, err := fetchWebpage(geekNewsRSSURL)
	if err != nil {
		return nil, err
	}

	return parseGeekNewsRSS(rssContent)
}

func createInputHandler(app *tview.Application, list *tview.List, articles []Article, pages *tview.Pages) func(event *tcell.EventKey) *tcell.EventKey {
	// Store articles in closure for refresh
	currentArticles := articles

	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		case tcell.KeyRight, tcell.KeyEnter:
			nextPage(pages, app, currentArticles, list)
			return nil
		case tcell.KeyLeft:
			backPage(pages)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 'l':
				nextPage(pages, app, currentArticles, list)
				return nil
			case 'h':
				backPage(pages)
				return nil
			case ' ':
				if article, ok := getCurrentArticle(currentArticles, list); ok {
					openArticleInBrowser(article)
				}
				return nil
			case 'c':
				if article, ok := getCurrentArticle(currentArticles, list); ok {
					openCommentsInBrowser(article)
				}
				return nil
			case 'r':
				newArticles, err := fetchArticles()
				if err != nil {
					displayError(app, pages, "목록을 새로고침할 수 없습니다: "+err.Error())
					return nil
				}
				currentArticles = newArticles
				list.Clear()
				for _, article := range newArticles {
					list.AddItem(article.Title, article.Domain, 0, nil)
				}
				pages.SwitchToPage("homepage")
				app.SetRoot(pages, true)
			}
		}

		return event
	}
}

func backPage(pages *tview.Pages) {
	currentPage, _ := pages.GetFrontPage()
	if currentPage == "comments" {
		pages.SwitchToPage("homepage")
	}
	if currentPage == "article" {
		pages.SwitchToPage("comments")
	}
}

func getCurrentArticle(articles []Article, list *tview.List) (Article, bool) {
	idx := list.GetCurrentItem()
	if idx < 0 || idx >= len(articles) {
		return Article{}, false
	}
	return articles[idx], true
}

func nextPage(pages *tview.Pages, app *tview.Application, articles []Article, list *tview.List) {
	article, ok := getCurrentArticle(articles, list)
	if !ok {
		return
	}
	currentPage, _ := pages.GetFrontPage()
	if currentPage == "comments" {
		openArticle(app, article, pages)
	} else {
		openComments(app, article, pages)
	}
}

func openComments(app *tview.Application, article Article, pages *tview.Pages) {
	topicID := extractTopicID(article.CommentsLink)
	if topicID == "" {
		displayComments(app, pages, "토픽 ID를 찾을 수 없습니다.")
		return
	}

	commentLines := fetchGeekNewsComments(topicID)
	commentsText := strings.Join(commentLines, "\n")

	displayComments(app, pages, commentsText)
}

func openArticle(app *tview.Application, article Article, pages *tview.Pages) {
	// Try to get external link - first check if we have it cached
	externalLink := article.Link

	// If no external link, fetch from topic page
	if externalLink == "" {
		var err error
		externalLink, err = fetchExternalLink(article.CommentsLink)
		if err != nil || externalLink == "" {
			displayArticle(app, pages, "기사 링크를 찾을 수 없습니다. 'c' 키를 눌러 GeekNews 페이지에서 확인하세요.")
			return
		}
	}

	// Check if it's an internal GeekNews link (Ask GN style posts)
	if !isValidHTTPURL(externalLink) {
		displayArticle(app, pages, "이 게시물은 외부 링크가 없습니다. 'c' 키를 눌러 GeekNews에서 확인하세요.")
		return
	}

	articleText, err := getArticleTextFromLink(externalLink)
	if err != nil || articleText == "" {
		displayArticle(app, pages, "기사 내용을 추출할 수 없습니다. 'space' 키를 눌러 브라우저에서 열어보세요.")
		return
	}

	displayArticle(app, pages, articleText)
}

func getArticleTextFromLink(url string) (string, error) {
	html, err := fetchWebpage(url)
	if err != nil {
		return "", err
	}

	article, err := articletext.GetArticleText(strings.NewReader(html))
	if err != nil {
		return "", err
	}
	return article, nil
}

func displayArticle(app *tview.Application, pages *tview.Pages, text string) {
	articleTextView := tview.NewTextView().
		SetText(text).
		SetDynamicColors(true).
		SetScrollable(true)

	pages.AddPage("article", articleTextView, true, true)
	app.SetRoot(pages, true)
}

func displayComments(app *tview.Application, pages *tview.Pages, text string) {
	commentsTextView := tview.NewTextView().
		SetText(text).
		SetDynamicColors(true).
		SetScrollable(true)

	pages.AddPage("comments", commentsTextView, true, true)
	app.SetRoot(pages, true)
}

func displayError(app *tview.Application, pages *tview.Pages, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"확인"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.SwitchToPage("homepage")
			app.SetRoot(pages, true)
		})
	pages.AddPage("error", modal, true, true)
	app.SetRoot(pages, true)
}

// openArticleInBrowser opens the article's external link in the browser
func openArticleInBrowser(article Article) {
	externalLink := article.Link

	// If no cached external link, fetch it
	if externalLink == "" {
		var err error
		externalLink, err = fetchExternalLink(article.CommentsLink)
		if err != nil || externalLink == "" {
			// Fall back to opening the topic page
			openURL(article.CommentsLink)
			return
		}
	}

	openURL(externalLink)
}

// openCommentsInBrowser opens the comments page in the browser
func openCommentsInBrowser(article Article) {
	topicID := extractTopicID(article.CommentsLink)
	if topicID == "" {
		openURL(article.CommentsLink)
		return
	}
	commentsURL := geekNewsBaseURL + "topic?go=comments&id=" + topicID
	openURL(commentsURL)
}

func isValidHTTPURL(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func openURL(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	exec.Command(cmd, args...).Start()
}
