# Original Design (hn-text)

## Purpose
- Terminal UI for browsing Hacker News (news.ycombinator.com) headlines, opening articles, and reading comments.

## High-level Flow
1. `main.go` sets up the TUI (`tview.Application`).
2. Fetches HN HTML with `fetchWebpage()`.
3. Parses article list via `parseArticles()`.
4. Builds a list UI with `createArticleList()` and starts the event loop.
5. Input handler routes navigation, refresh, open article, and open comments.

## Data Sources
- HN homepage HTML (`https://news.ycombinator.com/` or `/best`).
- HN Algolia API for comments: `https://hn.algolia.com/api/v1/items/{story_id}`.
- External article body via `github.com/gelembjuk/articletext`.

## Core Modules

### `main.go`
- Defines `hackerNewsURL`.
- Sets page ("" or "best") based on CLI args.
- Orchestrates fetch → parse → render.

### `web.go`
- `fetchWebpage(url)`: HTTP GET and returns response body as string.
- `fetchComments(storyID)`: Calls Algolia API, parses JSON, renders nested comment text.
- `appendComment(...)`: Recursive traversal of comment tree with indentation and HTML → text sanitization (`html2text`).
- `wrapText(...)`: Reflows comments to ~60 chars with indentation.

### `parser.go`
- `parseArticles(html)`: Uses goquery to find `tr.athing` rows and extract:
  - Title (`td.title > span.titleline > a`)
  - Link (`href`)
  - Comments count and comments link (from the following row’s `a[href^='item']`)
- `Article` model: `Title`, `Link`, `Comments`, `CommentsLink`.

### `ui.go`
- `createArticleList(...)`: Builds `tview.List`, shows title + domain + comment count; adds 🔥 for high comments.
- `createInputHandler(...)`: Keyboard actions and navigation.
- `openComments(...)`: Parses HN `item?id=...` to get story ID, then renders comments.
- `openArticle(...)`: Uses `articletext` to extract article body in text.
- `displayComments(...)` / `displayArticle(...)`: Add `tview.TextView` pages.
- `openURL(...)`: Opens article/comment in external browser.

## Keyboard Controls
- `j`/`k`: Move down/up.
- `h`/`l` or Left/Right arrows: Navigate comments ↔ article ↔ list.
- `space`: Open article link in browser.
- `c`: Open comments link in browser.
- `r`: Refresh list (re-fetches HN homepage).
- `q` or `Ctrl+C`: Quit.

## Tests
- `parser_test.go`: Parses sample HN HTML and validates titles/links/comments.
- `web_test.go`: Validates HTML sanitization for comment text.

## External Dependencies
- TUI: `github.com/rivo/tview`, `github.com/gdamore/tcell`.
- HTML parsing: `github.com/PuerkitoBio/goquery`.
- Comment sanitization: `jaytaylor.com/html2text`.
- Article extraction: `github.com/gelembjuk/articletext`.

## Assumptions / Coupling
- HN HTML structure matches selectors in `parser.go`.
- Comment retrieval relies on Algolia API schema.
- Comment links are `item?id=...` and are relative to `news.ycombinator.com`.
