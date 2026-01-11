package ai

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/tomswokowski/shippost/git"
)

// validGitHash matches a valid git commit hash (7-40 hex characters)
var validGitHash = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

// GeneratePostSuggestion uses Claude Code CLI to generate a post suggestion
// Returns a slice of posts (thread) - may be single post or multiple
func GeneratePostSuggestion(commits []git.Commit, prompt string, allowThread bool) ([]string, error) {
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits provided")
	}

	var context strings.Builder
	context.WriteString("Based on the following git commit(s), write an engaging post for X (formerly Twitter).\n\n")
	writePromptRules(&context, allowThread)

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

	writeOutputFormat(&context, allowThread)

	return runClaude(context.String())
}

// GenerateFromQuery uses natural language query to generate a post from commits
// Returns a slice of posts (thread) - may be single post or multiple
func GenerateFromQuery(query string, commits []git.Commit, allowThread bool) ([]string, error) {
	if query == "" {
		return nil, fmt.Errorf("no query provided")
	}

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

		// Get diff for this commit (validate hash first to prevent command injection)
		if validGitHash.MatchString(commit.Hash) {
			diffCmd := exec.Command("git", "show", "--stat", "--no-color", commit.Hash)
			if diffOutput, err := diffCmd.Output(); err == nil {
				context.WriteString(string(diffOutput))
			}
		}
		context.WriteString("\n")
	}

	context.WriteString("\n")
	writePromptRules(&context, allowThread)
	writeOutputFormat(&context, allowThread)

	return runClaude(context.String())
}

// IsClaudeAvailable checks if the claude CLI is installed and accessible
func IsClaudeAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// writePromptRules writes the common rules for post generation
func writePromptRules(b *strings.Builder, allowThread bool) {
	b.WriteString("CRITICAL RULES:\n")
	b.WriteString("- EACH post MUST be UNDER 280 characters - this is a hard limit, count carefully!\n")
	b.WriteString("- NEVER cut off in the middle of a word - if you're close to 280, end the sentence earlier\n")
	if allowThread {
		b.WriteString("- If the content is rich enough, write a thread (2-4 posts)\n")
		b.WriteString("- If a single post works, that's fine too\n")
	} else {
		b.WriteString("- Write exactly ONE post, not a thread\n")
	}
	b.WriteString("- Be concise and highlight what was accomplished\n")
	b.WriteString("- Don't use hashtags unless really relevant\n")
	b.WriteString("- Sound natural, not promotional\n\n")
}

// writeOutputFormat writes the output format instructions
func writeOutputFormat(b *strings.Builder, allowThread bool) {
	if allowThread {
		b.WriteString("Write only the post text. If writing a thread, separate posts with ---\n")
		b.WriteString("Example thread format:\n")
		b.WriteString("First post here\n---\nSecond post here\n---\nThird post here\n\n")
	} else {
		b.WriteString("Write only the post text (one post, no thread).\n\n")
	}
	b.WriteString("Output:")
}

// runClaude executes the claude CLI and parses the response
func runClaude(prompt string) ([]string, error) {
	cmd := exec.Command("claude", "-p", prompt)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("claude error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run claude: %w", err)
	}

	return parseThreadResponse(string(output)), nil
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
