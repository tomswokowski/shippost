package ai

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tomswokowski/shippost/git"
)

// GeneratePostSuggestion uses Claude Code CLI to generate a post suggestion
// Returns a slice of posts (thread) - may be single post or multiple
func GeneratePostSuggestion(commits []git.Commit, prompt string, allowThread bool) ([]string, error) {
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits provided")
	}

	// Build context from commits
	var context strings.Builder
	context.WriteString("Based on the following git commit(s), write an engaging post for X (formerly Twitter).\n\n")
	context.WriteString("CRITICAL RULES:\n")
	context.WriteString("- EACH post MUST be UNDER 280 characters - this is a hard limit, count carefully!\n")
	context.WriteString("- NEVER cut off in the middle of a word - if you're close to 280, end the sentence earlier\n")
	if allowThread {
		context.WriteString("- If the content is rich enough, write a thread (2-4 posts)\n")
		context.WriteString("- If a single post works, that's fine too\n")
	} else {
		context.WriteString("- Write exactly ONE post, not a thread\n")
	}
	context.WriteString("- Be concise and highlight what was accomplished\n")
	context.WriteString("- Don't use hashtags unless really relevant\n")
	context.WriteString("- Sound natural, not promotional\n\n")

	if prompt != "" {
		context.WriteString("User's guidance: ")
		context.WriteString(prompt)
		context.WriteString("\n\n")
	}

	for i, commit := range commits {
		context.WriteString(fmt.Sprintf("Commit %d:\n", i+1))
		context.WriteString(fmt.Sprintf("  Message: %s\n", commit.Subject))
		if commit.Body != "" {
			context.WriteString(fmt.Sprintf("  Details: %s\n", commit.Body))
		}
		context.WriteString(fmt.Sprintf("  When: %s\n", commit.Ago))
		context.WriteString("\n")
	}

	if allowThread {
		context.WriteString("Write only the post text. If writing a thread, separate posts with ---\n")
		context.WriteString("Example thread format:\n")
		context.WriteString("First post here\n---\nSecond post here\n---\nThird post here\n\n")
	} else {
		context.WriteString("Write only the post text (one post, no thread).\n\n")
	}
	context.WriteString("Output:")

	// Call claude CLI
	cmd := exec.Command("claude", "-p", context.String())
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("claude error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run claude: %w", err)
	}

	return parseThreadResponse(string(output)), nil
}

// GeneratePostFromDiff uses Claude Code CLI to generate a post from a diff
func GeneratePostFromDiff(commitHash string) (string, error) {
	// Get the diff for this commit
	diffCmd := exec.Command("git", "show", "--stat", commitHash)
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	// Build prompt
	var prompt strings.Builder
	prompt.WriteString("Based on this git commit diff, write a short, engaging post for X (formerly Twitter). ")
	prompt.WriteString("Keep it under 280 characters. Be concise and highlight what was accomplished. ")
	prompt.WriteString("Don't use hashtags unless they're really relevant. Sound natural, not promotional.\n\n")
	prompt.WriteString("```\n")
	prompt.WriteString(string(diffOutput))
	prompt.WriteString("```\n\n")
	prompt.WriteString("Write only the post text, nothing else:")

	// Call claude CLI
	cmd := exec.Command("claude", "-p", prompt.String())
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("claude error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to run claude: %w", err)
	}

	suggestion := strings.TrimSpace(string(output))
	suggestion = strings.Trim(suggestion, "\"'")

	return suggestion, nil
}

// GenerateFromQuery uses natural language query to generate a post from commits
// Returns a slice of posts (thread) - may be single post or multiple
func GenerateFromQuery(query string, commits []git.Commit, allowThread bool) ([]string, error) {
	if query == "" {
		return nil, fmt.Errorf("no query provided")
	}

	// Build context with commits and their diffs
	var context strings.Builder
	context.WriteString("You are helping a developer write an engaging X (Twitter) post about their coding work.\n\n")
	context.WriteString("Their question/request: ")
	context.WriteString(query)
	context.WriteString("\n\n")
	context.WriteString("Here are their recent git commits with diffs:\n\n")

	for i, commit := range commits {
		if i >= 20 { // Limit to 20 commits
			break
		}
		context.WriteString(fmt.Sprintf("--- Commit: %s (%s) ---\n", commit.Subject, commit.Ago))

		// Get diff for this commit
		diffCmd := exec.Command("git", "show", "--stat", "--no-color", commit.Hash)
		if diffOutput, err := diffCmd.Output(); err == nil {
			context.WriteString(string(diffOutput))
		}
		context.WriteString("\n")
	}

	context.WriteString("\nCRITICAL RULES:\n")
	context.WriteString("- EACH post MUST be UNDER 280 characters - this is a hard limit, count carefully!\n")
	context.WriteString("- NEVER cut off in the middle of a word - if you're close to 280, end the sentence earlier\n")
	if allowThread {
		context.WriteString("- If the content is rich enough, write a thread (2-4 posts)\n")
		context.WriteString("- If a single post works, that's fine too\n")
	} else {
		context.WriteString("- Write exactly ONE post, not a thread\n")
	}
	context.WriteString("- Be concise and insightful\n")
	context.WriteString("- Don't use hashtags unless really relevant\n")
	context.WriteString("- Sound natural, not promotional\n\n")
	if allowThread {
		context.WriteString("Write only the post text. If writing a thread, separate posts with ---\n")
	} else {
		context.WriteString("Write only the post text (one post, no thread).\n")
	}
	context.WriteString("Output:")

	// Call claude CLI
	cmd := exec.Command("claude", "-p", context.String())
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("claude error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run claude: %w", err)
	}

	return parseThreadResponse(string(output)), nil
}

// IsClaudeAvailable checks if the claude CLI is installed and accessible
func IsClaudeAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// parseThreadResponse splits the AI response into individual posts
func parseThreadResponse(output string) []string {
	output = strings.TrimSpace(output)
	output = strings.Trim(output, "\"'")

	// Split by --- separator
	parts := strings.Split(output, "---")

	var posts []string
	for _, part := range parts {
		post := strings.TrimSpace(part)
		if post != "" {
			posts = append(posts, post)
		}
	}

	// If no posts found, return the whole output as single post
	if len(posts) == 0 {
		return []string{output}
	}

	return posts
}
