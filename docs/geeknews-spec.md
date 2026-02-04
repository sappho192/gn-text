# GeekNews TUI Specification (gn-text)

**Version**: 1.0.0
**Date**: 2026-02-03
**Status**: Final Specification

## 1. Project Overview

### 1.1 Project Name
- **Name**: `gn-text` (GeekNews Text)
- **Description**: Terminal-based UI for browsing GeekNews (news.hada.io) headlines, articles, and comments.
- **Rationale**: Maintains naming consistency with `hn-text` while clearly indicating GeekNews as the target platform.

### 1.2 Goals
- Replace HN-specific implementation with GeekNews support
- Preserve existing TUI UX and keybindings
- Maintain simplicity and performance
- Support Korean text properly (encoding, wrapping, display)

### 1.3 Non-Goals (Out of Scope)
- Multi-site support (HN + GeekNews simultaneously)
- User authentication or posting comments
- Real-time notifications or auto-refresh
- Advanced filtering or search within the app

---

## 2. Architecture

### 2.1 Data Sources

#### Primary: RSS Feed
- **URL**: `https://news.hada.io/rss/news`
- **Format**: Atom/RSS feed
- **Usage**: Primary source for article list
- **Trade-off**: Comment count not available in RSS, but provides stable structure
- **Rationale**: RSS structure is more stable than HTML DOM, reducing breakage risk

#### Secondary: HTML Parsing
- **URL**: `https://news.hada.io/topic?go=comments&id={story_id}`
- **Usage**: Comment pages only
- **Parsing Library**: `github.com/PuerkitoBio/goquery`

#### Article Content Extraction
- **Primary**: `github.com/gelembjuk/articletext`
- **Fallback Chain**: Additional libraries to be tried in order:
  - `github.com/go-shiori/go-readability`
  - Others TBD during implementation
- **Strategy**: Dynamic prioritization based on success rate learning (see Section 3.5)

### 2.2 Core Modules

#### `main.go`
- Replace `hackerNewsURL` with `geekNewsRSSURL = "https://news.hada.io/rss/news"`
- Remove page selection logic (no "best" equivalent in GeekNews)
- Initialize cache manager
- Orchestrate: fetch RSS → parse → render

#### `parser.go`
- `parseGeekNewsRSS(xmlData string) ([]Article, error)`: Parse RSS/Atom feed
- `parseGeekNewsComments(html string) ([]Comment, error)`: Parse comment page HTML
- Update `Article` struct if needed (keep existing fields)
- **Comment Depth Limit**: Render up to 10 levels, display `[...]` for deeper comments
- **Selectors**: Use concrete selectors from `geeknews-design-plan.md` Section "Concrete Selectors"

#### `web.go`
- Keep `fetchWebpage(url string)` as-is
- Replace `fetchComments(storyID)` with `fetchGeekNewsComments(topicID string)`
  - Extract story ID from `topic?go=comments&id=...` URL
  - Call Algolia API → Replace with HTML fetch + parse
- Keep `sanitize()` and HTML-to-text utilities
- Add `wrapTextWithRuneWidth()`: Use `github.com/mattn/go-runewidth` for accurate CJK character width calculation

#### `ui.go`
- Update `openComments()` to parse GeekNews topic URLs
- Update `openURL()` base URL to `news.hada.io`
- **Browser**: Use OS default browser via `open` (macOS) or `xdg-open` (Linux)
- **Loading Indicator**: Display "Loading..." in status bar (top of screen)
- **Error Display**: Show error messages in status bar (non-blocking)
- **Empty State**: Display "내용 없음" (No content) for empty lists/comments/articles
- **Date Format**: Display dates as-is from GeekNews (no reformatting)

#### `cache.go` (New Module)
- In-memory cache with TTL: 5-10 minutes
- Disk cache with TTL: 30-60 minutes
- Cache keys: URL-based
- Storage location: `~/.cache/gn-text/` (or OS-specific cache dir)
- Invalidation: Manual via 'r' key, automatic on TTL expiry

