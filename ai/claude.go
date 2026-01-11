package ai

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tom/shippost/git"
)

// GeneratePostSuggestion uses Claude Code CLI to generate a post suggestion
func GeneratePostSuggestion(commits []git.Commit, prompt string) (string, error) {
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits provided")
	}

	// Build context from commits
	var context strings.Builder
	context.WriteString("Based on the following git commit(s), write a short, engaging post for X (formerly Twitter). ")
	context.WriteString("Keep it under 280 characters. Be concise and highlight what was accomplished. ")
	context.WriteString("Don't use hashtags unless they're really relevant. Sound natural, not promotional.\n\n")

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

	context.WriteString("Write only the post text, nothing else:")

	// Call claude CLI
	cmd := exec.Command("claude", "-p", context.String())
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("claude error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to run claude: %w", err)
	}

	// Clean up the response
	suggestion := strings.TrimSpace(string(output))

	// Remove quotes if the response is wrapped in them
	suggestion = strings.Trim(suggestion, "\"'")

	return suggestion, nil
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
func GenerateFromQuery(query string, commits []git.Commit) (string, error) {
	if query == "" {
		return "", fmt.Errorf("no query provided")
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

	context.WriteString("\nBased on the above commits and the user's query, write a short, engaging post for X (formerly Twitter). ")
	context.WriteString("Keep it under 280 characters. Be concise and insightful. ")
	context.WriteString("Don't use hashtags unless they're really relevant. Sound natural, not promotional.\n\n")
	context.WriteString("Write only the post text, nothing else:")

	// Call claude CLI
	cmd := exec.Command("claude", "-p", context.String())
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

// IsClaudeAvailable checks if the claude CLI is installed and accessible
func IsClaudeAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}
