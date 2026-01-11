# shippost

Post to X from your terminal.

## Installation

```bash
go install github.com/tom/shippost@latest
```

Or build from source:

```bash
git clone https://github.com/tom/shippost.git
cd shippost
go build -o shippost
```

## Setup

You need X API credentials with Read+Write permissions.

### 1. Create X Developer Account

1. Go to [developer.x.com](https://developer.x.com)
2. Sign up for a developer account (Free tier allows 1,500 posts/month)
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

**Security note:** Secrets are entered via hidden input and never logged.

## Usage

```bash
# Post to X
shippost "Just shipped a new feature!"

# Multiple words work without quotes too
shippost Working on some cool stuff

# Show help
shippost --help

# Show version
shippost --version
```

## Security

- Config file is stored with `0600` permissions (owner read/write only)
- Config directory uses `0700` permissions
- Credentials are never logged or printed
- Setup uses hidden input for secrets (won't appear in shell history)

## Future Plans

- TUI mode for browsing git history and drafting posts
- AI-powered post suggestions
- Post threads
- Media attachments