#### `extractor.go` (New Module)
- Article extraction with fallback chain
- Success rate tracking per library
- Dynamic prioritization based on historical success
- Timeout per library: 5-10 seconds
- Context-based cancellation support

---

## 3. Features

### 3.1 Article List (MVP)
- **Source**: RSS feed (`https://news.hada.io/rss/news`)
- **Display Format**: `{title} ({domain})`
- **Comment Count**: Not displayed (RSS does not provide this data)
- **Fire Emoji (🔥)**: Removed (no comment count available)
- **Sorting**: As provided by RSS feed (chronological)

### 3.2 Comments (MVP)
- **Source**: HTML parsing of `topic?go=comments&id={id}`
- **Display**: Nested/threaded comments with indentation
- **Indentation**: Use runewidth-based calculation for CJK characters
- **Depth Limit**: Display up to 10 levels, show `[...]` for deeper comments
- **Date Display**: Show timestamps as-is from GeekNews HTML
- **Pagination**: If comments are paginated:
  - Load first page only initially
  - User can press 'n' key to load next page
  - Show "Loading more comments..." in status bar

### 3.3 Article Viewing (MVP)
- **Extraction**: Use fallback chain (articletext → go-readability → ...)
- **Display**: Plain text with wrapping via runewidth
- **Failure Handling**: Show error in status bar, user can press 'o' to open in browser

### 3.4 Browser Integration (MVP)
- **Keys**: `space` (open article), `c` (open comments in browser)
- **Command**: OS default via `open` (macOS) or `xdg-open` (Linux)
- **No URL Clipboard**: Not implemented in MVP

### 3.5 Article Extraction Intelligence (Post-MVP)
- **Success Rate Tracking**:
  - Track success/failure per extractor library
  - Store stats in `~/.cache/gn-text/extractor-stats.json`
  - Update after each extraction attempt
- **Dynamic Prioritization**:
  - Reorder extraction libraries based on success rate
  - Re-evaluate every 100 attempts
  - Always try at least 2 libraries even if one has high success rate
- **Timeout**: 5-10 seconds per library attempt
- **Cancellation**: Support context cancellation if user navigates away

### 3.6 Caching (Post-MVP)
- **Memory Cache**:
  - TTL: 5-10 minutes
  - Scope: Article list, comments, article content
  - Eviction: LRU when memory limit reached
- **Disk Cache**:
  - TTL: 30-60 minutes
  - Storage: `~/.cache/gn-text/`
  - Format: JSON or gob encoding
  - Cleanup: On app start, remove expired entries
- **Invalidation**: Manual via 'r' key

### 3.7 Error Handling
- **Display**: Show error message in status bar (top of screen)
- **Retry Strategy**: Automatic retry 1-2 times with exponential backoff
- **Network Errors**: "Failed to fetch data. Retrying..."
- **Parse Errors**: "Failed to parse GeekNews data. Please try again."
- **Timeout Errors**: "Request timeout. Check your connection."
- **Detailed Logging**: Log errors to stderr with selector/URL info for debugging

---

## 4. Keyboard Controls

### 4.1 Navigation (Same as HN)
- `j` / Down Arrow: Move down in list
- `k` / Up Arrow: Move up in list
- `h` / Left Arrow: Go back (comments → list, article → list)
- `l` / Right Arrow: Enter/open selected item
- `Enter`: Open comments for selected article
- `Tab`: Toggle between article and comments view

### 4.2 Actions (Same as HN)
- `space`: Open article link in browser
- `c`: Open comments link in browser
- `o`: Open current article/comments in browser (alias for space/c)
- `r`: Refresh current page (invalidate cache, re-fetch)
- `q` / `Ctrl+C`: Quit

### 4.3 New Keys
- `n`: Load next page (for paginated comments)
- `?`: Show help (keybindings)

---

## 5. Text Handling (Korean Support)

