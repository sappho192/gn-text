# GeekNews Design Plan

## Goal
Replace HN-specific fetching/parsing with GeekNews (news.hada.io) while preserving the current TUI UX (list → comments → article) and keybindings.

## Key Decisions
- **Primary list source**: Prefer GeekNews RSS feed for stability and simpler parsing.
- **Fallback**: If RSS lacks fields (comment count, comment link), parse homepage HTML.
- **Comments**: Parse GeekNews comment page HTML (nested/threaded if present).
- **Keep UI**: Reuse `tview` layout, keybindings, and article extraction.

## Known URLs (to verify in implementation)
- Base: `https://news.hada.io/`
- RSS: `https://news.hada.io/rss/news`
- Topic: `https://news.hada.io/topic?id=...`
- Comments: `https://news.hada.io/topic?go=comments&id=...`

## Data Model Updates
- Extend `Article` to include fields if needed (author, points, time) but keep required fields:
  - `Title`, `Link` (external article), `Comments`, `CommentsLink` (GeekNews topic URL).
- Decide whether `CommentsLink` should store relative paths or full URLs; standardize in one place.

## Parsing Strategy

### 1) List Parsing
Option A (RSS-first):
- Fetch RSS feed and parse with `encoding/xml`.
- Map RSS items to `Article` fields.
- If RSS does not include comment count/link, enrich by requesting homepage HTML.

Option B (HTML-first):
- Fetch homepage HTML and extract:
  - Title text and external link.
  - Comment count text (e.g., “댓글 N개”).
  - Topic/comment URL for comments page.
- Use goquery selectors based on GeekNews DOM.

### 2) Comments Parsing
- Fetch comments page HTML.
- Use goquery to locate comment blocks.
- For each comment, extract:
  - Author
  - Body HTML → text (reuse `sanitize`)
  - Nesting level (if present in DOM or via tree structure).
- Render with the existing `appendComment` style (indent + wrapped lines).

## Code Changes (by file)

### `main.go`
- Replace `hackerNewsURL` with `geekNewsURL`.
- Allow page selection for GeekNews categories if applicable (future extension).

### `parser.go`
- Implement `parseGeekNewsArticles(...)`.
- If using RSS, add `parseGeekNewsRSS(...)` and call from `parseArticles(...)` or new entry point.
- Update tests to use GeekNews fixtures (RSS or HTML samples).

### `web.go`
- Add `fetchGeekNewsComments(topicID or URL)` to replace Algolia API usage.
- Keep `fetchWebpage()` and `sanitize()` as shared utilities.
- Adjust `openComments()` flow to use GeekNews identifiers.

### `ui.go`
- Update comments URL opener to use GeekNews base URL.
- Ensure `openComments()` can parse GeekNews topic IDs from URLs.
- Keep article opening logic the same (`articletext` still used).

## Implementation Steps
1. Inspect GeekNews DOM and RSS to confirm selectors and field availability.
2. Decide RSS-first vs HTML-first parsing based on completeness of fields.
3. Implement list parsing for GeekNews (new functions + tests).
4. Implement comment parsing for GeekNews (new functions + tests).
5. Wire parsing into `main.go` / `ui.go` and replace HN-specific constants.
6. Update tests and add fixtures for GeekNews HTML/RSS.

## Test Plan
- Unit tests for list parsing (RSS or HTML).
- Unit tests for comment parsing with sample GeekNews comments HTML.
- Manual smoke test: open list, comments, and article.

## Risks / Open Questions
- GeekNews DOM or RSS format might change; prefer RSS when possible.
- Comment nesting depth and structure may not map directly to HN; indentation rules may need adjustment.
- Some posts may not have external links (e.g., “Ask” style); decide how to handle in `openArticle()`.

## Concrete Selectors (Observed 2026-02-03)

### Homepage list (`https://news.hada.io/`)
- Container: `div.topics`
- Item: `div.topic_row`
- Title text: `.topictitle h1`
- External link (article URL): `.topictitle > a`
- Domain display: `.topictitle .topicurl`
- Topic page link: `.topicdesc > a[href^='topic?id=']`
- Points: `.topicinfo span[id^='tp']`
- Author: `.topicinfo a[href^='/user?id=']`
- Comment link: `.topicinfo a[href*='go=comments']`
- Comment count text: `.topicinfo` includes either `댓글 N개` or `댓글과 토론`
  - Parse digits when present; treat “댓글과 토론” as 0.

### Topic header (`https://news.hada.io/topic?id=...`)
- Container: `div.topic-table`
- External link (article URL): `.topictitle.link > a`
- Title text: `.topictitle.link h1`
- Comment link/count: `.topicinfo a[href^='topic?id=']` (text like `댓글 14개`)

### Comments (`https://news.hada.io/topic?go=comments&id=...`)
- Container: `#comment_thread`
- Comment row: `.comment_row`
- Nesting depth: `style="--depth:N"` on `.comment_row` (parse `N` from style)
- Author: `.commentinfo a[href^='/user?id=']` (first anchor)
- Time link: `.commentinfo a[href^='comment?id=']`
- Body HTML: `.commentTD .comment_contents` (sanitize to text)

## Fixtures Drafted
- `testdata/geeknews_feed.xml`: Atom feed header + 2 entries.
- `testdata/geeknews_homepage_topics.html`: `div.topics` block with multiple `topic_row` items.
- `testdata/geeknews_topic_header.html`: topic header block including external link and title.
- `testdata/geeknews_topic_comments.html`: `#comment_thread` block with multiple nested comments.
