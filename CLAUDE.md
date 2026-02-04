# CLAUDE.md

This file provides guidance for Claude Code when working with this repository.

## Project Overview

gn-text is a terminal-based GeekNews (news.hada.io) reader written in Go. It provides a TUI (Terminal User Interface) for browsing Korean tech news.

## Tech Stack

- **Language**: Go 1.21+
- **TUI Framework**: [tview](https://github.com/rivo/tview) with tcell
- **HTML Parsing**: goquery, html2text
- **Article Extraction**: articletext

## Project Structure

```
.
├── main.go          # Entry point, app initialization, version flag
├── ui.go            # UI components, input handling, page navigation
├── parser.go        # RSS/HTML parsing, data structures (Article, Comment)
├── web.go           # HTTP fetching, text formatting, terminal width handling
├── *_test.go        # Test files
├── chocolatey/      # Chocolatey package files for Windows distribution
└── .github/
    └── workflows/
        └── release.yml  # GoReleaser + Chocolatey release workflow
```

## Key Components

### Data Flow
1. `fetchWebpage()` fetches RSS from `https://news.hada.io/rss/news`
2. `parseGeekNewsRSS()` parses Atom feed into `[]Article`
3. `createArticleList()` creates tview List widget
4. User navigation triggers `openComments()` or `openArticle()`

### Main Types
- `Article`: Title, Link, CommentsLink, Domain
- `Comment`: Author, Body, Depth, Time, ID
- `TopicContent`: Title, ExternalLink, Body, Author, Time, Points

## Build & Test

```bash
# Build
go build

# Test
go test ./...

# Build with version
go build -ldflags "-X main.version=v1.0.0"

# Local GoReleaser test
goreleaser release --snapshot --clean
```

## Release Process

1. Commit changes to main
2. Create and push tag: `git tag v0.x.x && git push origin v0.x.x`
3. GitHub Actions automatically:
   - Builds binaries for linux/darwin/windows (amd64/arm64)
   - Creates GitHub Release
   - Updates Homebrew tap (sappho192/homebrew-tap)
   - Publishes Chocolatey package

## Code Conventions

- Korean comments are acceptable (target audience is Korean)
- Error messages displayed to users are in Korean
- Use `tcell.EventKey` for keyboard handling
- Terminal width is dynamically calculated via `term.GetSize()`
