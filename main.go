package main

import (
	"log"

	"github.com/rivo/tview"
)

var geekNewsRSSURL = "https://news.hada.io/rss/news"

func main() {
	app := tview.NewApplication()

	rssContent, err := fetchWebpage(geekNewsRSSURL)
	if err != nil {
		log.Fatal(err)
	}

	articles, err := parseGeekNewsRSS(rssContent)
	if err != nil {
		log.Fatal(err)
	}

	list := createArticleList(articles)
	pages := tview.NewPages()
	pages.AddPage("homepage", list, true, false)

	app.SetInputCapture(createInputHandler(app, list, articles, pages))

	if err := app.SetRoot(list, true).Run(); err != nil {
		log.Fatal(err)
	}
}