### 5.1 Encoding
- **Input**: UTF-8 from GeekNews RSS/HTML
- **Display**: UTF-8 in terminal via tcell
- **No Conversion**: Assume terminal supports UTF-8

### 5.2 Line Wrapping
- **Algorithm**: Use `github.com/mattn/go-runewidth` to calculate display width
- **Width Calculation**: Account for full-width CJK characters (2 cells) vs. ASCII (1 cell)
- **Target Width**: 60-80 characters (adaptive based on terminal width if possible)
- **Indentation Preservation**: Maintain comment indentation after wrap

### 5.3 HTML Sanitization
- **Library**: `jaytaylor.com/html2text` (same as HN version)
- **Korean Handling**: Library should handle UTF-8 correctly (verify in tests)

---

## 6. Data Model

### 6.1 Article Struct
```go
type Article struct {
    Title        string  // From RSS <title>
    Link         string  // From RSS <link> (external article URL)
    Comments     string  // Not available in RSS, set to ""
    CommentsLink string  // From RSS <guid> or <id> (GeekNews topic URL)
    Domain       string  // Extracted from Link
    // Optional fields (can add later):
    // Points       int
    // Author       string
    // Time         time.Time
}
```

### 6.2 Comment Struct
```go
type Comment struct {
    Author  string
    Body    string  // HTML converted to plain text
    Depth   int     // Nesting level (0-based)
    Time    string  // Display as-is from GeekNews
    ID      string  // Comment ID (if available)
    // Children []Comment  // Not used; flat list with depth field
}
```

### 6.3 Cache Entry
```go
type CacheEntry struct {
    Key       string
    Data      interface{}  // []Article, []Comment, or string (article body)
    Timestamp time.Time
    TTL       time.Duration
}
```

---

## 7. Parsing Strategy

### 7.1 Article List (RSS-First)
1. Fetch `https://news.hada.io/rss/news`
2. Parse XML using `encoding/xml`
3. Extract fields:
   - `<title>` → `Article.Title`
   - `<link>` → `Article.Link`
   - `<guid>` or `<id>` → `Article.CommentsLink` (GeekNews topic URL)
4. Extract domain from `Link` using `net/url`
5. Set `Comments` to empty string (no count available)

### 7.2 Comments (HTML Parsing)
1. Fetch `https://news.hada.io/topic?go=comments&id={id}`
2. Parse HTML using `goquery`
3. Select comments: `#comment_thread .comment_row`
4. For each comment:
   - Extract depth from `style="--depth:N"` attribute
   - Extract author from `.commentinfo a[href^='/user?id=']`
   - Extract body from `.commentTD .comment_contents` (sanitize HTML → text)
   - Extract time from `.commentinfo a[href^='comment?id=']` text
5. Limit depth to 10; append `[...]` for deeper comments
6. Return flat list of `Comment` structs with `Depth` field

### 7.3 Article Content (Fallback Chain)
1. Try `articletext.Extract(url)`
2. If fails or timeout, try `readability.FromURL(url, timeout)`
3. If all fail, return error and show in status bar
4. Track success rate per library for future prioritization

---

## 8. Error Handling & Resilience

### 8.1 Network Errors
- **Automatic Retry**: 1-2 attempts with exponential backoff (e.g., 1s, 2s)
- **Timeout**: 10 seconds per HTTP request
- **Display**: "Failed to fetch data. Retrying..." in status bar
- **After Retries**: Show final error message, allow manual retry via 'r' key

### 8.2 Parsing Errors
- **Selector Fallback**: Try multiple selectors if primary fails (implement per design plan)
- **Logging**: Log failed selectors with sample HTML snippet (to stderr)
- **Display**: "Failed to parse GeekNews data. Please report this issue."
- **Graceful Degradation**: Show partial data if some fields are missing

### 8.3 DOM Structure Changes
- **Mitigation**: Use RSS as primary source (more stable)
- **Detection**: Unit tests with real GeekNews fixtures (run in CI)
- **Logging**: Log detailed error with URL and selector used
- **User Guidance**: Show error with link to GitHub issues

