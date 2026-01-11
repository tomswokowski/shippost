package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tomswokowski/shippost/ai"
	"github.com/tomswokowski/shippost/git"
	"github.com/tomswokowski/shippost/x"
)

// Message types for async operations

type postResultMsg struct {
	urls []string
	err  error
}

type mediaUploadMsg struct {
	mediaID string
	path    string
	err     error
}

type commitsLoadedMsg struct {
	commits []git.Commit
	err     error
}

type aiSuggestionMsg struct {
	suggestions []string
	err         error
}

// Command functions

func (m Model) loadCommits() tea.Cmd {
	return func() tea.Msg {
		commits, err := git.GetRecentCommits(50)
		if err != nil {
			return commitsLoadedMsg{err: err}
		}
		return commitsLoadedMsg{commits: commits}
	}
}

func (m Model) generateSuggestion() tea.Cmd {
	prompt := m.commitPromptInput.Value()
	allowThread := m.allowThread
	return func() tea.Msg {
		var selectedCommits []git.Commit
		for _, idx := range m.selectedCommits {
			if idx < len(m.commits) {
				selectedCommits = append(selectedCommits, m.commits[idx])
			}
		}

		if !ai.IsClaudeAvailable() {
			return aiSuggestionMsg{err: fmt.Errorf("claude CLI not found - install Claude Code first")}
		}

		suggestions, err := ai.GeneratePostSuggestion(selectedCommits, prompt, allowThread)
		if err != nil {
			return aiSuggestionMsg{err: err}
		}
		return aiSuggestionMsg{suggestions: suggestions}
	}
}

func (m Model) generateFromQuery() tea.Cmd {
	query := m.askQuery
	commits := m.commits
	allowThread := m.allowThread
	return func() tea.Msg {
		if !ai.IsClaudeAvailable() {
			return aiSuggestionMsg{err: fmt.Errorf("claude CLI not found - install Claude Code first")}
		}

		suggestions, err := ai.GenerateFromQuery(query, commits, allowThread)
		if err != nil {
			return aiSuggestionMsg{err: err}
		}
		return aiSuggestionMsg{suggestions: suggestions}
	}
}

func (m Model) uploadMedia(path string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.xClient.UploadMedia(path)
		if err != nil {
			return mediaUploadMsg{err: err}
		}
		return mediaUploadMsg{mediaID: resp.MediaIDString, path: path}
	}
}

func (m Model) doPost() tea.Cmd {
	return func() tea.Msg {
		var posts []x.ThreadPost
		for i, item := range m.thread {
			text := item.text
			if i == m.currentPost {
				text = m.textarea.Value()
			}
			if strings.TrimSpace(text) == "" && len(item.mediaIDs) == 0 {
				continue
			}
			posts = append(posts, x.ThreadPost{
				Text:     text,
				MediaIDs: item.mediaIDs,
			})
		}

		if len(posts) == 0 {
			return postResultMsg{err: fmt.Errorf("no content to post")}
		}

		if len(posts) == 1 {
			resp, err := m.xClient.PostWithOptions(posts[0].Text, &x.PostOptions{
				MediaIDs: posts[0].MediaIDs,
			})
			if err != nil {
				return postResultMsg{err: err}
			}
			return postResultMsg{urls: []string{fmt.Sprintf("https://x.com/i/status/%s", resp.Data.ID)}}
		}

		responses, err := m.xClient.PostThread(posts)
		if err != nil {
			return postResultMsg{err: err}
		}

		var urls []string
		for _, resp := range responses {
			urls = append(urls, fmt.Sprintf("https://x.com/i/status/%s", resp.Data.ID))
		}
		return postResultMsg{urls: urls}
	}
}
