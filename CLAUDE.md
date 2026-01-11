# shippost Development Guide

## Overview

shippost is a Go TUI application for posting to X (Twitter) from the terminal. It features AI-powered post suggestions generated from git commits using Claude Code CLI.

## Project Structure

```
shippost/
├── main.go              # Entry point, CLI flag parsing
├── config/
│   └── config.go        # Credential storage (~/.config/shippost/config.json)
├── tui/
│   └── tui.go           # Bubbletea TUI (all screens and views)
├── x/
│   └── client.go        # X API client (OAuth 1.0a, posting, media upload)
├── ai/
│   └── claude.go        # Claude Code CLI integration for AI suggestions
├── git/
│   └── git.go           # Git operations (commit history, diffs)
├── .goreleaser.yaml     # Release automation config
├── .github/workflows/
│   └── release.yml      # GitHub Actions for releases
└── install.sh           # Quick install script
```

## Architecture

### TUI (tui/tui.go)

Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

**Screens (viewState enum):**
- `viewHome` - Main menu (Quick Post, Smart Post)
- `viewQuickPost` - Direct text composition
- `viewSmartMenu` - Smart Post sub-menu (Browse Commits, Ask)
- `viewSmartBrowse` - Commit browser with multi-select
- `viewSmartAsk` - Natural language query input
- `viewSmartCompose` - Edit AI-generated posts before sending
- `viewPosting` - Loading state while posting
- `viewSuccess` - Post confirmation with URL

**Key Model Fields:**
- `posts []postItem` - Thread posts (text + optional media)
- `currentPost int` - Active post in thread
- `commits []git.Commit` - Loaded git commits
- `selectedCommits map[int]bool` - Multi-select for Browse mode
- `allowThread bool` - Single post vs thread toggle

**Theme Detection:**
- `lipgloss.HasDarkBackground()` auto-detects terminal theme
- Two complete color palettes (dark/light) in `initStyles()`

### X Client (x/client.go)

- OAuth 1.0a authentication via `github.com/dghubble/oauth1`
- Posts to X API v2 (`https://api.x.com/2/tweets`)
- Media upload via v1.1 API (`https://upload.twitter.com/1.1/media/upload.json`)
- Thread support via `in_reply_to_tweet_id`
- Supports images (jpg, png, gif, webp)

### AI Integration (ai/claude.go)

- Requires Claude Code CLI (`claude`) in PATH
- Three generation modes:
  - `GeneratePostSuggestion()` - From selected commits
  - `GenerateFromQuery()` - Natural language questions
  - `GeneratePostFromDiff()` - From commit diff (unused currently)
- Thread posts separated by `---` in AI response
- Prompts enforce 280 char limit and natural tone

### Git Integration (git/git.go)

- `IsGitRepo()` - Checks if in git repository
- `GetRecentCommits(limit)` - Fetches commit history
- Uses null-byte separators for reliable parsing of multi-line commits
- Human-readable time ago formatting

## Key Dependencies

```
github.com/charmbracelet/bubbletea  - TUI framework
github.com/charmbracelet/bubbles    - TUI components (textarea, textinput)
github.com/charmbracelet/lipgloss   - TUI styling
github.com/dghubble/oauth1          - OAuth 1.0a for X API
golang.org/x/term                   - Hidden password input
```

## Development

### Build and Run
```bash
go build -o shippost && ./shippost
```

### Version Management
- Local builds use `version = "dev"` (set in main.go)
- Release versions injected by GoReleaser via `-ldflags`
- Current release: v0.0.1 (do not change without user approval)

### Creating Releases
```bash
git tag v0.0.2
git push origin v0.0.2
```
GitHub Actions runs GoReleaser which:
1. Builds binaries for Linux/macOS/Windows (amd64/arm64)
2. Creates GitHub release with assets
3. Updates Homebrew formula in `tomswokowski/homebrew-tap`

### Testing Locally
```bash
# Setup credentials first
./shippost --setup

# Run the TUI
./shippost

# Clean up credentials
./shippost --cleanup
```

## Configuration

Credentials stored at `~/.config/shippost/config.json`:
```json
{
  "api_key": "...",
  "api_secret": "...",
  "access_token": "...",
  "access_secret": "..."
}
```
- File permissions: 0600
- Directory permissions: 0700

## Common Tasks

### Adding a New Screen
1. Add constant to `viewState` enum in tui.go
2. Add case in `View()` method
3. Add keyboard handling in `Update()` method
4. Add help bar in relevant view function

### Modifying AI Prompts
Edit prompts in `ai/claude.go`. Key considerations:
- Enforce 280 char limit explicitly
- Request `---` separators for threads
- Keep prompts concise to reduce latency

### Adding New Keyboard Shortcuts
1. Handle in appropriate `case` in `Update()` method
2. Add to help bar in relevant view
3. Document in README.md

## Distribution

**Installation methods:**
- Homebrew: `brew tap tomswokowski/tap && brew install shippost`
- Script: `curl -fsSL .../install.sh | bash`
- Go: `go install github.com/tomswokowski/shippost@latest`
- Binary: Download from GitHub Releases

**Homebrew tap:** `github.com/tomswokowski/homebrew-tap`
- Formula auto-updated by GoReleaser on each release

## External Requirements

- X API credentials (Free tier: 500 posts/month)
- Claude Code CLI (optional, for Smart Post features)
- Git repository (optional, for Smart Post features)
