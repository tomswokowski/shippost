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
│   ├── tui.go           # Model, state machine, key handlers
│   ├── styles.go        # Lipgloss styles (dark/light themes)
│   ├── views.go         # View rendering for all screens
│   └── commands.go      # Async commands (API calls, AI generation)
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

### TUI (tui/)

Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

**File Organization:**
- `tui.go` - Model struct, New(), Init(), Update(), key handlers
- `styles.go` - All lipgloss style definitions, theme detection
- `views.go` - View() and all screen rendering functions
- `commands.go` - Async tea.Cmd functions for API/AI calls

**States (state enum):**
- `stateHome` - Main menu (Quick Post, Smart Post)
- `stateCompose` - Quick Post text composition
- `stateSmartMenu` - Smart Post sub-menu (Browse Commits, Ask)
- `stateCommitBrowser` - Commit browser with multi-select
- `stateAskInput` - Natural language query input
- `stateGenerating` - AI generation in progress
- `stateSmartCompose` - Edit AI-generated posts before sending
- `stateMediaInput` - File path input for attachments
- `statePosting` - Posting in progress
- `statePosted` - Success confirmation with URL

**Key Model Fields:**
- `thread []threadItem` - Thread posts (text + optional media)
- `currentPost int` - Active post in thread
- `commits []git.Commit` - Loaded git commits
- `selectedCommits []int` - Multi-select indices for Browse mode
- `allowThread bool` - Single post vs thread toggle

**Commit Browser Keys:**
- `space` - Toggle single commit selection
- `a` - Toggle select all (works with filtered results)

**Theme Detection:**
- `lipgloss.HasDarkBackground()` auto-detects terminal theme
- Two complete color palettes (dark/light) in `styles.go`

**Terminal Size:**
- Minimum: 128×30 characters (shows warning if smaller)
- Commit browser shows 5 commits at a time (scrollable, loads 50 total)

### X Client (x/client.go)

- OAuth 1.0a authentication via `github.com/dghubble/oauth1`
- Posts to X API v2 (`https://api.x.com/2/tweets`)
- Media upload via v1.1 API (`https://upload.twitter.com/1.1/media/upload.json`)
- Thread support via `in_reply_to_tweet_id`
- Supports images (jpg, png, gif, webp)

### AI Integration (ai/claude.go)

- Requires Claude Code CLI (`claude`) in PATH
- Two generation modes:
  - `GeneratePostSuggestion()` - From selected commits
  - `GenerateFromQuery()` - Natural language questions
- Thread posts separated by `---` in AI response
- Prompts enforce 280 char limit and natural tone
- Git hash validation to prevent command injection

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
1. Add constant to `state` enum in tui/tui.go
2. Add view function in tui/views.go
3. Add case in `View()` switch in tui/views.go
4. Add key handler function in tui/tui.go
5. Add case in `Update()` switch to call the handler

### Modifying AI Prompts
Edit prompts in `ai/claude.go`. Key considerations:
- Enforce 280 char limit explicitly
- Request `---` separators for threads
- Keep prompts concise to reduce latency

### Adding New Keyboard Shortcuts
1. Handle in appropriate handler function in tui/tui.go
2. Add to help bar in relevant view function in tui/views.go
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
