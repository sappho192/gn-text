package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rivo/tview"
)

var version = "dev"

var geekNewsRSSURL = "https://news.hada.io/rss/news"

func main() {
	versionFlag := flag.Bool("v", false, "Print version and exit")
	flag.BoolVar(versionFlag, "version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println("gn-text version", version)
		os.Exit(0)
	}

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
