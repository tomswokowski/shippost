package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Commit represents a git commit
type Commit struct {
	Hash      string
	Subject   string
	Body      string
	Author    string
	Timestamp time.Time
	Ago       string
}

// GetRecentCommits returns the most recent commits from the current repo
func GetRecentCommits(limit int) ([]Commit, error) {
	// Check if we're in a git repo
	if err := exec.Command("git", "rev-parse", "--git-dir").Run(); err != nil {
		return nil, fmt.Errorf("not a git repository")
	}

	// Get commits with format: hash|subject|body|author|timestamp
	format := "%H|%s|%b|%an|%at"
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", limit), fmt.Sprintf("--format=%s", format), "--no-merges")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return nil, fmt.Errorf("no commits found")
	}

	var commits []Commit
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		var timestamp time.Time
		if ts, err := parseUnixTimestamp(parts[4]); err == nil {
			timestamp = ts
		}

		commits = append(commits, Commit{
			Hash:      parts[0][:7], // Short hash
			Subject:   parts[1],
			Body:      strings.TrimSpace(parts[2]),
			Author:    parts[3],
			Timestamp: timestamp,
			Ago:       timeAgo(timestamp),
		})
	}

	return commits, nil
}

// GetCommitDiff returns the diff for a specific commit
func GetCommitDiff(hash string) (string, error) {
	cmd := exec.Command("git", "show", "--stat", hash)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return string(output), nil
}

func parseUnixTimestamp(s string) (time.Time, error) {
	var ts int64
	_, err := fmt.Sscanf(s, "%d", &ts)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts, 0), nil
}

func timeAgo(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}
