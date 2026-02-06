# gn-text

Terminal-based [GeekNews](https://news.hada.io) reader for Korean tech news.

![Go Version](https://img.shields.io/github/go-mod/go-version/sappho192/gn-text)
![License](https://img.shields.io/github/license/sappho192/gn-text)
![Release](https://img.shields.io/github/v/release/sappho192/gn-text)

## Installation

### Go

```bash
go install github.com/sappho192/gn-text@latest
```

### Homebrew (macOS/Linux)

```bash
brew install sappho192/tap/gn-text
```

### Chocolatey (Windows)

```bash
choco install gn-text
choco install gn-text --version 0.1.3 # Use this until the package validation pass
```

### Manual Download

Download the binary from [GitHub Releases](https://github.com/sappho192/gn-text/releases).

## Usage

```bash
gn-text
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `l` / `→` | View comments / article |
| `h` / `←` | Go back |
| `Space` | Open article in browser |
| `c` | Open comments in browser |
| `r` | Refresh |
| `q` / `Ctrl+C` | Quit |

### Navigation Flow

```
Article List → Comments → Article Content
     ←─────────────←──────────────←
```

- Press `l` or `→` on the article list to view comments
- Press `l` or `→` on comments to view the article content
- Press `h` or `←` to go back

## Version

```bash
gn-text -v
# or
gn-text --version
```

## License

MIT