### 8.4 Article Extraction Failures
- **Fallback Chain**: Try multiple libraries (Section 3.3)
- **Display**: "Failed to extract article content. Press 'o' to open in browser."
- **No Blank Page**: Always show error message, never empty content

---

## 9. Testing Strategy

### 9.1 Unit Tests (Required for MVP)
- **Fixtures**: Store sample GeekNews HTML/RSS in `testdata/`
  - `geeknews_feed.xml`: RSS feed with 2-3 articles
  - `geeknews_homepage_topics.html`: Homepage article list
  - `geeknews_topic_comments.html`: Comment page with nested comments
- **Test Cases**:
  - `TestParseGeekNewsRSS`: Verify article fields parsed correctly
  - `TestParseGeekNewsComments`: Verify comment depth, author, body
  - `TestWrapTextWithRuneWidth`: Verify Korean text wrapping
  - `TestHTMLSanitization`: Verify Korean HTML → text conversion

### 9.2 Integration Tests (Optional, Post-MVP)
- **Live Requests**: Fetch actual GeekNews data in CI (run periodically)
- **Purpose**: Detect DOM structure changes early
- **Frequency**: Daily or weekly (not on every commit)
- **Alerts**: Notify maintainers if parsing fails

### 9.3 Manual Testing Checklist
- [ ] List view loads and displays articles
- [ ] Selecting article opens comments
- [ ] Comments display with correct indentation
- [ ] Korean text displays without corruption
- [ ] Pressing 'space' opens article in browser
- [ ] Pressing 'c' opens comments in browser
- [ ] Pressing 'r' refreshes list
- [ ] Error messages appear in status bar
- [ ] Pressing 'q' quits the app
- [ ] Article extraction works for Korean sites

---

## 10. Build & Deployment

### 10.1 Build System
- **Tool**: `goreleaser` for multi-platform builds
- **Platforms**:
  - macOS (darwin/amd64, darwin/arm64)
  - Linux (linux/amd64, linux/arm64)
  - Windows (windows/amd64) - optional
- **Output**: Binaries in `dist/` directory
- **Makefile**: Provide targets for common tasks
  ```makefile
  .PHONY: build test install clean

  build:
      go build -o gn-text .

  test:
      go test -v ./...

  install:
      go install .

  clean:
      rm -f gn-text
      rm -rf dist/
  ```

### 10.2 Release Process
1. Update `VERSION` in `main.go`
2. Tag release: `git tag v1.0.0`
3. Push tag: `git push origin v1.0.0`
4. GitHub Actions runs `goreleaser`
5. Binaries uploaded to GitHub Releases

### 10.3 Distribution
- **Primary**: GitHub Releases (download binary)
- **Go Install**: `go install github.com/piqoni/gn-text@latest`
- **Homebrew**: Future enhancement (create tap)

---

## 11. Documentation

### 11.1 README.md (English)
- **Sections**:
  - Project description
  - Features
  - Installation (go install, binary download)
  - Usage (basic commands)
  - Keyboard shortcuts (table)
  - Screenshots/GIFs (demo)
  - Contributing
  - License

### 11.2 README.ko.md (Korean)
- **Sections**: Same as README.md, translated to Korean
- **Rationale**: GeekNews is a Korean site, Korean documentation improves accessibility

### 11.3 ARCHITECTURE.md
- **Purpose**: Developer documentation for contributors
- **Sections**:
  - Project structure (files and responsibilities)
  - Data flow (fetch → parse → render)
  - Adding new features
  - Testing guidelines
  - Code style

### 11.4 Screenshots/GIFs
- **Tool**: `asciinema` or `terminalizer`
- **Content**:
  - List view with articles
  - Comment view with nested comments
  - Article view with Korean text
  - Error handling example

---

## 12. Configuration

### 12.1 No Config File (MVP)
- **Rationale**: Keep it simple, most users don't need customization
- **Hardcoded Values**:
  - Wrap width: 60 characters (or dynamic based on terminal width)
  - Cache TTL: 10 minutes (memory), 1 hour (disk)
  - Comment depth limit: 10 levels
  - Request timeout: 10 seconds
  - Retry count: 2 attempts

