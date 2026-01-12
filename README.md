# shippost

Post to X from your terminal with AI-powered suggestions from your git commits.

## Features

- **Quick Post** - Write and post directly to X
- **Smart Post** - AI-powered posts from your git commits using Claude
  - Browse commits and select what to post about
  - Ask natural language questions like "What did I accomplish today?"
  - Generate threads or single posts
- **Thread support** - Create multi-post threads
- **Media attachments** - Attach images to your posts
- **Automatic theming** - Adapts to light or dark terminal backgrounds

## Installation

### Quick install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/tomswokowski/shippost/main/install.sh | bash
```

### Homebrew (macOS/Linux)

```bash
brew tap tomswokowski/tap
brew install shippost
```

### Go install

```bash
go install github.com/tomswokowski/shippost@latest
```

### Download binary

Download the latest release for your platform from [GitHub Releases](https://github.com/tomswokowski/shippost/releases).

## Setup

You need X API credentials with Read+Write permissions.

### 1. Create X Developer Account

1. Go to [developer.x.com](https://developer.x.com)
2. Sign up for a developer account (Free tier allows 500 posts/month)
3. Create a new Project and App

### 2. Generate Credentials

In your app settings:

1. Go to "Keys and tokens"
2. Generate **API Key and Secret** (also called Consumer Key/Secret)
3. Generate **Access Token and Secret** with Read+Write permissions

### 3. Configure shippost

```bash
shippost --setup
```

Enter your credentials when prompted. They're stored securely at `~/.config/shippost/config.json` with restricted permissions.

### 4. (Optional) Install Claude Code

For AI-powered Smart Post features, install [Claude Code](https://claude.ai/code):

```bash
npm install -g @anthropic-ai/claude-code
```

## Usage

```bash
# Launch the TUI
shippost

# Configure credentials
shippost --setup

# Remove stored credentials
shippost --cleanup

# Show help
shippost --help
```

### Keyboard shortcuts

**Home screen:**
- `↑/↓` or `j/k` - Navigate menu
- `Enter` - Select
- `q` - Quit

**Quick Post:**
- `ctrl+s` - Send post
- `ctrl+o` - Attach image
- `ctrl+n` - Add post to thread
- `ctrl+d` - Delete post from thread
- `ctrl+j/k` - Navigate thread
- `esc` - Back

**Smart Post (Browse Commits):**
- `↑/↓` - Navigate commits
- `Space` - Select/deselect commit
- `a` - Select/deselect all commits
- `/` - Search commits
- `Tab` - Focus prompt input
- `ctrl+t` - Toggle single/thread mode
- `Enter` - Generate post

**Smart Post (Ask Mode):**
- `ctrl+enter` - Submit question
- `esc` - Back

**Smart Post (Compose):**
- `ctrl+s` - Send
- `ctrl+r` - Regenerate
- `ctrl+n` - Add post
- `ctrl+b/f` - Navigate thread

## Security

- Config file is stored with `0600` permissions (owner read/write only)
- Config directory uses `0700` permissions
- Credentials are never logged or printed
- Setup uses hidden input for secrets

## Requirements

- X API credentials (Free tier available)
- Claude Code CLI (optional, for Smart Post features)
- Git repository (optional, for Smart Post features)
- Terminal size: minimum 128×30 characters

## License

MIT
