package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// IsGitRepo checks if the current directory is inside a git repository
func IsGitRepo() bool {
	err := exec.Command("git", "rev-parse", "--git-dir").Run()
	return err == nil
}

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

	// Get commits with format: hash|subject|author|timestamp
	// Use %x00 (null byte) as record separator to handle multi-line content
	// Skip body in the list view - we only need subject for display
	format := "%H%x01%s%x01%an%x01%at%x00"
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", limit), fmt.Sprintf("--format=%s", format), "--no-merges")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	// Split by null byte to get individual commits
	records := strings.Split(strings.TrimSpace(string(output)), "\x00")
	if len(records) == 0 || (len(records) == 1 && records[0] == "") {
		return nil, fmt.Errorf("no commits found")
	}

	var commits []Commit
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}

		// Split by unit separator (0x01)
		parts := strings.Split(record, "\x01")
		if len(parts) < 4 {
			continue
		}

		var timestamp time.Time
		if ts, err := parseUnixTimestamp(parts[3]); err == nil {
			timestamp = ts
		}

		commits = append(commits, Commit{
			Hash:      parts[0][:7], // Short hash
			Subject:   parts[1],
			Body:      "", // Body fetched separately if needed
			Author:    parts[2],
			Timestamp: timestamp,
			Ago:       timeAgo(timestamp),
		})
	}

	return commits, nil
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