### 12.2 Future Configuration (Post-MVP)
- **File**: `~/.config/gn-text/config.yaml`
- **Fields**:
  ```yaml
  wrap_width: 80
  cache_ttl_memory: 600  # seconds
  cache_ttl_disk: 3600   # seconds
  comment_depth_limit: 10
  request_timeout: 10    # seconds
  retry_count: 2
  ```

---

## 13. Dependencies

### 13.1 Required Libraries
- `github.com/rivo/tview` - TUI framework
- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/PuerkitoBio/goquery` - HTML parsing
- `github.com/mattn/go-runewidth` - CJK character width calculation
- `jaytaylor.com/html2text` - HTML to plain text conversion
- `github.com/gelembjuk/articletext` - Article content extraction (primary)

### 13.2 Optional Libraries (Fallback)
- `github.com/go-shiori/go-readability` - Article extraction fallback
- (Others TBD during implementation)

### 13.3 Standard Library
- `encoding/xml` - RSS parsing
- `net/http` - HTTP client
- `net/url` - URL parsing
- `time` - Timestamps and TTL
- `encoding/json` - Cache serialization

---

## 14. Implementation Phases

### 14.1 Phase 1: MVP (Priority)
**Goal**: Basic functionality working end-to-end

1. **Setup**:
   - Rename project to `gn-text`
   - Update module path in `go.mod`
   - Add `github.com/mattn/go-runewidth` dependency

2. **RSS Parsing**:
   - Implement `parseGeekNewsRSS()`
   - Add fixture `testdata/geeknews_feed.xml`
   - Write unit tests

3. **Comment Parsing**:
   - Implement `parseGeekNewsComments()`
   - Add fixture `testdata/geeknews_topic_comments.html`
   - Handle depth extraction from `style="--depth:N"`
   - Write unit tests

4. **UI Updates**:
   - Remove comment count from list view
   - Remove fire emoji (🔥)
   - Update URLs to GeekNews
   - Update status bar for loading/errors

5. **Text Handling**:
   - Implement `wrapTextWithRuneWidth()`
   - Test with Korean text samples

6. **Integration**:
   - Wire RSS parsing into `main.go`
   - Wire comment parsing into `ui.go`
   - Test end-to-end flow

7. **Manual Testing**:
   - Run through manual testing checklist (Section 9.3)

### 14.2 Phase 2: Caching (Post-MVP)
1. Implement `cache.go` module
2. Add memory cache with LRU eviction
3. Add disk cache with file-based storage
4. Integrate cache into fetch functions
5. Add 'r' key to invalidate cache

### 14.3 Phase 3: Article Extraction Intelligence (Post-MVP)
1. Add `extractor.go` module
2. Implement fallback chain for article extraction
3. Add success rate tracking
4. Implement dynamic prioritization algorithm
5. Store stats in `~/.cache/gn-text/extractor-stats.json`

### 14.4 Phase 4: Polish (Post-MVP)
1. Add `?` key for help screen
2. Improve error messages (more specific)
3. Add comment pagination ('n' key)
4. Optimize rendering performance
5. Add more unit tests

### 14.5 Phase 5: Documentation & Release (Post-MVP)
1. Write README.md (English)
2. Write README.ko.md (Korean)
3. Write ARCHITECTURE.md
4. Record demo GIF with `asciinema`
5. Set up `goreleaser` config
6. Set up GitHub Actions for releases
7. Tag v1.0.0 and release

---

## 15. Risk Mitigation

### 15.1 GeekNews DOM Changes
- **Risk**: HTML structure changes break parsing
- **Mitigation**:
  - Use RSS as primary source (more stable)
  - Implement selector fallback chains
  - Add integration tests to detect changes early
  - Provide clear error messages with GitHub issue link

### 15.2 RSS Feed Unavailability
- **Risk**: RSS feed goes down or is removed
- **Mitigation**:
  - Add fallback to HTML homepage parsing
  - Cache data aggressively (disk cache)
  - Show last cached data with staleness indicator

### 15.3 Korean Text Rendering Issues
- **Risk**: Terminal doesn't support UTF-8 or CJK characters
- **Mitigation**:
  - Document UTF-8 terminal requirement in README
  - Use runewidth for width calculation (handles fallback)
  - Test on multiple terminals (iTerm2, gnome-terminal, Windows Terminal)

### 15.4 Article Extraction Failures
- **Risk**: Extraction libraries fail on Korean sites
- **Mitigation**:
  - Implement fallback chain with multiple libraries
  - Learn success rates and prioritize working libraries
  - Always provide "open in browser" option

### 15.5 Performance Issues
- **Risk**: App feels slow due to network/parsing
- **Mitigation**:
  - Implement caching (memory + disk)
  - Show loading indicators immediately
  - Add request timeouts (10s)
  - Consider goroutines for parallel requests (future)

---

## 16. Open Questions & Future Enhancements

### 16.1 Resolved in Specification
- ✅ RSS vs HTML: RSS-first for list
- ✅ Caching strategy: Memory + disk with TTL
- ✅ Comment depth limit: 10 levels with `[...]`
- ✅ Korean text wrapping: Use runewidth
- ✅ Error handling: Status bar + auto retry
- ✅ MVP scope: List + comments + navigation only

### 16.2 Future Enhancements (Not in Spec)
- Search within article list
- Filter by date/score
- Bookmarking articles
- Export to markdown
- Multiple site support (HN + GeekNews)
- Real-time refresh (WebSocket)
- User authentication (if GeekNews adds API)

---

## 17. Success Criteria

### 17.1 Functional Requirements
- ✅ App launches without errors
- ✅ Article list loads from GeekNews RSS
- ✅ Selecting article opens comments
- ✅ Comments display with correct nesting (up to 10 levels)
- ✅ Korean text displays correctly without corruption
- ✅ Line wrapping respects CJK character widths
- ✅ Browser opens for article/comment links
- ✅ Refresh ('r') reloads data
- ✅ Quit ('q') exits cleanly
- ✅ Error messages display in status bar
- ✅ Retry logic handles transient network errors

### 17.2 Non-Functional Requirements
- ✅ App starts within 1 second
- ✅ List loads within 3 seconds (without cache)
- ✅ Comments load within 2 seconds
- ✅ Article extraction completes within 10 seconds (or fails gracefully)
- ✅ Memory usage under 100MB
- ✅ No crashes or panics during normal operation
- ✅ Code coverage >70% for parsing functions

### 17.3 User Experience
- ✅ Keyboard navigation feels responsive (<100ms latency)
- ✅ Error messages are clear and actionable
- ✅ Korean text is readable (no overlapping characters)
- ✅ Loading states are visible (not frozen)
- ✅ Empty states are clear ("내용 없음")

---

## 18. Glossary

- **GeekNews**: Korean tech news aggregator (news.hada.io)
- **TUI**: Terminal User Interface
- **RSS**: Really Simple Syndication (XML feed format)
- **CJK**: Chinese, Japanese, Korean character sets
- **Runewidth**: Display width of Unicode characters (1 or 2 cells)
- **Depth**: Nesting level of comments (0 = top-level)
- **TTL**: Time To Live (cache expiration duration)
- **LRU**: Least Recently Used (cache eviction policy)
- **MVP**: Minimum Viable Product (initial release scope)

---

## 19. References

- Original HN design: `docs/original-design.md`
- GeekNews research: `docs/geeknews-design-plan.md`
- GeekNews site: https://news.hada.io/
- GeekNews RSS: https://news.hada.io/rss/news
- tview docs: https://github.com/rivo/tview
- goquery docs: https://github.com/PuerkitoBio/goquery
- runewidth docs: https://github.com/mattn/go-runewidth

---

**Document End**
